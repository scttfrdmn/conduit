# GitHub Project Management Setup

This document describes how to manage the Conduit project using GitHub Projects, Milestones, Issues, and Labels.

## GitHub Project Board

Create a GitHub Project (Projects tab) with the following setup:

### Board Views

1. **Kanban Board** (Default)
   - Backlog
   - Todo
   - In Progress
   - In Review
   - Done

2. **Roadmap View** (Timeline)
   - Organized by milestones
   - Shows dependencies

3. **Priority View** (Table)
   - Sorted by priority and milestone
   - Shows assignees and status

### Custom Fields

- **Priority**: P0 (Critical), P1 (High), P2 (Medium), P3 (Low)
- **Size**: XS, S, M, L, XL (story points)
- **Domain**: Core, Protein Science, Materials, Drug Discovery, CLI, API, UI, Docs, DevOps
- **Status**: Backlog, Todo, In Progress, In Review, Done, Blocked

## Milestones

Create the following milestones in GitHub Issues:

### v0.1.0 - MVP (Target: 4 weeks)

**Goals:**
- Core CLI commands (init, validate, publish)
- Model.yaml parser and validator
- Basic Bedrock deployment
- PostgreSQL catalog backend
- Simple web catalog viewer

**Completion Criteria:**
- [ ] Can publish a model to catalog
- [ ] Can deploy model to Bedrock
- [ ] Can run inference via CLI
- [ ] Documentation complete
- [ ] Tests passing with >80% coverage

### v0.2.0 - Protein Science Suite (Target: 8 weeks)

**Goals:**
- 20+ protein science models published
- Studio Lab notebook generation
- Benchmarking framework
- Streamlit/Gradio UI generation
- Sigstore signing (basic)

**Completion Criteria:**
- [ ] 20+ models in catalog
- [ ] Auto-generated notebooks for all models
- [ ] Benchmark results for top 10 models
- [ ] Signing/verification working
- [ ] Example workflows documented

### v0.3.0 - Agentic Workflows (Target: 12 weeks)

**Goals:**
- Agentic workflow engine
- MCP server integration
- AWS Strands support
- Advanced deployment (spot instances, multi-region)
- Enterprise features

**Completion Criteria:**
- [ ] 3+ agentic workflows working
- [ ] MCP integration functional
- [ ] Cost optimization features
- [ ] Enterprise auth/security
- [ ] ISC paper draft complete

### v0.4.0 - Multi-Domain (Target: 16 weeks)

**Goals:**
- Materials science models
- Drug discovery models
- Climate science (initial)
- Advanced catalog features
- Community building

## Labels

Create the following labels in GitHub Issues:

### Type Labels (Mutually Exclusive)
- `type: bug` 🐛 - Something isn't working
- `type: feature` ✨ - New feature or request
- `type: enhancement` 🚀 - Improvement to existing feature
- `type: documentation` 📚 - Documentation improvements
- `type: refactor` 🔨 - Code refactoring
- `type: test` 🧪 - Test-related changes
- `type: ci/cd` ⚙️ - CI/CD pipeline changes
- `type: security` 🔒 - Security-related issues

### Priority Labels
- `priority: P0 - critical` 🔥 - Must be fixed immediately
- `priority: P1 - high` ⬆️ - Should be fixed soon
- `priority: P2 - medium` ➡️ - Should be fixed eventually
- `priority: P3 - low` ⬇️ - Nice to have

### Status Labels
- `status: blocked` 🚫 - Blocked by another issue
- `status: wip` 🚧 - Work in progress
- `status: needs-review` 👀 - Needs code review
- `status: needs-discussion` 💬 - Needs team discussion
- `status: ready` ✅ - Ready to be worked on

### Domain Labels
- `domain: core` - Core infrastructure
- `domain: cli` - Command-line interface
- `domain: api` - REST API
- `domain: catalog` - Model catalog
- `domain: deployment` - Deployment engine
- `domain: signing` - Sigstore integration
- `domain: notebooks` - Notebook generation
- `domain: ui` - Web interface
- `domain: docs` - Documentation
- `domain: protein-science` - Protein science models
- `domain: materials` - Materials science
- `domain: drug-discovery` - Drug discovery
- `domain: climate` - Climate science

### Size Labels (Story Points)
- `size: XS` - 1 point (< 2 hours)
- `size: S` - 2 points (2-4 hours)
- `size: M` - 3 points (4-8 hours)
- `size: L` - 5 points (1-2 days)
- `size: XL` - 8 points (2+ days)

