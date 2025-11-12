package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
	"github.com/scttfrdmn/conduit/internal/model"
)

var (
	publishDbPath string
)

var publishCmd = &cobra.Command{
	Use:   "publish [path]",
	Short: "Publish a model to the catalog",
	Long: `Publish a model to the Conduit catalog.

Reads a model.yaml file, validates it, and registers the model
in the local catalog database. This makes the model discoverable
via 'conduit search'.

Examples:
  conduit publish model.yaml
  conduit publish ./my-model/model.yaml
  conduit publish .  # looks for model.yaml in current directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPublish,
}

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().StringVar(&publishDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
}

func runPublish(cmd *cobra.Command, args []string) (err error) {
	// Determine the path to the model.yaml file
	path := "model.yaml"
	if len(args) > 0 {
		path = args[0]
	}

	// If path is a directory, look for model.yaml in it
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	if info.IsDir() {
		path = filepath.Join(path, "model.yaml")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	fmt.Printf("Publishing %s to catalog...\n", path)

	// Parse the model file
	parser := model.NewParser()
	m, err := parser.ParseFile(path)
	if err != nil {
		fmt.Printf("❌ Parsing failed:\n")
		return err
	}

	fmt.Printf("✓ Parsed successfully\n")

	// Validate the model
	validator := model.NewValidator()
	if err := validator.Validate(m); err != nil {
		fmt.Printf("❌ Validation failed:\n")
		return err
	}

	fmt.Printf("✓ Validation passed\n")

	// Open catalog database
	db, err := catalog.NewDB(publishDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Check if model already exists
	existing, err := db.GetModel(m.Name)
	if err == nil {
		// Model exists - check if version is different
		if existing.LatestVersion != nil && existing.LatestVersion.Version == m.Version {
			return fmt.Errorf("version %s already exists for model %s", m.Version, m.Name)
		}

		// Publishing new version
		fmt.Printf("Found existing model '%s'\n", m.Name)
		fmt.Printf("   Current version: %s\n", existing.LatestVersion.Version)
		fmt.Printf("   New version: %s\n", m.Version)
		fmt.Printf("Publishing new version...\n")

		if err := db.CreateModelVersion(existing.ID, m); err != nil {
			return fmt.Errorf("failed to create new version: %w", err)
		}

		fmt.Printf("✓ Published new version\n")
		fmt.Printf("\n✅ Successfully published %s v%s!\n\n", m.Name, m.Version)
		fmt.Printf("View it with: conduit info %s\n", m.Name)
		return nil
	}

	// Model doesn't exist - create it
	fmt.Printf("Registering new model '%s'...\n", m.Name)
	id, err := db.CreateModel(m)
	if err != nil {
		return fmt.Errorf("failed to register model: %w", err)
	}

	fmt.Printf("✓ Registered in catalog (ID: %d)\n", id)

	fmt.Printf("\n✅ Successfully published %s v%s!\n\n", m.Name, m.Version)
	fmt.Printf("Search for it with: conduit search %s\n", m.Name)

	return nil
}
