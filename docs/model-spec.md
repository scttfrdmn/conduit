# Model Specification Reference

This document provides a complete reference for the `model.yaml` specification format used by Conduit.

## Table of Contents

- [Overview](#overview)
- [Complete Example](#complete-example)
- [Required Fields](#required-fields)
- [Optional Fields](#optional-fields)
- [Field Reference](#field-reference)
- [Runtime Configuration](#runtime-configuration)
- [Inference Configuration](#inference-configuration)
- [Hardware Requirements](#hardware-requirements)
- [Benchmarks](#benchmarks)
- [Tags and Labels](#tags-and-labels)
- [Validation Rules](#validation-rules)
- [Best Practices](#best-practices)
- [Examples by Domain](#examples-by-domain)

---

## Overview

The `model.yaml` file is the central specification for your ML model in Conduit. It describes:

- **Metadata**: Name, version, description, author information
- **Runtime**: Framework, Python version, dependencies
- **Inference**: How to run predictions
- **Weights**: Where model artifacts are stored
- **Hardware**: CPU/GPU requirements and instance recommendations
- **Benchmarks**: Performance metrics (optional)
- **Tags**: Organization and discovery (optional)

---

## Complete Example

Here's a fully-documented example with all available fields:

```yaml
# ============================================================
# Basic Information (Required)
# ============================================================

name: "alphafold2"
version: "2.3.2"
domain: "protein-science"
description: "Protein structure prediction using deep learning"

# ============================================================
# Author Information (Optional but recommended)
# ============================================================

author: "DeepMind"
email: "alphafold@deepmind.com"
organization: "DeepMind Technologies"

# ============================================================
# Runtime Configuration (Required)
# ============================================================

runtime:
  framework: "jax"                    # pytorch, tensorflow, jax, onnx
  python_version: "3.10"              # Python version required
  dependencies: "requirements.txt"    # Path to requirements file
  system_packages:                    # Optional: apt packages
    - "libhdf5-dev"
    - "libffi-dev"

# ============================================================
# Inference Configuration (Required for deployment)
# ============================================================

inference:
  entrypoint: "inference/predict.py"  # Main inference script
  handler: "predict"                  # Function name in entrypoint
  input_schema:                       # Optional: input format
    type: "object"
    properties:
      sequence:
        type: "string"
        description: "Amino acid sequence"
  output_schema:                      # Optional: output format
    type: "object"
    properties:
      structure:
        type: "array"
        description: "Predicted 3D coordinates"

# ============================================================
# Model Artifacts (Required)
# ============================================================

weights_uri: "s3://deepmind-models/alphafold2/v2.3.2/params.npz"
weights_size_gb: 3.7
checksum_sha256: "a1b2c3d4e5f6789012345678901234567890123456789012345678901234"

# Optional: Additional artifacts
artifacts:
  - uri: "s3://deepmind-models/alphafold2/v2.3.2/config.json"
    type: "config"
    size_mb: 2.1
  - uri: "s3://deepmind-models/alphafold2/v2.3.2/database.tar.gz"
    type: "database"
    size_gb: 12.0

# ============================================================
# Hardware Requirements (Required for deployment)
# ============================================================

hardware:
  gpu_required: true
  recommended_instance: "ml.g5.2xlarge"
  min_memory_gb: 16
  min_gpu_memory_gb: 24
  min_storage_gb: 50                  # Optional: disk space
  gpu_count: 1                        # Optional: number of GPUs
  gpu_type: "nvidia-a10g"             # Optional: specific GPU

# ============================================================
# Performance Benchmarks (Optional)
# ============================================================

benchmarks:
  - dataset: "CASP14"
    metric: "GDT_TS"
    result: 92.4
    description: "Template-free modeling accuracy"
  - dataset: "CASP14"
    metric: "inference_time"
    result: 45.2
    unit: "seconds"
    description: "Average inference time per protein"

# ============================================================
# Tags and Labels (Optional but recommended)
# ============================================================

tags:
  - protein-folding
  - structure-prediction
  - deep-learning
  - research

labels:
  category: "structural-biology"
  maturity: "production"
  complexity: "high"

# ============================================================
# Licensing and Repository (Optional but recommended)
# ============================================================

license: "Apache-2.0"
github_repo: "github.com/deepmind/alphafold"
paper_url: "https://doi.org/10.1038/s41586-021-03819-2"
documentation_url: "https://github.com/deepmind/alphafold#readme"

# ============================================================
# Training Information (Optional)
# ============================================================

training:
  dataset: "PDB + BFD + Uniclust30 + UniRef90"
  num_samples: 170000
  training_time_hours: 720
  hardware_used: "128 TPUv3 cores"

# ============================================================
# Environment Variables (Optional)
# ============================================================

environment:
  ALPHAFOLD_DATA_DIR: "/opt/ml/model/data"
  MAX_TEMPLATE_DATE: "2020-05-14"
  NUM_PREDICTIONS: "1"

# ============================================================
# Health Check (Optional)
# ============================================================

health_check:
  endpoint: "/ping"
  timeout_seconds: 30
  interval_seconds: 60

# ============================================================
# Additional Metadata (Optional)
# ============================================================

created_at: "2023-11-15T10:30:00Z"
updated_at: "2024-01-20T14:45:00Z"
maintainer: "AlphaFold Team"
status: "stable"                      # alpha, beta, stable, deprecated
```

---

## Required Fields

These fields must be present in every `model.yaml`:

### name
- **Type**: String
- **Description**: Unique identifier for the model
- **Rules**:
  - Must be lowercase
  - Can contain letters, numbers, hyphens, underscores
  - Must start with a letter
  - Max length: 63 characters
- **Example**: `"alphafold2"`, `"esm-2-large"`, `"protein_mpnn"`

### version
- **Type**: String
- **Description**: Semantic version of the model
- **Rules**:
  - Should follow semantic versioning (MAJOR.MINOR.PATCH)
  - Examples: `"1.0.0"`, `"2.3.1"`, `"0.1.0-beta"`
- **Best Practice**: Use semantic versioning consistently

### domain
- **Type**: String
- **Description**: Scientific domain or field
- **Examples**:
  - `"protein-science"`
  - `"computer-vision"`
  - `"nlp"`
  - `"drug-discovery"`
  - `"materials-science"`

### description
- **Type**: String
- **Description**: Clear, concise description of what the model does
- **Best Practice**: 1-2 sentences explaining the model's purpose
- **Example**: `"Predicts 3D protein structure from amino acid sequence using deep learning"`

---

## Optional Fields

### author
- **Type**: String
- **Description**: Name of the model creator or team
- **Example**: `"DeepMind"`, `"John Doe"`

### email
- **Type**: String
- **Description**: Contact email for model maintainer
- **Example**: `"model-support@example.com"`

### organization
- **Type**: String
- **Description**: Organization or institution
- **Example**: `"Stanford University"`, `"MIT CSAIL"`

### license
- **Type**: String
- **Description**: Software license for the model
- **Common Values**: `"Apache-2.0"`, `"MIT"`, `"GPL-3.0"`, `"BSD-3-Clause"`
- **Important**: Should match the actual license file

### github_repo
- **Type**: String
- **Description**: GitHub repository URL (without https://)
- **Example**: `"github.com/organization/repo"`

### paper_url
- **Type**: String
- **Description**: Link to research paper or publication
- **Example**: `"https://doi.org/10.1038/s41586-021-03819-2"`

### documentation_url
- **Type**: String
- **Description**: Link to detailed documentation
- **Example**: `"https://docs.example.com/models/alphafold2"`

---

## Runtime Configuration

The `runtime` section specifies the execution environment.

```yaml
runtime:
  framework: "pytorch"           # Required
  python_version: "3.11"         # Required
  dependencies: "requirements.txt" # Required
  system_packages:               # Optional
    - "libgomp1"
    - "libhdf5-dev"
```

### framework
- **Type**: String
- **Required**: Yes
- **Allowed Values**: `"pytorch"`, `"tensorflow"`, `"jax"`, `"onnx"`, `"scikit-learn"`, `"xgboost"`
- **Description**: ML framework used by the model

### python_version
- **Type**: String
- **Required**: Yes
- **Format**: `"MAJOR.MINOR"` (e.g., `"3.11"`, `"3.10"`)
- **Supported Versions**: `"3.8"`, `"3.9"`, `"3.10"`, `"3.11"`, `"3.12"`

### dependencies
- **Type**: String
- **Required**: Yes (for deployment)
- **Description**: Path to pip requirements file
- **Example**: `"requirements.txt"`, `"deps/requirements.txt"`

### system_packages
- **Type**: Array of strings
- **Required**: No
- **Description**: System packages to install via apt-get
- **Example**: `["libgomp1", "libhdf5-dev", "ffmpeg"]`

---

## Inference Configuration

The `inference` section defines how predictions are made.

```yaml
inference:
  entrypoint: "predict.py"
  handler: "predict"
  input_schema:
    type: "object"
    properties:
      text:
        type: "string"
  output_schema:
    type: "object"
    properties:
      prediction:
        type: "number"
```

### entrypoint
- **Type**: String
- **Required**: Yes (for deployment)
- **Description**: Path to the main inference script
- **Example**: `"predict.py"`, `"src/inference.py"`

### handler
- **Type**: String
- **Required**: Yes (for deployment)
- **Description**: Function name to call for predictions
- **Example**: `"predict"`, `"run_inference"`
- **Expected Signature**:
  ```python
  def predict(input_data):
      # Process input_data
      return prediction_result
  ```

### input_schema
- **Type**: Object (JSON Schema format)
- **Required**: No
- **Description**: Expected input format
- **Use**: Documentation and validation

### output_schema
- **Type**: Object (JSON Schema format)
- **Required**: No
- **Description**: Expected output format
- **Use**: Documentation and validation

---

## Hardware Requirements

The `hardware` section specifies compute requirements.

```yaml
hardware:
  gpu_required: true
  recommended_instance: "ml.g5.2xlarge"
  min_memory_gb: 16
  min_gpu_memory_gb: 24
  min_storage_gb: 100
  gpu_count: 1
  gpu_type: "nvidia-a10g"
```

### gpu_required
- **Type**: Boolean
- **Required**: Yes (for deployment)
- **Description**: Whether GPU is required
- **Values**: `true` or `false`

### recommended_instance
- **Type**: String
- **Required**: Yes (for deployment)
- **Description**: Recommended AWS instance type
- **Examples**:
  - CPU: `"ml.m5.xlarge"`, `"ml.m5.2xlarge"`
  - GPU: `"ml.g5.xlarge"`, `"ml.g5.2xlarge"`, `"ml.p4d.24xlarge"`
- **Reference**: [AWS SageMaker Instance Types](https://aws.amazon.com/sagemaker/pricing/)

### min_memory_gb
- **Type**: Number
- **Required**: Yes (for deployment)
- **Description**: Minimum RAM in gigabytes
- **Example**: `8`, `16`, `32`, `64`

### min_gpu_memory_gb
- **Type**: Number
- **Required**: If `gpu_required: true`
- **Description**: Minimum GPU memory in gigabytes
- **Example**: `8`, `16`, `24`, `40`

### min_storage_gb
- **Type**: Number
- **Required**: No
- **Description**: Minimum disk space in gigabytes
- **Use**: For models with large databases or temporary files

### gpu_count
- **Type**: Integer
- **Required**: No
- **Default**: `1`
- **Description**: Number of GPUs required

### gpu_type
- **Type**: String
- **Required**: No
- **Description**: Specific GPU architecture
- **Examples**: `"nvidia-a10g"`, `"nvidia-v100"`, `"nvidia-a100"`

---

## Benchmarks

Document model performance with the `benchmarks` array.

```yaml
benchmarks:
  - dataset: "ImageNet-1K"
    metric: "top1_accuracy"
    result: 84.2
    description: "Top-1 accuracy on validation set"

  - dataset: "ImageNet-1K"
    metric: "inference_time"
    result: 12.5
    unit: "ms"
    description: "Average inference time per image"

  - dataset: "Custom Test Set"
    metric: "f1_score"
    result: 0.95
    conditions:
      batch_size: 32
      hardware: "V100"
```

### Benchmark Fields

- **dataset**: Name of the benchmark dataset
- **metric**: Metric name (e.g., `"accuracy"`, `"f1_score"`, `"inference_time"`)
- **result**: Numerical result
- **unit**: Optional unit (e.g., `"seconds"`, `"ms"`, `"%"`)
- **description**: Optional explanation
- **conditions**: Optional object with testing conditions

---

## Tags and Labels

### tags
- **Type**: Array of strings
- **Required**: No
- **Description**: Keywords for search and organization
- **Best Practices**:
  - Use lowercase
  - Use hyphens for multi-word tags
  - Be specific but not overly narrow
  - Common tags: `ml`, `deep-learning`, `computer-vision`, `nlp`, `research`, `production`

```yaml
tags:
  - protein-folding
  - structure-prediction
  - research
  - gpu-required
```

### labels
- **Type**: Object (key-value pairs)
- **Required**: No
- **Description**: Structured metadata
- **Use**: More specific categorization than tags

```yaml
labels:
  category: "structural-biology"
  maturity: "production"
  complexity: "high"
  deployment: "cloud-only"
```

---

## Validation Rules

Conduit validates `model.yaml` files at two levels:

### Basic Validation

Run with `conduit validate model.yaml`:

- Required fields present
- Field types correct
- Format conventions followed
- Version is valid semantic version
- Framework is supported

### Strict Validation

Run with `conduit validate --strict model.yaml`:

All basic validation plus:
- Weights URI is accessible
- Dependencies file exists
- Entrypoint file exists
- Handler function exists in entrypoint
- Checksum matches weights file
- Instance type is valid for AWS
- All referenced files exist

---

## Best Practices

### 1. Version Management

```yaml
# Good: Semantic versioning
version: "2.1.0"

# Avoid: Non-semantic versions
version: "latest"
version: "v2"
```

### 2. Descriptive Names

```yaml
# Good: Clear and descriptive
name: "protein-structure-predictor"
description: "Predicts 3D protein structures from amino acid sequences using transformer architecture"

# Avoid: Vague or generic
name: "model1"
description: "A model"
```

### 3. Complete Hardware Specs

```yaml
# Good: Complete requirements
hardware:
  gpu_required: true
  recommended_instance: "ml.g5.2xlarge"
  min_memory_gb: 16
  min_gpu_memory_gb: 24
  min_storage_gb: 50

# Avoid: Incomplete specs
hardware:
  gpu_required: true
```

### 4. Accessible Weights URIs

```yaml
# Good: S3 URI that the deployment environment can access
weights_uri: "s3://my-models/model-v1/weights.pt"

# Avoid: Local paths or inaccessible URLs
weights_uri: "/home/user/models/weights.pt"
weights_uri: "http://localhost:8000/weights.pt"
```

### 5. Proper Dependencies

```yaml
# Good: requirements.txt with pinned versions
# In requirements.txt:
# torch==2.0.0
# numpy==1.24.2
# transformers==4.28.0

# Avoid: Unpinned or missing versions
# In requirements.txt:
# torch
# numpy
# transformers
```

### 6. Comprehensive Tags

```yaml
# Good: Specific and discoverable
tags:
  - protein-folding
  - alphafold
  - structure-prediction
  - deep-learning
  - research

# Avoid: Too few or too generic
tags:
  - ai
  - model
```

---

## Examples by Domain

### Protein Science Model

```yaml
name: "esm2-large"
version: "1.0.0"
domain: "protein-science"
description: "Large-scale protein language model for sequence analysis"

runtime:
  framework: "pytorch"
  python_version: "3.10"
  dependencies: "requirements.txt"

inference:
  entrypoint: "predict.py"
  handler: "predict_structure"

weights_uri: "s3://fair-esm/esm2_t33_650M_UR50D.pt"
weights_size_gb: 2.4

hardware:
  gpu_required: true
  recommended_instance: "ml.g5.xlarge"
  min_memory_gb: 16
  min_gpu_memory_gb: 16

tags:
  - protein-language-model
  - sequence-analysis
  - meta-research
```

### Computer Vision Model

```yaml
name: "resnet50-imagenet"
version: "2.0.0"
domain: "computer-vision"
description: "ResNet-50 trained on ImageNet for image classification"

runtime:
  framework: "pytorch"
  python_version: "3.11"
  dependencies: "requirements.txt"

inference:
  entrypoint: "inference.py"
  handler: "classify"

weights_uri: "s3://torchvision-models/resnet50-19c8e357.pth"
weights_size_gb: 0.1

hardware:
  gpu_required: false
  recommended_instance: "ml.m5.xlarge"
  min_memory_gb: 4

benchmarks:
  - dataset: "ImageNet-1K"
    metric: "top1_accuracy"
    result: 76.15

tags:
  - computer-vision
  - image-classification
  - resnet
  - production-ready
```

### NLP Model

```yaml
name: "bert-base-uncased"
version: "1.0.0"
domain: "nlp"
description: "BERT base model for text classification and embeddings"

runtime:
  framework: "pytorch"
  python_version: "3.10"
  dependencies: "requirements.txt"

inference:
  entrypoint: "predict.py"
  handler: "encode"

weights_uri: "s3://huggingface-models/bert-base-uncased/pytorch_model.bin"
weights_size_gb: 0.4

hardware:
  gpu_required: false
  recommended_instance: "ml.m5.large"
  min_memory_gb: 8

tags:
  - nlp
  - bert
  - text-classification
  - embeddings
```

---

## Related Documentation

- [Getting Started Guide](getting-started.md) - Learn how to create your first model
- [Command Reference](commands.md) - All CLI commands
- [Deployment Guide](deployment.md) - Deploying models to AWS
- [Validation Guide](validation.md) - Understanding validation rules

---

## Schema Version

Current schema version: `1.0.0`

Future versions may add new fields while maintaining backward compatibility with existing `model.yaml` files.
