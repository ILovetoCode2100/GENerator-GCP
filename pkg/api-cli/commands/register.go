package commands

import (
	"github.com/spf13/cobra"
)

// RegisterCommands registers all CLI commands to the root command
func RegisterCommands(rootCmd *cobra.Command) {
	// ========================================
	// CORE COMMANDS - Configuration and Session
	// ========================================
	rootCmd.AddCommand(newValidateConfigCmd())
	// Set checkpoint functionality not implemented yet
	// rootCmd.AddCommand(newSetCheckpointCmd())

	// ========================================
	// PROJECT MANAGEMENT COMMANDS (Consolidated)
	// ========================================
	// From manage_projects.go
	rootCmd.AddCommand(newCreateProjectCmd())
	rootCmd.AddCommand(newCreateGoalCmd())
	rootCmd.AddCommand(newCreateJourneyCmd())
	rootCmd.AddCommand(newCreateCheckpointCmd())
	rootCmd.AddCommand(newUpdateJourneyCmd())
	rootCmd.AddCommand(newGetStepCmd())
	rootCmd.AddCommand(newUpdateNavigationCmd())

	// ========================================
	// LIST COMMANDS (Consolidated)
	// ========================================
	// From list.go
	rootCmd.AddCommand(NewListProjectsCmd())
	rootCmd.AddCommand(NewListGoalsCmd())
	rootCmd.AddCommand(NewListJourneysCmd())
	rootCmd.AddCommand(NewListCheckpointsCmd())

	// ========================================
	// YAML COMMAND GROUP
	// ========================================
	rootCmd.AddCommand(NewYAMLCmd())

	// ========================================
	// TEST STEP COMMANDS (9 Groups - All create test steps)
	// ========================================
	// All commands use unified positional argument syntax with 'step-' prefix

	// 1. STEP-ASSERT - All assertion operations
	rootCmd.AddCommand(newStepAssertCmd()) // equals, not-equals, exists, not-exists, gt, gte, lt, lte, matches, checked, selected, variable

	// 2. STEP-INTERACT - All user interaction actions (consolidated from interact.go, mouse.go, select.go)
	rootCmd.AddCommand(StepInteractionCmd()) // click, double-click, right-click, hover, write, key, mouse operations, select dropdowns

	// 3. STEP-NAVIGATE - Browser navigation and scrolling
	rootCmd.AddCommand(StepNavigateCmd()) // to, scroll-top, scroll-bottom, scroll-element, scroll-position, scroll-by, scroll-up, scroll-down

	// 4. STEP-WINDOW - Window management operations
	rootCmd.AddCommand(newStepWindowCmd()) // resize, maximize, switch tab/iframe/parent-frame

	// 5. STEP-DATA - Data storage and cookies
	rootCmd.AddCommand(newStepDataCmd()) // store element-text/attribute/literal, cookie create/delete/clear-all

	// 6. STEP-DIALOG - Dialog handling
	rootCmd.AddCommand(newStepDialogCmd()) // dismiss-alert, dismiss-confirm, dismiss-prompt, dismiss-prompt-with-text

	// 7. STEP-WAIT - Wait operations
	rootCmd.AddCommand(newStepWaitCmd()) // element, time

	// 8. STEP-FILE - File operations
	rootCmd.AddCommand(StepFileCmd()) // upload, upload-url

	// 9. STEP-MISC - Miscellaneous actions
	rootCmd.AddCommand(newStepMiscCmd()) // comment, execute

	// ========================================
	// LIBRARY MANAGEMENT COMMANDS
	// ========================================
	rootCmd.AddCommand(LibraryCmd()) // add, get, attach, move-step, remove-step, update

	// ========================================
	// EXECUTION AND ANALYSIS COMMANDS (Consolidated)
	// ========================================
	// From execution_management.go
	rootCmd.AddCommand(NewCreateEnvironmentCmd())
	rootCmd.AddCommand(NewManageTestDataCmd())
	rootCmd.AddCommand(NewExecuteGoalCmd())
	rootCmd.AddCommand(NewMonitorExecutionCmd())
	rootCmd.AddCommand(NewGetExecutionAnalysisCmd())

	// ========================================
	// TEST TEMPLATE COMMANDS (AI Integration)
	// ========================================
	rootCmd.AddCommand(LoadTestTemplateCmd)
	rootCmd.AddCommand(GenerateCommandsCmd)
	rootCmd.AddCommand(GetTestTemplatesCmd)

	// ========================================
	// UNIFIED TEST INTERFACE
	// ========================================
	rootCmd.AddCommand(newRunTestCmd()) // New simplified run-test command
}
