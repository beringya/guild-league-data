package domain

import "time"

const (
	MetricKills             = "kills"
	MetricAssists           = "assists"
	MetricPlayerDamage      = "player_damage"
	MetricBuildingDamage    = "building_damage"
	MetricHealing           = "healing"
	MetricDamageTaken       = "damage_taken"
	MetricQingdeng          = "qingdeng"
	MetricRevive            = "revive"
	MetricControl           = "control"
	MetricKDARatio          = "kda_ratio"
	MetricParticipationRate = "participation_rate"
)

type RawStat struct {
	RowNumber      int       `json:"row_number"`
	GuildName      string    `json:"guild_name"`
	PlayerName     string    `json:"player_name"`
	Level          int       `json:"level"`
	Career         string    `json:"career"`
	TeamLeader     string    `json:"team_leader"`
	Kills          int       `json:"kills"`
	Assists        int       `json:"assists"`
	Logistics      int64     `json:"logistics"`
	PlayerDamage   int64     `json:"player_damage"`
	BuildingDamage int64     `json:"building_damage"`
	Healing        int64     `json:"healing"`
	DamageTaken    int64     `json:"damage_taken"`
	Deaths         int       `json:"deaths"`
	Qingdeng       int       `json:"qingdeng"`
	Revive         int       `json:"revive"`
	Control        int       `json:"control"`
	BattleAt       time.Time `json:"battle_at,omitempty"`
}

type ImportMessage struct {
	Level     string `json:"level"`
	Code      string `json:"code"`
	RowNumber int    `json:"row_number,omitempty"`
	Message   string `json:"message"`
}

type GuildPreview struct {
	Name        string         `json:"name"`
	MemberCount int           `json:"member_count"`
	Careers     map[string]int `json:"careers"`
	Teams       map[string]int `json:"teams"`
}

type ImportPreview struct {
	Token            string           `json:"token,omitempty"`
	SourceFilename   string           `json:"source_filename"`
	SourceSHA256     string           `json:"source_sha256"`
	OriginalRowCount  int              `json:"original_row_count"`
	ValidRowCount     int              `json:"valid_row_count"`
	InferredBattleAt  *time.Time       `json:"inferred_battle_at,omitempty"`
	Guilds            []GuildPreview   `json:"guilds"`
	Rows              []RawStat        `json:"rows,omitempty"`
	PreviewRows       []RawStat        `json:"preview_rows"`
	Warnings          []ImportMessage  `json:"warnings"`
	Errors            []ImportMessage  `json:"errors"`
	UnknownCareers    []string         `json:"unknown_careers"`
	RangeSuggestions  []RangeSuggestion `json:"range_suggestions,omitempty"`
}

type Dimension struct {
	Slot          int     `json:"slot"`
	Metric        string  `json:"metric"`
	Label         string  `json:"label"`
	Enabled       bool    `json:"enabled"`
	Direction     string  `json:"direction"`
	RankingWeight float64 `json:"ranking_weight"`
}

type CareerProfile struct {
	Career         string      `json:"career"`
	Status         string      `json:"status"`
	RankingEnabled bool       `json:"ranking_enabled"`
	Dimensions     []Dimension `json:"dimensions"`
}

type MetricRange struct {
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Source    string  `json:"source"`
	SampleSize int     `json:"sample_size,omitempty"`
	Method    string  `json:"method,omitempty"`
}

type RangeSuggestion struct {
	Career       string  `json:"career"`
	Metric       string  `json:"metric"`
	SampleSize   int     `json:"sample_size"`
	Method       string  `json:"method"`
	SuggestedMin float64 `json:"suggested_min"`
	SuggestedMax float64 `json:"suggested_max"`
	Warning      string  `json:"warning,omitempty"`
}

type ScoredStat struct {
	RawStat
	Side                         string                 `json:"side"`
	KDARatio                     float64                `json:"kda_ratio"`
	ParticipationRate            float64                `json:"participation_rate"`
	PlayerDamageShare            float64                `json:"player_damage_share"`
	BuildingDamageShare          float64                `json:"building_damage_share"`
	PlayerDamageConversionRate   float64                `json:"player_damage_conversion_rate"`
	BuildingDamageConversionRate float64                `json:"building_damage_conversion_rate"`
	CompositeScore               float64                `json:"composite_score"`
	GuildRank                    int                    `json:"guild_rank"`
	CareerRank                   int                    `json:"career_rank"`
	TeamRank                     int                    `json:"team_rank"`
	SixDimensions                []DimensionScore       `json:"six_dimensions"`
	ScoreDetail                  map[string]interface{} `json:"score_detail"`
}

type DimensionScore struct {
	Slot          int     `json:"slot"`
	Metric        string  `json:"metric"`
	Label         string  `json:"label"`
	Value         float64 `json:"value"`
	Score         float64 `json:"score"`
	Weight        float64 `json:"weight"`
	Contribution  float64 `json:"contribution"`
	Percentile    float64 `json:"percentile"`
	RangeMin      float64 `json:"range_min"`
	RangeMax      float64 `json:"range_max"`
	RangeSource   string  `json:"range_source"`
}

type BattleAnalysis struct {
	HomeGuild        string            `json:"home_guild"`
	OpponentGuild    string            `json:"opponent_guild"`
	ScoringRule      string            `json:"scoring_rule"`
	ScoringRange     string            `json:"scoring_range"`
	Rows             []ScoredStat      `json:"rows"`
	GuildTotals      map[string]Totals `json:"guild_totals"`
	TeamTop3         map[string][]ScoredStat `json:"team_top3"`
	RangeSuggestions []RangeSuggestion `json:"range_suggestions"`
}

type Totals struct {
	MemberCount     int     `json:"member_count"`
	Kills           int64   `json:"kills"`
	Assists         int64   `json:"assists"`
	PlayerDamage    int64   `json:"player_damage"`
	BuildingDamage  int64   `json:"building_damage"`
	Healing         int64   `json:"healing"`
	DamageTaken     int64   `json:"damage_taken"`
	Deaths          int64   `json:"deaths"`
	Control         int64   `json:"control"`
	CompositeAvg    float64 `json:"composite_avg"`
}
