package merger

import (
	"fmt"
	"strings"

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

	// Check if we have a base template with placeholders
	baseConfig := m.findBaseTemplate(configs)
	if baseConfig != nil {
		// Use template-based merging
		m.mergeWithTemplate(result, baseConfig, configs)
	} else {
		// Use standard priority-based merging
		for _, cfg := range configs {
			m.mergeMetadata(result, cfg)
			m.mergeSections(result, cfg)
			m.mergeMergePoints(result, cfg)
			m.mergeMergeTargets(result, cfg)
		}
	}

	return result, nil
}

// mergeMetadata merges metadata using priority rules
func (m *PriorityMerger) mergeMetadata(result *config.Config, incoming *config.Config) {
	// Merge title
	if result.Metadata.Title == "" || incoming.Metadata.Priority.TakesPrecedenceOverOrEqual(result.Metadata.Priority) {
		if incoming.Metadata.Title != "" {
			result.Metadata.Title = incoming.Metadata.Title
			result.Metadata.Priority = incoming.Metadata.Priority
		}
	}

	// Merge description (only if higher priority or empty)
	if result.Metadata.Description == "" || incoming.Metadata.Priority.TakesPrecedenceOverOrEqual(result.Metadata.Priority) {
		if incoming.Metadata.Description != "" {
			result.Metadata.Description = incoming.Metadata.Description
		}
	}

	// Merge version (only if higher priority or empty)
	if result.Metadata.Version == "" || incoming.Metadata.Priority.TakesPrecedenceOverOrEqual(result.Metadata.Priority) {
		if incoming.Metadata.Version != "" {
			result.Metadata.Version = incoming.Metadata.Version
		}
	}

	// Merge language (only if higher priority or empty)
	if result.Metadata.Language == "" || incoming.Metadata.Priority.TakesPrecedenceOverOrEqual(result.Metadata.Priority) {
		if incoming.Metadata.Language != "" {
			result.Metadata.Language = incoming.Metadata.Language
		}
	}

	// Merge extends (only if higher priority or empty)
	if result.Metadata.Extends == "" || incoming.Metadata.Priority.TakesPrecedenceOverOrEqual(result.Metadata.Priority) {
		if incoming.Metadata.Extends != "" {
			result.Metadata.Extends = incoming.Metadata.Extends
		}
	}
}

// mergeSections merges sections using priority rules
func (m *PriorityMerger) mergeSections(result *config.Config, incoming *config.Config) {
	for name, section := range incoming.Sections {
		existing, exists := result.Sections[name]

		if !exists || section.Priority.TakesPrecedenceOverOrEqual(existing.Priority) {
			if m.debug {
				fmt.Printf("Merging section %s from %s\n", name, incoming.SourceFile)
			}
			result.Sections[name] = section
		} else if m.debug {
			fmt.Printf("Skipping section %s (lower priority)\n", name)
		}
	}
}

// mergeMergePoints merges merge points using priority rules
func (m *PriorityMerger) mergeMergePoints(result *config.Config, incoming *config.Config) {
	for name, point := range incoming.MergePoints {
		existing, exists := result.MergePoints[name]
		if !exists || point.Priority.TakesPrecedenceOverOrEqual(existing.Priority) {
			if m.debug {
				fmt.Printf("Merging merge point %s from %s\n", name, incoming.SourceFile)
			}
			result.MergePoints[name] = point
		} else if m.debug {
			fmt.Printf("Skipping merge point %s (lower priority)\n", name)
		}
	}
}

// mergeMergeTargets merges merge targets using priority rules
func (m *PriorityMerger) mergeMergeTargets(result *config.Config, incoming *config.Config) {
	for name, target := range incoming.MergeTargets {
		existing, exists := result.MergeTargets[name]
		if !exists || target.Priority.TakesPrecedenceOverOrEqual(existing.Priority) {
			if m.debug {
				fmt.Printf("Merging merge target %s from %s\n", name, incoming.SourceFile)
			}
			result.MergeTargets[name] = target
		} else if m.debug {
			fmt.Printf("Skipping merge target %s (lower priority)\n", name)
		}
	}
}

// applyPlaceholderReplacements handles special placeholder replacements for markdown
func (m *PriorityMerger) applyPlaceholderReplacements(result *config.Config, configs []*config.Config) {
	// Collect content for placeholders from all configs
	replacements := make(map[string]string)

	// Process each config to find content for placeholders
	for _, cfg := range configs {
		if m.debug {
			fmt.Printf("Processing config: %s, Language: %s\n", cfg.SourceFile, cfg.Metadata.Language)
		}
		// Look for specific sections that might contain replacement content
		for _, section := range cfg.Sections {
			content := section.Content

			// Extract test commands
			if containsTestCommands(content) {
				testCommands := extractTestCommands(content)
				if m.debug {
					fmt.Printf("Found test commands: %s\n", testCommands)
				}
				replacements["test-commands"] = testCommands
			}

			// Extract documentation standards
			if containsDocumentationStandards(content) {
				docStandards := extractDocumentationStandards(content)
				if m.debug {
					fmt.Printf("Found documentation standards: %d chars\n", len(docStandards))
				}
				replacements["documentation-standards"] = docStandards
			}
		}
	}

	// Apply replacements to all sections
	for name, section := range result.Sections {
		content := section.Content

		// Replace placeholder blocks (including content between tags)
		if replacements["test-commands"] != "" {
			content = replacePlaceholderBlock(content, 
				"<language-specific-test-commands-here>", 
				"</language-specific-test-commands-here>", 
				replacements["test-commands"])
		} else {
			content = replacePlaceholderBlock(content, 
				"<language-specific-test-commands-here>", 
				"</language-specific-test-commands-here>", 
				"")
		}

		if replacements["documentation-standards"] != "" {
			content = replacePlaceholderBlock(content, 
				"<language-specific-documentation-standards>", 
				"</language-specific-documentation-standards>", 
				replacements["documentation-standards"])
		} else {
			content = replacePlaceholderBlock(content, 
				"<language-specific-documentation-standards>", 
				"</language-specific-documentation-standards>", 
				"")
		}

		section.Content = content
		result.Sections[name] = section
	}
}

