package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/cicd"
)

var (
	workflowAll bool
)

var workflowCmd = &cobra.Command{
	Use:   "workflow [type]",
	Short: "Generate GitHub Actions workflows for CI/CD",
	Long: `Generate GitHub Actions workflows for CI/CD automation.

Available workflow types:
  validate  - Validate model.yaml on pull requests
  publish   - Publish models to catalog on releases
  deploy    - Deploy models on version tags
  all       - Generate all workflows

The workflows are created in .github/workflows/ directory.

Examples:
  conduit workflow all                # Generate all workflows
  conduit workflow validate           # Generate validate workflow only
  conduit workflow publish            # Generate publish workflow only
  conduit workflow deploy             # Generate deploy workflow only`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWorkflow,
}

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.Flags().BoolVar(&workflowAll, "all", false, "Generate all workflow files")
}

func runWorkflow(cmd *cobra.Command, args []string) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fmt.Println("Generating GitHub Actions workflows...")
	fmt.Printf("Directory: %s\n\n", currentDir)

	// Determine which workflows to generate
	if workflowAll || (len(args) > 0 && args[0] == "all") {
		return generateAllWorkflows(currentDir)
	}

	if len(args) == 0 {
		return fmt.Errorf("workflow type required: validate, publish, deploy, or all")
	}

	workflowType := args[0]
	var wf cicd.WorkflowType

	switch workflowType {
	case "validate":
		wf = cicd.WorkflowValidate
	case "publish":
		wf = cicd.WorkflowPublish
	case "deploy":
		wf = cicd.WorkflowDeploy
	default:
		return fmt.Errorf("unknown workflow type: %s (use: validate, publish, deploy, or all)", workflowType)
	}

	return generateSingleWorkflow(currentDir, wf, workflowType)
}

func generateAllWorkflows(dir string) error {
	workflows := []struct {
		name string
		wf   cicd.WorkflowType
	}{
		{"validate", cicd.WorkflowValidate},
		{"publish", cicd.WorkflowPublish},
		{"deploy", cicd.WorkflowDeploy},
	}

	generated := 0
	for _, w := range workflows {
		if err := cicd.GenerateWorkflow(w.wf, dir); err != nil {
			fmt.Printf("  ⚠️  Failed to create %s workflow (it may already exist)\n", w.name)
			continue
		}
		fmt.Printf("  ✓ Created %s workflow (.github/workflows/conduit-%s.yml)\n", w.name, w.name)
		generated++
	}

	if generated == 0 {
		return fmt.Errorf("no workflows were generated (they may already exist)")
	}

	fmt.Printf("\n✅ Generated %d workflow(s)\n\n", generated)
	printNextSteps()
	return nil
}

func generateSingleWorkflow(dir string, workflowType cicd.WorkflowType, name string) error {
	if err := cicd.GenerateWorkflow(workflowType, dir); err != nil {
		return fmt.Errorf("failed to generate %s workflow: %w", name, err)
	}

	fmt.Printf("  ✓ Created %s workflow (.github/workflows/conduit-%s.yml)\n", name, name)
	fmt.Printf("\n✅ Workflow generated successfully\n\n")
	printNextSteps()
	return nil
}

func printNextSteps() {
	fmt.Println("Next steps:")
	fmt.Println("  1. Review the generated workflow files in .github/workflows/")
	fmt.Println("  2. Commit the workflow files to your repository")
	fmt.Println("  3. Configure required secrets in GitHub repository settings:")
	fmt.Println("     - CONDUIT_CATALOG_PATH (for publish workflow)")
	fmt.Println("     - AWS credentials (for deploy workflow)")
	fmt.Println("  4. Create a pull request to test the workflows")
}
