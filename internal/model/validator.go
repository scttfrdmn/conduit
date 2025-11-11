package model

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/scttfrdmn/conduit/pkg/types"
)

// Validator validates Model structs
type Validator struct {
	// Can add configuration options here if needed
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("found %d validation errors:\n", len(e)))
	for _, err := range e {
		sb.WriteString(fmt.Sprintf("  - %s\n", err.Error()))
	}
	return sb.String()
}

// Validate validates a Model struct
func (v *Validator) Validate(model *types.Model) error {
	var errors ValidationErrors

	// Validate required fields
	if model.Name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "required field missing"})
	}

	if model.Version == "" {
		errors = append(errors, ValidationError{Field: "version", Message: "required field missing"})
	}

	if model.Domain == "" {
		errors = append(errors, ValidationError{Field: "domain", Message: "required field missing"})
	}

	if model.Description == "" {
		errors = append(errors, ValidationError{Field: "description", Message: "required field missing"})
	}

	// Validate name format (lowercase, hyphens only)
	if model.Name != "" {
		if !isValidName(model.Name) {
			errors = append(errors, ValidationError{
				Field:   "name",
				Message: "must be lowercase with hyphens only (e.g., 'alphafold2-multimer')",
			})
		}
	}

	// Validate version format (semver-like)
	if model.Version != "" {
		if !isValidVersion(model.Version) {
			errors = append(errors, ValidationError{
				Field:   "version",
				Message: "must be semantic version format (e.g., '1.0.0', 'v2.3.1')",
			})
		}
	}

	// Validate runtime
	if err := v.validateRuntime(&model.Runtime); err != nil {
		errors = append(errors, err...)
	}

	// Validate inference
	if err := v.validateInference(&model.Inference); err != nil {
		errors = append(errors, err...)
	}

	// Validate weights URI if present
	if model.WeightsURI != "" {
		if err := v.validateURI(model.WeightsURI); err != nil {
			errors = append(errors, ValidationError{
				Field:   "weights_uri",
				Message: fmt.Sprintf("invalid URI: %v", err),
			})
		}
	}

	// Validate hardware
	if err := v.validateHardware(&model.Hardware); err != nil {
		errors = append(errors, err...)
	}

	// Validate benchmarks
	for i, benchmark := range model.Benchmarks {
		if err := v.validateBenchmark(&benchmark, i); err != nil {
			errors = append(errors, err...)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (v *Validator) validateRuntime(runtime *types.Runtime) ValidationErrors {
	var errors ValidationErrors

	if runtime.Framework == "" {
		errors = append(errors, ValidationError{
			Field:   "runtime.framework",
			Message: "required field missing",
		})
	}

	validFrameworks := []string{"pytorch", "tensorflow", "jax", "onnx", "scikit-learn", "custom"}
	if runtime.Framework != "" && !contains(validFrameworks, runtime.Framework) {
		errors = append(errors, ValidationError{
			Field:   "runtime.framework",
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validFrameworks, ", ")),
		})
	}

	if runtime.PythonVersion == "" {
		errors = append(errors, ValidationError{
			Field:   "runtime.python_version",
			Message: "required field missing",
		})
	}

	return errors
}

func (v *Validator) validateInference(inference *types.Inference) ValidationErrors {
	var errors ValidationErrors

	if inference.Entrypoint == "" {
		errors = append(errors, ValidationError{
			Field:   "inference.entrypoint",
			Message: "required field missing",
		})
	}

	if inference.Handler == "" {
		errors = append(errors, ValidationError{
			Field:   "inference.handler",
			Message: "required field missing",
		})
	}

	return errors
}

func (v *Validator) validateHardware(hardware *types.Hardware) ValidationErrors {
	var errors ValidationErrors

	if hardware.RecommendedInstance == "" {
		errors = append(errors, ValidationError{
			Field:   "hardware.recommended_instance",
			Message: "required field missing",
		})
	}

	if hardware.MinMemoryGB < 0 {
		errors = append(errors, ValidationError{
			Field:   "hardware.min_memory_gb",
			Message: "must be non-negative",
		})
	}

	if hardware.MinGPUMemoryGB < 0 {
		errors = append(errors, ValidationError{
			Field:   "hardware.min_gpu_memory_gb",
			Message: "must be non-negative",
		})
	}

	return errors
}

func (v *Validator) validateBenchmark(benchmark *types.Benchmark, index int) ValidationErrors {
	var errors ValidationErrors
	prefix := fmt.Sprintf("benchmarks[%d]", index)

	if benchmark.Dataset == "" {
		errors = append(errors, ValidationError{
			Field:   prefix + ".dataset",
			Message: "required field missing",
		})
	}

	if benchmark.Metric == "" {
		errors = append(errors, ValidationError{
			Field:   prefix + ".metric",
			Message: "required field missing",
		})
	}

	if benchmark.Result < 0 {
		errors = append(errors, ValidationError{
			Field:   prefix + ".result",
			Message: "must be non-negative",
		})
	}

	return errors
}

func (v *Validator) validateURI(uri string) error {
	// Check if it's a valid URL or S3 URI
	if strings.HasPrefix(uri, "s3://") {
		// Basic S3 URI validation
		if len(uri) < 6 {
			return fmt.Errorf("invalid S3 URI")
		}
		return nil
	}

	if strings.HasPrefix(uri, "hf://") {
		// Basic HuggingFace URI validation
		if len(uri) < 6 {
			return fmt.Errorf("invalid HuggingFace URI")
		}
		return nil
	}

	// Validate as URL
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	return nil
}

// Helper functions

func isValidName(name string) bool {
	// Name must be lowercase letters, numbers, and hyphens only
	match, _ := regexp.MatchString(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, name)
	return match
}

func isValidVersion(version string) bool {
	// Accept semver format with or without 'v' prefix
	match, _ := regexp.MatchString(`^v?\d+\.\d+(\.\d+)?(-[a-zA-Z0-9.-]+)?$`, version)
	return match
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
