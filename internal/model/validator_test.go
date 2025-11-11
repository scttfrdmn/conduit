package model

import (
	"strings"
	"testing"

	"github.com/scttfrdmn/conduit/pkg/types"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		model   *types.Model
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid model",
			model: &types.Model{
				Name:        "valid-model",
				Version:     "1.0.0",
				Domain:      "test",
				Description: "A valid test model",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.11",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				Hardware: types.Hardware{
					RecommendedInstance: "ml.g5.2xlarge",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			model: &types.Model{
				Version:     "1.0.0",
				Domain:      "test",
				Description: "Missing name",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.11",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				Hardware: types.Hardware{
					RecommendedInstance: "ml.g5.2xlarge",
				},
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "missing version",
			model: &types.Model{
				Name:        "test-model",
				Domain:      "test",
				Description: "Missing version",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.11",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				Hardware: types.Hardware{
					RecommendedInstance: "ml.g5.2xlarge",
				},
			},
			wantErr: true,
			errMsg:  "version",
		},
		{
			name: "invalid name format",
			model: &types.Model{
				Name:        "Invalid_Model_Name",
				Version:     "1.0.0",
				Domain:      "test",
				Description: "Invalid name",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.11",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				Hardware: types.Hardware{
					RecommendedInstance: "ml.g5.2xlarge",
				},
			},
			wantErr: true,
			errMsg:  "lowercase with hyphens",
		},
		{
			name: "invalid version format",
			model: &types.Model{
				Name:        "test-model",
				Version:     "not-a-version",
				Domain:      "test",
				Description: "Invalid version",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.11",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				Hardware: types.Hardware{
					RecommendedInstance: "ml.g5.2xlarge",
				},
			},
			wantErr: true,
			errMsg:  "semantic version",
		},
		{
			name: "missing runtime framework",
			model: &types.Model{
				Name:        "test-model",
				Version:     "1.0.0",
				Domain:      "test",
				Description: "Missing framework",
				Runtime: types.Runtime{
					PythonVersion: "3.11",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				Hardware: types.Hardware{
					RecommendedInstance: "ml.g5.2xlarge",
				},
			},
			wantErr: true,
			errMsg:  "runtime.framework",
		},
		{
			name: "invalid framework",
			model: &types.Model{
				Name:        "test-model",
				Version:     "1.0.0",
				Domain:      "test",
				Description: "Invalid framework",
				Runtime: types.Runtime{
					Framework:     "invalid-framework",
					PythonVersion: "3.11",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				Hardware: types.Hardware{
					RecommendedInstance: "ml.g5.2xlarge",
				},
			},
			wantErr: true,
			errMsg:  "must be one of",
		},
		{
			name: "negative memory",
			model: &types.Model{
				Name:        "test-model",
				Version:     "1.0.0",
				Domain:      "test",
				Description: "Negative memory",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.11",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				Hardware: types.Hardware{
					RecommendedInstance: "ml.g5.2xlarge",
					MinMemoryGB:         -1,
				},
			},
			wantErr: true,
			errMsg:  "non-negative",
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidator_ValidateURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{"valid s3", "s3://bucket/path/to/weights", false},
		{"valid https", "https://example.com/weights.tar.gz", false},
		{"valid http", "http://example.com/weights.tar.gz", false},
		{"valid hf", "hf://org/model", false},
		{"invalid s3", "s3://", true},
		{"invalid url", "not-a-url", true},
		{"empty", "", true},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid simple", "model", true},
		{"valid with hyphen", "my-model", true},
		{"valid with number", "model2", true},
		{"valid complex", "alphafold2-multimer", true},
		{"invalid uppercase", "MyModel", false},
		{"invalid underscore", "my_model", false},
		{"invalid space", "my model", false},
		{"invalid start hyphen", "-model", false},
		{"invalid end hyphen", "model-", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidName(tt.input)
			if got != tt.want {
				t.Errorf("isValidName(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid simple", "1.0.0", true},
		{"valid with v", "v1.0.0", true},
		{"valid two part", "1.0", true},
		{"valid with v two part", "v2.3", true},
		{"valid with prerelease", "1.0.0-alpha", true},
		{"valid with prerelease and v", "v1.0.0-beta.1", true},
		{"invalid letters", "version", false},
		{"invalid format", "1", false},
		{"invalid dots", "1..0", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidVersion(tt.input)
			if got != tt.want {
				t.Errorf("isValidVersion(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
