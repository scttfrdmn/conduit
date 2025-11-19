package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// EnvironmentType represents where the code is running
type EnvironmentType string

const (
	// EnvStudioLab indicates SageMaker Studio Lab environment
	EnvStudioLab EnvironmentType = "studio-lab"
	// EnvStudio indicates SageMaker Studio environment
	EnvStudio EnvironmentType = "studio"
	// EnvNotebookInstance indicates SageMaker Notebook Instance environment
	EnvNotebookInstance EnvironmentType = "notebook-instance"
	// EnvLocal indicates local development environment
	EnvLocal EnvironmentType = "local"
)

// DetectEnvironment determines where code is running
func DetectEnvironment() EnvironmentType {
	// Check for SageMaker Studio
	if os.Getenv("SAGEMAKER_INTERNAL_IMAGE_URI") != "" {
		// Further check if Studio Lab vs Studio
		if os.Getenv("STUDIO_LAB_USER_ID") != "" {
			return EnvStudioLab
		}
		return EnvStudio
	}

	// Check for Notebook Instance
	if os.Getenv("SM_TRAINING_ENV") != "" {
		return EnvNotebookInstance
	}

	return EnvLocal
}

// GetAWSConfig returns appropriate AWS config based on environment
func GetAWSConfig(ctx context.Context) (aws.Config, error) {
	env := DetectEnvironment()

	switch env {
	case EnvStudio, EnvNotebookInstance:
		// Use SageMaker execution role automatically
		return config.LoadDefaultConfig(ctx)

	case EnvStudioLab:
		// Studio Lab: Need user to configure credentials
		// But can deploy to their AWS account if they provide them
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return aws.Config{}, fmt.Errorf("AWS credentials not configured. Please configure your AWS credentials to deploy from Studio Lab: %w", err)
		}
		return cfg, nil

	case EnvLocal:
		// Local: Use standard AWS credential chain
		return config.LoadDefaultConfig(ctx)
	}

	return config.LoadDefaultConfig(ctx)
}

// GetExecutionRole returns SageMaker execution role if available
func GetExecutionRole() (string, error) {
	env := DetectEnvironment()

	switch env {
	case EnvStudio, EnvNotebookInstance:
		// Get from SageMaker metadata
		role := os.Getenv("SAGEMAKER_ROLE_ARN")
		if role == "" {
			return "", fmt.Errorf("SAGEMAKER_ROLE_ARN not found in environment")
		}
		return role, nil

	default:
		return "", fmt.Errorf("not running in SageMaker environment")
	}
}

// GetRegion returns the AWS region
func GetRegion(ctx context.Context) (string, error) {
	// Try environment variable first
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region, nil
	}

	// Try AWS config
	cfg, err := GetAWSConfig(ctx)
	if err != nil {
		return "", err
	}

	if cfg.Region == "" {
		return "us-east-1", nil // Default region
	}

	return cfg.Region, nil
}

// HasAWSCredentials checks if AWS credentials are configured
func HasAWSCredentials(ctx context.Context) bool {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return false
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return false
	}

	return creds.AccessKeyID != "" && creds.SecretAccessKey != ""
}
