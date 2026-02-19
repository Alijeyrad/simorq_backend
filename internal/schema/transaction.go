package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// Transaction is an append-only ledger entry for a wallet.
type Transaction struct {
	ent.Schema
}

func (Transaction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("wallet_id", uuid.UUID{}).
			Comment("FK → wallets.id"),

		field.Enum("type").
			Values("credit", "debit"),

		field.Int64("amount").
			Comment("Amount in Rials (always positive)"),

		field.Int64("balance_before").
			Comment("Wallet balance before this transaction"),

		field.Int64("balance_after").
			Comment("Wallet balance after this transaction"),

		field.String("entity_type").
			MaxLen(30).
			Optional().
			Nillable().
			Comment("Type of the related entity (e.g. appointment, payment_request)"),

		field.UUID("entity_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID of the related entity"),

		field.String("description").
			MaxLen(500).
			Optional().
			Nillable(),
	}
}

func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("wallet", Wallet.Type).
			Ref("transactions").
			Unique().
			Required().
			Field("wallet_id"),
	}
}

func (Transaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("wallet_id", "created_at"),
		index.Fields("entity_type", "entity_id"),
	}
}
