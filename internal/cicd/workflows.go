package cicd

import (
	"fmt"
	"os"
	"path/filepath"
)

// WorkflowType represents the type of CI/CD workflow
type WorkflowType string

const (
	// WorkflowValidate validates model.yaml on pull requests
	WorkflowValidate WorkflowType = "validate"
	// WorkflowPublish publishes model to catalog on release
	WorkflowPublish WorkflowType = "publish"
	// WorkflowDeploy deploys model on version tags
	WorkflowDeploy WorkflowType = "deploy"
)

// GenerateWorkflow creates a GitHub Actions workflow file
func GenerateWorkflow(workflowType WorkflowType, outputDir string) error {
	// Create .github/workflows directory if it doesn't exist
	workflowDir := filepath.Join(outputDir, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil { //nolint:gosec // Standard directory permissions for workflow files
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	var content string
	var filename string

	switch workflowType {
	case WorkflowValidate:
		filename = "conduit-validate.yml"
		content = validateWorkflowTemplate
	case WorkflowPublish:
		filename = "conduit-publish.yml"
		content = publishWorkflowTemplate
	case WorkflowDeploy:
		filename = "conduit-deploy.yml"
		content = deployWorkflowTemplate
	default:
		return fmt.Errorf("unknown workflow type: %s", workflowType)
	}

	filePath := filepath.Join(workflowDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("workflow file already exists: %s", filePath)
	}

	// Write workflow file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil { //nolint:gosec // Standard file permissions for workflow files
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	return nil
}

// GenerateAllWorkflows creates all workflow files
func GenerateAllWorkflows(outputDir string) error {
	workflows := []WorkflowType{
		WorkflowValidate,
		WorkflowPublish,
		WorkflowDeploy,
	}

	for _, workflow := range workflows {
		if err := GenerateWorkflow(workflow, outputDir); err != nil {
			// If file already exists, skip it
			if _, ok := err.(*os.PathError); !ok {
				return err
			}
		}
	}

	return nil
}

const validateWorkflowTemplate = `name: Validate Model

on:
  pull_request:
    paths:
      - 'model.yaml'
      - 'models/**'
  push:
    branches:
      - main
      - develop

jobs:
  validate:
    name: Validate Model Configuration
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install Conduit
        run: |
          go install github.com/scttfrdmn/conduit@latest

      - name: Validate model.yaml
        run: |
          conduit validate model.yaml --strict

      - name: Comment on PR
        if: failure() && github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '❌ Model validation failed. Please check the logs for details.'
            })
`

const publishWorkflowTemplate = `name: Publish Model

on:
  release:
    types: [published]
  workflow_dispatch:
    inputs:
      version:
        description: 'Model version to publish'
        required: true
        type: string

jobs:
  publish:
    name: Publish to Conduit Catalog
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install Conduit
        run: |
          go install github.com/scttfrdmn/conduit@latest

      - name: Validate model
        run: |
          conduit validate model.yaml --strict

      - name: Publish to catalog
        env:
          CONDUIT_CATALOG_PATH: ${{ secrets.CONDUIT_CATALOG_PATH }}
        run: |
          # Set version from release tag or manual input
          VERSION="${{ github.event.release.tag_name || github.event.inputs.version }}"

          # Update version in model.yaml
          sed -i "s/version:.*/version: ${VERSION}/" model.yaml

          # Publish to catalog
          conduit publish model.yaml

      - name: Create release notes
        if: github.event_name == 'release'
        run: |
          echo "Model published successfully"
`

const deployWorkflowTemplate = `name: Deploy Model

on:
  push:
    tags:
      - 'v*.*.*'
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deployment environment'
        required: true
        type: choice
        options:
          - development
          - staging
          - production

jobs:
  deploy:
    name: Deploy Model
    runs-on: ubuntu-latest
    environment: ${{ github.event.inputs.environment || 'production' }}

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install Conduit
        run: |
          go install github.com/scttfrdmn/conduit@latest

      - name: Validate model
        run: |
          conduit validate model.yaml --strict

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ secrets.AWS_REGION || 'us-east-1' }}

      - name: Deploy to AWS
        run: |
          # Parse model configuration
          MODEL_NAME=$(grep 'name:' model.yaml | awk '{print $2}')
          VERSION=$(grep 'version:' model.yaml | awk '{print $2}')

          echo "Deploying ${MODEL_NAME} v${VERSION}..."

          # TODO: Add actual deployment commands here
          # Example: conduit deploy model.yaml --platform bedrock
          # Example: conduit deploy model.yaml --platform sagemaker

          echo "Deployment placeholder - configure with your deployment commands"

      - name: Create deployment record
        run: |
          echo "Deployment completed at $(date)" >> deployment-history.txt

      - name: Notify deployment status
        if: always()
        run: |
          echo "Deployment status: ${{ job.status }}"
          echo "Environment: ${{ github.event.inputs.environment || 'production' }}"
`
