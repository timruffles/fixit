package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Attachment struct {
	ent.Schema
}

func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		uuidField(),
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
		edge.To("post", Post.Type).
			Unique().Required(),
		edge.To("file", File.Type).
			Unique().Required(),
	}
}

func (Attachment) Annotations() []entsql.Annotation {
	return []entsql.Annotation{
		{Table: "attachment"},
	}
}
