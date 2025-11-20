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
	pushRegistry string
	pushForce    bool
)

var pushCmd = &cobra.Command{
	Use:   "push <model-name>[@version]",
	Short: "Push a model to a remote registry",
	Long: `Push a model from your local catalog to a remote registry.

This allows you to share models with your team or publish them publicly.

If no version is specified, the latest version will be pushed.
If no registry is specified, the default registry will be used.

Examples:
  conduit push alphafold2
  conduit push alphafold2@v2.3.2
  conduit push alphafold2 --registry myteam
  conduit push alphafold2 --force  # Overwrite if already exists`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVar(&pushRegistry, "registry", "", "Registry to push to (default: use default registry)")
	pushCmd.Flags().BoolVar(&pushForce, "force", false, "Overwrite model if it already exists in registry")
}

func runPush(cmd *cobra.Command, args []string) error {
	modelRef := args[0]

	// Parse model reference (name@version)
	name, version := parseModelRef(modelRef)

	// Load registry config
	regConfig, err := registry.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load registry config: %w", err)
	}

	// Get the registry to push to
	var targetRegistry *registry.Registry
	if pushRegistry != "" {
		targetRegistry, err = regConfig.GetRegistry(pushRegistry)
		if err != nil {
			return err
		}
	} else {
		targetRegistry, err = regConfig.GetDefaultRegistry()
		if err != nil {
			return fmt.Errorf("%w\n\nAdd a registry with: conduit registry add <name> <url>", err)
		}
	}

	fmt.Printf("Pushing to registry: %s (%s)\n", targetRegistry.Name, targetRegistry.URL)
	fmt.Printf("Model: %s", name)
	if version != "" {
		fmt.Printf("@%s", version)
	}
	fmt.Println()
	fmt.Println()

	// Open local catalog
	db, err := catalog.NewDB("")
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer db.Close() //nolint:errcheck // Defer close is safe

	// Export model to temporary file
	tempDir, err := os.MkdirTemp("", "conduit-push-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck // Best effort cleanup

	exportPath := filepath.Join(tempDir, "model-export.json")

	fmt.Println("Exporting model...")
	if err := db.ExportModel(name, exportPath); err != nil {
		return fmt.Errorf("failed to export model: %w", err)
	}

	// Push to registry
	fmt.Println("Uploading to registry...")

	client, err := registry.NewClient(targetRegistry)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	if err := client.Push(exportPath, name, version, pushForce); err != nil {
		return fmt.Errorf("failed to push to registry: %w", err)
	}

	fmt.Println()
	fmt.Printf("✅ Successfully pushed %s to %s\n", modelRef, targetRegistry.Name)

	return nil
}

// parseModelRef parses a model reference like "name@version" into name and version
func parseModelRef(ref string) (name string, version string) {
	// Simple parsing for now - split on @
	for i, c := range ref {
		if c == '@' {
			return ref[:i], ref[i+1:]
		}
	}
	return ref, ""
}
