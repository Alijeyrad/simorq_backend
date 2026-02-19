package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// Conversation holds a direct messaging thread between two clinic members or a member and a patient.
type Conversation struct {
	ent.Schema
}

func (Conversation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (Conversation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.UUID("participant_a", uuid.UUID{}).
			Comment("First participant (user id)"),

		field.UUID("participant_b", uuid.UUID{}).
			Comment("Second participant (user id)"),

		field.UUID("patient_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Optional linked patient"),

		field.Time("last_message_at").
			Optional().
			Nillable(),

		field.Bool("is_active").
			Default(true),
	}
}

func (Conversation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("clinic_id", "participant_a", "participant_b").Unique(),
		index.Fields("clinic_id", "participant_a"),
		index.Fields("clinic_id", "participant_b"),
	}
}
