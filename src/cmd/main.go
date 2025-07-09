package main

import (
	"fmt"
	"os"

	"github.com/marklovelady/api-cli-generator/pkg/config"
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
- Batch creation from JSON/YAML files  
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
		cfg.Output.DefaultFormat = output
	}
}

func main() {
	// Add commands
	rootCmd.AddCommand(newValidateConfigCmd())
	rootCmd.AddCommand(newCreateProjectCmd())
	rootCmd.AddCommand(newCreateGoalCmd())
	rootCmd.AddCommand(newCreateJourneyCmd())
	rootCmd.AddCommand(newUpdateJourneyCmd())
	rootCmd.AddCommand(newCreateCheckpointCmd())
	rootCmd.AddCommand(newAddStepCmd())
	rootCmd.AddCommand(newGetStepCmd())
	rootCmd.AddCommand(newUpdateNavigationCmd())
	rootCmd.AddCommand(newListProjectsCmd())
	rootCmd.AddCommand(newListGoalsCmd())
	rootCmd.AddCommand(newListJourneysCmd())
	rootCmd.AddCommand(newListCheckpointsCmd())
	// Use enhanced version that handles auto-creation behavior
	rootCmd.AddCommand(newCreateStructureEnhancedCmd())
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

