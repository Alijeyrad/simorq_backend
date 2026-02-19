package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// UserDevice stores push notification device tokens per user.
type UserDevice struct {
	ent.Schema
}

func (UserDevice) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (UserDevice) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("FK → users.id"),

		field.String("device_token").
			MaxLen(512),

		field.Enum("platform").
			Values("web", "android", "ios"),

		field.Bool("is_active").
			Default(true),
	}
}

func (UserDevice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "device_token").Unique(),
		index.Fields("user_id", "is_active"),
	}
}
