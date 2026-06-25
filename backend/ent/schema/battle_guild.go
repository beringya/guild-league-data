package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type BattleGuild struct {
	ent.Schema
}

func (BattleGuild) Fields() []ent.Field {
	return []ent.Field{
		field.String("side"),
		field.Int("member_count"),
	}
}

func (BattleGuild) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("battle", Battle.Type).Ref("battle_guilds").Unique().Required(),
		edge.From("guild", Guild.Type).Ref("battle_guilds").Unique().Required(),
		edge.To("stats", BattlePlayerStat.Type),
	}
}

func (BattleGuild) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("side").Edges("battle").Unique(),
		index.Edges("battle", "guild").Unique(),
	}
}
