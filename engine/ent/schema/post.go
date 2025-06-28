package schema

import (
	"github.com/gofrs/uuid/v5"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Post struct {
	ent.Schema
}

func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				v7, _ := uuid.NewV7()
				return v7
			}).
			StorageKey("id"),
		field.String("title").
			MinLen(5).
			MaxLen(128),
	}
}

func (Post) Edges() []ent.Edge {
	return nil
}