package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type BattlePlayerStat struct {
	ent.Schema
}

func (BattlePlayerStat) Fields() []ent.Field {
	return []ent.Field{
		field.String("guild_name_snapshot"),
		field.String("player_name_snapshot"),
		field.Int("level").Optional(),
		field.String("career"),
		field.String("team_leader_snapshot"),
		field.Int("kills").Default(0),
		field.Int("assists").Default(0),
		field.Int64("logistics").Default(0),
		field.Int64("player_damage").Default(0),
		field.Int64("building_damage").Default(0),
		field.Int64("healing").Default(0),
		field.Int64("damage_taken").Default(0),
		field.Int("deaths").Default(0),
		field.Int("qingdeng").Default(0),
		field.Int("revive").Default(0),
		field.Int("control").Default(0),
		field.Float("kda_ratio").Default(0),
		field.Float("player_damage_share").Optional(),
		field.Float("building_damage_share").Optional(),
		field.Float("player_damage_conversion_rate").Optional(),
		field.Float("building_damage_conversion_rate").Optional(),
		field.Float("participation_rate").Optional(),
		field.Float("composite_score").Optional(),
		field.Int("guild_rank").Optional(),
		field.Int("career_rank").Optional(),
		field.Int("team_rank").Optional(),
		field.JSON("six_dimension_json", []map[string]interface{}{}).Optional(),
		field.JSON("score_detail_json", map[string]interface{}{}).Optional(),
	}
}

func (BattlePlayerStat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("battle", Battle.Type).Ref("stats").Unique().Required(),
		edge.From("battle_guild", BattleGuild.Type).Ref("stats").Unique().Required(),
		edge.From("player", Player.Type).Ref("stats").Unique(),
	}
}

func (BattlePlayerStat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("career").Edges("battle"),
		index.Fields("team_leader_snapshot").Edges("battle"),
		index.Fields("player_name_snapshot").Edges("battle", "battle_guild").Unique(),
	}
}
