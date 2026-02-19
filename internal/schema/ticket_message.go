package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// TicketMessage is a reply within a support Ticket thread.
type TicketMessage struct {
	ent.Schema
}

func (TicketMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (TicketMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("ticket_id", uuid.UUID{}).
			Comment("FK → tickets.id"),

		field.UUID("sender_id", uuid.UUID{}).
			Comment("User id of the sender"),

		field.Text("content"),

		field.String("file_key").
			Optional().
			Nillable(),

		field.String("file_name").
			Optional().
			Nillable(),
	}
}

func (TicketMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ticket_id", "created_at"),
	}
}
