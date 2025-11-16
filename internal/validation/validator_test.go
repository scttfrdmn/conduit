package validation

import (
	"testing"

	"github.com/scttfrdmn/conduit/pkg/types"
)

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name          string
		model         *types.Model
		expectedError int
	}{
		{
			name: "Valid model with all required fields",
			model: &types.Model{
				Name:       "test-model",
				Version:    "1.0.0",
				Domain:     "test-domain",
				WeightsURI: "s3://bucket/weights",
			},
			expectedError: 0,
		},
		{
			name: "Missing name",
			model: &types.Model{
				Version:    "1.0.0",
				Domain:     "test-domain",
				WeightsURI: "s3://bucket/weights",
			},
			expectedError: 1,
		},
		{
			name: "Missing version",
			model: &types.Model{
				Name:       "test-model",
				Domain:     "test-domain",
				WeightsURI: "s3://bucket/weights",
			},
			expectedError: 1,
		},
		{
			name: "Missing domain",
			model: &types.Model{
				Name:       "test-model",
				Version:    "1.0.0",
				WeightsURI: "s3://bucket/weights",
			},
			expectedError: 1,
		},
		{
			name: "Missing weights_uri",
			model: &types.Model{
				Name:    "test-model",
				Version: "1.0.0",
				Domain:  "test-domain",
			},
			expectedError: 1,
		},
		{
			name:          "All required fields missing",
			model:         &types.Model{},
			expectedError: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			validateRequired(tt.model, result)

			if len(result.Errors) != tt.expectedError {
				t.Errorf("Expected %d errors, got %d", tt.expectedError, len(result.Errors))
			}
		})
	}
}

func TestValidateRuntime(t *testing.T) {
	tests := []struct {
		name            string
		model           *types.Model
		expectedErrors  int
		expectedWarnings int
	}{
		{
			name: "Valid runtime with common framework",
			model: &types.Model{
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.9",
				},
			},
			expectedErrors:  0,
			expectedWarnings: 0,
		},
		{
			name: "Missing framework",
			model: &types.Model{
				Runtime: types.Runtime{
					PythonVersion: "3.9",
				},
			},
			expectedErrors:  1,
			expectedWarnings: 0,
		},
		{
			name: "Missing python version",
			model: &types.Model{
				Runtime: types.Runtime{
					Framework: "pytorch",
				},
			},
			expectedErrors:  1,
			expectedWarnings: 0,
		},
		{
			name: "Uncommon framework",
			model: &types.Model{
				Runtime: types.Runtime{
					Framework:     "custom-framework",
					PythonVersion: "3.9",
				},
			},
			expectedErrors:  0,
			expectedWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			validateRuntime(tt.model, result)

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(result.Errors))
			}

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
			}
		})
	}
}

func TestValidateInference(t *testing.T) {
	tests := []struct {
		name          string
		model         *types.Model
		expectedError int
	}{
		{
			name: "Valid inference configuration",
			model: &types.Model{
				Inference: types.Inference{
					Entrypoint: "predict.py",
					Handler:    "predict",
				},
			},
			expectedError: 0,
		},
		{
			name: "Missing entrypoint",
			model: &types.Model{
				Inference: types.Inference{
					Handler: "predict",
				},
			},
			expectedError: 1,
		},
		{
			name: "Missing handler",
			model: &types.Model{
				Inference: types.Inference{
					Entrypoint: "predict.py",
				},
			},
			expectedError: 1,
		},
		{
			name: "Both missing",
			model: &types.Model{
				Inference: types.Inference{},
			},
			expectedError: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			validateInference(tt.model, result)

			if len(result.Errors) != tt.expectedError {
				t.Errorf("Expected %d errors, got %d", tt.expectedError, len(result.Errors))
			}
		})
	}
}

func TestValidateHardware(t *testing.T) {
	tests := []struct {
		name             string
		model            *types.Model
		expectedWarnings int
	}{
		{
			name: "GPU required with memory specified",
			model: &types.Model{
				Hardware: types.Hardware{
					GPURequired:    true,
					MinGPUMemoryGB: 8,
					MinMemoryGB:    16,
				},
			},
			expectedWarnings: 0,
		},
		{
			name: "GPU required without memory specified",
			model: &types.Model{
				Hardware: types.Hardware{
					GPURequired: true,
					MinMemoryGB: 16,
				},
			},
			expectedWarnings: 1,
		},
		{
			name: "No memory specified",
			model: &types.Model{
				Hardware: types.Hardware{
					GPURequired: false,
				},
			},
			expectedWarnings: 1,
		},
		{
			name: "Both warnings",
			model: &types.Model{
				Hardware: types.Hardware{
					GPURequired: true,
				},
			},
			expectedWarnings: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			validateHardware(tt.model, result)

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
			}
		})
	}
}

