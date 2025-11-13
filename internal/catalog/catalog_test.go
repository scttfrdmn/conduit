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

func TestGetModel_WithCitation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	model := &types.Model{
		Name:        "cited-model",
		Version:     "1.0.0",
		Domain:      "protein-science",
		Description: "Model with citation",
		Runtime: types.Runtime{
			Framework:     "pytorch",
			PythonVersion: "3.11",
		},
		Inference: types.Inference{
			Entrypoint: "inference.py",
			Handler:    "predict",
		},
		WeightsURI: "s3://test/weights",
		Citation: types.Citation{
			PaperTitle: "Test Paper",
			Authors:    []string{"Author A", "Author B"},
			Year:       2024,
			DOI:        "10.1234/test",
			PaperURL:   "https://example.com/paper",
			BibTeX:     "@article{test2024, title={Test Paper}}",
		},
	}

	_, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Retrieve and check citation
	retrieved, err := db.GetModel("cited-model")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}

	if retrieved.Citation == nil {
		t.Fatal("Citation is nil")
	}

	c := retrieved.Citation
	if c.PaperTitle != "Test Paper" {
		t.Errorf("PaperTitle = %v, want Test Paper", c.PaperTitle)
	}

	if c.Year != 2024 {
		t.Errorf("Year = %v, want 2024", c.Year)
	}

	if c.DOI != "10.1234/test" {
		t.Errorf("DOI = %v, want 10.1234/test", c.DOI)
	}

	// Authors should be joined by comma in database
	if c.Authors != "Author A, Author B" {
		t.Errorf("Authors = %v, want 'Author A, Author B'", c.Authors)
	}
}

func TestCreateModelVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create initial model
	model := createTestModel("versioned-model")
	model.Version = "1.0.0"

	modelID, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Create a new version
	model.Version = "2.0.0"
	model.Runtime.PythonVersion = "3.12" // Change something

	err = db.CreateModelVersion(modelID, model)
	if err != nil {
		t.Fatalf("CreateModelVersion() error = %v", err)
	}

	// Verify new version is latest
	retrieved, err := db.GetModel("versioned-model")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}

	if retrieved.LatestVersion.Version != "2.0.0" {
		t.Errorf("Latest version = %v, want 2.0.0", retrieved.LatestVersion.Version)
	}

	if retrieved.LatestVersion.PythonVersion != "3.12" {
		t.Errorf("PythonVersion = %v, want 3.12", retrieved.LatestVersion.PythonVersion)
	}

	// Try to create same version again - should fail
	err = db.CreateModelVersion(modelID, model)
	if err == nil {
		t.Error("CreateModelVersion() should fail for duplicate version")
	}
}

func TestListModelVersions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create model with multiple versions
	model := createTestModel("multi-version-model")
	model.Version = "1.0.0"

	modelID, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Add more versions
	versions := []string{"1.1.0", "2.0.0", "2.1.0"}
	for _, v := range versions {
		model.Version = v
		if err := db.CreateModelVersion(modelID, model); err != nil {
			t.Fatalf("CreateModelVersion(%s) error = %v", v, err)
		}
	}

	// List all versions
	allVersions, err := db.ListModelVersions("multi-version-model")
	if err != nil {
		t.Fatalf("ListModelVersions() error = %v", err)
	}

	// Should have 4 versions total (1.0.0 + 3 added)
	if len(allVersions) != 4 {
		t.Errorf("ListModelVersions() returned %d versions, want 4", len(allVersions))
	}

	// Verify all expected versions are present
	versionSet := make(map[string]bool)
	for _, v := range allVersions {
		versionSet[v.Version] = true
	}

	expectedVersions := []string{"1.0.0", "1.1.0", "2.0.0", "2.1.0"}
	for _, expected := range expectedVersions {
		if !versionSet[expected] {
			t.Errorf("Expected version %s not found", expected)
		}
	}

	// Only the latest should have IsLatest = true
	latestCount := 0
	var latestVersion string
	for _, v := range allVersions {
		if v.IsLatest {
			latestCount++
			latestVersion = v.Version
		}
	}

	if latestCount != 1 {
		t.Errorf("Found %d versions marked as latest, want 1", latestCount)
	}

	if latestVersion != "2.1.0" {
		t.Errorf("Latest version is %v, want 2.1.0", latestVersion)
	}
}

func TestGetModelVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create model with versions
	model := createTestModel("specific-version-model")
	model.Version = "1.0.0"

	modelID, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	model.Version = "2.0.0"
	model.Runtime.Framework = "tensorflow"
	if err := db.CreateModelVersion(modelID, model); err != nil {
		t.Fatalf("CreateModelVersion() error = %v", err)
	}

	// Get specific old version
	v1, err := db.GetModelVersion("specific-version-model", "1.0.0")
	if err != nil {
		t.Fatalf("GetModelVersion(1.0.0) error = %v", err)
	}

	if v1.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", v1.Version)
	}

	if v1.Framework != "pytorch" {
		t.Errorf("Framework = %v, want pytorch", v1.Framework)
	}

	if v1.IsLatest {
		t.Error("Version 1.0.0 should not be latest")
	}

	// Get latest version
	v2, err := db.GetModelVersion("specific-version-model", "2.0.0")
	if err != nil {
		t.Fatalf("GetModelVersion(2.0.0) error = %v", err)
	}

	if v2.Version != "2.0.0" {
		t.Errorf("Version = %v, want 2.0.0", v2.Version)
	}

	if v2.Framework != "tensorflow" {
		t.Errorf("Framework = %v, want tensorflow", v2.Framework)
	}

	if !v2.IsLatest {
		t.Error("Version 2.0.0 should be latest")
	}

	// Try to get non-existent version
	_, err = db.GetModelVersion("specific-version-model", "3.0.0")
	if err == nil {
		t.Error("GetModelVersion(3.0.0) should fail")
	}
}

func TestDeleteModelVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create model with multiple versions
	model := createTestModel("version-delete-test")
	model.Version = "1.0.0"

	modelID, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Add more versions
	model.Version = "2.0.0"
	if err := db.CreateModelVersion(modelID, model); err != nil {
		t.Fatalf("CreateModelVersion(2.0.0) error = %v", err)
	}

	model.Version = "3.0.0"
	if err := db.CreateModelVersion(modelID, model); err != nil {
		t.Fatalf("CreateModelVersion(3.0.0) error = %v", err)
	}

	// Delete middle version
	if err := db.DeleteModelVersion("version-delete-test", "2.0.0"); err != nil {
		t.Fatalf("DeleteModelVersion(2.0.0) error = %v", err)
	}

	// Verify it's gone
	_, err = db.GetModelVersion("version-delete-test", "2.0.0")
	if err == nil {
		t.Error("GetModelVersion(2.0.0) should fail after deletion")
	}

	// Verify other versions still exist
	v1, err := db.GetModelVersion("version-delete-test", "1.0.0")
	if err != nil {
		t.Fatalf("GetModelVersion(1.0.0) error = %v", err)
	}

	v3, err := db.GetModelVersion("version-delete-test", "3.0.0")
	if err != nil {
		t.Fatalf("GetModelVersion(3.0.0) error = %v", err)
	}

	// Verify latest flag is still correct (3.0.0 should be latest)
	if !v3.IsLatest {
		t.Error("Version 3.0.0 should still be latest")
	}

	if v1.IsLatest {
		t.Error("Version 1.0.0 should not be latest")
	}
}

func TestDeleteModelVersion_LastVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create model with single version
	model := createTestModel("single-version-test")
	_, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Try to delete the only version - should fail
	err = db.DeleteModelVersion("single-version-test", "1.0.0")
	if err == nil {
		t.Error("DeleteModelVersion() should fail when deleting the only version")
	}
}

func TestDeleteModelVersion_ReassignLatest(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create model with multiple versions
	model := createTestModel("latest-reassign-test")
	model.Version = "1.0.0"

	modelID, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	model.Version = "2.0.0"
	if err := db.CreateModelVersion(modelID, model); err != nil {
		t.Fatalf("CreateModelVersion(2.0.0) error = %v", err)
	}

	// 2.0.0 should be latest now
	v2, err := db.GetModelVersion("latest-reassign-test", "2.0.0")
	if err != nil {
		t.Fatalf("GetModelVersion(2.0.0) error = %v", err)
	}
	if !v2.IsLatest {
		t.Error("Version 2.0.0 should be latest")
	}

	// Delete the latest version
	if err := db.DeleteModelVersion("latest-reassign-test", "2.0.0"); err != nil {
		t.Fatalf("DeleteModelVersion(2.0.0) error = %v", err)
	}

	// 1.0.0 should now be marked as latest
	v1, err := db.GetModelVersion("latest-reassign-test", "1.0.0")
	if err != nil {
		t.Fatalf("GetModelVersion(1.0.0) error = %v", err)
	}
	if !v1.IsLatest {
		t.Error("Version 1.0.0 should be marked as latest after deleting 2.0.0")
	}
}

