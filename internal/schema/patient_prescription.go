package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/google/uuid"
)

// PatientPrescription is a prescription or homework assignment issued to a patient.
type PatientPrescription struct {
	ent.Schema
}

func (PatientPrescription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (PatientPrescription) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("patient_id", uuid.UUID{}).
			Comment("FK → patients.id"),

		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id (for tenant-scoped queries)"),

		field.UUID("therapist_id", uuid.UUID{}).
			Comment("FK → clinic_members.id"),

		field.String("title").
			Optional().
			Nillable().
			MaxLen(255),

		field.Text("notes").
			Optional().
			Nillable(),

		field.String("file_key").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("Optional attached file (S3 key)"),

		field.String("file_name").
			Optional().
			Nillable().
			MaxLen(255),

		field.Time("prescribed_date").
			Default(time.Now).
			Comment("Date the prescription was issued"),
	}
}

func (PatientPrescription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("patient", Patient.Type).
			Ref("prescriptions").
			Unique().
			Required().
			Field("patient_id"),
		edge.To("therapist", ClinicMember.Type).
			Unique().
			Required().
			Field("therapist_id"),
	}
}
