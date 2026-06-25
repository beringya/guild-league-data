package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nsh-guild-analytics/backend/internal/config"
	"nsh-guild-analytics/backend/internal/domain"
)

type Store struct {
	cfg  config.Config
	pool *pgxpool.Pool
}

type ConfirmImportRequest struct {
	HomeGuild string    `json:"home_guild"`
	BattleAt  time.Time `json:"battle_at"`
	UserID    int64     `json:"user_id"`
}

type BattleSummary struct {
	ID             int64     `json:"id"`
	BattleAt       time.Time `json:"battle_at"`
	SourceFilename string    `json:"source_filename"`
	SourceSHA256   string    `json:"source_sha256"`
	HomeGuild      string    `json:"home_guild"`
	OpponentGuild  string    `json:"opponent_guild"`
	ValidRowCount  int       `json:"valid_row_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type BattleDetail struct {
	BattleSummary
	ScoringRuleVersion  string `json:"scoring_rule_version"`
	ScoringRangeVersion string `json:"scoring_range_version"`
	ImportStatus        string `json:"import_status"`
}

type Overview struct {
	Battle      BattleSummary        `json:"battle"`
	Totals      map[string]domain.Totals `json:"totals"`
	Advantages []Insight            `json:"advantages"`
	Weaknesses  []Insight            `json:"weaknesses"`
	TopHome     []domain.ScoredStat  `json:"top_home"`
	TopOpponent []domain.ScoredStat  `json:"top_opponent"`
}

type Insight struct {
	Metric  string  `json:"metric"`
	Label   string  `json:"label"`
	Home    float64 `json:"home"`
	Opponent float64 `json:"opponent"`
	Delta   float64 `json:"delta"`
	Message string  `json:"message"`
}

type RankingResponse struct {
	Items []domain.ScoredStat `json:"items"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

type PlayerDetail struct {
	Player       domain.ScoredStat       `json:"player"`
	Dimensions   []domain.DimensionScore  `json:"dimensions"`
	SameCareer   map[string]interface{}    `json:"same_career"`
	Trend        []map[string]interface{}  `json:"trend"`
	ScoreDetail  map[string]interface{}    `json:"score_detail"`
}

type GuildComparison struct {
	Battle  BattleSummary       `json:"battle"`
	Totals  map[string]domain.Totals `json:"totals"`
	PerCapita map[string]domain.Totals `json:"per_capita"`
	Careers []map[string]interface{} `json:"careers"`
	Insights []Insight `json:"insights"`
}

type SquadComparison struct {
	Battle BattleSummary `json:"battle"`
	Squads []map[string]interface{} `json:"squads"`
}

type HistoryRankingItem struct {
	PlayerID          int64     `json:"player_id"`
	PlayerName        string    `json:"player_name"`
	GuildName         string    `json:"guild_name"`
	Career            string    `json:"career"`
	MatchCount        int       `json:"match_count"`
	AverageScore      float64   `json:"average_score"`
	CumulativeScore   float64   `json:"cumulative_score"`
	BestScore         float64   `json:"best_score"`
	LatestScore       float64   `json:"latest_score"`
	LatestBattleAt    time.Time `json:"latest_battle_at"`
	RecentTrend       []float64 `json:"recent_trend"`
}

func NewStore(cfg config.Config, pool *pgxpool.Pool) *Store {
	return &Store{cfg: cfg, pool: pool}
}

func (s *Store) EnsureDefaultScoring(ctx context.Context) error {
	configBytes, err := json.Marshal(domain.CareerProfiles())
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO scoring_rule(version, name, status, config_json, published_at, is_active)
		VALUES('default-v1.5-final', '默认十三职业六维评分规则', 'published', $1::jsonb, now(), TRUE)
		ON CONFLICT (version) DO NOTHING
	`, string(configBytes))
	if err != nil {
		return err
	}
	rangeConfig := map[string]interface{}{"source": "import_snapshot", "published_ranges_are_frozen": true}
	rangeBytes, _ := json.Marshal(rangeConfig)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO scoring_range_version(version, name, config_json, source_method, sample_summary_json, is_active, is_frozen)
		VALUES('import-snapshot', '导入样本即时范围', $1::jsonb, 'same_battle_same_career_both_guilds', '{}'::jsonb, TRUE, TRUE)
		ON CONFLICT (version) DO NOTHING
	`, string(rangeBytes))
	return err
}

