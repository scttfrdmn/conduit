package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/conduit/pkg/types"
	"gopkg.in/yaml.v3"
)

func TestCreateDirectoryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test-model")

	err := createDirectoryStructure(testDir)
	if err != nil {
		t.Fatalf("createDirectoryStructure() error = %v", err)
	}

	// Check that all directories were created
	dirs := []string{
		testDir,
		filepath.Join(testDir, "inference"),
		filepath.Join(testDir, "weights"),
		filepath.Join(testDir, "tests"),
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Directory %s was not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}

	// Check .gitkeep exists
	gitkeepPath := filepath.Join(testDir, "weights", ".gitkeep")
	if _, err := os.Stat(gitkeepPath); err != nil {
		t.Errorf(".gitkeep was not created in weights/: %v", err)
	}
}

func TestGenerateModelYAML(t *testing.T) {
	tmpDir := t.TempDir()

	model := &types.Model{
		Name:        "test-model",
		Version:     "1.0.0",
		Domain:      "test",
		Description: "Test model",
		Runtime: types.Runtime{
			Framework:     "pytorch",
			PythonVersion: "3.11",
			Dependencies:  "requirements.txt",
		},
		Inference: types.Inference{
			Entrypoint: "inference.py",
			Handler:    "predict",
		},
		WeightsURI: "s3://test/weights",
		Hardware: types.Hardware{
			GPURequired:         true,
			RecommendedInstance: "ml.g5.xlarge",
			MinCPU:              4,
			MinMemoryGB:         16,
		},
	}

	err := generateModelYAML(tmpDir, model)
	if err != nil {
		t.Fatalf("generateModelYAML() error = %v", err)
	}

	// Check file exists
	yamlPath := filepath.Join(tmpDir, "model.yaml")
	if _, err := os.Stat(yamlPath); err != nil {
		t.Fatalf("model.yaml was not created: %v", err)
	}

	// Parse the generated YAML to verify it's valid
	data, err := os.ReadFile(yamlPath) //nolint:gosec // Test file reading from temp directory
	if err != nil {
		t.Fatalf("Failed to read model.yaml: %v", err)
	}

	var parsedModel types.Model
	if err := yaml.Unmarshal(data, &parsedModel); err != nil {
		t.Fatalf("Generated YAML is not valid: %v", err)
	}

	// Verify key fields
	if parsedModel.Name != model.Name {
		t.Errorf("Name = %v, want %v", parsedModel.Name, model.Name)
	}
	if parsedModel.Version != model.Version {
		t.Errorf("Version = %v, want %v", parsedModel.Version, model.Version)
	}
}

func TestCreateTemplateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure first
	if err := createDirectoryStructure(tmpDir); err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	model := &types.Model{
		Name:        "test-model",
		Version:     "1.0.0",
		Domain:      "test",
		Description: "Test model",
		Runtime: types.Runtime{
			Framework:     "pytorch",
			PythonVersion: "3.11",
			Dependencies:  "requirements.txt",
		},
		Inference: types.Inference{
			Entrypoint: "inference.py",
			Handler:    "predict",
		},
		WeightsURI: "s3://test/weights",
		Hardware: types.Hardware{
			GPURequired:         true,
			RecommendedInstance: "ml.g5.xlarge",
			MinCPU:              4,
			MinMemoryGB:         16,
		},
	}

	err := createTemplateFiles(tmpDir, model)
	if err != nil {
		t.Fatalf("createTemplateFiles() error = %v", err)
	}

	// Check that all files were created
	files := []string{
		filepath.Join(tmpDir, "inference", "inference.py"),
		filepath.Join(tmpDir, "inference", "requirements.txt"),
		filepath.Join(tmpDir, "tests", "test_inference.py"),
		filepath.Join(tmpDir, "README.md"),
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			t.Errorf("File %s was not created: %v", file, err)
			continue
		}
		if info.IsDir() {
			t.Errorf("%s is a directory, expected file", file)
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", file)
		}
	}
}

