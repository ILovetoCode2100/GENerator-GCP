package main

import (
	"fmt"
	"os"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/commands"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "1.0.0"

	// Global config
	cfg *config.VirtuosoConfig

	// Flags
	cfgFile string
	verbose bool
	output  string
)

var rootCmd = &cobra.Command{
	Use:   "api-cli",
	Short: "Virtuoso API CLI - Intelligent orchestration for test automation",
	Long: `A command-line interface for Virtuoso API that provides:
- Automated multi-step workflows
- AI-friendly interfaces and output
- Business rule enforcement`,
	Version: Version,
}

func init() {
	cobra.OnInitialize(loadConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config/virtuoso-config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "human", "output format (json, yaml, human, ai)")
}

// loadConfig centralizes all configuration loading using Viper
// Priority order: command-line flags > environment variables > config file > defaults
// This function is crucial for AI-driven test building as it sets up:
// - API endpoints for test submission
// - Output formats for AI parsing
// - Test template directories for batch processing
// - Session context for multi-step test workflows
func loadConfig() {
	var err error

	// Load configuration using the enhanced LoadConfig function
	cfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		// Use consistent exit code for configuration errors
		os.Exit(commands.ExitGeneralError)
	}

	// Apply command line flags (highest priority)
	if verbose {
		cfg.Output.Verbose = true
	}
	if output != "" {
		// Validate output format - AI format is crucial for test generation
		if err := validateOutputFormat(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			// Use consistent exit code for validation errors
			os.Exit(commands.ExitValidationError)
		}
		cfg.Output.DefaultFormat = output
	}

	// Set config in commands package for global access
	// This enables all commands to use consistent configuration
	// for API endpoints, authentication, and AI-specific settings
	commands.SetConfig(cfg)

	// Log configuration source for debugging (useful for AI troubleshooting)
	if cfg.Output.Verbose {
		fmt.Printf("Config loaded from: %s\n", config.GetConfigSource())
		fmt.Printf("Output format: %s\n", cfg.Output.DefaultFormat)
		if cfg.Test.BatchDir != "" {
			fmt.Printf("Batch test directory: %s\n", cfg.Test.BatchDir)
		}
	}
}

func validateOutputFormat(format string) error {
	validFormats := []string{"human", "json", "yaml", "ai"}
	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid output format: %s (must be one of: human, json, yaml, ai)", format)
}

func main() {
	// Register all commands
	commands.RegisterCommands(rootCmd)

	// Execute root command (this triggers initConfig via cobra.OnInitialize)
	if err := rootCmd.Execute(); err != nil {
		// Use appropriate exit code based on error type
		exitCode := commands.ExitGeneralError

		// Print error to stderr
		fmt.Fprintln(os.Stderr, err)

		// Exit with the determined code
		os.Exit(exitCode)
	}

	// Successful execution
	os.Exit(0)
}
