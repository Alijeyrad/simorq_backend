package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/google/uuid"
)

// Appointment is a booked session between a therapist and a patient.
type Appointment struct {
	ent.Schema
}

func (Appointment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UUIDV7Mixin{},
		TimeStampedMixin{},
	}
}

func (Appointment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("clinic_id", uuid.UUID{}).
			Comment("FK → clinics.id"),

		field.UUID("therapist_id", uuid.UUID{}).
			Comment("FK → clinic_members.id"),

		field.UUID("patient_id", uuid.UUID{}).
			Comment("FK → patients.id"),

		field.UUID("time_slot_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Snapshot ref to time_slots.id (nullable non-FK — allows slot deletion)"),

		field.Time("start_time"),

		field.Time("end_time"),

		field.Enum("status").
			Values("scheduled", "completed", "cancelled", "no_show").
			Default("scheduled"),

		field.Int64("session_price").
			Comment("Snapshotted session price in Rials"),

		field.Int64("reservation_fee").
			Default(0).
			Comment("Snapshotted reservation fee in Rials"),

		field.Enum("payment_status").
			Values("unpaid", "reservation_paid", "fully_paid", "refunded").
			Default("unpaid"),

		field.Text("notes").
			Optional().
			Nillable(),

		field.Text("cancellation_reason").
			Optional().
			Nillable(),

		field.Enum("cancel_requested_by").
			Values("patient", "therapist", "clinic").
			Optional().
			Nillable(),

		field.Time("cancelled_at").
			Optional().
			Nillable(),

		field.Int64("cancellation_fee").
			Default(0),

		field.Time("completed_at").
			Optional().
			Nillable(),
	}
}

func (Appointment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("clinic_id", "therapist_id", "start_time"),
		index.Fields("clinic_id", "patient_id"),
		index.Fields("therapist_id", "status", "start_time"),
		index.Fields("patient_id", "status"),
	}
}
