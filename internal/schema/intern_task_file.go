package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// InternTaskFile is a file attachment on an InternTask.
type InternTaskFile struct {
	ent.Schema
}

func (InternTaskFile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		CreatedAtMixin{},
	}
}

func (InternTaskFile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("task_id", uuid.UUID{}).
			Comment("FK → intern_tasks.id"),

		field.String("file_key").
			MaxLen(500).
			Comment("S3 key"),

		field.String("file_name").
			MaxLen(255),

		field.Int64("file_size").
			Optional().
			Nillable(),

		field.String("mime_type").
			MaxLen(100).
			Optional().
			Nillable(),
	}
}

func (InternTaskFile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
	}
}
