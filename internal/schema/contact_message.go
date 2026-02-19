package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// ContactMessage stores messages submitted via the public contact form.
type ContactMessage struct {
	ent.Schema
}

func (ContactMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (ContactMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(255),

		field.String("email").
			MaxLen(255),

		field.String("subject").
			MaxLen(255),

		field.Text("message"),
	}
}