func (s *Store) SourceExists(ctx context.Context, sha256 string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM battle WHERE source_sha256=$1)`, sha256).Scan(&exists)
	return exists, err
}

func (s *Store) ConfirmImport(ctx context.Context, preview domain.ImportPreview, req ConfirmImportRequest) (int64, error) {
	if len(preview.Errors) > 0 {
		return 0, errors.New("preview contains validation errors")
	}
	if req.HomeGuild == "" {
		return 0, errors.New("home guild is required")
	}
	known := false
	for _, guild := range preview.Guilds {
		if guild.Name == req.HomeGuild {
			known = true
			break
		}
	}
	if !known {
		return 0, errors.New("home guild is not in preview")
	}
	if req.BattleAt.IsZero() {
		if preview.InferredBattleAt != nil {
			req.BattleAt = *preview.InferredBattleAt
		} else {
			req.BattleAt = time.Now()
		}
	}
	analysis := AnalyzeRows(preview.Rows, req.HomeGuild)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var battleID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO battle(
			battle_at, source_filename, source_sha256, original_row_count, valid_row_count,
			scoring_rule_version, scoring_range_version, import_status, match_result_json, created_by
		)
		VALUES($1,$2,$3,$4,$5,$6,$7,'analyzed',$8::jsonb,$9)
		RETURNING id
	`, req.BattleAt, preview.SourceFilename, preview.SourceSHA256, preview.OriginalRowCount, preview.ValidRowCount,
		analysis.ScoringRule, analysis.ScoringRange, mustJSON(map[string]interface{}{"home_guild": analysis.HomeGuild, "opponent_guild": analysis.OpponentGuild}), req.UserID).Scan(&battleID)
	if err != nil {
		return 0, err
	}

	guildIDs := map[string]int64{}
	battleGuildIDs := map[string]int64{}
	for _, gp := range preview.Guilds {
		guildID, err := upsertGuild(ctx, tx, gp.Name)
		if err != nil {
			return 0, err
		}
		guildIDs[gp.Name] = guildID
		side := "opponent"
		if gp.Name == req.HomeGuild {
			side = "home"
		}
		var battleGuildID int64
		err = tx.QueryRow(ctx, `
			INSERT INTO battle_guild(battle_id, guild_id, side, member_count)
			VALUES($1,$2,$3,$4)
			RETURNING id
		`, battleID, guildID, side, gp.MemberCount).Scan(&battleGuildID)
		if err != nil {
			return 0, err
		}
		battleGuildIDs[gp.Name] = battleGuildID
	}

	for _, row := range analysis.Rows {
		guildID := guildIDs[row.GuildName]
		playerID, err := upsertPlayer(ctx, tx, guildID, row.PlayerName)
		if err != nil {
			return 0, err
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO battle_player_stat(
				battle_id, battle_guild_id, player_id, guild_name_snapshot, player_name_snapshot,
				level, career, team_leader_snapshot, kills, assists, logistics, player_damage,
				building_damage, healing, damage_taken, deaths, qingdeng, revive, control,
				kda_ratio, player_damage_share, building_damage_share, player_damage_conversion_rate,
				building_damage_conversion_rate, participation_rate, composite_score, guild_rank,
				career_rank, team_rank, six_dimension_json, score_detail_json
			)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30::jsonb,$31::jsonb)
		`, battleID, battleGuildIDs[row.GuildName], playerID, row.GuildName, row.PlayerName, row.Level, row.Career, row.TeamLeader,
			row.Kills, row.Assists, row.Logistics, row.PlayerDamage, row.BuildingDamage, row.Healing, row.DamageTaken, row.Deaths,
			row.Qingdeng, row.Revive, row.Control, row.KDARatio, row.PlayerDamageShare, row.BuildingDamageShare,
			row.PlayerDamageConversionRate, row.BuildingDamageConversionRate, row.ParticipationRate, row.CompositeScore,
			row.GuildRank, row.CareerRank, row.TeamRank, mustJSON(row.SixDimensions), mustJSON(row.ScoreDetail)); err != nil {
			return 0, err
		}
	}
	for _, suggestion := range analysis.RangeSuggestions {
		if _, err = tx.Exec(ctx, `
			INSERT INTO range_suggestion(battle_id, career, metric, sample_size, method, suggested_min, suggested_max, margin_ratio, metadata_json)
			VALUES($1,$2,$3,$4,$5,$6,$7,0.1,$8::jsonb)
			ON CONFLICT (battle_id, career, metric) DO NOTHING
		`, battleID, suggestion.Career, suggestion.Metric, suggestion.SampleSize, suggestion.Method, suggestion.SuggestedMin, suggestion.SuggestedMax, mustJSON(map[string]interface{}{"warning": suggestion.Warning})); err != nil {
			return 0, err
		}
	}
	for _, msg := range append(preview.Warnings, preview.Errors...) {
		if _, err = tx.Exec(ctx, `
			INSERT INTO import_log(battle_id, level, code, row_number, message)
			VALUES($1,$2,$3,$4,$5)
		`, battleID, msg.Level, msg.Code, nullableInt(msg.RowNumber), msg.Message); err != nil {
			return 0, err
		}
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO app_setting(key, value_json, updated_at)
		VALUES('default_home_guild', to_jsonb($1::text), now())
		ON CONFLICT (key) DO UPDATE SET value_json=excluded.value_json, updated_at=now()
	`, req.HomeGuild); err != nil {
		return 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return battleID, nil
}

