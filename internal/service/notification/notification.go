package notification

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entnotif "github.com/Alijeyrad/simorq_backend/internal/repo/notification"
	entpref "github.com/Alijeyrad/simorq_backend/internal/repo/notificationpref"
	entdevice "github.com/Alijeyrad/simorq_backend/internal/repo/userdevice"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateRequest struct {
	UserID uuid.UUID
	Type   string
	Title  string
	Body   *string
	Data   map[string]any
}

type UpsertPrefsRequest struct {
	AppointmentSMS  bool
	AppointmentPush bool
	MessagePush     bool
	TicketReplyPush bool
}

type RegisterDeviceRequest struct {
	UserID      uuid.UUID
	DeviceToken string
	Platform    string // web | android | ios
}

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*repo.Notification, error)
	List(ctx context.Context, userID uuid.UUID, unreadOnly bool, page, perPage int) ([]*repo.Notification, error)
	MarkRead(ctx context.Context, notifID, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	GetPrefs(ctx context.Context, userID uuid.UUID) (*repo.NotificationPref, error)
	UpsertPrefs(ctx context.Context, userID uuid.UUID, req UpsertPrefsRequest) (*repo.NotificationPref, error)
	RegisterDevice(ctx context.Context, req RegisterDeviceRequest) (*repo.UserDevice, error)
	DeactivateDevice(ctx context.Context, userID uuid.UUID, deviceToken string) error
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type notificationService struct {
	db *repo.Client
}

func New(db *repo.Client) Service {
	return &notificationService{db: db}
}

func (s *notificationService) Create(ctx context.Context, req CreateRequest) (*repo.Notification, error) {
	c := s.db.Notification.Create().
		SetUserID(req.UserID).
		SetType(req.Type).
		SetTitle(req.Title)

	if req.Body != nil {
		c = c.SetBody(*req.Body)
	}
	if req.Data != nil {
		c = c.SetData(req.Data)
	}

	n, err := c.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return n, nil
}

func (s *notificationService) List(ctx context.Context, userID uuid.UUID, unreadOnly bool, page, perPage int) ([]*repo.Notification, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	q := s.db.Notification.Query().
		Where(entnotif.UserID(userID))

	if unreadOnly {
		q = q.Where(entnotif.IsRead(false))
	}

	notifs, err := q.
		Order(entnotif.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return notifs, nil
}

func (s *notificationService) MarkRead(ctx context.Context, notifID, userID uuid.UUID) error {
	n, err := s.db.Notification.Query().
		Where(entnotif.ID(notifID), entnotif.UserID(userID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("get notification: %w", err)
	}

	return s.db.Notification.UpdateOne(n).
		SetIsRead(true).
		Exec(ctx)
}

func (s *notificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.db.Notification.Update().
		Where(entnotif.UserID(userID), entnotif.IsRead(false)).
		SetIsRead(true).
		Exec(ctx)
}

func (s *notificationService) GetPrefs(ctx context.Context, userID uuid.UUID) (*repo.NotificationPref, error) {
	pref, err := s.db.NotificationPref.Query().
		Where(entpref.UserID(userID)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			// Return defaults without persisting
			return &repo.NotificationPref{
				UserID:          userID,
				AppointmentSms:  true,
				AppointmentPush: true,
				MessagePush:     true,
				TicketReplyPush: true,
			}, nil
		}
		return nil, fmt.Errorf("get notification prefs: %w", err)
	}
	return pref, nil
}

func (s *notificationService) UpsertPrefs(ctx context.Context, userID uuid.UUID, req UpsertPrefsRequest) (*repo.NotificationPref, error) {
	existing, err := s.db.NotificationPref.Query().
		Where(entpref.UserID(userID)).
		Only(ctx)
	if err != nil {
		if !repo.IsNotFound(err) {
			return nil, fmt.Errorf("get notification prefs: %w", err)
		}
		// Not found — create
		pref, cErr := s.db.NotificationPref.Create().
			SetUserID(userID).
			SetAppointmentSms(req.AppointmentSMS).
			SetAppointmentPush(req.AppointmentPush).
			SetMessagePush(req.MessagePush).
			SetTicketReplyPush(req.TicketReplyPush).
			Save(ctx)
		if cErr != nil {
			return nil, fmt.Errorf("create notification prefs: %w", cErr)
		}
		return pref, nil
	}

	pref, err := s.db.NotificationPref.UpdateOne(existing).
		SetAppointmentSms(req.AppointmentSMS).
		SetAppointmentPush(req.AppointmentPush).
		SetMessagePush(req.MessagePush).
		SetTicketReplyPush(req.TicketReplyPush).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update notification prefs: %w", err)
	}
	return pref, nil
}

func (s *notificationService) RegisterDevice(ctx context.Context, req RegisterDeviceRequest) (*repo.UserDevice, error) {
	existing, err := s.db.UserDevice.Query().
		Where(
			entdevice.UserID(req.UserID),
			entdevice.DeviceToken(req.DeviceToken),
		).
		Only(ctx)
	if err == nil {
		d, uErr := s.db.UserDevice.UpdateOne(existing).
			SetPlatform(entdevice.Platform(req.Platform)).
			SetIsActive(true).
			Save(ctx)
		if uErr != nil {
			return nil, fmt.Errorf("update device: %w", uErr)
		}
		return d, nil
	}

	if !repo.IsNotFound(err) {
		return nil, fmt.Errorf("check device: %w", err)
	}

	d, err := s.db.UserDevice.Create().
		SetUserID(req.UserID).
		SetDeviceToken(req.DeviceToken).
		SetPlatform(entdevice.Platform(req.Platform)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("register device: %w", err)
	}
	return d, nil
}

func (s *notificationService) DeactivateDevice(ctx context.Context, userID uuid.UUID, deviceToken string) error {
	return s.db.UserDevice.Update().
		Where(
			entdevice.UserID(userID),
			entdevice.DeviceToken(deviceToken),
		).
		SetIsActive(false).
		Exec(ctx)
}
