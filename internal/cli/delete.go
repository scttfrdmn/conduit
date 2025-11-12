package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

var (
	deleteDbPath string
	deleteVersion string
	deleteForce bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <model-name>",
	Short: "Delete a model from the catalog",
	Long: `Delete a model or specific version from the catalog.

By default, deletes the entire model and all its versions.
Use --version to delete a specific version only.

Examples:
  conduit delete alphafold2              # Delete entire model
  conduit delete alphafold2 --version 1.0.0  # Delete specific version
  conduit delete alphafold2 --force      # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&deleteDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
	deleteCmd.Flags().StringVar(&deleteVersion, "version", "", "Delete specific version only")
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) (err error) {
	modelName := args[0]

	// Open catalog database
	db, err := catalog.NewDB(deleteDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Check if model exists
	_, err = db.GetModel(modelName)
	if err != nil {
		return fmt.Errorf("model not found: %s", modelName)
	}

	// Determine what we're deleting
	var confirmMsg string
	if deleteVersion != "" {
		// Deleting specific version
		confirmMsg = fmt.Sprintf("Delete version %s of model '%s'?", deleteVersion, modelName)
	} else {
		// Deleting entire model
		versions, err := db.ListModelVersions(modelName)
		if err != nil {
			return fmt.Errorf("failed to list versions: %w", err)
		}
		confirmMsg = fmt.Sprintf("Delete model '%s' and all %d version(s)?", modelName, len(versions))
	}

	// Confirm deletion unless --force
	if !deleteForce {
		fmt.Printf("\n%s\n", confirmMsg)
		fmt.Printf("This action cannot be undone.\n\n")
		fmt.Printf("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("\nDeletion cancelled.")
			return nil
		}
	}

	// Perform deletion
	if deleteVersion != "" {
		// Delete specific version
		if err := db.DeleteModelVersion(modelName, deleteVersion); err != nil {
			return fmt.Errorf("failed to delete version: %w", err)
		}
		fmt.Printf("\n✓ Deleted version %s of %s\n", deleteVersion, modelName)
		fmt.Printf("\nView remaining versions: conduit versions %s\n", modelName)
	} else {
		// Delete entire model
		if err := db.DeleteModel(modelName); err != nil {
			return fmt.Errorf("failed to delete model: %w", err)
		}
		fmt.Printf("\n✓ Deleted model '%s'\n", modelName)

		// Show total count
		total, err := db.CountModels()
		if err == nil {
			fmt.Printf("\nRemaining models in catalog: %d\n", total)
		}
	}

	return nil
}
