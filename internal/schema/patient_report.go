package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// PatientReport is a clinical session report written by a therapist.
type PatientReport struct {
	ent.Schema
}

func (PatientReport) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (PatientReport) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("patient_id", uuid.UUID{}).
			Comment("FK → patients.id"),

		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id (for tenant-scoped queries)"),

		field.UUID("therapist_id", uuid.UUID{}).
			Comment("FK → clinic_members.id (author of the report)"),

		field.UUID("appointment_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("FK → appointments.id (optional link to session)"),

		field.String("title").
			Optional().
			Nillable().
			MaxLen(255),

		field.Text("content").
			Optional().
			Nillable(),

		field.Time("report_date").
			Default(time.Now).
			Comment("Date of the session or report"),
	}
}

func (PatientReport) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("patient", Patient.Type).
			Ref("reports").
			Unique().
			Required().
			Field("patient_id"),
		edge.To("therapist", ClinicMember.Type).
			Unique().
			Required().
			Field("therapist_id"),
	}
}

func (PatientReport) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("patient_id"),
		index.Fields("clinic_id"),
		index.Fields("therapist_id"),
		index.Fields("report_date"),
	}
}
