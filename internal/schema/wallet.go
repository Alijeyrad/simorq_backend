package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// Wallet holds a balance for a user, clinic, or platform account.
type Wallet struct {
	ent.Schema
}

func (Wallet) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (Wallet) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("owner_type").
			Values("user", "clinic", "platform").
			Comment("Discriminator for polymorphic ownership"),

		field.UUID("owner_id", uuid.UUID{}).
			Comment("ID of the owning entity (user_id, clinic_id, or platform sentinel)"),

		field.Int64("balance").
			Default(0).
			Comment("Current balance in Rials"),

		field.String("iban_encrypted").
			MaxLen(1000).
			Optional().
			Nillable().
			Comment("AES-256-GCM encrypted IBAN"),

		field.String("iban_hash").
			MaxLen(64).
			Optional().
			Nillable().
			Comment("SHA-256 hash of the plaintext IBAN for uniqueness lookup"),

		field.String("account_holder").
			MaxLen(200).
			Optional().
			Nillable(),
	}
}

func (Wallet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("transactions", Transaction.Type),
		edge.To("withdrawals", WithdrawalRequest.Type),
	}
}

func (Wallet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_type", "owner_id").Unique(),
	}
}
