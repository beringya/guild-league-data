package services

import (
	"math"
	"sort"

	"nsh-guild-analytics/backend/internal/domain"
)

type metricRange struct {
	Min       float64
	Max       float64
	Source    string
	SampleSize int
	Method    string
}

func AnalyzeRows(rows []domain.RawStat, homeGuild string) domain.BattleAnalysis {
	opponentGuild := ""
	for _, row := range rows {
		if row.GuildName != homeGuild {
			opponentGuild = row.GuildName
			break
		}
	}
	ranges := buildRanges(rows)
	totals := buildTotals(rows)
	profiles := domain.CareerProfiles()
	scored := make([]domain.ScoredStat, 0, len(rows))
	for _, row := range rows {
		side := "opponent"
		if row.GuildName == homeGuild {
			side = "home"
		}
		stat := domain.ScoredStat{RawStat: row, Side: side}
		stat.KDARatio = float64(row.Kills+row.Assists) / math.Max(float64(row.Deaths), 1)
		if guildTotal := totals[row.GuildName]; guildTotal.Kills > 0 {
			stat.ParticipationRate = float64(row.Kills+row.Assists) / float64(guildTotal.Kills)
		}
		if guildTotal := totals[row.GuildName]; guildTotal.PlayerDamage > 0 {
			stat.PlayerDamageShare = float64(row.PlayerDamage) / float64(guildTotal.PlayerDamage)
		}
		if guildTotal := totals[row.GuildName]; guildTotal.BuildingDamage > 0 {
			stat.BuildingDamageShare = float64(row.BuildingDamage) / float64(guildTotal.BuildingDamage)
		}
		profile := profiles[row.Career]
		rangeMap := ranges[row.Career]
		var composite float64
		var weightSum float64
		for _, dim := range profile.Dimensions {
			value := metricValue(stat, dim.Metric)
			r := rangeMap[dim.Metric]
			score := normalize(value, r.Min, r.Max, dim.Direction)
			contribution := score * dim.RankingWeight
			if dim.Enabled && dim.RankingWeight > 0 {
				composite += contribution
				weightSum += dim.RankingWeight
			}
			stat.SixDimensions = append(stat.SixDimensions, domain.DimensionScore{
				Slot: dim.Slot, Metric: dim.Metric, Label: dim.Label, Value: value, Score: score,
				Weight: dim.RankingWeight, Contribution: contribution, RangeMin: r.Min, RangeMax: r.Max,
				RangeSource: r.Source, Percentile: percentile(rows, row.Career, dim.Metric, value),
			})
			if dim.Metric == domain.MetricPlayerDamage {
				stat.PlayerDamageConversionRate = score / 100
			}
			if dim.Metric == domain.MetricBuildingDamage {
				stat.BuildingDamageConversionRate = score / 100
			}
		}
		if weightSum > 0 {
			stat.CompositeScore = round2(composite / weightSum)
		}
		stat.ScoreDetail = map[string]interface{}{
			"career": row.Career, "rule_version": "default-v1.5-final", "range_version": "import-snapshot",
			"kda_formula": "(击败 + 助攻) / max(重伤, 1)",
			"participation_formula": "(击败 + 助攻) / 所在帮会总击败",
			"dimensions": stat.SixDimensions,
		}
		scored = append(scored, stat)
	}
	assignRanks(scored)
	totals = buildScoredTotals(scored)
	return domain.BattleAnalysis{
		HomeGuild: homeGuild, OpponentGuild: opponentGuild,
		ScoringRule: "default-v1.5-final", ScoringRange: "import-snapshot",
		Rows: scored, GuildTotals: totals, TeamTop3: buildTeamTop3(scored), RangeSuggestions: SuggestRanges(rows),
	}
}

