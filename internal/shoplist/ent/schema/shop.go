package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Shop holds the schema definition for the Shop entity.
type Shop struct {
	ent.Schema
}

// Fields of the Shop.
func (Shop) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
	}
}

// Edges of the Shop.
func (Shop) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("shopping", Shopping.Type),
	}
}
