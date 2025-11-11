package model

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/scttfrdmn/conduit/pkg/types"
)

// Parser handles parsing of model.yaml files
type Parser struct {
	// Can add configuration options here if needed
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile reads and parses a model.yaml file
func (p *Parser) ParseFile(path string) (*types.Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses YAML data into a Model struct
func (p *Parser) Parse(data []byte) (*types.Model, error) {
	var model types.Model

	// Use strict unmarshaling to catch unknown fields
	decoder := yaml.NewDecoder(nil)
	decoder.KnownFields(true)

	if err := yaml.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &model, nil
}

// ParseString parses a YAML string into a Model struct
func (p *Parser) ParseString(yamlString string) (*types.Model, error) {
	return p.Parse([]byte(yamlString))
}
