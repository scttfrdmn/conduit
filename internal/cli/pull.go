package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
	"github.com/scttfrdmn/conduit/internal/registry"
)

var (
	pullRegistry string
	pullStrategy string
)

var pullCmd = &cobra.Command{
	Use:   "pull <model-name>[@version]",
	Short: "Pull a model from a remote registry",
	Long: `Pull a model from a remote registry to your local catalog.

If no version is specified, the latest version will be pulled.
If no registry is specified, the default registry will be used.

Conflict strategy determines what happens if the model already exists locally:
  skip      - Skip if model exists (default)
  overwrite - Replace local model with remote version
  merge     - Merge versions (keep both)

Examples:
  conduit pull alphafold2
  conduit pull alphafold2@v2.3.2
  conduit pull alphafold2 --registry myteam
  conduit pull alphafold2 --strategy overwrite`,
	Args: cobra.ExactArgs(1),
	RunE: runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringVar(&pullRegistry, "registry", "", "Registry to pull from (default: use default registry)")
	pullCmd.Flags().StringVar(&pullStrategy, "strategy", "skip", "Conflict resolution strategy (skip, overwrite, merge)")
}

func runPull(cmd *cobra.Command, args []string) error {
	modelRef := args[0]

	// Parse model reference (name@version)
	name, version := parseModelRef(modelRef)

	// Validate strategy
	strategy := catalog.ConflictStrategy(pullStrategy)
	if strategy != catalog.ConflictSkip && strategy != catalog.ConflictOverwrite && strategy != catalog.ConflictMerge {
		return fmt.Errorf("invalid strategy: %s (must be skip, overwrite, or merge)", pullStrategy)
	}

	// Load registry config
	regConfig, err := registry.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load registry config: %w", err)
	}

	// Get the registry to pull from
	var sourceRegistry *registry.Registry
	if pullRegistry != "" {
		sourceRegistry, err = regConfig.GetRegistry(pullRegistry)
		if err != nil {
			return err
		}
	} else {
		sourceRegistry, err = regConfig.GetDefaultRegistry()
		if err != nil {
			return fmt.Errorf("%w\n\nAdd a registry with: conduit registry add <name> <url>", err)
		}
	}

	fmt.Printf("Pulling from registry: %s (%s)\n", sourceRegistry.Name, sourceRegistry.URL)
	fmt.Printf("Model: %s", name)
	if version != "" {
		fmt.Printf("@%s", version)
	}
	fmt.Println()
	fmt.Println()

	// Download from registry
	tempDir, err := os.MkdirTemp("", "conduit-pull-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck // Best effort cleanup

	downloadPath := filepath.Join(tempDir, "model-export.json")

	fmt.Println("Downloading from registry...")

	client, err := registry.NewClient(sourceRegistry)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	if err := client.Pull(name, version, downloadPath); err != nil {
		return fmt.Errorf("failed to pull from registry: %w", err)
	}

	// Open local catalog
	db, err := catalog.NewDB("")
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer db.Close() //nolint:errcheck // Defer close is safe

	// Import into local catalog
	fmt.Println("Importing into local catalog...")

	result, err := db.ImportCatalog(downloadPath, strategy)
	if err != nil {
		return fmt.Errorf("failed to import model: %w", err)
	}

	// Print results
	fmt.Println()
	fmt.Println("Import Results:")
	fmt.Printf("  Imported: %d model(s)\n", result.ImportedModels)
	fmt.Printf("  Updated:  %d model(s)\n", result.UpdatedModels)
	fmt.Printf("  Skipped:  %d model(s)\n", result.SkippedModels)
	fmt.Println()

	if result.ImportedModels > 0 || result.UpdatedModels > 0 {
		fmt.Printf("✅ Successfully pulled %s from %s\n", modelRef, sourceRegistry.Name)
	} else if result.SkippedModels > 0 {
		fmt.Printf("⚠️  Model already exists locally (use --strategy overwrite to replace)\n")
	}

	return nil
}
