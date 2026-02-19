package router

import (
	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/api/http/handler"
	"github.com/Alijeyrad/simorq_backend/internal/api/http/middleware"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
	"github.com/Alijeyrad/simorq_backend/internal/service/appointment"
	"github.com/Alijeyrad/simorq_backend/internal/service/auth"
	"github.com/Alijeyrad/simorq_backend/internal/service/clinic"
	"github.com/Alijeyrad/simorq_backend/internal/service/contact"
	"github.com/Alijeyrad/simorq_backend/internal/service/conversation"
	"github.com/Alijeyrad/simorq_backend/internal/service/file"
	"github.com/Alijeyrad/simorq_backend/internal/service/intern"
	"github.com/Alijeyrad/simorq_backend/internal/service/notification"
	"github.com/Alijeyrad/simorq_backend/internal/service/patient"
	"github.com/Alijeyrad/simorq_backend/internal/service/payment"
	"github.com/Alijeyrad/simorq_backend/internal/service/psychtest"
	"github.com/Alijeyrad/simorq_backend/internal/service/scheduling"
	"github.com/Alijeyrad/simorq_backend/internal/service/ticket"
	"github.com/Alijeyrad/simorq_backend/internal/service/user"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

// Module provides the Router to the fx graph.
var Module = fx.Module("router", fx.Provide(NewRouter))

type Params struct {
	fx.In

	Cfg             *config.Config
	Redis           *redis.Client
	Auth            authorize.IAuthorization
	DB              *repo.Client
	UserSvc         user.Service
	AuthSvc         auth.Service
	ClinicSvc       clinic.Service
	PatientSvc      patient.Service
	FileSvc         file.Service
	PsychTestSvc    psychtest.Service
	SchedulingSvc   scheduling.Service
	AppointmentSvc  appointment.Service
	PaymentSvc      payment.Service
	ConversationSvc conversation.Service
	TicketSvc       ticket.Service
	NotificationSvc notification.Service
	ContactSvc      contact.Service
	InternSvc       intern.Service
	PasetoMgr       *pasetotoken.Manager
}

type Router struct {
	p Params
}

func NewRouter(p Params) *Router {
	return &Router{p: p}
}

func (r *Router) Register(app *fiber.App) {
	// 1. Health & Metrics
	r.registerSystemRoutes(app)

	// 2. Initialize Middlewares
	authRequired := middleware.AuthRequired(r.p.PasetoMgr, r.p.Redis)
	clinicCtx := middleware.ClinicContext(r.p.DB)
	clinicHeader := middleware.ClinicHeader(r.p.DB)

	// Permission helper
	requirePerm := func(res authorize.Resource, act authorize.Action) fiber.Handler {
		return middleware.RequirePermission(r.p.Auth, res, act)
	}

	// 3. Initialize Handlers
	authH := handler.NewAuthHandler(r.p.AuthSvc)
	userH := handler.NewUserHandler(r.p.UserSvc)
	clinicH := handler.NewClinicHandler(r.p.ClinicSvc)
	patientH := handler.NewPatientHandler(r.p.PatientSvc)
	fileH := handler.NewFileHandler(r.p.FileSvc)
	testH := handler.NewTestHandler(r.p.PsychTestSvc)
	scheduleH := handler.NewScheduleHandler(r.p.SchedulingSvc)
	appointmentH := handler.NewAppointmentHandler(r.p.AppointmentSvc)
	paymentH := handler.NewPaymentHandler(r.p.PaymentSvc)
	conversationH := handler.NewConversationHandler(r.p.ConversationSvc)
	ticketH := handler.NewTicketHandler(r.p.TicketSvc)
	notificationH := handler.NewNotificationHandler(r.p.NotificationSvc)
	contactH := handler.NewContactHandler(r.p.ContactSvc)
	internH := handler.NewInternHandler(r.p.InternSvc)

	api := app.Group("/api/v1")

	// 4. Delegate to sub-files
	r.registerAuthRoutes(api, authH, authRequired)
	r.registerUserRoutes(api, userH, authRequired)
	r.registerClinicRoutes(api, clinicH, authRequired, clinicCtx, requirePerm)
	r.registerPatientRoutes(api, patientH, fileH, authRequired, clinicHeader, requirePerm)
	r.registerFileRoutes(api, fileH, authRequired, clinicHeader)
	r.registerTestRoutes(api, testH, authRequired)
	r.registerScheduleRoutes(api, scheduleH, authRequired, clinicHeader, requirePerm)
	r.registerAppointmentRoutes(api, appointmentH, authRequired, clinicHeader, requirePerm)
	r.registerPaymentRoutes(api, paymentH, authRequired, clinicHeader)
	r.registerConversationRoutes(api, conversationH, authRequired, clinicHeader, requirePerm)
	r.registerTicketRoutes(api, ticketH, authRequired)
	r.registerNotificationRoutes(api, notificationH, authRequired)
	r.registerContactRoutes(api, contactH)
	r.registerInternRoutes(api, internH, authRequired, clinicHeader, requirePerm)
}

func (r *Router) registerSystemRoutes(app *fiber.App) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool { return authorize.IsPolicyHealthy() },
	}))
	app.Get(healthcheck.StartupEndpoint, healthcheck.New())

	if r.p.Cfg.Observability.Enabled && r.p.Cfg.Observability.Metrics.Enabled {
		path := r.p.Cfg.Observability.Metrics.Path
		if path == "" {
			path = "/metrics"
		}
		app.Get(path, adaptor.HTTPHandler(promhttp.Handler()))
	}
}
