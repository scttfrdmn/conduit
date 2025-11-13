package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

var (
	searchDbPath      string
	searchDomain      string
	searchFramework   string
	searchGPU         string
	searchTags        []string
	searchLicense     string
	searchAuthor      string
	searchAfter       string
	searchBefore      string
	searchLimit       int
	searchSortBy      string
	searchFuzzy       bool
	searchMinScore    float64
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for models in the catalog",
	Long: `Search for models in the Conduit catalog.

Searches across model names, domains, and descriptions.
Can filter by domain, framework, GPU requirements, tags, license, author, and date.
Supports fuzzy matching for handling typos.

Examples:
  conduit search protein
  conduit search "protein folding"
  conduit search --domain protein-science
  conduit search --framework pytorch --gpu required
  conduit search --tag production --tag verified
  conduit search alphafold --limit 5
  conduit search --sort-by popular
  conduit search --fuzzy "alphafld"  # Finds alphafold2
  conduit search --license apache --author deepmind
  conduit search --after 2024-01-01 --before 2024-12-31`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVar(&searchDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
	searchCmd.Flags().StringVar(&searchDomain, "domain", "", "Filter by domain")
	searchCmd.Flags().StringVar(&searchFramework, "framework", "", "Filter by framework (pytorch, tensorflow, jax, onnx)")
	searchCmd.Flags().StringVar(&searchGPU, "gpu", "", "Filter by GPU requirement (required, optional, none)")
	searchCmd.Flags().StringSliceVar(&searchTags, "tag", []string{}, "Filter by tag (can be specified multiple times)")
	searchCmd.Flags().StringVar(&searchLicense, "license", "", "Filter by license")
	searchCmd.Flags().StringVar(&searchAuthor, "author", "", "Filter by author/organization in GitHub repo")
	searchCmd.Flags().StringVar(&searchAfter, "after", "", "Filter by created after date (YYYY-MM-DD)")
	searchCmd.Flags().StringVar(&searchBefore, "before", "", "Filter by created before date (YYYY-MM-DD)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 20, "Maximum number of results")
	searchCmd.Flags().StringVar(&searchSortBy, "sort-by", "updated", "Sort results by (popular, name, created, updated)")
	searchCmd.Flags().BoolVar(&searchFuzzy, "fuzzy", false, "Enable fuzzy matching for typos")
	searchCmd.Flags().Float64Var(&searchMinScore, "min-score", 0.6, "Minimum relevance score for fuzzy matching (0.0-1.0)")
}

func runSearch(cmd *cobra.Command, args []string) (err error) {
	// Get query from args
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	// Open catalog database
	db, err := catalog.NewDB(searchDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Build search options
	opts := catalog.SearchOptions{
		Query:         query,
		Domain:        searchDomain,
		Framework:     searchFramework,
		Tags:          searchTags,
		License:       searchLicense,
		Author:        searchAuthor,
		CreatedAfter:  searchAfter,
		CreatedBefore: searchBefore,
		Limit:         searchLimit,
		Offset:        0,
		SortBy:        searchSortBy,
		FuzzyMatch:    searchFuzzy,
		MinScore:      searchMinScore,
	}

	// Handle GPU filter
	if searchGPU != "" {
		switch strings.ToLower(searchGPU) {
		case "required", "yes", "true":
			opts.GPURequired = boolPtr(true)
		case "optional", "no", "false":
			opts.GPURequired = boolPtr(false)
		default:
			return fmt.Errorf("invalid GPU filter: %s (use: required, optional)", searchGPU)
		}
	}

	// Execute search
	results, err := db.Search(opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Display results
	if len(results) == 0 {
		fmt.Println("No models found matching your criteria.")
		return nil
	}

	fmt.Printf("Found %d model(s):\n\n", len(results))

	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Name)
		fmt.Printf("   Domain: %s\n", result.Domain)
		if result.Description != "" {
			// Truncate long descriptions
			desc := result.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("   Description: %s\n", desc)
		}
		if result.GitHubRepo != "" {
			fmt.Printf("   GitHub: %s\n", result.GitHubRepo)
		}
		fmt.Printf("   Updated: %s\n", result.UpdatedAt.Format("2006-01-02"))
		fmt.Println()
	}

	// Show command to get more details
	if len(results) > 0 {
		fmt.Printf("View details with: conduit info <model-name>\n")
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
