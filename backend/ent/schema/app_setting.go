package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type AppSetting struct {
	ent.Schema
}

func (AppSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").Unique().Immutable(),
		field.JSON("value_json", map[string]interface{}{}),
		field.Time("updated_at"),
	}
}
