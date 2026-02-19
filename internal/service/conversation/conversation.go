package conversation

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entconv "github.com/Alijeyrad/simorq_backend/internal/repo/conversation"
	entmsg "github.com/Alijeyrad/simorq_backend/internal/repo/message"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateRequest struct {
	ParticipantA uuid.UUID
	ParticipantB uuid.UUID
	PatientID    *uuid.UUID
}

type SendMessageRequest struct {
	SenderID uuid.UUID
	Content  *string
	FileKey  *string
	FileName *string
	FileMime *string
}

type ListMessagesRequest struct {
	Page    int
	PerPage int
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type Service interface {
	List(ctx context.Context, clinicID, userID uuid.UUID, page, perPage int) ([]*repo.Conversation, error)
	GetByID(ctx context.Context, clinicID, convID, userID uuid.UUID) (*repo.Conversation, error)
	Create(ctx context.Context, clinicID uuid.UUID, req CreateRequest) (*repo.Conversation, error)
	ListMessages(ctx context.Context, convID, userID uuid.UUID, req ListMessagesRequest) ([]*repo.Message, error)
	SendMessage(ctx context.Context, convID uuid.UUID, req SendMessageRequest) (*repo.Message, error)
	DeleteMessage(ctx context.Context, convID, messageID, userID uuid.UUID) error
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type conversationService struct {
	db *repo.Client
	nc *nats.Conn
}

func New(db *repo.Client, nc *nats.Conn) Service {
	return &conversationService{db: db, nc: nc}
}

func (s *conversationService) List(ctx context.Context, clinicID, userID uuid.UUID, page, perPage int) ([]*repo.Conversation, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	convs, err := s.db.Conversation.Query().
		Where(
			entconv.ClinicID(clinicID),
			entconv.Or(
				entconv.ParticipantA(userID),
				entconv.ParticipantB(userID),
			),
		).
		Order(entconv.ByLastMessageAt(sql.OrderDesc())).
		Offset(offset).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	return convs, nil
}

func (s *conversationService) GetByID(ctx context.Context, clinicID, convID, userID uuid.UUID) (*repo.Conversation, error) {
	conv, err := s.db.Conversation.Query().
		Where(
			entconv.ID(convID),
			entconv.ClinicID(clinicID),
			entconv.Or(
				entconv.ParticipantA(userID),
				entconv.ParticipantB(userID),
			),
		).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	return conv, nil
}

func (s *conversationService) Create(ctx context.Context, clinicID uuid.UUID, req CreateRequest) (*repo.Conversation, error) {
	// Check if conversation already exists
	exists, err := s.db.Conversation.Query().
		Where(
			entconv.ClinicID(clinicID),
			entconv.Or(
				entconv.And(
					entconv.ParticipantA(req.ParticipantA),
					entconv.ParticipantB(req.ParticipantB),
				),
				entconv.And(
					entconv.ParticipantA(req.ParticipantB),
					entconv.ParticipantB(req.ParticipantA),
				),
			),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("check existing conversation: %w", err)
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	c := s.db.Conversation.Create().
		SetClinicID(clinicID).
		SetParticipantA(req.ParticipantA).
		SetParticipantB(req.ParticipantB)

	if req.PatientID != nil {
		c = c.SetPatientID(*req.PatientID)
	}

	conv, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}
	return conv, nil
}

func (s *conversationService) ListMessages(ctx context.Context, convID, userID uuid.UUID, req ListMessagesRequest) ([]*repo.Message, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 50
	}
	offset := (req.Page - 1) * req.PerPage

	msgs, err := s.db.Message.Query().
		Where(
			entmsg.ConversationID(convID),
			entmsg.DeletedAtIsNil(),
		).
		Order(entmsg.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(req.PerPage).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return msgs, nil
}

func (s *conversationService) SendMessage(ctx context.Context, convID uuid.UUID, req SendMessageRequest) (*repo.Message, error) {
	c := s.db.Message.Create().
		SetConversationID(convID).
		SetSenderID(req.SenderID)

	if req.Content != nil {
		c = c.SetContent(*req.Content)
	}
	if req.FileKey != nil {
		c = c.SetFileKey(*req.FileKey)
	}
	if req.FileName != nil {
		c = c.SetFileName(*req.FileName)
	}
	if req.FileMime != nil {
		c = c.SetFileMime(*req.FileMime)
	}

	msg, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	// Update last_message_at on the conversation
	_ = s.db.Conversation.Update().
		Where(entconv.ID(convID)).
		SetLastMessageAt(msg.CreatedAt).
		Exec(ctx)

	// Publish NATS event
	if s.nc != nil {
		subject := fmt.Sprintf("simorgh.message.new.%s", convID.String())
		_ = s.nc.Publish(subject, []byte(msg.ID.String()))
	}

	return msg, nil
}

func (s *conversationService) DeleteMessage(ctx context.Context, convID, messageID, userID uuid.UUID) error {
	msg, err := s.db.Message.Query().
		Where(
			entmsg.ID(messageID),
			entmsg.ConversationID(convID),
			entmsg.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return ErrMessageNotFound
		}
		return fmt.Errorf("get message: %w", err)
	}

	if msg.SenderID != userID {
		return ErrUnauthorized
	}

	return s.db.Message.UpdateOne(msg).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
