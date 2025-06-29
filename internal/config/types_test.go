package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_UnmarshalTOML(t *testing.T) {
	tomlContent := `
[metadata]
title = "Test Config"
priority = { type = "explicit", value = 10 }

[sections.header]
order = 1
priority = { type = "explicit", value = 5 }
content = "# Header"
`

	config, err := ParseConfig([]byte(tomlContent), FormatTOML)
	require.NoError(t, err)

	assert.Equal(t, "Test Config", config.Metadata.Title)
	assert.Equal(t, PriorityExplicit, config.Metadata.Priority.Type)
	assert.Equal(t, 10, config.Metadata.Priority.Value)
	assert.Equal(t, "# Header", config.Sections["header"].Content)
}

func TestConfig_UnmarshalYAML(t *testing.T) {
	yamlContent := `
metadata:
  title: "Test Config"
  priority:
    type: "explicit"
    value: 10
sections:
  header:
    order: 1
    priority:
      type: "explicit"
      value: 5
    content: "# Header"
`

	config, err := ParseConfig([]byte(yamlContent), FormatYAML)
	require.NoError(t, err)

	assert.Equal(t, "Test Config", config.Metadata.Title)
	assert.Equal(t, PriorityExplicit, config.Metadata.Priority.Type)
	assert.Equal(t, 10, config.Metadata.Priority.Value)
}

func TestConfig_ParseMarkdown(t *testing.T) {
	mdContent := `---
title: "Test Config"
priority:
  type: "explicit"
  value: 10
---

# Header Content

This is markdown content that should be treated as a section.

## Subsection

More content here.

### Another Level

Even more content.

- List item 1
- List item 2

1. Ordered item 1
2. Ordered item 2
`

	config, err := ParseConfig([]byte(mdContent), FormatMarkdown)
	require.NoError(t, err)

	assert.Equal(t, "Test Config", config.Metadata.Title)
	assert.Equal(t, FormatMarkdown, config.SourceFormat)
	assert.NotEmpty(t, config.Sections)

	// Should have a single content section with all markdown content
	assert.Len(t, config.Sections, 1, "Should have exactly one section")
	
	contentSection, exists := config.Sections["content"]
	assert.True(t, exists, "Should have a 'content' section")
	
	// Verify all content is preserved in the single section
	assert.Contains(t, contentSection.Content, "# Header Content")
	assert.Contains(t, contentSection.Content, "## Subsection")
	assert.Contains(t, contentSection.Content, "### Another Level")
	assert.Contains(t, contentSection.Content, "- List item 1")
	assert.Contains(t, contentSection.Content, "1. Ordered item 1")
}

func TestPriority_Comparison(t *testing.T) {
	tests := []struct {
		name     string
		p1, p2   Priority
		expected bool
	}{
		{"explicit beats relative", NewExplicitPriority(5), NewRelativePriority(10), true},
		{"higher explicit beats lower", NewExplicitPriority(10), NewExplicitPriority(5), true},
		{"higher relative beats lower", NewRelativePriority(10), NewRelativePriority(5), true},
		{"explicit beats none", NewExplicitPriority(1), Priority{}, true},
		{"relative beats none", NewRelativePriority(1), Priority{}, true},
		{"same explicit values", NewExplicitPriority(5), NewExplicitPriority(5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.p1.TakesPrecedenceOver(tt.p2))
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		expected FileFormat
		wantErr  bool
	}{
		{"test.toml", FormatTOML, false},
		{"test.yaml", FormatYAML, false},
		{"test.yml", FormatYAML, false},
		{"test.md", FormatMarkdown, false},
		{"test.markdown", FormatMarkdown, false},
		{"test.txt", FormatTOML, true}, // Should error on unsupported format
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			format, err := DetectFormat(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, format)
			}
		})
	}
}

func TestPriority_StringRepresentation(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{NewExplicitPriority(10), "explicit(10)"},
		{NewRelativePriority(5), "relative(5)"},
		{Priority{}, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.String())
		})
	}
}