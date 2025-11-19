package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	awsenv "github.com/scttfrdmn/conduit/internal/aws"
	"github.com/scttfrdmn/conduit/internal/deployment"
	"github.com/scttfrdmn/conduit/internal/model"
)

var (
	deployPlatform     string
	deployInstanceType string
	deployEndpointName string
	deployRegion       string
	deployRole         string
	deployDryRun       bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy [model.yaml]",
	Short: "Deploy a model to AWS (Bedrock or SageMaker)",
	Long: `Deploy a model to AWS Bedrock or SageMaker.

Platforms:
  bedrock    - AWS Bedrock (serverless, default)
  sagemaker  - SageMaker Endpoints (managed instances)

The deploy command will:
  1. Build a Docker container from your model specification
  2. Push the container to ECR (Elastic Container Registry)
  3. Create an endpoint configuration
  4. Deploy the endpoint
  5. Perform health checks

Environment Detection:
  - Local: Requires AWS credentials configured
  - SageMaker Studio: Uses execution role automatically
  - SageMaker Studio Lab: Prompts for AWS credentials
  - SageMaker Notebook Instance: Uses execution role automatically

Examples:
  conduit deploy model.yaml
  conduit deploy model.yaml --platform bedrock
  conduit deploy model.yaml --platform sagemaker --instance-type ml.g5.2xlarge
  conduit deploy model.yaml --endpoint-name my-model-endpoint
  conduit deploy model.yaml --dry-run  # Preview without deploying`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVar(&deployPlatform, "platform", "bedrock", "Deployment platform (bedrock or sagemaker)")
	deployCmd.Flags().StringVar(&deployInstanceType, "instance-type", "", "Instance type (e.g., ml.g5.2xlarge)")
	deployCmd.Flags().StringVar(&deployEndpointName, "endpoint-name", "", "Custom endpoint name")
	deployCmd.Flags().StringVar(&deployRegion, "region", "", "AWS region (default: from AWS config)")
	deployCmd.Flags().StringVar(&deployRole, "role", "", "IAM role ARN (default: auto-detect)")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Preview deployment without executing")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine the path to model.yaml
	path := "model.yaml"
	if len(args) > 0 {
		path = args[0]
	}

	// If path is a directory, look for model.yaml in it
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	if info.IsDir() {
		path = filepath.Join(path, "model.yaml")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	// Detect environment
	env := awsenv.DetectEnvironment()
	fmt.Printf("Environment: %s\n", env)

	// Check AWS credentials
	if !awsenv.HasAWSCredentials(ctx) {
		return printCredentialHelp(env)
	}

	// Load model configuration
	parser := model.NewParser()
	m, err := parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("failed to parse model file: %w", err)
	}

	fmt.Printf("Model: %s v%s\n", m.Name, m.Version)
	fmt.Printf("Platform: %s\n", deployPlatform)
	fmt.Println()

	// Determine instance type
	instanceType := deployInstanceType
	if instanceType == "" {
		instanceType = m.Hardware.RecommendedInstance
		if instanceType == "" {
			instanceType = "ml.g5.2xlarge" // Default
		}
	}

	// Determine endpoint name
	endpointName := deployEndpointName
	if endpointName == "" {
		endpointName = fmt.Sprintf("%s-%s", m.Name, m.Version)
	}

	// Get region
	region := deployRegion
	if region == "" {
		region, err = awsenv.GetRegion(ctx)
		if err != nil {
			return fmt.Errorf("failed to get AWS region: %w", err)
		}
	}

	// Get role (if needed)
	role := deployRole
	if role == "" && (env == awsenv.EnvStudio || env == awsenv.EnvNotebookInstance) {
		role, err = awsenv.GetExecutionRole()
		if err != nil {
			fmt.Printf("Warning: Could not auto-detect execution role: %v\n", err)
		}
	}

	// Create deployment config
	deployConfig := &deployment.Config{
		Model:        m,
		Platform:     deployment.Platform(deployPlatform),
		InstanceType: instanceType,
		EndpointName: endpointName,
		Region:       region,
		Role:         role,
		DryRun:       deployDryRun,
	}

	// Validate deployment config
	if err := deployConfig.Validate(); err != nil {
		return fmt.Errorf("invalid deployment configuration: %w", err)
	}

	// Print deployment plan
	fmt.Println("Deployment Plan:")
	fmt.Printf("  Platform:      %s\n", deployConfig.Platform)
	fmt.Printf("  Instance Type: %s\n", deployConfig.InstanceType)
	fmt.Printf("  Endpoint Name: %s\n", deployConfig.EndpointName)
	fmt.Printf("  Region:        %s\n", deployConfig.Region)
	if deployConfig.Role != "" {
		fmt.Printf("  IAM Role:      %s\n", deployConfig.Role)
	}
	fmt.Println()

	if deployDryRun {
		fmt.Println("Dry run mode - no resources will be created")
		return nil
	}

	// Create deployer
	deployer, err := deployment.NewDeployer(ctx, deployConfig)
	if err != nil {
		return fmt.Errorf("failed to create deployer: %w", err)
	}

	// Execute deployment
	fmt.Println("Starting deployment...")
	result, err := deployer.Deploy(ctx)
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// Print results
	fmt.Println()
	fmt.Println("✅ Deployment successful!")
	fmt.Printf("Endpoint Name: %s\n", result.EndpointName)
	fmt.Printf("Endpoint ARN:  %s\n", result.EndpointARN)
	fmt.Printf("Status:        %s\n", result.Status)
	if result.URL != "" {
		fmt.Printf("URL:           %s\n", result.URL)
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Test your endpoint with sample predictions")
	fmt.Println("  - Monitor endpoint metrics in AWS Console")
	fmt.Println("  - Set up auto-scaling policies if needed")

	return nil
}

func printCredentialHelp(env awsenv.EnvironmentType) error {
	fmt.Println()
	fmt.Println("❌ AWS credentials not configured")
	fmt.Println()

	switch env {
	case awsenv.EnvStudioLab:
		fmt.Println("To deploy from SageMaker Studio Lab:")
		fmt.Println("  1. Click 'AWS' in the Studio Lab menu")
		fmt.Println("  2. Enter your AWS Access Key ID and Secret Access Key")
		fmt.Println("  3. Run this command again")
		fmt.Println()
		fmt.Println("Don't have an AWS account? Create one at: https://aws.amazon.com/free/")

	case awsenv.EnvLocal:
		fmt.Println("To configure AWS credentials locally:")
		fmt.Println("  1. Install AWS CLI: https://aws.amazon.com/cli/")
		fmt.Println("  2. Run: aws configure")
		fmt.Println("  3. Enter your AWS Access Key ID and Secret Access Key")
		fmt.Println()
		fmt.Println("Or set environment variables:")
		fmt.Println("  export AWS_ACCESS_KEY_ID=your_key_id")
		fmt.Println("  export AWS_SECRET_ACCESS_KEY=your_secret_key")
		fmt.Println("  export AWS_REGION=us-east-1")

	default:
		fmt.Println("AWS credentials not found. Please configure AWS credentials.")
	}

	return fmt.Errorf("AWS credentials required")
}
