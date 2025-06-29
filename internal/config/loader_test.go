package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_TOML(t *testing.T) {
	content := `
[metadata]
title = "Test Config"
version = "1.0.0"

[sections.test]
order = 1
content = "Test content"
`
	config, err := testLoadFromContent(t, content, "test.toml")
	require.NoError(t, err)

	assert.Equal(t, "Test Config", config.Metadata.Title)
	assert.Equal(t, "Test content", config.Sections["test"].Content)
	assert.Equal(t, FormatTOML, config.SourceFormat)
}

func TestLoadConfig_YAML(t *testing.T) {
	content := `
metadata:
  title: "Test Config"
  version: "1.0.0"
sections:
  test:
    order: 1
    content: "Test content"
`
	config, err := testLoadFromContent(t, content, "test.yaml")
	require.NoError(t, err)

	assert.Equal(t, "Test Config", config.Metadata.Title)
	assert.Equal(t, "Test content", config.Sections["test"].Content)
	assert.Equal(t, FormatYAML, config.SourceFormat)
}

func TestLoadConfig_Markdown(t *testing.T) {
	content := `---
title: "Test Config"
version: "1.0.0"
---

# Test Header

This is test content that should become a section.

## Another Section

More content here.
`
	config, err := testLoadFromContent(t, content, "test.md")
	require.NoError(t, err)

	assert.Equal(t, "Test Config", config.Metadata.Title)
	assert.Equal(t, FormatMarkdown, config.SourceFormat)
	assert.NotEmpty(t, config.Sections)
}

func TestLoadConfig_WithPriorities(t *testing.T) {
	content := `
[metadata]
title = "Priority Test"
priority = { type = "explicit", value = 10 }

[sections.high_priority]
order = 1
content = "High priority content"
priority = { type = "explicit", value = 20 }

[sections.low_priority]
order = 2
content = "Low priority content"
priority = { type = "relative", value = 5 }
`
	config, err := testLoadFromContent(t, content, "test.toml")
	require.NoError(t, err)

	assert.Equal(t, PriorityExplicit, config.Metadata.Priority.Type)
	assert.Equal(t, 10, config.Metadata.Priority.Value)
	assert.Equal(t, PriorityExplicit, config.Sections["high_priority"].Priority.Type)
	assert.Equal(t, PriorityRelative, config.Sections["low_priority"].Priority.Type)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.toml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestLoadConfig_UnsupportedFormat(t *testing.T) {
	content := "some content"
	tmpfile, err := os.CreateTemp("", "test*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	tmpfile.Close()

	_, err = LoadConfig(tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file format")
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	content := `
[metadata
title = "Broken TOML"
`
	tmpfile, err := os.CreateTemp("", "test*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	tmpfile.Close()

	_, err = LoadConfig(tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TOML parse error")
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	content := `
metadata:
  title: "Test"
  - invalid YAML structure
`
	tmpfile, err := os.CreateTemp("", "test*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	tmpfile.Close()

	_, err = LoadConfig(tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YAML parse error")
}

// Helper function for testing
func testLoadFromContent(t *testing.T, content, filename string) (*Config, error) {
	tmpfile, err := os.CreateTemp("", "test_*_"+filename)
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	tmpfile.Close()

	return LoadConfig(tmpfile.Name())
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Metadata: Metadata{Title: "Test"},
				Sections: map[string]Section{"test": {Content: "content"}},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			config: &Config{
				Sections: map[string]Section{"test": {Content: "content"}},
			},
			wantErr: true,
		},
		{
			name: "no sections",
			config: &Config{
				Metadata: Metadata{Title: "Test"},
			},
			wantErr: true,
		},
		{
			name: "invalid priority value",
			config: &Config{
				Metadata: Metadata{Title: "Test"},
				Sections: map[string]Section{
					"test": {
						Content:  "content",
						Priority: Priority{Type: PriorityExplicit, Value: -1},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}