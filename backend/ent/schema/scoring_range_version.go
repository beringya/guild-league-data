package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type ScoringRangeVersion struct {
	ent.Schema
}

func (ScoringRangeVersion) Fields() []ent.Field {
	return []ent.Field{
		field.String("version").Unique().Immutable(),
		field.String("name"),
		field.JSON("config_json", map[string]interface{}{}),
		field.String("source_method"),
		field.Int64("source_battle_id").Optional().Nillable(),
		field.JSON("sample_summary_json", map[string]interface{}{}),
		field.Int64("created_by").Optional().Nillable(),
		field.Time("created_at"),
		field.Bool("is_active").Default(false),
		field.Bool("is_frozen").Default(true),
	}
}
