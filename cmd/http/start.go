package http

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/api/http"
	"github.com/Alijeyrad/simorq_backend/internal/app"
	"github.com/Alijeyrad/simorq_backend/pkg/logs"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func NewStartCommand() *cobra.Command {
	var shutdownTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the HTTP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return err
			}

			cfg, err := config.ReadConfig(filepath.Dir(cfgPath))
			if err != nil {
				return err
			}

			// Set up structured logger before fx starts so all logs use it.
			slog.SetDefault(logs.New(cfg))

			fxApp := fx.New(
				fx.Supply(cfg),
				app.InfraModule,
				app.ServiceModule,
				http.Module,
				fx.Invoke(func(*http.Server) {}),
				fx.StopTimeout(shutdownTimeout),
				fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
			)

			fxApp.Run()
			return nil
		},
	}

	cmd.Flags().DurationVar(&shutdownTimeout, "shutdown-timeout", 30*time.Second, "Maximum time to wait for graceful shutdown")

	return cmd
}
