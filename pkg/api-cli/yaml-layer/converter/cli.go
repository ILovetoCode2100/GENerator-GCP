package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/yaml-layer/detector"
	"gopkg.in/yaml.v3"
)

// CLI provides command-line interface for the converter
type CLI struct {
	converter *FormatConverter
	detector  *detector.FormatDetector
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		converter: NewFormatConverter(),
		detector:  detector.NewFormatDetector(),
	}
}

// ConvertFile converts a YAML file from one format to another
func (c *CLI) ConvertFile(inputPath string, outputPath string, targetFormat detector.YAMLFormat) error {
	// Read input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Convert
	result, err := c.converter.Convert(content, targetFormat)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Print warnings
	for _, warning := range result.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	}

	// Write output
	if outputPath == "-" {
		_, err = os.Stdout.Write(result.Output)
	} else {
		err = os.WriteFile(outputPath, result.Output, 0644)
	}

	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// DetectFormat detects the format of a YAML file
func (c *CLI) DetectFormat(inputPath string) (*detector.DetectionResult, error) {
	// Read input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	return c.detector.DetectFormat(content)
}

// ConvertStream converts YAML from stdin to stdout
func (c *CLI) ConvertStream(targetFormat detector.YAMLFormat) error {
	// Read from stdin
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Convert
	result, err := c.converter.Convert(content, targetFormat)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Print warnings to stderr
	for _, warning := range result.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	}

	// Write output to stdout
	_, err = os.Stdout.Write(result.Output)
	return err
}

// ShowFormats displays information about supported formats
func (c *CLI) ShowFormats() {
	formats := []detector.YAMLFormat{
		detector.FormatCompact,
		detector.FormatSimplified,
		detector.FormatExtended,
	}

	for _, format := range formats {
		fmt.Printf("\n=== %s ===\n", detector.GetFormatDescription(format))
		fmt.Printf("Format ID: %s\n", format)
		fmt.Printf("CLI Support: %v\n", detector.IsFormatSupported(format))

		if cmd := detector.GetSupportedCommand(format); cmd != "" {
			fmt.Printf("CLI Command: %s\n", cmd)
		}

		fmt.Printf("\nFeatures:\n")
		features := GetFormatFeatures()[string(format)]
		for _, feature := range features {
			fmt.Printf("  - %s\n", feature)
		}

		fmt.Printf("\nExample:\n%s\n", detector.GetFormatExample(format))
	}

	fmt.Printf("\n=== Conversion Capabilities ===\n")
	caps := GetConversionCapabilities()
	for from, targets := range caps {
		fmt.Printf("\nFrom %s:\n", from)
		for to, supported := range targets {
			if supported {
				fmt.Printf("  ✓ Can convert to %s\n", to)
			}
		}
	}
}

// ValidateFile validates a YAML file format
func (c *CLI) ValidateFile(inputPath string, expectedFormat detector.YAMLFormat) error {
	// Read input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Validate
	if err := c.detector.ValidateFormat(content, expectedFormat); err != nil {
		return err
	}

	fmt.Printf("✓ File is valid %s format\n", detector.GetFormatDescription(expectedFormat))
	return nil
}

// AnalyzeFile provides detailed analysis of a YAML file
func (c *CLI) AnalyzeFile(inputPath string) error {
	// Read input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Detect format
	result, err := c.detector.DetectFormat(content)
	if err != nil {
		return fmt.Errorf("failed to analyze file: %w", err)
	}

	// Print analysis
	fmt.Printf("File: %s\n", inputPath)
	fmt.Printf("Detected Format: %s\n", detector.GetFormatDescription(result.Format))
	fmt.Printf("Confidence: %.2f\n", result.Confidence)

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
	}

	fmt.Printf("\nDetected Features:\n")
	for feature, present := range result.Features {
		if present {
			fmt.Printf("  ✓ %s\n", feature)
		}
	}

	// Parse and show structure
	var data interface{}
	if err := yaml.Unmarshal(content, &data); err == nil {
		fmt.Printf("\nStructure:\n")
		c.printStructure(data, "  ")
	}

	return nil
}

