# Scientific Models Platform - Go Implementation Plan

## Project Structure

```
conduit/                          # Or whatever name you choose
├── cmd/
│   ├── conduit/                 # Main CLI binary
│   │   └── main.go
│   ├── conduit-server/          # API server
│   │   └── main.go
│   ├── conduit-worker/          # Background jobs (benchmarks, validation)
│   │   └── main.go
│   └── conduit-crawler/         # GitHub repo crawler
│       └── main.go
│
├── internal/                     # Private application code
│   ├── cli/                     # CLI commands
│   │   ├── init.go             # conduit init
│   │   ├── validate.go         # conduit validate
│   │   ├── benchmark.go        # conduit benchmark
│   │   ├── publish.go          # conduit publish
│   │   ├── deploy.go           # conduit deploy
│   │   ├── install.go          # conduit install <model>
│   │   └── search.go           # conduit search
│   │
│   ├── model/                   # Core domain models
│   │   ├── model.go            # Model struct and methods
│   │   ├── workflow.go         # Pipeline/agentic workflows
│   │   ├── benchmark.go        # Benchmark results
│   │   └── validation.go       # Validation results
│   │
│   ├── parser/                  # model.yaml parser
│   │   ├── yaml.go             # YAML parsing
│   │   ├── schema.go           # JSON schema validation
│   │   └── validator.go        # Semantic validation
│   │
│   ├── bedrock/                 # AWS Bedrock integration
│   │   ├── client.go           # Bedrock API client
│   │   ├── deploy.go           # Model deployment
│   │   ├── endpoint.go         # Endpoint management
│   │   └── inference.go        # Run inference
│   │
│   ├── weights/                 # Weight resolution
│   │   ├── resolver.go         # Main resolver interface
│   │   ├── s3.go               # S3 resolver
│   │   ├── huggingface.go      # HuggingFace resolver
│   │   ├── http.go             # HTTP/HTTPS resolver
│   │   └── cache.go            # Local caching
│   │
│   ├── github/                  # GitHub integration
│   │   ├── client.go           # GitHub API client
│   │   ├── repo.go             # Repo operations
│   │   ├── releases.go         # Release management
│   │   └── webhook.go          # Webhook handling
│   │
│   ├── zenodo/                  # Zenodo DOI integration
│   │   ├── client.go           # Zenodo API client
│   │   └── doi.go              # DOI minting
│   │
│   ├── catalog/                 # Model catalog
│   │   ├── store.go            # Database interface
│   │   ├── search.go           # Search implementation
│   │   ├── index.go            # Search indexing
│   │   └── cache.go            # Redis caching
│   │
│   ├── benchmark/               # Benchmark framework
│   │   ├── runner.go           # Benchmark execution
│   │   ├── datasets.go         # Standard datasets
│   │   ├── metrics.go          # Metric calculators
│   │   └── reporter.go         # Results reporting
│   │
│   ├── ui/                      # UI generation
│   │   ├── streamlit.go        # Streamlit generator
│   │   ├── templates/          # UI templates
│   │   └── deployer.go         # UI deployment
│   │
│   ├── worker/                  # Background jobs
│   │   ├── jobs.go             # Job definitions
│   │   ├── queue.go            # Queue interface (SQS)
│   │   └── executor.go         # Job execution
│   │
│   └── api/                     # REST API handlers
│       ├── server.go           # HTTP server
│       ├── handlers/
│       │   ├── models.go       # Model endpoints
│       │   ├── search.go       # Search endpoints
│       │   ├── deploy.go       # Deployment endpoints
│       │   └── webhooks.go     # GitHub webhooks
│       └── middleware/
│           ├── auth.go         # Authentication
│           ├── cors.go         # CORS
│           └── logging.go      # Request logging
│
├── pkg/                         # Public libraries (can be imported by others)
│   ├── sdk/                     # Go SDK for users
│   │   ├── client.go
│   │   ├── model.go
│   │   └── deploy.go
│   └── types/                   # Shared types
│       └── model.go
│
├── web/                         # Web frontend (Next.js)
│   ├── package.json
│   ├── next.config.js
│   ├── app/
│   │   ├── page.tsx            # Homepage
│   │   ├── models/
│   │   │   └── [id]/page.tsx  # Model detail page
│   │   ├── search/
│   │   │   └── page.tsx       # Search page
│   │   └── docs/
│   │       └── page.tsx       # Documentation
│   └── components/
│       ├── ModelCard.tsx
│       ├── SearchBar.tsx
│       └── DeployButton.tsx
│
├── migrations/                  # Database migrations
│   ├── 001_initial.sql
│   ├── 002_add_benchmarks.sql
│   └── ...
│
├── config/                      # Configuration
│   ├── config.go               # Config loader
│   └── config.yaml             # Default config
│
├── scripts/                     # Utility scripts
│   ├── seed-catalog.sh         # Seed initial models
│   ├── deploy.sh               # Deployment script
│   └── benchmark-all.sh        # Run all benchmarks
│
├── templates/                   # Model repo templates
│   ├── protein-folding/
│   │   ├── model.yaml
│   │   ├── inference.py
│   │   └── requirements.txt
│   ├── materials-science/
│   └── drug-discovery/
│
├── testdata/                    # Test fixtures
│   ├── models/
│   │   └── alphafold2.yaml
│   └── responses/
│       └── bedrock_response.json
│
├── docs/                        # Documentation
│   ├── README.md
│   ├── architecture.md
│   ├── api.md
│   └── cli.md
│
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── .github/
│   └── workflows/
│       ├── test.yml
│       ├── build.yml
│       └── deploy.yml
└── README.md
```

