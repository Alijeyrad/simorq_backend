package app

import (
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
	"github.com/Alijeyrad/simorq_backend/internal/service/appointment"
	"github.com/Alijeyrad/simorq_backend/internal/service/auth"
	"github.com/Alijeyrad/simorq_backend/internal/service/clinic"
	"github.com/Alijeyrad/simorq_backend/internal/service/contact"
	"github.com/Alijeyrad/simorq_backend/internal/service/conversation"
	svcfile "github.com/Alijeyrad/simorq_backend/internal/service/file"
	"github.com/Alijeyrad/simorq_backend/internal/service/intern"
	"github.com/Alijeyrad/simorq_backend/internal/service/notification"
	"github.com/Alijeyrad/simorq_backend/internal/service/patient"
	"github.com/Alijeyrad/simorq_backend/internal/service/payment"
	"github.com/Alijeyrad/simorq_backend/internal/service/psychtest"
	"github.com/Alijeyrad/simorq_backend/internal/service/scheduling"
	"github.com/Alijeyrad/simorq_backend/internal/service/ticket"
	"github.com/Alijeyrad/simorq_backend/internal/service/user"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/Alijeyrad/simorq_backend/pkg/email"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
	s3pkg "github.com/Alijeyrad/simorq_backend/pkg/s3"
	"github.com/Alijeyrad/simorq_backend/pkg/sms"
	zarinpalpkg "github.com/Alijeyrad/simorq_backend/pkg/zarinpal"
)

// ServiceModule provides all application service dependencies.
var ServiceModule = fx.Module("services",
	fx.Provide(
		ProvideUserService,
		ProvideAuthService,
		ProvideClinicService,
		ProvidePatientService,
		ProvideFileService,
		ProvidePsychTestService,
		ProvideSchedulingService,
		ProvideAppointmentService,
		ProvidePaymentService,
		ProvideConversationService,
		ProvideTicketService,
		ProvideNotificationService,
		ProvideContactService,
		ProvideInternService,
		ProvidePasetoManager,
	),
)

func ProvideUserService(client *repo.Client, emailClient *email.Client, cfg *config.Config, authz authorize.IAuthorization) user.Service {
	return user.New(client, emailClient, cfg, authz)
}

func ProvideAuthService(
	db *repo.Client,
	rdb *redis.Client,
	smsCli *sms.Client,
	paseto *pasetotoken.Manager,
	cfg *config.Config,
) (auth.Service, error) {
	return auth.New(db, rdb, smsCli, paseto, cfg)
}

func ProvideClinicService(db *repo.Client, authz authorize.IAuthorization) clinic.Service {
	return clinic.New(db, authz)
}

func ProvidePatientService(db *repo.Client) patient.Service {
	return patient.New(db)
}

func ProvideFileService(db *repo.Client, s3 *s3pkg.Client) svcfile.Service {
	return svcfile.New(db, s3)
}

func ProvidePsychTestService(db *repo.Client) psychtest.Service {
	return psychtest.New(db)
}

func ProvideSchedulingService(db *repo.Client) scheduling.Service {
	return scheduling.New(db)
}

func ProvideAppointmentService(db *repo.Client, nc *nats.Conn) appointment.Service {
	return appointment.New(db, nc)
}

func ProvidePaymentService(db *repo.Client, zp *zarinpalpkg.Client, cfg *config.Config, nc *nats.Conn) payment.Service {
	return payment.New(db, zp, cfg, nc)
}

func ProvideConversationService(db *repo.Client, nc *nats.Conn) conversation.Service {
	return conversation.New(db, nc)
}

func ProvideTicketService(db *repo.Client, nc *nats.Conn) ticket.Service {
	return ticket.New(db, nc)
}

func ProvideNotificationService(db *repo.Client) notification.Service {
	return notification.New(db)
}

func ProvideContactService(db *repo.Client) contact.Service {
	return contact.New(db)
}

func ProvideInternService(db *repo.Client) intern.Service {
	return intern.New(db)
}

func ProvidePasetoManager(cfg *config.Config) (*pasetotoken.Manager, error) {
	return pasetotoken.NewPasetoManager(cfg)
}
