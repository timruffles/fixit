package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

func newUUID() uuid.UUID {
	v7, _ := uuid.NewV7()
	return v7
}

func uuidField() ent.Field {
	return field.UUID("id", uuid.UUID{}).
		Default(newUUID).
		Unique().
		Immutable()
}
