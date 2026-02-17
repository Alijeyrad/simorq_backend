package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	httpcmd "github.com/Alijeyrad/simorq_backend/cmd/http"
	systemcmd "github.com/Alijeyrad/simorq_backend/cmd/system"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "simorq",
	Short: "Simorq Multi-tenant SaaS platform for psychology and therapy clinics.",
	Long: `Simorgh is a multi-tenant SaaS platform for psychology and therapy clinics. 
It connects clinics with clients through a single web application, one unified deployment.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global config flag, available for all commands.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file path")

	// Attach top-level command trees.
	rootCmd.AddCommand(systemcmd.NewSystemCommand())
	rootCmd.AddCommand(httpcmd.NewHTTPCommand())
}
