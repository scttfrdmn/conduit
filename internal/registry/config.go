package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Registry represents a remote registry configuration
type Registry struct {
	Name     string `yaml:"name" json:"name"`
	URL      string `yaml:"url" json:"url"`
	Type     string `yaml:"type" json:"type"` // http, s3, git
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Token    string `yaml:"token,omitempty" json:"token,omitempty"`
	Default  bool   `yaml:"default" json:"default"`
}

// Config holds all registry configurations
type Config struct {
	Registries []Registry `yaml:"registries" json:"registries"`
}

// GetConfigPath returns the path to the registry config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".conduit")
	if err := os.MkdirAll(configDir, 0755); err != nil { //nolint:gosec // Standard user config directory permissions
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "registries.yaml"), nil
}

// LoadConfig loads the registry configuration from disk
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{Registries: []Registry{}}, nil
	}

	data, err := os.ReadFile(configPath) //nolint:gosec // User's own config file
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the registry configuration to disk
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil { //nolint:gosec // Config file may contain tokens
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddRegistry adds a new registry to the configuration
func (c *Config) AddRegistry(reg Registry) error {
	// Check if registry with same name already exists
	for i, existing := range c.Registries {
		if existing.Name == reg.Name {
			// Update existing registry
			c.Registries[i] = reg
			return nil
		}
	}

	// If this is the first registry or marked as default, make it default
	if len(c.Registries) == 0 || reg.Default {
		// Unset any existing defaults
		for i := range c.Registries {
			c.Registries[i].Default = false
		}
		reg.Default = true
	}

	c.Registries = append(c.Registries, reg)
	return nil
}

// RemoveRegistry removes a registry by name
func (c *Config) RemoveRegistry(name string) error {
	for i, reg := range c.Registries {
		if reg.Name == name {
			c.Registries = append(c.Registries[:i], c.Registries[i+1:]...)

			// If we removed the default and there are other registries, make the first one default
			if reg.Default && len(c.Registries) > 0 {
				c.Registries[0].Default = true
			}

			return nil
		}
	}
	return fmt.Errorf("registry not found: %s", name)
}

// GetRegistry returns a registry by name
func (c *Config) GetRegistry(name string) (*Registry, error) {
	for _, reg := range c.Registries {
		if reg.Name == name {
			return &reg, nil
		}
	}
	return nil, fmt.Errorf("registry not found: %s", name)
}

// GetDefaultRegistry returns the default registry
func (c *Config) GetDefaultRegistry() (*Registry, error) {
	for _, reg := range c.Registries {
		if reg.Default {
			return &reg, nil
		}
	}

	// If no default is set but we have registries, return the first one
	if len(c.Registries) > 0 {
		return &c.Registries[0], nil
	}

	return nil, fmt.Errorf("no registries configured")
}

// SetDefault sets a registry as the default
func (c *Config) SetDefault(name string) error {
	found := false
	for i := range c.Registries {
		if c.Registries[i].Name == name {
			c.Registries[i].Default = true
			found = true
		} else {
			c.Registries[i].Default = false
		}
	}

	if !found {
		return fmt.Errorf("registry not found: %s", name)
	}

	return nil
}
