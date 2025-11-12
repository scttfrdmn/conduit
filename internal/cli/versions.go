package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

var versionsDbPath string

var versionsCmd = &cobra.Command{
	Use:   "versions <model-name>",
	Short: "List all versions of a model",
	Long: `Display all versions of a model in the catalog.

Shows version history with creation dates and latest flag.

Examples:
  conduit versions alphafold2
  conduit versions esm2-650m --db /custom/path/catalog.db`,
	Args: cobra.ExactArgs(1),
	RunE: runVersions,
}

func init() {
	rootCmd.AddCommand(versionsCmd)
	versionsCmd.Flags().StringVar(&versionsDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
}

func runVersions(cmd *cobra.Command, args []string) (err error) {
	modelName := args[0]

	// Open catalog database
	db, err := catalog.NewDB(versionsDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Get all versions
	versions, err := db.ListModelVersions(modelName)
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Printf("No versions found for model: %s\n", modelName)
		return nil
	}

	fmt.Printf("\nVersions of %s:\n\n", modelName)

	for _, v := range versions {
		latest := ""
		if v.IsLatest {
			latest = " (latest)"
		}

		fmt.Printf("• v%s%s\n", v.Version, latest)
		fmt.Printf("  Created:   %s\n", v.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Framework: %s\n", v.Framework)
		fmt.Printf("  GPU:       %v\n", v.GPURequired)

		if len(v.Benchmarks) > 0 {
			fmt.Printf("  Benchmarks: %d\n", len(v.Benchmarks))
		}

		fmt.Println()
	}

	fmt.Printf("Total versions: %d\n\n", len(versions))
	fmt.Printf("View details: conduit info %s\n", modelName)

	return nil
}