### Special Labels
- `good first issue` 👶 - Good for newcomers
- `help wanted` 🙏 - Extra attention needed
- `dependencies` 📦 - Dependency updates
- `breaking change` 💥 - Breaking API changes
- `needs-triage` 🔍 - Needs initial review

## Issue Templates

We have two issue templates:

1. **Bug Report** (`.github/ISSUE_TEMPLATE/bug_report.yml`)
   - Automatically adds `type: bug` and `needs-triage` labels

2. **Feature Request** (`.github/ISSUE_TEMPLATE/feature_request.yml`)
   - Automatically adds `type: feature` and `needs-triage` labels

## Pull Request Process

### PR Labels

PRs inherit labels from linked issues, plus:
- `size: XS/S/M/L/XL` - Based on lines changed
- `needs-review` - Waiting for review
- `changes-requested` - Reviewer requested changes
- `approved` - Approved and ready to merge

### PR Workflow

1. **Create PR** - Use template, link issues
2. **CI Runs** - Tests, linting, build
3. **Review** - At least one approval required
4. **Merge** - Squash and merge to main
5. **Auto-close** - Linked issues close automatically

### Branch Naming Convention

- `feat/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation
- `refactor/description` - Refactoring
- `test/description` - Tests
- `chore/description` - Maintenance

## Automation

### GitHub Actions Workflows

We have automation for:

1. **CI Pipeline** (`.github/workflows/ci.yml`)
   - Runs on every push and PR
   - Tests, linting, build
   - Code coverage reporting

2. **Release** (`.github/workflows/release.yml`)
   - Triggers on version tags
   - Creates GitHub release
   - Publishes binaries via GoReleaser
   - Builds Docker images

### Automated Labels

GitHub Actions can automatically:
- Add `size: XS/S/M/L/XL` based on PR diff
- Add `needs-review` when PR is opened
- Add `approved` when PR is approved
- Update project board status

## Project Board Automation

Set up these automations in GitHub Projects:

1. **New Issue** → Backlog column
2. **Issue assigned** → Todo column
3. **PR opened** → In Progress column
4. **PR approved** → In Review column
5. **PR merged** → Done column
6. **Issue closed** → Done column

## Weekly Workflow

### Monday
- Review backlog
- Prioritize issues for the week
- Assign issues to team members
- Update milestone progress

### Wednesday
- Check blocked issues
- Review in-progress work
- Update documentation

### Friday
- Review completed work
- Update CHANGELOG.md
- Plan next week's priorities
- Demo new features

## Release Process

1. **Create Release Branch**
   ```bash
   git checkout -b release/v0.1.0
   ```

2. **Update Version Files**
   - Update CHANGELOG.md
   - Version in code (handled by goreleaser)

3. **Create PR**
   - Label: `type: release`
   - Milestone: v0.1.0
   - Get approval

4. **Merge and Tag**
   ```bash
   git checkout main
   git merge release/v0.1.0
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin main --tags
   ```

5. **GitHub Actions Auto-Release**
   - GoReleaser creates release
   - Binaries built and uploaded
   - Docker images published
   - Changelog generated

6. **Announce**
   - Update README with new version
   - Post in Discussions
   - Tweet/blog post

## Metrics to Track

### Development Metrics
- Issues opened vs closed per week
- PR merge time (target: <24 hours)
- CI success rate (target: >95%)
- Test coverage (target: >80%)
- Go Report Card (target: A+)

### Product Metrics
- Models published
- Deployments created
- CLI downloads
- GitHub stars
- Community contributions

## Communication

### GitHub Discussions
- **Announcements** - Project updates
- **General** - Open discussion
- **Ideas** - Feature proposals
- **Q&A** - Help and questions
- **Show and Tell** - Community models

### Issue Comments
- Keep discussion on-topic
- Tag relevant people with @mentions
- Reference related issues with #123
- Link to code with line numbers

### PR Reviews
- Be constructive and respectful
- Suggest improvements clearly
- Approve promptly when ready
- Use GitHub suggestions for small fixes

## Resources

- [GitHub Projects Documentation](https://docs.github.com/en/issues/planning-and-tracking-with-projects)
- [GitHub Labels Best Practices](https://github.com/joelparkerhenderson/github-special-files-and-paths)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)

---

**Last Updated**: 2024-11-10
