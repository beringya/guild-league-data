package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type ImportLog struct {
	ent.Schema
}

func (ImportLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("level"),
		field.String("code"),
		field.Int("row_number").Optional().Nillable(),
		field.String("message"),
		field.Time("created_at"),
	}
}

func (ImportLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("battle", Battle.Type).Ref("import_logs").Unique(),
	}
}
