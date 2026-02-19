package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// NotificationPref holds per-user notification preferences.
type NotificationPref struct {
	ent.Schema
}

func (NotificationPref) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (NotificationPref) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Unique().
			Comment("FK → users.id"),

		field.Bool("appointment_sms").
			Default(true),

		field.Bool("appointment_push").
			Default(true),

		field.Bool("message_push").
			Default(true),

		field.Bool("ticket_reply_push").
			Default(true),
	}
}

func (NotificationPref) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
	}
}
