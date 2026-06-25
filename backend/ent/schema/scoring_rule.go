package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type ScoringRule struct {
	ent.Schema
}

func (ScoringRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("version").Unique().Immutable(),
		field.String("name"),
		field.String("status").Default("draft"),
		field.JSON("config_json", map[string]interface{}{}),
		field.Time("created_at"),
		field.Time("published_at").Optional().Nillable(),
		field.Int64("published_by").Optional().Nillable(),
		field.Bool("is_active").Default(false),
	}
}
