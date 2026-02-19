package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// Notification is an in-app or push notification for a user.
type Notification struct {
	ent.Schema
}

func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("Target user"),

		field.String("type").
			MaxLen(64).
			Comment("e.g. message_new, ticket_replied, appointment_created"),

		field.String("title").
			MaxLen(255),

		field.Text("body").
			Optional().
			Nillable(),

		field.JSON("data", map[string]any{}).
			Optional().
			Comment("Arbitrary JSON payload"),

		field.Bool("is_read").
			Default(false),

		field.Bool("is_pushed").
			Default(false).
			Comment("Whether a push notification was sent"),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "is_read", "created_at"),
	}
}
