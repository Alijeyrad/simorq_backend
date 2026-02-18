package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/google/uuid"
)

// PatientTest records the administration of a psychological test to a patient.
type PatientTest struct {
	ent.Schema
}

func (PatientTest) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (PatientTest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("patient_id", uuid.UUID{}).
			Comment("FK → patients.id"),

		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id (for tenant-scoped queries)"),

		field.UUID("test_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("FK → psych_tests.id (NULL if free-text test)"),

		field.UUID("administered_by", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("FK → clinic_members.id"),

		field.String("test_name").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Free-text name when test_id is NULL"),

		field.JSON("raw_scores", map[string]any{}).
			Optional().
			Comment("Raw test scores as JSONB"),

		field.JSON("computed_scores", map[string]any{}).
			Optional().
			Comment("Computed/normalised scores as JSONB"),

		field.Text("interpretation").
			Optional().
			Nillable(),

		field.Time("test_date").
			Default(time.Now),

		field.Enum("status").
			Values("assigned", "completed", "reviewed").
			Default("assigned"),
	}
}

func (PatientTest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("patient", Patient.Type).
			Ref("tests").
			Unique().
			Required().
			Field("patient_id"),
		edge.To("psych_test", PsychTest.Type).
			Unique().
			Field("test_id"),
		edge.To("administrator", ClinicMember.Type).
			Unique().
			Field("administered_by"),
	}
}
