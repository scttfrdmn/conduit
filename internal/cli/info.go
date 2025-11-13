package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

var infoDbPath string

var infoCmd = &cobra.Command{
	Use:   "info <model-name>",
	Short: "Show detailed information about a model",
	Long: `Display comprehensive information about a model from the catalog.

Shows model metadata, version details, hardware requirements,
benchmarks, and citation information if available.

Examples:
  conduit info alphafold2
  conduit info esm2-650m --db /custom/path/catalog.db`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringVar(&infoDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
}

func runInfo(cmd *cobra.Command, args []string) (err error) {
	modelName := args[0]

	// Open catalog database
	db, err := catalog.NewDB(infoDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Get model
	model, err := db.GetModel(modelName)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	// Display model information
	fmt.Printf("\n=== %s ===\n\n", model.Name)

	// Basic info
	fmt.Printf("Domain:      %s\n", model.Domain)
	if model.Description != "" {
		fmt.Printf("Description: %s\n", model.Description)
	}
	if len(model.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(model.Tags, ", "))
	}
	if model.GitHubRepo != "" {
		fmt.Printf("GitHub:      %s\n", model.GitHubRepo)
	}
	if model.License != "" {
		fmt.Printf("License:     %s\n", model.License)
	}
	fmt.Printf("Created:     %s\n", model.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:     %s\n", model.UpdatedAt.Format("2006-01-02 15:04:05"))

	// Version info
	if model.LatestVersion != nil {
		v := model.LatestVersion
		fmt.Printf("\n--- Latest Version: %s ---\n\n", v.Version)

		fmt.Printf("Weights URI:  %s\n", v.WeightsURI)
		if v.WeightsSizeGB > 0 {
			fmt.Printf("Weights Size: %.2f GB\n", v.WeightsSizeGB)
		}
		if v.ChecksumSHA256 != "" {
			fmt.Printf("Checksum:     %s\n", v.ChecksumSHA256)
		}

		// Runtime
		fmt.Printf("\n--- Runtime ---\n\n")
		fmt.Printf("Framework:      %s\n", v.Framework)
		fmt.Printf("Python Version: %s\n", v.PythonVersion)
		if v.Dependencies != "" {
			fmt.Printf("Dependencies:   %s\n", v.Dependencies)
		}
		if v.CustomImage != "" {
			fmt.Printf("Custom Image:   %s\n", v.CustomImage)
		}

		// Inference
		fmt.Printf("\n--- Inference ---\n\n")
		fmt.Printf("Entrypoint: %s\n", v.Entrypoint)
		fmt.Printf("Handler:    %s\n", v.Handler)

		// Hardware
		fmt.Printf("\n--- Hardware Requirements ---\n\n")
		fmt.Printf("GPU Required: %v\n", v.GPURequired)
		if v.RecommendedInstance != "" {
			fmt.Printf("Recommended:  %s\n", v.RecommendedInstance)
		}
		if v.MinCPU > 0 {
			fmt.Printf("Min CPU:      %d cores\n", v.MinCPU)
		}
		if v.MinMemoryGB > 0 {
			fmt.Printf("Min Memory:   %d GB\n", v.MinMemoryGB)
		}
		if v.MinGPUMemoryGB > 0 {
			fmt.Printf("Min GPU VRAM: %d GB\n", v.MinGPUMemoryGB)
		}

		// Benchmarks
		if len(v.Benchmarks) > 0 {
			fmt.Printf("\n--- Benchmarks ---\n\n")
			for _, b := range v.Benchmarks {
				fmt.Printf("Dataset: %s\n", b.Dataset)
				fmt.Printf("  Metric:  %s = %.4f\n", b.Metric, b.Result)
				if b.Instance != "" {
					fmt.Printf("  Instance: %s\n", b.Instance)
				}
				if b.CostPerPrediction != "" {
					fmt.Printf("  Cost/Prediction: %s\n", b.CostPerPrediction)
				}
				if b.WalltimeSeconds > 0 {
					fmt.Printf("  Walltime: %.2fs\n", b.WalltimeSeconds)
				}
				fmt.Println()
			}
		}
	}

	// Citation
	if model.Citation != nil {
		c := model.Citation
		fmt.Printf("\n--- Citation ---\n\n")
		if c.PaperTitle != "" {
			fmt.Printf("Title:   %s\n", c.PaperTitle)
		}
		if c.Authors != "" {
			// Split authors by comma and format nicely
			authors := strings.Split(c.Authors, ",")
			fmt.Printf("Authors: %s\n", strings.TrimSpace(authors[0]))
			for i := 1; i < len(authors); i++ {
				fmt.Printf("         %s\n", strings.TrimSpace(authors[i]))
			}
		}
		if c.Year > 0 {
			fmt.Printf("Year:    %d\n", c.Year)
		}
		if c.DOI != "" {
			fmt.Printf("DOI:     %s\n", c.DOI)
		}
		if c.PaperURL != "" {
			fmt.Printf("URL:     %s\n", c.PaperURL)
		}
		if c.BibTeX != "" {
			fmt.Printf("\nBibTeX:\n%s\n", c.BibTeX)
		}
	}

	fmt.Println()
	return nil
}