## Technology Stack

### Backend (Go)
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra) - industry standard for CLI apps
- **Web Framework:** [Gin](https://github.com/gin-gonic/gin) or [Fiber](https://gofiber.io/) - high performance
- **Database:** PostgreSQL (models, benchmarks, metadata)
- **Search:** OpenSearch or Elasticsearch (model discovery)
- **Cache:** Redis (API responses, search results)
- **Queue:** AWS SQS (background jobs)
- **Storage:** AWS S3 (artifacts, benchmarks)

### AWS SDK
- **AWS SDK for Go v2** - Bedrock, S3, SQS, Cognito

### Frontend (Next.js/React)
- **Framework:** Next.js 14 (App Router)
- **UI Components:** shadcn/ui or Tailwind UI
- **State Management:** React Query
- **Visualization:** Recharts, Plotly

### Other
- **Container Registry:** ECR (optional - since we're Bedrock-native)
- **CI/CD:** GitHub Actions
- **Monitoring:** CloudWatch, Prometheus
- **Logging:** CloudWatch Logs, structured logging

## Core Data Models

```go
// pkg/types/model.go
package types

import "time"

type Model struct {
    ID          string    `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Version     string    `json:"version" db:"version"`
    Domain      string    `json:"domain" db:"domain"`
    Type        ModelType `json:"type" db:"type"`
    Description string    `json:"description" db:"description"`
    
    // Citation
    DOI         string   `json:"doi" db:"doi"`
    Paper       string   `json:"paper" db:"paper"`
    Authors     []Author `json:"authors"`
    
    // Source
    GitHubRepo  string `json:"github_repo" db:"github_repo"`
    GitHubTag   string `json:"github_tag" db:"github_tag"`
    
    // Artifacts
    Weights     Weights `json:"weights"`
    
    // Runtime
    Runtime     Runtime `json:"runtime"`
    
    // Hardware
    Hardware    Hardware `json:"hardware"`
    
    // Benchmarks
    Benchmarks  []Benchmark `json:"benchmarks"`
    
    // Quality
    Validation  ValidationStatus `json:"validation"`
    Certification *Certification `json:"certification,omitempty"`
    
    // UI
    UI          *UIConfig `json:"ui,omitempty"`
    
    // Metadata
    Tags        []string  `json:"tags"`
    License     string    `json:"license" db:"license"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
    Downloads   int64     `json:"downloads" db:"downloads"`
    Stars       int       `json:"stars" db:"stars"`
}

type ModelType string

const (
    ModelTypeSingle  ModelType = "single-model"
    ModelTypePipeline ModelType = "pipeline"
    ModelTypeEnsemble ModelType = "ensemble"
    ModelTypeMoE     ModelType = "model-of-experts"
    ModelTypeHybrid  ModelType = "hybrid"
    ModelTypeAgentic ModelType = "agentic"
)

type Weights struct {
    URI           string  `json:"uri" yaml:"uri"`
    DOI           string  `json:"doi,omitempty" yaml:"doi,omitempty"`
    HuggingFaceID string  `json:"huggingface_id,omitempty" yaml:"huggingface_id,omitempty"`
    SizeGB        float64 `json:"size_gb" yaml:"size_gb"`
    Format        string  `json:"format" yaml:"format"`
    ChecksumSHA256 string `json:"checksum_sha256,omitempty" yaml:"checksum_sha256,omitempty"`
    OpenDataRegistry bool `json:"open_data_registry" yaml:"open_data_registry"`
}

type Runtime struct {
    Framework     string   `json:"framework" yaml:"framework"`
    PythonVersion string   `json:"python_version" yaml:"python_version"`
    Dependencies  string   `json:"dependencies" yaml:"dependencies"`
    Entrypoint    string   `json:"entrypoint" yaml:"entrypoint"`
    Handler       string   `json:"handler" yaml:"handler"`
}

type Hardware struct {
    GPURequired      bool   `json:"gpu_required" yaml:"gpu_required"`
    MinGPUMemoryGB   int    `json:"min_gpu_memory_gb" yaml:"min_gpu_memory_gb"`
    RecommendedInstance string `json:"recommended_instance" yaml:"recommended_instance"`
    SupportsBatch    bool   `json:"supports_batch" yaml:"supports_batch"`
    OptimalBatchSize int    `json:"optimal_batch_size,omitempty" yaml:"optimal_batch_size,omitempty"`
}

type Benchmark struct {
    Dataset        string    `json:"dataset"`
    Metric         string    `json:"metric"`
    Result         float64   `json:"result"`
    Instance       string    `json:"instance"`
    CostPer1k      string    `json:"cost_per_1k"`
    LatencyP50Ms   int       `json:"latency_p50_ms"`
    LatencyP99Ms   int       `json:"latency_p99_ms"`
    ThroughputPerHour int    `json:"throughput_per_hour"`
    Date           time.Time `json:"date"`
}

type ValidationStatus struct {
    Status       string    `json:"status"` // "pending", "passed", "failed"
    Date         time.Time `json:"date"`
    Checks       []Check   `json:"checks"`
}

type Check struct {
    Name    string `json:"name"`
    Status  string `json:"status"` // "pass", "fail", "warning"
    Message string `json:"message,omitempty"`
}

type Certification struct {
    Status     string    `json:"status"` // "requested", "certified", "rejected"
    Date       time.Time `json:"date"`
    Reviewers  []Reviewer `json:"reviewers"`
}
```

## Implementation Phases

### Phase 1: MVP (Weeks 1-2)

**Week 1: Core CLI + Parser**
```bash
Day 1-2: Project setup, Cobra CLI skeleton
Day 3-4: model.yaml parser + validation
Day 5-7: Basic Bedrock integration (deploy single model)
```

**Week 2: Basic Catalog**
```bash
Day 1-3: Database schema, PostgreSQL setup
Day 4-5: GitHub integration (fetch repos, parse model.yaml)
Day 6-7: Basic REST API + search
```

**Deliverable:** Can parse model.yaml, deploy to Bedrock, basic catalog

### Phase 2: Publisher Tools (Weeks 3-4)

**Week 3: Publishing Flow**
```bash
Day 1-2: conduit init (generate templates)
Day 3-4: conduit validate (run checks)
Day 5-7: conduit publish (GitHub integration, Zenodo DOI)
```

**Week 4: Benchmarking**
```bash
Day 1-3: Benchmark framework
Day 4-5: Standard datasets (CASP15, etc.)
Day 6-7: conduit benchmark command
```

**Deliverable:** Full publisher workflow working

### Phase 3: Consumer Tools (Weeks 5-6)

**Week 5: Discovery + Deployment**
```bash
Day 1-3: Advanced search (filters, facets)
Day 4-5: conduit search + install commands
Day 6-7: Batch deployment
```

**Week 6: SDK**
```bash
Day 1-4: Go SDK for programmatic access
Day 5-7: Python SDK (wrapper around API)
```

**Deliverable:** Consumer can discover and deploy models easily

### Phase 4: Web Frontend (Weeks 7-8)

**Week 7: Basic UI**
```bash
Day 1-3: Next.js setup, homepage, search
Day 4-7: Model detail pages, deployment UI
```

**Week 8: Auto-generated UIs**
```bash
Day 1-4: Streamlit generator
Day 5-7: Deployment of generated UIs
```

**Deliverable:** Full web experience

### Phase 5: Advanced Features (Weeks 9-12)

**Week 9-10: Workflows**
```bash
- Pipeline support
- Multi-model composition
- Workflow execution
```

**Week 11-12: Agentic + MCP**
```bash
- Foundation model integration
- MCP server support
- AWS Strands integration
```

**Deliverable:** Advanced workflows working

## Key Files to Start With

### 1. cmd/conduit/main.go
```go
package main

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/yourusername/conduit/internal/cli"
)

var rootCmd = &cobra.Command{
    Use:   "conduit",
    Short: "Scientific model publishing and deployment platform",
    Long:  `Conduit makes it easy to publish, discover, and deploy scientific ML models.`,
}

func main() {
    // Add commands
    rootCmd.AddCommand(cli.InitCmd())
    rootCmd.AddCommand(cli.ValidateCmd())
    rootCmd.AddCommand(cli.BenchmarkCmd())
    rootCmd.AddCommand(cli.PublishCmd())
    rootCmd.AddCommand(cli.DeployCmd())
    rootCmd.AddCommand(cli.SearchCmd())
    rootCmd.AddCommand(cli.InstallCmd())
    
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

### 2. internal/parser/yaml.go
```go
package parser

import (
    "os"
    
    "gopkg.in/yaml.v3"
    "github.com/yourusername/conduit/pkg/types"
)

type Parser struct {
    // Configuration
}

func New() *Parser {
    return &Parser{}
}

func (p *Parser) ParseFile(path string) (*types.Model, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    return p.Parse(data)
}

func (p *Parser) Parse(data []byte) (*types.Model, error) {
    var model types.Model
    
    if err := yaml.Unmarshal(data, &model); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }
    
    // Validate
    if err := p.Validate(&model); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    return &model, nil
}

func (p *Parser) Validate(model *types.Model) error {
    // Schema validation
    // Semantic validation
    // Required fields check
    return nil
}
```

### 3. internal/bedrock/deploy.go
```go
package bedrock

import (
    "context"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/bedrock"
    "github.com/yourusername/conduit/pkg/types"
)

type Deployer struct {
    client *bedrock.Client
}

func NewDeployer(ctx context.Context) (*Deployer, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, err
    }
    
    return &Deployer{
        client: bedrock.NewFromConfig(cfg),
    }, nil
}