func TestValidateBestPractices(t *testing.T) {
	tests := []struct {
		name             string
		model            *types.Model
		expectedWarnings int
	}{
		{
			name: "Complete model with best practices",
			model: &types.Model{
				Description:    "A test model",
				License:        "Apache-2.0",
				GitHubRepo:     "github.com/org/repo",
				Tags:           []string{"test", "ml"},
				ChecksumSHA256: "abc123",
				Benchmarks: []types.Benchmark{
					{Dataset: "test", Metric: "accuracy", Result: 0.95},
				},
				Citation: types.Citation{
					PaperTitle: "Test Paper",
				},
			},
			expectedWarnings: 0,
		},
		{
			name: "Missing description",
			model: &types.Model{
				License:        "Apache-2.0",
				GitHubRepo:     "github.com/org/repo",
				Tags:           []string{"test"},
				ChecksumSHA256: "abc123",
				Benchmarks: []types.Benchmark{
					{Dataset: "test", Metric: "accuracy", Result: 0.95},
				},
				Citation: types.Citation{
					PaperTitle: "Test Paper",
				},
			},
			expectedWarnings: 1,
		},
		{
			name: "Missing all best practices",
			model: &types.Model{
				Name: "test-model",
			},
			expectedWarnings: 7, // description, license, github_repo, tags, benchmarks, citation, checksum
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			validateBestPractices(tt.model, result)

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
			}
		})
	}
}

func TestValidateVersionFormat(t *testing.T) {
	tests := []struct {
		name             string
		version          string
		expectedWarnings int
	}{
		{"Valid semantic version", "1.0.0", 0},
		{"Valid with v prefix", "v1.0.0", 0},
		{"Valid with suffix", "1.0.0-beta", 0},
		{"Valid with complex suffix", "1.2.3-alpha.1", 0},
		{"Invalid - missing patch", "1.0", 1},
		{"Invalid - no dots", "100", 1},
		{"Invalid - text only", "latest", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			model := &types.Model{
				Version: tt.version,
			}

			validateVersionFormat(model, result)

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
			}
		})
	}
}

func TestValidateModelBasic(t *testing.T) {
	tests := []struct {
		name          string
		model         *types.Model
		expectValid   bool
		expectedErrors int
	}{
		{
			name: "Valid minimal model",
			model: &types.Model{
				Name:       "test-model",
				Version:    "1.0.0",
				Domain:     "test",
				WeightsURI: "s3://bucket/weights",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.9",
				},
				Inference: types.Inference{
					Entrypoint: "predict.py",
					Handler:    "predict",
				},
			},
			expectValid:   true,
			expectedErrors: 0,
		},
		{
			name: "Invalid - missing required fields",
			model: &types.Model{
				Name: "test-model",
			},
			expectValid:   false,
			expectedErrors: 7, // version, domain, weights_uri, framework, python_version, entrypoint, handler
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateModel(tt.model, ValidationBasic)

			if result.IsValid() != tt.expectValid {
				t.Errorf("Expected IsValid()=%v, got %v", tt.expectValid, result.IsValid())
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(result.Errors))
				for _, err := range result.Errors {
					t.Logf("  - [%s] %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidateModelStrict(t *testing.T) {
	tests := []struct {
		name             string
		model            *types.Model
		expectValid      bool
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "Valid complete model",
			model: &types.Model{
				Name:           "test-model",
				Version:        "1.0.0",
				Domain:         "test",
				Description:    "A test model",
				License:        "Apache-2.0",
				GitHubRepo:     "github.com/org/repo",
				WeightsURI:     "https://example.com/weights",
				ChecksumSHA256: "abc123",
				Tags:           []string{"test"},
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.9",
				},
				Inference: types.Inference{
					Entrypoint: "https://example.com/predict.py",
					Handler:    "predict",
				},
				Benchmarks: []types.Benchmark{
					{Dataset: "test", Metric: "accuracy", Result: 0.95},
				},
				Citation: types.Citation{
					PaperTitle: "Test Paper",
					PaperURL:   "https://example.com/paper",
				},
			},
			expectValid:      true,
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "Valid but missing best practices",
			model: &types.Model{
				Name:       "test-model",
				Version:    "1.0.0",
				Domain:     "test",
				WeightsURI: "https://example.com/weights",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.9",
				},
				Inference: types.Inference{
					Entrypoint: "https://example.com/predict.py",
					Handler:    "predict",
				},
			},
			expectValid:      true,
			expectedErrors:   0,
			expectedWarnings: 6, // description, license, github_repo, tags, benchmarks, citation, checksum
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateModel(tt.model, ValidationStrict)

			if result.IsValid() != tt.expectValid {
				t.Errorf("Expected IsValid()=%v, got %v", tt.expectValid, result.IsValid())
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(result.Errors))
				for _, err := range result.Errors {
					t.Logf("  ERROR - [%s] %s", err.Field, err.Message)
				}
			}

			if len(result.Warnings) < tt.expectedWarnings {
				t.Errorf("Expected at least %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
				for _, warn := range result.Warnings {
					t.Logf("  WARNING - [%s] %s", warn.Field, warn.Message)
				}
			}
		})
	}
}