// containsTestCommands checks if content has test commands section
func containsTestCommands(content string) bool {
	return strings.Contains(content, "Testing commands") || 
		strings.Contains(content, "### Testing commands") ||
		strings.Contains(content, "test ./...")
}

// extractTestCommands extracts test commands from content
func extractTestCommands(content string) string {
	lines := strings.Split(content, "\n")
	inTestSection := false
	var result []string

	for _, line := range lines {
		if strings.Contains(line, "Testing commands") || strings.Contains(line, "### Testing commands") {
			inTestSection = true
			continue
		}

		if inTestSection {
			// Stop at next section or empty line followed by non-test content
			if strings.HasPrefix(line, "#") && !strings.Contains(line, "Testing") {
				break
			}
			if line == "" && len(result) > 4 {
				// Check if we've collected enough test commands
				break
			}
			if line != "" {
				result = append(result, line)
			}
		}
	}

	return strings.Join(result, "\n")
}

// containsDocumentationStandards checks if content has documentation standards
func containsDocumentationStandards(content string) bool {
	return strings.Contains(content, "Documentation Standards") || 
		strings.Contains(content, "### Documentation Standards")
}

// extractDocumentationStandards extracts documentation standards from content
func extractDocumentationStandards(content string) string {
	lines := strings.Split(content, "\n")
	inDocSection := false
	var result []string
	codeBlockCount := 0

	for _, line := range lines {
		if strings.Contains(line, "Documentation Standards") || strings.Contains(line, "### Documentation Standards") {
			inDocSection = true
			result = append(result, "")  // Add empty line before standards
			continue
		}

		if inDocSection {
			// Count code blocks to know when to stop
			if strings.HasPrefix(line, "```") {
				codeBlockCount++
			}

			// Stop after closing the godoc example code block
			if codeBlockCount >= 2 && strings.HasPrefix(line, "```") {
				result = append(result, line)
				break
			}

			// Stop at next major section
			if strings.HasPrefix(line, "##") && !strings.Contains(line, "Documentation") {
				break
			}

			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// findBaseTemplate identifies if any config contains placeholders and should be used as base
func (m *PriorityMerger) findBaseTemplate(configs []*config.Config) *config.Config {
	for _, cfg := range configs {
		for _, section := range cfg.Sections {
			if containsPlaceholders(section.Content) {
				return cfg
			}
		}
	}
	return nil
}

// containsPlaceholders checks if content has any placeholder tags
func containsPlaceholders(content string) bool {
	return strings.Contains(content, "<language-specific-") && strings.Contains(content, "</language-specific-")
}

// mergeWithTemplate handles template-based merging where base has placeholders
func (m *PriorityMerger) mergeWithTemplate(result *config.Config, baseConfig *config.Config, configs []*config.Config) {
	// Start with base config
	result.Metadata = baseConfig.Metadata
	for name, section := range baseConfig.Sections {
		result.Sections[name] = section
	}
	for name, mp := range baseConfig.MergePoints {
		result.MergePoints[name] = mp
	}
	for name, mt := range baseConfig.MergeTargets {
		result.MergeTargets[name] = mt
	}

	// Apply placeholder replacements
	m.applyPlaceholderReplacements(result, configs)

	// Merge metadata from other configs if they have higher priority
	for _, cfg := range configs {
		if cfg != baseConfig {
			// Only merge metadata, not sections
			if cfg.Metadata.Priority.TakesPrecedenceOverOrEqual(result.Metadata.Priority) {
				if cfg.Metadata.Title != "" && cfg.Metadata.Title != result.Metadata.Title {
					// Don't override the title from base unless explicitly higher priority
					if cfg.Metadata.Priority.TakesPrecedenceOver(result.Metadata.Priority) {
						result.Metadata.Title = cfg.Metadata.Title
					}
				}
			}
		}
	}
}

// replacePlaceholderBlock replaces content between opening and closing tags
func replacePlaceholderBlock(content, openTag, closeTag, replacement string) string {
	startIdx := strings.Index(content, openTag)
	if startIdx == -1 {
		return content
	}

	endIdx := strings.Index(content, closeTag)
	if endIdx == -1 {
		return content
	}

	// Replace everything from openTag to closeTag (inclusive) with replacement
	before := content[:startIdx]
	after := content[endIdx+len(closeTag):]
	
	return before + replacement + after
}