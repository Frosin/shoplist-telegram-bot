package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Shopping holds the schema definition for the Shopping entity.
type Shopping struct {
	ent.Schema
}

// Fields of the Shopping.
func (Shopping) Fields() []ent.Field {
	return []ent.Field{
		field.Time("date").Default(time.Now),
		field.Int("sum").Default(0),
		field.Bool("complete").Default(false),
		field.Int("type").Default(0),
	}
}

// Edges of the Shopping.
func (Shopping) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("item", Item.Type),
		edge.From("shop", Shop.Type).Ref("shopping").Unique(),
		edge.From("user", User.Type).Ref("shopping").Unique(),
	}
}
