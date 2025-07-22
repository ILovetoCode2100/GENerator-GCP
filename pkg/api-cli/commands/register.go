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
	// From project_management.go
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
	// STEP COMMANDS (8 Groups - Further Consolidated)
	// ========================================
	// All commands use unified positional argument syntax

	// 1. ASSERT - All assertion operations
	rootCmd.AddCommand(newAssertCmd()) // equals, not-equals, exists, not-exists, gt, gte, lt, lte, matches, checked, selected, variable

	// 2. INTERACT - All user interaction actions (consolidated from interact.go, mouse.go, select.go)
	rootCmd.AddCommand(InteractionCmd()) // click, double-click, right-click, hover, write, key, mouse operations, select dropdowns

	// 3. BROWSER COMMANDS (consolidated from navigate.go, window.go)
	rootCmd.AddCommand(NavigateCmd())  // to, scroll-to, scroll-top, scroll-bottom, scroll-element, scroll-by, scroll-up, scroll-down
	rootCmd.AddCommand(newWindowCmd()) // resize, maximize, switch-tab, switch-iframe, switch-parent-frame

	// 4. DATA - Data storage and cookies
	rootCmd.AddCommand(newDataCmd()) // store-text, store-value, store-attribute, cookie-create, cookie-delete, cookie-clear

	// 5. DIALOG - Dialog handling
	rootCmd.AddCommand(newDialogCmd()) // alert-accept, alert-dismiss, confirm-accept, confirm-reject, prompt

	// 6. WAIT - Wait operations
	rootCmd.AddCommand(newWaitCmd()) // element, element-not-visible, time

	// 7. FILE - File operations
	rootCmd.AddCommand(FileCmd()) // upload, upload-url

	// 8. MISC - Miscellaneous actions
	rootCmd.AddCommand(newMiscCmd()) // comment, execute

	// 9. LIBRARY - Library checkpoint operations
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
}
