package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

var (
	listDbPath string
	listLimit  int
	listOffset int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all models in the catalog",
	Long: `Display all models in the catalog with pagination.

Shows model names, domains, and latest versions.

Examples:
  conduit list
  conduit list --limit 50
  conduit list --limit 10 --offset 20`,
	Args: cobra.NoArgs,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
	listCmd.Flags().IntVar(&listLimit, "limit", 20, "Maximum number of models to display")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "Number of models to skip")
}

func runList(cmd *cobra.Command, args []string) (err error) {
	// Open catalog database
	db, err := catalog.NewDB(listDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Get total count
	total, err := db.CountModels()
	if err != nil {
		return fmt.Errorf("failed to count models: %w", err)
	}

	if total == 0 {
		fmt.Println("No models in catalog.")
		fmt.Println("\nPublish a model with: conduit publish model.yaml")
		return nil
	}

	// List models
	models, err := db.ListModels(listLimit, listOffset)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		fmt.Printf("No models found at offset %d.\n", listOffset)
		return nil
	}

	// Display header
	fmt.Printf("\nModels in catalog (%d total):\n\n", total)

	// Display models
	for i, model := range models {
		displayNum := listOffset + i + 1
		fmt.Printf("%d. %s\n", displayNum, model.Name)
		fmt.Printf("   Domain:  %s\n", model.Domain)

		if model.Description != "" {
			desc := model.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("   Desc:    %s\n", desc)
		}

		fmt.Printf("   Updated: %s\n", model.UpdatedAt.Format("2006-01-02"))
		fmt.Println()
	}

	// Show pagination info
	showing := len(models)
	rangeStart := listOffset + 1
	rangeEnd := listOffset + showing

	fmt.Printf("Showing %d-%d of %d models\n\n", rangeStart, rangeEnd, total)

	// Show navigation hints
	if rangeEnd < int(total) {
		nextOffset := listOffset + listLimit
		fmt.Printf("View more: conduit list --offset %d\n", nextOffset)
	}

	fmt.Printf("View details: conduit info <model-name>\n")
	fmt.Printf("Search: conduit search <query>\n")

	return nil
}