func TestGenerateInferenceTemplate(t *testing.T) {
	model := &types.Model{
		Name:        "test-model",
		Description: "Test model description",
		Runtime: types.Runtime{
			Framework: "pytorch",
		},
	}

	content := generateInferenceTemplate(model)

	// Check for key elements
	if content == "" {
		t.Error("Generated inference template is empty")
	}

	// Should contain the predict function
	if !containsString(content, "def predict(") {
		t.Error("Inference template missing predict function")
	}

	// Should contain the load_model function
	if !containsString(content, "def load_model(") {
		t.Error("Inference template missing load_model function")
	}

	// Should contain the description
	if !containsString(content, model.Description) {
		t.Error("Inference template missing model description")
	}
}

func TestGenerateRequirementsTemplate(t *testing.T) {
	tests := []struct {
		name      string
		framework string
		want      string
	}{
		{
			name:      "pytorch",
			framework: "pytorch",
			want:      "torch>=2.0.0",
		},
		{
			name:      "tensorflow",
			framework: "tensorflow",
			want:      "tensorflow>=2.13.0",
		},
		{
			name:      "jax",
			framework: "jax",
			want:      "jax[cpu]>=0.4.0",
		},
		{
			name:      "onnx",
			framework: "onnx",
			want:      "onnxruntime>=1.15.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &types.Model{
				Runtime: types.Runtime{
					Framework: tt.framework,
				},
			}

			content := generateRequirementsTemplate(model)

			if content == "" {
				t.Error("Generated requirements template is empty")
			}

			if !containsString(content, tt.want) {
				t.Errorf("Requirements template missing %s", tt.want)
			}

			// All should contain numpy
			if !containsString(content, "numpy") {
				t.Error("Requirements template missing numpy")
			}
		})
	}
}

func TestGenerateTestTemplate(t *testing.T) {
	model := &types.Model{
		Name: "test-model",
	}

	content := generateTestTemplate(model)

	if content == "" {
		t.Error("Generated test template is empty")
	}

	// Should contain unittest
	if !containsString(content, "unittest") {
		t.Error("Test template missing unittest import")
	}

	// Should contain test class
	if !containsString(content, "class TestInference") {
		t.Error("Test template missing TestInference class")
	}

	// Should contain model name
	if !containsString(content, model.Name) {
		t.Error("Test template missing model name")
	}
}

func TestGenerateReadmeTemplate(t *testing.T) {
	model := &types.Model{
		Name:        "test-model",
		Version:     "1.0.0",
		Domain:      "test",
		Description: "Test model description",
		Runtime: types.Runtime{
			Framework:     "pytorch",
			PythonVersion: "3.11",
		},
		WeightsURI: "s3://test/weights",
		Hardware: types.Hardware{
			GPURequired:         true,
			RecommendedInstance: "ml.g5.xlarge",
			MinCPU:              4,
			MinMemoryGB:         16,
		},
	}

	content := generateReadmeTemplate(model)

	if content == "" {
		t.Error("Generated README template is empty")
	}

	// Check for key sections
	requiredSections := []string{
		"# " + model.Name,
		model.Description,
		"## Overview",
		"## Installation",
		"## Usage",
		"## Deployment",
		"## Model Details",
		model.WeightsURI,
	}

	for _, section := range requiredSections {
		if !containsString(content, section) {
			t.Errorf("README template missing section: %s", section)
		}
	}
}

func TestCheckDirectory(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(string) error
		dir     string
		wantErr bool
	}{
		{
			name: "new directory - should pass",
			setup: func(base string) error {
				return nil
			},
			dir:     "new-model",
			wantErr: false,
		},
		{
			name: "empty existing directory - should pass",
			setup: func(base string) error {
				return os.MkdirAll(filepath.Join(base, "empty-model"), 0755) //nolint:gosec // Test directory creation
			},
			dir:     "empty-model",
			wantErr: false,
		},
		{
			name: "non-empty directory - should fail",
			setup: func(base string) error {
				dir := filepath.Join(base, "non-empty-model")
				if err := os.MkdirAll(dir, 0755); err != nil { //nolint:gosec // Test directory creation
					return err
				}
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644) //nolint:gosec // Test file creation
			},
			dir:     "non-empty-model",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := tt.setup(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			testDir := filepath.Join(tmpDir, tt.dir)
			err := checkDirectory(testDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