func (s *Store) ListBattles(ctx context.Context) ([]BattleSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT b.id, b.battle_at, b.source_filename, b.source_sha256, b.valid_row_count, b.created_at,
		       COALESCE(MAX(CASE WHEN bg.side='home' THEN g.name END), '') AS home_guild,
		       COALESCE(MAX(CASE WHEN bg.side='opponent' THEN g.name END), '') AS opponent_guild
		FROM battle b
		LEFT JOIN battle_guild bg ON bg.battle_id=b.id
		LEFT JOIN guild g ON g.id=bg.guild_id
		GROUP BY b.id
		ORDER BY b.battle_at DESC, b.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BattleSummary
	for rows.Next() {
		var item BattleSummary
		if err = rows.Scan(&item.ID, &item.BattleAt, &item.SourceFilename, &item.SourceSHA256, &item.ValidRowCount, &item.CreatedAt, &item.HomeGuild, &item.OpponentGuild); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) LatestBattleID(ctx context.Context) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `SELECT id FROM battle ORDER BY battle_at DESC, id DESC LIMIT 1`).Scan(&id)
	return id, err
}

func (s *Store) BattleDetail(ctx context.Context, id int64) (BattleDetail, error) {
	var detail BattleDetail
	err := s.pool.QueryRow(ctx, `
		SELECT b.id, b.battle_at, b.source_filename, b.source_sha256, b.valid_row_count, b.created_at,
		       b.scoring_rule_version, b.scoring_range_version, b.import_status,
		       COALESCE(MAX(CASE WHEN bg.side='home' THEN g.name END), '') AS home_guild,
		       COALESCE(MAX(CASE WHEN bg.side='opponent' THEN g.name END), '') AS opponent_guild
		FROM battle b
		LEFT JOIN battle_guild bg ON bg.battle_id=b.id
		LEFT JOIN guild g ON g.id=bg.guild_id
		WHERE b.id=$1
		GROUP BY b.id
	`, id).Scan(&detail.ID, &detail.BattleAt, &detail.SourceFilename, &detail.SourceSHA256, &detail.ValidRowCount, &detail.CreatedAt,
		&detail.ScoringRuleVersion, &detail.ScoringRangeVersion, &detail.ImportStatus, &detail.HomeGuild, &detail.OpponentGuild)
	return detail, err
}

