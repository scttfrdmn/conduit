package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/model"
	"github.com/scttfrdmn/conduit/internal/validation"
)

var (
	validateStrict bool
)

var validateCmd = &cobra.Command{
	Use:   "validate [model.yaml]",
	Short: "Validate a model configuration file",
	Long: `Validate a model configuration file (model.yaml) for correctness.

By default, performs basic validation checking required fields.
Use --strict flag for comprehensive validation including best practices.

Basic validation checks:
  - Required fields (name, version, domain, weights_uri)
  - Runtime configuration (framework, python_version)
  - Inference configuration (entrypoint, handler)
  - Hardware requirements (GPU memory warnings)

Strict validation additionally checks:
  - Best practices (description, license, tags, benchmarks)
  - File existence (dependencies, entrypoint files)
  - URL formats (weights_uri, github_repo, citation)
  - Version format (semantic versioning)

Examples:
  conduit validate model.yaml
  conduit validate model.yaml --strict
  conduit validate /path/to/model.yaml --strict
  conduit validate .  # looks for model.yaml in current directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Enable strict validation including best practices")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Determine the path to validate
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

	fmt.Printf("Validating %s...\n\n", path)

	// Load model configuration
	parser := model.NewParser()
	m, err := parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("failed to parse model file: %w", err)
	}

	// Determine validation level
	level := validation.ValidationBasic
	if validateStrict {
		level = validation.ValidationStrict
		fmt.Println("Using strict validation mode")
		fmt.Println()
	}

	// Validate model
	result := validation.ValidateModel(m, level)

	// Display results
	if len(result.Errors) > 0 {
		fmt.Println("❌ Validation failed with errors:")
		fmt.Println()
		for _, err := range result.Errors {
			fmt.Printf("  ERROR [%s]: %s\n", err.Field, err.Message)
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Println("⚠️  Validation warnings:")
		fmt.Println()
		for _, warn := range result.Warnings {
			fmt.Printf("  WARNING [%s]: %s\n", warn.Field, warn.Message)
		}
		fmt.Println()
	}

	// Summary
	if result.IsValid() {
		if len(result.Warnings) > 0 {
			fmt.Printf("✅ Model is valid with %d warning(s)\n", len(result.Warnings))
		} else {
			fmt.Println("✅ Model is valid")
		}
		return nil
	}

	fmt.Printf("❌ Model validation failed with %d error(s)\n", len(result.Errors))
	os.Exit(1)
	return nil
}
