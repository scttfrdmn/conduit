package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/conduit/pkg/types"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    *types.Model
		wantErr bool
	}{
		{
			name: "valid minimal model",
			yaml: `
name: test-model
version: "1.0.0"
domain: test
description: A test model

runtime:
  framework: pytorch
  python_version: "3.11"
  dependencies: requirements.txt

inference:
  entrypoint: inference.py
  handler: predict

weights_uri: s3://bucket/weights/
hardware:
  gpu_required: true
  recommended_instance: ml.g5.2xlarge
`,
			want: &types.Model{
				Name:        "test-model",
				Version:     "1.0.0",
				Domain:      "test",
				Description: "A test model",
				Runtime: types.Runtime{
					Framework:     "pytorch",
					PythonVersion: "3.11",
					Dependencies:  "requirements.txt",
				},
				Inference: types.Inference{
					Entrypoint: "inference.py",
					Handler:    "predict",
				},
				WeightsURI: "s3://bucket/weights/",
				Hardware: types.Hardware{
					GPURequired:         true,
					RecommendedInstance: "ml.g5.2xlarge",
				},
			},
			wantErr: false,
		},
		{
			name: "model with benchmarks",
			yaml: `
name: benchmark-model
version: "2.0.0"
domain: protein-science
description: Model with benchmarks

runtime:
  framework: jax
  python_version: "3.11"
  dependencies: requirements.txt

inference:
  entrypoint: run.py
  handler: main

weights_uri: https://example.com/weights.tar.gz
hardware:
  gpu_required: false
  recommended_instance: ml.c5.xlarge
  min_memory_gb: 16

benchmarks:
  - dataset: CASP15
    metric: GDT-TS
    result: 92.4
    instance: ml.g5.2xlarge
    cost_per_prediction: "$0.15"
    walltime_seconds: 120.5
`,
			want: &types.Model{
				Name:        "benchmark-model",
				Version:     "2.0.0",
				Domain:      "protein-science",
				Description: "Model with benchmarks",
				Runtime: types.Runtime{
					Framework:     "jax",
					PythonVersion: "3.11",
					Dependencies:  "requirements.txt",
				},
				Inference: types.Inference{
					Entrypoint: "run.py",
					Handler:    "main",
				},
				WeightsURI: "https://example.com/weights.tar.gz",
				Hardware: types.Hardware{
					GPURequired:         false,
					RecommendedInstance: "ml.c5.xlarge",
					MinMemoryGB:         16,
				},
				Benchmarks: []types.Benchmark{
					{
						Dataset:           "CASP15",
						Metric:            "GDT-TS",
						Result:            92.4,
						Instance:          "ml.g5.2xlarge",
						CostPerPrediction: "$0.15",
						WalltimeSeconds:   120.5,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			yaml:    `invalid: yaml: content:`,
			wantErr: true,
		},
		{
			name:    "empty yaml",
			yaml:    ``,
			want:    &types.Model{}, // Empty struct, validation will catch required fields
			wantErr: false,
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseString(tt.yaml)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if tt.want == nil {
				return
			}

			// Basic validation of parsed content
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.want.Version)
			}
			if got.Domain != tt.want.Domain {
				t.Errorf("Domain = %v, want %v", got.Domain, tt.want.Domain)
			}
		})
	}
}

func TestParser_ParseFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "model.yaml")

	content := `
name: file-test
version: "1.0.0"
domain: test
description: Test from file

runtime:
  framework: pytorch
  python_version: "3.11"
  dependencies: requirements.txt

inference:
  entrypoint: inference.py
  handler: predict

weights_uri: s3://test/weights/
hardware:
  gpu_required: true
  recommended_instance: ml.g5.xlarge
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	model, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if model.Name != "file-test" {
		t.Errorf("Name = %v, want %v", model.Name, "file-test")
	}

	if model.Version != "1.0.0" {
		t.Errorf("Version = %v, want %v", model.Version, "1.0.0")
	}
}

func TestParser_ParseFile_NotFound(t *testing.T) {
	parser := NewParser()
	_, err := parser.ParseFile("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}