func (s *Store) DeleteBattle(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM battle WHERE id=$1`, id)
	return err
}

func (s *Store) BattleRows(ctx context.Context, battleID int64) ([]domain.ScoredStat, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT bps.id, bg.side, bps.guild_name_snapshot, bps.player_name_snapshot, bps.level, bps.career, bps.team_leader_snapshot,
		       bps.kills, bps.assists, bps.logistics, bps.player_damage, bps.building_damage, bps.healing, bps.damage_taken,
		       bps.deaths, bps.qingdeng, bps.revive, bps.control, bps.kda_ratio, bps.player_damage_share,
		       bps.building_damage_share, bps.player_damage_conversion_rate, bps.building_damage_conversion_rate,
		       bps.participation_rate, bps.composite_score, bps.guild_rank, bps.career_rank, bps.team_rank,
		       bps.six_dimension_json, bps.score_detail_json
		FROM battle_player_stat bps
		JOIN battle_guild bg ON bg.id=bps.battle_guild_id
		WHERE bps.battle_id=$1
		ORDER BY bps.composite_score DESC NULLS LAST, bps.id ASC
	`, battleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ScoredStat
	for rows.Next() {
		var statID int64
		var dimsRaw, detailRaw []byte
		var item domain.ScoredStat
		if err = rows.Scan(&statID, &item.Side, &item.GuildName, &item.PlayerName, &item.Level, &item.Career, &item.TeamLeader,
			&item.Kills, &item.Assists, &item.Logistics, &item.PlayerDamage, &item.BuildingDamage, &item.Healing, &item.DamageTaken,
			&item.Deaths, &item.Qingdeng, &item.Revive, &item.Control, &item.KDARatio, &item.PlayerDamageShare,
			&item.BuildingDamageShare, &item.PlayerDamageConversionRate, &item.BuildingDamageConversionRate, &item.ParticipationRate,
			&item.CompositeScore, &item.GuildRank, &item.CareerRank, &item.TeamRank, &dimsRaw, &detailRaw); err != nil {
			return nil, err
		}
		item.RowNumber = int(statID)
		_ = json.Unmarshal(dimsRaw, &item.SixDimensions)
		_ = json.Unmarshal(detailRaw, &item.ScoreDetail)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) Overview(ctx context.Context, battleID int64) (Overview, error) {
	battle, err := s.BattleDetail(ctx, battleID)
	if err != nil {
		return Overview{}, err
	}
	rows, err := s.BattleRows(ctx, battleID)
	if err != nil {
		return Overview{}, err
	}
	totals := buildScoredTotals(rows)
	homeTop := filterTop(rows, "home", 3)
	opponentTop := filterTop(rows, "opponent", 3)
	insights := compareTotals(totals, battle.HomeGuild, battle.OpponentGuild)
	var advantages, weaknesses []Insight
	for _, item := range insights {
		if item.Delta >= 0 {
			advantages = append(advantages, item)
		} else {
			weaknesses = append(weaknesses, item)
		}
	}
	return Overview{Battle: battle.BattleSummary, Totals: totals, Advantages: advantages, Weaknesses: weaknesses, TopHome: homeTop, TopOpponent: opponentTop}, nil
}

func (s *Store) Rankings(ctx context.Context, battleID int64, side, career, team, search string, page, size int) (RankingResponse, error) {
	rows, err := s.BattleRows(ctx, battleID)
	if err != nil {
		return RankingResponse{}, err
	}
	filtered := filterRows(rows, side, career, team, search)
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].CompositeScore != filtered[j].CompositeScore {
			return filtered[i].CompositeScore > filtered[j].CompositeScore
		}
		return filtered[i].PlayerName < filtered[j].PlayerName
	})
	return paginate(filtered, page, size), nil
}

func (s *Store) PlayerDetail(ctx context.Context, battleID, statID int64) (PlayerDetail, error) {
	rows, err := s.BattleRows(ctx, battleID)
	if err != nil {
		return PlayerDetail{}, err
	}
	var player domain.ScoredStat
	found := false
	for _, row := range rows {
		if int64(row.RowNumber) == statID {
			player = row
			found = true
			break
		}
	}
	if !found {
		return PlayerDetail{}, pgx.ErrNoRows
	}
	sameCareer := sameCareerSummary(rows, player)
	trend, _ := s.PlayerTrend(ctx, player.GuildName, player.PlayerName, player.Career)
	return PlayerDetail{Player: player, Dimensions: player.SixDimensions, SameCareer: sameCareer, Trend: trend, ScoreDetail: player.ScoreDetail}, nil
}

func (s *Store) TeamTop3(ctx context.Context, battleID int64, side string) (map[string][]domain.ScoredStat, error) {
	rows, err := s.BattleRows(ctx, battleID)
	if err != nil {
		return nil, err
	}
	if side != "" {
		rows = filterRows(rows, side, "", "", "")
	}
	return buildTeamTop3(rows), nil
}

func (s *Store) GuildComparison(ctx context.Context, battleID int64) (GuildComparison, error) {
	battle, err := s.BattleDetail(ctx, battleID)
	if err != nil {
		return GuildComparison{}, err
	}
	rows, err := s.BattleRows(ctx, battleID)
	if err != nil {
		return GuildComparison{}, err
	}
	totals := buildScoredTotals(rows)
	perCapita := map[string]domain.Totals{}
	for guild, total := range totals {
		if total.MemberCount > 0 {
			div := int64(total.MemberCount)
			perCapita[guild] = domain.Totals{
				MemberCount: total.MemberCount,
				Kills: total.Kills / div, Assists: total.Assists / div, PlayerDamage: total.PlayerDamage / div,
				BuildingDamage: total.BuildingDamage / div, Healing: total.Healing / div, DamageTaken: total.DamageTaken / div,
				Deaths: total.Deaths / div, Control: total.Control / div, CompositeAvg: total.CompositeAvg,
			}
		}
	}
	careers := careerComparison(rows, battle.HomeGuild, battle.OpponentGuild)
	return GuildComparison{Battle: battle.BattleSummary, Totals: totals, PerCapita: perCapita, Careers: careers, Insights: compareTotals(totals, battle.HomeGuild, battle.OpponentGuild)}, nil
}

