# Step-by-Step TDD Implementation Plan: claude-merge Tool

## Overview for Junior Developer
You will be building a Go command-line tool that reads configuration files in multiple formats (TOML, YAML, Markdown) and combines them using a priority-based merging system to create a single CLAUDE.md file. Think of it like a template system where you have base configuration files (common files) and specialized additions (language-specific, project-specific, etc.) that get intelligently merged together based on explicit priorities, relative priorities, and fallback ordering.

**Key Principles:**
- **TDD-First**: Write failing tests before any implementation code
- **Multi-format Support**: Handle TOML, YAML, and Markdown inputs seamlessly
- **Priority-based Merging**: Respect explicit priorities, relative priorities, and file order
- **Schema Consistency**: All formats use the same logical schema structure

## Phase 1: Initial Project Setup (Start Here!)

### Step 1: Create the Basic Directory Structure
**Intent**: We're organizing our code into standard Go project folders. This makes it easier for other Go developers to understand our project.

```bash
# Create the main directories
mkdir -p cmd/claude-merge
mkdir -p internal/config
mkdir -p internal/merger
mkdir -p internal/generator
mkdir -p pkg/claudemd
mkdir -p examples
mkdir -p testdata
```

### Step 2: Initialize the Go Module
**Intent**: This tells Go that our directory is a Go project and helps manage dependencies.

```bash
# Initialize the module
go mod init github.com/arustydev/claude-merge
```

### Step 3: Create the Main Entry Point
**Intent**: Every Go program needs a main function. This is where our program starts.

Create `cmd/claude-merge/main.go`:
```go
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("claude-merge tool - Starting development")
    // We'll add the actual logic here later
    os.Exit(0)
}
```

### Step 4: Test Your Setup
**Intent**: Make sure everything is working before we continue.

```bash
# Build and run the program
go run cmd/claude-merge/main.go

# You should see: "claude-merge tool - Starting development"
```

## Phase 2: Test-Driven Data Structure Design

### Step 5: Write Tests for Type Definitions (TDD)
**Intent**: Define our data structures through tests first, ensuring they work correctly across all supported formats.

Create `internal/config/types_test.go`:
```go
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
priority = 10

[sections.header]
order = 1
priority = 5
content = "# Header"
`
    
    config, err := ParseConfig([]byte(tomlContent), FormatTOML)
    require.NoError(t, err)
    
    assert.Equal(t, "Test Config", config.Metadata.Title)
    assert.Equal(t, 10, config.Metadata.Priority)
    assert.Equal(t, "# Header", config.Sections["header"].Content)
}

func TestConfig_UnmarshalYAML(t *testing.T) {
    yamlContent := `
metadata:
  title: "Test Config"
  priority: 10
sections:
  header:
    order: 1
    priority: 5
    content: "# Header"
`
    
    config, err := ParseConfig([]byte(yamlContent), FormatYAML)
    require.NoError(t, err)
    
    assert.Equal(t, "Test Config", config.Metadata.Title)
    assert.Equal(t, 10, config.Metadata.Priority)
}

func TestConfig_ParseMarkdown(t *testing.T) {
    mdContent := `---
title: "Test Config"
priority: 10
---

# Header Content

This is markdown content that should be treated as a section.
`
    
    config, err := ParseConfig([]byte(mdContent), FormatMarkdown)
    require.NoError(t, err)
    
    assert.Equal(t, "Test Config", config.Metadata.Title)
    assert.Contains(t, config.Sections, "content")
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
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.expected, tt.p1.TakesPrecedenceOver(tt.p2))
        })
    }
}
```

### Step 6: Implement Type Definitions
**Intent**: Now implement the data structures to make our tests pass, supporting all formats with priority-based merging.

Create `internal/config/types.go`:
```go
package config

import (
    "fmt"
    "strings"
)

// FileFormat represents supported input formats
type FileFormat int

const (
    FormatTOML FileFormat = iota
    FormatYAML
    FormatMarkdown
)

