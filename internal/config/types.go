package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
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

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (pt *PriorityType) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "explicit":
		*pt = PriorityExplicit
	case "relative":
		*pt = PriorityRelative
	case "none", "":
		*pt = PriorityNone
	default:
		return fmt.Errorf("unknown priority type: %s", string(text))
	}
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface
func (pt PriorityType) MarshalText() ([]byte, error) {
	switch pt {
	case PriorityExplicit:
		return []byte("explicit"), nil
	case PriorityRelative:
		return []byte("relative"), nil
	default:
		return []byte("none"), nil
	}
}

// String implements the Stringer interface for Priority
func (p Priority) String() string {
	switch p.Type {
	case PriorityExplicit:
		return fmt.Sprintf("explicit(%d)", p.Value)
	case PriorityRelative:
		return fmt.Sprintf("relative(%d)", p.Value)
	default:
		return "none"
	}
}

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

// TakesPrecedenceOverOrEqual determines if this priority beats or equals another
// Used for file order precedence where later files should override earlier ones
func (p Priority) TakesPrecedenceOverOrEqual(other Priority) bool {
	// Explicit always beats relative or none
	if p.Type == PriorityExplicit && other.Type != PriorityExplicit {
		return true
	}
	if other.Type == PriorityExplicit && p.Type != PriorityExplicit {
		return false
	}

	// Same type, compare values (higher wins, equal is also acceptable for file order)
	if p.Type == other.Type {
		return p.Value >= other.Value
	}

	// Relative beats none, none equals none (for file order)
	if p.Type == PriorityRelative && other.Type == PriorityNone {
		return true
	}
	if p.Type == PriorityNone && other.Type == PriorityNone {
		return true // Equal priorities, file order wins
	}

	return false
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

	config.SourceFormat = format
	return &config, nil
}

// parseMarkdown handles markdown files with frontmatter and parses headers/lists
func parseMarkdown(data []byte) (Config, error) {
	var config Config
	content := string(data)

	// Check for frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			// Parse frontmatter as YAML into metadata
			frontmatter := strings.TrimSpace(parts[1])
			var metadata map[string]interface{}
			err := yaml.Unmarshal([]byte(frontmatter), &metadata)
			if err != nil {
				return config, fmt.Errorf("failed to parse frontmatter: %w", err)
			}

			// Extract metadata fields
			if title, ok := metadata["title"].(string); ok {
				config.Metadata.Title = title
			}
			if description, ok := metadata["description"].(string); ok {
				config.Metadata.Description = description
			}
			if version, ok := metadata["version"].(string); ok {
				config.Metadata.Version = version
			}
			if language, ok := metadata["language"].(string); ok {
				config.Metadata.Language = language
			}

			// Parse priority if present
			if priorityMap, ok := metadata["priority"].(map[string]interface{}); ok {
				if priorityType, typeOk := priorityMap["type"].(string); typeOk {
					if priorityValue, valueOk := priorityMap["value"].(int); valueOk {
						config.Metadata.Priority.Value = priorityValue
						switch strings.ToLower(priorityType) {
						case "explicit":
							config.Metadata.Priority.Type = PriorityExplicit
						case "relative":
							config.Metadata.Priority.Type = PriorityRelative
						}
					}
				}
			}

			// For markdown files, treat the entire content as a single section
			markdownContent := strings.TrimSpace(parts[2])
			if markdownContent != "" {
				if config.Sections == nil {
					config.Sections = make(map[string]Section)
				}
				config.Sections["content"] = Section{
					Order:   1,
					Content: markdownContent,
				}
			}
		}
	} else {
		// No frontmatter, treat entire content as a single section
		if config.Sections == nil {
			config.Sections = make(map[string]Section)
		}
		config.Sections["content"] = Section{
			Order:   1,
			Content: strings.TrimSpace(content),
		}
		// Set a default title if none exists
		if config.Metadata.Title == "" {
			config.Metadata.Title = "Untitled Document"
		}
	}

	return config, nil
}

// parseMarkdownSections parses markdown content into sections based on headers and lists
func parseMarkdownSections(content string) map[string]Section {
	sections := make(map[string]Section)
	lines := strings.Split(content, "\n")

	currentSection := ""
	currentContent := ""
	order := 1

	// Regex patterns for different markdown elements
	headerRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	unorderedListRegex := regexp.MustCompile(`^[\s]*[-*+]\s+(.+)$`)
	orderedListRegex := regexp.MustCompile(`^[\s]*(\d+)\.\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for headers
		if headerRegex.MatchString(line) {
			// Save previous section if exists
			if currentSection != "" && currentContent != "" {
				sections[currentSection] = Section{
					Order:   order,
					Content: strings.TrimSpace(currentContent),
				}
				order++
			}

			// Start new section
			matches := headerRegex.FindStringSubmatch(line)
			level := len(matches[1])
			title := matches[2]
			currentSection = fmt.Sprintf("header_%d_%s", level, sanitizeName(title))
			currentContent = line
			continue
		}

		// Check for unordered lists
		if unorderedListRegex.MatchString(line) {
			// Save previous section if exists
			if currentSection != "" && currentContent != "" {
				sections[currentSection] = Section{
					Order:   order,
					Content: strings.TrimSpace(currentContent),
				}
				order++
			}

			// Create list section
			sectionName := fmt.Sprintf("list_%d", order)
			sections[sectionName] = Section{
				Order:   order,
				Content: line,
			}
			order++
			currentSection = ""
			currentContent = ""
			continue
		}

		// Check for ordered lists
		if orderedListRegex.MatchString(line) {
			// Save previous section if exists
			if currentSection != "" && currentContent != "" {
				sections[currentSection] = Section{
					Order:   order,
					Content: strings.TrimSpace(currentContent),
				}
				order++
			}

			// Create ordered list section
			matches := orderedListRegex.FindStringSubmatch(line)
			listNum := matches[1]
			sectionName := fmt.Sprintf("ordered_list_%s", listNum)
			sections[sectionName] = Section{
				Order:   order,
				Content: line,
			}
			order++
			currentSection = ""
			currentContent = ""
			continue
		}

		// Regular content line
		if currentContent != "" {
			currentContent += "\n" + line
		} else if line != "" {
			// Start content section if no header found yet
			if currentSection == "" {
				currentSection = "content"
			}
			currentContent = line
		}
	}

	// Save final section
	if currentSection != "" && currentContent != "" {
		sections[currentSection] = Section{
			Order:   order,
			Content: strings.TrimSpace(currentContent),
		}
	}

	return sections
}

// sanitizeName converts a title to a valid section name
func sanitizeName(title string) string {
	// Convert to lowercase and replace spaces/special chars with underscores
	name := strings.ToLower(title)
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	return name
}