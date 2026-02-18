package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/google/uuid"
)

type UUIDV7Mixin struct {
	mixin.Schema
}

func (UUIDV7Mixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				id, err := uuid.NewV7()
				if err != nil {
					panic(err)
				}
				return id
			}).
			Immutable(),
	}
}

type TimeStampedMixin struct {
	mixin.Schema
}

func (TimeStampedMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

type SoftDeleteMixin struct {
	mixin.Schema
}

func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			Optional().
			Nillable().
			Default(nil),
	}
}

// CreatedAtMixin provides only created_at (for append-only records).
type CreatedAtMixin struct {
	mixin.Schema
}

func (CreatedAtMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}