// printStructure recursively prints YAML structure
func (c *CLI) printStructure(data interface{}, indent string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fmt.Printf("%s%s:", indent, key)
			if _, ok := value.(map[string]interface{}); ok {
				fmt.Println()
				c.printStructure(value, indent+"  ")
			} else if _, ok := value.([]interface{}); ok {
				fmt.Printf(" [%d items]\n", len(value.([]interface{})))
				if len(value.([]interface{})) > 0 && len(value.([]interface{})) <= 3 {
					c.printStructure(value, indent+"  ")
				}
			} else {
				fmt.Printf(" %v\n", value)
			}
		}
	case []interface{}:
		for i, item := range v {
			if i >= 3 {
				fmt.Printf("%s... (%d more items)\n", indent, len(v)-3)
				break
			}
			fmt.Printf("%s- ", indent)
			if _, ok := item.(map[string]interface{}); ok {
				fmt.Println()
				c.printStructure(item, indent+"  ")
			} else {
				fmt.Printf("%v\n", item)
			}
		}
	}
}

// ConvertBatch converts multiple files
func (c *CLI) ConvertBatch(inputFiles []string, outputDir string, targetFormat detector.YAMLFormat) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	results := make(map[string]error)

	for _, inputFile := range inputFiles {
		outputFile := fmt.Sprintf("%s/%s.%s.yaml", outputDir, getBaseName(inputFile), targetFormat)
		err := c.ConvertFile(inputFile, outputFile, targetFormat)
		results[inputFile] = err

		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", inputFile, err)
		} else {
			fmt.Printf("✓ %s → %s\n", inputFile, outputFile)
		}
	}

	// Summary
	succeeded := 0
	for _, err := range results {
		if err == nil {
			succeeded++
		}
	}

	fmt.Printf("\nSummary: %d/%d files converted successfully\n", succeeded, len(inputFiles))

	if succeeded < len(inputFiles) {
		return fmt.Errorf("some conversions failed")
	}

	return nil
}

// getBaseName extracts base name from file path
func getBaseName(path string) string {
	base := path
	if idx := len(path) - 1; idx >= 0 && path[idx] == '/' {
		base = path[:idx]
	}
	if idx := len(base) - 1; idx >= 0 {
		for i := idx; i >= 0; i-- {
			if base[i] == '/' {
				base = base[i+1:]
				break
			}
		}
	}
	if idx := len(base) - 1; idx >= 0 {
		for i := idx; i >= 0; i-- {
			if base[i] == '.' {
				base = base[:i]
				break
			}
		}
	}
	return base
}

// ExportConversionResult exports the full conversion result as JSON
func (c *CLI) ExportConversionResult(inputPath string, targetFormat detector.YAMLFormat) error {
	// Read input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Detect source format
	detection, err := c.detector.DetectFormat(content)
	if err != nil {
		return fmt.Errorf("failed to detect format: %w", err)
	}

	// Convert
	result, err := c.converter.Convert(content, targetFormat)
	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Create export data
	export := map[string]interface{}{
		"source": map[string]interface{}{
			"file":       inputPath,
			"format":     detection.Format,
			"confidence": detection.Confidence,
			"features":   detection.Features,
		},
		"target": map[string]interface{}{
			"format": targetFormat,
		},
		"conversion": map[string]interface{}{
			"success":  true,
			"warnings": result.Warnings,
		},
		"output": string(result.Output),
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = os.Stdout.Write(jsonData)
	return err
}

// GetFormatFeatures returns features for each format
func GetFormatFeatures() map[string][]string {
	return map[string][]string{
		string(detector.FormatCompact): {
			"Minimal token usage (59% reduction)",
			"Abbreviated action syntax (c, t, ch)",
			"Inline variable references with $",
			"Compact conditional logic",
			"Native support for data-driven tests",
		},
		string(detector.FormatSimplified): {
			"Human-readable action names",
			"Clear step structure",
			"Explicit selectors and values",
			"Infrastructure configuration",
			"Variables with template syntax {{}}",
		},
		string(detector.FormatExtended): {
			"Full metadata support",
			"Type and command structure",
			"Advanced test configuration",
			"Detailed step annotations",
			"Complete feature coverage",
		},
	}
}

// GetConversionCapabilities returns conversion matrix
func GetConversionCapabilities() map[string]map[string]bool {
	return map[string]map[string]bool{
		string(detector.FormatCompact): {
			string(detector.FormatSimplified): true,
			string(detector.FormatExtended):   true,
		},
		string(detector.FormatSimplified): {
			string(detector.FormatCompact):  true,
			string(detector.FormatExtended): true,
		},
		string(detector.FormatExtended): {
			string(detector.FormatCompact):    true,
			string(detector.FormatSimplified): true,
		},
	}
}
