package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// WithdrawalRequest is a payout request from a clinic wallet.
type WithdrawalRequest struct {
	ent.Schema
}

func (WithdrawalRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (WithdrawalRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("wallet_id", uuid.UUID{}).
			Comment("FK → wallets.id"),

		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.Int64("amount").
			Comment("Requested payout amount in Rials"),

		field.Enum("status").
			Values("pending", "processing", "completed", "failed", "cancelled").
			Default("pending"),

		field.String("iban_encrypted").
			MaxLen(1000).
			Comment("AES-256-GCM encrypted IBAN at request time"),

		field.String("account_holder").
			MaxLen(200),

		field.String("bank_ref").
			MaxLen(100).
			Optional().
			Nillable().
			Comment("Bank transfer reference number"),

		field.Time("requested_at").
			Default(time.Now).
			Immutable(),

		field.Time("processed_at").
			Optional().
			Nillable(),

		field.Text("failure_reason").
			Optional().
			Nillable(),
	}
}

func (WithdrawalRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("wallet", Wallet.Type).
			Ref("withdrawals").
			Unique().
			Required().
			Field("wallet_id"),
	}
}

func (WithdrawalRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("wallet_id", "status"),
		index.Fields("clinic_id", "status", "requested_at"),
	}
}
