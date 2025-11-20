# Registry Guide

Complete guide to using Conduit registries for team collaboration and model sharing.

## Table of Contents

- [Overview](#overview)
- [Registry Types](#registry-types)
- [HTTP Registries](#http-registries)
- [Setting Up Registries](#setting-up-registries)
- [Authentication](#authentication)
- [Pushing Models](#pushing-models)
- [Pulling Models](#pulling-models)
- [Conflict Resolution](#conflict-resolution)
- [Search and Discovery](#search-and-discovery)
- [Best Practices](#best-practices)
- [Building a Registry Server](#building-a-registry-server)
- [S3 Registries](#s3-registries-future)
- [Git Registries](#git-registries-future)
- [Troubleshooting](#troubleshooting)

---

## Overview

Conduit registries enable team collaboration by providing centralized storage and distribution of ML models. Think of them like Docker registries, but for ML models.

**Key Features**:
- **Push/Pull Workflow** - Similar to Docker/Git
- **Version Management** - Track model versions across team
- **Conflict Resolution** - Handle simultaneous updates
- **Authentication** - Secure access control
- **Search** - Discover models across registries

**Workflow**:
```
Developer A                Registry              Developer B
    |                          |                       |
    |--- Push model v1.0 ----->|                       |
    |                          |<--- Pull model v1.0 --|
    |                          |                       |
    |                          |<--- Push model v1.1 --|
    |<--- Pull model v1.1 -----|                       |
```

---

## Registry Types

Conduit supports three types of registries:

### 1. HTTP Registries (Available Now)

Simple HTTP-based registries with REST API:
- Web servers with API endpoints
- Easiest to set up and use
- Good for small to medium teams
- Examples: Custom server, company intranet

### 2. S3 Registries (Coming Soon)

AWS S3-backed registries:
- Use S3 buckets as storage
- Serverless (no server to maintain)
- Cost-effective for large models
- Built-in versioning and lifecycle policies

### 3. Git Registries (Coming Soon)

GitHub/GitLab-based registries:
- Use Git releases for model storage
- Leverage existing Git infrastructure
- Full audit trail via Git history
- Good for open-source projects

---

## HTTP Registries

### Registry API Specification

HTTP registries must implement these endpoints:

```
POST   /models/{name}                    # Push model (latest version)
POST   /models/{name}/versions/{version} # Push specific version
GET    /models/{name}                    # Pull model (latest version)
GET    /models/{name}/versions/{version} # Pull specific version
GET    /search?q={query}                 # Search models
GET    /models                            # List all models
DELETE /models/{name}/versions/{version} # Delete version
```

### Request/Response Format

**Push Model** (`POST /models/{name}`):
- **Request Headers**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <token>` (if authentication enabled)
- **Request Body**: Model export JSON (from `conduit export`)
- **Response**: `200 OK` or `201 Created`

**Pull Model** (`GET /models/{name}`):
- **Request Headers**:
  - `Authorization: Bearer <token>` (if authentication enabled)
- **Response**: Model export JSON
- **Status Codes**:
  - `200 OK` - Model found
  - `404 Not Found` - Model doesn't exist
  - `401 Unauthorized` - Invalid credentials

**Search** (`GET /search?q={query}`):
- **Query Parameters**:
  - `q` - Search query
  - `tags` - Filter by tags (comma-separated)
  - `limit` - Max results (default: 20)
- **Response**: Array of search results

```json
[
  {
    "name": "alphafold2",
    "version": "2.3.2",
    "description": "Protein structure prediction",
    "tags": ["protein-folding", "deep-learning"]
  }
]
```

---

## Setting Up Registries

### Adding a Registry

```bash
# Add HTTP registry
conduit registry add myteam https://models.example.com

# With authentication
conduit registry add myteam https://models.example.com \
  --username john \
  --token abc123xyz

# Set as default
conduit registry add myteam https://models.example.com --set-default
```

### Listing Registries

```bash
conduit registry list
```

Output:
```
NAME      URL                           TYPE   DEFAULT
myteam    https://models.example.com   http   ✓
backup    https://backup.example.com   http
public    https://public-models.org    http
```

### Setting Default Registry

```bash
# Set default
conduit registry set-default myteam

# Now push/pull use this registry by default
conduit push my-model
conduit pull other-model
```

### Removing a Registry

```bash
conduit registry remove old-registry
```

---

## Authentication

### Token-Based Authentication

Most common authentication method:

```bash
# Add registry with bearer token
conduit registry add myteam https://models.example.com \
  --token your-api-token-here
```

Token is sent as `Authorization: Bearer <token>` header.

### Basic Authentication

Username/password authentication:

```bash
# Add registry with username and password
conduit registry add myteam https://models.example.com \
  --username john \
  --token mypassword
```

Credentials are sent as Basic Auth header.

### No Authentication

For public or internal registries:

```bash
# Add registry without authentication
conduit registry add public https://public-models.org
```

### Storing Credentials Securely

Credentials are stored in `~/.conduit/registries.json`:

```json
{
  "registries": [
    {
      "name": "myteam",
      "url": "https://models.example.com",
      "type": "http",
      "username": "john",
      "token": "encrypted_token_here",
      "default": true
    }
  ]
}
```

**Security Note**: Tokens are encrypted at rest. For additional security:
- Use environment variables:
  ```bash
  export CONDUIT_REGISTRY_TOKEN=your_token
  conduit push my-model
  ```
- Use IAM roles (for AWS-based registries)
- Rotate tokens regularly

---

## Pushing Models

### Basic Push

Push latest version to default registry:

```bash
conduit push my-model
```

### Push Specific Version

```bash
# Using --version flag
conduit push my-model --version 1.5.0

# Using version suffix
conduit push my-model@1.5.0
```

### Push to Specific Registry

```bash
conduit push my-model --registry backup
```

### Force Push (Overwrite)

```bash
# Overwrite if model already exists on registry
conduit push my-model --force
```

### Push All Versions

```bash
# Push all versions of a model
for version in $(conduit version list my-model --format json | jq -r '.[].version'); do
  conduit push my-model@$version
done
```

### Push Workflow

```
1. Export model from local catalog
2. Serialize to JSON
3. Send HTTP POST to registry
4. Registry stores model data
5. Return success/failure
```

**Example**:
```bash
conduit push alphafold2

# Output:
Pushing "alphafold2" (v2.3.2) to registry "myteam"...
✓ Export successful (3.2 KB)
✓ Upload successful
✓ Model pushed to https://models.example.com/models/alphafold2
```

---

## Pulling Models

### Basic Pull

Pull latest version from default registry:

```bash
conduit pull my-model
```

### Pull Specific Version

```bash
# Using --version flag
conduit pull my-model --version 1.5.0

# Using version suffix
conduit pull my-model@1.5.0
```

### Pull from Specific Registry

```bash
conduit pull my-model --registry public
```

### Pull Workflow

```
1. Send HTTP GET to registry
2. Download model JSON
3. Import to local catalog (using conflict strategy)
4. Return success/failure
```

**Example**:
```bash
conduit pull alphafold2

# Output:
Pulling "alphafold2" from registry "myteam"...
✓ Download successful (3.2 KB)
✓ Import successful
✓ Model "alphafold2" (v2.3.2) added to catalog
```

---

## Conflict Resolution

When pulling a model that already exists locally, choose a resolution strategy:

### Skip (Default)

Skip if model already exists:

```bash
conduit pull my-model --strategy skip
```

**Behavior**:
- If model exists locally: Do nothing, keep local version
- If model doesn't exist: Import from registry

**Use When**: You want to preserve local changes

### Overwrite

Replace local version with registry version:

```bash
conduit pull my-model --strategy overwrite
```

**Behavior**:
- Always replace local version with registry version
- Local changes are lost

**Use When**: You want registry to be source of truth

### Merge

Merge registry version with local version:

```bash
conduit pull my-model --strategy merge
```

**Behavior**:
- Import registry version as new version
- Keep local version
- Both versions coexist

**Use When**: You want to keep both versions

### Examples

**Scenario 1: Local changes, don't want to lose them**
```bash
# You modified a model locally
# Someone else pushed updates to registry
# You want to keep your local changes

conduit pull my-model --strategy skip
# Local version preserved, registry version ignored
```

**Scenario 2: Want latest from registry**
```bash
# You want to sync with team's latest version
# Don't care about local changes

conduit pull my-model --strategy overwrite
# Local version replaced with registry version
```

**Scenario 3: Want both versions**
```bash
# You want to compare local and registry versions
# Keep both for testing

conduit pull my-model --strategy merge
# Now you have both versions:
# - my-model v1.0.0 (local)
# - my-model v1.1.0 (from registry)
```

---

## Search and Discovery

### Search Registry

Search for models in a registry:

```bash
# Search in default registry
conduit search "protein" --registry myteam

# Search specific registry
conduit search "folding" --registry public
```

**Note**: Currently searches local catalog. Future versions will search remote registries directly.

### List Registry Contents

```bash
# Coming soon
conduit registry browse myteam
```

### Compare Local and Registry

```bash
# Coming soon
conduit diff my-model --registry myteam
```

---

## Best Practices

### 1. Registry Naming

Use clear, descriptive names:
```bash
# Good
conduit registry add production https://models.prod.example.com
conduit registry add staging https://models.staging.example.com
conduit registry add public https://public-models.org

# Avoid
conduit registry add reg1 https://...
conduit registry add temp https://...
```

### 2. Version Management

- ✅ Always push specific versions: `conduit push model@1.0.0`
- ✅ Use semantic versioning: `1.0.0`, `1.1.0`, `2.0.0`
- ✅ Tag releases appropriately
- ❌ Don't push without versioning

### 3. Access Control

- ✅ Use token authentication for private registries
- ✅ Rotate tokens regularly (quarterly)
- ✅ Use separate tokens for CI/CD and developers
- ✅ Revoke tokens when team members leave
- ❌ Don't commit tokens to Git

### 4. Backup

- ✅ Maintain backup registry: `conduit registry add backup ...`
- ✅ Periodically push to backup: `conduit push model --registry backup`
- ✅ Export critical models: `conduit export model -o backups/`

### 5. Team Workflow

**Model Owner**:
```bash
# 1. Develop model locally
conduit add model.yaml

# 2. Test thoroughly
conduit validate --strict model.yaml

# 3. Version and push
conduit version create my-model 1.0.0
conduit push my-model@1.0.0 --registry team
```

**Team Member**:
```bash
# 1. Pull latest
conduit pull my-model --registry team

# 2. Test locally
conduit info my-model

# 3. Deploy if needed
conduit deploy model.yaml
```

### 6. Environments

Separate registries for different environments:

```bash
# Development registry
conduit registry add dev https://models.dev.example.com

# Staging registry
conduit registry add staging https://models.staging.example.com

# Production registry
conduit registry add prod https://models.prod.example.com

# Workflow:
# 1. Push to dev: conduit push model --registry dev
# 2. Test in dev environment
# 3. Promote to staging: conduit pull model --registry dev && conduit push model --registry staging
# 4. Test in staging
# 5. Promote to prod: conduit pull model --registry staging && conduit push model --registry prod
```

---

## Building a Registry Server

### Simple Python Registry Server

Here's a basic HTTP registry server implementation:

```python
# registry_server.py
from flask import Flask, request, jsonify
import json
import os
from pathlib import Path

app = Flask(__name__)

# Storage directory
STORAGE_DIR = Path("./registry_storage")
STORAGE_DIR.mkdir(exist_ok=True)

# Authentication (simple token-based)
API_TOKEN = os.environ.get("REGISTRY_TOKEN", "changeme")

def check_auth():
    """Check bearer token authentication"""
    auth_header = request.headers.get("Authorization")
    if not auth_header:
        return False

    if not auth_header.startswith("Bearer "):
        return False

    token = auth_header[7:]  # Remove "Bearer " prefix
    return token == API_TOKEN

@app.route("/models/<model_name>", methods=["POST"])
def push_model(model_name):
    """Push a model to the registry"""
    if not check_auth():
        return jsonify({"error": "Unauthorized"}), 401

    # Get model data
    model_data = request.get_json()

    # Save to file
    model_dir = STORAGE_DIR / model_name
    model_dir.mkdir(exist_ok=True)

    version = model_data.get("version", "latest")
    file_path = model_dir / f"{version}.json"

    with open(file_path, "w") as f:
        json.dump(model_data, f, indent=2)

    # Also save as latest
    latest_path = model_dir / "latest.json"
    with open(latest_path, "w") as f:
        json.dump(model_data, f, indent=2)

    return jsonify({"message": "Model pushed successfully"}), 201

@app.route("/models/<model_name>", methods=["GET"])
def pull_model(model_name):
    """Pull a model from the registry"""
    if not check_auth():
        return jsonify({"error": "Unauthorized"}), 401

    # Get version (default to latest)
    version = request.args.get("version", "latest")

    # Load from file
    file_path = STORAGE_DIR / model_name / f"{version}.json"

    if not file_path.exists():
        return jsonify({"error": "Model not found"}), 404

    with open(file_path) as f:
        model_data = json.load(f)

    return jsonify(model_data), 200

@app.route("/models/<model_name>/versions/<version>", methods=["POST"])
def push_model_version(model_name, version):
    """Push a specific version"""
    if not check_auth():
        return jsonify({"error": "Unauthorized"}), 401

    model_data = request.get_json()

    model_dir = STORAGE_DIR / model_name
    model_dir.mkdir(exist_ok=True)

    file_path = model_dir / f"{version}.json"

    with open(file_path, "w") as f:
        json.dump(model_data, f, indent=2)

    return jsonify({"message": "Model version pushed successfully"}), 201

@app.route("/models/<model_name>/versions/<version>", methods=["GET"])
def pull_model_version(model_name, version):
    """Pull a specific version"""
    if not check_auth():
        return jsonify({"error": "Unauthorized"}), 401

    file_path = STORAGE_DIR / model_name / f"{version}.json"

    if not file_path.exists():
        return jsonify({"error": "Model version not found"}), 404

    with open(file_path) as f:
        model_data = json.load(f)

    return jsonify(model_data), 200

@app.route("/search", methods=["GET"])
def search_models():
    """Search for models"""
    if not check_auth():
        return jsonify({"error": "Unauthorized"}), 401

    query = request.args.get("q", "").lower()

    results = []

    # Search through all models
    for model_dir in STORAGE_DIR.iterdir():
        if not model_dir.is_dir():
            continue

        latest_file = model_dir / "latest.json"
        if not latest_file.exists():
            continue

        with open(latest_file) as f:
            model_data = json.load(f)

        # Simple search: check if query in name or description
        name = model_data.get("name", "").lower()
        description = model_data.get("description", "").lower()

        if query in name or query in description:
            results.append({
                "name": model_data.get("name"),
                "version": model_data.get("version"),
                "description": model_data.get("description"),
                "tags": model_data.get("tags", [])
            })

    return jsonify(results), 200

@app.route("/models", methods=["GET"])
def list_models():
    """List all models"""
    if not check_auth():
        return jsonify({"error": "Unauthorized"}), 401

    models = []

    for model_dir in STORAGE_DIR.iterdir():
        if not model_dir.is_dir():
            continue

        latest_file = model_dir / "latest.json"
        if not latest_file.exists():
            continue

        with open(latest_file) as f:
            model_data = json.load(f)

        models.append({
            "name": model_data.get("name"),
            "version": model_data.get("version"),
            "description": model_data.get("description")
        })

    return jsonify(models), 200

if __name__ == "__main__":
    print("Starting Conduit Registry Server")
    print(f"Storage directory: {STORAGE_DIR.absolute()}")
    print(f"API Token: {API_TOKEN}")
    app.run(host="0.0.0.0", port=8080)
```

### Running the Registry Server

```bash
# Install dependencies
pip install flask

# Set authentication token
export REGISTRY_TOKEN=your-secret-token-here

# Run server
python registry_server.py

# Server runs on http://localhost:8080
```

### Using the Registry

```bash
# Add registry to Conduit
conduit registry add local http://localhost:8080 \
  --token your-secret-token-here

# Push model
conduit push my-model --registry local

# Pull model
conduit pull my-model --registry local
```

### Production Deployment

For production, enhance the server:

1. **Use a production WSGI server**:
   ```bash
   pip install gunicorn
   gunicorn -w 4 -b 0.0.0.0:8080 registry_server:app
   ```

2. **Add HTTPS** (use nginx or traefik as reverse proxy)

3. **Add database** (PostgreSQL instead of file storage)

4. **Add user management** (multiple users with different permissions)

5. **Add rate limiting** (prevent abuse)

6. **Add logging and monitoring**

---

## S3 Registries (Future)

Coming in future release:

```bash
# Add S3 registry
conduit registry add backup s3://my-bucket/models --type s3

# Push to S3
conduit push my-model --registry backup

# Pull from S3
conduit pull my-model --registry backup
```

**Benefits**:
- Serverless (no server to maintain)
- Automatic versioning
- Lifecycle policies (auto-delete old versions)
- Cost-effective for large models
- High availability

---

## Git Registries (Future)

Coming in future release:

```bash
# Add Git registry
conduit registry add public github.com/org/models --type git

# Push to GitHub releases
conduit push my-model --registry public

# Pull from GitHub releases
conduit pull my-model --registry public
```

**Benefits**:
- Full audit trail via Git history
- GitHub's infrastructure
- Good for open-source projects
- Built-in access control (GitHub permissions)

---

## Troubleshooting

### Connection Refused

**Error**:
```
Error: Failed to connect to registry: connection refused
```

**Solutions**:
- Verify registry URL is correct
- Check registry server is running
- Check firewall rules
- Verify network connectivity

### Authentication Failed

**Error**:
```
Error: Registry returned error: Unauthorized (status: 401)
```

**Solutions**:
- Verify token is correct
- Check token hasn't expired
- Ensure token has required permissions
- Re-add registry with correct credentials:
  ```bash
  conduit registry remove myteam
  conduit registry add myteam https://... --token new-token
  ```

### Model Not Found

**Error**:
```
Error: Registry returned error: Model not found (status: 404)
```

**Solutions**:
- Verify model name is correct (case-sensitive)
- Check model exists on registry: `conduit search model-name --registry myteam`
- Try pulling without version: `conduit pull model-name`

### Push Failed

**Error**:
```
Error: Failed to push model to registry
```

**Solutions**:
- Check you have write permissions
- Verify model is valid: `conduit validate model.yaml`
- Try force push: `conduit push model --force`
- Check registry server logs

### Slow Push/Pull

**Symptoms**:
- Push or pull takes very long time

**Solutions**:
- Check network speed
- Model might be very large (check size with `conduit info`)
- Try compressing model data
- Use registry closer to your location

---

## Related Documentation

- [Getting Started Guide](getting-started.md) - Learn Conduit basics
- [Model Specification](model-spec.md) - model.yaml reference
- [Command Reference](commands.md) - All CLI commands
- [Deployment Guide](deployment.md) - AWS deployment
- [CI/CD Guide](cicd.md) - Automating workflows