func TestModelWithTags(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create model with tags
	model := createTestModel("tagged-model")
	model.Tags = []string{"production", "verified", "protein-folding"}

	_, err := db.CreateModel(model)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	// Retrieve model
	retrieved, err := db.GetModel("tagged-model")
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}

	// Verify tags are loaded and sorted
	if len(retrieved.Tags) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(retrieved.Tags))
	}

	expectedTags := []string{"production", "protein-folding", "verified"}
	for i, expected := range expectedTags {
		if retrieved.Tags[i] != expected {
			t.Errorf("Tag[%d] = %v, want %v", i, retrieved.Tags[i], expected)
		}
	}

	// Test search by single tag
	results, err := db.Search(SearchOptions{
		Tags:   []string{"production"},
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Name != "tagged-model" {
		t.Errorf("Search result name = %v, want tagged-model", results[0].Name)
	}

	// Test search by multiple tags (AND logic)
	results, err = db.Search(SearchOptions{
		Tags:   []string{"production", "verified"},
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Search(multiple tags) error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result for multiple tags, got %d", len(results))
	}

	// Test search by non-existent tag
	results, err = db.Search(SearchOptions{
		Tags:   []string{"nonexistent"},
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Search(nonexistent tag) error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for nonexistent tag, got %d", len(results))
	}

	// Create model without tags
	untaggedModel := createTestModel("untagged-model")
	_, err = db.CreateModel(untaggedModel)
	if err != nil {
		t.Fatalf("CreateModel(untagged) error = %v", err)
	}

	// Verify untagged model is not returned in tag search
	results, err = db.Search(SearchOptions{
		Tags:   []string{"production"},
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Search after adding untagged model error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected only tagged model in results, got %d", len(results))
	}
}

func TestModelStatistics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create two models
	model1 := createTestModel("model-1")
	id1, err := db.CreateModel(model1)
	if err != nil {
		t.Fatalf("CreateModel(model-1) error = %v", err)
	}

	model2 := createTestModel("model-2")
	id2, err := db.CreateModel(model2)
	if err != nil {
		t.Fatalf("CreateModel(model-2) error = %v", err)
	}

	// Verify stats are initialized
	retrieved1, err := db.GetModel("model-1")
	if err != nil {
		t.Fatalf("GetModel(model-1) error = %v", err)
	}

	if retrieved1.Stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if retrieved1.Stats.TotalDeployments != 0 || retrieved1.Stats.TotalPredictions != 0 || retrieved1.Stats.ViewCount != 0 {
		t.Errorf("Initial stats should be zero, got deployments=%d, predictions=%d, views=%d",
			retrieved1.Stats.TotalDeployments, retrieved1.Stats.TotalPredictions, retrieved1.Stats.ViewCount)
	}

	// Test IncrementViewCount
	if err := db.IncrementViewCount(id1); err != nil {
		t.Fatalf("IncrementViewCount() error = %v", err)
	}

	retrieved1, err = db.GetModel("model-1")
	if err != nil {
		t.Fatalf("GetModel(model-1) after view error = %v", err)
	}

	if retrieved1.Stats.ViewCount != 1 {
		t.Errorf("ViewCount = %d, want 1", retrieved1.Stats.ViewCount)
	}

	if retrieved1.Stats.LastViewedAt.IsZero() {
		t.Error("LastViewedAt should be set after view")
	}

	// Test TrackDeployment
	if err := db.TrackDeployment(id1); err != nil {
		t.Fatalf("TrackDeployment() error = %v", err)
	}

	retrieved1, err = db.GetModel("model-1")
	if err != nil {
		t.Fatalf("GetModel(model-1) after deployment error = %v", err)
	}

	if retrieved1.Stats.TotalDeployments != 1 {
		t.Errorf("TotalDeployments = %d, want 1", retrieved1.Stats.TotalDeployments)
	}

	if retrieved1.Stats.LastDeployedAt.IsZero() {
		t.Error("LastDeployedAt should be set after deployment")
	}

	// Test TrackPrediction
	if err := db.TrackPrediction(id1, 100); err != nil {
		t.Fatalf("TrackPrediction() error = %v", err)
	}

	retrieved1, err = db.GetModel("model-1")
	if err != nil {
		t.Fatalf("GetModel(model-1) after prediction error = %v", err)
	}

	if retrieved1.Stats.TotalPredictions != 100 {
		t.Errorf("TotalPredictions = %d, want 100", retrieved1.Stats.TotalPredictions)
	}

	// Add more predictions
	if err := db.TrackPrediction(id1, 50); err != nil {
		t.Fatalf("TrackPrediction(50) error = %v", err)
	}

	retrieved1, err = db.GetModel("model-1")
	if err != nil {
		t.Fatalf("GetModel(model-1) after second prediction error = %v", err)
	}

	if retrieved1.Stats.TotalPredictions != 150 {
		t.Errorf("TotalPredictions = %d, want 150", retrieved1.Stats.TotalPredictions)
	}

	// Make model-2 more popular for sorting test
	if err := db.TrackDeployment(id2); err != nil {
		t.Fatalf("TrackDeployment(model-2) error = %v", err)
	}
	if err := db.TrackDeployment(id2); err != nil {
		t.Fatalf("TrackDeployment(model-2) #2 error = %v", err)
	}
	if err := db.TrackPrediction(id2, 200); err != nil {
		t.Fatalf("TrackPrediction(model-2) error = %v", err)
	}

	// Test sorting by popularity
	results, err := db.Search(SearchOptions{
		Limit:  10,
		Offset: 0,
		SortBy: "popular",
	})
	if err != nil {
		t.Fatalf("Search(sort by popular) error = %v", err)
	}

	if len(results) < 2 {
		t.Fatalf("Expected at least 2 results, got %d", len(results))
	}

	// model-2 should come first (2 deployments, 200 predictions)
	// model-1 should come second (1 deployment, 150 predictions)
	if results[0].Name != "model-2" {
		t.Errorf("First result = %v, want model-2 (most popular)", results[0].Name)
	}

	if results[1].Name != "model-1" {
		t.Errorf("Second result = %v, want model-1", results[1].Name)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
