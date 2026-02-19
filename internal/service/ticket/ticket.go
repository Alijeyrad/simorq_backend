package ticket

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entticket "github.com/Alijeyrad/simorq_backend/internal/repo/ticket"
	enttm "github.com/Alijeyrad/simorq_backend/internal/repo/ticketmessage"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type ListRequest struct {
	Status  *string
	Page    int
	PerPage int
}

type CreateRequest struct {
	UserID   uuid.UUID
	ClinicID *uuid.UUID
	Subject  string
	Priority string
	Content  string // first message body
}

type UpdateStatusRequest struct {
	Status string // open | answered | closed
}

type ReplyRequest struct {
	SenderID uuid.UUID
	Content  string
	FileKey  *string
	FileName *string
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type Service interface {
	List(ctx context.Context, userID uuid.UUID, req ListRequest) ([]*repo.Ticket, error)
	GetByID(ctx context.Context, ticketID, userID uuid.UUID) (*repo.Ticket, error)
	Create(ctx context.Context, req CreateRequest) (*repo.Ticket, error)
	UpdateStatus(ctx context.Context, ticketID uuid.UUID, req UpdateStatusRequest) error
	ListMessages(ctx context.Context, ticketID, userID uuid.UUID, page, perPage int) ([]*repo.TicketMessage, error)
	Reply(ctx context.Context, ticketID uuid.UUID, req ReplyRequest) (*repo.TicketMessage, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type ticketService struct {
	db *repo.Client
	nc *nats.Conn
}

func New(db *repo.Client, nc *nats.Conn) Service {
	return &ticketService{db: db, nc: nc}
}

func (s *ticketService) List(ctx context.Context, userID uuid.UUID, req ListRequest) ([]*repo.Ticket, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}
	offset := (req.Page - 1) * req.PerPage

	q := s.db.Ticket.Query().
		Where(entticket.UserID(userID))

	if req.Status != nil {
		q = q.Where(entticket.StatusEQ(entticket.Status(*req.Status)))
	}

	tickets, err := q.
		Order(entticket.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(req.PerPage).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	return tickets, nil
}

func (s *ticketService) GetByID(ctx context.Context, ticketID, userID uuid.UUID) (*repo.Ticket, error) {
	t, err := s.db.Ticket.Query().
		Where(entticket.ID(ticketID), entticket.UserID(userID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}
	return t, nil
}

func (s *ticketService) Create(ctx context.Context, req CreateRequest) (*repo.Ticket, error) {
	priority := entticket.PriorityNormal
	if req.Priority != "" {
		priority = entticket.Priority(req.Priority)
	}

	c := s.db.Ticket.Create().
		SetUserID(req.UserID).
		SetSubject(req.Subject).
		SetPriority(priority)

	if req.ClinicID != nil {
		c = c.SetClinicID(*req.ClinicID)
	}

	t, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	// Create the first message
	_, err = s.db.TicketMessage.Create().
		SetTicketID(t.ID).
		SetSenderID(req.UserID).
		SetContent(req.Content).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create initial ticket message: %w", err)
	}

	return t, nil
}

func (s *ticketService) UpdateStatus(ctx context.Context, ticketID uuid.UUID, req UpdateStatusRequest) error {
	return s.db.Ticket.UpdateOneID(ticketID).
		SetStatus(entticket.Status(req.Status)).
		Exec(ctx)
}

func (s *ticketService) ListMessages(ctx context.Context, ticketID, userID uuid.UUID, page, perPage int) ([]*repo.TicketMessage, error) {
	// Verify access
	_, err := s.GetByID(ctx, ticketID, userID)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	msgs, err := s.db.TicketMessage.Query().
		Where(enttm.TicketID(ticketID)).
		Order(enttm.ByCreatedAt(sql.OrderAsc())).
		Offset(offset).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list ticket messages: %w", err)
	}
	return msgs, nil
}

func (s *ticketService) Reply(ctx context.Context, ticketID uuid.UUID, req ReplyRequest) (*repo.TicketMessage, error) {
	c := s.db.TicketMessage.Create().
		SetTicketID(ticketID).
		SetSenderID(req.SenderID).
		SetContent(req.Content)

	if req.FileKey != nil {
		c = c.SetFileKey(*req.FileKey)
	}
	if req.FileName != nil {
		c = c.SetFileName(*req.FileName)
	}

	msg, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create ticket reply: %w", err)
	}

	// Mark ticket as answered
	_ = s.db.Ticket.UpdateOneID(ticketID).
		SetStatus(entticket.StatusAnswered).
		Exec(ctx)

	// Publish NATS event
	if s.nc != nil {
		subject := fmt.Sprintf("simorgh.ticket.replied.%s", ticketID.String())
		_ = s.nc.Publish(subject, []byte(msg.ID.String()))
	}

	return msg, nil
}
