package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// PsychTest is a platform-wide psychological test catalog entry.
// Tests are managed by the platform superadmin and referenced by PatientTest.
type PsychTest struct {
	ent.Schema
}

func (PsychTest) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (PsychTest) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(255).
			NotEmpty().
			Comment("Test name in Latin/English"),

		field.String("name_fa").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Test name in Persian"),

		field.Text("description").
			Optional().
			Nillable(),

		field.String("category").
			Optional().
			Nillable().
			MaxLen(100).
			Comment("e.g. 'anxiety', 'depression', 'personality'"),

		field.String("age_range").
			Optional().
			Nillable().
			MaxLen(50).
			Comment("e.g. '6-18', 'adult'"),

		field.JSON("schema_data", map[string]any{}).
			Optional().
			Comment("Test schema / questions (flexible JSONB)"),

		field.String("scoring_method").
			Optional().
			Nillable().
			MaxLen(50),

		field.Bool("is_active").
			Default(true),
	}
}
