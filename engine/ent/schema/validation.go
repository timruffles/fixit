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

// Validate
type Validation struct {
	ent.Schema
}

func (Validation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("validation"),
	}
}

func (Validation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				v7, _ := uuid.NewV7()
				return v7
			}).
			Unique().
			Immutable(),
		field.Bool("is_valid").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (Validation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("validations").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("validations").
			Unique().
			Required(),
	}
}
