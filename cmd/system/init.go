package system

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/pkg/database"
)

func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize all databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config flag: %w", err)
			}
			cfg, err := config.ReadConfig(filepath.Dir(cfgPath))
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}

			fmt.Println("Initializing databases...")
			err = database.InitializeDatabases(cfg)
			if err != nil {
				return fmt.Errorf("failed to initialize databases: %w", err)
			}
			fmt.Println("Databases Initialized successfully.")
			return nil
		},
	}

	return cmd
}
