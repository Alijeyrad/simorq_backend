package contact

import (
	"context"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateRequest struct {
	Name    string
	Email   string
	Subject string
	Message string
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type Service interface {
	Submit(ctx context.Context, req CreateRequest) (*repo.ContactMessage, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type contactService struct {
	db *repo.Client
}

func New(db *repo.Client) Service {
	return &contactService{db: db}
}

func (s *contactService) Submit(ctx context.Context, req CreateRequest) (*repo.ContactMessage, error) {
	msg, err := s.db.ContactMessage.Create().
		SetName(req.Name).
		SetEmail(req.Email).
		SetSubject(req.Subject).
		SetMessage(req.Message).
		Save(ctx)
	if err != nil {
		return nil, ErrInternal
	}
	return msg, nil
}
