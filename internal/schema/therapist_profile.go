package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/google/uuid"
)

// TherapistProfile extends a ClinicMember (role=therapist) with clinical credentials
// and public-facing profile information.
type TherapistProfile struct {
	ent.Schema
}

func (TherapistProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (TherapistProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_member_id", uuid.UUID{}).
			Unique().
			Comment("FK → clinic_members.id (1:1)"),

		field.String("education").
			Optional().
			Nillable().
			MaxLen(255),

		field.String("psychology_license").
			Optional().
			Nillable().
			MaxLen(50).
			Comment("سازمان نظام روانشناسی license number"),

		field.String("approach").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Therapeutic approach, e.g. CBT, ACT"),

		field.JSON("specialties", []string{}).
			Optional().
			Comment("List of specialty tags"),

		field.Text("bio").
			Optional().
			Nillable(),

		field.Float("rating").
			Default(0).
			Comment("Aggregated rating (0–5)"),

		field.Int64("session_price").
			Optional().
			Nillable().
			Comment("Session price in Rials (for display; reservation fee set per slot)"),

		field.Int("session_duration_min").
			Optional().
			Nillable().
			Comment("Default session duration in minutes"),

		field.Bool("is_accepting").
			Default(true).
			Comment("Whether this therapist is accepting new patients"),
	}
}

func (TherapistProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", ClinicMember.Type).
			Ref("therapist_profile").
			Unique().
			Required().
			Field("clinic_member_id"),
	}
}
