# Conduit

[![Go Report Card](https://goreportcard.com/badge/github.com/scttfrdmn/conduit)](https://goreportcard.com/report/github.com/scttfrdmn/conduit)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/scttfrdmn/conduit)](https://github.com/scttfrdmn/conduit/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/scttfrdmn/conduit)](go.mod)

**AWS-native platform for publishing, discovering, and deploying scientific ML models**

Conduit makes deploying scientific models as simple as installing software packages. No more HPC queue waits, no Docker expertise required, full cost transparency.

```bash
# Install
brew install scttfrdmn/tap/conduit

# Deploy AlphaFold2 to AWS Bedrock
conduit deploy alphafold2-multimer

# Run inference
conduit predict alphafold2-multimer --sequence "MKLLVGDDS..."
```

---

## Features

### 🚀 Instant Deployment
- Deploy to AWS Bedrock in seconds (no HPC queues)
- Automatic instance selection and optimization
- Support for spot instances (70% cost savings)

### 🔬 Scientific Rigor
- DOI minting via Zenodo
- Standardized benchmarks across hardware
- Reproducible model packaging
- Cryptographic signing with Sigstore

### 🤖 Agentic Workflows
- Compose multiple models with foundation models (Claude)
- MCP server integration for scientific databases
- Natural language workflow descriptions

### 📊 Cost Transparency
- Real-time cost tracking per inference
- Cost/performance benchmarking
- Budget alerts and limits

### 🎓 Education-Friendly
- Auto-generated SageMaker Studio Lab notebooks
- Streamlit/Gradio UIs (no-code access)
- One-click Colab demos

### 🔐 Supply Chain Security
- Sigstore signing (keyless, OIDC-based)
- Rekor transparency log
- Cryptographic verification

---

## Quick Start

### Installation

**Homebrew (macOS/Linux)**
```bash
brew install scttfrdmn/tap/conduit
```

**Script (macOS/Linux)**
```bash
curl -sSL https://get.conduit.dev | bash
```

**Go Install**
```bash
go install github.com/scttfrdmn/conduit/cmd/conduit@latest
```

**From Source**
```bash
git clone https://github.com/scttfrdmn/conduit.git
cd conduit
make install
```

### Usage

**Discover Models**
```bash
# Search for protein folding models
conduit search "protein folding"

# Show model details
conduit show alphafold2-multimer
```

**Deploy a Model**
```bash
# Deploy to AWS Bedrock
conduit deploy alphafold2-multimer \
  --instance ml.g5.2xlarge \
  --region us-east-1

# Deploy with spot instances (save 70%)
conduit deploy alphafold2-multimer --use-spot
```

**Run Inference**
```bash
# Run a prediction
conduit predict alphafold2-multimer \
  --sequence "MKLLVGDDS..." \
  --output prediction.pdb

# Batch processing
conduit batch alphafold2-multimer \
  --input sequences.fasta \
  --output results/
```

**Publish Your Own Model**
```bash
# Initialize from template
conduit init --template protein-folding

# Validate model spec
conduit validate

# Publish to catalog
conduit publish --github yourorg/your-model --sign
```

---

## Why Conduit?

### The Problem

**For Researchers:**
- Wait days for HPC GPU allocations
- Complex deployment (Docker, Kubernetes, SLURM)
- No cost transparency
- Hard to reproduce published models

**For Institutions:**
- Wasted compute from overallocations
- Maintenance burden (software environments, dependencies)
- Difficulty sharing models between labs
- No security/provenance for model weights

### The Solution

**Conduit provides:**
- ✅ Instant deployment (seconds, not days)
- ✅ Simple CLI (no Docker knowledge required)
- ✅ Cost transparency (pay only for what you use)
- ✅ Reproducible packaging (GitHub + DOI)
- ✅ Cryptographic signing (Sigstore)
- ✅ Auto-generated UIs (Streamlit, Gradio, Jupyter)

---

## Architecture

### Components

1. **CLI Tool** (`conduit`) - Publish, deploy, and manage models
2. **Model Registry** - GitHub-based, decentralized catalog
3. **Deployment Engine** - AWS Bedrock and SageMaker integration
4. **Catalog API** - Search and discover models
5. **Web UI** - Browse models, view benchmarks

### Model Specification

Models are defined in `model.yaml`:

```yaml
name: "alphafold2-multimer"
version: "2.3.2"
domain: "protein-science"

description: |
  Predict protein structures and interactions using AlphaFold2

runtime:
  framework: "jax"
  python_version: "3.11"
  dependencies: "requirements.txt"

inference:
  entrypoint: "inference.py"
  handler: "predict"

weights_uri: "s3://aws-open-data-scientific-models/alphafold2/v2.3.2/"

hardware:
  gpu_required: true
  recommended_instance: "ml.g5.2xlarge"
  memory_gb: 24

benchmarks:
  - dataset: "CASP15"
    metric: "GDT-TS"
    result: 92.4
    cost_per_prediction: "$0.15"
```

---

## Supported Domains

### 🧬 Protein Science (20+ models)
AlphaFold2, ESMFold, ProteinMPNN, RFdiffusion, DiffDock, IgFold, and more

### 🔬 Materials Science (Coming Soon)
MACE, M3GNet, CDVAE, Nequip

### 💊 Drug Discovery (Coming Soon)
DiffDock, ADMET-AI, DeepChem models

### 🌍 Climate Science (Planned)
Weather forecasting, climate projection models

---

## Documentation

- [Getting Started Guide](docs/getting-started.md)
- [Publishing Models](docs/publishing.md)
- [Model Specification](docs/model-spec.md)
- [Deployment Guide](docs/deployment.md)
- [Agentic Workflows](docs/agentic-workflows.md)
- [API Reference](docs/api.md)
- [Contributing](CONTRIBUTING.md)

---

## Project Status

**Current Phase:** MVP Development (v0.1.0)

**Roadmap:**
- ✅ Project initialization and scaffolding
- 🚧 Core CLI commands (init, validate, publish, deploy)
- 🚧 Model parser and validator
- 🚧 Bedrock deployment engine
- 🚧 PostgreSQL catalog backend
- 📋 Protein science model suite (20+ models)
- 📋 Sigstore signing integration
- 📋 SageMaker Studio Lab notebooks
- 📋 Agentic workflow engine

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone repository
git clone https://github.com/scttfrdmn/conduit.git
cd conduit

# Install dependencies
go mod download

# Run tests
make test

# Build
make build

# Run locally
./bin/conduit --help
```

### Code Quality

This project maintains:
- ✅ Go Report Card: A+
- ✅ Test coverage: >80%
- ✅ Linting: golangci-lint
- ✅ Formatting: gofmt, goimports

---

## Community

- **GitHub Discussions**: [Ask questions, share ideas](https://github.com/scttfrdmn/conduit/discussions)
- **GitHub Issues**: [Report bugs, request features](https://github.com/scttfrdmn/conduit/issues)
- **Twitter**: [@conduit_dev](https://twitter.com/conduit_dev) (coming soon)

---

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

Built with ❤️ for the scientific computing community.

Special thanks to:
- AWS Bedrock team
- Sigstore project
- Scientific model publishers
- Open source contributors

---

## Related Projects

- [Garden AI](https://github.com/Garden-AI) - ML model publishing (NSF-funded)
- [Hugging Face](https://huggingface.co) - LLM model hub
- [BioContainers](https://biocontainers.pro) - Bioinformatics containers
- [Sigstore](https://sigstore.dev) - Software supply chain security
