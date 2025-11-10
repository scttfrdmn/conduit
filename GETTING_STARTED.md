# Getting Started with Conduit Development

## ✅ What's Been Set Up

Your professional Go project is ready! Here's what we've built:

### Core Infrastructure
- ✅ Git repository initialized with professional structure
- ✅ Go module (`github.com/scttfrdmn/conduit`)
- ✅ Semantic versioning (semver 2.0) - Tagged as v0.0.1
- ✅ Keep a Changelog format (CHANGELOG.md)
- ✅ GoReleaser configuration
- ✅ Apache 2.0 License

### Code Quality
- ✅ golangci-lint configured and passing (0 issues)
- ✅ Makefile with standard targets
- ✅ Dockerfile for containerization
- ✅ Tests configured (go test with race detector)
- ✅ Code validated and building successfully

### GitHub Integration
- ✅ CI/CD workflows (test, lint, build)
- ✅ Release automation via GitHub Actions
- ✅ Issue templates (bug report, feature request)
- ✅ Pull request template
- ✅ Project management guide (.github/PROJECT_SETUP.md)

### Documentation
- ✅ Comprehensive README.md
- ✅ CONTRIBUTING.md with dev workflow
- ✅ PROJECT_OVERVIEW.md (your excellent summary!)
- ✅ Design docs preserved (DESIGN_CONVO.md, specs)

### Initial Code
- ✅ CLI framework (Cobra-based)
- ✅ Version management with ldflags
- ✅ Core types (Model, Runtime, Inference, etc.)
- ✅ Project structure following Go best practices

---

## 📦 Project Structure

```
conduit/
├── cmd/
│   └── conduit/           # Main CLI application
│       └── main.go        # Entry point
├── internal/              # Private application code
│   ├── cli/              # CLI commands (root.go)
│   ├── version/          # Version info
│   ├── model/            # Model parsing (TODO)
│   ├── catalog/          # Catalog backend (TODO)
│   ├── deploy/           # Deployment engine (TODO)
│   └── sigstore/         # Signing/verification (TODO)
├── pkg/
│   └── types/            # Public API types (model.go)
├── docs/                 # Documentation
├── examples/             # Example models
├── scripts/              # Build scripts
├── .github/
│   ├── workflows/        # CI/CD
│   ├── ISSUE_TEMPLATE/   # Issue templates
│   └── PULL_REQUEST_TEMPLATE/
├── Makefile              # Build automation
├── Dockerfile            # Container image
├── .golangci.yml         # Linter config
├── .goreleaser.yaml      # Release automation
└── go.mod                # Go dependencies
```

---

## 🚀 Quick Commands

### Build & Run
```bash
make build              # Build binary to ./bin/conduit
./bin/conduit --help    # Run CLI
./bin/conduit --version # Show version
```

### Development
```bash
make deps               # Download dependencies
make test               # Run tests
make lint               # Run linter
make vet                # Run go vet
make fmt                # Format code
make validate           # Run all checks
make clean              # Clean build artifacts
```

### Tools Installation
```bash
make tools              # Install dev tools (golangci-lint, etc.)
```

### CI Simulation
```bash
make ci                 # Run full CI pipeline locally
```

### Release
```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0  # Triggers automatic release
```

---

## 📋 Next Steps

### Immediate (Week 1-2)

1. **Set up GitHub**
   ```bash
   # Create GitHub repo
   gh repo create conduit --public --source=. --remote=origin

   # Push code
   git branch -M main
   git push -u origin main --tags

   # Set up Project board
   # Follow .github/PROJECT_SETUP.md
   ```

2. **Implement Core Parser** (`internal/model/parser.go`)
   - Parse model.yaml files
   - Validate against schema
   - Add comprehensive tests

3. **Add CLI Commands**
   - `conduit init` - Create new model from template
   - `conduit validate` - Validate model.yaml
   - `conduit info` - Show model information

4. **Database Schema**
   - Set up PostgreSQL migrations
   - Create catalog tables
   - Add connection management

### Phase 1 (Week 3-4) - MVP

5. **Catalog Backend**
   - Model registration
   - Search functionality
   - Metadata indexing

6. **Deployment Engine**
   - AWS Bedrock integration
   - Instance selection logic
   - Cost calculation

7. **Web UI**
   - Simple Next.js catalog viewer
   - Model detail pages
   - Search interface

### Phase 2 (Week 5-8) - Protein Science

