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
	rootCmd.AddCommand(newCreateStepNavigateCmd())
	rootCmd.AddCommand(newCreateStepWaitTimeCmd())
	rootCmd.AddCommand(newCreateStepWaitElementCmd())
	rootCmd.AddCommand(newCreateStepWindowCmd())
	rootCmd.AddCommand(newCreateStepClickCmd())
	rootCmd.AddCommand(newCreateStepDoubleClickCmd())
	rootCmd.AddCommand(newCreateStepHoverCmd())
	rootCmd.AddCommand(newCreateStepRightClickCmd())
	rootCmd.AddCommand(newCreateStepWriteCmd())
	rootCmd.AddCommand(newCreateStepKeyCmd())
	rootCmd.AddCommand(newCreateStepPickCmd())
	rootCmd.AddCommand(newCreateStepUploadCmd())
	rootCmd.AddCommand(newCreateStepScrollTopCmd())
	rootCmd.AddCommand(newCreateStepScrollBottomCmd())
	rootCmd.AddCommand(newCreateStepScrollElementCmd())
	rootCmd.AddCommand(newCreateStepAssertExistsCmd())
	rootCmd.AddCommand(newCreateStepAssertNotExistsCmd())
	rootCmd.AddCommand(newCreateStepAssertEqualsCmd())
	rootCmd.AddCommand(newCreateStepAssertCheckedCmd())
	rootCmd.AddCommand(newCreateStepStoreCmd())
	rootCmd.AddCommand(newCreateStepExecuteJsCmd())
	rootCmd.AddCommand(newCreateStepAddCookieCmd())
	rootCmd.AddCommand(newCreateStepDismissAlertCmd())
	rootCmd.AddCommand(newCreateStepCommentCmd())
	rootCmd.AddCommand(newCreateStepAssertLessThanOrEqualCmd())
	rootCmd.AddCommand(newCreateStepAssertSelectedCmd())
	rootCmd.AddCommand(newCreateStepAssertVariableCmd())
	rootCmd.AddCommand(newCreateStepDismissConfirmCmd())
	rootCmd.AddCommand(newCreateStepDismissPromptCmd())
	rootCmd.AddCommand(newCreateStepClearCookiesCmd())
	rootCmd.AddCommand(newCreateStepDeleteCookieCmd())
	rootCmd.AddCommand(newCreateStepMouseDownCmd())
	rootCmd.AddCommand(newCreateStepMouseUpCmd())
	rootCmd.AddCommand(newCreateStepMouseMoveCmd())
	rootCmd.AddCommand(newCreateStepMouseEnterCmd())
	rootCmd.AddCommand(newCreateStepPickValueCmd())
	rootCmd.AddCommand(newCreateStepPickTextCmd())
	rootCmd.AddCommand(newCreateStepScrollPositionCmd())
	rootCmd.AddCommand(newCreateStepStoreValueCmd())
	// Use enhanced version that handles auto-creation behavior
	rootCmd.AddCommand(newCreateStructureEnhancedCmd())
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

