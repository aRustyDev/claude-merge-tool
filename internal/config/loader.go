package config

import (
	"fmt"
	"os"
)

// LoadConfig reads a configuration file and returns a Config struct
// Supports TOML, YAML, and Markdown formats
func LoadConfig(filename string) (*Config, error) {
	// Step 1: Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Step 2: Detect format
	format, err := DetectFormat(filename)
	if err != nil {
		return nil, err
	}

	// Step 3: Parse based on format
	config, err := ParseConfig(data, format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	// Step 4: Set source metadata
	config.SourceFile = filename
	config.SourceFormat = format

	return config, nil
}

// ValidateConfig checks if a config is valid
func ValidateConfig(config *Config) error {
	if config.Metadata.Title == "" {
		return fmt.Errorf("config missing title in metadata")
	}

	if len(config.Sections) == 0 {
		return fmt.Errorf("config has no sections")
	}

	// Validate priorities
	for name, section := range config.Sections {
		if section.Priority.Type != PriorityNone && section.Priority.Value < 0 {
			return fmt.Errorf("section %s has invalid priority value: %d", name, section.Priority.Value)
		}
	}

	return nil
}