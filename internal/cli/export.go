package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

var (
	exportDbPath string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export [model-name]",
	Short: "Export catalog or model data",
	Long: `Export catalog data to JSON format for backup, migration, or sharing.

Can export the entire catalog or a single model with all versions and metadata.
The export file includes all model data: versions, benchmarks, citations, and tags.

Examples:
  # Export entire catalog
  conduit export --output catalog-backup.json

  # Export single model
  conduit export alphafold2 --output alphafold2.json

  # Export with custom database
  conduit export --db /path/to/catalog.db --output backup.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (required)")
	_ = exportCmd.MarkFlagRequired("output") // Error only occurs if flag doesn't exist
}

func runExport(cmd *cobra.Command, args []string) (err error) {
	// Get model name if specified
	var modelName string
	if len(args) > 0 {
		modelName = args[0]
	}

	// Open catalog database
	db, err := catalog.NewDB(exportDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Export catalog or single model
	if modelName != "" {
		fmt.Printf("Exporting model: %s\n", modelName)
		if err := db.ExportModel(modelName, exportOutput); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		fmt.Printf("Model exported to: %s\n", exportOutput)
	} else {
		fmt.Println("Exporting entire catalog...")
		if err := db.ExportCatalog(exportOutput); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		fmt.Printf("Catalog exported to: %s\n", exportOutput)
	}

	return nil
}
