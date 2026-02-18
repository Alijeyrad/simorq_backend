package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// ClinicPermission stores per-user permission overrides within a clinic.
// These override the default Casbin role-based policies.
type ClinicPermission struct {
	ent.Schema
}

func (ClinicPermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (ClinicPermission) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("FK → users.id"),

		field.String("resource_type").
			MaxLen(50).
			NotEmpty().
			Comment("Casbin resource type, e.g. 'patient', 'patient_file'"),

		field.UUID("resource_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Specific resource ID for per-resource overrides; NULL = all"),

		field.String("action").
			MaxLen(20).
			NotEmpty().
			Comment("Casbin action, e.g. 'read', 'update', 'manage'"),

		field.Bool("granted").
			Default(true).
			Comment("true = explicitly allow, false = explicitly deny"),
	}
}

func (ClinicPermission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("clinic", Clinic.Type).
			Ref("permissions").
			Unique().
			Required().
			Field("clinic_id"),
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
	}
}

func (ClinicPermission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("clinic_id", "user_id"),
	}
}