func (d *Deployer) Deploy(ctx context.Context, model *types.Model, opts DeployOptions) (*Endpoint, error) {
    // 1. Resolve and stage weights
    weightsURI, err := d.stageWeights(ctx, model.Weights)
    if err != nil {
        return nil, err
    }
    
    // 2. Create Bedrock model
    modelArn, err := d.createBedrockModel(ctx, model, weightsURI)
    if err != nil {
        return nil, err
    }
    
    // 3. Create inference endpoint
    endpoint, err := d.createEndpoint(ctx, modelArn, opts)
    if err != nil {
        return nil, err
    }
    
    return endpoint, nil
}

func (d *Deployer) stageWeights(ctx context.Context, weights types.Weights) (string, error) {
    // Handle different weight sources (S3, HuggingFace, HTTP)
    // Return S3 URI that Bedrock can access
    return "", nil
}
```

### 4. internal/cli/init.go
```go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/yourusername/conduit/internal/templates"
)

func InitCmd() *cobra.Command {
    var templateName string
    
    cmd := &cobra.Command{
        Use:   "init [path]",
        Short: "Initialize a new model repository",
        Args:  cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            path := "."
            if len(args) > 0 {
                path = args[0]
            }
            
            // Get template
            tmpl, err := templates.Get(templateName)
            if err != nil {
                return err
            }
            
            // Generate files
            if err := tmpl.Generate(path); err != nil {
                return err
            }
            
            cmd.Printf("✓ Initialized model repository at %s\n", path)
            cmd.Println("Next steps:")
            cmd.Println("  1. Edit model.yaml with your model details")
            cmd.Println("  2. Add your inference code to src/inference.py")
            cmd.Println("  3. Run 'conduit validate' to check your setup")
            
            return nil
        },
    }
    
    cmd.Flags().StringVarP(&templateName, "template", "t", "protein-folding", 
        "Template to use (protein-folding, materials-science, drug-discovery)")
    
    return cmd
}
```

## Database Schema

```sql
-- migrations/001_initial.sql

