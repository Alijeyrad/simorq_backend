package psychtest

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entpsych "github.com/Alijeyrad/simorq_backend/internal/repo/psychtest"
)

var ErrNotFound = errors.New("psych test not found")

// ---------------------------------------------------------------------------
// Service interface
// ---------------------------------------------------------------------------

type Service interface {
	List(ctx context.Context) ([]*repo.PsychTest, error)
	GetByID(ctx context.Context, id uuid.UUID) (*repo.PsychTest, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type service struct {
	db *repo.Client
}

func New(db *repo.Client) Service {
	return &service{db: db}
}

func (s *service) List(ctx context.Context) ([]*repo.PsychTest, error) {
	return s.db.PsychTest.Query().
		Where(entpsych.IsActive(true)).
		Order(entpsych.ByName()).
		All(ctx)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*repo.PsychTest, error) {
	t, err := s.db.PsychTest.Query().
		Where(entpsych.ID(id), entpsych.IsActive(true)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get psych test: %w", err)
	}
	return t, nil
}
