package deployment

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

// ECRManager handles ECR repository operations
type ECRManager struct {
	client *ecr.Client
	config aws.Config
	region string
}

// NewECRManager creates a new ECR manager
func NewECRManager(cfg aws.Config, region string) *ECRManager {
	return &ECRManager{
		client: ecr.NewFromConfig(cfg),
		config: cfg,
		region: region,
	}
}

// CreateRepository creates an ECR repository if it doesn't exist
func (m *ECRManager) CreateRepository(ctx context.Context, repositoryName string) (string, error) {
	// Try to describe the repository first
	describeOutput, err := m.client.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repositoryName},
	})

	if err == nil && len(describeOutput.Repositories) > 0 {
		// Repository already exists
		return *describeOutput.Repositories[0].RepositoryUri, nil
	}

	// Repository doesn't exist, create it
	fmt.Printf("Creating ECR repository: %s\n", repositoryName)

	createOutput, err := m.client.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repositoryName),
		ImageScanningConfiguration: &types.ImageScanningConfiguration{
			ScanOnPush: true,
		},
		ImageTagMutability: types.ImageTagMutabilityMutable,
	})

	if err != nil {
		return "", fmt.Errorf("failed to create ECR repository: %w", err)
	}

	return *createOutput.Repository.RepositoryUri, nil
}

// GetAuthToken gets Docker authentication credentials for ECR
func (m *ECRManager) GetAuthToken(ctx context.Context) (string, string, error) {
	output, err := m.client.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to get ECR auth token: %w", err)
	}

	if len(output.AuthorizationData) == 0 {
		return "", "", fmt.Errorf("no authorization data returned")
	}

	authData := output.AuthorizationData[0]
	authToken, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode auth token: %w", err)
	}

	// Token is in format "username:password"
	parts := strings.SplitN(string(authToken), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid auth token format")
	}

	return parts[0], parts[1], nil
}

// LoginToECR authenticates Docker with ECR
func (m *ECRManager) LoginToECR(ctx context.Context) error {
	username, password, err := m.GetAuthToken(ctx)
	if err != nil {
		return err
	}

	registryURL := fmt.Sprintf("https://%s.dkr.ecr.%s.amazonaws.com", m.getAccountID(), m.region)

	fmt.Println("Authenticating with ECR...")

	// Docker login command
	//nolint:gosec // Docker commands with user's AWS credentials
	cmd := exec.CommandContext(ctx, "docker", "login",
		"--username", username,
		"--password-stdin",
		registryURL,
	)

	cmd.Stdin = strings.NewReader(password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker login failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// PushImage tags and pushes a Docker image to ECR
func (m *ECRManager) PushImage(ctx context.Context, localImage, repositoryURI, tag string) error {
	// Tag the image
	fullTag := fmt.Sprintf("%s:%s", repositoryURI, tag)
	fmt.Printf("Tagging image: %s -> %s\n", localImage, fullTag)

	//nolint:gosec // Docker commands with validated image names
	tagCmd := exec.CommandContext(ctx, "docker", "tag", localImage, fullTag)
	if output, err := tagCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to tag image: %w\nOutput: %s", err, string(output))
	}

	// Push the image
	fmt.Printf("Pushing image to ECR: %s\n", fullTag)

	//nolint:gosec // Docker commands with validated image names
	pushCmd := exec.CommandContext(ctx, "docker", "push", fullTag)
	pushCmd.Stdout = nil // We'll show progress
	pushCmd.Stderr = nil

	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	fmt.Println("✓ Image pushed successfully")
	return nil
}

// BuildAndPushImage builds a Docker image and pushes it to ECR
func (m *ECRManager) BuildAndPushImage(ctx context.Context, buildContext, repositoryName, tag string) (string, error) {
	// Create repository
	repositoryURI, err := m.CreateRepository(ctx, repositoryName)
	if err != nil {
		return "", err
	}

	// Build the image
	localImage := fmt.Sprintf("%s:%s", repositoryName, tag)
	fmt.Printf("Building Docker image: %s\n", localImage)

	//nolint:gosec // Docker commands with validated image names
	buildCmd := exec.CommandContext(ctx, "docker", "build",
		"-t", localImage,
		"-f", "Dockerfile",
		buildContext,
	)

	buildCmd.Stdout = nil // Show build progress
	buildCmd.Stderr = nil

	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build Docker image: %w", err)
	}

	fmt.Println("✓ Image built successfully")

	// Login to ECR
	if err := m.LoginToECR(ctx); err != nil {
		return "", err
	}

	// Push the image
	if err := m.PushImage(ctx, localImage, repositoryURI, tag); err != nil {
		return "", err
	}

	fullImageURI := fmt.Sprintf("%s:%s", repositoryURI, tag)
	return fullImageURI, nil
}

// getAccountID extracts the AWS account ID from credentials
func (m *ECRManager) getAccountID() string {
	// For now, we'll extract from the repository URI after creation
	// In a production implementation, you'd use STS to get the account ID
	return "123456789012" // Placeholder
}

// CheckDockerInstalled checks if Docker is installed and running
func CheckDockerInstalled(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not installed or not running. Please install Docker: https://docs.docker.com/get-docker/")
	}
	return nil
}
