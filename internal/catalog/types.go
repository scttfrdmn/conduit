package catalog

import "time"

// Model represents a model in the catalog
type Model struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Domain      string    `db:"domain"`
	Description string    `db:"description"`
	GitHubRepo  string    `db:"github_repo"`
	License     string    `db:"license"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	// Related data (not from DB directly)
	LatestVersion *ModelVersion `db:"-"`
	Citation      *Citation     `db:"-"`
	Tags          []string      `db:"-"`
}

// ModelVersion represents a specific version of a model
type ModelVersion struct {
	ID          int64   `db:"id"`
	ModelID     int64   `db:"model_id"`
	Version     string  `db:"version"`
	WeightsURI  string  `db:"weights_uri"`
	WeightsSizeGB float64 `db:"weights_size_gb"`
	ChecksumSHA256 string `db:"checksum_sha256"`

	// Runtime
	Framework     string `db:"framework"`
	PythonVersion string `db:"python_version"`
	Dependencies  string `db:"dependencies"`
	CustomImage   string `db:"custom_image"`

	// Inference
	Entrypoint string `db:"entrypoint"`
	Handler    string `db:"handler"`

	// Hardware
	GPURequired         bool   `db:"gpu_required"`
	RecommendedInstance string `db:"recommended_instance"`
	MinCPU              int    `db:"min_cpu"`
	MinMemoryGB         int    `db:"min_memory_gb"`
	MinGPUMemoryGB      int    `db:"min_gpu_memory_gb"`

	// Metadata
	CreatedAt time.Time `db:"created_at"`
	IsLatest  bool      `db:"is_latest"`

	// Related data
	Benchmarks []Benchmark `db:"-"`
}

// Benchmark represents a performance metric
type Benchmark struct {
	ID               int64     `db:"id"`
	ModelVersionID   int64     `db:"model_version_id"`
	Dataset          string    `db:"dataset"`
	Metric           string    `db:"metric"`
	Result           float64   `db:"result"`
	Instance         string    `db:"instance"`
	CostPerPrediction string   `db:"cost_per_prediction"`
	WalltimeSeconds  float64   `db:"walltime_seconds"`
	CreatedAt        time.Time `db:"created_at"`
}

// Citation represents publication information
type Citation struct {
	ID         int64  `db:"id"`
	ModelID    int64  `db:"model_id"`
	PaperTitle string `db:"paper_title"`
	PaperURL   string `db:"paper_url"`
	DOI        string `db:"doi"`
	Authors    string `db:"authors"`
	Year       int    `db:"year"`
	BibTeX     string `db:"bibtex"`
}

// SearchResult represents a model search result
type SearchResult struct {
	Model
	Rank  float64 `db:"rank"` // FTS rank score
	Score float64 // Calculated relevance score
}

// SearchOptions configures search behavior
type SearchOptions struct {
	Query      string
	Domain     string
	Framework  string
	GPURequired *bool
	Tags       []string
	Limit      int
	Offset     int
	SortBy     string // "relevance", "name", "created_at", "updated_at"
}
