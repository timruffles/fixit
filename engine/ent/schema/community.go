package schema

import (
	"github.com/gofrs/uuid/v5"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Community holds the schema definition for the Community entity.
type Community struct {
	ent.Schema
}

func (Community) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				v7, _ := uuid.NewV7()
				return v7
			}).
			StorageKey("id"),
		field.String("name").
			MinLen(5).
			MaxLen(128),
		field.String("title").
			MinLen(5).
			MaxLen(128),
	}
}

func (Community) Edges() []ent.Edge {
	return nil
}
