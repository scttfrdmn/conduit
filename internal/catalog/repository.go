package catalog

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/scttfrdmn/conduit/pkg/types"
)

// CreateModel inserts a new model into the catalog
func (db *DB) CreateModel(m *types.Model) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // Rollback is safe to call even after commit

	// Insert model
	result, err := tx.Exec(`
		INSERT INTO models (name, domain, description, github_repo, license)
		VALUES (?, ?, ?, ?, ?)
	`, m.Name, m.Domain, m.Description, m.GitHubRepo, m.License)
	if err != nil {
		return 0, fmt.Errorf("failed to insert model: %w", err)
	}

	modelID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get model ID: %w", err)
	}

	// Insert model version
	versionID, err := db.insertModelVersion(tx, modelID, m)
	if err != nil {
		return 0, fmt.Errorf("failed to insert model version: %w", err)
	}

	// Insert benchmarks
	for _, benchmark := range m.Benchmarks {
		if err := db.insertBenchmark(tx, versionID, &benchmark); err != nil {
			return 0, fmt.Errorf("failed to insert benchmark: %w", err)
		}
	}

	// Insert citation if present
	if m.Citation.PaperTitle != "" {
		if err := db.insertCitation(tx, modelID, &m.Citation); err != nil {
			return 0, fmt.Errorf("failed to insert citation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return modelID, nil
}

// insertModelVersion inserts a model version within a transaction
func (db *DB) insertModelVersion(tx *sql.Tx, modelID int64, m *types.Model) (int64, error) {
	result, err := tx.Exec(`
		INSERT INTO model_versions (
			model_id, version, weights_uri, weights_size_gb, checksum_sha256,
			framework, python_version, dependencies, custom_image,
			entrypoint, handler,
			gpu_required, recommended_instance, min_cpu, min_memory_gb, min_gpu_memory_gb,
			is_latest
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, TRUE)
	`,
		modelID, m.Version, m.WeightsURI, m.WeightsSizeGB, m.ChecksumSHA256,
		m.Runtime.Framework, m.Runtime.PythonVersion, m.Runtime.Dependencies, m.Runtime.CustomImage,
		m.Inference.Entrypoint, m.Inference.Handler,
		m.Hardware.GPURequired, m.Hardware.RecommendedInstance,
		m.Hardware.MinCPU, m.Hardware.MinMemoryGB, m.Hardware.MinGPUMemoryGB,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// insertBenchmark inserts a benchmark within a transaction
func (db *DB) insertBenchmark(tx *sql.Tx, versionID int64, b *types.Benchmark) error {
	_, err := tx.Exec(`
		INSERT INTO benchmarks (
			model_version_id, dataset, metric, result,
			instance, cost_per_prediction, walltime_seconds
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		versionID, b.Dataset, b.Metric, b.Result,
		b.Instance, b.CostPerPrediction, b.WalltimeSeconds,
	)
	return err
}

// insertCitation inserts a citation within a transaction
func (db *DB) insertCitation(tx *sql.Tx, modelID int64, c *types.Citation) error {
	authorsStr := strings.Join(c.Authors, ", ")
	_, err := tx.Exec(`
		INSERT INTO citations (
			model_id, paper_title, paper_url, doi, authors, year, bibtex
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		modelID, c.PaperTitle, c.PaperURL, c.DOI, authorsStr, c.Year, c.BibTeX,
	)
	return err
}

// GetModel retrieves a model by name
func (db *DB) GetModel(name string) (*Model, error) {
	var m Model
	err := db.conn.QueryRow(`
		SELECT id, name, domain, description, github_repo, license, created_at, updated_at
		FROM models
		WHERE name = ?
	`, name).Scan(
		&m.ID, &m.Name, &m.Domain, &m.Description,
		&m.GitHubRepo, &m.License, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found: %s", name)
		}
		return nil, fmt.Errorf("failed to query model: %w", err)
	}

	// Load latest version
	version, err := db.getLatestVersion(m.ID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to load latest version: %w", err)
	}
	m.LatestVersion = version

	// Load citation
	citation, err := db.getCitation(m.ID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to load citation: %w", err)
	}
	m.Citation = citation

	return &m, nil
}

// getLatestVersion retrieves the latest version of a model
func (db *DB) getLatestVersion(modelID int64) (*ModelVersion, error) {
	var v ModelVersion
	err := db.conn.QueryRow(`
		SELECT id, model_id, version, weights_uri, weights_size_gb, checksum_sha256,
		       framework, python_version, dependencies, custom_image,
		       entrypoint, handler,
		       gpu_required, recommended_instance, min_cpu, min_memory_gb, min_gpu_memory_gb,
		       created_at, is_latest
		FROM model_versions
		WHERE model_id = ? AND is_latest = TRUE
		LIMIT 1
	`, modelID).Scan(
		&v.ID, &v.ModelID, &v.Version, &v.WeightsURI, &v.WeightsSizeGB, &v.ChecksumSHA256,
		&v.Framework, &v.PythonVersion, &v.Dependencies, &v.CustomImage,
		&v.Entrypoint, &v.Handler,
		&v.GPURequired, &v.RecommendedInstance, &v.MinCPU, &v.MinMemoryGB, &v.MinGPUMemoryGB,
		&v.CreatedAt, &v.IsLatest,
	)
	if err != nil {
		return nil, err
	}

	// Load benchmarks
	benchmarks, err := db.getBenchmarks(v.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load benchmarks: %w", err)
	}
	v.Benchmarks = benchmarks

	return &v, nil
}

// getBenchmarks retrieves benchmarks for a model version
func (db *DB) getBenchmarks(versionID int64) ([]Benchmark, error) {
	rows, err := db.conn.Query(`
		SELECT id, model_version_id, dataset, metric, result,
		       instance, cost_per_prediction, walltime_seconds, created_at
		FROM benchmarks
		WHERE model_version_id = ?
	`, versionID)
	if err != nil {
		return nil, err
	}

	var benchmarks []Benchmark
	for rows.Next() {
		var b Benchmark
		if err := rows.Scan(
			&b.ID, &b.ModelVersionID, &b.Dataset, &b.Metric, &b.Result,
			&b.Instance, &b.CostPerPrediction, &b.WalltimeSeconds, &b.CreatedAt,
		); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				return nil, fmt.Errorf("scan error: %w, close error: %v", err, closeErr)
			}
			return nil, err
		}
		benchmarks = append(benchmarks, b)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return benchmarks, rows.Err()
}

// getCitation retrieves citation information for a model
func (db *DB) getCitation(modelID int64) (*Citation, error) {
	var c Citation
	err := db.conn.QueryRow(`
		SELECT id, model_id, paper_title, paper_url, doi, authors, year, bibtex
		FROM citations
		WHERE model_id = ?
	`, modelID).Scan(
		&c.ID, &c.ModelID, &c.PaperTitle, &c.PaperURL,
		&c.DOI, &c.Authors, &c.Year, &c.BibTeX,
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// ListModels retrieves all models with pagination
func (db *DB) ListModels(limit, offset int) ([]Model, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, domain, description, github_repo, license, created_at, updated_at
		FROM models
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query models: %w", err)
	}

	var models []Model
	for rows.Next() {
		var m Model
		if err := rows.Scan(
			&m.ID, &m.Name, &m.Domain, &m.Description,
			&m.GitHubRepo, &m.License, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				return nil, fmt.Errorf("scan error: %w, close error: %v", err, closeErr)
			}
			return nil, err
		}
		models = append(models, m)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return models, rows.Err()
}

// Search performs full-text search on models
func (db *DB) Search(opts SearchOptions) ([]SearchResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}

	var conditions []string
	var args []interface{}

	// Build WHERE clause
	if opts.Query != "" {
		// Use LIKE for text search across name, domain, and description
		conditions = append(conditions, "(m.name LIKE ? OR m.domain LIKE ? OR m.description LIKE ?)")
		searchPattern := "%" + opts.Query + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if opts.Domain != "" {
		conditions = append(conditions, "m.domain = ?")
		args = append(args, opts.Domain)
	}

	if opts.Framework != "" {
		conditions = append(conditions, `
			m.id IN (
				SELECT model_id FROM model_versions
				WHERE framework = ? AND is_latest = TRUE
			)
		`)
		args = append(args, opts.Framework)
	}

	if opts.GPURequired != nil {
		conditions = append(conditions, `
			m.id IN (
				SELECT model_id FROM model_versions
				WHERE gpu_required = ? AND is_latest = TRUE
			)
		`)
		args = append(args, *opts.GPURequired)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "m.updated_at DESC"
	switch opts.SortBy {
	case "name":
		orderBy = "m.name ASC"
	case "created_at":
		orderBy = "m.created_at DESC"
	}

	// Build query using strings.Builder to avoid gosec G201
	var query strings.Builder
	query.WriteString("SELECT m.id, m.name, m.domain, m.description, ")
	query.WriteString("m.github_repo, m.license, m.created_at, m.updated_at ")
	query.WriteString("FROM models m ")
	if len(conditions) > 0 {
		query.WriteString(whereClause)
		query.WriteString(" ")
	}
	query.WriteString("ORDER BY ")
	query.WriteString(orderBy)
	query.WriteString(" LIMIT ? OFFSET ?")

	args = append(args, opts.Limit, opts.Offset)

	rows, err := db.conn.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Domain, &r.Description,
			&r.GitHubRepo, &r.License, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				return nil, fmt.Errorf("scan error: %w, close error: %v", err, closeErr)
			}
			return nil, err
		}
		results = append(results, r)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return results, rows.Err()
}

// UpdateModel updates an existing model's metadata
func (db *DB) UpdateModel(name string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Whitelist of allowed columns for update
	allowedColumns := map[string]bool{
		"description": true,
		"github_repo": true,
		"license":     true,
	}

	var setClauses []string
	var args []interface{}

	for key, value := range updates {
		if !allowedColumns[key] {
			return fmt.Errorf("invalid column for update: %s", key)
		}
		setClauses = append(setClauses, key+" = ?")
		args = append(args, value)
	}

	// Add updated_at
	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now())

	// Add WHERE condition
	args = append(args, name)

	// Build query using strings.Builder
	var query strings.Builder
	query.WriteString("UPDATE models SET ")
	query.WriteString(strings.Join(setClauses, ", "))
	query.WriteString(" WHERE name = ?")

	result, err := db.conn.Exec(query.String(), args...)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("model not found: %s", name)
	}

	return nil
}

// DeleteModel removes a model and all its versions from the catalog
func (db *DB) DeleteModel(name string) error {
	result, err := db.conn.Exec("DELETE FROM models WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("model not found: %s", name)
	}

	return nil
}

// CountModels returns the total number of models in the catalog
func (db *DB) CountModels() (int64, error) {
	var count int64
	err := db.conn.QueryRow("SELECT COUNT(*) FROM models").Scan(&count)
	return count, err
}

// CreateModelVersion adds a new version to an existing model
func (db *DB) CreateModelVersion(modelID int64, m *types.Model) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			err = fmt.Errorf("transaction error: %w, rollback error: %v", err, rollbackErr)
		}
	}()

	// Check if this version already exists
	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM model_versions
			WHERE model_id = ? AND version = ?
		)
	`, modelID, m.Version).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check version existence: %w", err)
	}

	if exists {
		return fmt.Errorf("version %s already exists for this model", m.Version)
	}

	// Mark all existing versions as not latest
	_, err = tx.Exec(`
		UPDATE model_versions
		SET is_latest = FALSE
		WHERE model_id = ?
	`, modelID)
	if err != nil {
		return fmt.Errorf("failed to update existing versions: %w", err)
	}

	// Insert new version
	versionID, err := db.insertModelVersion(tx, modelID, m)
	if err != nil {
		return fmt.Errorf("failed to insert model version: %w", err)
	}

	// Insert benchmarks
	for _, benchmark := range m.Benchmarks {
		if err := db.insertBenchmark(tx, versionID, &benchmark); err != nil {
			return fmt.Errorf("failed to insert benchmark: %w", err)
		}
	}

	// Update model updated_at timestamp
	_, err = tx.Exec(`
		UPDATE models
		SET updated_at = ?
		WHERE id = ?
	`, time.Now(), modelID)
	if err != nil {
		return fmt.Errorf("failed to update model timestamp: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetModelVersion retrieves a specific version of a model
func (db *DB) GetModelVersion(modelName, version string) (*ModelVersion, error) {
	// First get model ID
	var modelID int64
	err := db.conn.QueryRow(`
		SELECT id FROM models WHERE name = ?
	`, modelName).Scan(&modelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		return nil, fmt.Errorf("failed to query model: %w", err)
	}

	// Get specific version
	var v ModelVersion
	err = db.conn.QueryRow(`
		SELECT id, model_id, version, weights_uri, weights_size_gb, checksum_sha256,
		       framework, python_version, dependencies, custom_image,
		       entrypoint, handler,
		       gpu_required, recommended_instance, min_cpu, min_memory_gb, min_gpu_memory_gb,
		       created_at, is_latest
		FROM model_versions
		WHERE model_id = ? AND version = ?
	`, modelID, version).Scan(
		&v.ID, &v.ModelID, &v.Version, &v.WeightsURI, &v.WeightsSizeGB, &v.ChecksumSHA256,
		&v.Framework, &v.PythonVersion, &v.Dependencies, &v.CustomImage,
		&v.Entrypoint, &v.Handler,
		&v.GPURequired, &v.RecommendedInstance, &v.MinCPU, &v.MinMemoryGB, &v.MinGPUMemoryGB,
		&v.CreatedAt, &v.IsLatest,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("version not found: %s", version)
		}
		return nil, err
	}

	// Load benchmarks
	benchmarks, err := db.getBenchmarks(v.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load benchmarks: %w", err)
	}
	v.Benchmarks = benchmarks

	return &v, nil
}

// ListModelVersions retrieves all versions of a model
func (db *DB) ListModelVersions(modelName string) ([]ModelVersion, error) {
	// First get model ID
	var modelID int64
	err := db.conn.QueryRow(`
		SELECT id FROM models WHERE name = ?
	`, modelName).Scan(&modelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		return nil, fmt.Errorf("failed to query model: %w", err)
	}

	// Get all versions
	rows, err := db.conn.Query(`
		SELECT id, model_id, version, weights_uri, weights_size_gb, checksum_sha256,
		       framework, python_version, dependencies, custom_image,
		       entrypoint, handler,
		       gpu_required, recommended_instance, min_cpu, min_memory_gb, min_gpu_memory_gb,
		       created_at, is_latest
		FROM model_versions
		WHERE model_id = ?
		ORDER BY created_at DESC
	`, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}

	var versions []ModelVersion
	for rows.Next() {
		var v ModelVersion
		if err := rows.Scan(
			&v.ID, &v.ModelID, &v.Version, &v.WeightsURI, &v.WeightsSizeGB, &v.ChecksumSHA256,
			&v.Framework, &v.PythonVersion, &v.Dependencies, &v.CustomImage,
			&v.Entrypoint, &v.Handler,
			&v.GPURequired, &v.RecommendedInstance, &v.MinCPU, &v.MinMemoryGB, &v.MinGPUMemoryGB,
			&v.CreatedAt, &v.IsLatest,
		); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				return nil, fmt.Errorf("scan error: %w, close error: %v", err, closeErr)
			}
			return nil, err
		}

		// Load benchmarks
		benchmarks, err := db.getBenchmarks(v.ID)
		if err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				return nil, fmt.Errorf("benchmark load error: %w, close error: %v", err, closeErr)
			}
			return nil, fmt.Errorf("failed to load benchmarks: %w", err)
		}
		v.Benchmarks = benchmarks

		versions = append(versions, v)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return versions, rows.Err()
}
