package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Vote
type Vote struct {
	ent.Schema
}

// ENUM(interesting,solved)
type VoteKind string

func (Vote) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("vote"),
	}
}

func (Vote) Fields() []ent.Field {
	return []ent.Field{
		uuidField(),
		field.String("kind"),
		field.Int("value"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Vote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kind"),
		// each user can only have one type of each vote
		index.Fields("kind").
			Edges("user").
			Unique(),
	}
}

func (Vote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("post", Post.Type).Unique().Required(),
		edge.To("user", User.Type).Unique().Required(),
	}
}
