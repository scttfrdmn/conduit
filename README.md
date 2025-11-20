# Conduit

[![Go Report Card](https://goreportcard.com/badge/github.com/scttfrdmn/conduit)](https://goreportcard.com/report/github.com/scttfrdmn/conduit)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/scttfrdmn/conduit)](https://github.com/scttfrdmn/conduit/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/scttfrdmn/conduit)](go.mod)

**A comprehensive catalog and deployment platform for scientific ML models**

Conduit simplifies the management, distribution, and deployment of scientific machine learning models. Discover models, share them with your team, and deploy to AWS with a single command.

```bash
# Search for models
conduit search "protein folding" --fuzzy

# Deploy to AWS SageMaker
conduit deploy model.yaml --platform sagemaker

# Share with your team
conduit push alphafold2 --registry myteam
```

---

## ✨ Features

### 📦 Model Catalog Management
- **Local catalog** with SQLite storage
- **Full-text search** with fuzzy matching for typos
- **Tags and labels** for organization
- **Usage statistics** and popularity tracking
- **Model versioning** with complete history

### 🌐 Team Collaboration
- **Remote registries** for sharing models (HTTP, S3, Git)
- **Push/pull workflows** like Docker
- **Conflict resolution** strategies (skip, overwrite, merge)
- **Export/import** for backup and migration

### 🚀 AWS Deployment
- **SageMaker endpoints** with full automation
- **Automatic Dockerfile generation** from model specifications
- **ECR integration** (build, push, deploy)
- **Environment detection** (Studio Lab, Studio, Local)
- **Status monitoring** with health checks

### ✅ CI/CD Integration
- **Model validation** (basic and strict modes)
- **GitHub Actions workflows** generation
- **Automated testing** on PRs
- **Publishing workflows** for releases

### 🔍 Advanced Search
- **Fuzzy matching** with Levenshtein distance
- **Multiple filters** (tags, license, author, dates)
- **Relevance scoring** with weighted matching
- **Sort by popularity** based on usage stats

---

## 🚀 Quick Start

### Installation

**From Source** (requires Go 1.23+):
```bash
git clone https://github.com/scttfrdmn/conduit.git
cd conduit
go build -o conduit ./cmd/conduit
sudo mv conduit /usr/local/bin/
```

**Verify Installation**:
```bash
conduit --version
```

### Basic Usage

**Initialize Catalog**:
```bash
# Catalog is created automatically in ~/.conduit/catalog.db
conduit list
```

**Add a Model**:
```bash
# Create a model specification
conduit init my-model

# Edit model.yaml with your model details
# Then add to catalog
conduit add model.yaml
```

**Search and Discover**:
```bash
# Search for models
conduit search "protein"

# Fuzzy search (handles typos)
conduit search "alphafld" --fuzzy

# Filter by tags
conduit search --tags ml,biology

# View model details
conduit info alphafold2
```

**Team Collaboration**:
```bash
# Configure a registry
conduit registry add myteam https://registry.example.com

# Push a model
conduit push alphafold2 --registry myteam

# Pull a model
conduit pull esm2-large --registry myteam
```

**AWS Deployment**:
```bash
# Deploy to SageMaker
conduit deploy model.yaml --platform sagemaker

# With custom configuration
conduit deploy model.yaml \
  --platform sagemaker \
  --instance-type ml.g5.2xlarge \
  --region us-west-2
```

---

## 📚 Documentation

### Guides
- [**Getting Started**](docs/getting-started.md) - Complete walkthrough
- [**Model Specification**](docs/model-spec.md) - model.yaml reference
- [**Command Reference**](docs/commands.md) - All CLI commands
- [**Deployment Guide**](docs/deployment.md) - AWS deployment
- [**Registry Guide**](docs/registry.md) - Team collaboration
- [**CI/CD Guide**](docs/cicd.md) - Automation workflows

### Examples
- [**Model Examples**](examples/) - Sample model.yaml files
- [**Inference Code**](examples/inference/) - Example handlers
- [**Workflows**](examples/workflows/) - GitHub Actions examples

---

## 🏗️ Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Conduit CLI                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Catalog    │  │   Registry   │  │  Deployment  │     │
│  │  (SQLite)    │  │   (HTTP/S3)  │  │  (AWS)       │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Validation  │  │    Search    │  │    Export    │     │
│  │  (Strict)    │  │   (Fuzzy)    │  │   (JSON)     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

**Adding a Model**:
```
model.yaml → Parser → Validator → Catalog (SQLite)
```

**Deploying a Model**:
```
model.yaml → Dockerfile Generator → Docker Build → ECR Push →
SageMaker Model → Endpoint Config → Endpoint Creation
```

**Sharing a Model**:
```
Local Catalog → Export (JSON) → Push to Registry → Team Access
```

---

## 📖 Command Reference

### Catalog Management
```bash
conduit add <model.yaml>              # Add model to catalog
conduit list                          # List all models
conduit info <model-name>             # Show model details
conduit delete <model-name>           # Remove model
conduit search <query>                # Search models
```

### Model Operations
```bash
conduit init [path]                   # Initialize new model project
conduit validate <model.yaml>         # Validate model specification
conduit publish <model.yaml>          # Publish model (placeholder)
```

### Version Management
```bash
conduit version create <model> <ver>  # Create new version
conduit version list <model>          # List versions
conduit version set-latest <m> <ver>  # Set latest version
```

### Export/Import
```bash
conduit export <model> -o file.json   # Export model
conduit import <file.json>            # Import model
```

### Registry Operations
```bash
conduit registry add <name> <url>     # Add registry
conduit registry list                 # List registries
conduit registry remove <name>        # Remove registry
conduit registry set-default <name>   # Set default

conduit push <model>                  # Push to registry
conduit pull <model>                  # Pull from registry
```

### AWS Deployment
```bash
conduit deploy <model.yaml>           # Deploy to AWS
conduit workflow <type>               # Generate CI/CD workflows
```

### CI/CD
```bash
conduit validate --strict             # Strict validation
conduit workflow validate             # Generate validate workflow
conduit workflow publish              # Generate publish workflow
conduit workflow deploy               # Generate deploy workflow
conduit workflow all                  # Generate all workflows
```

---

## 🔧 Model Specification

Models are defined in `model.yaml`:

```yaml
name: "my-model"
version: "1.0.0"
domain: "protein-science"
description: "A brief description of your model"

# Runtime configuration
runtime:
  framework: "pytorch"            # pytorch, tensorflow, jax, onnx
  python_version: "3.11"
  dependencies: "requirements.txt"

# Inference configuration
inference:
  entrypoint: "predict.py"
  handler: "predict"

# Model artifacts
weights_uri: "s3://bucket/weights"
weights_size_gb: 5.2
checksum_sha256: "abc123..."

# Hardware requirements
hardware:
  gpu_required: true
  recommended_instance: "ml.g5.2xlarge"
  min_memory_gb: 16
  min_gpu_memory_gb: 24

# Optional: benchmarks
benchmarks:
  - dataset: "TestSet"
    metric: "accuracy"
    result: 0.95

# Optional: tags
tags:
  - ml
  - protein-folding

# Optional: metadata
license: "Apache-2.0"
github_repo: "github.com/org/repo"
```

See [Model Specification Guide](docs/model-spec.md) for complete reference.

---

## 🌟 Key Features in Detail

### Fuzzy Search
Handle typos and variations in search queries:
```bash
# These all find "alphafold2"
conduit search "alphafld" --fuzzy
conduit search "alpha fold" --fuzzy
conduit search "alfafold" --fuzzy --min-score 0.7
```

### Usage Statistics
Track model popularity and usage:
```bash
conduit info alphafold2
# Shows:
# - Total deployments
# - Total predictions
# - View count
# - Last deployed/viewed dates

# Search by popularity
conduit search --sort-by popular
```

### Model Versioning
Complete version management:
```bash
# Create versions
conduit version create alphafold2 2.3.2
conduit version create alphafold2 2.3.3

# List versions
conduit version list alphafold2

# Set latest
conduit version set-latest alphafold2 2.3.3
```

### Remote Registries
Share models with your team:
```bash
# HTTP registry
conduit registry add myteam https://registry.example.com --type http

# S3 registry (planned)
conduit registry add backup s3://bucket/models --type s3

# Git registry (planned)
conduit registry add public github.com/org/models --type git

# Push and pull
conduit push alphafold2@2.3.2
conduit pull esm2-large --strategy overwrite
```

### AWS Deployment
Full SageMaker integration:
```bash
# Deploys in 4 steps:
# 1. Generate Dockerfile
# 2. Build and push to ECR
# 3. Create SageMaker model
# 4. Deploy endpoint

conduit deploy model.yaml --platform sagemaker

# Monitors progress and waits for InService
```

### CI/CD Automation
Generate GitHub Actions workflows:
```bash
conduit workflow all

# Creates:
# .github/workflows/conduit-validate.yml
# .github/workflows/conduit-publish.yml
# .github/workflows/conduit-deploy.yml
```

---

## 🛠️ Development

### Prerequisites
- Go 1.23+
- Docker (for deployment features)
- AWS CLI (for deployment features)

### Setup
```bash
git clone https://github.com/scttfrdmn/conduit.git
cd conduit

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linting
golangci-lint run

# Build
go build -o conduit ./cmd/conduit
```

### Project Structure
```
conduit/
├── cmd/conduit/          # Main entry point
├── internal/
│   ├── catalog/          # SQLite catalog
│   ├── cli/              # CLI commands
│   ├── deployment/       # AWS deployment
│   ├── registry/         # Remote registries
│   ├── validation/       # Model validation
│   ├── aws/              # AWS helpers
│   └── cicd/             # CI/CD workflows
├── pkg/types/            # Shared types
├── docs/                 # Documentation
└── examples/             # Example models
```

### Running Tests
```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/catalog/...

# Verbose
go test -v ./...
```

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Areas for Contribution
- 🐛 Bug fixes
- 📝 Documentation improvements
- ✨ New features
- 🧪 Test coverage
- 📦 Example models
- 🌍 Internationalization

### Development Workflow
1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes
4. Run tests and linting
5. Commit your changes (follow conventional commits)
6. Push to your fork
7. Open a Pull Request

---

## 📊 Project Status

**Current Version**: v0.2.0 (Beta)

### Completed Features ✅
- ✅ Local catalog with SQLite
- ✅ Model versioning
- ✅ Tags and labels
- ✅ Usage statistics
- ✅ Export/import
- ✅ Fuzzy search
- ✅ Remote registries (HTTP)
- ✅ AWS SageMaker deployment
- ✅ CI/CD workflow generation
- ✅ Model validation (basic & strict)

### In Progress 🚧
- 🚧 S3 and Git registry backends
- 🚧 Endpoint management commands
- 🚧 Cost estimation for deployments

### Planned 📋
- 📋 Web UI for catalog browsing
- 📋 Bedrock custom model deployment
- 📋 Batch inference support
- 📋 Model performance monitoring
- 📋 Multi-model endpoints
- 📋 Auto-scaling configuration

---

## 📝 License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built for the scientific computing community with:
- Go 1.23+
- AWS SDK for Go v2
- Cobra CLI framework
- SQLite
- Docker

---

## 📧 Contact & Support

- **GitHub Issues**: [Report bugs, request features](https://github.com/scttfrdmn/conduit/issues)
- **GitHub Discussions**: [Ask questions, share ideas](https://github.com/scttfrdmn/conduit/discussions)
- **Email**: support@conduit.dev (coming soon)

---

## 🔗 Links

- [Documentation](docs/)
- [Examples](examples/)
- [Changelog](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)
- [License](LICENSE)
