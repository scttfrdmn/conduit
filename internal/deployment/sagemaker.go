package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
)

// SageMakerDeployer handles deployment to SageMaker Endpoints
type SageMakerDeployer struct {
	awsConfig aws.Config
	config    *Config
	client    *sagemaker.Client
}

// NewSageMakerDeployer creates a new SageMaker deployer
func NewSageMakerDeployer(awsConfig aws.Config, config *Config) *SageMakerDeployer {
	return &SageMakerDeployer{
		awsConfig: awsConfig,
		config:    config,
		client:    sagemaker.NewFromConfig(awsConfig),
	}
}

// Validate checks if the deployment configuration is valid for SageMaker
func (d *SageMakerDeployer) Validate(ctx context.Context) error {
	// Validate model configuration
	if d.config.Model.WeightsURI == "" {
		return fmt.Errorf("weights_uri is required for deployment")
	}

	if d.config.Model.Inference.Entrypoint == "" {
		return fmt.Errorf("inference entrypoint is required for deployment")
	}

	if d.config.Model.Inference.Handler == "" {
		return fmt.Errorf("inference handler is required for deployment")
	}

	// SageMaker requires a role
	if d.config.Role == "" {
		return fmt.Errorf("IAM role is required for SageMaker deployment")
	}

	return nil
}

// Deploy deploys the model to SageMaker
func (d *SageMakerDeployer) Deploy(ctx context.Context) (*DeploymentResult, error) {
	// Validate first
	if err := d.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if Docker is installed
	if err := CheckDockerInstalled(ctx); err != nil {
		return nil, err
	}

	// Step 1: Build and push Docker image
	fmt.Println("Step 1/4: Building and pushing Docker image...")
	imageURI, err := d.buildAndPushDockerImage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build/push image: %w", err)
	}
	fmt.Printf("✓ Image URI: %s\n\n", imageURI)

	// Step 2: Create SageMaker model
	fmt.Println("Step 2/4: Creating SageMaker model...")
	modelName, err := d.createModel(ctx, imageURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}
	fmt.Printf("✓ Model created: %s\n\n", modelName)

	// Step 3: Create endpoint configuration
	fmt.Println("Step 3/4: Creating endpoint configuration...")
	endpointConfigName, err := d.createEndpointConfig(ctx, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint config: %w", err)
	}
	fmt.Printf("✓ Endpoint config created: %s\n\n", endpointConfigName)

	// Step 4: Create and wait for endpoint
	fmt.Println("Step 4/4: Creating endpoint...")
	endpointARN, err := d.createEndpoint(ctx, endpointConfigName)
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}
	fmt.Printf("✓ Endpoint created: %s\n\n", d.config.EndpointName)

	// Return result
	return &DeploymentResult{
		EndpointName: d.config.EndpointName,
		EndpointARN:  endpointARN,
		Status:       "InService",
		URL:          fmt.Sprintf("https://runtime.sagemaker.%s.amazonaws.com/endpoints/%s/invocations", d.config.Region, d.config.EndpointName),
		Region:       d.config.Region,
		Platform:     PlatformSageMaker,
	}, nil
}

// buildAndPushDockerImage builds a Docker image and pushes it to ECR
func (d *SageMakerDeployer) buildAndPushDockerImage(ctx context.Context) (string, error) {
	// Create temporary directory for build context
	buildDir, err := os.MkdirTemp("", "conduit-build-*")
	if err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}
	defer os.RemoveAll(buildDir) //nolint:errcheck // Best effort cleanup

	// Generate Dockerfile
	generator := NewDockerfileGenerator(d.config.Model)
	if err := generator.WriteDockerfile(buildDir); err != nil {
		return "", fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Generate .dockerignore
	if err := GenerateDockerignore(buildDir); err != nil {
		return "", fmt.Errorf("failed to generate .dockerignore: %w", err)
	}

	// Copy inference code and dependencies to build context
	if err := d.copyInferenceFiles(buildDir); err != nil {
		return "", fmt.Errorf("failed to copy inference files: %w", err)
	}

	// Build and push to ECR
	ecrManager := NewECRManager(d.awsConfig, d.config.Region)
	repositoryName := fmt.Sprintf("conduit/%s", d.config.Model.Name)
	tag := d.config.Model.Version

	imageURI, err := ecrManager.BuildAndPushImage(ctx, buildDir, repositoryName, tag)
	if err != nil {
		return "", err
	}

	return imageURI, nil
}

