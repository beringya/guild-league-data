package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type CareerAvatar struct {
	ent.Schema
}

func (CareerAvatar) Fields() []ent.Field {
	return []ent.Field{
		field.String("career").Unique(),
		field.String("asset_path"),
		field.String("content_sha256").Optional(),
		field.Int64("updated_by").Optional().Nillable(),
		field.Time("updated_at"),
	}
}
