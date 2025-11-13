package catalog

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/scttfrdmn/conduit/pkg/types"
)

// ConflictStrategy defines how to handle conflicts during import
type ConflictStrategy string

const (
	ConflictSkip      ConflictStrategy = "skip"      // Skip models that already exist
	ConflictOverwrite ConflictStrategy = "overwrite" // Replace existing models
	ConflictMerge     ConflictStrategy = "merge"     // Add new versions, keep existing
)

// ImportResult tracks the results of an import operation
type ImportResult struct {
	TotalModels    int
	ImportedModels int
	SkippedModels  int
	UpdatedModels  int
	Errors         []ImportError
}

// ImportError represents an error that occurred during import
type ImportError struct {
	ModelName string
	Error     string
}

// ImportStatus tracks what happened to a single model during import
type ImportStatus int

const (
	ImportStatusCreated ImportStatus = iota
	ImportStatusSkipped
	ImportStatusUpdated
)

// ImportCatalog imports models from an export file
func (db *DB) ImportCatalog(inputPath string, strategy ConflictStrategy) (*ImportResult, error) {
	// Read export file
	export, err := readExportFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read export file: %w", err)
	}

	result := &ImportResult{
		TotalModels: len(export.Models),
	}

	// Import each model
	for _, model := range export.Models {
		status, err := db.importModel(&model, strategy)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				ModelName: model.Name,
				Error:     err.Error(),
			})
			continue
		}

		// Track result based on status
		switch status {
		case ImportStatusCreated:
			result.ImportedModels++
		case ImportStatusSkipped:
			result.SkippedModels++
		case ImportStatusUpdated:
			result.UpdatedModels++
		}
	}

	return result, nil
}

// importModel imports a single model with the specified conflict strategy
func (db *DB) importModel(model *ExportedModel, strategy ConflictStrategy) (ImportStatus, error) {
	// Check if model exists
	exists, existingID := db.modelExists(model.Name)

	if exists {
		switch strategy {
		case ConflictSkip:
			return ImportStatusSkipped, nil // Skip, no error
		case ConflictOverwrite:
			// Delete existing model and re-create
			if err := db.deleteModelByID(existingID); err != nil {
				return ImportStatusCreated, fmt.Errorf("failed to delete existing model: %w", err)
			}
			if err := db.createImportedModel(model); err != nil {
				return ImportStatusCreated, err
			}
			return ImportStatusUpdated, nil
		case ConflictMerge:
			// Add new versions to existing model
			if err := db.mergeModelVersions(existingID, model); err != nil {
				return ImportStatusUpdated, err
			}
			return ImportStatusUpdated, nil
		default:
			return ImportStatusCreated, fmt.Errorf("unknown conflict strategy: %s", strategy)
		}
	}

	// Model doesn't exist, create it
	if err := db.createImportedModel(model); err != nil {
		return ImportStatusCreated, err
	}
	return ImportStatusCreated, nil
}

// createImportedModel creates a new model from imported data
func (db *DB) createImportedModel(model *ExportedModel) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Insert model
	result, err := tx.Exec(`
		INSERT INTO models (name, domain, description, github_repo, license, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, model.Name, model.Domain, model.Description, model.GitHubRepo, model.License,
		model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert model: %w", err)
	}

	modelID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get model ID: %w", err)
	}

	// Insert versions
	for _, version := range model.Versions {
		if err := db.importVersion(tx, modelID, &version); err != nil {
			return fmt.Errorf("failed to import version %s: %w", version.Version, err)
		}
	}

	// Insert citation
	if model.Citation != nil {
		if err := db.insertCitation(tx, modelID, model.Citation); err != nil {
			return fmt.Errorf("failed to insert citation: %w", err)
		}
	}

	// Insert tags
	if len(model.Tags) > 0 {
		if err := db.insertTags(tx, modelID, model.Tags); err != nil {
			return fmt.Errorf("failed to insert tags: %w", err)
		}
	}

	// Initialize statistics
	if err := db.initializeStats(tx, modelID); err != nil {
		return fmt.Errorf("failed to initialize stats: %w", err)
	}

	return tx.Commit()
}

// importVersion imports a single version with benchmarks
func (db *DB) importVersion(tx *sql.Tx, modelID int64, version *ExportedVersion) error {
	result, err := tx.Exec(`
		INSERT INTO model_versions (
			model_id, version, weights_uri, weights_size_gb, checksum_sha256,
			framework, python_version, dependencies, custom_image,
			entrypoint, handler,
			gpu_required, recommended_instance, min_cpu, min_memory_gb, min_gpu_memory_gb,
			is_latest, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		modelID, version.Version, version.WeightsURI, version.WeightsSizeGB, version.ChecksumSHA256,
		version.Framework, version.PythonVersion, version.Dependencies, version.CustomImage,
		version.Entrypoint, version.Handler,
		version.GPURequired, version.RecommendedInstance,
		version.MinCPU, version.MinMemoryGB, version.MinGPUMemoryGB,
		version.IsLatest, version.CreatedAt,
	)
	if err != nil {
		return err
	}

	versionID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Insert benchmarks
	for _, benchmark := range version.Benchmarks {
		b := types.Benchmark{
			Dataset:           benchmark.Dataset,
			Metric:            benchmark.Metric,
			Result:            benchmark.Result,
			Instance:          benchmark.Instance,
			CostPerPrediction: benchmark.CostPerPrediction,
			WalltimeSeconds:   benchmark.WalltimeSeconds,
		}
		if err := db.insertBenchmark(tx, versionID, &b); err != nil {
			return fmt.Errorf("failed to insert benchmark: %w", err)
		}
	}

	return nil
}

