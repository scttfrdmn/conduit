package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

var (
	importDbPath      string
	importInput       string
	importOnConflict  string
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import catalog data from backup",
	Long: `Import models from an export file into the catalog.

Supports conflict resolution strategies for handling existing models:
- skip: Skip models that already exist (default)
- overwrite: Replace existing models completely
- merge: Add new versions to existing models

Examples:
  # Import catalog (skip existing models)
  conduit import catalog-backup.json

  # Import and overwrite existing models
  conduit import catalog-backup.json --on-conflict overwrite

  # Import and merge new versions
  conduit import catalog-backup.json --on-conflict merge

  # Import to custom database
  conduit import backup.json --db /path/to/catalog.db`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&importDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
	importCmd.Flags().StringVar(&importOnConflict, "on-conflict", "skip", "Conflict resolution strategy (skip, overwrite, merge)")
}

func runImport(cmd *cobra.Command, args []string) (err error) {
	importInput = args[0]

	// Validate conflict strategy
	var strategy catalog.ConflictStrategy
	switch importOnConflict {
	case "skip":
		strategy = catalog.ConflictSkip
	case "overwrite":
		strategy = catalog.ConflictOverwrite
	case "merge":
		strategy = catalog.ConflictMerge
	default:
		return fmt.Errorf("invalid conflict strategy: %s (use: skip, overwrite, merge)", importOnConflict)
	}

	// Open catalog database
	db, err := catalog.NewDB(importDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Import catalog
	fmt.Printf("Importing from: %s\n", importInput)
	fmt.Printf("Conflict strategy: %s\n\n", strategy)

	result, err := db.ImportCatalog(importInput, strategy)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	// Display results
	fmt.Printf("Import complete:\n")
	fmt.Printf("  Total models: %d\n", result.TotalModels)
	fmt.Printf("  Imported: %d\n", result.ImportedModels)
	fmt.Printf("  Updated: %d\n", result.UpdatedModels)
	fmt.Printf("  Skipped: %d\n", result.SkippedModels)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %s: %s\n", e.ModelName, e.Error)
		}
		return fmt.Errorf("import completed with %d error(s)", len(result.Errors))
	}

	return nil
}
