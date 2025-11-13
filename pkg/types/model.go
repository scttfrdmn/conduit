package types

import "time"

// Model represents a scientific model specification
type Model struct {
	// Metadata
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Domain      string   `yaml:"domain" json:"domain"`
	Description string   `yaml:"description" json:"description"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// Publishing info
	GitHubRepo   string    `yaml:"github_repo,omitempty" json:"github_repo,omitempty"`
	DOI          string    `yaml:"doi,omitempty" json:"doi,omitempty"`
	License      string    `yaml:"license,omitempty" json:"license,omitempty"`
	PublishedAt  time.Time `yaml:"published_at,omitempty" json:"published_at,omitempty"`

	// Runtime configuration
	Runtime  Runtime  `yaml:"runtime" json:"runtime"`

	// Model artifacts
	WeightsURI   string `yaml:"weights_uri" json:"weights_uri"`
	WeightsSizeGB float64 `yaml:"weights_size_gb,omitempty" json:"weights_size_gb,omitempty"`
	ChecksumSHA256 string `yaml:"checksum_sha256,omitempty" json:"checksum_sha256,omitempty"`

	// Inference configuration
	Inference Inference `yaml:"inference" json:"inference"`

	// Hardware requirements
	Hardware Hardware `yaml:"hardware" json:"hardware"`

	// Benchmarks
	Benchmarks []Benchmark `yaml:"benchmarks,omitempty" json:"benchmarks,omitempty"`

	// Signatures (Sigstore)
	Signatures *ModelSignatures `yaml:"signatures,omitempty" json:"signatures,omitempty"`

	// Citations
	Citation Citation `yaml:"citation,omitempty" json:"citation,omitempty"`

	// Usage statistics (not included in YAML, only for display/API)
	Stats *ModelStats `yaml:"-" json:"stats,omitempty"`
}

// Runtime defines the model runtime environment
type Runtime struct {
	Framework     string   `yaml:"framework" json:"framework"` // pytorch, tensorflow, jax, onnx
	PythonVersion string   `yaml:"python_version" json:"python_version"`
	Dependencies  string   `yaml:"dependencies" json:"dependencies"` // path to requirements.txt
	CustomImage   string   `yaml:"custom_image,omitempty" json:"custom_image,omitempty"`
}

// Inference defines how to run inference
type Inference struct {
	Entrypoint string                 `yaml:"entrypoint" json:"entrypoint"` // e.g., inference.py
	Handler    string                 `yaml:"handler" json:"handler"`       // e.g., predict
	InputSchema  map[string]interface{} `yaml:"input_schema,omitempty" json:"input_schema,omitempty"`
	OutputSchema map[string]interface{} `yaml:"output_schema,omitempty" json:"output_schema,omitempty"`
}

// Hardware specifies hardware requirements
type Hardware struct {
	GPURequired         bool    `yaml:"gpu_required" json:"gpu_required"`
	RecommendedInstance string  `yaml:"recommended_instance" json:"recommended_instance"`
	MinCPU              int     `yaml:"min_cpu,omitempty" json:"min_cpu,omitempty"`
	MinMemoryGB         int     `yaml:"min_memory_gb,omitempty" json:"min_memory_gb,omitempty"`
	MinGPUMemoryGB      int     `yaml:"min_gpu_memory_gb,omitempty" json:"min_gpu_memory_gb,omitempty"`
}

// Benchmark represents a benchmark result
type Benchmark struct {
	Dataset          string  `yaml:"dataset" json:"dataset"`
	Metric           string  `yaml:"metric" json:"metric"`
	Result           float64 `yaml:"result" json:"result"`
	Instance         string  `yaml:"instance,omitempty" json:"instance,omitempty"`
	CostPerPrediction string  `yaml:"cost_per_prediction,omitempty" json:"cost_per_prediction,omitempty"`
	WalltimeSeconds  float64 `yaml:"walltime_seconds,omitempty" json:"walltime_seconds,omitempty"`
}

// ModelSignatures contains cryptographic signatures (Sigstore)
type ModelSignatures struct {
	WeightsSignature  *Signature   `yaml:"weights_signature,omitempty" json:"weights_signature,omitempty"`
	CodeSignature     *Signature   `yaml:"code_signature,omitempty" json:"code_signature,omitempty"`
	MetadataSignature *Signature   `yaml:"metadata_signature,omitempty" json:"metadata_signature,omitempty"`
	SignedAt          time.Time    `yaml:"signed_at" json:"signed_at"`
}

// Signature represents a Sigstore signature
type Signature struct {
	ArtifactHash    string    `yaml:"artifact_hash" json:"artifact_hash"`
	SignatureBundle string    `yaml:"signature_bundle" json:"signature_bundle"`
	RekorLogEntry   string    `yaml:"rekor_log_entry" json:"rekor_log_entry"`
	SignedBy        string    `yaml:"signed_by" json:"signed_by"`
	SignedAt        time.Time `yaml:"signed_at" json:"signed_at"`
	CertIdentity    string    `yaml:"cert_identity" json:"cert_identity"`
	Issuer          string    `yaml:"issuer" json:"issuer"`
}

// Citation contains citation information
type Citation struct {
	PaperTitle  string   `yaml:"paper_title,omitempty" json:"paper_title,omitempty"`
	PaperURL    string   `yaml:"paper_url,omitempty" json:"paper_url,omitempty"`
	DOI         string   `yaml:"doi,omitempty" json:"doi,omitempty"`
	Authors     []string `yaml:"authors,omitempty" json:"authors,omitempty"`
	Year        int      `yaml:"year,omitempty" json:"year,omitempty"`
	BibTeX      string   `yaml:"bibtex,omitempty" json:"bibtex,omitempty"`
}

// ModelStats tracks usage and popularity metrics
type ModelStats struct {
	TotalDeployments int       `json:"total_deployments"`
	TotalPredictions int       `json:"total_predictions"`
	ViewCount        int       `json:"view_count"`
	LastDeployedAt   time.Time `json:"last_deployed_at,omitempty"`
	LastViewedAt     time.Time `json:"last_viewed_at,omitempty"`
}