// mergeModelVersions adds new versions to an existing model
func (db *DB) mergeModelVersions(modelID int64, model *ExportedModel) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Get existing version numbers
	existingVersions, err := db.getExistingVersionNumbers(modelID)
	if err != nil {
		return fmt.Errorf("failed to get existing versions: %w", err)
	}

	// Import only new versions
	addedCount := 0
	for _, version := range model.Versions {
		if !contains(existingVersions, version.Version) {
			// Clear is_latest flag if this version has it
			// We'll reassign it later
			importVersion := version
			importVersion.IsLatest = false

			if err := db.importVersion(tx, modelID, &importVersion); err != nil {
				return fmt.Errorf("failed to import version %s: %w", version.Version, err)
			}
			addedCount++
		}
	}

	// If we added versions, update the latest flag
	if addedCount > 0 {
		// Find the most recent version
		if err := db.updateLatestVersion(tx, modelID); err != nil {
			return fmt.Errorf("failed to update latest version: %w", err)
		}

		// Update model's updated_at timestamp
		_, err = tx.Exec(`
			UPDATE models SET updated_at = CURRENT_TIMESTAMP WHERE id = ?
		`, modelID)
		if err != nil {
			return fmt.Errorf("failed to update model timestamp: %w", err)
		}
	}

	return tx.Commit()
}

// modelExists checks if a model exists and returns its ID
func (db *DB) modelExists(name string) (bool, int64) {
	var id int64
	err := db.conn.QueryRow(`SELECT id FROM models WHERE name = ?`, name).Scan(&id)
	if err != nil {
		return false, 0
	}
	return true, id
}

// deleteModelByID deletes a model and all related data
func (db *DB) deleteModelByID(modelID int64) error {
	_, err := db.conn.Exec(`DELETE FROM models WHERE id = ?`, modelID)
	return err
}

// getExistingVersionNumbers returns all version numbers for a model
func (db *DB) getExistingVersionNumbers(modelID int64) ([]string, error) {
	rows, err := db.conn.Query(`
		SELECT version FROM model_versions WHERE model_id = ?
	`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // Error handling done at end with rows.Err()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// updateLatestVersion updates the is_latest flag for a model
func (db *DB) updateLatestVersion(tx *sql.Tx, modelID int64) error {
	// Clear all is_latest flags
	_, err := tx.Exec(`
		UPDATE model_versions SET is_latest = FALSE WHERE model_id = ?
	`, modelID)
	if err != nil {
		return err
	}

	// Set the most recent version as latest
	_, err = tx.Exec(`
		UPDATE model_versions
		SET is_latest = TRUE
		WHERE id = (
			SELECT id FROM model_versions
			WHERE model_id = ?
			ORDER BY created_at DESC
			LIMIT 1
		)
	`, modelID)
	return err
}

// readExportFile reads and parses an export file
func readExportFile(path string) (*CatalogExport, error) {
	file, err := os.Open(path) //nolint:gosec // Path is provided by user, legitimate file operation
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close() //nolint:errcheck // Error checked below in Decode

	var export CatalogExport
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&export); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Validate export version
	if export.Version != "1.0" {
		return nil, fmt.Errorf("unsupported export version: %s", export.Version)
	}

	return &export, nil
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
