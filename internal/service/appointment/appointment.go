package appointment

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entappt "github.com/Alijeyrad/simorq_backend/internal/repo/appointment"
	entslot "github.com/Alijeyrad/simorq_backend/internal/repo/timeslot"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type ListRequest struct {
	TherapistID *uuid.UUID
	PatientID   *uuid.UUID
	Status      *string
	From        *time.Time
	To          *time.Time
	Page        int
	PerPage     int
}

type BookRequest struct {
	TherapistID    uuid.UUID
	PatientID      uuid.UUID
	TimeSlotID     *uuid.UUID
	StartTime      time.Time
	EndTime        time.Time
	SessionPrice   int64
	ReservationFee int64
	Notes          *string
}

type CancelRequest struct {
	Reason           *string
	RequestedBy      string // "patient" | "therapist" | "clinic"
	CancellationFee  int64
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type Service interface {
	List(ctx context.Context, clinicID uuid.UUID, req ListRequest) ([]*repo.Appointment, error)
	GetByID(ctx context.Context, clinicID, apptID uuid.UUID) (*repo.Appointment, error)
	Book(ctx context.Context, clinicID uuid.UUID, req BookRequest) (*repo.Appointment, error)
	Cancel(ctx context.Context, clinicID, apptID uuid.UUID, req CancelRequest) error
	Complete(ctx context.Context, clinicID, therapistMemberID, apptID uuid.UUID) error
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type appointmentService struct {
	db *repo.Client
}

func New(db *repo.Client) Service {
	return &appointmentService{db: db}
}

func (s *appointmentService) List(ctx context.Context, clinicID uuid.UUID, req ListRequest) ([]*repo.Appointment, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}
	offset := (req.Page - 1) * req.PerPage

	q := s.db.Appointment.Query().
		Where(entappt.ClinicID(clinicID))

	if req.TherapistID != nil {
		q = q.Where(entappt.TherapistID(*req.TherapistID))
	}
	if req.PatientID != nil {
		q = q.Where(entappt.PatientID(*req.PatientID))
	}
	if req.Status != nil {
		q = q.Where(entappt.StatusEQ(entappt.Status(*req.Status)))
	}
	if req.From != nil {
		q = q.Where(entappt.StartTimeGTE(*req.From))
	}
	if req.To != nil {
		q = q.Where(entappt.StartTimeLT(*req.To))
	}

	q = q.Order(entappt.ByStartTime(sql.OrderDesc()))

	appts, err := q.Offset(offset).Limit(req.PerPage).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list appointments: %w", err)
	}
	return appts, nil
}

func (s *appointmentService) GetByID(ctx context.Context, clinicID, apptID uuid.UUID) (*repo.Appointment, error) {
	appt, err := s.db.Appointment.Query().
		Where(entappt.ID(apptID), entappt.ClinicID(clinicID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get appointment: %w", err)
	}
	return appt, nil
}

func (s *appointmentService) Book(ctx context.Context, clinicID uuid.UUID, req BookRequest) (*repo.Appointment, error) {
	// If a time slot ID is provided, lock the slot atomically
	if req.TimeSlotID != nil {
		updated, err := s.db.TimeSlot.Update().
			Where(
				entslot.ID(*req.TimeSlotID),
				entslot.ClinicID(clinicID),
				entslot.StatusEQ(entslot.StatusAvailable),
			).
			SetStatus(entslot.StatusBooked).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("lock slot: %w", err)
		}
		if updated == 0 {
			return nil, ErrSlotNotAvailable
		}
	}

	c := s.db.Appointment.Create().
		SetClinicID(clinicID).
		SetTherapistID(req.TherapistID).
		SetPatientID(req.PatientID).
		SetStartTime(req.StartTime).
		SetEndTime(req.EndTime).
		SetSessionPrice(req.SessionPrice).
		SetReservationFee(req.ReservationFee)

	if req.TimeSlotID != nil {
		c = c.SetTimeSlotID(*req.TimeSlotID)
	}
	if req.Notes != nil {
		c = c.SetNillableNotes(req.Notes)
	}

	appt, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create appointment: %w", err)
	}
	return appt, nil
}

func (s *appointmentService) Cancel(ctx context.Context, clinicID, apptID uuid.UUID, req CancelRequest) error {
	appt, err := s.GetByID(ctx, clinicID, apptID)
	if err != nil {
		return err
	}

	if appt.Status == entappt.StatusCancelled {
		return ErrAlreadyCancelled
	}
	if appt.Status == entappt.StatusCompleted {
		return ErrAlreadyCompleted
	}

	now := time.Now()
	upd := s.db.Appointment.UpdateOne(appt).
		SetStatus(entappt.StatusCancelled).
		SetCancelledAt(now).
		SetCancellationFee(req.CancellationFee).
		SetCancelRequestedBy(entappt.CancelRequestedBy(req.RequestedBy))

	if req.Reason != nil {
		upd = upd.SetCancellationReason(*req.Reason)
	}

	if err := upd.Exec(ctx); err != nil {
		return fmt.Errorf("cancel appointment: %w", err)
	}

	// Restore slot to available if this appointment had a slot reference
	if appt.TimeSlotID != nil {
		_ = s.db.TimeSlot.Update().
			Where(
				entslot.ID(*appt.TimeSlotID),
				entslot.StatusEQ(entslot.StatusBooked),
			).
			SetStatus(entslot.StatusAvailable).
			Exec(ctx)
	}

	return nil
}

func (s *appointmentService) Complete(ctx context.Context, clinicID, therapistMemberID, apptID uuid.UUID) error {
	appt, err := s.GetByID(ctx, clinicID, apptID)
	if err != nil {
		return err
	}

	if appt.Status == entappt.StatusCompleted {
		return ErrAlreadyCompleted
	}
	if appt.Status == entappt.StatusCancelled {
		return ErrAlreadyCancelled
	}

	now := time.Now()
	return s.db.Appointment.UpdateOne(appt).
		SetStatus(entappt.StatusCompleted).
		SetCompletedAt(now).
		Exec(ctx)
}
