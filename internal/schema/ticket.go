package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// Ticket is a support ticket submitted by a user.
type Ticket struct {
	ent.Schema
}

func (Ticket) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (Ticket) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Optional clinic scope"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("Submitting user"),

		field.String("subject").
			MaxLen(255),

		field.Enum("status").
			Values("open", "answered", "closed").
			Default("open"),

		field.Enum("priority").
			Values("low", "normal", "high", "urgent").
			Default("normal"),
	}
}

func (Ticket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "status"),
		index.Fields("clinic_id", "status"),
	}
}
