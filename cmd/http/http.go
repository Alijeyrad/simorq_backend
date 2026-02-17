package http

import "github.com/spf13/cobra"

func NewHTTPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "HTTP server commands",
	}

	cmd.AddCommand(NewStartCommand())

	return cmd
}
