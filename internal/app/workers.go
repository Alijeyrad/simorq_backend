package app

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/fx"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entappt "github.com/Alijeyrad/simorq_backend/internal/repo/appointment"
	entconv "github.com/Alijeyrad/simorq_backend/internal/repo/conversation"
	entmsg "github.com/Alijeyrad/simorq_backend/internal/repo/message"
	entticket "github.com/Alijeyrad/simorq_backend/internal/repo/ticket"
	"github.com/Alijeyrad/simorq_backend/internal/service/notification"
	svcsms "github.com/Alijeyrad/simorq_backend/pkg/sms"
)

// WorkerModule registers all NATS event workers.
var WorkerModule = fx.Module("workers",
	fx.Invoke(RegisterWorkers),
)

type WorkerParams struct {
	fx.In

	Lc       fx.Lifecycle
	NC       *nats.Conn
	DB       *repo.Client
	NotifSvc notification.Service
	SMS      *svcsms.Client
}

func RegisterWorkers(p WorkerParams) {
	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			startNotificationWorker(p.NC, p.DB, p.NotifSvc)
			startSMSWorker(p.NC, p.DB, p.SMS)
			startWalletWorker(p.NC, p.DB)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Drain handled by ProvideNatsClient
			return nil
		},
	})
}

// ---------------------------------------------------------------------------
// notification_worker
// ---------------------------------------------------------------------------

func startNotificationWorker(nc *nats.Conn, db *repo.Client, notifSvc notification.Service) {
	// New message notifications
	_, err := nc.Subscribe("simorgh.message.new.*", func(msg *nats.Msg) {
		parts := strings.Split(msg.Subject, ".")
		if len(parts) < 4 {
			return
		}
		convIDStr := parts[3]
		convID, err := uuid.Parse(convIDStr)
		if err != nil {
			return
		}

		msgIDStr := strings.TrimSpace(string(msg.Data))
		msgID, err := uuid.Parse(msgIDStr)
		if err != nil {
			return
		}

		ctx := context.Background()

		conv, err := db.Conversation.Query().
			Where(entconv.ID(convID)).
			Only(ctx)
		if err != nil {
			slog.Warn("notification_worker: conversation not found", "id", convIDStr, "err", err)
			return
		}

		message, err := db.Message.Query().
			Where(entmsg.ID(msgID)).
			Only(ctx)
		if err != nil {
			slog.Warn("notification_worker: message not found", "id", msgIDStr, "err", err)
			return
		}

		recipientID := conv.ParticipantA
		if conv.ParticipantA == message.SenderID {
			recipientID = conv.ParticipantB
		}

		_, err = notifSvc.Create(ctx, notification.CreateRequest{
			UserID: recipientID,
			Type:   "message_new",
			Title:  "پیام جدید",
			Data:   map[string]any{"conversation_id": conv.ID.String()},
		})
		if err != nil {
			slog.Warn("notification_worker: create notification failed", "err", err)
		}
	})
	if err != nil {
		slog.Error("notification_worker: subscribe message.new failed", "err", err)
	}

	// Ticket reply notifications
	_, err = nc.Subscribe("simorgh.ticket.replied.*", func(msg *nats.Msg) {
		parts := strings.Split(msg.Subject, ".")
		if len(parts) < 4 {
			return
		}
		ticketIDStr := parts[3]
		ticketID, err := uuid.Parse(ticketIDStr)
		if err != nil {
			return
		}

		ctx := context.Background()

		ticket, err := db.Ticket.Query().
			Where(entticket.ID(ticketID)).
			Only(ctx)
		if err != nil {
			slog.Warn("notification_worker: ticket not found", "id", ticketIDStr, "err", err)
			return
		}

		_, err = notifSvc.Create(ctx, notification.CreateRequest{
			UserID: ticket.UserID,
			Type:   "ticket_replied",
			Title:  "پاسخ به تیکت",
			Data:   map[string]any{"ticket_id": ticket.ID.String()},
		})
		if err != nil {
			slog.Warn("notification_worker: create ticket notification failed", "err", err)
		}
	})
	if err != nil {
		slog.Error("notification_worker: subscribe ticket.replied failed", "err", err)
	}

	// Appointment created notifications
	_, err = nc.Subscribe("simorgh.appointment.created.*", func(msg *nats.Msg) {
		apptIDStr := strings.TrimSpace(string(msg.Data))
		apptID, err := uuid.Parse(apptIDStr)
		if err != nil {
			return
		}

		ctx := context.Background()

		appt, err := db.Appointment.Query().
			Where(entappt.ID(apptID)).
			Only(ctx)
		if err != nil {
			slog.Warn("notification_worker: appointment not found", "id", apptIDStr, "err", err)
			return
		}

		_, err = notifSvc.Create(ctx, notification.CreateRequest{
			UserID: appt.PatientID,
			Type:   "appointment_created",
			Title:  "نوبت جدید ثبت شد",
			Data:   map[string]any{"appointment_id": appt.ID.String()},
		})
		if err != nil {
			slog.Warn("notification_worker: create appt notification failed", "err", err)
		}
	})
	if err != nil {
		slog.Error("notification_worker: subscribe appointment.created failed", "err", err)
	}

	slog.Info("notification_worker: started")
}

