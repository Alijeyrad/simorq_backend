package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// RecurringRule defines a weekly recurring schedule for a therapist.
type RecurringRule struct {
	ent.Schema
}

func (RecurringRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (RecurringRule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("therapist_id", uuid.UUID{}).
			Comment("FK → clinic_members.id"),

		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.Int8("day_of_week").
			Comment("0=Sunday, 1=Monday … 6=Saturday"),

		field.Int8("start_hour"),

		field.Int8("start_minute"),

		field.Int8("end_hour"),

		field.Int8("end_minute"),

		field.Int64("session_price").
			Optional().
			Nillable(),

		field.Int64("reservation_fee").
			Optional().
			Nillable(),

		field.Time("valid_from").
			Comment("Rule takes effect from this date"),

		field.Time("valid_until").
			Optional().
			Nillable().
			Comment("Rule expires after this date; nil = no expiry"),

		field.Bool("is_active").
			Default(true),
	}
}

func (RecurringRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("therapist_id", "day_of_week", "is_active"),
		index.Fields("clinic_id"),
	}
}