8. **Example Models**
   - AlphaFold2-multimer
   - ESMFold
   - ProteinMPNN
   - 17 more models

9. **Benchmarking**
   - Standard datasets
   - Cost/performance tracking
   - Results storage

10. **Notebooks & UIs**
    - Studio Lab notebooks
    - Streamlit generation
    - Gradio interfaces

---

## 🧪 Testing Strategy

### Current Status
Tests are configured but no test files yet.

### Add Tests For
```go
// internal/model/parser_test.go
func TestParseModelYAML(t *testing.T) {
    // Test valid model.yaml parsing
}

// internal/model/validator_test.go
func TestValidateModel(t *testing.T) {
    // Test validation logic
}

// pkg/types/model_test.go
func TestModelSerialization(t *testing.T) {
    // Test YAML ↔ struct conversion
}
```

### Coverage Goal
- Target: >80% coverage
- Run: `make coverage` to generate HTML report

---

## 🎯 GitHub Project Setup

Follow `.github/PROJECT_SETUP.md` to set up:

1. **GitHub Project Board**
   - Kanban view
   - Roadmap timeline
   - Priority table

2. **Milestones**
   - v0.1.0 - MVP (4 weeks)
   - v0.2.0 - Protein Science (8 weeks)
   - v0.3.0 - Agentic Workflows (12 weeks)

3. **Labels**
   - Type: bug, feature, docs, etc.
   - Priority: P0-P3
   - Domain: cli, api, catalog, etc.
   - Size: XS, S, M, L, XL

4. **Issue Templates**
   - Already configured!
   - Bug reports
   - Feature requests

---

## 📚 Key Documents

| Document | Purpose |
|----------|---------|
| PROJECT_OVERVIEW.md | High-level project vision |
| DESIGN_CONVO.md | Complete design discussion |
| go-implementation-plan.md | Technical implementation plan |
| protein-science-suite-spec.md | Exemplar domain spec |
| sagemaker-integration-spec.md | Studio Lab/SageMaker integration |
| sigstore-integration-spec.md | Cryptographic signing |
| .github/PROJECT_SETUP.md | GitHub project management |
| CONTRIBUTING.md | Contribution guidelines |
| README.md | User-facing documentation |

---

## 🔧 Development Workflow

### 1. Pick an Issue
```bash
# View issues in GitHub or create one
gh issue list

# Create new issue
gh issue create --title "Implement model parser" \
  --label "type: feature,domain: core,priority: P1"
```

### 2. Create Branch
```bash
git checkout -b feat/model-parser
```

### 3. Develop
```bash
# Write code
# Add tests
make test
make lint
```

### 4. Commit
```bash
git add .
git commit -m "feat: implement model.yaml parser

- Add YAML unmarshaling
- Validate required fields
- Add comprehensive tests"
```

### 5. Push & PR
```bash
git push origin feat/model-parser
gh pr create --title "feat: implement model parser" \
  --body "Implements #123"
```

### 6. CI Checks
- Tests run automatically
- Linter runs automatically
- Must pass before merge

### 7. Review & Merge
- Get approval
- Squash and merge
- Issue auto-closes

---

## 🎓 Learning Resources

### Go
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Project Management
- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

### AWS
- [Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)
- [SageMaker Studio Lab](https://studiolab.sagemaker.aws/)
- [AWS SDK for Go V2](https://aws.github.io/aws-sdk-go-v2/)

---

## 🐛 Troubleshooting

### Build Issues
```bash
# Clean and rebuild
make clean
make deps
make build
```

### Linter Errors
```bash
# Format code
make fmt

# Run linter
make lint

# Check specific file
golangci-lint run path/to/file.go
```

### Test Failures
```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./internal/model -run TestParseModel
```

---

## 🎉 You're Ready!

Your project is fully set up with professional standards:
- ✅ Semantic versioning
- ✅ Keep a Changelog
- ✅ GoReleaser for releases
- ✅ Go Report Card ready (will be A+ with your code)
- ✅ GitHub Projects/Issues/Milestones ready
- ✅ CI/CD automated

**Next Action**: Push to GitHub and start coding!

```bash
# Create GitHub repo and push
gh repo create conduit --public --source=. --remote=origin
git push -u origin main --tags

# Start development
gh issue create --title "Implement model.yaml parser" \
  --label "type: feature,domain: core,priority: P0,size: M"

# Begin coding!
git checkout -b feat/model-parser
```

---

**Happy coding! Let's build something amazing! 🚀**