// copyInferenceFiles copies inference code and dependencies to build context
func (d *SageMakerDeployer) copyInferenceFiles(buildDir string) error {
	// Copy dependencies file if specified
	if d.config.Model.Runtime.Dependencies != "" {
		srcPath := d.config.Model.Runtime.Dependencies
		dstPath := filepath.Join(buildDir, filepath.Base(srcPath))
		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy dependencies: %w", err)
		}
	}

	// Copy inference code
	entrypointDir := filepath.Dir(d.config.Model.Inference.Entrypoint)
	if entrypointDir == "." || entrypointDir == "" {
		entrypointDir = "."
	}

	// For simplicity, copy all Python files from the entrypoint directory
	return filepath.Walk(entrypointDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only copy Python files and the entrypoint
		if filepath.Ext(path) == ".py" || path == d.config.Model.Inference.Entrypoint {
			relPath, err := filepath.Rel(entrypointDir, path)
			if err != nil {
				return err
			}

			dstPath := filepath.Join(buildDir, relPath)
			if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil { //nolint:gosec // Standard directory permissions
				return err
			}

			return copyFile(path, dstPath)
		}

		return nil
	})
}

// createModel creates a SageMaker model
func (d *SageMakerDeployer) createModel(ctx context.Context, imageURI string) (string, error) {
	modelName := fmt.Sprintf("%s-%s", d.config.Model.Name, d.config.Model.Version)

	input := &sagemaker.CreateModelInput{
		ModelName: aws.String(modelName),
		PrimaryContainer: &types.ContainerDefinition{
			Image: aws.String(imageURI),
			Environment: map[string]string{
				"MODEL_NAME":    d.config.Model.Name,
				"MODEL_VERSION": d.config.Model.Version,
				"WEIGHTS_URI":   d.config.Model.WeightsURI,
			},
		},
		ExecutionRoleArn: aws.String(d.config.Role),
	}

	_, err := d.client.CreateModel(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create SageMaker model: %w", err)
	}

	return modelName, nil
}

// createEndpointConfig creates a SageMaker endpoint configuration
func (d *SageMakerDeployer) createEndpointConfig(ctx context.Context, modelName string) (string, error) {
	configName := fmt.Sprintf("%s-config", d.config.EndpointName)

	input := &sagemaker.CreateEndpointConfigInput{
		EndpointConfigName: aws.String(configName),
		ProductionVariants: []types.ProductionVariant{
			{
				VariantName:          aws.String("AllTraffic"),
				ModelName:            aws.String(modelName),
				InitialInstanceCount: aws.Int32(1),
				InstanceType:         types.ProductionVariantInstanceType(d.config.InstanceType),
				InitialVariantWeight: aws.Float32(1.0),
			},
		},
	}

	_, err := d.client.CreateEndpointConfig(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create endpoint config: %w", err)
	}

	return configName, nil
}

// createEndpoint creates a SageMaker endpoint
func (d *SageMakerDeployer) createEndpoint(ctx context.Context, endpointConfigName string) (string, error) {
	input := &sagemaker.CreateEndpointInput{
		EndpointName:       aws.String(d.config.EndpointName),
		EndpointConfigName: aws.String(endpointConfigName),
	}

	output, err := d.client.CreateEndpoint(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create endpoint: %w", err)
	}

	// Wait for endpoint to be in service
	fmt.Println("Waiting for endpoint to be in service (this may take several minutes)...")
	if err := d.waitForEndpoint(ctx); err != nil {
		return "", err
	}

	return *output.EndpointArn, nil
}

// waitForEndpoint waits for the endpoint to reach InService status
func (d *SageMakerDeployer) waitForEndpoint(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	timeout := time.After(20 * time.Minute)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for endpoint to be in service")

		case <-ticker.C:
			output, err := d.client.DescribeEndpoint(ctx, &sagemaker.DescribeEndpointInput{
				EndpointName: aws.String(d.config.EndpointName),
			})

			if err != nil {
				return fmt.Errorf("failed to describe endpoint: %w", err)
			}

			status := output.EndpointStatus
			fmt.Printf("Endpoint status: %s\n", status)

			switch status {
			case types.EndpointStatusInService:
				return nil
			case types.EndpointStatusFailed:
				return fmt.Errorf("endpoint creation failed: %s", aws.ToString(output.FailureReason))
			case types.EndpointStatusRollingBack:
				return fmt.Errorf("endpoint is rolling back")
			}
		}
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src) //nolint:gosec // User's own files
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644) //nolint:gosec // Standard file permissions
}
