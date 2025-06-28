package schema

import (
	"time"

	"entgo.io/ent/dialect/entsql"
	"github.com/gofrs/uuid/v5"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Solution struct {
	ent.Schema
}

func (Solution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("solution"),
	}
}

func (Solution) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				v7, _ := uuid.NewV7()
				return v7
			}).
			Unique().
			Immutable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (Solution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("solution").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("solutions").
			Unique().
			Required(),
	}
}