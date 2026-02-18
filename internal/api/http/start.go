package http

import (
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/api/http/router"
	"github.com/Alijeyrad/simorq_backend/internal/app"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Start(cfg *config.Config, timeout time.Duration) {
	fx.New(
		fx.Supply(cfg),
		app.InfraModule,
		app.ServiceModule,
		router.Module,
		Module, // This is the http.Module from server.go

		// IMPORTANT: Invoke *fiber.App because that's what NewServer returns
		// This forces the creation of fiber.App, triggering the OnStart hook
		fx.Invoke(func(*fiber.App) {}),

		fx.StopTimeout(timeout),
	).Run()
}
