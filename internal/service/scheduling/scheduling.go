package scheduling

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entrecrule "github.com/Alijeyrad/simorq_backend/internal/repo/recurringrule"
	entprofile "github.com/Alijeyrad/simorq_backend/internal/repo/therapistprofile"
	entslot "github.com/Alijeyrad/simorq_backend/internal/repo/timeslot"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateSlotRequest struct {
	StartTime       time.Time
	EndTime         time.Time
	SessionPrice    *int64
	ReservationFee  *int64
	IsRecurring     bool
	RecurringRuleID *uuid.UUID
}

type CreateRecurringRuleRequest struct {
	DayOfWeek      int8
	StartHour      int8
	StartMinute    int8
	EndHour        int8
	EndMinute      int8
	SessionPrice   *int64
	ReservationFee *int64
	ValidFrom      time.Time
	ValidUntil     *time.Time
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type Service interface {
	// Slot management
	ListSlots(ctx context.Context, clinicID, therapistMemberID uuid.UUID, from, to time.Time) ([]*repo.TimeSlot, error)
	CreateSlot(ctx context.Context, clinicID, therapistMemberID uuid.UUID, req CreateSlotRequest) (*repo.TimeSlot, error)
	DeleteSlot(ctx context.Context, clinicID, therapistMemberID, slotID uuid.UUID) error

	// Recurring rule management
	ListRecurringRules(ctx context.Context, clinicID, therapistMemberID uuid.UUID) ([]*repo.RecurringRule, error)
	CreateRecurringRule(ctx context.Context, clinicID, therapistMemberID uuid.UUID, req CreateRecurringRuleRequest) (*repo.RecurringRule, error)
	DeleteRecurringRule(ctx context.Context, clinicID, therapistMemberID, ruleID uuid.UUID) error

	// Schedule toggle (updates TherapistProfile.is_accepting)
	ToggleSchedule(ctx context.Context, clinicID, therapistMemberID uuid.UUID, enabled bool) error

	// Public — no auth required
	ListPublicSlots(ctx context.Context, therapistMemberID uuid.UUID, from, to time.Time) ([]*repo.TimeSlot, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type schedulingService struct {
	db *repo.Client
}

func New(db *repo.Client) Service {
	return &schedulingService{db: db}
}

// ---------------------------------------------------------------------------
// Slots
// ---------------------------------------------------------------------------

func (s *schedulingService) ListSlots(ctx context.Context, clinicID, therapistMemberID uuid.UUID, from, to time.Time) ([]*repo.TimeSlot, error) {
	slots, err := s.db.TimeSlot.Query().
		Where(
			entslot.ClinicID(clinicID),
			entslot.TherapistID(therapistMemberID),
			entslot.StartTimeGTE(from),
			entslot.StartTimeLT(to),
		).
		Order(entslot.ByStartTime()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list slots: %w", err)
	}
	return slots, nil
}

func (s *schedulingService) CreateSlot(ctx context.Context, clinicID, therapistMemberID uuid.UUID, req CreateSlotRequest) (*repo.TimeSlot, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, ErrInvalidTimeRange
	}

	// Overlap check: existing non-cancelled slots for this therapist that overlap
	overlaps, err := s.db.TimeSlot.Query().
		Where(
			entslot.TherapistID(therapistMemberID),
			entslot.StatusNotIn(entslot.StatusCancelled),
			entslot.StartTimeLT(req.EndTime),
			entslot.EndTimeGTE(req.StartTime),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check overlap: %w", err)
	}
	if overlaps {
		return nil, ErrOverlappingSlot
	}

	c := s.db.TimeSlot.Create().
		SetClinicID(clinicID).
		SetTherapistID(therapistMemberID).
		SetStartTime(req.StartTime).
		SetEndTime(req.EndTime).
		SetIsRecurring(req.IsRecurring)

	if req.SessionPrice != nil {
		c = c.SetSessionPrice(*req.SessionPrice)
	}
	if req.ReservationFee != nil {
		c = c.SetReservationFee(*req.ReservationFee)
	}
	if req.RecurringRuleID != nil {
		c = c.SetRecurringRuleID(*req.RecurringRuleID)
	}

	slot, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create slot: %w", err)
	}
	return slot, nil
}

func (s *schedulingService) DeleteSlot(ctx context.Context, clinicID, therapistMemberID, slotID uuid.UUID) error {
	slot, err := s.db.TimeSlot.Query().
		Where(entslot.ID(slotID), entslot.ClinicID(clinicID), entslot.TherapistID(therapistMemberID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return ErrSlotNotFound
		}
		return fmt.Errorf("get slot: %w", err)
	}
	if slot.Status == entslot.StatusBooked {
		return ErrSlotAlreadyBooked
	}
	return s.db.TimeSlot.DeleteOne(slot).Exec(ctx)
}

// ---------------------------------------------------------------------------
// Recurring rules
// ---------------------------------------------------------------------------

func (s *schedulingService) ListRecurringRules(ctx context.Context, clinicID, therapistMemberID uuid.UUID) ([]*repo.RecurringRule, error) {
	rules, err := s.db.RecurringRule.Query().
		Where(
			entrecrule.ClinicID(clinicID),
			entrecrule.TherapistID(therapistMemberID),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recurring rules: %w", err)
	}
	return rules, nil
}

func (s *schedulingService) CreateRecurringRule(ctx context.Context, clinicID, therapistMemberID uuid.UUID, req CreateRecurringRuleRequest) (*repo.RecurringRule, error) {
	c := s.db.RecurringRule.Create().
		SetClinicID(clinicID).
		SetTherapistID(therapistMemberID).
		SetDayOfWeek(req.DayOfWeek).
		SetStartHour(req.StartHour).
		SetStartMinute(req.StartMinute).
		SetEndHour(req.EndHour).
		SetEndMinute(req.EndMinute).
		SetValidFrom(req.ValidFrom)

	if req.SessionPrice != nil {
		c = c.SetSessionPrice(*req.SessionPrice)
	}
	if req.ReservationFee != nil {
		c = c.SetReservationFee(*req.ReservationFee)
	}
	if req.ValidUntil != nil {
		c = c.SetValidUntil(*req.ValidUntil)
	}

	rule, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create recurring rule: %w", err)
	}
	return rule, nil
}

