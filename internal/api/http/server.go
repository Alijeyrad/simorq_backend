package http

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/api/http/middleware"

	"github.com/Alijeyrad/simorq_backend/internal/repo"
	"github.com/Alijeyrad/simorq_backend/internal/service/user"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/Alijeyrad/simorq_backend/pkg/observability"
)

// Module provides the HTTP server to the fx dependency graph.
var Module = fx.Module("http", fx.Provide(NewServer))

// Params holds all dependencies for the HTTP server.
type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Cfg       *config.Config
	Redis     *redis.Client
	Auth      authorize.IAuthorization
	OTel      *observability.Provider `optional:"true"`
	DB        *repo.Client

	// Services
	UserSvc user.Service
}

type Server struct {
	app     *fiber.App
	cfg     *config.Config
	rdb     *redis.Client
	auth    authorize.IAuthorization
	db      *repo.Client
	userSvc user.Service
}

func NewServer(p Params) *Server {
	app := fiber.New()

	// OTel tracing middleware (before other middleware)
	if p.OTel != nil && p.Cfg.Observability.Tracing.Enabled {
		app.Use(observability.FiberMiddleware(p.Cfg.Observability.ServiceName))
	}

	configureMiddleware(app, p.Cfg, p.Redis)

	s := &Server{
		app:     app,
		cfg:     p.Cfg,
		rdb:     p.Redis,
		auth:    p.Auth,
		db:      p.DB,
		userSvc: p.UserSvc,
	}

	s.registerRoutes()

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			port := s.cfg.Server.Port
			if port == 0 {
				return fmt.Errorf("server port is not configured (got 0)")
			}
			go func() {
				addr := fmt.Sprintf(":%d", port)
				if err := s.app.Listen(addr); err != nil {
					slog.Error("HTTP server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Debug("shutting down Fiber application")
			return s.app.ShutdownWithContext(ctx)
		},
	})

	return s
}

func (s *Server) registerRoutes() {
	s.app.Get(healthcheck.LivenessEndpoint, healthcheck.New()) // /livesz
	s.app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool {
			return authorize.IsPolicyHealthy()
		},
	})) // /readyz
	s.app.Get(healthcheck.StartupEndpoint, healthcheck.New()) // /startupz

	// Prometheus metrics endpoint
	if s.cfg.Observability.Enabled && s.cfg.Observability.Metrics.Enabled {
		metricsPath := s.cfg.Observability.Metrics.Path
		if metricsPath == "" {
			metricsPath = "/metrics"
		}
		s.app.Get(metricsPath, adaptor.HTTPHandler(promhttp.Handler()))
		slog.Info("metrics endpoint registered", "path", metricsPath)
	}

	s.app.Group("/api/v1")

}

func configureMiddleware(app *fiber.App, cfg *config.Config, rdb *redis.Client) {
	// Request ID must be first
	app.Use(middleware.RequestID())

	app.Use(recoverer.New())

	if cfg.Server.Environment == "production" {
		headersCfg := cfg.Server.Headers
		app.Use(helmet.New(helmet.Config{
			XSSProtection:             headersCfg.XSSProtection,
			ContentTypeNosniff:        headersCfg.ContentTypeNosniff,
			XFrameOptions:             headersCfg.XFrameOptions,
			ReferrerPolicy:            headersCfg.ReferrerPolicy,
			CrossOriginEmbedderPolicy: headersCfg.CrossOriginEmbedderPolicy,
			CrossOriginOpenerPolicy:   headersCfg.CrossOriginOpenerPolicy,
			CrossOriginResourcePolicy: headersCfg.CrossOriginResourcePolicy,
			OriginAgentCluster:        headersCfg.OriginAgentCluster,
			XDNSPrefetchControl:       headersCfg.XDNSPrefetchControl,
			XDownloadOptions:          headersCfg.XDownloadOptions,
			XPermittedCrossDomain:     headersCfg.XPermittedCrossDomain,
		}))

		if cfg.Server.CORS.Enabled {
			origins := cfg.Server.CORS.AllowOrigins
			app.Use(cors.New(cors.Config{
				AllowOrigins:     origins,
				AllowMethods:     cfg.Server.CORS.AllowMethods,
				AllowHeaders:     cfg.Server.CORS.AllowHeaders,
				ExposeHeaders:    cfg.Server.CORS.ExposeHeaders,
				AllowCredentials: cfg.Server.CORS.AllowCredentials,
				MaxAge:           cfg.Server.CORS.MaxAgeSeconds,
			}))
		}

		app.Use(middleware.NewLimiterWithRedis(rdb))
	}

	app.Use(logger.New(logger.Config{
		CustomTags: map[string]logger.LogFunc{
			"requestId": func(output logger.Buffer, c fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				if rid, ok := middleware.RequestIDFromFiber(c); ok {
					return output.WriteString(rid)
				}
				return 0, nil
			},
		},
		Format: "${ip} - - [${time}] [req_id=${requestId}] \"${method} ${url} ${protocol}\" ${status} ${bytesSent} \"${referer}\" \"${ua}\"\n",
	}))
}
