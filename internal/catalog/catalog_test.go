package catalog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scttfrdmn/conduit/pkg/types"
)

func TestNewDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	if db.Path() != dbPath {
		t.Errorf("Path() = %v, want %v", db.Path(), dbPath)
	}

	if err := db.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestNewDB_DefaultPath(t *testing.T) {
	// Test with empty path (should use default)
	db, err := NewDB("")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer func() {
		dbPath := db.Path()
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
		// Clean up the default database
		if err := os.Remove(dbPath); err != nil {
			t.Errorf("Remove() error = %v", err)
		}
	}()

	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".conduit", "catalog.db")
	if db.Path() != expectedPath {
		t.Errorf("Path() = %v, want %v", db.Path(), expectedPath)
	}
}

func TestCreateAndGetModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

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

	// Create model
	id, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	if id == 0 {
		t.Error("CreateModel() returned ID = 0")
	}

	// Get model
	retrieved, err := db.GetModel("test-model")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}

	if retrieved.Name != model.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, model.Name)
	}

	if retrieved.Domain != model.Domain {
		t.Errorf("Domain = %v, want %v", retrieved.Domain, model.Domain)
	}

	// Check latest version was loaded
	if retrieved.LatestVersion == nil {
		t.Fatal("LatestVersion is nil")
	}

	if retrieved.LatestVersion.Version != model.Version {
		t.Errorf("Version = %v, want %v", retrieved.LatestVersion.Version, model.Version)
	}

	if retrieved.LatestVersion.Framework != model.Runtime.Framework {
		t.Errorf("Framework = %v, want %v", retrieved.LatestVersion.Framework, model.Runtime.Framework)
	}
}

func TestCreateModel_WithBenchmarks(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &types.Model{
		Name:        "benchmark-model",
		Version:     "1.0.0",
		Domain:      "protein-science",
		Description: "Model with benchmarks",
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
			RecommendedInstance: "ml.g5.2xlarge",
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
	}

	_, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Retrieve and check benchmarks
	retrieved, err := db.GetModel("benchmark-model")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}

	if len(retrieved.LatestVersion.Benchmarks) != 1 {
		t.Fatalf("Expected 1 benchmark, got %d", len(retrieved.LatestVersion.Benchmarks))
	}

	b := retrieved.LatestVersion.Benchmarks[0]
	if b.Dataset != "CASP15" {
		t.Errorf("Dataset = %v, want CASP15", b.Dataset)
	}

	if b.Result != 92.4 {
		t.Errorf("Result = %v, want 92.4", b.Result)
	}
}

func TestListModels(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create multiple models
	models := []string{"model-a", "model-b", "model-c"}
	for _, name := range models {
		model := createTestModel(name)
		if _, err := db.CreateModel(model); err != nil {
			t.Fatalf("CreateModel(%s) error = %v", name, err)
		}
	}

	// List all models
	results, err := db.ListModels(10, 0)
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("ListModels() returned %d models, want 3", len(results))
	}
}

func TestSearch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test models
	models := []*types.Model{
		createTestModelWithDetails("protein-fold-1", "protein-science", "pytorch", true),
		createTestModelWithDetails("protein-fold-2", "protein-science", "tensorflow", true),
		createTestModelWithDetails("genomics-tool", "genomics", "jax", false),
	}

	for _, m := range models {
		if _, err := db.CreateModel(m); err != nil {
			t.Fatalf("CreateModel() error = %v", err)
		}
	}

	tests := []struct {
		name     string
		opts     SearchOptions
		wantLen  int
		wantName string
	}{
		{
			name: "search by domain",
			opts: SearchOptions{
				Domain: "protein-science",
				Limit:  10,
			},
			wantLen: 2,
		},
		{
			name: "search by framework",
			opts: SearchOptions{
				Framework: "pytorch",
				Limit:     10,
			},
			wantLen:  1,
			wantName: "protein-fold-1",
		},
		{
			name: "search by GPU requirement",
			opts: SearchOptions{
				GPURequired: boolPtr(false),
				Limit:       10,
			},
			wantLen:  1,
			wantName: "genomics-tool",
		},
		{
			name: "search with query",
			opts: SearchOptions{
				Query: "protein",
				Limit: 10,
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.Search(tt.opts)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) != tt.wantLen {
				t.Errorf("Search() returned %d results, want %d", len(results), tt.wantLen)
			}

			if tt.wantName != "" && len(results) > 0 {
				if results[0].Name != tt.wantName {
					t.Errorf("First result name = %v, want %v", results[0].Name, tt.wantName)
				}
			}
		})
	}
}

func TestUpdateModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create model
	model := createTestModel("update-test")
	if _, err := db.CreateModel(model); err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Update description
	updates := map[string]interface{}{
		"description": "Updated description",
	}

	if err := db.UpdateModel("update-test", updates); err != nil {
		t.Fatalf("UpdateModel() error = %v", err)
	}

	// Verify update
	retrieved, err := db.GetModel("update-test")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}

	if retrieved.Description != "Updated description" {
		t.Errorf("Description = %v, want 'Updated description'", retrieved.Description)
	}
}

func TestDeleteModel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create model
	model := createTestModel("delete-test")
	if _, err := db.CreateModel(model); err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Delete model
	if err := db.DeleteModel("delete-test"); err != nil {
		t.Fatalf("DeleteModel() error = %v", err)
	}

	// Verify deletion
	_, err := db.GetModel("delete-test")
	if err == nil {
		t.Error("GetModel() should return error for deleted model")
	}
}

func TestCountModels(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Initially should be 0
	count, err := db.CountModels()
	if err != nil {
		t.Fatalf("CountModels() error = %v", err)
	}

	if count != 0 {
		t.Errorf("CountModels() = %d, want 0", count)
	}

	// Create models
	for i := 0; i < 5; i++ {
		model := createTestModel("model-" + string(rune('a'+i)))
		if _, err := db.CreateModel(model); err != nil {
			t.Fatalf("CreateModel() error = %v", err)
		}
	}

	// Count should be 5
	count, err = db.CountModels()
	if err != nil {
		t.Fatalf("CountModels() error = %v", err)
	}

	if count != 5 {
		t.Errorf("CountModels() = %d, want 5", count)
	}
}

// Helper functions

func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}

	return db, cleanup
}

func createTestModel(name string) *types.Model {
	return &types.Model{
		Name:        name,
		Version:     "1.0.0",
		Domain:      "test",
		Description: "Test model " + name,
		Runtime: types.Runtime{
			Framework:     "pytorch",
			PythonVersion: "3.11",
			Dependencies:  "requirements.txt",
		},
		Inference: types.Inference{
			Entrypoint: "inference.py",
			Handler:    "predict",
		},
		WeightsURI: "s3://test/weights/" + name,
		Hardware: types.Hardware{
			GPURequired:         true,
			RecommendedInstance: "ml.g5.xlarge",
			MinCPU:              4,
			MinMemoryGB:         16,
		},
	}
}

func createTestModelWithDetails(name, domain, framework string, gpuRequired bool) *types.Model {
	m := createTestModel(name)
	m.Domain = domain
	m.Runtime.Framework = framework
	m.Hardware.GPURequired = gpuRequired
	return m
}

func boolPtr(b bool) *bool {
	return &b
}
