# Deployment Guide

Complete guide to deploying ML models with Conduit, focusing on AWS SageMaker deployment.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [AWS Setup](#aws-setup)
- [Basic Deployment](#basic-deployment)
- [Deployment Process](#deployment-process)
- [Configuration Options](#configuration-options)
- [Instance Types](#instance-types)
- [Dockerfile Generation](#dockerfile-generation)
- [Testing Endpoints](#testing-endpoints)
- [Monitoring](#monitoring)
- [Cost Management](#cost-management)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)
- [Advanced Topics](#advanced-topics)

---

## Overview

Conduit automates the deployment of ML models to AWS SageMaker endpoints. The deployment process includes:

1. **Dockerfile Generation** - Creates optimized Docker images from model specs
2. **Image Building** - Builds and pushes Docker images to Amazon ECR
3. **Model Creation** - Creates SageMaker model resources
4. **Endpoint Configuration** - Configures instance types and scaling
5. **Endpoint Deployment** - Deploys and monitors the live endpoint

```
model.yaml → Dockerfile → Docker Build → ECR Push →
SageMaker Model → Endpoint Config → Endpoint
```

---

## Prerequisites

### Local Requirements

1. **Docker** (required)
   ```bash
   docker --version
   # Docker version 20.10.0 or later
   ```

   Install: https://docs.docker.com/get-docker/

2. **AWS CLI** (required)
   ```bash
   aws --version
   # aws-cli/2.x or later
   ```

   Install: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html

3. **Conduit** (required)
   ```bash
   conduit --version
   ```

### AWS Requirements

1. **AWS Account** with:
   - SageMaker permissions
   - ECR (Elastic Container Registry) permissions
   - IAM role creation permissions
   - S3 access (for model weights)

2. **AWS Credentials** configured:
   ```bash
   aws configure
   # or use environment variables:
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_DEFAULT_REGION=us-east-1
   ```

3. **IAM Role** for SageMaker with permissions:
   - `AmazonSageMakerFullAccess`
   - `AmazonEC2ContainerRegistryFullAccess`
   - S3 read access to model weights

---

## AWS Setup

### Step 1: Create IAM Role

Create an IAM role for SageMaker to use:

**Using AWS Console**:
1. Go to IAM → Roles → Create Role
2. Select "AWS Service" → "SageMaker"
3. Attach policies:
   - `AmazonSageMakerFullAccess`
   - `AmazonEC2ContainerRegistryFullAccess`
4. Add inline policy for S3 access:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "s3:GetObject",
           "s3:ListBucket"
         ],
         "Resource": [
           "arn:aws:s3:::your-model-bucket",
           "arn:aws:s3:::your-model-bucket/*"
         ]
       }
     ]
   }
   ```
5. Name the role: `ConduitSageMakerRole`
6. Note the ARN: `arn:aws:iam::123456789012:role/ConduitSageMakerRole`

**Using AWS CLI**:
```bash
# Create trust policy
cat > trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "sagemaker.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

# Create role
aws iam create-role \
  --role-name ConduitSageMakerRole \
  --assume-role-policy-document file://trust-policy.json

# Attach policies
aws iam attach-role-policy \
  --role-name ConduitSageMakerRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonSageMakerFullAccess

aws iam attach-role-policy \
  --role-name ConduitSageMakerRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryFullAccess
```

### Step 2: Upload Model Weights to S3

```bash
# Create S3 bucket for models
aws s3 mb s3://my-ml-models

# Upload model weights
aws s3 cp model-weights.pt s3://my-ml-models/my-model/v1.0.0/weights.pt

# Update model.yaml with S3 URI
# weights_uri: "s3://my-ml-models/my-model/v1.0.0/weights.pt"
```

### Step 3: Verify Docker is Running

```bash
docker ps
# Should show running containers or empty list (not an error)
```

---

## Basic Deployment

### Simple Deployment

Deploy a model with default settings:

```bash
conduit deploy model.yaml --platform sagemaker
```

This uses:
- Default region (from AWS config)
- Instance type from `model.yaml`
- Endpoint name = model name
- Auto-detected IAM role

### Deploy with Options

```bash
conduit deploy model.yaml \
  --platform sagemaker \
  --region us-west-2 \
  --instance-type ml.g5.2xlarge \
  --endpoint-name my-production-model \
  --role arn:aws:iam::123456789012:role/ConduitSageMakerRole
```

---

## Deployment Process

### Step 1: Dockerfile Generation

Conduit generates a Dockerfile based on your model specification:

```dockerfile
# Example generated Dockerfile for PyTorch GPU model

FROM 763104351884.dkr.ecr.us-east-1.amazonaws.com/pytorch-inference:2.0.0-gpu-py310

WORKDIR /opt/ml/model

# GPU support
ENV NVIDIA_VISIBLE_DEVICES all
ENV NVIDIA_DRIVER_CAPABILITIES compute,utility

# Install Python dependencies
COPY requirements.txt requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Copy inference code
COPY . ./

# Environment variables
ENV PYTHONUNBUFFERED=1
ENV MODEL_NAME=my-model
ENV MODEL_VERSION=1.0.0
ENV WEIGHTS_URI=s3://my-models/weights.pt

# Entry point
ENV SAGEMAKER_PROGRAM=predict.py

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
  CMD python -c "import sys; sys.exit(0)"

# Expose port
EXPOSE 8080

# Command
CMD ["python", "-m", "sagemaker_inference", "serve"]
```

**Base Image Selection**:
- GPU models: AWS SageMaker GPU images with CUDA
- CPU models: Python slim images
- Framework-specific: PyTorch, TensorFlow, JAX optimized images

### Step 2: Docker Build and Push

```bash
# Conduit automatically:
1. Creates ECR repository: conduit/my-model
2. Builds Docker image locally
3. Authenticates with ECR
4. Tags image with model version
5. Pushes to ECR
```

**Output**:
```
Building Docker image: conduit/my-model:1.0.0
[Docker build logs...]
✓ Image built successfully

Authenticating with ECR...
✓ Authenticated

Tagging image: 123456789012.dkr.ecr.us-west-2.amazonaws.com/conduit/my-model:1.0.0
Pushing image to ECR...
[Push progress...]
✓ Image pushed successfully
```

### Step 3: SageMaker Model Creation

Creates a SageMaker model resource:

```python
{
  "ModelName": "my-model-1.0.0",
  "PrimaryContainer": {
    "Image": "123456789012.dkr.ecr.us-west-2.amazonaws.com/conduit/my-model:1.0.0",
    "Environment": {
      "MODEL_NAME": "my-model",
      "MODEL_VERSION": "1.0.0",
      "WEIGHTS_URI": "s3://my-models/weights.pt"
    }
  },
  "ExecutionRoleArn": "arn:aws:iam::123456789012:role/ConduitSageMakerRole"
}
```

### Step 4: Endpoint Configuration

Creates endpoint configuration with instance settings:

```python
{
  "EndpointConfigName": "my-model-config",
  "ProductionVariants": [
    {
      "VariantName": "AllTraffic",
      "ModelName": "my-model-1.0.0",
      "InitialInstanceCount": 1,
      "InstanceType": "ml.g5.2xlarge",
      "InitialVariantWeight": 1.0
    }
  ]
}
```

### Step 5: Endpoint Creation

Deploys the endpoint and waits for it to be InService:

```bash
Creating endpoint: my-model
Waiting for endpoint to be in service (this may take several minutes)...

Endpoint status: Creating
Endpoint status: Creating
Endpoint status: Creating
Endpoint status: InService

✓ Endpoint created successfully!
```

**Typical Deployment Times**:
- CPU instances: 3-5 minutes
- GPU instances: 5-8 minutes
- Large models: 8-12 minutes

---

## Configuration Options

### Instance Type Selection

Choose instance type based on model requirements:

**CPU Instances** (for CPU-only models):
```yaml
hardware:
  gpu_required: false
  recommended_instance: "ml.m5.xlarge"
  min_memory_gb: 8
```

Common CPU instances:
- `ml.m5.large` - 2 vCPU, 8 GB RAM
- `ml.m5.xlarge` - 4 vCPU, 16 GB RAM
- `ml.m5.2xlarge` - 8 vCPU, 32 GB RAM
- `ml.m5.4xlarge` - 16 vCPU, 64 GB RAM

**GPU Instances** (for GPU-accelerated models):
```yaml
hardware:
  gpu_required: true
  recommended_instance: "ml.g5.xlarge"
  min_memory_gb: 16
  min_gpu_memory_gb: 24
```

Common GPU instances:
- `ml.g5.xlarge` - 1× A10G (24 GB), 4 vCPU, 16 GB RAM
- `ml.g5.2xlarge` - 1× A10G (24 GB), 8 vCPU, 32 GB RAM
- `ml.g5.4xlarge` - 1× A10G (24 GB), 16 vCPU, 64 GB RAM
- `ml.g5.12xlarge` - 4× A10G (96 GB), 48 vCPU, 192 GB RAM
- `ml.p4d.24xlarge` - 8× A100 (320 GB), 96 vCPU, 1152 GB RAM

See [full pricing](https://aws.amazon.com/sagemaker/pricing/) for all instance types.

### Region Selection

```bash
# Deploy to specific region
conduit deploy model.yaml --region us-west-2

# Or set default region
export AWS_DEFAULT_REGION=us-west-2
conduit deploy model.yaml
```

**Recommended Regions** (for GPU availability):
- `us-east-1` (N. Virginia) - Most services available
- `us-west-2` (Oregon) - Good availability
- `eu-west-1` (Ireland) - Europe
- `ap-northeast-1` (Tokyo) - Asia Pacific

### Custom Endpoint Names

```bash
# Default: uses model name
conduit deploy model.yaml

# Custom name
conduit deploy model.yaml --endpoint-name production-v1

# Use model version in name
conduit deploy model.yaml --endpoint-name my-model-$(date +%Y%m%d)
```

### Environment Variables

Pass custom environment variables to the container:

```yaml
# In model.yaml
environment:
  MAX_BATCH_SIZE: "32"
  INFERENCE_TIMEOUT: "60"
  DEBUG_MODE: "false"
```

These are available in your inference code:
```python
import os

max_batch = int(os.environ.get("MAX_BATCH_SIZE", "16"))
timeout = int(os.environ.get("INFERENCE_TIMEOUT", "30"))
```

---

## Instance Types

### Choosing the Right Instance

**Decision Tree**:

1. **Does model need GPU?**
   - No → Use `ml.m5.*` (CPU)
   - Yes → Continue to step 2

2. **How much GPU memory needed?**
   - < 24 GB → `ml.g5.xlarge` or `ml.g5.2xlarge` (1× A10G, 24 GB)
   - 24-40 GB → `ml.g5.12xlarge` (4× A10G, 96 GB total)
   - > 40 GB → `ml.p4d.24xlarge` (8× A100, 320 GB total)

3. **What's your budget?**
   - Budget-friendly → `ml.g5.xlarge` ($1.41/hr)
   - Balanced → `ml.g5.2xlarge` ($1.69/hr)
   - High performance → `ml.g5.12xlarge` ($5.67/hr)
   - Maximum performance → `ml.p4d.24xlarge` ($32.77/hr)

### Instance Comparison

| Instance | vCPU | RAM | GPU | GPU RAM | Price/hr |
|----------|------|-----|-----|---------|----------|
| ml.m5.large | 2 | 8 GB | - | - | $0.12 |
| ml.m5.xlarge | 4 | 16 GB | - | - | $0.23 |
| ml.m5.2xlarge | 8 | 32 GB | - | - | $0.46 |
| ml.g5.xlarge | 4 | 16 GB | 1× A10G | 24 GB | $1.41 |
| ml.g5.2xlarge | 8 | 32 GB | 1× A10G | 24 GB | $1.69 |
| ml.g5.4xlarge | 16 | 64 GB | 1× A10G | 24 GB | $2.25 |
| ml.g5.12xlarge | 48 | 192 GB | 4× A10G | 96 GB | $5.67 |
| ml.p4d.24xlarge | 96 | 1152 GB | 8× A100 | 320 GB | $32.77 |

*Prices as of January 2024, may vary by region*

### Testing Instance Selection

Test with smaller instance first:

```bash
# Start with smaller instance
conduit deploy model.yaml --instance-type ml.g5.xlarge

# Test inference
# (see Testing Endpoints section)

# If OOM or slow, upgrade
# Delete endpoint first, then redeploy:
aws sagemaker delete-endpoint --endpoint-name my-model

conduit deploy model.yaml --instance-type ml.g5.2xlarge
```

---

## Dockerfile Generation

### Generated Dockerfile Structure

Conduit generates Dockerfiles with these sections:

1. **Base Image** - Framework and hardware-optimized
2. **Working Directory** - `/opt/ml/model`
3. **GPU Setup** - NVIDIA environment variables (if GPU required)
4. **Dependencies** - Python packages from requirements.txt
5. **Code Copy** - Inference scripts and handlers
6. **Environment** - Model metadata and configuration
7. **Health Check** - Container health monitoring
8. **Port Exposure** - SageMaker inference port (8080)
9. **Command** - SageMaker inference server

### Customizing the Dockerfile

**Option 1**: Modify model.yaml (recommended)

```yaml
runtime:
  framework: "pytorch"
  python_version: "3.11"
  dependencies: "requirements.txt"
  system_packages:
    - "libgomp1"
    - "libhdf5-dev"
```

**Option 2**: Provide custom Dockerfile

Create `Dockerfile` in your model directory:

```dockerfile
FROM pytorch/pytorch:2.0.0-cuda11.7-cudnn8-runtime

WORKDIR /opt/ml/model

# Custom system setup
RUN apt-get update && apt-get install -y \
    libgomp1 \
    libhdf5-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy inference code
COPY . .

# Environment
ENV PYTHONUNBUFFERED=1
ENV MODEL_NAME=my-model

# Health check
HEALTHCHECK --interval=30s CMD python -c "import sys; sys.exit(0)"

# Expose port
EXPOSE 8080

# Command
CMD ["python", "-m", "sagemaker_inference", "serve"]
```

Then deploy:
```bash
# Conduit will use existing Dockerfile if found
conduit deploy model.yaml
```

---

## Testing Endpoints

### Using AWS CLI

```bash
# Invoke endpoint
aws sagemaker-runtime invoke-endpoint \
  --endpoint-name my-model \
  --body '{"sequence": "MKTAYIAKQRQISFVK"}' \
  --content-type application/json \
  output.json

# View response
cat output.json
```

### Using Python

```python
import boto3
import json

# Create client
runtime = boto3.client('sagemaker-runtime', region_name='us-west-2')

# Prepare input
input_data = {
    "sequence": "MKTAYIAKQRQISFVKSHFSRQLEERLGLIEVQAPIL"
}

# Invoke endpoint
response = runtime.invoke_endpoint(
    EndpointName='my-model',
    ContentType='application/json',
    Body=json.dumps(input_data)
)

# Parse response
result = json.loads(response['Body'].read())
print(result)
```

### Using cURL

```bash
# Get invocation URL
ENDPOINT_URL="https://runtime.sagemaker.us-west-2.amazonaws.com/endpoints/my-model/invocations"

# Get temporary credentials
AWS_ACCESS_KEY=$(aws configure get aws_access_key_id)
AWS_SECRET_KEY=$(aws configure get aws_secret_access_key)

# Sign request and invoke (using awscurl or similar)
```

### Load Testing

Use Apache Bench or similar tools:

```bash
# Install Apache Bench
# sudo apt-get install apache2-utils

# Create test payload
echo '{"sequence": "MKTAYIAKQRQISFVK"}' > payload.json

# Load test (requires proper AWS signature)
# Use Python script for proper authentication

# Python load test script
python <<EOF
import boto3
import json
import time
from concurrent.futures import ThreadPoolExecutor

runtime = boto3.client('sagemaker-runtime')

def invoke():
    start = time.time()
    response = runtime.invoke_endpoint(
        EndpointName='my-model',
        Body=json.dumps({"sequence": "MKTAYIAKQRQISFVK"})
    )
    duration = time.time() - start
    return duration

# Run 100 requests with 10 concurrent threads
with ThreadPoolExecutor(max_workers=10) as executor:
    durations = list(executor.map(lambda _: invoke(), range(100)))

print(f"Average: {sum(durations)/len(durations):.2f}s")
print(f"Min: {min(durations):.2f}s")
print(f"Max: {max(durations):.2f}s")
EOF
```

---

## Monitoring

### CloudWatch Metrics

SageMaker automatically publishes metrics to CloudWatch:

**Invocation Metrics**:
- `Invocations` - Total invocations
- `Invocation4XXErrors` - Client errors
- `Invocation5XXErrors` - Server errors
- `ModelLatency` - Time spent in model inference
- `OverheadLatency` - Time spent in infrastructure

**Resource Metrics**:
- `CPUUtilization` - CPU usage percentage
- `MemoryUtilization` - Memory usage percentage
- `GPUUtilization` - GPU usage percentage (if GPU instance)
- `GPUMemoryUtilization` - GPU memory usage

### Viewing Metrics

**AWS Console**:
1. Go to CloudWatch → Metrics → SageMaker
2. Select "Endpoint Metrics"
3. Choose your endpoint name

**AWS CLI**:
```bash
# Get invocation count
aws cloudwatch get-metric-statistics \
  --namespace AWS/SageMaker \
  --metric-name Invocations \
  --dimensions Name=EndpointName,Value=my-model \
  --start-time 2024-01-20T00:00:00Z \
  --end-time 2024-01-20T23:59:59Z \
  --period 3600 \
  --statistics Sum

# Get average latency
aws cloudwatch get-metric-statistics \
  --namespace AWS/SageMaker \
  --metric-name ModelLatency \
  --dimensions Name=EndpointName,Value=my-model \
  --start-time 2024-01-20T00:00:00Z \
  --end-time 2024-01-20T23:59:59Z \
  --period 300 \
  --statistics Average
```

### Setting Up Alarms

```bash
# Create alarm for high error rate
aws cloudwatch put-metric-alarm \
  --alarm-name my-model-high-errors \
  --alarm-description "Alert on high error rate" \
  --metric-name Invocation5XXErrors \
  --namespace AWS/SageMaker \
  --statistic Sum \
  --period 300 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --dimensions Name=EndpointName,Value=my-model
```

### Logging

View endpoint logs in CloudWatch Logs:

```bash
# List log streams
aws logs describe-log-streams \
  --log-group-name /aws/sagemaker/Endpoints/my-model

# Get logs
aws logs tail /aws/sagemaker/Endpoints/my-model --follow
```

---

## Cost Management

### Understanding Costs

**Endpoint Costs** = Instance Cost × Uptime

Examples:
- `ml.g5.xlarge`: $1.41/hr × 730 hrs/month = **$1,029/month**
- `ml.m5.xlarge`: $0.23/hr × 730 hrs/month = **$168/month**

**Additional Costs**:
- ECR storage: ~$0.10/GB-month
- S3 storage: $0.023/GB-month
- Data transfer: $0.09/GB (out of region)

### Cost Optimization

**1. Use Appropriate Instance Sizes**

```bash
# Don't over-provision
# If ml.g5.xlarge is sufficient, don't use ml.g5.4xlarge

# Monitor utilization
aws cloudwatch get-metric-statistics \
  --namespace AWS/SageMaker \
  --metric-name GPUUtilization \
  --dimensions Name=EndpointName,Value=my-model \
  --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average
```

**2. Delete Unused Endpoints**

```bash
# List all endpoints
aws sagemaker list-endpoints

# Delete unused endpoints
aws sagemaker delete-endpoint --endpoint-name old-model
```

**3. Use Auto-scaling** (future feature)

```bash
# Coming soon: auto-scaling based on traffic
conduit deploy model.yaml --auto-scale --min-instances 1 --max-instances 5
```

**4. Use Spot Instances** (not available for SageMaker endpoints)

**5. Schedule Endpoints** (for dev/test)

```bash
# Stop endpoint (via Lambda, scheduled)
aws sagemaker delete-endpoint --endpoint-name dev-model

# Redeploy when needed
conduit deploy model.yaml --endpoint-name dev-model
```

### Cost Estimation

Before deploying:

```bash
# Estimate monthly cost
# Instance: ml.g5.xlarge = $1.41/hr
# Uptime: 730 hrs/month
# Cost: $1.41 × 730 = $1,029/month

# Add ECR storage: ~$1/month for typical model
# Add S3 storage: model size × $0.023/GB-month
# Total: ~$1,030-1,050/month
```

---

## Troubleshooting

### Common Issues

#### 1. Docker Not Running

**Error**:
```
Error: Cannot connect to Docker daemon
```

**Solution**:
```bash
# Start Docker
sudo systemctl start docker

# Or Docker Desktop on Mac/Windows

# Verify
docker ps
```

#### 2. AWS Credentials Not Configured

**Error**:
```
Error: Unable to locate credentials
```

**Solution**:
```bash
aws configure
# Enter access key, secret key, region
```

#### 3. IAM Permission Denied

**Error**:
```
Error: User is not authorized to perform: sagemaker:CreateModel
```

**Solution**:
- Verify IAM user has SageMaker permissions
- Attach `AmazonSageMakerFullAccess` policy
- Check IAM role exists and has correct policies

#### 4. ECR Push Failed

**Error**:
```
Error: failed to push image: access denied
```

**Solution**:
```bash
# Re-authenticate with ECR
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin \
  123456789012.dkr.ecr.us-west-2.amazonaws.com
```

#### 5. Endpoint Creation Failed

**Error**:
```
Error: Endpoint creation failed: Failed to download model
```

**Possible Causes**:
- Model weights URI is incorrect
- IAM role doesn't have S3 read permissions
- S3 bucket is in different region

**Solution**:
```bash
# Verify weights URI
aws s3 ls s3://my-bucket/model/weights.pt

# Check IAM role policies
aws iam get-role --role-name ConduitSageMakerRole
aws iam list-attached-role-policies --role-name ConduitSageMakerRole
```

#### 6. Endpoint InService but Errors on Invoke

**Error**:
```
Error: Model inference error
```

**Solution**:
```bash
# Check CloudWatch logs
aws logs tail /aws/sagemaker/Endpoints/my-model --follow

# Common issues:
# - Missing Python dependencies
# - Incorrect handler function name
# - Model weights not loading correctly
# - Memory issues (try larger instance)
```

#### 7. Out of Memory (OOM)

**Error** (in CloudWatch logs):
```
MemoryError: Unable to allocate tensor
```

**Solution**:
```bash
# Delete endpoint
aws sagemaker delete-endpoint --endpoint-name my-model

# Deploy with larger instance
conduit deploy model.yaml --instance-type ml.g5.2xlarge
```

### Debugging Tips

**1. Enable Verbose Logging**

```bash
conduit deploy --verbose model.yaml
```

**2. Check Docker Image Locally**

```bash
# Build image
docker build -t test-model .

# Run locally
docker run -p 8080:8080 test-model

# Test
curl -X POST http://localhost:8080/invocations \
  -H "Content-Type: application/json" \
  -d '{"sequence": "MKTAYIAKQRQISFVK"}'
```

**3. Validate Model Specification**

```bash
conduit validate --strict model.yaml
```

**4. Check SageMaker Logs**

```bash
# Real-time logs
aws logs tail /aws/sagemaker/Endpoints/my-model --follow

# Specific time range
aws logs filter-log-events \
  --log-group-name /aws/sagemaker/Endpoints/my-model \
  --start-time $(date -d '1 hour ago' +%s)000 \
  --end-time $(date +%s)000
```

---

## Best Practices

### 1. Model Preparation

- ✅ Test inference code locally before deploying
- ✅ Use strict validation: `conduit validate --strict`
- ✅ Pin dependency versions in requirements.txt
- ✅ Optimize model size (quantization, pruning)
- ✅ Upload weights to S3 before deployment

### 2. Deployment

- ✅ Start with smaller instances, scale up if needed
- ✅ Use meaningful endpoint names
- ✅ Tag resources for cost tracking
- ✅ Deploy to region closest to users
- ✅ Test thoroughly before production

### 3. Production

- ✅ Set up CloudWatch alarms
- ✅ Monitor costs regularly
- ✅ Implement health checks
- ✅ Use multiple availability zones
- ✅ Version your endpoints (e.g., `model-v1`, `model-v2`)
- ✅ Blue/green deployments for updates

### 4. Security

- ✅ Use IAM roles with least privilege
- ✅ Enable VPC for endpoints (enterprise)
- ✅ Encrypt data at rest and in transit
- ✅ Rotate AWS credentials regularly
- ✅ Use private S3 buckets for weights

### 5. Maintenance

- ✅ Delete unused endpoints
- ✅ Clean up old ECR images
- ✅ Update dependencies regularly
- ✅ Keep Conduit updated
- ✅ Document endpoint configurations

---

## Advanced Topics

### Multi-Model Endpoints

Deploy multiple models to a single endpoint (future feature):

```bash
# Coming soon
conduit deploy --multi-model \
  model1.yaml \
  model2.yaml \
  model3.yaml
```

### A/B Testing

Deploy multiple variants for A/B testing (future feature):

```bash
# Coming soon
conduit deploy model.yaml \
  --variant-a model-v1:50 \
  --variant-b model-v2:50
```

### Auto-scaling

Configure automatic scaling based on traffic (future feature):

```bash
# Coming soon
conduit deploy model.yaml \
  --auto-scale \
  --min-instances 1 \
  --max-instances 10 \
  --target-invocations 1000
```

### Custom Inference Container

Use completely custom Docker containers:

```bash
# Build custom container
docker build -t my-custom-inference .

# Push to ECR
# Deploy using AWS SageMaker directly or via Conduit
```

### Batch Transform Jobs

For batch inference (not real-time):

```bash
# Coming soon
conduit batch-transform model.yaml \
  --input s3://bucket/input/ \
  --output s3://bucket/output/
```

---

## Related Documentation

- [Getting Started Guide](getting-started.md) - Learn Conduit basics
- [Model Specification](model-spec.md) - model.yaml reference
- [Command Reference](commands.md) - All CLI commands
- [Registry Guide](registry.md) - Sharing models
- [CI/CD Guide](cicd.md) - Automating deployments

---

## External Resources

- [AWS SageMaker Documentation](https://docs.aws.amazon.com/sagemaker/)
- [SageMaker Pricing](https://aws.amazon.com/sagemaker/pricing/)
- [SageMaker Instance Types](https://aws.amazon.com/sagemaker/pricing/instance-types/)
- [Docker Documentation](https://docs.docker.com/)
- [AWS CLI Reference](https://docs.aws.amazon.com/cli/)
