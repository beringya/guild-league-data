package domain

func CareerProfiles() map[string]CareerProfile {
	damage := func(career string) CareerProfile {
		return CareerProfile{
			Career: career, Status: "confirmed", RankingEnabled: true,
			Dimensions: []Dimension{
				{Slot: 1, Metric: MetricPlayerDamage, Label: "对玩家伤害", Enabled: true, Direction: "higher", RankingWeight: 0.5},
				{Slot: 2, Metric: MetricBuildingDamage, Label: "对建筑伤害", Enabled: true, Direction: "higher", RankingWeight: 0.5},
				{Slot: 3, Metric: MetricKills, Label: "击败", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 4, Metric: MetricKDARatio, Label: "KDA", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 5, Metric: MetricParticipationRate, Label: "参团率", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 6, Metric: MetricControl, Label: "控制", Enabled: true, Direction: "higher", RankingWeight: 0},
			},
		}
	}

	return map[string]CareerProfile{
		"素问": {
			Career: "素问", Status: "confirmed", RankingEnabled: true,
			Dimensions: []Dimension{
				{Slot: 1, Metric: MetricHealing, Label: "治疗值", Enabled: true, Direction: "higher", RankingWeight: 0.55},
				{Slot: 2, Metric: MetricDamageTaken, Label: "承受伤害", Enabled: true, Direction: "higher", RankingWeight: 0.25},
				{Slot: 3, Metric: MetricRevive, Label: "化羽", Enabled: true, Direction: "higher", RankingWeight: 0.20},
				{Slot: 4, Metric: MetricAssists, Label: "助攻", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 5, Metric: MetricParticipationRate, Label: "参团率", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 6, Metric: MetricKDARatio, Label: "KDA", Enabled: true, Direction: "higher", RankingWeight: 0},
			},
		},
		"铁衣": {
			Career: "铁衣", Status: "confirmed", RankingEnabled: true,
			Dimensions: []Dimension{
				{Slot: 1, Metric: MetricControl, Label: "控制", Enabled: true, Direction: "higher", RankingWeight: 0.60},
				{Slot: 2, Metric: MetricDamageTaken, Label: "承受伤害", Enabled: true, Direction: "higher", RankingWeight: 0.40},
				{Slot: 3, Metric: MetricKDARatio, Label: "KDA", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 4, Metric: MetricParticipationRate, Label: "参团率", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 5, Metric: MetricPlayerDamage, Label: "对玩家伤害", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 6, Metric: MetricBuildingDamage, Label: "对建筑伤害", Enabled: true, Direction: "higher", RankingWeight: 0},
			},
		},
		"神相": damage("神相"),
		"血河": damage("血河"),
		"沧澜": damage("沧澜"),
		"玄机": damage("玄机"),
		"云瑶": damage("云瑶"),
		"潮光": damage("潮光"),
		"荒羽": damage("荒羽"),
		"龙吟": damage("龙吟"),
		"碎梦": {
			Career: "碎梦", Status: "confirmed", RankingEnabled: true,
			Dimensions: []Dimension{
				{Slot: 1, Metric: MetricKills, Label: "击败", Enabled: true, Direction: "higher", RankingWeight: 0.50},
				{Slot: 2, Metric: MetricPlayerDamage, Label: "对玩家伤害", Enabled: true, Direction: "higher", RankingWeight: 0.30},
				{Slot: 3, Metric: MetricBuildingDamage, Label: "对建筑伤害", Enabled: true, Direction: "higher", RankingWeight: 0.20},
				{Slot: 4, Metric: MetricKDARatio, Label: "KDA", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 5, Metric: MetricAssists, Label: "助攻", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 6, Metric: MetricParticipationRate, Label: "参团率", Enabled: true, Direction: "higher", RankingWeight: 0},
			},
		},
		"九灵": {
			Career: "九灵", Status: "confirmed", RankingEnabled: true,
			Dimensions: []Dimension{
				{Slot: 1, Metric: MetricQingdeng, Label: "青灯焚骨", Enabled: true, Direction: "higher", RankingWeight: 0.60},
				{Slot: 2, Metric: MetricPlayerDamage, Label: "对玩家伤害", Enabled: true, Direction: "higher", RankingWeight: 0.20},
				{Slot: 3, Metric: MetricBuildingDamage, Label: "对建筑伤害", Enabled: true, Direction: "higher", RankingWeight: 0.20},
				{Slot: 4, Metric: MetricKDARatio, Label: "KDA", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 5, Metric: MetricParticipationRate, Label: "参团率", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 6, Metric: MetricControl, Label: "控制", Enabled: true, Direction: "higher", RankingWeight: 0},
			},
		},
		"鸿音": {
			Career: "鸿音", Status: "confirmed", RankingEnabled: true,
			Dimensions: []Dimension{
				{Slot: 1, Metric: MetricControl, Label: "控制", Enabled: true, Direction: "higher", RankingWeight: 0.55},
				{Slot: 2, Metric: MetricHealing, Label: "治疗值", Enabled: true, Direction: "higher", RankingWeight: 0.45},
				{Slot: 3, Metric: MetricRevive, Label: "化羽", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 4, Metric: MetricKDARatio, Label: "KDA", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 5, Metric: MetricParticipationRate, Label: "参团率", Enabled: true, Direction: "higher", RankingWeight: 0},
				{Slot: 6, Metric: MetricDamageTaken, Label: "承受伤害", Enabled: true, Direction: "higher", RankingWeight: 0},
			},
		},
	}
}

func CareerNames() []string {
	profiles := CareerProfiles()
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	return names
}
