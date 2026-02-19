package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// InternTask is a task submitted by an intern for review.
type InternTask struct {
	ent.Schema
}

func (InternTask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (InternTask) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.UUID("intern_id", uuid.UUID{}).
			Comment("FK → clinic_members.id (intern)"),

		field.String("title").
			MaxLen(255),

		field.Text("caption").
			Optional().
			Nillable(),

		field.Time("submitted_at").
			Default(time.Now),

		field.Enum("review_status").
			Values("pending", "reviewed", "needs_revision").
			Default("pending"),

		field.UUID("reviewed_by", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("FK → clinic_members.id (reviewer)"),

		field.Text("review_comment").
			Optional().
			Nillable(),

		field.String("grade").
			MaxLen(20).
			Optional().
			Nillable(),

		field.Time("reviewed_at").
			Optional().
			Nillable(),
	}
}

func (InternTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("clinic_id", "intern_id", "submitted_at"),
		index.Fields("intern_id", "review_status"),
	}
}
