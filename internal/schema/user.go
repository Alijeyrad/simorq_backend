package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
		SoftDeleteMixin{},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("first_name").
			Optional().
			Nillable().
			MaxLen(100),

		field.String("last_name").
			Optional().
			Nillable().
			MaxLen(100),

		field.String("phone").
			Optional().Nillable().
			Unique().
			MaxLen(20),

		field.String("email").
			Optional().
			Nillable().
			Unique().
			MaxLen(255),

		// National ID stored AES-256-GCM encrypted; nullable because clients
		// may register via phone only and fill national_id later.
		field.String("national_id").
			Optional().
			Nillable().
			MaxLen(500).
			Sensitive().
			Comment("AES-256-GCM encrypted national ID (کد ملی)"),

		// SHA-256 hex of raw national_id for fast uniqueness lookups.
		field.String("national_id_hash").
			Optional().
			Nillable().
			Unique().
			MaxLen(64).
			Comment("SHA-256 hex of national_id for indexed lookup"),

		field.String("gender").
			Optional().
			Nillable().
			MaxLen(10),

		field.String("marital_status").
			Optional().
			Nillable().
			MaxLen(20),

		// Jalali (Shamsi) birth year, e.g. 1370
		field.Int("birth_year").
			Optional().
			Nillable(),

		// S3 key for the user's avatar image
		field.String("avatar_key").
			Optional().
			Nillable().
			MaxLen(500),

		field.String("password_hash").
			Optional().
			Nillable().
			Sensitive(),

		field.Bool("must_change_password").
			Default(true),

		field.Enum("status").
			Values("ACTIVE", "SUSPENDED").
			Default("ACTIVE"),

		field.Bool("phone_verified").Default(false),
		field.Bool("email_verified").Default(false),

		field.Bool("twofa_phone_enabled").Default(false),
		field.Bool("twofa_email_enabled").Default(false),

		// Audit
		field.Time("last_login_at").
			Optional().
			Nillable(),

		field.Int("failed_login_attempts").
			Default(0).
			NonNegative(),

		field.Time("locked_until").
			Optional().
			Nillable().
			Comment("Account locked until this time after repeated login failures"),

		field.Time("last_failed_login_at").
			Optional().
			Nillable(),

		field.JSON("metadata", map[string]any{}).
			Optional().
			Default(map[string]any{}),

		field.Time("suspended_at").
			Optional().
			Nillable(),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("phone"),
		index.Fields("national_id_hash"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{}
}
