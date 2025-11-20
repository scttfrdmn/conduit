package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/registry"
)

var (
	registryType    string
	registryDefault bool
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage remote registries",
	Long: `Manage remote registry configurations for model sharing and collaboration.

Registries allow you to push and pull models to/from remote locations,
enabling team collaboration and model distribution.

Examples:
  conduit registry list
  conduit registry add myteam https://registry.example.com
  conduit registry remove myteam
  conduit registry set-default myteam`,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured registries",
	Long:  `List all configured registries with their URLs and types.`,
	RunE:  runRegistryList,
}

var registryAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a new registry",
	Long: `Add a new registry configuration.

The registry type can be:
  http  - HTTP-based registry server (default)
  s3    - S3-backed static registry
  git   - Git-backed registry (GitHub releases)

Examples:
  conduit registry add myteam https://registry.example.com
  conduit registry add myteam https://registry.example.com --type http
  conduit registry add backup s3://my-bucket/models --type s3
  conduit registry add public github.com/org/models --type git`,
	Args: cobra.ExactArgs(2),
	RunE: runRegistryAdd,
}

var registryRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registry",
	Long:  `Remove a registry configuration by name.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRegistryRemove,
}

var registrySetDefaultCmd = &cobra.Command{
	Use:   "set-default <name>",
	Short: "Set the default registry",
	Long: `Set a registry as the default for push/pull operations.

When no registry is specified in push/pull commands, the default registry will be used.`,
	Args: cobra.ExactArgs(1),
	RunE: runRegistrySetDefault,
}

func init() {
	rootCmd.AddCommand(registryCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryRemoveCmd)
	registryCmd.AddCommand(registrySetDefaultCmd)

	registryAddCmd.Flags().StringVar(&registryType, "type", "http", "Registry type (http, s3, git)")
	registryAddCmd.Flags().BoolVar(&registryDefault, "default", false, "Set as default registry")
}

func runRegistryList(cmd *cobra.Command, args []string) error {
	config, err := registry.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(config.Registries) == 0 {
		fmt.Println("No registries configured.")
		fmt.Println()
		fmt.Println("Add a registry with:")
		fmt.Println("  conduit registry add <name> <url>")
		return nil
	}

	fmt.Println("Configured Registries:")
	fmt.Println()

	for _, reg := range config.Registries {
		defaultMarker := ""
		if reg.Default {
			defaultMarker = " (default)"
		}

		fmt.Printf("  %s%s\n", reg.Name, defaultMarker)
		fmt.Printf("    URL:  %s\n", reg.URL)
		fmt.Printf("    Type: %s\n", reg.Type)
		if reg.Username != "" {
			fmt.Printf("    User: %s\n", reg.Username)
		}
		fmt.Println()
	}

	return nil
}

func runRegistryAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	url := args[1]

	// Validate registry type
	validTypes := map[string]bool{"http": true, "s3": true, "git": true}
	if !validTypes[registryType] {
		return fmt.Errorf("invalid registry type: %s (must be http, s3, or git)", registryType)
	}

	config, err := registry.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reg := registry.Registry{
		Name:    name,
		URL:     url,
		Type:    registryType,
		Default: registryDefault,
	}

	if err := config.AddRegistry(reg); err != nil {
		return fmt.Errorf("failed to add registry: %w", err)
	}

	if err := registry.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Added registry: %s\n", name)
	fmt.Printf("  URL:  %s\n", url)
	fmt.Printf("  Type: %s\n", registryType)

	if reg.Default {
		fmt.Println("  Set as default registry")
	}

	return nil
}

func runRegistryRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	config, err := registry.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.RemoveRegistry(name); err != nil {
		return err
	}

	if err := registry.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Removed registry: %s\n", name)
	return nil
}

func runRegistrySetDefault(cmd *cobra.Command, args []string) error {
	name := args[0]

	config, err := registry.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.SetDefault(name); err != nil {
		return err
	}

	if err := registry.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Set %s as default registry\n", name)
	return nil
}
