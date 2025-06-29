# claude-merge-tool

A powerful configuration file merger designed specifically for creating CLAUDE.md files by combining base templates with language-specific configurations.

## Overview

`claude-merge-tool` helps you maintain consistent development guidelines across multiple programming languages by merging common guidelines with language-specific details. It supports multiple file formats (TOML, YAML, Markdown) and uses a sophisticated priority-based merging system.

## Features

- **Multi-format support**: Merge TOML, YAML, and Markdown files
- **Priority-based merging**: Control which configurations take precedence
- **Placeholder replacement**: Automatically inject language-specific content into templates
- **Flexible merge strategies**: Replace, append, or prepend content
- **Validation support**: Ensure configurations are valid before merging
- **Debug mode**: Trace the merging process for troubleshooting

## Installation

```bash
go install github.com/arustydev/claude-merge/cmd/claude-merge@latest
```

Or clone and build from source:

```bash
git clone https://github.com/arustydev/claude-merge-tool.git
cd claude-merge-tool
go build ./cmd/claude-merge
```

## Usage

### Basic Usage

Merge multiple configuration files:

```bash
claude-merge -files common.md,golang.md
```

This will create `CLAUDE.merged.md` by default.

### Command Line Options

```
-files string    Comma-separated paths to configuration files (required)
-output string   Output filename (default: CLAUDE.merged.md)
-order string    Comma-separated file order for merging (optional)
-validate        Validate only, don't generate output
-debug          Enable debug output
-help           Show help message
```

### Examples

#### Merge with custom output file
```bash
claude-merge -files base.toml,python.yaml -output CLAUDE.python.md
```

#### Specify merge order
```bash
claude-merge -files common.md,go.md,team.md -order common.md,team.md,go.md
```

#### Validate configurations without generating output
```bash
claude-merge -files config1.yaml,config2.yaml -validate
```

#### Debug mode to trace merging
```bash
claude-merge -files common.md,rust.md -debug
```

## Configuration Formats

### Markdown with Frontmatter

```markdown
---
title: "Python Development Guidelines"
priority:
  type: "explicit"
  value: 10
---

# Python Guidelines

Your markdown content here...
```

### TOML Configuration

```toml
[metadata]
title = "Go Development Guidelines"
language = "go"
priority = { type = "relative", value = 5 }

[sections.testing]
content = """
### Testing
Run tests with `go test ./...`
"""
priority = { type = "explicit", value = 10 }
```

### YAML Configuration

```yaml
metadata:
  title: "JavaScript Guidelines"
  language: "javascript"
  priority:
    type: "explicit"
    value: 8

sections:
  linting:
    content: |
      ### Linting
      Use ESLint with the team configuration
    priority:
      type: "relative"
      value: 5
```

## Placeholder System

The tool supports automatic placeholder replacement for language-specific content. This is particularly useful for maintaining a common template with language-specific sections.

### Example Template (common.md)

```markdown
# Development Guidelines

## Testing

Run the following commands for testing:
<language-specific-test-commands-here>
</language-specific-test-commands-here>

## Documentation

Follow these documentation standards:
<language-specific-documentation-standards>
</language-specific-documentation-standards>
```

### Language File (golang.md)

```markdown
# Go Specific Guidelines

### Testing commands

- Unit tests: `go test ./...`
- Race condition tests: `go test -race ./...`
- Coverage: `go test -cover ./...`

### Documentation Standards:

```go
// Package example provides...
//
// This package implements...
package example

// Function documents what it does
func Function() error {
    return nil
}
```
```

When merged, the placeholders in `common.md` will be replaced with the appropriate content from `golang.md`.

## Priority System

The tool uses a three-tier priority system:

1. **Explicit Priority**: Highest precedence, always wins
2. **Relative Priority**: Middle precedence, prevents override by lower priorities  
3. **File Order**: Lowest precedence, later files override earlier ones

### Setting Priorities

In YAML/TOML:
```yaml
priority:
  type: "explicit"  # or "relative"
  value: 10         # higher values take precedence
```

## Merge Strategies

When using merge targets and merge points, you can specify different strategies:

- **replace**: Replace the entire content (default)
- **append**: Add content after existing content
- **prepend**: Add content before existing content

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
```

### Project Structure

```
claude-merge-tool/
├── cmd/claude-merge/      # CLI application
├── internal/
│   ├── config/           # Configuration parsing and types
│   ├── merger/           # Merging logic and strategies
│   └── generator/        # Output generation
├── examples/data/        # Example configuration files
└── test/e2e/            # End-to-end tests
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.