func SuggestRanges(rows []domain.RawStat) []domain.RangeSuggestion {
	profiles := domain.CareerProfiles()
	seen := map[string]bool{}
	var suggestions []domain.RangeSuggestion
	for _, row := range rows {
		profile, ok := profiles[row.Career]
		if !ok {
			continue
		}
		for _, dim := range profile.Dimensions {
			key := row.Career + "|" + dim.Metric
			if seen[key] {
				continue
			}
			seen[key] = true
			values := metricValuesForCareer(rows, row.Career, dim.Metric)
			if len(values) == 0 {
				continue
			}
			sort.Float64s(values)
			minValue, maxValue := suggestBounds(values)
			method := "min_max_margin"
			warning := ""
			if len(values) >= 20 {
				method = "p05_p95"
			} else if len(values) < 3 {
				method = "very_small_sample"
				warning = "同职业样本少于 3 人，建议人工复核范围。"
			}
			suggestions = append(suggestions, domain.RangeSuggestion{
				Career: row.Career, Metric: dim.Metric, SampleSize: len(values), Method: method,
				SuggestedMin: round2(minValue), SuggestedMax: round2(maxValue), Warning: warning,
			})
		}
	}
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Career == suggestions[j].Career {
			return suggestions[i].Metric < suggestions[j].Metric
		}
		return suggestions[i].Career < suggestions[j].Career
	})
	return suggestions
}

func buildRanges(rows []domain.RawStat) map[string]map[string]metricRange {
	out := map[string]map[string]metricRange{}
	profiles := domain.CareerProfiles()
	for career, profile := range profiles {
		out[career] = map[string]metricRange{}
		for _, dim := range profile.Dimensions {
			values := metricValuesForCareer(rows, career, dim.Metric)
			minValue, maxValue := suggestBounds(values)
			if minValue == maxValue {
				maxValue = minValue + 1
			}
			method := "min_max_margin"
			if len(values) >= 20 {
				method = "p05_p95"
			} else if len(values) < 3 {
				method = "very_small_sample"
			}
			out[career][dim.Metric] = metricRange{Min: minValue, Max: maxValue, Source: "import_suggestion", SampleSize: len(values), Method: method}
		}
	}
	return out
}

func metricValuesForCareer(rows []domain.RawStat, career, metric string) []float64 {
	values := []float64{}
	totals := buildTotals(rows)
	for _, row := range rows {
		if row.Career != career {
			continue
		}
		stat := domain.ScoredStat{RawStat: row}
		stat.KDARatio = float64(row.Kills+row.Assists) / math.Max(float64(row.Deaths), 1)
		if guildTotal := totals[row.GuildName]; guildTotal.Kills > 0 {
			stat.ParticipationRate = float64(row.Kills+row.Assists) / float64(guildTotal.Kills)
		}
		values = append(values, metricValue(stat, metric))
	}
	return values
}

func suggestBounds(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 1
	}
	sort.Float64s(values)
	if len(values) >= 20 {
		return values[int(math.Floor(float64(len(values)-1)*0.05))], values[int(math.Ceil(float64(len(values)-1)*0.95))]
	}
	minValue, maxValue := values[0], values[len(values)-1]
	margin := math.Max((maxValue-minValue)*0.1, 1)
	return math.Max(minValue-margin, 0), maxValue + margin
}

