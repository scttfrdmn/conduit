# Conduit GitHub Configuration

This directory contains GitHub-specific configuration files for the Conduit project.

## Directory Structure

```
.github/
├── workflows/              # GitHub Actions CI/CD
│   ├── ci.yml             # Continuous Integration
│   └── release.yml        # Release automation
├── ISSUE_TEMPLATE/        # Issue templates
│   ├── bug_report.yml     # Bug report form
│   └── feature_request.yml # Feature request form
├── PULL_REQUEST_TEMPLATE/ # PR templates
│   └── pull_request_template.md
├── PROJECT_SETUP.md       # Project management guide
└── README.md             # This file
```

## Workflows

### CI Workflow (`workflows/ci.yml`)

Runs on every push and pull request to `main` and `develop` branches.

**Jobs:**
- **Test**: Run tests on Ubuntu, macOS, and Windows with Go 1.22 and 1.23
- **Lint**: Run golangci-lint
- **Build**: Build the binary and verify it works

**Triggers:**
- Push to main/develop
- Pull request to main/develop

### Release Workflow (`workflows/release.yml`)

Runs when a version tag is pushed (e.g., `v0.1.0`).

**Jobs:**
- Run tests
- Build binaries for all platforms via GoReleaser
- Create GitHub release
- Build and push Docker images to GHCR

**Triggers:**
- Push of tags matching `v*`

## Issue Templates

### Bug Report (`ISSUE_TEMPLATE/bug_report.yml`)

Structured form for bug reports including:
- Bug description
- Steps to reproduce
- Expected vs actual behavior
- Version information
- System details
- Log output

**Automatic Labels:**
- `bug`
- `needs-triage`

### Feature Request (`ISSUE_TEMPLATE/feature_request.yml`)

Structured form for feature requests including:
- Problem statement
- Proposed solution
- Alternatives considered
- Scientific domain
- Impact areas

**Automatic Labels:**
- `enhancement`
- `needs-triage`

## Pull Request Template

Located in `PULL_REQUEST_TEMPLATE/pull_request_template.md`.

Includes sections for:
- Description
- Related issues
- Type of change
- Changes made
- Testing performed
- Checklist (style, tests, docs, etc.)

## Project Management

See `PROJECT_SETUP.md` for comprehensive guide on:
- GitHub Projects setup
- Milestones
- Labels
- Workflows
- Development process

## Setup Instructions

### 1. Create GitHub Repository

```bash
# Using GitHub CLI
gh repo create conduit --public --source=. --remote=origin

# Push code
git push -u origin main --tags
```

### 2. Enable GitHub Actions

Actions should be enabled by default. Verify at:
- `https://github.com/scttfrdmn/conduit/actions`

### 3. Configure Repository Secrets

For release workflow:

```bash
# Generate GitHub token with packages:write permission
gh auth token

# Add as secret (done automatically if using gh)
# Name: GITHUB_TOKEN (available by default)
```

### 4. Set Up GitHub Project

1. Go to `https://github.com/scttfrdmn/conduit/projects`
2. Click "New project"
3. Choose "Board" view
4. Follow `PROJECT_SETUP.md` for configuration

### 5. Create Milestones

1. Go to `https://github.com/scttfrdmn/conduit/milestones`
2. Click "New milestone"
3. Create milestones as described in `PROJECT_SETUP.md`:
   - v0.1.0 - MVP
   - v0.2.0 - Protein Science Suite
   - v0.3.0 - Agentic Workflows

### 6. Add Labels

GitHub provides default labels. Add custom ones via:

```bash
# Using GitHub CLI
gh label create "domain: core" --description "Core infrastructure" --color "0052cc"
gh label create "priority: P0 - critical" --description "Must fix immediately" --color "d73a4a"
# ... etc (see PROJECT_SETUP.md for full list)
```

Or use the web interface:
- Go to `https://github.com/scttfrdmn/conduit/labels`
- Create labels as per `PROJECT_SETUP.md`

### 7. Branch Protection

Configure branch protection for `main`:

1. Go to Settings → Branches
2. Add rule for `main`
3. Enable:
   - ✓ Require pull request reviews before merging
   - ✓ Require status checks to pass before merging
     - Select: `test`, `lint`, `build`
   - ✓ Require branches to be up to date before merging
   - ✓ Include administrators

### 8. Enable Discussions

1. Go to Settings → Features
2. Enable "Discussions"
3. Create categories:
   - Announcements
   - General
   - Ideas
   - Q&A
   - Show and Tell

## Automated Processes

### When You Push a Tag

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

**What happens automatically:**
1. Release workflow triggers
2. Tests run
3. Binaries built for all platforms
4. GitHub release created with changelog
5. Docker images built and pushed to GHCR
6. Release artifacts uploaded

### When You Open a PR

**What happens automatically:**
1. CI workflow triggers
2. Tests run on all platforms
3. Linter checks code
4. Build verifies compilation
5. Status checks appear on PR
6. PR can't merge until checks pass

### When You Create an Issue

**What happens automatically:**
1. Issue template loads based on type
2. Labels auto-applied based on template
3. Issue appears in Project board (if connected)

## Maintenance

### Updating Workflows

Edit files in `workflows/` directory and commit. Changes take effect immediately on next trigger.

### Updating Templates

Edit files in `ISSUE_TEMPLATE/` or `PULL_REQUEST_TEMPLATE/` and commit. Changes take effect immediately.

### Adding New Workflows

Create new `.yml` files in `workflows/`. Examples:

```yaml
# .github/workflows/security.yml
name: Security Scan
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Gosec
        run: gosec ./...
```

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GoReleaser Documentation](https://goreleaser.com/)
- [GitHub Issue Templates](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests)
- [Branch Protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches)

---

**Last Updated**: 2024-11-10
