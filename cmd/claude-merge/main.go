package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/arustydev/claude-merge/internal/config"
	"github.com/arustydev/claude-merge/internal/generator"
	"github.com/arustydev/claude-merge/internal/merger"
)

func main() {
	// Define command-line flags
	var (
		files      = flag.String("files", "", "Comma-separated paths to configuration files (required)")
		outputFile = flag.String("output", "CLAUDE.merged.md", "Output filename (default: CLAUDE.merged.md)")
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
	fmt.Println("  -output string   Output filename (default: CLAUDE.merged.md)")
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