func (s *Store) SquadComparison(ctx context.Context, battleID int64, side string) (SquadComparison, error) {
	battle, err := s.BattleDetail(ctx, battleID)
	if err != nil {
		return SquadComparison{}, err
	}
	rows, err := s.BattleRows(ctx, battleID)
	if err != nil {
		return SquadComparison{}, err
	}
	if side != "" {
		rows = filterRows(rows, side, "", "", "")
	}
	group := map[string][]domain.ScoredStat{}
	for _, row := range rows {
		key := row.GuildName + "|" + row.TeamLeader
		group[key] = append(group[key], row)
	}
	var squads []map[string]interface{}
	for key, items := range group {
		parts := strings.SplitN(key, "|", 2)
		total := buildScoredTotals(items)[parts[0]]
		squads = append(squads, map[string]interface{}{
			"guild_name": parts[0], "team_leader": parts[1], "member_count": len(items),
			"totals": total, "top3": buildTeamTop3(items)[parts[0]+" / "+parts[1]],
		})
	}
	sort.Slice(squads, func(i, j int) bool {
		return fmt.Sprint(squads[i]["guild_name"], squads[i]["team_leader"]) < fmt.Sprint(squads[j]["guild_name"], squads[j]["team_leader"])
	})
	return SquadComparison{Battle: battle.BattleSummary, Squads: squads}, nil
}

func (s *Store) HistoryRankings(ctx context.Context, guild, career, search string, minMatches int) ([]HistoryRankingItem, error) {
	if minMatches <= 0 {
		minMatches = 3
	}
	rows, err := s.pool.Query(ctx, `
		SELECT COALESCE(bps.player_id, 0), bps.player_name_snapshot, bps.guild_name_snapshot, bps.career,
		       COUNT(*), AVG(bps.composite_score), SUM(bps.composite_score), MAX(bps.composite_score),
		       (ARRAY_AGG(bps.composite_score ORDER BY b.battle_at DESC))[1],
		       MAX(b.battle_at)
		FROM battle_player_stat bps
		JOIN battle b ON b.id=bps.battle_id
		WHERE ($1='' OR bps.guild_name_snapshot=$1)
		  AND ($2='' OR bps.career=$2)
		  AND ($3='' OR bps.player_name_snapshot ILIKE '%' || $3 || '%')
		GROUP BY COALESCE(bps.player_id, 0), bps.player_name_snapshot, bps.guild_name_snapshot, bps.career
		HAVING COUNT(*) >= $4
		ORDER BY AVG(bps.composite_score) DESC, bps.player_name_snapshot ASC
		LIMIT 200
	`, guild, career, search, minMatches)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []HistoryRankingItem
	for rows.Next() {
		var item HistoryRankingItem
		if err = rows.Scan(&item.PlayerID, &item.PlayerName, &item.GuildName, &item.Career, &item.MatchCount, &item.AverageScore,
			&item.CumulativeScore, &item.BestScore, &item.LatestScore, &item.LatestBattleAt); err != nil {
			return nil, err
		}
		trend, _ := s.scoreTrend(ctx, item.PlayerID, item.Career)
		item.RecentTrend = trend
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) PlayerTrend(ctx context.Context, guildName, playerName, career string) ([]map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT b.battle_at, bps.composite_score
		FROM battle_player_stat bps
		JOIN battle b ON b.id=bps.battle_id
		WHERE bps.guild_name_snapshot=$1 AND bps.player_name_snapshot=$2 AND bps.career=$3
		ORDER BY b.battle_at ASC
		LIMIT 50
	`, guildName, playerName, career)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var at time.Time
		var score float64
		if err = rows.Scan(&at, &score); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{"battle_at": at, "score": score})
	}
	return out, rows.Err()
}

func (s *Store) Settings(ctx context.Context) (map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `SELECT key, value_json FROM app_setting ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]interface{}{}
	for rows.Next() {
		var key string
		var raw []byte
		if err = rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		var value interface{}
		if err = json.Unmarshal(raw, &value); err != nil {
			value = string(raw)
		}
		out[key] = value
	}
	return out, rows.Err()
}

func (s *Store) PutSettings(ctx context.Context, settings map[string]interface{}) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for key, value := range settings {
		raw, _ := json.Marshal(value)
		if _, err = tx.Exec(ctx, `
			INSERT INTO app_setting(key, value_json, updated_at)
			VALUES($1,$2::jsonb,now())
			ON CONFLICT(key) DO UPDATE SET value_json=excluded.value_json, updated_at=now()
		`, key, string(raw)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) ScoringRules(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `SELECT version, name, status, config_json, created_at, published_at, is_active FROM scoring_rule ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var version, name, status string
		var configRaw []byte
		var createdAt time.Time
		var publishedAt *time.Time
		var active bool
		if err = rows.Scan(&version, &name, &status, &configRaw, &createdAt, &publishedAt, &active); err != nil {
			return nil, err
		}
		var config interface{}
		_ = json.Unmarshal(configRaw, &config)
		out = append(out, map[string]interface{}{"version": version, "name": name, "status": status, "config": config, "created_at": createdAt, "published_at": publishedAt, "is_active": active})
	}
	return out, rows.Err()
}

func (s *Store) CreateScoringRule(ctx context.Context, name string, configValue interface{}) (string, error) {
	if name == "" {
		name = "自定义评分规则"
	}
	version := "rule-" + time.Now().Format("20060102150405")
	raw, _ := json.Marshal(configValue)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO scoring_rule(version, name, status, config_json)
		VALUES($1,$2,'draft',$3::jsonb)
	`, version, name, string(raw))
	return version, err
}

