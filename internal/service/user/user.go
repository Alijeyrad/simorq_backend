package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
	"github.com/Alijeyrad/simorq_backend/internal/repo/user"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/Alijeyrad/simorq_backend/pkg/email"
)

type Service interface {
	GetByID(ctx context.Context, id string) (*repo.User, error)
}

type UserService struct {
	client      *repo.Client
	emailClient *email.Client
	cfg         *config.Config
	authorize   authorize.IAuthorization
}

func New(client *repo.Client, emailClient *email.Client, cfg *config.Config, authz authorize.IAuthorization) *UserService {
	return &UserService{
		client:      client,
		emailClient: emailClient,
		cfg:         cfg,
		authorize:   authz,
	}
}

// GetByID retrieves a user by ID with their profile, excluding soft-deleted users
func (s *UserService) GetByID(ctx context.Context, id string) (*repo.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	u, err := s.client.User.Query().
		Where(
			user.ID(uid),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return u, nil
}
