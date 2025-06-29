package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arustydev/claude-merge/internal/config"
	"github.com/arustydev/claude-merge/internal/generator"
	"github.com/arustydev/claude-merge/internal/merger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_CommonPlusGolangMerge(t *testing.T) {
	// Load COMMON.1.md
	commonPath := filepath.Join("..", "..", "examples", "data", "COMMON.1.md")
	commonConfig, err := config.LoadConfig(commonPath)
	require.NoError(t, err, "Failed to load COMMON.1.md")

	// Load GOLANG.1.md
	golangPath := filepath.Join("..", "..", "examples", "data", "GOLANG.1.md")
	golangConfig, err := config.LoadConfig(golangPath)
	require.NoError(t, err, "Failed to load GOLANG.1.md")

	// Merge configurations
	m := merger.NewPriorityMerger(false)
	merged, err := m.MergeAll([]*config.Config{commonConfig, golangConfig})
	require.NoError(t, err, "Failed to merge configurations")

	// Generate markdown
	output := generator.GenerateMarkdown(merged)

	// Verify key aspects of the merge
	// 1. Title should be from COMMON.1.md
	assert.Contains(t, output, "Claude General Development Guidelines")

	// 2. Structure from COMMON.1.md should be preserved
	assert.Contains(t, output, "## IMPORTANT!!")
	assert.Contains(t, output, "## Quick Reference")
	assert.Contains(t, output, "### Using the Justfile")
	assert.Contains(t, output, "## STRICT REQUIREMENTS")
	assert.Contains(t, output, "### 1. Test Driven Development (TDD)")

	// 3. Placeholders should be replaced with Go-specific content
	// Test commands should be inserted
	assert.Contains(t, output, "- Unit tests: `go test ./...`")
	assert.Contains(t, output, "- Race condition tests: `go test -race ./...`")
	assert.Contains(t, output, "- End-to-end tests: `just test-e2e`")
	assert.Contains(t, output, "- Integration tests: `go test -tags=integration ./...`")

	// Documentation standards should be inserted
	assert.Contains(t, output, "// Package mcp provides a Model Context Protocol server implementation.")
	assert.Contains(t, output, "// Server represents an MCP server instance.")
	assert.Contains(t, output, "server := mcp.NewServer()")

	// 4. Placeholders themselves should be removed
	assert.NotContains(t, output, "<language-specific-test-commands-here>")
	assert.NotContains(t, output, "</language-specific-test-commands-here>")
	assert.NotContains(t, output, "<language-specific-documentation-standards>")
	assert.NotContains(t, output, "</language-specific-documentation-standards>")

	// 5. Go-specific sections should NOT override common sections
	assert.NotContains(t, output, "Claude Development Guidelines for Go Projects") // This is from GOLANG.1.md

	// 6. Content order should be preserved
	importantIndex := strings.Index(output, "## IMPORTANT!!")
	quickRefIndex := strings.Index(output, "## Quick Reference")
	strictReqIndex := strings.Index(output, "## STRICT REQUIREMENTS")
	assert.Less(t, importantIndex, quickRefIndex, "IMPORTANT!! should come before Quick Reference")
	assert.Less(t, quickRefIndex, strictReqIndex, "Quick Reference should come before STRICT REQUIREMENTS")
}

