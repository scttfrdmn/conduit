package deployment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scttfrdmn/conduit/pkg/types"
)

// DockerfileGenerator generates Dockerfiles from model specifications
type DockerfileGenerator struct {
	model *types.Model
}

// NewDockerfileGenerator creates a new Dockerfile generator
func NewDockerfileGenerator(model *types.Model) *DockerfileGenerator {
	return &DockerfileGenerator{
		model: model,
	}
}

// Generate generates a Dockerfile for the model
func (g *DockerfileGenerator) Generate() (string, error) {
	var dockerfile strings.Builder

	// Select base image based on framework
	baseImage := g.getBaseImage()
	dockerfile.WriteString(fmt.Sprintf("FROM %s\n\n", baseImage))

	// Set working directory
	dockerfile.WriteString("WORKDIR /opt/ml/model\n\n")

	// Install system dependencies
	if g.model.Hardware.GPURequired {
		dockerfile.WriteString("# GPU support\n")
		dockerfile.WriteString("ENV NVIDIA_VISIBLE_DEVICES all\n")
		dockerfile.WriteString("ENV NVIDIA_DRIVER_CAPABILITIES compute,utility\n\n")
	}

	// Copy and install Python dependencies
	if g.model.Runtime.Dependencies != "" {
		dockerfile.WriteString("# Install Python dependencies\n")
		dockerfile.WriteString(fmt.Sprintf("COPY %s requirements.txt\n", g.model.Runtime.Dependencies))
		dockerfile.WriteString("RUN pip install --no-cache-dir -r requirements.txt\n\n")
	}

	// Copy inference code
	dockerfile.WriteString("# Copy inference code\n")
	entrypointDir := filepath.Dir(g.model.Inference.Entrypoint)
	if entrypointDir != "." && entrypointDir != "" {
		dockerfile.WriteString(fmt.Sprintf("COPY %s/ ./\n\n", entrypointDir))
	} else {
		dockerfile.WriteString("COPY . ./\n\n")
	}

	// Set environment variables
	dockerfile.WriteString("# Environment variables\n")
	dockerfile.WriteString("ENV PYTHONUNBUFFERED=1\n")
	dockerfile.WriteString(fmt.Sprintf("ENV MODEL_NAME=%s\n", g.model.Name))
	dockerfile.WriteString(fmt.Sprintf("ENV MODEL_VERSION=%s\n", g.model.Version))
	dockerfile.WriteString(fmt.Sprintf("ENV WEIGHTS_URI=%s\n", g.model.WeightsURI))
	dockerfile.WriteString("\n")

	// Set up entry point
	dockerfile.WriteString("# Entry point\n")
	entrypointBase := filepath.Base(g.model.Inference.Entrypoint)
	dockerfile.WriteString(fmt.Sprintf("ENV SAGEMAKER_PROGRAM=%s\n", entrypointBase))
	dockerfile.WriteString("\n")

	// Health check
	dockerfile.WriteString("# Health check\n")
	dockerfile.WriteString("HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \\\n")
	dockerfile.WriteString("  CMD python -c \"import sys; sys.exit(0)\"\n\n")

	// Expose port for SageMaker
	dockerfile.WriteString("# Expose port\n")
	dockerfile.WriteString("EXPOSE 8080\n\n")

	// Default command
	dockerfile.WriteString("# Command\n")
	dockerfile.WriteString("CMD [\"python\", \"-m\", \"sagemaker_inference\", \"serve\"]\n")

	return dockerfile.String(), nil
}

// getBaseImage returns the appropriate base image for the framework
func (g *DockerfileGenerator) getBaseImage() string {
	framework := strings.ToLower(g.model.Runtime.Framework)
	pythonVersion := g.model.Runtime.PythonVersion
	if pythonVersion == "" {
		pythonVersion = "3.9"
	}

	// Use official SageMaker images when possible
	if g.model.Hardware.GPURequired {
		switch framework {
		case "pytorch":
			return fmt.Sprintf("763104351884.dkr.ecr.us-east-1.amazonaws.com/pytorch-inference:%s-gpu-py%s", "2.0.0", strings.Replace(pythonVersion, ".", "", 1))
		case "tensorflow":
			return fmt.Sprintf("763104351884.dkr.ecr.us-east-1.amazonaws.com/tensorflow-inference:%s-gpu-py%s", "2.13.0", strings.Replace(pythonVersion, ".", "", 1))
		default:
			// Fall back to NVIDIA CUDA base image
			return "nvidia/cuda:11.8.0-cudnn8-runtime-ubuntu22.04"
		}
	}

	// CPU images
	switch framework {
	case "pytorch":
		return fmt.Sprintf("python:%s-slim", pythonVersion)
	case "tensorflow":
		return fmt.Sprintf("python:%s-slim", pythonVersion)
	case "jax":
		return fmt.Sprintf("python:%s-slim", pythonVersion)
	case "onnx":
		return fmt.Sprintf("python:%s-slim", pythonVersion)
	default:
		return fmt.Sprintf("python:%s-slim", pythonVersion)
	}
}

// WriteDockerfile writes the Dockerfile to disk
func (g *DockerfileGenerator) WriteDockerfile(outputDir string) error {
	content, err := g.Generate()
	if err != nil {
		return err
	}

	dockerfilePath := filepath.Join(outputDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil { //nolint:gosec // Standard file permissions
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	return nil
}

// GenerateDockerignore generates a .dockerignore file
func GenerateDockerignore(outputDir string) error {
	content := `# Ignore git and version control
.git
.gitignore
.gitattributes

# Ignore Python cache
__pycache__
*.pyc
*.pyo
*.pyd
.Python
*.so

# Ignore test files
tests/
test_*.py
*_test.py

# Ignore documentation
*.md
docs/

# Ignore IDE files
.vscode/
.idea/
*.swp
*.swo

# Ignore OS files
.DS_Store
Thumbs.db

# Ignore build artifacts
build/
dist/
*.egg-info/

# Ignore large files that shouldn't be in the image
*.tar.gz
*.zip
*.ckpt
*.pt
*.pth
*.h5
*.pb
weights/
checkpoints/
`

	dockerignorePath := filepath.Join(outputDir, ".dockerignore")
	if err := os.WriteFile(dockerignorePath, []byte(content), 0644); err != nil { //nolint:gosec // Standard file permissions
		return fmt.Errorf("failed to write .dockerignore: %w", err)
	}

	return nil
}
