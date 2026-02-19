package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/google/uuid"
)

// CommissionRule defines the platform/clinic fee split for a specific clinic.
type CommissionRule struct {
	ent.Schema
}

func (CommissionRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (CommissionRule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Unique().
			Comment("FK → clinics.id (one rule per clinic)"),

		field.Int("platform_fee_percent").
			Default(0).
			Comment("Platform commission percentage (0–100)"),

		field.Int("clinic_fee_percent").
			Default(0).
			Comment("Clinic fee percentage retained before therapist payout (0–100)"),

		field.Bool("is_flat_fee").
			Default(false).
			Comment("If true, flat_fee_amount is used instead of percentage"),

		field.Int64("flat_fee_amount").
			Default(0).
			Comment("Fixed platform fee in Rials (used when is_flat_fee=true)"),

		field.Bool("is_active").
			Default(true),
	}
}
