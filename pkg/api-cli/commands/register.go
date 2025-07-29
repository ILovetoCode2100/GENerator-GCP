package commands

import (
	"github.com/spf13/cobra"
)

// RegisterCommands registers all CLI commands to the root command
func RegisterCommands(rootCmd *cobra.Command) {
	// Create command validator instance
	validator := NewCommandValidator()

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
	rootCmd.AddCommand(NewListCheckpointStepsCmd())

	// ========================================
	// YAML COMMAND GROUP
	// ========================================
	rootCmd.AddCommand(NewYAMLCmd())

	// ========================================
	// TEST STEP COMMANDS (9 Groups - All create test steps)
	// ========================================
	// All commands use unified positional argument syntax with 'step-' prefix
	// Apply validator middleware to all step commands

	// 1. STEP-ASSERT - All assertion operations
	assertCmd := newStepAssertCmd()
	applyValidatorToCommand(assertCmd, validator)
	rootCmd.AddCommand(assertCmd) // equals, not-equals, exists, not-exists, gt, gte, lt, lte, matches, checked, selected, variable

	// 2. STEP-INTERACT - All user interaction actions (consolidated from interact.go, mouse.go, select.go)
	interactCmd := StepInteractionCmd()
	applyValidatorToCommand(interactCmd, validator)
	rootCmd.AddCommand(interactCmd) // click, double-click, right-click, hover, write, key, mouse operations, select dropdowns

	// 3. STEP-NAVIGATE - Browser navigation and scrolling
	navigateCmd := StepNavigateCmd()
	applyValidatorToCommand(navigateCmd, validator)
	rootCmd.AddCommand(navigateCmd) // to, scroll-top, scroll-bottom, scroll-element, scroll-position, scroll-by, scroll-up, scroll-down

	// 4. STEP-WINDOW - Window management operations
	windowCmd := newStepWindowCmd()
	applyValidatorToCommand(windowCmd, validator)
	rootCmd.AddCommand(windowCmd) // resize, maximize, switch tab/iframe/parent-frame

	// 5. STEP-DATA - Data storage and cookies
	dataCmd := newStepDataCmd()
	applyValidatorToCommand(dataCmd, validator)
	rootCmd.AddCommand(dataCmd) // store element-text/attribute/literal, cookie create/delete/clear-all

	// 6. STEP-DIALOG - Dialog handling
	dialogCmd := newStepDialogCmd()
	applyValidatorToCommand(dialogCmd, validator)
	rootCmd.AddCommand(dialogCmd) // dismiss-alert, dismiss-confirm, dismiss-prompt, dismiss-prompt-with-text

	// 7. STEP-WAIT - Wait operations
	waitCmd := newStepWaitCmd()
	applyValidatorToCommand(waitCmd, validator)
	rootCmd.AddCommand(waitCmd) // element, time

	// 8. STEP-FILE - File operations
	fileCmd := StepFileCmd()
	applyValidatorToCommand(fileCmd, validator)
	rootCmd.AddCommand(fileCmd) // upload, upload-url

	// 9. STEP-MISC - Miscellaneous actions
	miscCmd := newStepMiscCmd()
	applyValidatorToCommand(miscCmd, validator)
	rootCmd.AddCommand(miscCmd) // comment, execute

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

// applyValidatorToCommand applies the command validator middleware to a command and all its subcommands
func applyValidatorToCommand(cmd *cobra.Command, validator *CommandValidator) {
	// TODO: Implement ApplyAsMiddleware method
	// validator.ApplyAsMiddleware(cmd)

	// Apply to all subcommands recursively
	for _, subCmd := range cmd.Commands() {
		applyValidatorToCommand(subCmd, validator)
	}
}
