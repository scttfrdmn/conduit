package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// Client handles communication with a remote registry
type Client struct {
	registry *Registry
	client   *http.Client
}

// NewClient creates a new registry client
func NewClient(reg *Registry) (*Client, error) {
	return &Client{
		registry: reg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Push uploads a model export to the registry
func (c *Client) Push(exportPath, modelName, version string, force bool) error {
	switch c.registry.Type {
	case "http":
		return c.pushHTTP(exportPath, modelName, version, force)
	case "s3":
		return c.pushS3(exportPath, modelName, version, force)
	case "git":
		return c.pushGit(exportPath, modelName, version, force)
	default:
		return fmt.Errorf("unsupported registry type: %s", c.registry.Type)
	}
}

// Pull downloads a model export from the registry
func (c *Client) Pull(modelName, version, downloadPath string) error {
	switch c.registry.Type {
	case "http":
		return c.pullHTTP(modelName, version, downloadPath)
	case "s3":
		return c.pullS3(modelName, version, downloadPath)
	case "git":
		return c.pullGit(modelName, version, downloadPath)
	default:
		return fmt.Errorf("unsupported registry type: %s", c.registry.Type)
	}
}

// pushHTTP pushes to an HTTP registry
func (c *Client) pushHTTP(exportPath, modelName, version string, force bool) error {
	// Read the export file
	data, err := os.ReadFile(exportPath) //nolint:gosec // User-provided export file
	if err != nil {
		return fmt.Errorf("failed to read export file: %w", err)
	}

	// Build the URL
	pushURL := fmt.Sprintf("%s/models/%s", c.registry.URL, modelName)
	if version != "" {
		pushURL = fmt.Sprintf("%s/versions/%s", pushURL, version)
	}

	if force {
		pushURL += "?force=true"
	}

	// Create request
	req, err := http.NewRequest(http.MethodPut, pushURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication if configured
	if c.registry.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.registry.Token)
	} else if c.registry.Username != "" {
		req.SetBasicAuth(c.registry.Username, c.registry.Token)
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to push to registry: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry returned error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// pullHTTP pulls from an HTTP registry
func (c *Client) pullHTTP(modelName, version, downloadPath string) error {
	// Build the URL
	pullURL := fmt.Sprintf("%s/models/%s", c.registry.URL, modelName)
	if version != "" {
		pullURL = fmt.Sprintf("%s/versions/%s", pullURL, version)
	}

	// Create request
	req, err := http.NewRequest(http.MethodGet, pullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if configured
	if c.registry.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.registry.Token)
	} else if c.registry.Username != "" {
		req.SetBasicAuth(c.registry.Username, c.registry.Token)
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pull from registry: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry returned error: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Write to file
	file, err := os.Create(downloadPath) //nolint:gosec // User-provided download path
	if err != nil {
		return fmt.Errorf("failed to create download file: %w", err)
	}
	defer file.Close() //nolint:errcheck // Best effort close

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write downloaded data: %w", err)
	}

	return nil
}

// pushS3 pushes to an S3-backed registry
func (c *Client) pushS3(exportPath, modelName, version string, force bool) error {
	// TODO: Implement S3 push
	// This would use the AWS SDK to upload the export file to S3
	return fmt.Errorf("s3 registry type not yet implemented")
}

// pullS3 pulls from an S3-backed registry
func (c *Client) pullS3(modelName, version, downloadPath string) error {
	// TODO: Implement S3 pull
	// This would use the AWS SDK to download the export file from S3
	return fmt.Errorf("s3 registry type not yet implemented")
}

// pushGit pushes to a Git-backed registry
func (c *Client) pushGit(exportPath, modelName, version string, force bool) error {
	// TODO: Implement Git push
	// This would create a GitHub release with the export file
	return fmt.Errorf("git registry type not yet implemented")
}

// pullGit pulls from a Git-backed registry
func (c *Client) pullGit(modelName, version, downloadPath string) error {
	// TODO: Implement Git pull
	// This would download a release asset from GitHub
	return fmt.Errorf("git registry type not yet implemented")
}

// Search searches for models in the registry
func (c *Client) Search(query string) ([]SearchResult, error) {
	switch c.registry.Type {
	case "http":
		return c.searchHTTP(query)
	case "s3":
		return c.searchS3(query)
	case "git":
		return c.searchGit(query)
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", c.registry.Type)
	}
}

// SearchResult represents a search result from a registry
type SearchResult struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
}

// searchHTTP searches an HTTP registry
func (c *Client) searchHTTP(query string) ([]SearchResult, error) {
	// Build the URL
	searchURL := fmt.Sprintf("%s/search?q=%s", c.registry.URL, url.QueryEscape(query))

	// Create request
	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if configured
	if c.registry.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.registry.Token)
	} else if c.registry.Username != "" {
		req.SetBasicAuth(c.registry.Username, c.registry.Token)
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search registry: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registry returned error: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	var results []SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	return results, nil
}

// searchS3 searches an S3-backed registry
func (c *Client) searchS3(query string) ([]SearchResult, error) {
	// TODO: Implement S3 search
	// This would list S3 objects and filter by query
	return nil, fmt.Errorf("s3 registry search not yet implemented")
}

// searchGit searches a Git-backed registry
func (c *Client) searchGit(query string) ([]SearchResult, error) {
	// TODO: Implement Git search
	// This would search GitHub releases
	return nil, fmt.Errorf("git registry search not yet implemented")
}

// GetModelPath returns the path where a model would be stored in the registry
func GetModelPath(modelName, version string) string {
	if version == "" {
		return filepath.Join("models", modelName, "latest.json")
	}
	return filepath.Join("models", modelName, "versions", version+".json")
}