CREATE TABLE models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    domain VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    
    doi VARCHAR(255),
    paper_url TEXT,
    
    github_repo VARCHAR(500) NOT NULL,
    github_tag VARCHAR(100),
    
    weights_uri TEXT NOT NULL,
    weights_size_gb DECIMAL(10,2),
    weights_format VARCHAR(50),
    
    framework VARCHAR(50),
    python_version VARCHAR(20),
    
    gpu_required BOOLEAN DEFAULT false,
    min_gpu_memory_gb INTEGER,
    recommended_instance VARCHAR(100),
    
    license VARCHAR(100),
    
    validation_status VARCHAR(50) DEFAULT 'pending',
    certification_status VARCHAR(50),
    
    downloads BIGINT DEFAULT 0,
    stars INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(github_repo, version)
);

CREATE INDEX idx_models_domain ON models(domain);
CREATE INDEX idx_models_type ON models(type);
CREATE INDEX idx_models_github_repo ON models(github_repo);
CREATE INDEX idx_models_downloads ON models(downloads DESC);

CREATE TABLE model_authors (
    model_id UUID REFERENCES models(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    orcid VARCHAR(50),
    institution VARCHAR(255),
    PRIMARY KEY (model_id, orcid)
);

CREATE TABLE model_tags (
    model_id UUID REFERENCES models(id) ON DELETE CASCADE,
    tag VARCHAR(100) NOT NULL,
    PRIMARY KEY (model_id, tag)
);

CREATE INDEX idx_model_tags_tag ON model_tags(tag);

CREATE TABLE benchmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID REFERENCES models(id) ON DELETE CASCADE,
    dataset VARCHAR(255) NOT NULL,
    metric VARCHAR(100) NOT NULL,
    result DECIMAL(10,6),
    instance VARCHAR(100),
    cost_per_1k VARCHAR(20),
    latency_p50_ms INTEGER,
    latency_p99_ms INTEGER,
    throughput_per_hour INTEGER,
    date TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_benchmarks_model_id ON benchmarks(model_id);
CREATE INDEX idx_benchmarks_dataset ON benchmarks(dataset);

CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID REFERENCES models(id),
    user_id VARCHAR(255) NOT NULL,
    endpoint_name VARCHAR(255) NOT NULL,
    instance_type VARCHAR(100),
    status VARCHAR(50),
    bedrock_endpoint_arn TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_deployments_user_id ON deployments(user_id);
CREATE INDEX idx_deployments_model_id ON deployments(model_id);
```

## Quick Start Guide

### For a Developer Starting Tomorrow

```bash
# Day 1 Morning: Project setup
mkdir conduit && cd conduit
go mod init github.com/yourusername/conduit
go get github.com/spf13/cobra
go get github.com/gin-gonic/gin
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/bedrock
go get gopkg.in/yaml.v3

# Create structure
mkdir -p cmd/conduit internal/{cli,parser,model,bedrock} pkg/types

# Day 1 Afternoon: Basic CLI
# Implement cmd/conduit/main.go with Cobra
# Implement internal/cli/init.go
# Test: conduit init --template protein-folding

# Day 2: Parser
# Implement internal/parser/yaml.go
# Implement pkg/types/model.go
# Test: Parse sample model.yaml

# Day 3-4: Bedrock Integration
# Implement internal/bedrock/client.go
# Implement internal/bedrock/deploy.go
# Test: Deploy a simple model

# Day 5: Database
# Setup PostgreSQL
# Run migrations
# Implement internal/catalog/store.go

# Week 2: API Server
# Implement cmd/conduit-server/main.go
# Implement internal/api/handlers/
# Test: curl http://localhost:8080/api/models

# You're off to the races!
```

## Configuration

```yaml
# config/config.yaml
server:
  port: 8080
  host: 0.0.0.0

database:
  host: localhost
  port: 5432
  name: conduit
  user: conduit
  password: ${DB_PASSWORD}

aws:
  region: us-east-1
  bedrock:
    default_instance: ml.g5.2xlarge

redis:
  host: localhost
  port: 6379

github:
  token: ${GITHUB_TOKEN}
  webhook_secret: ${GITHUB_WEBHOOK_SECRET}

zenodo:
  token: ${ZENODO_TOKEN}
  sandbox: true  # Use sandbox for testing

catalog:
  index_name: scientific_models
  refresh_interval: 5m
```

## Testing Strategy

```go
// internal/parser/yaml_test.go
package parser

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestParseValidModel(t *testing.T) {
    parser := New()
    
    model, err := parser.ParseFile("../../testdata/models/alphafold2.yaml")
    assert.NoError(t, err)
    assert.Equal(t, "alphafold2-multimer", model.Name)
    assert.Equal(t, "2.3.2", model.Version)
}

func TestParseInvalidModel(t *testing.T) {
    parser := New()
    
    _, err := parser.ParseFile("../../testdata/models/invalid.yaml")
    assert.Error(t, err)
}
```

## Deployment

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /conduit-server ./cmd/conduit-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /conduit-server /conduit-server

EXPOSE 8080
CMD ["/conduit-server"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: conduit
      POSTGRES_USER: conduit
      POSTGRES_PASSWORD: conduit
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  server:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    environment:
      DB_PASSWORD: conduit
      AWS_REGION: us-east-1

volumes:
  postgres_data:
```

## Makefile

```makefile
.PHONY: build test run clean

build:
	go build -o bin/conduit ./cmd/conduit
	go build -o bin/conduit-server ./cmd/conduit-server

test:
	go test -v ./...

run-server:
	go run ./cmd/conduit-server

run-cli:
	go run ./cmd/conduit

docker-build:
	docker build -t conduit-server .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	psql -h localhost -U conduit -d conduit -f migrations/001_initial.sql

clean:
	rm -rf bin/
	docker-compose down -v
```

## Next Steps

1. **Day 1:** Set up project structure, implement basic CLI skeleton
2. **Day 2:** YAML parser + validation
3. **Day 3-4:** Bedrock integration (deploy single model)
4. **Day 5:** Database setup + basic catalog
5. **Week 2:** REST API + GitHub integration
6. **Week 3:** Publishing flow (init, validate, publish)
7. **Week 4:** Benchmark framework
8. **Week 5-6:** Consumer tools (search, install, deploy)
9. **Week 7-8:** Web frontend
10. **Week 9-12:** Advanced features (workflows, agentic)

With Go, you can move incredibly fast. The CLI tool is a single binary that users can download - no dependencies. The backend is fast and can handle high load. You could have an MVP in 2 weeks working full-time, or 4-6 weeks part-time.

Want me to generate specific starter files (main.go, parser.go, etc.) to get you coding immediately?
