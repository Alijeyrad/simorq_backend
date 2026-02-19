package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// InternPatientAccess grants an intern access to a specific patient record.
type InternPatientAccess struct {
	ent.Schema
}

func (InternPatientAccess) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (InternPatientAccess) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("intern_id", uuid.UUID{}).
			Comment("FK → clinic_members.id (intern)"),

		field.UUID("patient_id", uuid.UUID{}).
			Comment("FK → patients.id"),

		field.UUID("granted_by", uuid.UUID{}).
			Comment("FK → clinic_members.id (who granted)"),

		field.Bool("can_view_files").
			Default(true),

		field.Bool("can_write_reports").
			Default(false),
	}
}

func (InternPatientAccess) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("intern_id", "patient_id").Unique(),
		index.Fields("intern_id"),
		index.Fields("patient_id"),
	}
}
