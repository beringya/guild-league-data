package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Battle struct {
	ent.Schema
}

func (Battle) Fields() []ent.Field {
	return []ent.Field{
		field.Time("battle_at"),
		field.String("source_filename"),
		field.String("source_sha256").Unique(),
		field.Int("original_row_count"),
		field.Int("valid_row_count"),
		field.String("scoring_rule_version"),
		field.String("scoring_range_version"),
		field.String("import_status"),
		field.JSON("match_result_json", map[string]interface{}{}).Optional(),
		field.Time("created_at"),
	}
}

func (Battle) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("created_by", User.Type).Ref("battles").Unique(),
		edge.To("battle_guilds", BattleGuild.Type),
		edge.To("stats", BattlePlayerStat.Type),
		edge.To("import_logs", ImportLog.Type),
	}
}
