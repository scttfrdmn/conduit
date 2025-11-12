# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Model.yaml parser with `ParseFile()`, `Parse()`, and `ParseString()` methods
- Comprehensive validator with clear error messages
  - Name format validation (lowercase-with-hyphens)
  - Semantic version validation (semver 2.0)
  - URI validation (s3://, hf://, https://)
  - Framework validation against allowed list
  - Hardware requirements validation
  - Benchmark data validation
- CLI `validate` command for model.yaml files
- CLI `init` command to scaffold new model projects
  - Interactive prompts for model metadata
  - Automatic directory structure generation (inference/, weights/, tests/)
  - Template files (inference.py, requirements.txt, test_inference.py, README.md)
  - Framework-specific requirements templates
  - Support for all major ML frameworks (PyTorch, TensorFlow, JAX, ONNX)
  - --force flag to overwrite existing files
- Example model.yaml demonstrating all fields
- Test suite with 73.3% coverage for model package
- Comprehensive test suite for init command
- Model catalog with SQLite database backend
  - Database schema for models, versions, benchmarks, and citations
  - CRUD operations for model registry
  - Full-text search across model metadata
  - Filter by domain, framework, and GPU requirements
  - Automatic schema initialization and migrations
- CLI `publish` command to register models in catalog
  - Validates model.yaml before registration
  - Supports versioning - can publish multiple versions of same model
  - Automatically marks new versions as latest
  - Prevents duplicate version publishing
  - Stores model metadata and benchmarks
- CLI `search` command to query catalog
  - Search by keyword across names, domains, descriptions
  - Filter by domain, framework, GPU requirement
  - Pagination and result limiting
  - User-friendly result display
- CLI `info` command to view detailed model information
  - Display model metadata, version details, and hardware requirements
  - Show benchmark results and performance metrics
  - Display citation information with formatted author list
  - Comprehensive view of all model attributes
- CLI `versions` command to list model version history
  - Shows all versions in reverse chronological order
  - Indicates which version is marked as latest
  - Displays creation dates and key attributes
- Model versioning support in catalog
  - CreateModelVersion() method to add new versions
  - GetModelVersion() to retrieve specific version
  - ListModelVersions() to show version history
  - Automatic latest version tracking
  - Version uniqueness validation
- Comprehensive test suite for catalog operations (13 tests)

### Changed
- CI pipeline now tests only Go 1.23 (previously tested 1.22 and 1.23)

### Deprecated

### Removed

### Fixed

### Security

---

## [0.0.1] - 2024-11-10

### Added
- Initial project structure with Go best practices
- Professional Makefile with common development targets
- golangci-lint configuration (0 issues, ready for A+ rating)
- GitHub Actions CI/CD workflows (test, lint, build, release)
- GoReleaser configuration for multi-platform releases
- Dockerfile for containerization
- Issue templates (bug report, feature request)
- Pull request template with comprehensive checklist
- Apache 2.0 license
- Semantic versioning setup (semver 2.0)
- Keep a Changelog format
- Basic CLI framework using Cobra
- Version management with build-time ldflags
- Core type definitions (Model, Runtime, Inference, Hardware, etc.)
- Comprehensive documentation:
  - PROJECT_OVERVIEW.md - Executive summary
  - GETTING_STARTED.md - Developer guide
  - CONTRIBUTING.md - Contribution workflow
  - .github/PROJECT_SETUP.md - GitHub project management
  - .github/README.md - GitHub configuration guide
- Design documentation preserved:
  - DESIGN_CONVO.md - Complete design conversation
  - go-implementation-plan.md - Technical implementation plan
  - protein-science-suite-spec.md - Protein science exemplar
  - sagemaker-integration-spec.md - SageMaker Studio Lab integration
  - sigstore-integration-spec.md - Cryptographic signing specification

### Changed
- N/A (initial release)

### Fixed
- golangci-lint configuration compatibility with latest version
- Error handling in CLI root command

[0.0.1]: https://github.com/scttfrdmn/conduit/releases/tag/v0.0.1

---

## Release Template

Use this template for new releases:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Removed features

### Fixed
- Bug fixes

### Security
- Security fixes
```
