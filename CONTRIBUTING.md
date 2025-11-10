# Contributing to Conduit

Thank you for your interest in contributing to Conduit! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the problem
- **Expected behavior** vs actual behavior
- **Version information** (`conduit --version`)
- **Environment details** (OS, Go version)
- **Relevant logs or screenshots**

Use the bug report template when creating issues.

### Suggesting Features

Feature suggestions are welcome! When suggesting a feature:

- **Use the feature request template**
- **Explain the problem** it solves
- **Describe the proposed solution** clearly
- **Consider alternatives** you've thought about
- **Indicate the impact** (which components it affects)

### Contributing Code

#### Getting Started

1. **Fork the repository**
   ```bash
   gh repo fork scttfrdmn/conduit
   ```

2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/conduit.git
   cd conduit
   ```

3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/scttfrdmn/conduit.git
   ```

4. **Install dependencies**
   ```bash
   make deps
   make tools
   ```

#### Development Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, readable code
   - Follow Go best practices
   - Add tests for new functionality
   - Update documentation as needed

3. **Run tests and linters**
   ```bash
   make test
   make lint
   make vet
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

   Follow [Conventional Commits](https://www.conventionalcommits.org/) format:
   - `feat:` new feature
   - `fix:` bug fix
   - `docs:` documentation changes
   - `style:` formatting changes
   - `refactor:` code refactoring
   - `test:` adding tests
   - `chore:` maintenance tasks

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**
   - Use the PR template
   - Link related issues
   - Provide clear description of changes

#### Code Style

- **Go Format**: Use `gofmt` and `goimports`
  ```bash
  make fmt
  ```

- **Linting**: Must pass `golangci-lint`
  ```bash
  make lint
  ```

- **Documentation**: Comment exported functions and types
  ```go
  // MethodName does something important.
  // It takes a parameter and returns a result.
  func MethodName(param string) error {
      // ...
  }
  ```

- **Error Handling**: Always handle errors appropriately
  ```go
  if err != nil {
      return fmt.Errorf("operation failed: %w", err)
  }
  ```

#### Testing

- **Write tests** for all new functionality
- **Maintain coverage** above 80%
- **Use table-driven tests** where appropriate

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "expected",
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

Run tests:
```bash
make test          # All tests
make test-short    # Short tests only
make coverage      # Generate coverage report
```

#### Project Structure

```
conduit/
├── cmd/                    # Command-line applications
│   └── conduit/           # Main CLI application
├── internal/              # Private application code
│   ├── catalog/           # Model catalog
│   ├── deploy/            # Deployment engine
│   ├── model/             # Model parsing and validation
│   ├── sigstore/          # Signing and verification
│   └── version/           # Version information
├── pkg/                   # Public API
│   └── types/             # Shared types
├── docs/                  # Documentation
├── examples/              # Example models
└── scripts/               # Build and utility scripts
```

### Contributing Documentation

Documentation improvements are always welcome!

- **Code comments**: Ensure all exported functions are documented
- **README updates**: Keep installation and usage instructions current
- **New guides**: Add guides to `docs/` directory
- **Examples**: Add practical examples to `examples/`

#### Documentation Style

- Use clear, concise language
- Include code examples
- Add screenshots where helpful
- Keep formatting consistent

### Contributing Models

To contribute example scientific models:

1. **Create model repository**
   - Use `conduit init --template <domain>`
   - Add complete `model.yaml`
   - Include inference code
   - Add requirements.txt

2. **Add benchmarks**
   - Run on standard datasets
   - Document methodology
   - Include cost metrics

3. **Test thoroughly**
   - Validate with `conduit validate`
   - Test deployment locally
   - Verify inference works

4. **Submit PR**
   - Link to model repository
   - Provide benchmark results
   - Include usage documentation

## Review Process

### Pull Request Reviews

- All PRs require at least one approval
- Maintainers will review within 48 hours
- Address review comments promptly
- Keep PRs focused and reasonably sized

### Review Criteria

- **Code quality**: Readable, maintainable, idiomatic Go
- **Tests**: Adequate coverage, meaningful assertions
- **Documentation**: Clear comments and docs
- **Performance**: No unnecessary performance degradation
- **Security**: No security vulnerabilities introduced

## Release Process

Releases follow semantic versioning (semver 2.0):

- **Major (v1.0.0)**: Breaking changes
- **Minor (v0.1.0)**: New features, backward compatible
- **Patch (v0.0.1)**: Bug fixes, backward compatible

### Creating a Release

1. Update CHANGELOG.md
2. Create version tag: `git tag -a v0.1.0 -m "Release v0.1.0"`
3. Push tag: `git push origin v0.1.0`
4. GitHub Actions automatically creates release

## Getting Help

- **GitHub Discussions**: Ask questions, share ideas
- **GitHub Issues**: Report bugs, request features
- **Pull Requests**: Contribute code and documentation

## Recognition

Contributors are recognized in:
- CHANGELOG.md for each release
- GitHub contributors page
- Project README

Thank you for contributing to Conduit! 🚀
