# Getting Started with Conduit

This guide will walk you through everything you need to know to start using Conduit for managing and deploying your scientific ML models.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [First Steps](#first-steps)
- [Creating Your First Model](#creating-your-first-model)
- [Managing Your Catalog](#managing-your-catalog)
- [Searching for Models](#searching-for-models)
- [Working with Versions](#working-with-versions)
- [Sharing with Your Team](#sharing-with-your-team)
- [Deploying to AWS](#deploying-to-aws)
- [Next Steps](#next-steps)

---

## Prerequisites

Before you begin, make sure you have:

1. **Go 1.23 or later** (for building from source)
   ```bash
   go version
   # Should show: go version go1.23.x or later
   ```

2. **Git** (for cloning the repository)
   ```bash
   git --version
   ```

3. **Docker** (optional, for deployment features)
   ```bash
   docker --version
   ```

4. **AWS CLI** (optional, for AWS deployment)
   ```bash
   aws --version
   ```

---

## Installation

### Option 1: Build from Source

This is currently the recommended installation method:

```bash
# Clone the repository
git clone https://github.com/scttfrdmn/conduit.git
cd conduit

# Build the binary
go build -o conduit ./cmd/conduit

# Move to your PATH (optional but recommended)
sudo mv conduit /usr/local/bin/

# Verify installation
conduit --version
```

### Option 2: Using Go Install

If you have Go installed, you can install directly:

```bash
go install github.com/scttfrdmn/conduit/cmd/conduit@latest
```

### Verify Installation

Check that Conduit is properly installed:

```bash
conduit --version
# Should output version information

conduit --help
# Should display available commands
```

---

## First Steps

### Initialize Your Catalog

Conduit automatically creates a local catalog when you first use it. The catalog is stored at `~/.conduit/catalog.db`.

List your (currently empty) catalog:

```bash
conduit list
```

You should see a message indicating no models are found yet.

### Understanding the Catalog

Your local catalog is a SQLite database that stores:
- Model metadata (name, version, description, domain)
- Runtime configuration (framework, Python version)
- Hardware requirements (CPU/GPU, memory)
- Usage statistics (deployments, predictions, views)
- Tags and labels for organization
- Version history

---

## Creating Your First Model

### Step 1: Initialize a Model Project

Create a new directory for your model and initialize it:

```bash
# Create project directory
mkdir my-first-model
cd my-first-model

# Initialize model specification
conduit init
```

This creates a `model.yaml` file with a basic template.

### Step 2: Edit the Model Specification

Open `model.yaml` in your editor and customize it:

```yaml
name: "my-protein-model"
version: "1.0.0"
domain: "protein-science"
description: "My first protein structure prediction model"

# Author information
author: "Your Name"
email: "your.email@example.com"

# Runtime configuration
runtime:
  framework: "pytorch"
  python_version: "3.11"
  dependencies: "requirements.txt"

# Inference configuration
inference:
  entrypoint: "predict.py"
  handler: "predict"

# Model weights
weights_uri: "s3://my-bucket/models/my-protein-model/weights.pt"
weights_size_gb: 2.5
checksum_sha256: "abc123...def456"

# Hardware requirements
hardware:
  gpu_required: true
  recommended_instance: "ml.g5.xlarge"
  min_memory_gb: 8
  min_gpu_memory_gb: 16

# Tags for organization
tags:
  - protein-folding
  - ml
  - research

# License
license: "Apache-2.0"
github_repo: "github.com/yourorg/my-protein-model"
```

### Step 3: Create Inference Code

Create a simple inference script (`predict.py`):

```python
import torch

def predict(input_data):
    """
    Main prediction handler.

    Args:
        input_data: Input sequence or structure

    Returns:
        Prediction results
    """
    # Your inference logic here
    print(f"Processing input: {input_data}")

    # Load model weights from WEIGHTS_URI environment variable
    # (Conduit sets this automatically during deployment)

    return {"prediction": "example_output"}

if __name__ == "__main__":
    # Test your handler locally
    result = predict("MKTAYIAKQRQISFVKSHFSRQLEERLGLIEVQAPIL")
    print(result)
```

### Step 4: Create Dependencies File

Create `requirements.txt`:

```
torch==2.0.0
numpy>=1.24.0
biopython>=1.81
```

### Step 5: Validate Your Model

Before adding to the catalog, validate the specification:

```bash
conduit validate model.yaml
```

If there are any issues, the validator will tell you what needs to be fixed.

### Step 6: Add to Catalog

Add your model to the local catalog:

```bash
conduit add model.yaml
```

You should see a success message confirming the model was added.

---

## Managing Your Catalog

### List All Models

View all models in your catalog:

```bash
conduit list
```

Output:
```
NAME                VERSION   DOMAIN            TAGS
my-protein-model    1.0.0     protein-science   protein-folding, ml, research
```

### View Model Details

Get detailed information about a specific model:

```bash
conduit info my-protein-model
```

This shows:
- Full metadata
- Runtime configuration
- Hardware requirements
- Usage statistics
- All versions
- Tags and labels

### Delete a Model

Remove a model from your catalog:

```bash
conduit delete my-protein-model
```

**Note**: This only removes the catalog entry, not the actual model weights or code.

---

## Searching for Models

Conduit provides powerful search capabilities to help you find models.

### Basic Search

Search by name or description:

```bash
conduit search "protein"
```

### Fuzzy Search

Handle typos and variations:

```bash
# These all find models with "alphafold" in the name
conduit search "alphafld" --fuzzy
conduit search "alfa fold" --fuzzy
conduit search "alphaflod" --fuzzy
```

### Filter by Tags

Find models with specific tags:

```bash
conduit search --tags ml,protein-folding
```

### Filter by Multiple Criteria

Combine multiple filters:

```bash
conduit search "protein" \
  --tags ml \
  --license Apache-2.0 \
  --author "John Doe"
```

### Sort by Popularity

Find the most-used models:

```bash
conduit search --sort-by popular
```

### Adjust Fuzzy Matching Threshold

Control how similar matches need to be:

```bash
# More strict matching (default: 0.6)
conduit search "alphafld" --fuzzy --min-score 0.8

# More lenient matching
conduit search "alphafld" --fuzzy --min-score 0.5
```

---

## Working with Versions

Conduit supports full version management for your models.

### Create a New Version

When you update your model, create a new version:

```bash
# Update model.yaml with version: "1.1.0"
# Then create the new version
conduit version create my-protein-model 1.1.0
```

### List All Versions

See all versions of a model:

```bash
conduit version list my-protein-model
```

Output:
```
VERSION   CREATED              LATEST
1.0.0     2024-01-15 10:30:00
1.1.0     2024-01-20 14:45:00  ✓
```

### Set Latest Version

Mark a version as the latest:

```bash
conduit version set-latest my-protein-model 1.1.0
```

### Query Specific Versions

Get information about a specific version:

```bash
conduit info my-protein-model@1.0.0
```

---

## Sharing with Your Team

Conduit's registry system lets you share models with your team like Docker images.

### Configure a Registry

Add a remote registry:

```bash
# HTTP registry
conduit registry add myteam https://models.mycompany.com

# With authentication
conduit registry add myteam https://models.mycompany.com \
  --username your-username \
  --token your-api-token
```

### List Registries

View all configured registries:

```bash
conduit registry list
```

### Set Default Registry

Set a default registry for push/pull operations:

```bash
conduit registry set-default myteam
```

### Push a Model

Upload a model to the registry:

```bash
# Push latest version
conduit push my-protein-model

# Push specific version
conduit push my-protein-model@1.0.0

# Push to specific registry
conduit push my-protein-model --registry myteam
```

### Pull a Model

Download a model from the registry:

```bash
# Pull latest version
conduit pull my-protein-model

# Pull specific version
conduit pull my-protein-model@1.0.0

# Pull from specific registry
conduit pull my-protein-model --registry myteam
```

### Conflict Resolution

When pulling a model that already exists locally, choose a strategy:

```bash
# Skip if exists (default)
conduit pull my-protein-model --strategy skip

# Overwrite local version
conduit pull my-protein-model --strategy overwrite

# Merge metadata (keeps local changes)
conduit pull my-protein-model --strategy merge
```

### Export and Import

Backup or share models as JSON files:

```bash
# Export a model
conduit export my-protein-model -o backup.json

# Import a model
conduit import backup.json
```

---

## Deploying to AWS

Conduit can automatically deploy your models to AWS SageMaker endpoints.

### Prerequisites for Deployment

1. **AWS Account** with SageMaker permissions
2. **Docker** installed and running
3. **AWS CLI** configured with credentials:
   ```bash
   aws configure
   ```

4. **IAM Role** for SageMaker with permissions:
   - AmazonSageMakerFullAccess
   - AmazonEC2ContainerRegistryFullAccess
   - S3 access to your model weights

### Deploy to SageMaker

Deploy your model with a single command:

```bash
conduit deploy model.yaml --platform sagemaker
```

This automatically:
1. Generates a Dockerfile from your model specification
2. Builds the Docker image
3. Pushes the image to Amazon ECR
4. Creates a SageMaker model
5. Creates an endpoint configuration
6. Deploys and monitors the endpoint

### Custom Deployment Configuration

Specify custom settings:

```bash
conduit deploy model.yaml \
  --platform sagemaker \
  --region us-west-2 \
  --instance-type ml.g5.2xlarge \
  --endpoint-name my-custom-endpoint \
  --role arn:aws:iam::123456789012:role/SageMakerRole
```

### Monitor Deployment

The deploy command shows real-time progress:

```
Step 1/4: Building and pushing Docker image...
Building Docker image: conduit/my-protein-model:1.0.0
✓ Image built successfully
Authenticating with ECR...
Pushing image to ECR: 123456789012.dkr.ecr.us-west-2.amazonaws.com/conduit/my-protein-model:1.0.0
✓ Image pushed successfully
✓ Image URI: 123456789012.dkr.ecr.us-west-2.amazonaws.com/conduit/my-protein-model:1.0.0

Step 2/4: Creating SageMaker model...
✓ Model created: my-protein-model-1.0.0

Step 3/4: Creating endpoint configuration...
✓ Endpoint config created: my-protein-model-config

Step 4/4: Creating endpoint...
Waiting for endpoint to be in service (this may take several minutes)...
Endpoint status: Creating
Endpoint status: Creating
Endpoint status: InService
✓ Endpoint created: my-protein-model

Deployment successful!
Endpoint Name: my-protein-model
Endpoint URL: https://runtime.sagemaker.us-west-2.amazonaws.com/endpoints/my-protein-model/invocations
Region: us-west-2
```

### Test Your Endpoint

Once deployed, test your endpoint:

```bash
aws sagemaker-runtime invoke-endpoint \
  --endpoint-name my-protein-model \
  --body '{"sequence": "MKTAYIAKQRQISFVK"}' \
  --content-type application/json \
  output.json

cat output.json
```

---

## Next Steps

Now that you understand the basics, explore these advanced topics:

### Documentation

- [**Model Specification**](model-spec.md) - Complete `model.yaml` reference
- [**Command Reference**](commands.md) - Detailed documentation for all commands
- [**Deployment Guide**](deployment.md) - Advanced AWS deployment topics
- [**Registry Guide**](registry.md) - Setting up and using registries
- [**CI/CD Guide**](cicd.md) - Automating validation and deployment

### Advanced Features

1. **Strict Validation** - Enforce additional requirements:
   ```bash
   conduit validate --strict model.yaml
   ```

2. **CI/CD Workflows** - Generate GitHub Actions workflows:
   ```bash
   conduit workflow all
   ```

3. **Usage Statistics** - Track model popularity:
   ```bash
   conduit info my-protein-model
   # Shows deployment count, predictions, views
   ```

4. **Advanced Search** - Use multiple filters and relevance scoring:
   ```bash
   conduit search "protein" \
     --fuzzy \
     --tags ml,biology \
     --author "Smith" \
     --license MIT \
     --sort-by popular
   ```

### Best Practices

1. **Version your models** - Always create new versions for changes
2. **Use semantic versioning** - Follow MAJOR.MINOR.PATCH format
3. **Tag appropriately** - Use consistent tags across your organization
4. **Document thoroughly** - Include clear descriptions and examples
5. **Test locally first** - Validate before deploying to production
6. **Use registries** - Share models through registries, not file transfers
7. **Monitor deployments** - Track usage statistics to understand adoption

### Getting Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/scttfrdmn/conduit/issues)
- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/scttfrdmn/conduit/discussions)
- **Documentation**: Browse the [docs/](.) directory

### Example Models

Check out the [examples/](../examples/) directory for:
- Sample model.yaml files
- Example inference code
- GitHub Actions workflows
- Common patterns and templates

---

## Common Workflows

### Daily Development Workflow

```bash
# 1. Search for existing models
conduit search "protein folding"

# 2. Pull a model to examine
conduit pull alphafold2

# 3. Create your own model
conduit init my-new-model
# Edit model.yaml and predict.py

# 4. Validate and add to catalog
conduit validate model.yaml
conduit add model.yaml

# 5. Share with team
conduit push my-new-model
```

### Production Deployment Workflow

```bash
# 1. Create and validate model
conduit init production-model
# Edit model.yaml with production settings
conduit validate --strict model.yaml

# 2. Add to catalog
conduit add model.yaml

# 3. Deploy to AWS
conduit deploy model.yaml \
  --platform sagemaker \
  --instance-type ml.g5.2xlarge \
  --region us-east-1

# 4. Monitor endpoint (AWS Console or CLI)
aws sagemaker describe-endpoint \
  --endpoint-name production-model
```

### Team Collaboration Workflow

```bash
# Team member 1: Create and share
conduit add model.yaml
conduit push new-model --registry team

# Team member 2: Pull and test
conduit pull new-model --registry team
conduit info new-model
conduit validate model.yaml --strict

# Team member 2: Make improvements and share new version
# Edit model.yaml with version: "1.1.0"
conduit version create new-model 1.1.0
conduit push new-model@1.1.0 --registry team

# Team member 1: Update to latest
conduit pull new-model@1.1.0 --strategy overwrite
```

---

Congratulations! You now have a solid understanding of Conduit's core features. Explore the other documentation guides to dive deeper into specific topics.
