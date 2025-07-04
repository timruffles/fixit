package schema

import (
	"time"

	"github.com/gofrs/uuid/v5"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Community holds the schema definition for the Community entity.
type Community struct {
	ent.Schema
}

func (Community) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "community"},
	}
}

func (Community) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(newUUID).
			Unique().
			Immutable(),
		field.String("name").
			MinLen(5).
			MaxLen(128),
		field.String("title").
			MinLen(5).
			MaxLen(128),
		field.String("location").
			Optional(),
		field.String("banner_image_url").
			Optional(),
		field.String("geography").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Community) Edges() []ent.Edge {
	return nil
}