func (s *schedulingService) DeleteRecurringRule(ctx context.Context, clinicID, therapistMemberID, ruleID uuid.UUID) error {
	rule, err := s.db.RecurringRule.Query().
		Where(entrecrule.ID(ruleID), entrecrule.ClinicID(clinicID), entrecrule.TherapistID(therapistMemberID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return ErrRuleNotFound
		}
		return fmt.Errorf("get recurring rule: %w", err)
	}
	return s.db.RecurringRule.DeleteOne(rule).Exec(ctx)
}

// ---------------------------------------------------------------------------
// Schedule toggle
// ---------------------------------------------------------------------------

func (s *schedulingService) ToggleSchedule(ctx context.Context, clinicID, therapistMemberID uuid.UUID, enabled bool) error {
	profile, err := s.db.TherapistProfile.Query().
		Where(entprofile.ClinicMemberID(therapistMemberID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return fmt.Errorf("therapist profile not found for member %s", therapistMemberID)
		}
		return fmt.Errorf("get therapist profile: %w", err)
	}

	return s.db.TherapistProfile.UpdateOne(profile).
		SetIsAccepting(enabled).
		Exec(ctx)
}

// ---------------------------------------------------------------------------
// Public listing
// ---------------------------------------------------------------------------

func (s *schedulingService) ListPublicSlots(ctx context.Context, therapistMemberID uuid.UUID, from, to time.Time) ([]*repo.TimeSlot, error) {
	slots, err := s.db.TimeSlot.Query().
		Where(
			entslot.TherapistID(therapistMemberID),
			entslot.StatusEQ(entslot.StatusAvailable),
			entslot.StartTimeGTE(from),
			entslot.StartTimeLT(to),
		).
		Order(entslot.ByStartTime()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list public slots: %w", err)
	}
	return slots, nil
}