func (s *Store) PublishScoringRule(ctx context.Context, version string, userID int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE scoring_rule SET is_active=FALSE WHERE is_active=TRUE`); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE scoring_rule SET status='published', published_at=now(), published_by=$2, is_active=TRUE WHERE version=$1`, version, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return tx.Commit(ctx)
}

func (s *Store) ScoringRanges(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `SELECT version, name, config_json, source_method, sample_summary_json, created_at, is_active, is_frozen FROM scoring_range_version ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var version, name, source string
		var configRaw, sampleRaw []byte
		var createdAt time.Time
		var active, frozen bool
		if err = rows.Scan(&version, &name, &configRaw, &source, &sampleRaw, &createdAt, &active, &frozen); err != nil {
			return nil, err
		}
		var cfg, sample interface{}
		_ = json.Unmarshal(configRaw, &cfg)
		_ = json.Unmarshal(sampleRaw, &sample)
		out = append(out, map[string]interface{}{"version": version, "name": name, "config": cfg, "source_method": source, "sample_summary": sample, "created_at": createdAt, "is_active": active, "is_frozen": frozen})
	}
	return out, rows.Err()
}

func (s *Store) PublishScoringRange(ctx context.Context, name string, configValue interface{}, userID int64) (string, error) {
	if name == "" {
		name = "职业范围版本"
	}
	version := "range-" + time.Now().Format("20060102150405")
	raw, _ := json.Marshal(configValue)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE scoring_range_version SET is_active=FALSE WHERE is_active=TRUE`); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO scoring_range_version(version, name, config_json, source_method, sample_summary_json, created_by, is_active, is_frozen)
		VALUES($1,$2,$3::jsonb,'admin_confirmed','{}'::jsonb,$4,TRUE,TRUE)
	`, version, name, string(raw), userID); err != nil {
		return "", err
	}
	return version, tx.Commit(ctx)
}

func (s *Store) ReanalyzeBattle(ctx context.Context, battleID int64) error {
	detail, err := s.BattleDetail(ctx, battleID)
	if err != nil {
		return err
	}
	rawRows, err := s.rawRows(ctx, battleID)
	if err != nil {
		return err
	}
	analysis := AnalyzeRows(rawRows, detail.HomeGuild)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, row := range analysis.Rows {
		if _, err = tx.Exec(ctx, `
			UPDATE battle_player_stat
			SET kda_ratio=$1, player_damage_share=$2, building_damage_share=$3,
			    player_damage_conversion_rate=$4, building_damage_conversion_rate=$5,
			    participation_rate=$6, composite_score=$7, guild_rank=$8,
			    career_rank=$9, team_rank=$10, six_dimension_json=$11::jsonb,
			    score_detail_json=$12::jsonb
			WHERE battle_id=$13 AND guild_name_snapshot=$14 AND player_name_snapshot=$15
		`, row.KDARatio, row.PlayerDamageShare, row.BuildingDamageShare, row.PlayerDamageConversionRate,
			row.BuildingDamageConversionRate, row.ParticipationRate, row.CompositeScore, row.GuildRank,
			row.CareerRank, row.TeamRank, mustJSON(row.SixDimensions), mustJSON(row.ScoreDetail), battleID, row.GuildName, row.PlayerName); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) SaveAvatar(ctx context.Context, kind string, id string, filename string, data []byte, userID int64) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
		return "", errors.New("avatar must be png, jpg, jpeg or webp")
	}
	dir := filepath.Join(s.cfg.UploadDir, "avatars", kind)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	sum := HashOpaque(string(data))
	assetPath := filepath.Join(dir, id+"-"+sum[:12]+ext)
	if err := os.WriteFile(assetPath, data, 0644); err != nil {
		return "", err
	}
	publicPath := "/uploads" + strings.TrimPrefix(strings.ReplaceAll(assetPath, "\\", "/"), strings.ReplaceAll(s.cfg.UploadDir, "\\", "/"))
	if kind == "players" {
		playerID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return "", errors.New("invalid player id")
		}
		if _, err := s.pool.Exec(ctx, `
			INSERT INTO player_avatar(player_id, source, asset_path, content_sha256, updated_by)
			VALUES($1,'uploaded',$2,$3,$4)
			ON CONFLICT(player_id) DO UPDATE SET source='uploaded', asset_path=excluded.asset_path, content_sha256=excluded.content_sha256, updated_by=excluded.updated_by, updated_at=now()
		`, playerID, publicPath, sum, userID); err != nil {
			return "", err
		}
	} else {
		if _, err := s.pool.Exec(ctx, `
			INSERT INTO career_avatar(career, asset_path, content_sha256, updated_by)
			VALUES($1,$2,$3,$4)
			ON CONFLICT(career) DO UPDATE SET asset_path=excluded.asset_path, content_sha256=excluded.content_sha256, updated_by=excluded.updated_by, updated_at=now()
		`, id, publicPath, sum, userID); err != nil {
			return "", err
		}
	}
	return publicPath, nil
}

func (s *Store) DeleteAvatar(ctx context.Context, kind string, id string) error {
	if kind == "players" {
		playerID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return errors.New("invalid player id")
		}
		_, err = s.pool.Exec(ctx, `DELETE FROM player_avatar WHERE player_id=$1`, playerID)
		return err
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM career_avatar WHERE career=$1`, id)
	return err
}

