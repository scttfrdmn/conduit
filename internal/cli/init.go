package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/scttfrdmn/conduit/internal/model"
	"github.com/scttfrdmn/conduit/pkg/types"
)

var (
	initForce bool
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new model project",
	Long: `Initialize a new model project with scaffolding.

Creates a new directory with model.yaml, inference code templates,
and a recommended project structure.

Examples:
  conduit init my-model
  conduit init ./alphafold-variant
  conduit init .  # initialize in current directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing files")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine target directory
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	// Check if directory exists and is empty (unless --force)
	if err := checkDirectory(targetDir); err != nil && !initForce {
		return err
	}

	fmt.Println("🚀 Initializing new Conduit model project")

	// Gather model information interactively
	modelData, err := promptModelInfo()
	if err != nil {
		return fmt.Errorf("failed to gather model information: %w", err)
	}

	// Validate the model data
	validator := model.NewValidator()
	if err := validator.Validate(modelData); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create directory structure
	if err := createDirectoryStructure(targetDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate model.yaml
	if err := generateModelYAML(targetDir, modelData); err != nil {
		return fmt.Errorf("failed to generate model.yaml: %w", err)
	}

	// Create template files
	if err := createTemplateFiles(targetDir, modelData); err != nil {
		return fmt.Errorf("failed to create template files: %w", err)
	}

	fmt.Printf("\n✅ Successfully initialized model project in %s\n\n", targetDir)
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit model.yaml to complete your model specification")
	fmt.Println("  2. Implement your inference logic in inference/inference.py")
	fmt.Println("  3. Add your dependencies to inference/requirements.txt")
	fmt.Println("  4. Test your model locally")
	fmt.Printf("  5. Validate with: conduit validate %s\n", filepath.Join(targetDir, "model.yaml"))

	return nil
}

func checkDirectory(dir string) error {
	if dir == "." {
		// Check if current directory has model.yaml
		if _, err := os.Stat("model.yaml"); err == nil {
			return fmt.Errorf("model.yaml already exists in current directory (use --force to overwrite)")
		}
		return nil
	}

	// Check if directory exists
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists and is not a directory", dir)
		}
		// Check if directory is empty
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("directory %s is not empty (use --force to proceed anyway)", dir)
		}
	}

	return nil
}

func promptModelInfo() (*types.Model, error) {
	var qs = []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Model name (lowercase-with-hyphens):",
				Help:    "e.g., alphafold2-multimer, esm2-large, protbert",
			},
			Validate: survey.Required,
		},
		{
			Name: "version",
			Prompt: &survey.Input{
				Message: "Version (semver):",
				Default: "1.0.0",
			},
			Validate: survey.Required,
		},
		{
			Name: "domain",
			Prompt: &survey.Select{
				Message: "Domain:",
				Options: []string{
					"protein-science",
					"genomics",
					"drug-discovery",
					"molecular-dynamics",
					"bioinformatics",
					"computational-biology",
					"other",
				},
			},
		},
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "Short description:",
			},
			Validate: survey.Required,
		},
		{
			Name: "framework",
			Prompt: &survey.Select{
				Message: "ML Framework:",
				Options: []string{"pytorch", "tensorflow", "jax", "onnx"},
				Default: "pytorch",
			},
		},
		{
			Name: "python_version",
			Prompt: &survey.Input{
				Message: "Python version:",
				Default: "3.11",
			},
		},
		{
			Name: "weights_uri",
			Prompt: &survey.Input{
				Message: "Weights URI (s3://, hf://, or https://):",
				Help:    "e.g., s3://mybucket/weights/, hf://username/model, https://example.com/weights.tar.gz",
			},
			Validate: survey.Required,
		},
		{
			Name: "gpu_required",
			Prompt: &survey.Confirm{
				Message: "Requires GPU?",
				Default: true,
			},
		},
		{
			Name: "recommended_instance",
			Prompt: &survey.Input{
				Message: "Recommended AWS instance type:",
				Default: "ml.g5.2xlarge",
			},
		},
	}

	answers := struct {
		Name                string
		Version             string
		Domain              string
		Description         string
		Framework           string
		PythonVersion       string `survey:"python_version"`
		WeightsURI          string `survey:"weights_uri"`
		GPURequired         bool   `survey:"gpu_required"`
		RecommendedInstance string `survey:"recommended_instance"`
	}{}

	if err := survey.Ask(qs, &answers); err != nil {
		return nil, err
	}

	// Handle "other" domain
	if answers.Domain == "other" {
		var customDomain string
		if err := survey.AskOne(&survey.Input{
			Message: "Enter custom domain:",
		}, &customDomain); err != nil {
			return nil, err
		}
		answers.Domain = customDomain
	}

	return &types.Model{
		Name:        answers.Name,
		Version:     answers.Version,
		Domain:      answers.Domain,
		Description: answers.Description,
		Runtime: types.Runtime{
			Framework:     answers.Framework,
			PythonVersion: answers.PythonVersion,
			Dependencies:  "requirements.txt",
		},
		Inference: types.Inference{
			Entrypoint: "inference.py",
			Handler:    "predict",
		},
		WeightsURI: answers.WeightsURI,
		Hardware: types.Hardware{
			GPURequired:         answers.GPURequired,
			RecommendedInstance: answers.RecommendedInstance,
			MinCPU:              4,
			MinMemoryGB:         16,
		},
	}, nil
}

func createDirectoryStructure(baseDir string) error {
	dirs := []string{
		baseDir,
		filepath.Join(baseDir, "inference"),
		filepath.Join(baseDir, "weights"),
		filepath.Join(baseDir, "tests"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil { //nolint:gosec // Standard permissions for generated directories
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create .gitkeep for empty directories
	gitkeepPath := filepath.Join(baseDir, "weights", ".gitkeep")
	if err := os.WriteFile(gitkeepPath, []byte(""), 0644); err != nil { //nolint:gosec // Standard permissions for generated files
		return fmt.Errorf("failed to create .gitkeep: %w", err)
	}

	return nil
}

func generateModelYAML(baseDir string, m *types.Model) error {
	yamlPath := filepath.Join(baseDir, "model.yaml")

	// Marshal to YAML
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal model to YAML: %w", err)
	}

	// Add header comment
	header := `# Conduit Model Specification
# See https://github.com/scttfrdmn/conduit for documentation

`
	fullContent := header + string(data)

	if err := os.WriteFile(yamlPath, []byte(fullContent), 0644); err != nil { //nolint:gosec // Standard permissions for generated files
		return fmt.Errorf("failed to write model.yaml: %w", err)
	}

	fmt.Printf("✓ Created %s\n", yamlPath)
	return nil
}

func createTemplateFiles(baseDir string, m *types.Model) error {
	// Create inference.py
	inferencePath := filepath.Join(baseDir, "inference", "inference.py")
	inferenceContent := generateInferenceTemplate(m)
	if err := os.WriteFile(inferencePath, []byte(inferenceContent), 0644); err != nil { //nolint:gosec // Standard permissions for generated files
		return fmt.Errorf("failed to write inference.py: %w", err)
	}
	fmt.Printf("✓ Created %s\n", inferencePath)

	// Create requirements.txt
	reqPath := filepath.Join(baseDir, "inference", "requirements.txt")
	reqContent := generateRequirementsTemplate(m)
	if err := os.WriteFile(reqPath, []byte(reqContent), 0644); err != nil { //nolint:gosec // Standard permissions for generated files
		return fmt.Errorf("failed to write requirements.txt: %w", err)
	}
	fmt.Printf("✓ Created %s\n", reqPath)

	// Create test file
	testPath := filepath.Join(baseDir, "tests", "test_inference.py")
	testContent := generateTestTemplate(m)
	if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil { //nolint:gosec // Standard permissions for generated files
		return fmt.Errorf("failed to write test_inference.py: %w", err)
	}
	fmt.Printf("✓ Created %s\n", testPath)

	// Create README.md
	readmePath := filepath.Join(baseDir, "README.md")
	readmeContent := generateReadmeTemplate(m)
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil { //nolint:gosec // Standard permissions for generated files
		return fmt.Errorf("failed to write README.md: %w", err)
	}
	fmt.Printf("✓ Created %s\n", readmePath)

	return nil
}

func generateInferenceTemplate(m *types.Model) string {
	return fmt.Sprintf(`"""
%s - Inference Implementation

Generated by Conduit
"""

import os
from typing import Any, Dict


def predict(input_data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Main prediction function.

    Args:
        input_data: Dictionary containing input data for prediction

    Returns:
        Dictionary containing prediction results
    """
    # TODO: Implement your inference logic here

    # Example structure:
    # 1. Load model weights from weights_dir
    # 2. Preprocess input_data
    # 3. Run inference
    # 4. Postprocess results
    # 5. Return predictions

    return {
        "status": "not_implemented",
        "message": "Please implement the predict() function"
    }


def load_model(weights_dir: str):
    """
    Load model weights.

    Args:
        weights_dir: Path to directory containing model weights

    Returns:
        Loaded model
    """
    # TODO: Implement model loading logic
    pass


if __name__ == "__main__":
    # Test the inference function
    test_input = {
        "example_key": "example_value"
    }

    result = predict(test_input)
    print(f"Result: {result}")
`, m.Description)
}

func generateRequirementsTemplate(m *types.Model) string {
	var reqs []string

	switch m.Runtime.Framework {
	case "pytorch":
		reqs = []string{
			"torch>=2.0.0",
			"numpy>=1.24.0",
		}
	case "tensorflow":
		reqs = []string{
			"tensorflow>=2.13.0",
			"numpy>=1.24.0",
		}
	case "jax":
		reqs = []string{
			"jax[cpu]>=0.4.0",
			"jaxlib>=0.4.0",
			"numpy>=1.24.0",
		}
	case "onnx":
		reqs = []string{
			"onnxruntime>=1.15.0",
			"numpy>=1.24.0",
		}
	default:
		reqs = []string{"numpy>=1.24.0"}
	}

	return strings.Join(reqs, "\n") + "\n"
}

func generateTestTemplate(m *types.Model) string {
	return fmt.Sprintf(`"""
Test suite for %s

Generated by Conduit
"""

import unittest
import sys
import os

# Add inference directory to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'inference'))

from inference import predict


class TestInference(unittest.TestCase):
    """Test cases for inference function."""

    def test_predict_basic(self):
        """Test basic prediction functionality."""
        input_data = {
            "example_key": "example_value"
        }

        result = predict(input_data)

        # TODO: Add your assertions here
        self.assertIsNotNone(result)
        self.assertIsInstance(result, dict)

    def test_predict_with_invalid_input(self):
        """Test prediction with invalid input."""
        # TODO: Implement error handling tests
        pass


if __name__ == "__main__":
    unittest.main()
`, m.Name)
}

func generateReadmeTemplate(m *types.Model) string {
	return fmt.Sprintf(`# %s

%s

## Overview

- **Version**: %s
- **Domain**: %s
- **Framework**: %s
- **Python Version**: %s
- **GPU Required**: %v

## Installation

### Prerequisites

- Python %s or higher
- %s

### Setup

`+"```bash"+`
# Install dependencies
pip install -r inference/requirements.txt

# Download weights
# TODO: Add instructions for downloading model weights to weights/
`+"```"+`

## Usage

### Local Inference

`+"```python"+`
from inference.inference import predict

input_data = {
    # TODO: Add example input structure
}

result = predict(input_data)
print(result)
`+"```"+`

### Testing

`+"```bash"+`
python -m pytest tests/
`+"```"+`

## Deployment

### Using Conduit

`+"```bash"+`
# Validate your model specification
conduit validate model.yaml

# Deploy to AWS Bedrock
conduit deploy

# Publish to Conduit catalog
conduit publish
`+"```"+`

## Model Details

### Hardware Requirements

- **Recommended Instance**: %s
- **Minimum CPU**: %d cores
- **Minimum Memory**: %d GB

### Weights

Weights are stored at: `+"`%s`"+`

## Contributing

TODO: Add contribution guidelines

## License

TODO: Add license information

## Citation

TODO: Add citation information if applicable

---

Generated by [Conduit](https://github.com/scttfrdmn/conduit)
`,
		m.Name,
		m.Description,
		m.Version,
		m.Domain,
		m.Runtime.Framework,
		m.Runtime.PythonVersion,
		m.Hardware.GPURequired,
		m.Runtime.PythonVersion,
		m.Runtime.Framework,
		m.Hardware.RecommendedInstance,
		m.Hardware.MinCPU,
		m.Hardware.MinMemoryGB,
		m.WeightsURI,
	)
}