// ---------------------------------------------------------------------------
// sms_worker
// ---------------------------------------------------------------------------

func startSMSWorker(nc *nats.Conn, db *repo.Client, smsCli *svcsms.Client) {
	_, err := nc.Subscribe("simorgh.appointment.created.*", func(msg *nats.Msg) {
		apptIDStr := strings.TrimSpace(string(msg.Data))
		apptID, err := uuid.Parse(apptIDStr)
		if err != nil {
			return
		}
		ctx := context.Background()

		appt, err := db.Appointment.Query().
			Where(entappt.ID(apptID)).
			Only(ctx)
		if err != nil {
			slog.Warn("sms_worker: appointment not found", "id", apptIDStr, "err", err)
			return
		}

		_ = appt
		_ = smsCli
		// Full SMS send deferred: requires patient phone number lookup
		slog.Debug("sms_worker: appointment created", "appointment_id", apptIDStr)
	})
	if err != nil {
		slog.Error("sms_worker: subscribe appointment.created failed", "err", err)
	}

	_, err = nc.Subscribe("simorgh.appointment.cancelled.*", func(msg *nats.Msg) {
		apptIDStr := strings.TrimSpace(string(msg.Data))
		apptID, err := uuid.Parse(apptIDStr)
		if err != nil {
			return
		}
		ctx := context.Background()

		appt, err := db.Appointment.Query().
			Where(entappt.ID(apptID)).
			Only(ctx)
		if err != nil {
			slog.Warn("sms_worker: appointment not found", "id", apptIDStr, "err", err)
			return
		}

		_ = appt
		_ = smsCli
		slog.Debug("sms_worker: appointment cancelled", "appointment_id", apptIDStr)
	})
	if err != nil {
		slog.Error("sms_worker: subscribe appointment.cancelled failed", "err", err)
	}

	slog.Info("sms_worker: started")
}

// ---------------------------------------------------------------------------
// wallet_worker (commission splitting)
// ---------------------------------------------------------------------------

func startWalletWorker(nc *nats.Conn, db *repo.Client) {
	_, err := nc.Subscribe("simorgh.payment.received.*", func(msg *nats.Msg) {
		paymentIDStr := strings.TrimSpace(string(msg.Data))
		// Commission splitting logic deferred to Phase 5 finance module:
		// 1. Load PaymentRequest by id
		// 2. Load CommissionRule for the clinic
		// 3. Credit clinic wallet (session_price - platform_fee)
		// 4. Credit platform wallet (platform_fee)
		// 5. Create Transaction records for both
		_ = db
		slog.Debug("wallet_worker: payment received", "payment_id", paymentIDStr)
	})
	if err != nil {
		slog.Error("wallet_worker: subscribe payment.received failed", "err", err)
	}

	slog.Info("wallet_worker: started")
}