// Config represents the entire configuration file
// Works across TOML, YAML, and Markdown formats
type Config struct {
    Metadata     Metadata                `toml:"metadata" yaml:"metadata"`
    Sections     map[string]Section      `toml:"sections" yaml:"sections"`
    MergePoints  map[string]MergePoint   `toml:"merge_points" yaml:"merge_points"`
    MergeTargets map[string]MergeTarget  `toml:"merge_targets" yaml:"merge_targets"`
    SourceFile   string                  // Track which file this came from
    SourceFormat FileFormat              // Track the original format
}

// Metadata contains information about the configuration
type Metadata struct {
    Title       string   `toml:"title" yaml:"title"`
    Description string   `toml:"description" yaml:"description"`
    Version     string   `toml:"version" yaml:"version"`
    Language    string   `toml:"language" yaml:"language"`
    Extends     string   `toml:"extends" yaml:"extends"`
    Priority    Priority `toml:"priority" yaml:"priority"`
}

// Section represents a piece of content in the final document
type Section struct {
    Order       int      `toml:"order" yaml:"order"`
    Parent      string   `toml:"parent" yaml:"parent"`
    MergeID     string   `toml:"merge_id" yaml:"merge_id"`
    Content     string   `toml:"content" yaml:"content"`
    MergePoints []string `toml:"merge_points" yaml:"merge_points"`
    Priority    Priority `toml:"priority" yaml:"priority"`
}

// MergePoint defines a place where content can be inserted
type MergePoint struct {
    Placeholder string   `toml:"placeholder" yaml:"placeholder"`
    Default     string   `toml:"default" yaml:"default"`
    Priority    Priority `toml:"priority" yaml:"priority"`
}

// MergeTarget is content that fills a merge point
type MergeTarget struct {
    Strategy string   `toml:"strategy" yaml:"strategy"`
    Content  string   `toml:"content" yaml:"content"`
    Priority Priority `toml:"priority" yaml:"priority"`
}

// Priority represents merge priority with explicit > relative > order-based
type Priority struct {
    Type  PriorityType `toml:"type" yaml:"type"`
    Value int          `toml:"value" yaml:"value"`
}

type PriorityType int

const (
    PriorityNone PriorityType = iota
    PriorityRelative
    PriorityExplicit
)

// NewExplicitPriority creates an explicit priority
func NewExplicitPriority(value int) Priority {
    return Priority{Type: PriorityExplicit, Value: value}
}

// NewRelativePriority creates a relative priority
func NewRelativePriority(value int) Priority {
    return Priority{Type: PriorityRelative, Value: value}
}

// TakesPrecedenceOver determines if this priority beats another
func (p Priority) TakesPrecedenceOver(other Priority) bool {
    // Explicit always beats relative or none
    if p.Type == PriorityExplicit && other.Type != PriorityExplicit {
        return true
    }
    if other.Type == PriorityExplicit && p.Type != PriorityExplicit {
        return false
    }
    
    // Same type, compare values (higher wins)
    if p.Type == other.Type {
        return p.Value > other.Value
    }
    
    // Relative beats none
    return p.Type == PriorityRelative && other.Type == PriorityNone
}

// DetectFormat determines file format from extension
func DetectFormat(filename string) (FileFormat, error) {
    lower := strings.ToLower(filename)
    switch {
    case strings.HasSuffix(lower, ".toml"):
        return FormatTOML, nil
    case strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml"):
        return FormatYAML, nil
    case strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown"):
        return FormatMarkdown, nil
    default:
        return FormatTOML, fmt.Errorf("unsupported file format for %s", filename)
    }
}
```

### Step 7: Write Tests for Merge Strategies (TDD)
**Intent**: Define merge behavior through tests first.

Create `internal/merger/strategies_test.go`:
```go
package merger

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMergeStrategy_IsValid(t *testing.T) {
    tests := []struct {
        strategy MergeStrategy
        valid    bool
    }{
        {StrategyReplace, true},
        {StrategyAppend, true},
        {StrategyPrepend, true},
        {"invalid", false},
        {"", false},
    }
    
    for _, tt := range tests {
        t.Run(string(tt.strategy), func(t *testing.T) {
            assert.Equal(t, tt.valid, tt.strategy.IsValid())
        })
    }
}

