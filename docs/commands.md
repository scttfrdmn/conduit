# Command Reference

Complete reference for all Conduit CLI commands.

## Table of Contents

- [Global Flags](#global-flags)
- [Catalog Management](#catalog-management)
- [Model Operations](#model-operations)
- [Version Management](#version-management)
- [Search](#search)
- [Export and Import](#export-and-import)
- [Registry Operations](#registry-operations)
- [Push and Pull](#push-and-pull)
- [Deployment](#deployment)
- [CI/CD](#cicd)
- [Utility Commands](#utility-commands)

---

## Global Flags

These flags work with any command:

```bash
--help, -h       Show help for a command
--version        Show version information
--verbose, -v    Enable verbose output
--quiet, -q      Suppress non-error output
```

**Examples**:
```bash
conduit --version
conduit search --help
conduit deploy --verbose model.yaml
```

---

## Catalog Management

### conduit add

Add a model to the local catalog.

**Usage**:
```bash
conduit add <model.yaml> [flags]
```

**Flags**:
```
--validate-strict    Run strict validation before adding
--update            Update if model already exists
--force             Skip confirmation prompts
```

**Examples**:
```bash
# Add a model
conduit add model.yaml

# Add with strict validation
conduit add model.yaml --validate-strict

# Update existing model
conduit add model.yaml --update

# Skip confirmation
conduit add model.yaml --force
```

**Output**:
```
✓ Model "my-model" (v1.0.0) added to catalog
```

**Exit Codes**:
- `0`: Success
- `1`: Validation failed
- `2`: Model already exists (without --update flag)

---

### conduit list

List all models in the catalog.

**Usage**:
```bash
conduit list [flags]
```

**Flags**:
```
--format string     Output format: table, json, yaml (default: table)
--tags strings      Filter by tags (comma-separated)
--domain string     Filter by domain
--sort-by string    Sort by: name, version, created, updated (default: name)
--limit int         Limit number of results (default: 0 = no limit)
```

**Examples**:
```bash
# List all models
conduit list

# List as JSON
conduit list --format json

# Filter by tags
conduit list --tags ml,protein-folding

# Filter by domain
conduit list --domain protein-science

# Sort by most recently updated
conduit list --sort-by updated

# Limit results
conduit list --limit 10
```

**Output (table format)**:
```
NAME                VERSION   DOMAIN            TAGS
alphafold2          2.3.2     protein-science   protein-folding, deep-learning
esm2-large          1.0.0     protein-science   protein-language-model
resnet50            2.0.0     computer-vision   image-classification
```

**Output (json format)**:
```json
[
  {
    "name": "alphafold2",
    "version": "2.3.2",
    "domain": "protein-science",
    "tags": ["protein-folding", "deep-learning"]
  }
]
```

---

### conduit info

Show detailed information about a model.

**Usage**:
```bash
conduit info <model-name> [flags]
```

**Flags**:
```
--version string    Show specific version (default: latest)
--format string     Output format: text, json, yaml (default: text)
```

**Examples**:
```bash
# Show latest version
conduit info alphafold2

# Show specific version
conduit info alphafold2 --version 2.3.1

# Output as JSON
conduit info alphafold2 --format json

# Using version suffix
conduit info alphafold2@2.3.1
```

**Output**:
```
Name:           alphafold2
Version:        2.3.2 (latest)
Domain:         protein-science
Description:    Protein structure prediction using deep learning
Author:         DeepMind
License:        Apache-2.0

Runtime:
  Framework:    jax
  Python:       3.10
  Dependencies: requirements.txt

Hardware:
  GPU Required: Yes
  Instance:     ml.g5.2xlarge
  Memory:       16 GB
  GPU Memory:   24 GB

Weights:
  URI:          s3://deepmind-models/alphafold2/v2.3.2/params.npz
  Size:         3.7 GB
  Checksum:     a1b2c3d4...

Usage Statistics:
  Total Deployments:  42
  Total Predictions:  1,523
  View Count:         89
  Last Deployed:      2024-01-15 10:30:00
  Last Viewed:        2024-01-20 14:45:00

Tags:
  protein-folding, structure-prediction, deep-learning

Versions:
  2.3.2 (latest) - 2024-01-20
  2.3.1          - 2024-01-10
  2.3.0          - 2023-12-15
```

---

### conduit delete

Remove a model from the catalog.

**Usage**:
```bash
conduit delete <model-name> [flags]
```

**Flags**:
```
--version string    Delete specific version only
--force            Skip confirmation prompt
--all-versions     Delete all versions
```

**Examples**:
```bash
# Delete latest version (with confirmation)
conduit delete my-model

# Delete specific version
conduit delete my-model --version 1.0.0

# Delete without confirmation
conduit delete my-model --force

# Delete all versions
conduit delete my-model --all-versions
```

**Output**:
```
Are you sure you want to delete "my-model" (v1.0.0)? [y/N]: y
✓ Model "my-model" (v1.0.0) deleted from catalog
```

**Note**: This only removes catalog entries, not actual model artifacts or weights.

---

## Model Operations

### conduit init

Initialize a new model project with template files.

**Usage**:
```bash
conduit init [path] [flags]
```

**Flags**:
```
--name string       Model name
--domain string     Model domain
--framework string  ML framework (pytorch, tensorflow, jax, onnx)
--template string   Template to use (basic, advanced, custom)
```

**Examples**:
```bash
# Initialize in current directory
conduit init

# Initialize in specific directory
conduit init my-new-model

# Specify model details
conduit init --name protein-predictor --domain protein-science --framework pytorch

# Use advanced template
conduit init --template advanced
```

**Output**:
```
✓ Created model.yaml
✓ Created predict.py
✓ Created requirements.txt
✓ Created README.md

Model project initialized at: ./my-new-model

Next steps:
1. Edit model.yaml with your model details
2. Implement predict() function in predict.py
3. Add dependencies to requirements.txt
4. Run: conduit validate model.yaml
5. Run: conduit add model.yaml
```

**Generated Files**:
- `model.yaml` - Model specification template
- `predict.py` - Inference code template
- `requirements.txt` - Empty dependencies file
- `README.md` - Basic documentation template

---

### conduit validate

Validate a model specification file.

**Usage**:
```bash
conduit validate <model.yaml> [flags]
```

**Flags**:
```
--strict           Enable strict validation
--quiet            Only show errors
--format string    Output format: text, json (default: text)
```

**Validation Levels**:

**Basic Validation**:
- Required fields present
- Field types correct
- Format conventions followed
- Version is semantic version
- Framework is supported

**Strict Validation** (with `--strict` flag):
- All basic validation
- Weights URI is accessible
- Dependencies file exists
- Entrypoint file exists
- Handler function exists
- Checksum matches (if provided)
- Instance type is valid

**Examples**:
```bash
# Basic validation
conduit validate model.yaml

# Strict validation
conduit validate --strict model.yaml

# Quiet mode (only errors)
conduit validate --quiet model.yaml

# JSON output
conduit validate --format json model.yaml
```

**Output (success)**:
```
✓ model.yaml is valid

Validation Summary:
  ✓ Required fields present
  ✓ Field types correct
  ✓ Version format valid
  ✓ Framework supported
  ✓ Hardware requirements complete
```

**Output (errors)**:
```
✗ model.yaml has 3 errors:

  ✗ Field 'weights_uri' is required
  ✗ Field 'version' must be semantic version (got: 'latest')
  ✗ Field 'hardware.gpu_required' is required for deployment
```

---

### conduit publish

Publish a model (placeholder for future functionality).

**Usage**:
```bash
conduit publish <model.yaml> [flags]
```

**Note**: Currently a placeholder. Future versions will support publishing to public registries or marketplaces.

---

## Version Management

### conduit version create

Create a new version of a model.

**Usage**:
```bash
conduit version create <model-name> <version> [flags]
```

**Flags**:
```
--from string      Copy from existing version (default: latest)
--set-latest       Set as latest version
--notes string     Version notes or changelog
```

**Examples**:
```bash
# Create new version
conduit version create my-model 2.0.0

# Create and set as latest
conduit version create my-model 2.0.0 --set-latest

# Copy from specific version
conduit version create my-model 2.1.0 --from 2.0.0

# Add version notes
conduit version create my-model 2.0.0 --notes "Fixed inference bug"
```

**Output**:
```
✓ Created version 2.0.0 for model "my-model"
✓ Set as latest version
```

---

### conduit version list

List all versions of a model.

**Usage**:
```bash
conduit version list <model-name> [flags]
```

**Flags**:
```
--format string    Output format: table, json, yaml (default: table)
--limit int        Limit number of results
```

**Examples**:
```bash
# List all versions
conduit version list my-model

# JSON output
conduit version list my-model --format json

# Limit to 5 most recent
conduit version list my-model --limit 5
```

**Output**:
```
VERSION   CREATED              LATEST   NOTES
2.1.0     2024-01-25 10:00:00  ✓       Performance improvements
2.0.0     2024-01-20 14:45:00          Fixed inference bug
1.5.0     2024-01-15 09:30:00          Added new features
1.0.0     2024-01-10 12:00:00          Initial release
```

---

### conduit version set-latest

Mark a version as the latest.

**Usage**:
```bash
conduit version set-latest <model-name> <version>
```

**Examples**:
```bash
# Set latest version
conduit version set-latest my-model 2.1.0

# Rollback to previous version
conduit version set-latest my-model 1.5.0
```

**Output**:
```
✓ Set version 2.1.0 as latest for model "my-model"
```

---

## Search

### conduit search

Search for models in the catalog.

**Usage**:
```bash
conduit search [query] [flags]
```

**Flags**:
```
--fuzzy                  Enable fuzzy matching
--min-score float        Minimum similarity score for fuzzy matching (0.0-1.0, default: 0.6)
--tags strings           Filter by tags (comma-separated)
--domain string          Filter by domain
--license string         Filter by license
--author string          Filter by author
--framework string       Filter by framework
--gpu-required bool      Filter by GPU requirement
--created-after string   Filter by creation date (YYYY-MM-DD)
--created-before string  Filter by creation date (YYYY-MM-DD)
--sort-by string         Sort by: relevance, name, created, updated, popular (default: relevance)
--limit int              Limit number of results (default: 20)
--format string          Output format: table, json, yaml (default: table)
```

**Search Behavior**:
- Without query: Lists all models (respecting filters)
- With query: Searches name, description, tags, author
- With `--fuzzy`: Handles typos and variations using Levenshtein distance

**Examples**:
```bash
# Basic search
conduit search "protein"

# Fuzzy search (handles typos)
conduit search "alphafld" --fuzzy

# Strict fuzzy matching
conduit search "alphafld" --fuzzy --min-score 0.8

# Filter by tags
conduit search --tags ml,protein-folding

# Filter by multiple criteria
conduit search "protein" \
  --tags ml \
  --framework pytorch \
  --gpu-required true \
  --license Apache-2.0

# Filter by date range
conduit search --created-after 2024-01-01

# Sort by popularity
conduit search --sort-by popular

# Limit results
conduit search "protein" --limit 5

# JSON output
conduit search "protein" --format json
```

**Output**:
```
Found 3 models matching "protein":

NAME                VERSION   DOMAIN            RELEVANCE   TAGS
alphafold2          2.3.2     protein-science   0.95       protein-folding, deep-learning
esm2-large          1.0.0     protein-science   0.87       protein-language-model, analysis
protein-mpnn        1.0.1     protein-science   0.92       protein-design, structure
```

**Fuzzy Matching Examples**:
```bash
# All these find "alphafold2":
conduit search "alphafld" --fuzzy
conduit search "alpha fold" --fuzzy
conduit search "alfafold" --fuzzy
```

---

## Export and Import

### conduit export

Export a model to a JSON file.

**Usage**:
```bash
conduit export <model-name> [flags]
```

**Flags**:
```
--version string      Export specific version (default: latest)
--output, -o string   Output file path (default: <model-name>.json)
--include-all-versions Export all versions
--pretty              Pretty-print JSON
```

**Examples**:
```bash
# Export latest version
conduit export my-model

# Export to specific file
conduit export my-model -o backup/model.json

# Export specific version
conduit export my-model --version 1.5.0

# Export all versions
conduit export my-model --include-all-versions

# Pretty-print
conduit export my-model --pretty
```

**Output**:
```
✓ Exported "my-model" (v1.0.0) to my-model.json
  Size: 2.3 KB
```

**Export Format**:
```json
{
  "name": "my-model",
  "version": "1.0.0",
  "exported_at": "2024-01-20T14:45:00Z",
  "metadata": { ... },
  "runtime": { ... },
  "hardware": { ... }
}
```

---

### conduit import

Import a model from a JSON file.

**Usage**:
```bash
conduit import <file> [flags]
```

**Flags**:
```
--strategy string   Conflict resolution: skip, overwrite, merge (default: skip)
--force            Skip confirmation prompts
--validate-strict  Run strict validation after import
```

**Strategies**:
- `skip`: Skip if model already exists (default)
- `overwrite`: Replace existing model
- `merge`: Merge with existing model (keeps local changes)

**Examples**:
```bash
# Import with default strategy (skip)
conduit import model.json

# Overwrite existing model
conduit import model.json --strategy overwrite

# Merge with existing
conduit import model.json --strategy merge

# Skip confirmation
conduit import model.json --force

# Import and validate
conduit import model.json --validate-strict
```

**Output**:
```
✓ Imported "my-model" (v1.0.0) from model.json
  Strategy: skip
  Status: Added (new model)
```

---

## Registry Operations

### conduit registry add

Add a remote registry.

**Usage**:
```bash
conduit registry add <name> <url> [flags]
```

**Flags**:
```
--type string       Registry type: http, s3, git (default: http)
--username string   Username for authentication
--token string      API token or password
--set-default      Set as default registry
```

**Examples**:
```bash
# Add HTTP registry
conduit registry add myteam https://models.mycompany.com

# Add with authentication
conduit registry add myteam https://models.mycompany.com \
  --username john \
  --token abc123

# Add and set as default
conduit registry add myteam https://models.mycompany.com --set-default

# Add S3 registry (future feature)
conduit registry add backup s3://my-bucket/models --type s3

# Add Git registry (future feature)
conduit registry add public github.com/org/models --type git
```

**Output**:
```
✓ Added registry "myteam"
  URL: https://models.mycompany.com
  Type: http
  Default: Yes
```

---

### conduit registry list

List all configured registries.

**Usage**:
```bash
conduit registry list [flags]
```

**Flags**:
```
--format string    Output format: table, json, yaml (default: table)
```

**Examples**:
```bash
# List registries
conduit registry list

# JSON output
conduit registry list --format json
```

**Output**:
```
NAME      URL                              TYPE   DEFAULT
myteam    https://models.mycompany.com    http   ✓
backup    https://backup.example.com      http
public    https://public-models.org       http
```

---

### conduit registry remove

Remove a registry.

**Usage**:
```bash
conduit registry remove <name> [flags]
```

**Flags**:
```
--force    Skip confirmation prompt
```

**Examples**:
```bash
# Remove registry (with confirmation)
conduit registry remove old-registry

# Skip confirmation
conduit registry remove old-registry --force
```

**Output**:
```
Are you sure you want to remove registry "old-registry"? [y/N]: y
✓ Removed registry "old-registry"
```

---

### conduit registry set-default

Set the default registry for push/pull operations.

**Usage**:
```bash
conduit registry set-default <name>
```

**Examples**:
```bash
# Set default registry
conduit registry set-default myteam
```

**Output**:
```
✓ Set "myteam" as default registry
```

---

## Push and Pull

### conduit push

Push a model to a remote registry.

**Usage**:
```bash
conduit push <model-name> [flags]
```

**Flags**:
```
--version string     Push specific version (default: latest)
--registry string    Target registry (default: default registry)
--force             Overwrite if exists on registry
```

**Examples**:
```bash
# Push to default registry
conduit push my-model

# Push specific version
conduit push my-model --version 1.5.0

# Push to specific registry
conduit push my-model --registry backup

# Force overwrite
conduit push my-model --force

# Using version suffix
conduit push my-model@1.5.0
```

**Output**:
```
Pushing "my-model" (v1.0.0) to registry "myteam"...
✓ Export successful (2.3 KB)
✓ Upload successful
✓ Model pushed to https://models.mycompany.com/models/my-model
```

---

### conduit pull

Pull a model from a remote registry.

**Usage**:
```bash
conduit pull <model-name> [flags]
```

**Flags**:
```
--version string     Pull specific version (default: latest)
--registry string    Source registry (default: default registry)
--strategy string    Conflict resolution: skip, overwrite, merge (default: skip)
--force             Skip confirmation prompts
```

**Examples**:
```bash
# Pull from default registry
conduit pull my-model

# Pull specific version
conduit pull my-model --version 1.5.0

# Pull from specific registry
conduit pull my-model --registry backup

# Overwrite local version
conduit pull my-model --strategy overwrite

# Using version suffix
conduit pull my-model@1.5.0
```

**Output**:
```
Pulling "my-model" from registry "myteam"...
✓ Download successful (2.3 KB)
✓ Import successful
✓ Model "my-model" (v1.0.0) added to catalog
```

---

## Deployment

### conduit deploy

Deploy a model to a cloud platform.

**Usage**:
```bash
conduit deploy <model.yaml> [flags]
```

**Flags**:
```
--platform string        Deployment platform: sagemaker (default: sagemaker)
--region string          AWS region (default: us-east-1)
--instance-type string   Instance type (default: from model.yaml)
--endpoint-name string   Endpoint name (default: model-name)
--role string           IAM role ARN (default: from AWS config)
--wait                  Wait for endpoint to be in service (default: true)
--no-wait               Don't wait for endpoint
--dry-run               Validate without deploying
```

**Examples**:
```bash
# Deploy with defaults
conduit deploy model.yaml

# Deploy to specific region
conduit deploy model.yaml --region us-west-2

# Custom instance type
conduit deploy model.yaml --instance-type ml.g5.4xlarge

# Custom endpoint name
conduit deploy model.yaml --endpoint-name production-model

# Specify IAM role
conduit deploy model.yaml --role arn:aws:iam::123456789012:role/SageMakerRole

# Don't wait for endpoint
conduit deploy model.yaml --no-wait

# Dry run (validate only)
conduit deploy model.yaml --dry-run
```

**Output**:
```
Step 1/4: Building and pushing Docker image...
Building Docker image: conduit/my-model:1.0.0
✓ Image built successfully
Authenticating with ECR...
✓ Image pushed successfully
✓ Image URI: 123456789012.dkr.ecr.us-west-2.amazonaws.com/conduit/my-model:1.0.0

Step 2/4: Creating SageMaker model...
✓ Model created: my-model-1.0.0

Step 3/4: Creating endpoint configuration...
✓ Endpoint config created: my-model-config

Step 4/4: Creating endpoint...
Waiting for endpoint to be in service (this may take several minutes)...
Endpoint status: Creating
Endpoint status: InService
✓ Endpoint created: my-model

Deployment successful!
  Endpoint Name: my-model
  Endpoint URL:  https://runtime.sagemaker.us-west-2.amazonaws.com/endpoints/my-model/invocations
  Region:        us-west-2
  Status:        InService
```

**Prerequisites**:
- Docker installed and running
- AWS CLI configured with credentials
- IAM role with SageMaker and ECR permissions

---

## CI/CD

### conduit workflow

Generate CI/CD workflow files.

**Usage**:
```bash
conduit workflow <type> [flags]
```

**Types**:
- `validate` - Workflow for model validation
- `publish` - Workflow for publishing on release
- `deploy` - Workflow for deployment
- `all` - Generate all workflows

**Flags**:
```
--output-dir string    Output directory (default: .github/workflows)
--force               Overwrite existing workflows
```

**Examples**:
```bash
# Generate validation workflow
conduit workflow validate

# Generate all workflows
conduit workflow all

# Custom output directory
conduit workflow all --output-dir .gitlab-ci

# Overwrite existing
conduit workflow all --force
```

**Output**:
```
✓ Created .github/workflows/conduit-validate.yml
✓ Created .github/workflows/conduit-publish.yml
✓ Created .github/workflows/conduit-deploy.yml

Workflows generated successfully!

Next steps:
1. Review and customize the generated workflows
2. Commit to your repository
3. GitHub Actions will run automatically on push/PR
```

**Generated Workflows**:

1. **conduit-validate.yml** - Runs on every PR
   - Validates model.yaml
   - Runs basic and strict validation
   - Reports errors as PR comments

2. **conduit-publish.yml** - Runs on release
   - Validates model
   - Adds/updates in catalog
   - Pushes to registry

3. **conduit-deploy.yml** - Manual or on release
   - Builds Docker image
   - Deploys to SageMaker
   - Reports endpoint URL

---

## Utility Commands

### conduit version

Show Conduit version information.

**Usage**:
```bash
conduit version
conduit --version
```

**Output**:
```
Conduit v0.2.0
Build: 83d8243
Go: go1.23.1
Platform: darwin/arm64
```

---

### conduit help

Show help information.

**Usage**:
```bash
conduit help
conduit help <command>
conduit <command> --help
```

**Examples**:
```bash
# General help
conduit help

# Command-specific help
conduit help search
conduit search --help
```

---

## Exit Codes

Conduit uses standard exit codes:

- `0` - Success
- `1` - General error (validation failed, file not found, etc.)
- `2` - Conflict (model already exists, registry conflict, etc.)
- `3` - Permission denied
- `4` - Network error
- `5` - AWS error
- `6` - Docker error

**Example Usage in Scripts**:
```bash
#!/bin/bash

conduit validate model.yaml
if [ $? -ne 0 ]; then
  echo "Validation failed"
  exit 1
fi

conduit deploy model.yaml
if [ $? -ne 0 ]; then
  echo "Deployment failed"
  exit 1
fi

echo "Success!"
```

---

## Environment Variables

Conduit recognizes these environment variables:

```bash
CONDUIT_HOME           # Catalog directory (default: ~/.conduit)
CONDUIT_REGISTRY       # Default registry name
AWS_REGION            # AWS region for deployment
AWS_PROFILE           # AWS profile to use
DOCKER_HOST           # Docker daemon address
```

**Example**:
```bash
# Use custom catalog location
export CONDUIT_HOME=/opt/conduit
conduit list

# Set default registry
export CONDUIT_REGISTRY=myteam
conduit push my-model

# Use specific AWS profile
export AWS_PROFILE=production
conduit deploy model.yaml
```

---

## Configuration File

Future versions will support a configuration file at `~/.conduit/config.yaml`:

```yaml
default_registry: myteam
default_region: us-west-2
default_instance_type: ml.g5.xlarge
validate_strict_by_default: true

registries:
  myteam:
    url: https://models.mycompany.com
    username: john
    token: ${REGISTRY_TOKEN}
```

---

## Related Documentation

- [Getting Started Guide](getting-started.md) - Learn Conduit basics
- [Model Specification](model-spec.md) - model.yaml reference
- [Deployment Guide](deployment.md) - AWS deployment details
- [Registry Guide](registry.md) - Registry setup and usage
- [CI/CD Guide](cicd.md) - Automation workflows
