package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/scttfrdmn/conduit/internal/version"
)

var (
	// Version flag
	versionFlag bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "conduit",
	Short: "AWS-native platform for scientific ML model publishing and deployment",
	Long: `Conduit makes deploying scientific models as simple as installing software packages.

Deploy models to AWS Bedrock, manage benchmarks, generate UIs, and more.

Examples:
  conduit search "protein folding"
  conduit deploy alphafold2-multimer
  conduit publish --github yourorg/your-model`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			info := version.Get()
			fmt.Println(info.String())
			return
		}
		_ = cmd.Help() // Ignore error as we're already in error handler context
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print version information")
}
