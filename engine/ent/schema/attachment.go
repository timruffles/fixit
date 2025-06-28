package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

type Attachment struct {
	ent.Schema
}

func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(newUUID).
			Unique().
			Immutable(),
		field.String("caption").
			Optional().
			MaxLen(500),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("attachments").
			Unique().
			Required(),
		edge.From("file", File.Type).
			Ref("attachments").
			Unique().
			Required(),
	}
}

func (Attachment) Annotations() []entsql.Annotation {
	return []entsql.Annotation{
		{Table: "attachment"},
	}
}
