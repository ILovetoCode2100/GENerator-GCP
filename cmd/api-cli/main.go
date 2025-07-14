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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config/virtuoso-config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "human", "output format (json, yaml, human, ai)")
}

func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Apply command line flags
	if verbose {
		cfg.Output.Verbose = true
	}
	if output != "" {
		// Validate output format
		if err := validateOutputFormat(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		cfg.Output.DefaultFormat = output
	}

	// Set config in commands package
	commands.SetConfig(cfg)
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