func (s *Store) Backup(ctx context.Context) (string, error) {
	if err := os.MkdirAll(s.cfg.BackupDir, 0755); err != nil {
		return "", err
	}
	payload := map[string]interface{}{"created_at": time.Now(), "type": "logical_metadata_backup"}
	var counts []map[string]interface{}
	for _, table := range []string{"battle", "guild", "player", "battle_player_stat", "scoring_rule", "scoring_range_version"} {
		var count int64
		if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			return "", err
		}
		counts = append(counts, map[string]interface{}{"table": table, "count": count})
	}
	payload["counts"] = counts
	raw, _ := json.MarshalIndent(payload, "", "  ")
	name := filepath.Join(s.cfg.BackupDir, "backup-"+time.Now().Format("20060102-150405")+".json")
	if err := os.WriteFile(name, raw, 0644); err != nil {
		return "", err
	}
	return name, nil
}

func (s *Store) rawRows(ctx context.Context, battleID int64) ([]domain.RawStat, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT guild_name_snapshot, player_name_snapshot, level, career, team_leader_snapshot,
		       kills, assists, logistics, player_damage, building_damage, healing, damage_taken,
		       deaths, qingdeng, revive, control
		FROM battle_player_stat
		WHERE battle_id=$1
	`, battleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RawStat
	for rows.Next() {
		var item domain.RawStat
		if err = rows.Scan(&item.GuildName, &item.PlayerName, &item.Level, &item.Career, &item.TeamLeader,
			&item.Kills, &item.Assists, &item.Logistics, &item.PlayerDamage, &item.BuildingDamage, &item.Healing,
			&item.DamageTaken, &item.Deaths, &item.Qingdeng, &item.Revive, &item.Control); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func upsertGuild(ctx context.Context, tx pgx.Tx, name string) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO guild(name) VALUES($1)
		ON CONFLICT(name) DO UPDATE SET name=excluded.name
		RETURNING id
	`, name).Scan(&id)
	return id, err
}

