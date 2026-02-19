package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// TimeSlot represents a bookable time block for a therapist.
type TimeSlot struct {
	ent.Schema
}

func (TimeSlot) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (TimeSlot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("therapist_id", uuid.UUID{}).
			Comment("FK → clinic_members.id"),

		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.Time("start_time"),

		field.Time("end_time"),

		field.Enum("status").
			Values("available", "booked", "blocked", "cancelled").
			Default("available"),

		field.Int64("session_price").
			Optional().
			Nillable().
			Comment("Override session price in Rials; nil = use therapist default"),

		field.Int64("reservation_fee").
			Optional().
			Nillable().
			Comment("Override reservation fee in Rials; nil = use clinic default"),

		field.Bool("is_recurring").
			Default(false),

		field.UUID("recurring_rule_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Non-FK reference to the recurring_rule that generated this slot"),
	}
}

func (TimeSlot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("therapist_id", "start_time"),
		index.Fields("clinic_id", "status", "start_time"),
	}
}
