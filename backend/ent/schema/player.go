package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Player struct {
	ent.Schema
}

func (Player) Fields() []ent.Field {
	return []ent.Field{
		field.String("canonical_name"),
		field.Time("created_at"),
	}
}

func (Player) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("guild", Guild.Type).Ref("players").Unique().Required(),
		edge.To("stats", BattlePlayerStat.Type),
		edge.To("avatar", PlayerAvatar.Type).Unique(),
	}
}

func (Player) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("canonical_name").Edges("guild").Unique(),
	}
}