func TestApplyStrategy(t *testing.T) {
    tests := []struct {
        name     string
        strategy MergeStrategy
        old      string
        new      string
        expected string
    }{
        {"replace", StrategyReplace, "old content", "new content", "new content"},
        {"append", StrategyAppend, "old content", "new content", "old content\nnew content"},
        {"prepend", StrategyPrepend, "old content", "new content", "new content\nold content"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ApplyStrategy(tt.strategy, tt.old, tt.new)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Step 8: Implement Merge Strategies
**Intent**: Now implement the strategies to make our tests pass.

Create `internal/merger/strategies.go`:
```go
package merger

// MergeStrategy defines how content should be combined
type MergeStrategy string

const (
    // StrategyReplace means the new content completely replaces the old
    StrategyReplace MergeStrategy = "replace"
    
    // StrategyAppend means the new content is added after the old
    StrategyAppend MergeStrategy = "append"
    
    // StrategyPrepend means the new content is added before the old
    StrategyPrepend MergeStrategy = "prepend"
)

// IsValid checks if a strategy string is valid
func (s MergeStrategy) IsValid() bool {
    switch s {
    case StrategyReplace, StrategyAppend, StrategyPrepend:
        return true
    default:
        return false
    }
}

// ApplyStrategy applies a merge strategy to combine old and new content
func ApplyStrategy(strategy MergeStrategy, old, new string) string {
    switch strategy {
    case StrategyReplace:
        return new
    case StrategyAppend:
        if old == "" {
            return new
        }
        return old + "\n" + new
    case StrategyPrepend:
        if old == "" {
            return new
        }
        return new + "\n" + old
    default:
        return new // Default to replace
    }
}
```

## Phase 3: Multi-Format Configuration Loading (TDD)

### Step 9: Write Tests for Configuration Loading
**Intent**: Define loading behavior for all formats through tests first.

First, add required dependencies:
```bash
go get github.com/BurntSushi/toml
go get gopkg.in/yaml.v3
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/require
```

Create `internal/config/loader_test.go`:
```go
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

// Helper function for testing
func testLoadFromContent(t *testing.T, content, filename string) (*Config, error) {
    tmpfile, err := os.CreateTemp("", filename)
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
```

### Step 10: Implement Multi-Format Configuration Loading
**Intent**: Now implement the loader to make our tests pass.

Create `internal/config/loader.go`:
```go
package config

import (
    "fmt"
    "os"
    "strings"
    "bufio"
    "bytes"
    
    "github.com/BurntSushi/toml"
    "gopkg.in/yaml.v3"
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

// ParseConfig parses configuration data based on format
func ParseConfig(data []byte, format FileFormat) (*Config, error) {
    var config Config
    
    switch format {
    case FormatTOML:
        err := toml.Unmarshal(data, &config)
        if err != nil {
            return nil, fmt.Errorf("TOML parse error: %w", err)
        }
        
    case FormatYAML:
        err := yaml.Unmarshal(data, &config)
        if err != nil {
            return nil, fmt.Errorf("YAML parse error: %w", err)
        }
        
    case FormatMarkdown:
        var err error
        config, err = parseMarkdown(data)
        if err != nil {
            return nil, fmt.Errorf("Markdown parse error: %w", err)
        }
        
    default:
        return nil, fmt.Errorf("unsupported format: %v", format)
    }
    
    // Initialize maps if nil
    if config.Sections == nil {
        config.Sections = make(map[string]Section)
    }
    if config.MergePoints == nil {
        config.MergePoints = make(map[string]MergePoint)
    }
    if config.MergeTargets == nil {
        config.MergeTargets = make(map[string]MergeTarget)
    }
    
    return &config, nil
}

// parseMarkdown handles markdown files with frontmatter
func parseMarkdown(data []byte) (Config, error) {
    var config Config
    
    content := string(data)
    
    // Check for frontmatter
    if strings.HasPrefix(content, "---") {
        parts := strings.SplitN(content, "---", 3)
        if len(parts) >= 3 {
            // Parse frontmatter as YAML
            frontmatter := parts[1]
            err := yaml.Unmarshal([]byte(frontmatter), &config)
            if err != nil {
                return config, fmt.Errorf("failed to parse frontmatter: %w", err)
            }
            
            // Rest is markdown content
            markdownContent := strings.TrimSpace(parts[2])
            if markdownContent != "" {
                config.Sections = map[string]Section{
                    "content": {
                        Order:   1,
                        Content: markdownContent,
                    },
                }
            }
        }
    } else {
        // No frontmatter, treat entire content as a section
        config.Sections = map[string]Section{
            "content": {
                Order:   1,
                Content: strings.TrimSpace(content),
            },
        }
    }
    
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
```

## Phase 4: Command-Line Interface with Priority Support

### Step 11: Write Tests for CLI (TDD)
**Intent**: Define CLI behavior through tests first.

Create `cmd/claude-merge/main_test.go`:
```go
package main

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestParseFiles_DefaultOrder(t *testing.T) {
    files := []string{"file1.toml", "file2.yaml", "file3.md"}
    order := parseFileOrder(files, "")
    
    expected := []string{"file1.toml", "file2.yaml", "file3.md"}
    assert.Equal(t, expected, order)
}

func TestParseFiles_CustomOrder(t *testing.T) {
    files := []string{"file1.toml", "file2.yaml", "file3.md"}
    order := parseFileOrder(files, "file3.md,file1.toml,file2.yaml")
    
    expected := []string{"file3.md", "file1.toml", "file2.yaml"}
    assert.Equal(t, expected, order)
}

func TestValidateArgs(t *testing.T) {
    tests := []struct {
        name    string
        files   []string
        wantErr bool
    }{
        {"valid files", []string{"test.toml"}, false},
        {"no files", []string{}, true},
        {"empty string", []string{""}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateArgs(tt.files, "CLAUDE.md")
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Step 12: Implement Enhanced CLI
**Intent**: Now implement the CLI to support multiple formats and merge ordering.

Update `cmd/claude-merge/main.go`:
```go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "strings"
    
    "github.com/arustydev/claude-merge/internal/config"
    "github.com/arustydev/claude-merge/internal/merger"
    "github.com/arustydev/claude-merge/internal/generator"
)

func main() {
    // Define command-line flags
    var (
        files      = flag.String("files", "", "Comma-separated paths to configuration files (required)")
        outputFile = flag.String("output", "CLAUDE.md", "Output filename (default: CLAUDE.md)")
        mergeOrder = flag.String("order", "", "Comma-separated file order for merging (optional)")
        validate   = flag.Bool("validate", false, "Validate only, don't generate output")
        debug      = flag.Bool("debug", false, "Enable debug output")
        help       = flag.Bool("help", false, "Show help message")
    )
    
    // Parse the flags
    flag.Parse()
    
    if *help {
        printHelp()
        return
    }
    
    // Split input files
    inputFiles := strings.Split(*files, ",")
    for i := range inputFiles {
        inputFiles[i] = strings.TrimSpace(inputFiles[i])
    }
    
    // Validate arguments
    err := validateArgs(inputFiles, *outputFile)
    if err != nil {
        log.Fatalf("Invalid arguments: %v", err)
    }
    
    // Determine merge order
    fileOrder := parseFileOrder(inputFiles, *mergeOrder)
    
    if *debug {
        fmt.Printf("Input files: %v\n", inputFiles)
        fmt.Printf("Merge order: %v\n", fileOrder)
        fmt.Printf("Output file: %s\n", *outputFile)
    }
    
    // Load all configurations
    configs := make([]*config.Config, 0, len(fileOrder))
    for _, filename := range fileOrder {
        cfg, err := config.LoadConfig(filename)
        if err != nil {
            log.Fatalf("Failed to load config %s: %v", filename, err)
        }
        
        if *validate {
            err = config.ValidateConfig(cfg)
            if err != nil {
                log.Fatalf("Invalid config %s: %v", filename, err)
            }
            fmt.Printf("✓ %s validated successfully\n", filename)
        }
        
        configs = append(configs, cfg)
        
        if *debug {
            fmt.Printf("✓ Loaded %s (%s format)\n", filename, formatName(cfg.SourceFormat))
        }
    }
    
    if *validate {
        fmt.Println("✓ All configurations validated successfully")
        return
    }
    
    // Merge configurations using priority-based merging
    m := merger.NewPriorityMerger(*debug)
    merged, err := m.MergeAll(configs)
    if err != nil {
        log.Fatalf("Failed to merge configurations: %v", err)
    }
    
    // Generate markdown
    markdown := generator.GenerateMarkdown(merged)
    
    // Write output
    err = os.WriteFile(*outputFile, []byte(markdown), 0644)
    if err != nil {
        log.Fatalf("Failed to write output: %v", err)
    }
    
    fmt.Printf("✓ Generated %s successfully\n", *outputFile)
}

// validateArgs validates command-line arguments
func validateArgs(files []string, output string) error {
    if len(files) == 0 || (len(files) == 1 && files[0] == "") {
        return fmt.Errorf("no input files specified")
    }
    
    if output == "" {
        return fmt.Errorf("output filename cannot be empty")
    }
    
    // Check if files exist
    for _, file := range files {
        if file == "" {
            continue
        }
        if _, err := os.Stat(file); os.IsNotExist(err) {
            return fmt.Errorf("file not found: %s", file)
        }
    }
    
    return nil
}

// parseFileOrder determines the order of files for merging
func parseFileOrder(files []string, orderSpec string) []string {
    if orderSpec == "" {
        return files // Default order
    }
    
    // Parse custom order
    customOrder := strings.Split(orderSpec, ",")
    for i := range customOrder {
        customOrder[i] = strings.TrimSpace(customOrder[i])
    }
    
    // Validate that all files in custom order exist in input files
    fileSet := make(map[string]bool)
    for _, file := range files {
        fileSet[file] = true
    }
    
    result := make([]string, 0, len(files))
    used := make(map[string]bool)
    
    // Add files in custom order first
    for _, file := range customOrder {
        if fileSet[file] && !used[file] {
            result = append(result, file)
            used[file] = true
        }
    }
    
    // Add any remaining files
    for _, file := range files {
        if !used[file] {
            result = append(result, file)
        }
    }
    
    return result
}

// formatName returns a readable format name
func formatName(format config.FileFormat) string {
    switch format {
    case config.FormatTOML:
        return "TOML"
    case config.FormatYAML:
        return "YAML"
    case config.FormatMarkdown:
        return "Markdown"
    default:
        return "Unknown"
    }
}

// printHelp displays usage information
func printHelp() {
    fmt.Println("claude-merge - Configuration file merger for CLAUDE.md generation")
    fmt.Println()
    fmt.Println("Usage:")
    fmt.Println("  claude-merge -files file1.toml,file2.yaml,file3.md [options]")
    fmt.Println()
    fmt.Println("Options:")
    fmt.Println("  -files string    Comma-separated paths to configuration files (required)")
    fmt.Println("  -output string   Output filename (default: CLAUDE.md)")
    fmt.Println("  -order string    Comma-separated file order for merging (optional)")
    fmt.Println("  -validate        Validate only, don't generate output")
    fmt.Println("  -debug          Enable debug output")
    fmt.Println("  -help           Show this help message")
    fmt.Println()
    fmt.Println("Supported formats: TOML (.toml), YAML (.yaml, .yml), Markdown (.md)")
    fmt.Println()
    fmt.Println("Priority-based merging:")
    fmt.Println("  1. Explicit priority values override everything else")
    fmt.Println("  2. Relative priority values prevent override by lower priorities")
    fmt.Println("  3. File order determines precedence when no priorities set")
}
```

## Phase 5: Create Example Files

### Step 9: Create Example TOML Files
**Intent**: We need example files to test our tool. These also serve as documentation for users.

Create `examples/common.toml`:
```toml
[metadata]
title = "Claude General Development Guidelines"
description = "This document contains critical guidelines and requirements"
version = "1.0.0"

[sections.header]
order = 1
content = """
# Claude Development Guidelines

This document contains critical guidelines for development projects.
"""

[sections.requirements]
order = 2
content = """
## STRICT REQUIREMENTS

### 1. Test Driven Development (TDD)
Follow strict TDD practices with RED-GREEN-REFACTOR pattern.

### 2. 100% Test Passing Rate
ALL tests must pass before any commit:
<!-- MERGE:test-commands -->
"""

[merge_points.test-commands]
placeholder = "<!-- MERGE:test-commands -->"
default = "Run all tests before committing"
```

Create `examples/golang.toml`:
```toml
[metadata]
language = "golang"
extends = "common.toml"

[merge_targets.test-commands]
strategy = "replace"
content = """
```bash
go test ./...
go test -race ./...
```
"""

[sections.golang-specific]
order = 10
content = """
## Go-Specific Guidelines

### Error Handling
Always wrap errors with context:
```go
if err != nil {
    return fmt.Errorf("failed to process: %w", err)
}
```
"""
```

## Phase 5: Priority-Based Merge Logic (TDD)

### Step 13: Write Tests for Priority Merger
**Intent**: Define the complex priority-based merging logic through tests first.

Create `internal/merger/priority_merger_test.go`:
```go
package merger

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/arustydev/claude-merge/internal/config"
)

func TestPriorityMerger_MergeAll(t *testing.T) {
    // Test explicit priority overrides
    config1 := &config.Config{
        Metadata: config.Metadata{
            Title: "Base Config",
            Priority: config.NewExplicitPriority(5),
        },
        Sections: map[string]config.Section{
            "section1": {
                Content: "Base content",
                Priority: config.NewExplicitPriority(5),
            },
        },
    }
    
    config2 := &config.Config{
        Metadata: config.Metadata{
            Title: "Override Config", 
            Priority: config.NewExplicitPriority(10),
        },
        Sections: map[string]config.Section{
            "section1": {
                Content: "Override content",
                Priority: config.NewExplicitPriority(10),
            },
        },
    }
    
    merger := NewPriorityMerger(false)
    result, err := merger.MergeAll([]*config.Config{config1, config2})
    require.NoError(t, err)
    
    assert.Equal(t, "Override Config", result.Metadata.Title)
    assert.Equal(t, "Override content", result.Sections["section1"].Content)
}

func TestPriorityMerger_RelativePriority(t *testing.T) {
    config1 := &config.Config{
        Sections: map[string]config.Section{
            "section1": {
                Content: "Relative content",
                Priority: config.NewRelativePriority(10),
            },
        },
    }
    
    config2 := &config.Config{
        Sections: map[string]config.Section{
            "section1": {
                Content: "Should not override",
                Priority: config.NewRelativePriority(5),
            },
        },
    }
    
    merger := NewPriorityMerger(false)
    result, err := merger.MergeAll([]*config.Config{config1, config2})
    require.NoError(t, err)
    
    assert.Equal(t, "Relative content", result.Sections["section1"].Content)
}

func TestPriorityMerger_FileOrder(t *testing.T) {
    config1 := &config.Config{
        Sections: map[string]config.Section{
            "section1": {Content: "First content"},
        },
    }
    
    config2 := &config.Config{
        Sections: map[string]config.Section{
            "section1": {Content: "Second content"},
        },
    }
    
    merger := NewPriorityMerger(false)
    result, err := merger.MergeAll([]*config.Config{config1, config2})
    require.NoError(t, err)
    
    // Second config should win with file order precedence
    assert.Equal(t, "Second content", result.Sections["section1"].Content)
}
```

### Step 14: Implement Priority-Based Merger
**Intent**: Implement the complex merging logic that respects all priority types.

Create `internal/merger/priority_merger.go`:
```go
package merger

import (
    "fmt"
    "github.com/arustydev/claude-merge/internal/config"
)

// PriorityMerger handles priority-based merging of multiple configurations
type PriorityMerger struct {
    debug bool
}

// NewPriorityMerger creates a new priority merger
func NewPriorityMerger(debug bool) *PriorityMerger {
    return &PriorityMerger{debug: debug}
}

// MergeAll merges multiple configurations using priority rules
func (m *PriorityMerger) MergeAll(configs []*config.Config) (*config.Config, error) {
    if len(configs) == 0 {
        return nil, fmt.Errorf("no configurations to merge")
    }
    
    result := &config.Config{
        Sections:     make(map[string]config.Section),
        MergePoints:  make(map[string]config.MergePoint),
        MergeTargets: make(map[string]config.MergeTarget),
    }
    
    // Merge in phases: explicit -> relative -> file order
    for _, cfg := range configs {
        m.mergeMetadata(result, cfg)
        m.mergeSections(result, cfg)
        m.mergeMergePoints(result, cfg)
        m.mergeMergeTargets(result, cfg)
    }
    
    return result, nil
}

// mergeMetadata merges metadata using priority rules
func (m *PriorityMerger) mergeMetadata(result *config.Config, incoming *config.Config) {
    if result.Metadata.Title == "" || incoming.Metadata.Priority.TakesPrecedenceOver(result.Metadata.Priority) {
        if incoming.Metadata.Title != "" {
            result.Metadata.Title = incoming.Metadata.Title
            result.Metadata.Priority = incoming.Metadata.Priority
        }
    }
    
    // Apply same logic to other metadata fields
    if result.Metadata.Description == "" || incoming.Metadata.Priority.TakesPrecedenceOver(result.Metadata.Priority) {
        if incoming.Metadata.Description != "" {
            result.Metadata.Description = incoming.Metadata.Description
        }
    }
}

// mergeSections merges sections using priority rules
func (m *PriorityMerger) mergeSections(result *config.Config, incoming *config.Config) {
    for name, section := range incoming.Sections {
        existing, exists := result.Sections[name]
        
        if !exists || section.Priority.TakesPrecedenceOver(existing.Priority) {
            if m.debug {
                fmt.Printf("Merging section %s from %s\n", name, incoming.SourceFile)
            }
            result.Sections[name] = section
        } else if m.debug {
            fmt.Printf("Skipping section %s (lower priority)\n", name)
        }
    }
}

// mergeMergePoints and mergeMergeTargets follow similar patterns
func (m *PriorityMerger) mergeMergePoints(result *config.Config, incoming *config.Config) {
    for name, point := range incoming.MergePoints {
        existing, exists := result.MergePoints[name]
        if !exists || point.Priority.TakesPrecedenceOver(existing.Priority) {
            result.MergePoints[name] = point
        }
    }
}

func (m *PriorityMerger) mergeMergeTargets(result *config.Config, incoming *config.Config) {
    for name, target := range incoming.MergeTargets {
        existing, exists := result.MergeTargets[name]
        if !exists || target.Priority.TakesPrecedenceOver(existing.Priority) {
            result.MergeTargets[name] = target
        }
    }
}
```

## Phase 6: Updated Example Files

### Step 15: Create Multi-Format Example Files
**Intent**: Provide examples that demonstrate the new capabilities.

Create `examples/common.toml`:
```toml
[metadata]
title = "Claude Development Guidelines"
description = "Multi-format configuration system"
version = "2.0.0"
priority = { type = "relative", value = 1 }

[sections.header]
order = 1
content = """# Claude Development Guidelines
Base guidelines that apply to all projects."""

[merge_points.test-commands]
placeholder = "<!-- MERGE:test-commands -->"
default = "Run basic tests"
```

Create `examples/golang.yaml`:
```yaml
metadata:
  language: "golang"
  extends: "common.toml"
  priority:
    type: "explicit"
    value: 10

sections:
  golang-specific:
    order: 10
    content: |
      ## Go-Specific Guidelines
      Use `go test ./...` for testing.
    priority:
      type: "explicit"
      value: 15

merge_targets:
  test-commands:
    strategy: "replace"
    content: "```bash\ngo test ./...\ngo test -race ./...\n```"
```

Create `examples/project-specific.md`:
```markdown
---
title: "Project Overrides"
priority:
  type: "explicit"
  value: 20
---

# Project-Specific Content

This content has the highest priority and will override others.
```

## Phase 7: TDD Markdown Generation

### Step 16: Write Tests for Markdown Generator
Create `internal/generator/markdown_test.go` then implement `internal/generator/markdown.go`.

## Phase 8: Integration and Testing

### Step 17: Build and Test Complete System
**Intent**: Verify everything works together with TDD approach.

```bash
# Install dependencies
go mod tidy

# Run all tests (should pass!)
go test ./...

# Build the tool
go build -o claude-merge cmd/claude-merge/main.go

# Test with multi-format examples
./claude-merge -files examples/common.toml,examples/golang.yaml,examples/project-specific.md -debug
```

## Key Changes from Original Plan

### 1. **TDD-First Approach**
- Write tests before implementation for all components
- RED-GREEN-REFACTOR cycle throughout development
- Comprehensive test coverage for priority logic

### 2. **Multi-Format Support**
- Auto-detection of TOML, YAML, and Markdown formats
- Consistent schema across all formats
- Markdown frontmatter support

### 3. **Priority-Based Merging**
- **Explicit Priority**: Highest precedence, numeric values
- **Relative Priority**: Prevents override by lower priorities
- **File Order**: Default fallback when no priorities set
- Complex merge logic that respects all priority types

### 4. **Enhanced CLI**
- `-files` flag for multiple input files
- `-order` flag for custom merge ordering
- Support for mixed file formats in single command
- Better validation and error handling

### 5. **Simplified Structure**
- Removed the distinction between "common" and "language" files
- All files are equal, differentiated only by priority and order
- More flexible configuration system

## Summary

The updated tool is now significantly more powerful:

✅ **Multi-format input** (TOML, YAML, Markdown)  
✅ **Priority-based merging** with three levels of precedence  
✅ **TDD development approach** ensuring robust code  
✅ **Flexible file ordering** with optional custom sequences  
✅ **Enhanced validation** and debugging capabilities  

The tool can now handle complex scenarios like:
- Base configuration files with common defaults
- Language-specific overrides with explicit priorities  
- Project-specific customizations with highest priority
- Mixed file formats in a single merge operation

## Next Steps for Enhancement

1. **Add more merge strategies** (like "merge" for combining lists)
2. **Support multiple language files** in one output
3. **Add template support** for custom output formats
4. **Create more comprehensive tests**
5. **Add a `convert` command** to convert existing .md files to .toml
6. **Implement section hierarchy** (parent/child relationships)
7. **Add watch mode** for development

## Common Issues and Solutions

1. **"module not found" error**: Make sure you run `go mod init` and update import paths
2. **"undefined" errors**: Check that package names match directory names
3. **Empty output**: Enable debug mode (-debug) to see what's happening
4. **Merge not working**: Check that placeholder names match exactly

## Summary

You've built a tool that:
1. Reads TOML configuration files
2. Merges language-specific content into common templates
3. Generates unified CLAUDE.md files
4. Supports different merge strategies
5. Can validate configurations

The tool is extensible and can grow with your needs!