func upsertPlayer(ctx context.Context, tx pgx.Tx, guildID int64, name string) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO player(guild_id, canonical_name) VALUES($1,$2)
		ON CONFLICT(guild_id, canonical_name) DO UPDATE SET canonical_name=excluded.canonical_name
		RETURNING id
	`, guildID, name).Scan(&id)
	return id, err
}

func mustJSON(value interface{}) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func nullableInt(value int) interface{} {
	if value == 0 {
		return nil
	}
	return value
}

func filterRows(rows []domain.ScoredStat, side, career, team, search string) []domain.ScoredStat {
	side = strings.TrimSpace(side)
	career = strings.TrimSpace(career)
	team = strings.TrimSpace(team)
	search = strings.TrimSpace(strings.ToLower(search))
	var out []domain.ScoredStat
	for _, row := range rows {
		if side != "" && side != "all" && row.Side != side {
			continue
		}
		if career != "" && row.Career != career {
			continue
		}
		if team != "" && row.TeamLeader != team {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(row.PlayerName), search) {
			continue
		}
		out = append(out, row)
	}
	return out
}

func paginate(rows []domain.ScoredStat, page, size int) RankingResponse {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}
	if size > 200 {
		size = 200
	}
	total := len(rows)
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return RankingResponse{Items: rows[start:end], Total: total, Page: page, Size: size}
}

func filterTop(rows []domain.ScoredStat, side string, n int) []domain.ScoredStat {
	filtered := filterRows(rows, side, "", "", "")
	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].CompositeScore > filtered[j].CompositeScore
	})
	if len(filtered) > n {
		return filtered[:n]
	}
	return filtered
}

func compareTotals(totals map[string]domain.Totals, home, opponent string) []Insight {
	metrics := []struct {
		key   string
		label string
		home  func(domain.Totals) float64
	}{
		{"kills", "击败", func(t domain.Totals) float64 { return float64(t.Kills) }},
		{"assists", "助攻", func(t domain.Totals) float64 { return float64(t.Assists) }},
		{"player_damage", "对玩家伤害", func(t domain.Totals) float64 { return float64(t.PlayerDamage) }},
		{"building_damage", "对建筑伤害", func(t domain.Totals) float64 { return float64(t.BuildingDamage) }},
		{"healing", "治疗值", func(t domain.Totals) float64 { return float64(t.Healing) }},
		{"control", "控制", func(t domain.Totals) float64 { return float64(t.Control) }},
		{"composite_avg", "平均综合分", func(t domain.Totals) float64 { return t.CompositeAvg }},
	}
	var insights []Insight
	for _, metric := range metrics {
		h := metric.home(totals[home])
		o := metric.home(totals[opponent])
		delta := 0.0
		if o != 0 {
			delta = (h - o) / abs(o)
		} else if h > 0 {
			delta = 1
		}
		label := "持平"
		if delta >= 0.05 {
			label = "优势"
		} else if delta <= -0.05 {
			label = "不足"
		}
		insights = append(insights, Insight{Metric: metric.key, Label: metric.label, Home: h, Opponent: o, Delta: round2(delta * 100), Message: fmt.Sprintf("%s：本帮 %.0f，对手 %.0f，判断为%s。", metric.label, h, o, label)})
	}
	sort.Slice(insights, func(i, j int) bool {
		return abs(insights[i].Delta) > abs(insights[j].Delta)
	})
	if len(insights) > 6 {
		return insights[:6]
	}
	return insights
}

func careerComparison(rows []domain.ScoredStat, home, opponent string) []map[string]interface{} {
	group := map[string]map[string][]domain.ScoredStat{}
	for _, row := range rows {
		if group[row.Career] == nil {
			group[row.Career] = map[string][]domain.ScoredStat{}
		}
		group[row.Career][row.GuildName] = append(group[row.Career][row.GuildName], row)
	}
	var out []map[string]interface{}
	for career, guilds := range group {
		item := map[string]interface{}{"career": career, "home_count": len(guilds[home]), "opponent_count": len(guilds[opponent])}
		item["home_avg_score"] = avgScore(guilds[home])
		item["opponent_avg_score"] = avgScore(guilds[opponent])
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]["career"]) < fmt.Sprint(out[j]["career"]) })
	return out
}

func sameCareerSummary(rows []domain.ScoredStat, player domain.ScoredStat) map[string]interface{} {
	var same []domain.ScoredStat
	var home []domain.ScoredStat
	var opponent []domain.ScoredStat
	for _, row := range rows {
		if row.Career != player.Career {
			continue
		}
		same = append(same, row)
		if row.Side == "home" {
			home = append(home, row)
		} else {
			opponent = append(opponent, row)
		}
	}
	percentile := 0.0
	for _, row := range same {
		if row.CompositeScore <= player.CompositeScore {
			percentile++
		}
	}
	if len(same) > 0 {
		percentile = percentile * 100 / float64(len(same))
	}
	return map[string]interface{}{"sample_size": len(same), "home_average": avgScore(home), "opponent_average": avgScore(opponent), "percentile": round2(percentile)}
}

func avgScore(rows []domain.ScoredStat) float64 {
	if len(rows) == 0 {
		return 0
	}
	sum := 0.0
	for _, row := range rows {
		sum += row.CompositeScore
	}
	return round2(sum / float64(len(rows)))
}

func (s *Store) scoreTrend(ctx context.Context, playerID int64, career string) ([]float64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT composite_score
		FROM battle_player_stat bps
		JOIN battle b ON b.id=bps.battle_id
		WHERE bps.player_id=$1 AND bps.career=$2
		ORDER BY b.battle_at DESC
		LIMIT 5
	`, playerID, career)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []float64
	for rows.Next() {
		var score float64
		if err = rows.Scan(&score); err != nil {
			return nil, err
		}
		values = append(values, score)
	}
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	return values, rows.Err()
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
