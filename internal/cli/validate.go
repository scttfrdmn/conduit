package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/scttfrdmn/conduit/internal/model"
)

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate a model.yaml file",
	Long: `Validate a model.yaml file for correctness.

Checks that all required fields are present, formats are correct,
and values are valid. Provides clear error messages for any issues found.

Examples:
  conduit validate model.yaml
  conduit validate ./my-model/model.yaml
  conduit validate .  # looks for model.yaml in current directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
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

	fmt.Printf("Validating %s...\n", path)

	// Parse the file
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

	fmt.Printf("✓ Validation passed\n\n")
	fmt.Printf("Model: %s v%s\n", m.Name, m.Version)
	fmt.Printf("Domain: %s\n", m.Domain)
	fmt.Printf("Framework: %s\n", m.Runtime.Framework)
	fmt.Printf("GPU Required: %v\n", m.Hardware.GPURequired)
	fmt.Printf("Recommended Instance: %s\n", m.Hardware.RecommendedInstance)

	if len(m.Benchmarks) > 0 {
		fmt.Printf("\nBenchmarks: %d\n", len(m.Benchmarks))
		for i, b := range m.Benchmarks {
			fmt.Printf("  %d. %s: %s = %.2f\n", i+1, b.Dataset, b.Metric, b.Result)
		}
	}

	fmt.Printf("\n✅ All checks passed! Model specification is valid.\n")
	return nil
}
