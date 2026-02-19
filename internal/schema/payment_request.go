package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// PaymentRequest tracks a single payment attempt via ZarinPal or wallet.
type PaymentRequest struct {
	ent.Schema
}

func (PaymentRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (PaymentRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("FK → users.id (the payer)"),

		field.UUID("appointment_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("FK → appointments.id (optional link)"),

		field.Int64("amount").
			Comment("Amount in Rials"),

		field.String("description").
			MaxLen(500),

		field.Enum("status").
			Values("pending", "success", "failed", "cancelled").
			Default("pending"),

		field.Enum("source").
			Values("zarinpal", "wallet").
			Default("zarinpal"),

		field.String("zarinpal_authority").
			MaxLen(200).
			Optional().
			Nillable(),

		field.String("zarinpal_ref_id").
			MaxLen(50).
			Optional().
			Nillable().
			Comment("Stores int64 ref_id as string"),

		field.String("zarinpal_card_pan").
			MaxLen(25).
			Optional().
			Nillable().
			Comment("Masked card number e.g. 502229******5995"),

		field.String("zarinpal_card_hash").
			MaxLen(70).
			Optional().
			Nillable(),

		field.Time("paid_at").
			Optional().
			Nillable(),
	}
}

func (PaymentRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "status", "created_at"),
		index.Fields("clinic_id", "status"),
		index.Fields("zarinpal_authority"),
	}
}
