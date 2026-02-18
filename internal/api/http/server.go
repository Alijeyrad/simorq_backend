package http

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/api/http/middleware"
	"github.com/Alijeyrad/simorq_backend/internal/api/http/router"
	"github.com/Alijeyrad/simorq_backend/pkg/observability"
)

// Module provides the HTTP Server to the fx graph.
var Module = fx.Module("http", fx.Provide(NewServer))

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Cfg       *config.Config
	Redis     *redis.Client
	Router    *router.Router
	OTel      *observability.Provider `optional:"true"`
}

func NewServer(p Params) *fiber.App {
	app := fiber.New()

	if p.OTel != nil && p.Cfg.Observability.Tracing.Enabled {
		app.Use(observability.FiberMiddleware(p.Cfg.Observability.ServiceName))
	}

	configureGlobalMiddleware(app, p.Cfg, p.Redis)

	p.Router.Register(app)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := fmt.Sprintf(":%d", p.Cfg.Server.Port)
			go func() {
				if err := app.Listen(addr); err != nil {
					slog.Error("HTTP server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.ShutdownWithContext(ctx)
		},
	})

	return app
}

func configureGlobalMiddleware(app *fiber.App, cfg *config.Config, rdb *redis.Client) {
	app.Use(middleware.RequestID())
	app.Use(recoverer.New())

	if cfg.Server.Environment == "production" {
		app.Use(helmet.New())
		if cfg.Server.CORS.Enabled {
			app.Use(cors.New(cors.Config{AllowOrigins: cfg.Server.CORS.AllowOrigins}))
		}
		app.Use(middleware.NewLimiterWithRedis(rdb))
	}

	app.Use(logger.New(logger.Config{
		Format: "${ip} - [${time}] [req_id=${requestId}] ${method} ${url} ${status}\n",
	}))
}
