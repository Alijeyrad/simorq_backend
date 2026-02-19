package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// Patient represents a per-clinic patient record.
// A user can be a patient in multiple clinics.
type Patient struct {
	ent.Schema
}

func (Patient) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
		SoftDeleteMixin{},
	}
}

func (Patient) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("FK → users.id (the patient's user account)"),

		field.UUID("primary_therapist_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("FK → clinic_members.id (assigned therapist)"),

		field.String("file_number").
			Optional().
			Nillable().
			MaxLen(50).
			Comment("Internal file/case number assigned by clinic"),

		field.Enum("status").
			Values("active", "waiting_reservation", "inactive", "discharged").
			Default("active"),

		field.Int("session_count").
			Default(0),

		field.Int("total_cancellations").
			Default(0),

		field.Text("last_cancel_reason").
			Optional().
			Nillable(),

		field.Bool("has_discount").
			Default(false),

		field.Int("discount_percent").
			Default(0),

		field.Enum("payment_status").
			Values("paid", "unpaid", "partial").
			Default("unpaid"),

		field.Int64("total_paid").
			Default(0).
			Comment("Total amount paid by patient in Rials"),

		field.Text("notes").
			Optional().
			Nillable(),

		field.String("referral_source").
			Optional().
			Nillable().
			MaxLen(255),

		field.Text("chief_complaint").
			Optional().
			Nillable().
			Comment("اشکال اصلی / presenting problem"),

		// Child-specific fields
		field.Bool("is_child").
			Default(false),

		field.Time("child_birth_date").
			Optional().
			Nillable(),

		field.String("child_school").
			Optional().
			Nillable().
			MaxLen(255),

		field.String("child_grade").
			Optional().
			Nillable().
			MaxLen(50),

		field.String("parent_name").
			Optional().
			Nillable().
			MaxLen(255),

		field.String("parent_phone").
			Optional().
			Nillable().
			MaxLen(11),

		field.String("parent_relation").
			Optional().
			Nillable().
			MaxLen(50).
			Comment("e.g. father, mother, guardian"),

		field.JSON("developmental_history", map[string]any{}).
			Optional().
			Comment("Free-form developmental history JSON"),
	}
}

func (Patient) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("clinic", Clinic.Type).
			Ref("patients").
			Unique().
			Required().
			Field("clinic_id"),
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
		edge.To("primary_therapist", ClinicMember.Type).
			Unique().
			Field("primary_therapist_id"),
		edge.To("reports", PatientReport.Type),
		edge.To("files", PatientFile.Type),
		edge.To("prescriptions", PatientPrescription.Type),
		edge.To("tests", PatientTest.Type),
	}
}

func (Patient) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("clinic_id", "user_id").Unique(),
		index.Fields("clinic_id"),
		index.Fields("user_id"),
		index.Fields("clinic_id", "status"),
		index.Fields("clinic_id", "file_number"),
	}
}
