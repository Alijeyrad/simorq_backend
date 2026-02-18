package http

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/api/http"
	"github.com/Alijeyrad/simorq_backend/pkg/logs"
	"github.com/spf13/cobra"
)

func NewStartCommand() *cobra.Command {
	var shutdownTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the HTTP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := cmd.Root().PersistentFlags().GetString("config")
			cfg, err := config.ReadConfig(filepath.Dir(cfgPath))
			if err != nil {
				return err
			}

			slog.SetDefault(logs.New(cfg))

			// Just start the app defined in internal/api/http
			http.Start(cfg, shutdownTimeout)
			return nil
		},
	}

	cmd.Flags().DurationVar(&shutdownTimeout, "shutdown-timeout", 30*time.Second, "Maximum time to wait for graceful shutdown")

	return cmd
}
