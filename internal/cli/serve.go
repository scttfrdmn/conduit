package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scttfrdmn/conduit/internal/catalog"
	"github.com/scttfrdmn/conduit/internal/web"
)

var (
	serveDbPath string
	serveAddr   string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web UI server",
	Long: `Start a web server providing a browser-based interface for the catalog.

The web UI provides:
  - Browse and search models with a visual interface
  - Model detail pages with comprehensive information
  - Real-time search with fuzzy matching
  - Filtering by domain, framework, and tags
  - Usage statistics and version history

The server will continue running until you stop it with Ctrl+C.

Examples:
  conduit serve
  conduit serve --addr localhost:8080
  conduit serve --addr 0.0.0.0:3000`,
	Args: cobra.NoArgs,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&serveDbPath, "db", "", "Path to catalog database (default: ~/.conduit/catalog.db)")
	serveCmd.Flags().StringVar(&serveAddr, "addr", "localhost:8080", "Address to listen on (host:port)")
}

func runServe(cmd *cobra.Command, args []string) (err error) {
	// Open catalog database
	db, err := catalog.NewDB(serveDbPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Create web server
	server, err := web.NewServer(db, serveAddr)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	// Start server (blocks until shutdown)
	if err := server.Start(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
