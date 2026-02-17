package system

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/pkg/authorize"
	"github.com/Alijeyrad/simorq_backend/pkg/database"
)

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config flag: %w", err)
			}
			cfg, err := config.ReadConfig(filepath.Dir(cfgPath))
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}

			// client db
			fmt.Println("Running Migrations For Client DB.")
			client, err := database.NewEntClient(cfg.Database)
			if err != nil {
				return fmt.Errorf("failed to create ent client: %w", err)
			}
			defer client.Close()

			timeout := time.Duration(cfg.Server.TimeoutSeconds) * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := database.MigrateEnt(ctx, client); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			// casbin db
			fmt.Println("Running Migrations For Casbin DB.")

			casbinDBDSN := database.NewDSN(cfg.CasbinDatabase)
			enforcer, cleanup, err := authorize.NewEnforcer(cfg.Authorization.CasbinModelPath, casbinDBDSN)
			if err != nil {
				return fmt.Errorf("failed to create enforcer: %w", err)
			}
			defer cleanup(context.Background())

			auth, err := authorize.NewAuthorization(enforcer)
			if err != nil {
				return fmt.Errorf("failed to create authorization: %w", err)
			}

			// Seed Casbin policies
			slog.Info("Seeding Casbin policies...")
			if err := authorize.SeedDefaultPolicies(context.Background(), auth); err != nil {
				return fmt.Errorf("failed to seed policies: %w", err)
			}

			fmt.Println("Migrations executed successfully.")
			return nil
		},
	}

	return cmd
}
