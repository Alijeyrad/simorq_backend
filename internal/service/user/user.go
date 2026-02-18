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

type UpdateMeRequest struct {
	FirstName     *string
	LastName      *string
	Gender        *string
	MaritalStatus *string
	BirthYear     *int
}

type Service interface {
	GetByID(ctx context.Context, id string) (*repo.User, error)
	GetMe(ctx context.Context, userID uuid.UUID) (*repo.User, error)
	UpdateMe(ctx context.Context, userID uuid.UUID, req UpdateMeRequest) (*repo.User, error)
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

func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (*repo.User, error) {
	u, err := s.client.User.Query().
		Where(
			user.ID(userID),
			user.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get me: %w", err)
	}
	return u, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID uuid.UUID, req UpdateMeRequest) (*repo.User, error) {
	u, err := s.GetMe(ctx, userID)
	if err != nil {
		return nil, err
	}

	upd := s.client.User.UpdateOne(u)
	if req.FirstName != nil {
		upd = upd.SetFirstName(*req.FirstName)
	}
	if req.LastName != nil {
		upd = upd.SetLastName(*req.LastName)
	}
	if req.Gender != nil {
		upd = upd.SetNillableGender(req.Gender)
	}
	if req.MaritalStatus != nil {
		upd = upd.SetNillableMaritalStatus(req.MaritalStatus)
	}
	if req.BirthYear != nil {
		upd = upd.SetNillableBirthYear(req.BirthYear)
	}

	result, err := upd.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update me: %w", err)
	}
	return result, nil
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