func normalize(value, minValue, maxValue float64, direction string) float64 {
	if maxValue <= minValue {
		return 100
	}
	var ratio float64
	if direction == "lower" {
		ratio = (maxValue - value) / (maxValue - minValue)
	} else {
		ratio = (value - minValue) / (maxValue - minValue)
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	return round2(ratio * 100)
}

func metricValue(stat domain.ScoredStat, metric string) float64 {
	switch metric {
	case domain.MetricKills:
		return float64(stat.Kills)
	case domain.MetricAssists:
		return float64(stat.Assists)
	case domain.MetricPlayerDamage:
		return float64(stat.PlayerDamage)
	case domain.MetricBuildingDamage:
		return float64(stat.BuildingDamage)
	case domain.MetricHealing:
		return float64(stat.Healing)
	case domain.MetricDamageTaken:
		return float64(stat.DamageTaken)
	case domain.MetricQingdeng:
		return float64(stat.Qingdeng)
	case domain.MetricRevive:
		return float64(stat.Revive)
	case domain.MetricControl:
		return float64(stat.Control)
	case domain.MetricKDARatio:
		return stat.KDARatio
	case domain.MetricParticipationRate:
		return stat.ParticipationRate
	default:
		return 0
	}
}

func percentile(rows []domain.RawStat, career, metric string, value float64) float64 {
	values := metricValuesForCareer(rows, career, metric)
	if len(values) == 0 {
		return 0
	}
	count := 0
	for _, candidate := range values {
		if candidate <= value {
			count++
		}
	}
	return round2(float64(count) * 100 / float64(len(values)))
}

func buildTotals(rows []domain.RawStat) map[string]domain.Totals {
	out := map[string]domain.Totals{}
	for _, row := range rows {
		total := out[row.GuildName]
		total.MemberCount++
		total.Kills += int64(row.Kills)
		total.Assists += int64(row.Assists)
		total.PlayerDamage += row.PlayerDamage
		total.BuildingDamage += row.BuildingDamage
		total.Healing += row.Healing
		total.DamageTaken += row.DamageTaken
		total.Deaths += int64(row.Deaths)
		total.Control += int64(row.Control)
		out[row.GuildName] = total
	}
	return out
}

func buildScoredTotals(rows []domain.ScoredStat) map[string]domain.Totals {
	out := map[string]domain.Totals{}
	scoreSum := map[string]float64{}
	for _, row := range rows {
		total := out[row.GuildName]
		total.MemberCount++
		total.Kills += int64(row.Kills)
		total.Assists += int64(row.Assists)
		total.PlayerDamage += row.PlayerDamage
		total.BuildingDamage += row.BuildingDamage
		total.Healing += row.Healing
		total.DamageTaken += row.DamageTaken
		total.Deaths += int64(row.Deaths)
		total.Control += int64(row.Control)
		scoreSum[row.GuildName] += row.CompositeScore
		out[row.GuildName] = total
	}
	for guild, total := range out {
		if total.MemberCount > 0 {
			total.CompositeAvg = round2(scoreSum[guild] / float64(total.MemberCount))
		}
		out[guild] = total
	}
	return out
}

func assignRanks(rows []domain.ScoredStat) {
	byGuild := map[string][]int{}
	byCareer := map[string][]int{}
	byTeam := map[string][]int{}
	for i, row := range rows {
		byGuild[row.GuildName] = append(byGuild[row.GuildName], i)
		byCareer[row.GuildName+"|"+row.Career] = append(byCareer[row.GuildName+"|"+row.Career], i)
		byTeam[row.GuildName+"|"+row.TeamLeader] = append(byTeam[row.GuildName+"|"+row.TeamLeader], i)
	}
	rank := func(indexes []int, set func(int, int)) {
		sort.SliceStable(indexes, func(i, j int) bool {
			a, b := rows[indexes[i]], rows[indexes[j]]
			if a.CompositeScore != b.CompositeScore {
				return a.CompositeScore > b.CompositeScore
			}
			if a.ParticipationRate != b.ParticipationRate {
				return a.ParticipationRate > b.ParticipationRate
			}
			if a.KDARatio != b.KDARatio {
				return a.KDARatio > b.KDARatio
			}
			return a.PlayerName < b.PlayerName
		})
		lastScore := math.NaN()
		currentRank := 0
		for i, idx := range indexes {
			if i == 0 || rows[idx].CompositeScore != lastScore {
				currentRank = i + 1
				lastScore = rows[idx].CompositeScore
			}
			set(idx, currentRank)
		}
	}
	for _, indexes := range byGuild {
		rank(indexes, func(idx, value int) { rows[idx].GuildRank = value })
	}
	for _, indexes := range byCareer {
		rank(indexes, func(idx, value int) { rows[idx].CareerRank = value })
	}
	for _, indexes := range byTeam {
		rank(indexes, func(idx, value int) { rows[idx].TeamRank = value })
	}
}

func buildTeamTop3(rows []domain.ScoredStat) map[string][]domain.ScoredStat {
	teams := map[string][]domain.ScoredStat{}
	for _, row := range rows {
		key := row.GuildName + " / " + row.TeamLeader
		teams[key] = append(teams[key], row)
	}
	for key, items := range teams {
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].CompositeScore != items[j].CompositeScore {
				return items[i].CompositeScore > items[j].CompositeScore
			}
			return items[i].PlayerName < items[j].PlayerName
		})
		if len(items) > 3 {
			items = items[:3]
		}
		teams[key] = items
	}
	return teams
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
