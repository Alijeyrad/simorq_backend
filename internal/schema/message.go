package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// Message is a single message within a Conversation.
type Message struct {
	ent.Schema
}

func (Message) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
		SoftDeleteMixin{},
	}
}

func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("conversation_id", uuid.UUID{}).
			Comment("FK → conversations.id"),

		field.UUID("sender_id", uuid.UUID{}).
			Comment("User id of the sender"),

		field.Text("content").
			Optional().
			Nillable(),

		field.String("file_key").
			Optional().
			Nillable(),

		field.String("file_name").
			Optional().
			Nillable(),

		field.String("file_mime").
			Optional().
			Nillable(),

		field.Bool("is_read").
			Default(false),

		field.Time("read_at").
			Optional().
			Nillable(),
	}
}

func (Message) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("conversation_id", "created_at"),
		index.Fields("sender_id"),
	}
}
