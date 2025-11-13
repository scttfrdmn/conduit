package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/conduit/pkg/types"
)

// CatalogExport represents an exported catalog
type CatalogExport struct {
	Version    string               `json:"version"`
	ExportedAt time.Time            `json:"exported_at"`
	Models     []ExportedModel      `json:"models"`
}

// ExportedModel represents a complete model with all versions and metadata
type ExportedModel struct {
	Name        string                `json:"name"`
	Domain      string                `json:"domain"`
	Description string                `json:"description"`
	GitHubRepo  string                `json:"github_repo,omitempty"`
	License     string                `json:"license,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Citation    *types.Citation       `json:"citation,omitempty"`
	Versions    []ExportedVersion     `json:"versions"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// ExportedVersion represents a model version with benchmarks
type ExportedVersion struct {
	Version         string            `json:"version"`
	IsLatest        bool              `json:"is_latest"`
	WeightsURI      string            `json:"weights_uri"`
	WeightsSizeGB   float64           `json:"weights_size_gb,omitempty"`
	ChecksumSHA256  string            `json:"checksum_sha256,omitempty"`
	Framework       string            `json:"framework"`
	PythonVersion   string            `json:"python_version"`
	Dependencies    string            `json:"dependencies,omitempty"`
	CustomImage     string            `json:"custom_image,omitempty"`
	Entrypoint      string            `json:"entrypoint"`
	Handler         string            `json:"handler"`
	GPURequired     bool              `json:"gpu_required"`
	RecommendedInstance string        `json:"recommended_instance,omitempty"`
	MinCPU          int               `json:"min_cpu,omitempty"`
	MinMemoryGB     int               `json:"min_memory_gb,omitempty"`
	MinGPUMemoryGB  int               `json:"min_gpu_memory_gb,omitempty"`
	Benchmarks      []types.Benchmark `json:"benchmarks,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

// ExportCatalog exports the entire catalog to a file
func (db *DB) ExportCatalog(outputPath string) error {
	models, err := db.getAllModelsForExport()
	if err != nil {
		return fmt.Errorf("failed to get models: %w", err)
	}

	export := CatalogExport{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Models:     models,
	}

	return writeExportFile(outputPath, export)
}

// ExportModel exports a single model to a file
func (db *DB) ExportModel(modelName, outputPath string) error {
	model, err := db.getModelForExport(modelName)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	export := CatalogExport{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Models:     []ExportedModel{*model},
	}

	return writeExportFile(outputPath, export)
}

// getAllModelsForExport retrieves all models with complete metadata
func (db *DB) getAllModelsForExport() ([]ExportedModel, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, domain, description, github_repo, license, created_at, updated_at
		FROM models
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // Error handling done at end with rows.Err()

	var models []ExportedModel
	for rows.Next() {
		var m Model
		if err := rows.Scan(
			&m.ID, &m.Name, &m.Domain, &m.Description,
			&m.GitHubRepo, &m.License, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}

		exportedModel, err := db.buildExportedModel(&m)
		if err != nil {
			return nil, fmt.Errorf("failed to build exported model %s: %w", m.Name, err)
		}

		models = append(models, *exportedModel)
	}

	return models, rows.Err()
}

// getModelForExport retrieves a single model with complete metadata
func (db *DB) getModelForExport(name string) (*ExportedModel, error) {
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
		return nil, err
	}

	return db.buildExportedModel(&m)
}

// buildExportedModel constructs a complete exported model with all related data
func (db *DB) buildExportedModel(m *Model) (*ExportedModel, error) {
	// Load tags
	tags, err := db.getTags(m.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}

	// Load citation
	citation, err := db.getCitation(m.ID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return nil, fmt.Errorf("failed to load citation: %w", err)
	}

	// Load all versions
	versions, err := db.getAllVersionsForExport(m.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load versions: %w", err)
	}

	// Convert citation to types.Citation if present
	var exportCitation *types.Citation
	if citation != nil {
		var authors []string
		if citation.Authors != "" {
			// Split comma-separated authors
			authors = splitAuthors(citation.Authors)
		}

		exportCitation = &types.Citation{
			PaperTitle: citation.PaperTitle,
			PaperURL:   citation.PaperURL,
			DOI:        citation.DOI,
			Authors:    authors,
			Year:       citation.Year,
			BibTeX:     citation.BibTeX,
		}
	}

	return &ExportedModel{
		Name:        m.Name,
		Domain:      m.Domain,
		Description: m.Description,
		GitHubRepo:  m.GitHubRepo,
		License:     m.License,
		Tags:        tags,
		Citation:    exportCitation,
		Versions:    versions,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

// getAllVersionsForExport retrieves all versions of a model with benchmarks
func (db *DB) getAllVersionsForExport(modelID int64) ([]ExportedVersion, error) {
	rows, err := db.conn.Query(`
		SELECT id, version, weights_uri, weights_size_gb, checksum_sha256,
		       framework, python_version, dependencies, custom_image,
		       entrypoint, handler,
		       gpu_required, recommended_instance, min_cpu, min_memory_gb, min_gpu_memory_gb,
		       created_at, is_latest
		FROM model_versions
		WHERE model_id = ?
		ORDER BY created_at DESC
	`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // Error handling done at end with rows.Err()

	var versions []ExportedVersion
	for rows.Next() {
		var v ModelVersion
		if err := rows.Scan(
			&v.ID, &v.Version, &v.WeightsURI, &v.WeightsSizeGB, &v.ChecksumSHA256,
			&v.Framework, &v.PythonVersion, &v.Dependencies, &v.CustomImage,
			&v.Entrypoint, &v.Handler,
			&v.GPURequired, &v.RecommendedInstance, &v.MinCPU, &v.MinMemoryGB, &v.MinGPUMemoryGB,
			&v.CreatedAt, &v.IsLatest,
		); err != nil {
			return nil, err
		}

		// Load benchmarks
		benchmarks, err := db.getBenchmarks(v.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load benchmarks: %w", err)
		}

		// Convert benchmarks to types.Benchmark
		var exportBenchmarks []types.Benchmark
		for _, b := range benchmarks {
			exportBenchmarks = append(exportBenchmarks, types.Benchmark{
				Dataset:           b.Dataset,
				Metric:            b.Metric,
				Result:            b.Result,
				Instance:          b.Instance,
				CostPerPrediction: b.CostPerPrediction,
				WalltimeSeconds:   b.WalltimeSeconds,
			})
		}

		versions = append(versions, ExportedVersion{
			Version:             v.Version,
			IsLatest:            v.IsLatest,
			WeightsURI:          v.WeightsURI,
			WeightsSizeGB:       v.WeightsSizeGB,
			ChecksumSHA256:      v.ChecksumSHA256,
			Framework:           v.Framework,
			PythonVersion:       v.PythonVersion,
			Dependencies:        v.Dependencies,
			CustomImage:         v.CustomImage,
			Entrypoint:          v.Entrypoint,
			Handler:             v.Handler,
			GPURequired:         v.GPURequired,
			RecommendedInstance: v.RecommendedInstance,
			MinCPU:              v.MinCPU,
			MinMemoryGB:         v.MinMemoryGB,
			MinGPUMemoryGB:      v.MinGPUMemoryGB,
			Benchmarks:          exportBenchmarks,
			CreatedAt:           v.CreatedAt,
		})
	}

	return versions, rows.Err()
}

// writeExportFile writes the export data to a JSON file
func writeExportFile(path string, export CatalogExport) error {
	file, err := os.Create(path) //nolint:gosec // Path is provided by user, legitimate file operation
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close() //nolint:errcheck // Error checked below in Encode

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(export); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// splitAuthors splits a comma-separated author string
func splitAuthors(authors string) []string {
	if authors == "" {
		return nil
	}

	var result []string
	for _, author := range splitByComma(authors) {
		trimmed := trimSpace(author)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Helper functions to avoid importing strings package unnecessarily
func splitByComma(s string) []string {
	var result []string
	var current string
	for _, ch := range s {
		if ch == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing spaces
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