func TestE2E_PlaceholderReplacementWithMultipleLanguages(t *testing.T) {
	// Create test files with placeholders
	tempDir := t.TempDir()

	// Create a common file with placeholders
	commonContent := `# Common Guidelines

## Development Standards

Follow these practices:
<language-specific-test-commands-here>
</language-specific-test-commands-here>

## Documentation

Write good docs:
<language-specific-documentation-standards>
</language-specific-documentation-standards>
`
	commonPath := filepath.Join(tempDir, "common.md")
	err := os.WriteFile(commonPath, []byte(commonContent), 0644)
	require.NoError(t, err)

	// Create a language-specific file
	langContent := `# Python Guidelines

## Development Standards

### Testing commands

- Unit tests: ` + "`pytest tests/`" + `
- Coverage: ` + "`pytest --cov=src tests/`" + `
- Integration: ` + "`pytest tests/integration/`" + `

### Documentation Standards:

` + "```python" + `
def example_function(param: str) -> str:
    """
    Brief description of function.

    Args:
        param: Description of parameter

    Returns:
        Description of return value

    Examples:
        >>> example_function("test")
        "test result"
    """
    return f"{param} result"
` + "```" + `
`
	langPath := filepath.Join(tempDir, "python.md")
	err = os.WriteFile(langPath, []byte(langContent), 0644)
	require.NoError(t, err)

	// Load and merge
	commonConfig, err := config.LoadConfig(commonPath)
	require.NoError(t, err)

	langConfig, err := config.LoadConfig(langPath)
	require.NoError(t, err)

	m := merger.NewPriorityMerger(false)
	merged, err := m.MergeAll([]*config.Config{commonConfig, langConfig})
	require.NoError(t, err)

	// Generate output
	output := generator.GenerateMarkdown(merged)

	// Verify placeholders were replaced
	assert.Contains(t, output, "- Unit tests: `pytest tests/`")
	assert.Contains(t, output, "- Coverage: `pytest --cov=src tests/`")
	assert.Contains(t, output, `def example_function(param: str) -> str:`)
	assert.Contains(t, output, `"""`)

	// Verify placeholders were removed
	assert.NotContains(t, output, "<language-specific-test-commands-here>")
	assert.NotContains(t, output, "</language-specific-test-commands-here>")
}

func TestE2E_NoPlaceholders(t *testing.T) {
	// Test merging files without placeholders
	tempDir := t.TempDir()

	// Create two files without placeholders
	file1Content := `# File 1

This is content from file 1.
`
	file1Path := filepath.Join(tempDir, "file1.md")
	err := os.WriteFile(file1Path, []byte(file1Content), 0644)
	require.NoError(t, err)

	file2Content := `# File 2

This is content from file 2.
`
	file2Path := filepath.Join(tempDir, "file2.md")
	err = os.WriteFile(file2Path, []byte(file2Content), 0644)
	require.NoError(t, err)

	// Load and merge
	config1, err := config.LoadConfig(file1Path)
	require.NoError(t, err)

	config2, err := config.LoadConfig(file2Path)
	require.NoError(t, err)

	m := merger.NewPriorityMerger(false)
	merged, err := m.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	// Generate output
	output := generator.GenerateMarkdown(merged)

	// In this case, file2 should override file1 (file order precedence)
	assert.Contains(t, output, "This is content from file 2")
	assert.NotContains(t, output, "This is content from file 1")
}

func TestE2E_ComplexMergeScenarios(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected []string
		notExpected []string
	}{
		{
			name: "multiple placeholders",
			files: map[string]string{
				"base.md": `# Base
<language-specific-test-commands-here>
</language-specific-test-commands-here>
<language-specific-documentation-standards>
</language-specific-documentation-standards>`,
				"lang.md": `# Language
### Testing commands
- test command 1
- test command 2

### Documentation Standards:
Doc standard content`,
			},
			expected: []string{
				"# Base",
				"- test command 1",
				"- test command 2",
				"Doc standard content",
			},
			notExpected: []string{
				"<language-specific-test-commands-here>",
				"# Language",
			},
		},
		{
			name: "nested placeholders",
			files: map[string]string{
				"base.md": `# Base
## Section
<language-specific-test-commands-here>
Nested content
</language-specific-test-commands-here>`,
				"lang.md": `# Lang
### Testing commands
- nested test`,
			},
			expected: []string{
				"# Base",
				"- nested test",
			},
			notExpected: []string{
				"Nested content", // This should be replaced
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configs := make([]*config.Config, 0, len(tt.files))

			// Create files and load configs
			for filename, content := range tt.files {
				path := filepath.Join(tempDir, filename)
				err := os.WriteFile(path, []byte(content), 0644)
				require.NoError(t, err)

				cfg, err := config.LoadConfig(path)
				require.NoError(t, err)
				configs = append(configs, cfg)
			}

			// Merge
			m := merger.NewPriorityMerger(false)
			merged, err := m.MergeAll(configs)
			require.NoError(t, err)

			// Generate output
			output := generator.GenerateMarkdown(merged)

			// Check expected content
			for _, expected := range tt.expected {
				assert.Contains(t, output, expected, "Should contain: %s", expected)
			}

			// Check not expected content
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected, "Should not contain: %s", notExpected)
			}
		})
	}
}