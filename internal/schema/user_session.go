package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// UserSession tracks PASETO sessions for audit and revocation.
// Redis is the primary session store; this table is the audit trail.
type UserSession struct {
	ent.Schema
}

func (UserSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (UserSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("FK → users.id"),

		// session_id is the "sid" claim in the PASETO token
		field.String("session_id").
			Unique().
			NotEmpty().
			MaxLen(36).
			Immutable().
			Comment("UUID stored in PASETO sid claim"),

		// sha-256 hex of the refresh token for look-up without storing plaintext
		field.String("refresh_token_hash").
			Optional().
			Nillable().
			MaxLen(64).
			Sensitive(),

		field.String("user_agent").
			Optional().
			Nillable(),

		field.String("ip_address").
			Optional().
			Nillable().
			MaxLen(45),

		field.Time("expires_at").
			Comment("When the refresh token (and thus the session) expires"),

		field.Time("last_used_at").
			Optional().
			Nillable(),

		field.Time("revoked_at").
			Optional().
			Nillable().
			Comment("Set on logout or token rotation invalidation"),
	}
}

func (UserSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id"),
		index.Fields("user_id"),
	}
}

func (UserSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
	}
}
