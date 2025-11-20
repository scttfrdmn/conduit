package deployment

import (
	"context"
	"fmt"

	awsenv "github.com/scttfrdmn/conduit/internal/aws"
)

// Deployer handles model deployment
type Deployer interface {
	Deploy(ctx context.Context) (*DeploymentResult, error)
	Validate(ctx context.Context) error
}

// NewDeployer creates a new deployer based on the platform
func NewDeployer(ctx context.Context, config *Config) (Deployer, error) {
	// Get AWS config
	awsConfig, err := awsenv.GetAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	switch config.Platform {
	case PlatformBedrock:
		return NewBedrockDeployer(awsConfig, config), nil
	case PlatformSageMaker:
		return NewSageMakerDeployer(awsConfig, config), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", config.Platform)
	}
}
