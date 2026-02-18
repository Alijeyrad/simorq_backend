package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Clinic
// ---------------------------------------------------------------------------

type Clinic struct {
	ent.Schema
}

func (Clinic) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
		SoftDeleteMixin{},
	}
}

func (Clinic) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(255).
			NotEmpty(),

		field.String("slug").
			MaxLen(100).
			NotEmpty().
			Unique().
			Comment("URL-friendly identifier for the clinic"),

		field.String("description").
			Optional().
			Nillable(),

		field.String("logo_key").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("S3 key for clinic logo"),

		field.String("phone").
			Optional().
			Nillable().
			MaxLen(20),

		field.String("address").
			Optional().
			Nillable(),

		field.String("city").
			Optional().
			Nillable().
			MaxLen(100),

		field.String("province").
			Optional().
			Nillable().
			MaxLen(100),

		field.Bool("is_active").Default(true),

		field.Bool("is_verified").
			Default(false).
			Comment("Platform-level verification status"),
	}
}

func (Clinic) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug"),
	}
}

func (Clinic) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", ClinicMember.Type),
		edge.To("settings", ClinicSettings.Type).Unique(),
	}
}

// ---------------------------------------------------------------------------
// ClinicMember — join table: user ↔ clinic with role
// ---------------------------------------------------------------------------

type ClinicMember struct {
	ent.Schema
}

func (ClinicMember) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
	}
}

func (ClinicMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.UUID("user_id", uuid.UUID{}).
			Comment("FK → users.id"),

		field.Enum("role").
			Values("owner", "admin", "therapist", "intern").
			Comment("Role of this user in the clinic"),

		field.Bool("is_active").Default(true),

		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

func (ClinicMember) Indexes() []ent.Index {
	return []ent.Index{
		// A user can only have one membership record per clinic
		index.Fields("clinic_id", "user_id").Unique(),
		index.Fields("clinic_id"),
		index.Fields("user_id"),
	}
}

func (ClinicMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("clinic", Clinic.Type).
			Ref("members").
			Unique().
			Required().
			Field("clinic_id"),
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
	}
}

// ---------------------------------------------------------------------------
// ClinicSettings — one-to-one with Clinic
// ---------------------------------------------------------------------------

type ClinicSettings struct {
	ent.Schema
}

func (ClinicSettings) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (ClinicSettings) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Unique().
			Comment("FK → clinics.id"),

		// Reservation policy
		field.Int64("reservation_fee_amount").Default(0).
			Comment("Fixed reservation fee in Rials"),

		field.Int("reservation_fee_percent").Default(0).
			Comment("Reservation fee as percentage of session price"),

		field.Int("cancellation_window_hours").Default(24).
			Comment("Hours before appointment when free cancellation is allowed"),

		field.Int64("cancellation_fee_amount").Default(0),

		field.Int("cancellation_fee_percent").Default(0),

		field.Bool("allow_client_self_book").Default(true).
			Comment("Clients can book slots without staff intervention"),

		// Session defaults
		field.Int("default_session_duration_min").Default(60),

		field.Int64("default_session_price").Default(0).
			Comment("Default session price in Rials; therapists can override"),

		// Working hours stored as JSONB: {"saturday": {"start": "08:00", "end": "20:00"}, ...}
		field.JSON("working_hours", map[string]any{}).
			Optional(),
	}
}

func (ClinicSettings) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("clinic", Clinic.Type).
			Ref("settings").
			Unique().
			Required().
			Field("clinic_id"),
	}
}
