package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type PlayerAvatar struct {
	ent.Schema
}

func (PlayerAvatar) Fields() []ent.Field {
	return []ent.Field{
		field.String("source"),
		field.String("seed").Optional(),
		field.String("asset_path").Optional(),
		field.String("content_sha256").Optional(),
		field.Int64("updated_by").Optional().Nillable(),
		field.Time("updated_at"),
	}
}

func (PlayerAvatar) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("player", Player.Type).Ref("avatar").Unique().Required(),
	}
}
