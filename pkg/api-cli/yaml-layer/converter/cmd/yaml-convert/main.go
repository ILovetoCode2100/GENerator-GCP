package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/your-org/virtuoso-generator/pkg/api-cli/yaml-layer/converter"
	"github.com/your-org/virtuoso-generator/pkg/api-cli/yaml-layer/detector"
)

func main() {
	// Define flags
	var (
		inputFile    = flag.String("i", "", "Input YAML file (use '-' for stdin)")
		outputFile   = flag.String("o", "-", "Output file (use '-' for stdout)")
		targetFormat = flag.String("f", "", "Target format: compact, simplified, extended")
		detect       = flag.Bool("detect", false, "Detect format of input file")
		analyze      = flag.Bool("analyze", false, "Analyze input file structure")
		validate     = flag.String("validate", "", "Validate file matches format: compact, simplified, extended")
		showFormats  = flag.Bool("formats", false, "Show information about all formats")
		batch        = flag.Bool("batch", false, "Batch convert multiple files")
		outputDir    = flag.String("d", "converted", "Output directory for batch conversion")
		exportJSON   = flag.Bool("json", false, "Export conversion result as JSON")
		quiet        = flag.Bool("q", false, "Quiet mode (suppress warnings)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "YAML Format Converter - Convert between Virtuoso test formats\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s -i input.yaml -f compact -o output.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -detect test.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -analyze test.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -formats\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nFormats:\n")
		fmt.Fprintf(os.Stderr, "  compact     - AI-optimized concise format (yaml command)\n")
		fmt.Fprintf(os.Stderr, "  simplified  - Human-readable format (run-test command)\n")
		fmt.Fprintf(os.Stderr, "  extended    - Verbose format with metadata (no CLI support)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Convert compact to simplified\n")
		fmt.Fprintf(os.Stderr, "  %s -i test.yaml -f simplified -o test-simple.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n  # Detect format\n")
		fmt.Fprintf(os.Stderr, "  %s -detect test.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n  # Convert from stdin to stdout\n")
		fmt.Fprintf(os.Stderr, "  cat test.yaml | %s -i - -f compact\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n  # Batch convert all YAML files\n")
		fmt.Fprintf(os.Stderr, "  %s -batch -f simplified -d output/ *.yaml\n", os.Args[0])
	}

	flag.Parse()

	// Create CLI instance
	cli := converter.NewCLI()

	// Suppress warnings in quiet mode
	if *quiet {
		// Redirect stderr temporarily
		oldStderr := os.Stderr
		os.Stderr = nil
		defer func() { os.Stderr = oldStderr }()
	}

	// Handle different modes
	switch {
	case *showFormats:
		cli.ShowFormats()
		return

	case *detect:
		if *inputFile == "" {
			fmt.Fprintf(os.Stderr, "Error: -i flag required for format detection\n")
			os.Exit(1)
		}
		result, err := cli.DetectFormat(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Format: %s\n", detector.GetFormatDescription(result.Format))
		fmt.Printf("Confidence: %.0f%%\n", result.Confidence*100)
		if len(result.Warnings) > 0 && !*quiet {
			fmt.Println("\nWarnings:")
			for _, w := range result.Warnings {
				fmt.Printf("  - %s\n", w)
			}
		}
		return

	case *analyze:
		if *inputFile == "" {
			fmt.Fprintf(os.Stderr, "Error: -i flag required for analysis\n")
			os.Exit(1)
		}
		if err := cli.AnalyzeFile(*inputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return

	case *validate != "":
		if *inputFile == "" {
			fmt.Fprintf(os.Stderr, "Error: -i flag required for validation\n")
			os.Exit(1)
		}
		format := parseFormat(*validate)
		if format == "" {
			fmt.Fprintf(os.Stderr, "Error: Invalid format '%s'\n", *validate)
			os.Exit(1)
		}
		if err := cli.ValidateFile(*inputFile, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return

	case *exportJSON:
		if *inputFile == "" || *targetFormat == "" {
			fmt.Fprintf(os.Stderr, "Error: -i and -f flags required for JSON export\n")
			os.Exit(1)
		}
		format := parseFormat(*targetFormat)
		if format == "" {
			fmt.Fprintf(os.Stderr, "Error: Invalid target format '%s'\n", *targetFormat)
			os.Exit(1)
		}
		if err := cli.ExportConversionResult(*inputFile, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return

	case *batch:
		if *targetFormat == "" {
			fmt.Fprintf(os.Stderr, "Error: -f flag required for batch conversion\n")
			os.Exit(1)
		}
		format := parseFormat(*targetFormat)
		if format == "" {
			fmt.Fprintf(os.Stderr, "Error: Invalid target format '%s'\n", *targetFormat)
			os.Exit(1)
		}
		files := flag.Args()
		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "Error: No input files specified\n")
			os.Exit(1)
		}
		if err := cli.ConvertBatch(files, *outputDir, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return

	default:
		// Regular conversion mode
		if *targetFormat == "" {
			fmt.Fprintf(os.Stderr, "Error: -f flag required for conversion\n")
			flag.Usage()
			os.Exit(1)
		}

		format := parseFormat(*targetFormat)
		if format == "" {
			fmt.Fprintf(os.Stderr, "Error: Invalid target format '%s'\n", *targetFormat)
			os.Exit(1)
		}

		// Handle stdin/stdout conversion
		if *inputFile == "" || *inputFile == "-" {
			if err := cli.ConvertStream(format); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// File conversion
		if err := cli.ConvertFile(*inputFile, *outputFile, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Success message (unless outputting to stdout)
		if *outputFile != "-" && !*quiet {
			fmt.Printf("✓ Converted %s to %s format\n", *inputFile, format)
			fmt.Printf("  Output: %s\n", *outputFile)
		}
	}
}

// parseFormat converts string to YAMLFormat
func parseFormat(s string) detector.YAMLFormat {
	switch strings.ToLower(s) {
	case "compact", "c":
		return detector.FormatCompact
	case "simplified", "simple", "s":
		return detector.FormatSimplified
	case "extended", "ext", "e":
		return detector.FormatExtended
	default:
		return ""
	}
}