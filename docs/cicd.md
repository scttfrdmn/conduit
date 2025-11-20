# CI/CD Guide

Complete guide to automating ML model workflows with Conduit and GitHub Actions.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Generated Workflows](#generated-workflows)
- [Validation Workflow](#validation-workflow)
- [Publish Workflow](#publish-workflow)
- [Deploy Workflow](#deploy-workflow)
- [Custom Workflows](#custom-workflows)
- [GitHub Secrets](#github-secrets)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Advanced Topics](#advanced-topics)

---

## Overview

Conduit's CI/CD features automate model validation, publishing, and deployment using GitHub Actions.

**Key Features**:
- **Automatic Validation** - Run on every PR to catch errors early
- **Automated Publishing** - Push to registries on release
- **Deployment Automation** - Deploy to AWS on release or manually
- **Status Checks** - Block PRs with invalid models
- **Audit Trail** - Track all changes via Git history

**Workflow Triggers**:
- **On Pull Request** → Validate models
- **On Release** → Publish to registry and/or deploy
- **Manual Trigger** → Deploy on demand

---

## Quick Start

### Step 1: Generate Workflows

```bash
# Generate all CI/CD workflows
conduit workflow all
```

This creates three GitHub Actions workflows:
- `.github/workflows/conduit-validate.yml` - Validates on PRs
- `.github/workflows/conduit-publish.yml` - Publishes on releases
- `.github/workflows/conduit-deploy.yml` - Deploys to AWS

### Step 2: Configure Secrets

Add required secrets in GitHub:

1. Go to your repository → Settings → Secrets → Actions
2. Click "New repository secret"
3. Add these secrets:

```
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=us-east-1
SAGEMAKER_ROLE=arn:aws:iam::123456789012:role/SageMakerRole
REGISTRY_TOKEN=your_registry_token  # Optional: for private registries
```

### Step 3: Commit and Push

```bash
# Add workflows to git
git add .github/workflows/

# Commit
git commit -m "feat: add CI/CD workflows"

# Push
git push origin main
```

### Step 4: Test

Create a pull request to test validation:

```bash
# Create feature branch
git checkout -b feat/update-model

# Modify model
vim model.yaml

# Commit and push
git add model.yaml
git commit -m "feat: update model configuration"
git push origin feat/update-model

# Create PR on GitHub
# Validation workflow runs automatically
```

---

## Generated Workflows

### conduit-validate.yml

Runs on every pull request to validate model specifications.

**Generated Workflow**:
```yaml
name: Validate Models

on:
  pull_request:
    branches: [main, master]
    paths:
      - '**/*.yaml'
      - '**/*.yml'
      - '**/model.*'

jobs:
  validate:
    name: Validate Model Specifications
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Install Conduit
        run: |
          go install github.com/scttfrdmn/conduit/cmd/conduit@latest

      - name: Find model files
        id: find-models
        run: |
          MODELS=$(find . -name 'model.yaml' -o -name 'model.yml')
          echo "models<<EOF" >> $GITHUB_OUTPUT
          echo "$MODELS" >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Validate models
        run: |
          EXIT_CODE=0
          while IFS= read -r MODEL; do
            if [ -n "$MODEL" ]; then
              echo "Validating $MODEL..."
              if ! conduit validate --strict "$MODEL"; then
                EXIT_CODE=1
              fi
            fi
          done <<< "${{ steps.find-models.outputs.models }}"
          exit $EXIT_CODE

      - name: Comment PR
        if: failure()
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '❌ Model validation failed. Please fix errors and push again.'
            })
```

**What It Does**:
1. Triggers on PRs that modify YAML files
2. Installs Conduit
3. Finds all `model.yaml` files
4. Runs strict validation on each
5. Fails PR if any validation errors
6. Comments on PR with results

### conduit-publish.yml

Publishes models to registry when a release is created.

**Generated Workflow**:
```yaml
name: Publish Models

on:
  release:
    types: [published]

jobs:
  publish:
    name: Publish to Registry
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Install Conduit
        run: |
          go install github.com/scttfrdmn/conduit/cmd/conduit@latest

      - name: Configure registry
        run: |
          conduit registry add production ${{ secrets.REGISTRY_URL }} \
            --token ${{ secrets.REGISTRY_TOKEN }} \
            --set-default

      - name: Find and publish models
        run: |
          find . -name 'model.yaml' -o -name 'model.yml' | while read MODEL; do
            if [ -n "$MODEL" ]; then
              echo "Publishing $MODEL..."
              conduit add "$MODEL"
              MODEL_NAME=$(yq eval '.name' "$MODEL")
              conduit push "$MODEL_NAME"
            fi
          done

      - name: Create summary
        run: |
          echo "### Published Models" >> $GITHUB_STEP_SUMMARY
          conduit list --format json | jq -r '.[] | "- \(.name) v\(.version)"' >> $GITHUB_STEP_SUMMARY
```

**What It Does**:
1. Triggers when you create a GitHub release
2. Installs Conduit
3. Configures registry with credentials
4. Finds all models
5. Adds to catalog and pushes to registry
6. Creates summary of published models

### conduit-deploy.yml

Deploys models to AWS SageMaker.

**Generated Workflow**:
```yaml
name: Deploy Models

on:
  release:
    types: [published]
  workflow_dispatch:
    inputs:
      model_path:
        description: 'Path to model.yaml'
        required: true
        default: 'model.yaml'
      instance_type:
        description: 'Instance type'
        required: false
        default: 'ml.g5.xlarge'
      endpoint_name:
        description: 'Endpoint name'
        required: false

jobs:
  deploy:
    name: Deploy to AWS SageMaker
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Install Conduit
        run: |
          go install github.com/scttfrdmn/conduit/cmd/conduit@latest

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ secrets.AWS_REGION }}

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Deploy model
        run: |
          MODEL_PATH="${{ github.event.inputs.model_path || 'model.yaml' }}"
          INSTANCE_TYPE="${{ github.event.inputs.instance_type || '' }}"
          ENDPOINT_NAME="${{ github.event.inputs.endpoint_name || '' }}"

          DEPLOY_CMD="conduit deploy $MODEL_PATH --platform sagemaker --role ${{ secrets.SAGEMAKER_ROLE }}"

          if [ -n "$INSTANCE_TYPE" ]; then
            DEPLOY_CMD="$DEPLOY_CMD --instance-type $INSTANCE_TYPE"
          fi

          if [ -n "$ENDPOINT_NAME" ]; then
            DEPLOY_CMD="$DEPLOY_CMD --endpoint-name $ENDPOINT_NAME"
          fi

          eval $DEPLOY_CMD

      - name: Create summary
        if: success()
        run: |
          echo "### Deployment Successful" >> $GITHUB_STEP_SUMMARY
          echo "Model deployed to AWS SageMaker" >> $GITHUB_STEP_SUMMARY
          echo "Region: ${{ secrets.AWS_REGION }}" >> $GITHUB_STEP_SUMMARY
```

**What It Does**:
1. Triggers on release or manually
2. Installs Conduit
3. Configures AWS credentials
4. Sets up Docker
5. Deploys model to SageMaker
6. Creates deployment summary

---

## Validation Workflow

### Purpose

Catch model specification errors before they reach production.

### When It Runs

- Every pull request
- When any YAML file is modified
- On push to main/master (optional)

### What It Validates

**Basic Checks**:
- Required fields present
- Field types correct
- Version format valid
- Framework supported

**Strict Checks** (with `--strict` flag):
- Files exist (requirements.txt, predict.py)
- Handler function exists
- Dependencies are installable
- Model spec is deployable

### Customizing Validation

Edit `.github/workflows/conduit-validate.yml`:

```yaml
# Add more file patterns
on:
  pull_request:
    paths:
      - '**/*.yaml'
      - '**/*.py'
      - '**/requirements.txt'

# Change validation level
- name: Validate models
  run: |
    conduit validate "$MODEL"  # Basic validation
    # or
    conduit validate --strict "$MODEL"  # Strict validation

# Add custom checks
- name: Custom checks
  run: |
    # Check model size
    MODEL_SIZE=$(stat -f%z "$MODEL")
    if [ $MODEL_SIZE -gt 1000000 ]; then
      echo "Error: Model spec too large"
      exit 1
    fi
```

### Integration with PR Checks

Configure branch protection rules:

1. Go to Settings → Branches → Branch protection rules
2. Add rule for `main`
3. Enable "Require status checks to pass before merging"
4. Select "Validate Model Specifications"

Now PRs cannot be merged if validation fails!

---

## Publish Workflow

### Purpose

Automatically publish validated models to team registry when releasing.

### When It Runs

- When you create a GitHub release
- Manually via workflow_dispatch (optional)

### What It Does

1. **Validates** models (ensures they pass validation)
2. **Adds** to catalog
3. **Pushes** to configured registry
4. **Creates** summary of published models

### Configuring Registry

Set these secrets:

```
REGISTRY_URL=https://models.example.com
REGISTRY_TOKEN=your_token_here
```

Or use environment-specific registries:

```yaml
# In workflow file
- name: Configure registry
  run: |
    if [ "${{ github.event.release.prerelease }}" == "true" ]; then
      REGISTRY_URL="${{ secrets.STAGING_REGISTRY_URL }}"
      REGISTRY_TOKEN="${{ secrets.STAGING_REGISTRY_TOKEN }}"
    else
      REGISTRY_URL="${{ secrets.PROD_REGISTRY_URL }}"
      REGISTRY_TOKEN="${{ secrets.PROD_REGISTRY_TOKEN }}"
    fi

    conduit registry add target $REGISTRY_URL --token $REGISTRY_TOKEN
```

### Versioning

Use Git tags for versions:

```bash
# Create release with version tag
git tag v1.0.0
git push origin v1.0.0

# Create GitHub release
gh release create v1.0.0 --title "Release v1.0.0" --notes "Bug fixes and improvements"
```

The publish workflow automatically picks up the version from the tag.

---

## Deploy Workflow

### Purpose

Automatically deploy models to AWS SageMaker on release or manually.

### Trigger Options

**Option 1: Automatic on Release**
```yaml
on:
  release:
    types: [published]
```

**Option 2: Manual Trigger**
```yaml
on:
  workflow_dispatch:
    inputs:
      model_path:
        description: 'Path to model.yaml'
        required: true
```

**Option 3: Both**
```yaml
on:
  release:
    types: [published]
  workflow_dispatch:
    inputs:
      model_path:
        description: 'Path to model.yaml'
        required: true
```

### Manual Deployment

Trigger manually via GitHub UI:

1. Go to Actions tab
2. Select "Deploy Models" workflow
3. Click "Run workflow"
4. Enter parameters:
   - Model path: `models/my-model/model.yaml`
   - Instance type: `ml.g5.xlarge`
   - Endpoint name: `my-production-model`
5. Click "Run workflow"

Or via GitHub CLI:

```bash
gh workflow run deploy \
  -f model_path=model.yaml \
  -f instance_type=ml.g5.xlarge \
  -f endpoint_name=my-production-model
```

### Environment-Specific Deployments

Deploy to different environments:

```yaml
on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment'
        required: true
        type: choice
        options:
          - dev
          - staging
          - production

jobs:
  deploy:
    name: Deploy to ${{ github.event.inputs.environment }}
    runs-on: ubuntu-latest

    steps:
      # ... setup steps ...

      - name: Deploy model
        run: |
          ENV="${{ github.event.inputs.environment }}"

          case $ENV in
            dev)
              REGION="us-east-1"
              INSTANCE="ml.m5.large"
              ENDPOINT="model-dev"
              ;;
            staging)
              REGION="us-west-2"
              INSTANCE="ml.g5.xlarge"
              ENDPOINT="model-staging"
              ;;
            production)
              REGION="us-west-2"
              INSTANCE="ml.g5.2xlarge"
              ENDPOINT="model-production"
              ;;
          esac

          conduit deploy model.yaml \
            --platform sagemaker \
            --region $REGION \
            --instance-type $INSTANCE \
            --endpoint-name $ENDPOINT \
            --role ${{ secrets.SAGEMAKER_ROLE }}
```

---

## Custom Workflows

### Model Testing Workflow

Test models before merging:

```yaml
name: Test Models

on:
  pull_request:
    branches: [main]

jobs:
  test:
    name: Run Model Tests
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install pytest

      - name: Run unit tests
        run: |
          pytest tests/ -v

      - name: Test inference locally
        run: |
          python predict.py --test
```

### Multi-Model Workflow

Deploy multiple models in sequence:

```yaml
name: Deploy All Models

on:
  workflow_dispatch:

jobs:
  deploy:
    name: Deploy All Models
    runs-on: ubuntu-latest

    strategy:
      matrix:
        model: [model1, model2, model3]

    steps:
      - uses: actions/checkout@v4

      - name: Install Conduit
        run: |
          go install github.com/scttfrdmn/conduit/cmd/conduit@latest

      - name: Configure AWS
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ secrets.AWS_REGION }}

      - name: Deploy ${{ matrix.model }}
        run: |
          conduit deploy models/${{ matrix.model }}/model.yaml \
            --endpoint-name ${{ matrix.model }}-production
```

### Scheduled Validation

Validate models on a schedule:

```yaml
name: Scheduled Validation

on:
  schedule:
    - cron: '0 0 * * 0'  # Every Sunday at midnight

jobs:
  validate:
    name: Weekly Validation
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Install Conduit
        run: |
          go install github.com/scttfrdmn/conduit/cmd/conduit@latest

      - name: Validate all models
        run: |
          find . -name 'model.yaml' | while read MODEL; do
            conduit validate --strict "$MODEL"
          done

      - name: Create issue if failed
        if: failure()
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: 'Weekly validation failed',
              body: 'Automated validation found errors. Please review.'
            })
```

---

## GitHub Secrets

### Required Secrets

For AWS deployment:

```
AWS_ACCESS_KEY_ID          # AWS access key
AWS_SECRET_ACCESS_KEY      # AWS secret key
AWS_REGION                 # Default: us-east-1
SAGEMAKER_ROLE            # IAM role ARN
```

For registry publishing:

```
REGISTRY_URL              # Registry endpoint
REGISTRY_TOKEN            # Authentication token
```

### Setting Secrets

**Via GitHub UI**:
1. Repository → Settings → Secrets → Actions
2. Click "New repository secret"
3. Enter name and value
4. Click "Add secret"

**Via GitHub CLI**:
```bash
gh secret set AWS_ACCESS_KEY_ID --body "AKIA..."
gh secret set AWS_SECRET_ACCESS_KEY --body "..."
gh secret set AWS_REGION --body "us-west-2"
gh secret set SAGEMAKER_ROLE --body "arn:aws:iam::..."
```

### Environment Secrets

For different environments (dev/staging/prod):

1. Settings → Environments → New environment
2. Create environments: `dev`, `staging`, `production`
3. Add environment-specific secrets
4. Reference in workflows:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: production  # Uses production secrets

    steps:
      - name: Deploy
        run: conduit deploy model.yaml
```

---

## Best Practices

### 1. Always Validate on PRs

```yaml
# Required: Validate before merging
on:
  pull_request:
    branches: [main]

# Enable branch protection
# Settings → Branches → Add rule → Require status checks
```

### 2. Version Everything

```bash
# Use semantic versioning
git tag v1.0.0
git push origin v1.0.0

# Update model.yaml version to match
version: "1.0.0"
```

### 3. Separate Environments

```
dev → staging → production

# Use separate:
# - AWS accounts/regions
# - Registries
# - Endpoints
```

### 4. Test Before Deploying

```yaml
# Add tests to workflow
jobs:
  test:
    # ... run tests ...

  deploy:
    needs: test  # Only deploy if tests pass
    # ... deploy ...
```

### 5. Monitor Deployments

```yaml
# Add monitoring step
- name: Monitor endpoint
  run: |
    sleep 60  # Wait for endpoint to stabilize

    # Check endpoint health
    aws sagemaker describe-endpoint --endpoint-name my-model

    # Run smoke test
    python tests/smoke_test.py
```

### 6. Rollback Strategy

```yaml
# Keep previous version
- name: Deploy with blue-green
  run: |
    # Deploy new version
    conduit deploy model.yaml --endpoint-name model-v2

    # Test new version
    python tests/validate_endpoint.py model-v2

    # If success, switch traffic
    # If failure, keep model-v1
```

### 7. Secrets Management

- ✅ Use GitHub Secrets (never commit credentials)
- ✅ Rotate secrets regularly
- ✅ Use environment-specific secrets
- ✅ Use OIDC for AWS (instead of long-lived keys)

### 8. Notifications

```yaml
# Add Slack notification
- name: Notify on failure
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {
        "text": "Deployment failed for ${{ github.repository }}"
      }
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

---

## Troubleshooting

### Workflow Not Triggering

**Check**:
- Workflow file in `.github/workflows/`
- File is valid YAML
- Trigger conditions match (e.g., PR to `main` branch)
- GitHub Actions enabled (Settings → Actions)

### Validation Failing

**Check**:
- Run locally: `conduit validate --strict model.yaml`
- Check workflow logs for specific errors
- Verify all files referenced in model.yaml exist
- Check Python dependencies are installable

### Deployment Failing

**Check**:
- AWS credentials are correct
- IAM role has required permissions
- Docker is set up correctly in workflow
- Check CloudWatch logs for specific errors

### Secrets Not Working

**Check**:
- Secret names match exactly (case-sensitive)
- Secrets are set at repository level (not user level)
- For environment secrets, workflow references environment
- Re-create secrets if suspicious

---

## Advanced Topics

### Matrix Builds

Deploy to multiple regions:

```yaml
strategy:
  matrix:
    region: [us-east-1, us-west-2, eu-west-1]

steps:
  - name: Deploy to ${{ matrix.region }}
    run: |
      conduit deploy model.yaml \
        --region ${{ matrix.region }} \
        --endpoint-name model-${{ matrix.region }}
```

### Reusable Workflows

Create reusable workflow:

`.github/workflows/deploy-model.yml`:
```yaml
name: Reusable Deploy

on:
  workflow_call:
    inputs:
      model_path:
        required: true
        type: string
      environment:
        required: true
        type: string

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ inputs.environment }}

    steps:
      - uses: actions/checkout@v4
      # ... deployment steps ...
```

Use in another workflow:

```yaml
jobs:
  deploy-model1:
    uses: ./.github/workflows/deploy-model.yml
    with:
      model_path: models/model1/model.yaml
      environment: production
```

### Conditional Deployments

Only deploy if changes to specific paths:

```yaml
on:
  push:
    branches: [main]
    paths:
      - 'models/**'

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Check changed files
        id: changed
        uses: tj-actions/changed-files@v40

      - name: Deploy if model changed
        if: contains(steps.changed.outputs.modified_files, 'model.yaml')
        run: conduit deploy model.yaml
```

---

## Related Documentation

- [Getting Started Guide](getting-started.md) - Learn Conduit basics
- [Model Specification](model-spec.md) - model.yaml reference
- [Command Reference](commands.md) - All CLI commands
- [Deployment Guide](deployment.md) - AWS deployment details
- [Registry Guide](registry.md) - Team collaboration

---

## External Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Actions Marketplace](https://github.com/marketplace?type=actions)
- [Workflow Syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions)
- [AWS Actions](https://github.com/aws-actions)
