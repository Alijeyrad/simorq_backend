package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// InternProfile extends a clinic member with intern-specific metadata.
type InternProfile struct {
	ent.Schema
}

func (InternProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (InternProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_member_id", uuid.UUID{}).
			Unique().
			Comment("FK → clinic_members.id"),

		field.Int("internship_year").
			Optional().
			Nillable(),

		field.JSON("supervisor_ids", []uuid.UUID{}).
			Optional().
			Comment("List of clinic_member UUIDs who supervise this intern"),
	}
}

func (InternProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("clinic_member_id").Unique(),
	}
}
