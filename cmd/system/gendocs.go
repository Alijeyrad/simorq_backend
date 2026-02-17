package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewGenDocsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gendocs",
		Short: "Generate CLI documentation in Markdown format",
		Long: `Generate Markdown documentation for all Simorgh CLI commands.

By default, docs are written to ./docs/cli. You can override this with --outdir.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir, err := cmd.Flags().GetString("outdir")
			if err != nil {
				return fmt.Errorf("failed to read outdir flag: %w", err)
			}
			if outDir == "" {
				outDir = "docs/cli"
			}

			// Ensure output directory exists
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return fmt.Errorf("failed to create docs directory %q: %w", outDir, err)
			}

			// Root() gives us the full command tree (rootCmd) at runtime.
			root := cmd.Root()

			// Normalize to absolute path (optional, but nice)
			absOutDir, err := filepath.Abs(outDir)
			if err != nil {
				return fmt.Errorf("failed to resolve absolute path for %q: %w", outDir, err)
			}

			if err := doc.GenMarkdownTree(root, absOutDir); err != nil {
				return fmt.Errorf("failed to generate CLI docs: %w", err)
			}

			fmt.Printf("CLI docs generated in %s\n", absOutDir)
			return nil
		},
	}

	cmd.Flags().String("outdir", "docs/cli", "Output directory for generated CLI docs")

	return cmd
}
