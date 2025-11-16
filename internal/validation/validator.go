package validation

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/scttfrdmn/conduit/pkg/types"
)

// ValidationLevel defines the strictness of validation
type ValidationLevel int

const (
	// ValidationBasic performs basic required field checks
	ValidationBasic ValidationLevel = iota
	// ValidationStrict performs comprehensive validation including best practices
	ValidationStrict
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Level   string // "error" or "warning"
}

// ValidationResult contains all validation errors and warnings
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
}

// IsValid returns true if there are no errors
func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// ValidateModel validates a model according to the specified level
func ValidateModel(model *types.Model, level ValidationLevel) *ValidationResult {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Basic validation (always performed)
	validateRequired(model, result)
	validateRuntime(model, result)
	validateInference(model, result)
	validateHardware(model, result)

	// Strict validation (additional checks)
	if level == ValidationStrict {
		validateBestPractices(model, result)
		validateFiles(model, result)
		validateURLs(model, result)
		validateVersionFormat(model, result)
	}

	return result
}

// validateRequired checks required fields
func validateRequired(model *types.Model, result *ValidationResult) {
	if model.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Message: "model name is required",
			Level:   "error",
		})
	}

	if model.Version == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "version",
			Message: "version is required",
			Level:   "error",
		})
	}

	if model.Domain == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "domain",
			Message: "domain is required",
			Level:   "error",
		})
	}

	if model.WeightsURI == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "weights_uri",
			Message: "weights_uri is required",
			Level:   "error",
		})
	}
}

// validateRuntime checks runtime configuration
func validateRuntime(model *types.Model, result *ValidationResult) {
	if model.Runtime.Framework == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "runtime.framework",
			Message: "framework is required",
			Level:   "error",
		})
	}

	validFrameworks := []string{"pytorch", "tensorflow", "jax", "onnx", "keras", "scikit-learn"}
	if model.Runtime.Framework != "" && !contains(validFrameworks, strings.ToLower(model.Runtime.Framework)) {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "runtime.framework",
			Message: fmt.Sprintf("framework %q is not a common framework (expected: %v)", model.Runtime.Framework, validFrameworks),
			Level:   "warning",
		})
	}

	if model.Runtime.PythonVersion == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "runtime.python_version",
			Message: "python_version is required",
			Level:   "error",
		})
	}
}

// validateInference checks inference configuration
func validateInference(model *types.Model, result *ValidationResult) {
	if model.Inference.Entrypoint == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "inference.entrypoint",
			Message: "entrypoint is required",
			Level:   "error",
		})
	}

	if model.Inference.Handler == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "inference.handler",
			Message: "handler is required",
			Level:   "error",
		})
	}
}

// validateHardware checks hardware requirements
func validateHardware(model *types.Model, result *ValidationResult) {
	if model.Hardware.GPURequired && model.Hardware.MinGPUMemoryGB == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "hardware.min_gpu_memory_gb",
			Message: "GPU is required but min_gpu_memory_gb is not specified",
			Level:   "warning",
		})
	}

	if model.Hardware.MinMemoryGB == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "hardware.min_memory_gb",
			Message: "min_memory_gb is not specified - consider adding memory requirements",
			Level:   "warning",
		})
	}
}

// validateBestPractices checks for best practices
func validateBestPractices(model *types.Model, result *ValidationResult) {
	if model.Description == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "description",
			Message: "description is recommended for discoverability",
			Level:   "warning",
		})
	}

	if model.License == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "license",
			Message: "license is recommended for legal clarity",
			Level:   "warning",
		})
	}

	if model.GitHubRepo == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "github_repo",
			Message: "github_repo is recommended for source code access",
			Level:   "warning",
		})
	}

	if len(model.Tags) == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "tags",
			Message: "tags are recommended for categorization and search",
			Level:   "warning",
		})
	}

	if len(model.Benchmarks) == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "benchmarks",
			Message: "benchmarks are recommended to demonstrate model performance",
			Level:   "warning",
		})
	}

	if model.Citation.PaperTitle == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "citation",
			Message: "citation is recommended for academic attribution",
			Level:   "warning",
		})
	}

	if model.ChecksumSHA256 == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "checksum_sha256",
			Message: "checksum is recommended for integrity verification",
			Level:   "warning",
		})
	}
}

// validateFiles checks if referenced files exist (for local paths)
func validateFiles(model *types.Model, result *ValidationResult) {
	// Check if dependencies file exists (if it's a local path)
	if model.Runtime.Dependencies != "" && !strings.Contains(model.Runtime.Dependencies, "://") {
		if _, err := os.Stat(model.Runtime.Dependencies); os.IsNotExist(err) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "runtime.dependencies",
				Message: fmt.Sprintf("dependencies file not found: %s", model.Runtime.Dependencies),
				Level:   "error",
			})
		}
	}

	// Check if entrypoint file exists (if it's a local path)
	if model.Inference.Entrypoint != "" && !strings.Contains(model.Inference.Entrypoint, "://") {
		if _, err := os.Stat(model.Inference.Entrypoint); os.IsNotExist(err) {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "inference.entrypoint",
				Message: fmt.Sprintf("entrypoint file not found: %s", model.Inference.Entrypoint),
				Level:   "warning",
			})
		}
	}
}

// validateURLs checks if URLs are valid
func validateURLs(model *types.Model, result *ValidationResult) {
	// Validate weights URI
	if model.WeightsURI != "" {
		if _, err := url.Parse(model.WeightsURI); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "weights_uri",
				Message: fmt.Sprintf("invalid weights_uri: %v", err),
				Level:   "error",
			})
		}
	}

	// Validate GitHub repo
	if model.GitHubRepo != "" {
		if !strings.HasPrefix(model.GitHubRepo, "github.com/") && !strings.HasPrefix(model.GitHubRepo, "https://github.com/") {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "github_repo",
				Message: "github_repo should be in format 'github.com/org/repo' or 'https://github.com/org/repo'",
				Level:   "warning",
			})
		}
	}

	// Validate citation URL
	if model.Citation.PaperURL != "" {
		if _, err := url.Parse(model.Citation.PaperURL); err != nil {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "citation.paper_url",
				Message: fmt.Sprintf("invalid paper_url: %v", err),
				Level:   "warning",
			})
		}
	}
}

// validateVersionFormat checks version format (semantic versioning)
func validateVersionFormat(model *types.Model, result *ValidationResult) {
	// Semantic versioning regex: X.Y.Z or X.Y.Z-suffix
	semverRegex := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$`)

	if model.Version != "" && !semverRegex.MatchString(model.Version) {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("version %q does not follow semantic versioning (X.Y.Z)", model.Version),
			Level:   "warning",
		})
	}
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}
