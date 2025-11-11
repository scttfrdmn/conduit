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
- Example model.yaml demonstrating all fields
- Test suite with 73.3% coverage for model package

### Changed

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
