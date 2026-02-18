package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// PatientFile tracks files uploaded for a patient (stored in S3).
type PatientFile struct {
	ent.Schema
}

func (PatientFile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (PatientFile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("patient_id", uuid.UUID{}).
			Comment("FK → patients.id"),

		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id (for tenant-scoped queries)"),

		field.UUID("uploaded_by", uuid.UUID{}).
			Comment("FK → users.id"),

		field.String("linked_type").
			Optional().
			Nillable().
			MaxLen(30).
			Comment("'report', 'test_result', 'prescription', or NULL for standalone"),

		field.UUID("linked_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID of the linked entity (report, test, prescription)"),

		field.String("file_name").
			MaxLen(255).
			NotEmpty(),

		field.String("file_key").
			MaxLen(500).
			NotEmpty().
			Comment("S3 object key"),

		field.Int64("file_size").
			Optional().
			Nillable().
			Comment("File size in bytes"),

		field.String("mime_type").
			Optional().
			Nillable().
			MaxLen(100),

		field.Text("description").
			Optional().
			Nillable(),
	}
}

func (PatientFile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("patient", Patient.Type).
			Ref("files").
			Unique().
			Required().
			Field("patient_id"),
		edge.To("uploader", User.Type).
			Unique().
			Required().
			Field("uploaded_by"),
	}
}

func (PatientFile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("patient_id"),
		index.Fields("linked_type", "linked_id"),
		index.Fields("clinic_id"),
	}
}
