package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
)

// File holds the schema definition for the File entity.
type File struct {
	ent.Schema
}

func (File) Fields() []ent.Field {
	return []ent.Field{
		uuidField(),
		field.String("filename").
			MinLen(1).
			MaxLen(255),
		field.String("extension").
			MinLen(1).
			MaxLen(10),
		field.Bytes("data"),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (File) Edges() []ent.Edge {
	return nil
}

func (File) Annotations() []entsql.Annotation {
	return []entsql.Annotation{
		{Table: "file"},
	}
}
