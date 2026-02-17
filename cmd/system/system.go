package system

import "github.com/spf13/cobra"

func NewSystemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Maintenance and tooling commands",
	}

	cmd.AddCommand(NewMigrateCommand())
	cmd.AddCommand(NewGenDocsCommand())
	cmd.AddCommand(NewInitCommand())

	return cmd
}
