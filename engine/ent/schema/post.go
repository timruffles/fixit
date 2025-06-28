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

type Post struct {
	ent.Schema
}

func (Post) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("post"),
	}
}

func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				v7, _ := uuid.NewV7()
				return v7
			}).
			Unique().
			Immutable(),
		field.String("title").
			MinLen(5).
			MaxLen(128),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.JSON("tags", []string{}).
			Optional().
			Default([]string{}),
		field.UUID("reply_to", uuid.UUID{}).
			Optional(),
	}
}

func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		// o2o
		edge.To("user", User.Type).Unique(),
		edge.To("parent", Post.Type).
			Field("reply_to").
			Unique().
			From("replies"),
		// o2m
		edge.From("votes", Vote.Type).Ref("post"),
		// TODO - getting "entc/gen: type "Attachment" does not exist for edge", not required for now
		//edge.From("attachments", Attachment.Type).Ref("post"),
	}
}
