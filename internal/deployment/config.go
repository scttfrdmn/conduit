package deployment

import (
	"fmt"

	"github.com/scttfrdmn/conduit/pkg/types"
)

// Platform represents the deployment platform
type Platform string

const (
	// PlatformBedrock deploys to AWS Bedrock
	PlatformBedrock Platform = "bedrock"
	// PlatformSageMaker deploys to SageMaker Endpoints
	PlatformSageMaker Platform = "sagemaker"
)

// Config holds deployment configuration
type Config struct {
	Model        *types.Model
	Platform     Platform
	InstanceType string
	EndpointName string
	Region       string
	Role         string
	DryRun       bool
}

// Validate checks if the deployment configuration is valid
func (c *Config) Validate() error {
	if c.Model == nil {
		return fmt.Errorf("model is required")
	}

	if c.Model.Name == "" {
		return fmt.Errorf("model name is required")
	}

	if c.Model.Version == "" {
		return fmt.Errorf("model version is required")
	}

	if c.Platform != PlatformBedrock && c.Platform != PlatformSageMaker {
		return fmt.Errorf("invalid platform: %s (must be 'bedrock' or 'sagemaker')", c.Platform)
	}

	if c.InstanceType == "" {
		return fmt.Errorf("instance type is required")
	}

	if c.EndpointName == "" {
		return fmt.Errorf("endpoint name is required")
	}

	if c.Region == "" {
		return fmt.Errorf("region is required")
	}

	// SageMaker requires a role
	if c.Platform == PlatformSageMaker && c.Role == "" {
		return fmt.Errorf("IAM role is required for SageMaker deployment")
	}

	return nil
}

// DeploymentResult contains the result of a deployment
type DeploymentResult struct {
	EndpointName string
	EndpointARN  string
	Status       string
	URL          string
	Region       string
	Platform     Platform
}
