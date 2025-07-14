package commands

import (
	"github.com/spf13/cobra"
)

// RegisterCommands registers all CLI commands to the root command
func RegisterCommands(rootCmd *cobra.Command) {
	// ========================================
	// CORE COMMANDS - Project/Journey/Goal Management
	// ========================================
	rootCmd.AddCommand(newValidateConfigCmd())
	rootCmd.AddCommand(newCreateProjectCmd())
	rootCmd.AddCommand(newCreateGoalCmd())
	rootCmd.AddCommand(newCreateJourneyCmd())
	rootCmd.AddCommand(newUpdateJourneyCmd())
	rootCmd.AddCommand(newCreateCheckpointCmd())
	rootCmd.AddCommand(newSetCheckpointCmd())
	rootCmd.AddCommand(newAddStepCmd())
	rootCmd.AddCommand(newGetStepCmd())
	rootCmd.AddCommand(newUpdateNavigationCmd())
	rootCmd.AddCommand(newListProjectsCmd())
	rootCmd.AddCommand(newListGoalsCmd())
	rootCmd.AddCommand(newListJourneysCmd())
	rootCmd.AddCommand(newListCheckpointsCmd())

	// ========================================
	// CONSOLIDATED STEP COMMANDS (11 Groups)
	// ========================================
	// These commands replace 54 individual legacy commands with a more organized structure

	// 1. ASSERT - All assertion operations (12 legacy commands → 1 group)
	rootCmd.AddCommand(newAssertCmd()) // equals, not-equals, exists, not-exists, gt, gte, lt, lte, matches, checked, selected, variable

	// 2. INTERACT - User interaction actions (6 legacy commands → 1 group)
	rootCmd.AddCommand(InteractCmd()) // click, double-click, right-click, hover, write, key

	// 3. NAVIGATE - Navigation and scrolling (5 legacy commands → 1 group)
	rootCmd.AddCommand(NavigateCmd()) // url (navigate), scroll-to, scroll-top, scroll-bottom, scroll-element

	// 4. DATA - Data storage and cookies (5 legacy commands → 1 group)
	rootCmd.AddCommand(newDataCmd()) // store-text, store-value, cookie-create, cookie-delete, cookie-clear

	// 5. DIALOG - Dialog handling (4 legacy commands → 1 group)
	rootCmd.AddCommand(newDialogCmd()) // dismiss-alert, dismiss-confirm, dismiss-prompt, dismiss-prompt-with-text

	// 6. WAIT - Wait operations (4 legacy commands → 1 group)
	rootCmd.AddCommand(newWaitCmd()) // element, time (replaces wait-element, wait-for-element-default/timeout, wait-time)

	// 7. WINDOW - Window and frame management (5 legacy commands → 1 group)
	rootCmd.AddCommand(newWindowCmd()) // resize, switch-tab (next/prev), switch-frame (iframe/parent)

	// 8. MOUSE - Advanced mouse operations (6 legacy commands → 1 group)
	rootCmd.AddCommand(newMouseCmd()) // move-to, move-by, move, down, up, enter

	// 9. SELECT - Dropdown selection (3 legacy commands → 1 group)
	rootCmd.AddCommand(newSelectCmd()) // option, index, last

	// 10. FILE - File operations (2 legacy commands → 1 group)
	rootCmd.AddCommand(FileCmd()) // upload, upload-url

	// 11. MISC - Miscellaneous actions (2 legacy commands → 1 group)
	rootCmd.AddCommand(newMiscCmd()) // comment, execute (script)

	// ========================================
	// LEGACY COMMAND SUPPORT
	// ========================================
	// Register all 54 legacy commands with deprecation warnings
	// These provide backward compatibility and guide users to new commands
	registerLegacyCommands(rootCmd)

	// ========================================
	// LEGACY COMMANDS (ALL DEPRECATED - Use consolidated commands above)
	// ========================================
	// The following 54 individual commands are now replaced by the 11 consolidated groups above.
	// They remain available through the legacy wrapper for backward compatibility.
	// rootCmd.AddCommand(newCreateStepAssertExistsCmd())
	// rootCmd.AddCommand(newCreateStepAssertNotExistsCmd())
	// rootCmd.AddCommand(newCreateStepAssertEqualsCmd())
	// rootCmd.AddCommand(newCreateStepAssertNotEqualsCmd())
	// rootCmd.AddCommand(newCreateStepAssertGreaterThanCmd())
	// rootCmd.AddCommand(newCreateStepAssertGreaterThanOrEqualCmd())
	// rootCmd.AddCommand(newCreateStepAssertMatchesCmd())
	// rootCmd.AddCommand(newCreateStepAssertCheckedCmd())
	// rootCmd.AddCommand(newCreateStepDismissAlertCmd())  // Replaced by 'dialog dismiss alert'
	// rootCmd.AddCommand(newCreateStepCommentCmd())  // Replaced by 'misc comment'
	// rootCmd.AddCommand(newCreateStepAssertLessThanOrEqualCmd())
	// rootCmd.AddCommand(newCreateStepAssertLessThanCmd())
	// rootCmd.AddCommand(newCreateStepAssertSelectedCmd())
	// rootCmd.AddCommand(newCreateStepAssertVariableCmd())
	// rootCmd.AddCommand(newCreateStepDismissConfirmCmd())  // Replaced by 'dialog dismiss confirm'
	// rootCmd.AddCommand(newCreateStepDismissPromptCmd())   // Replaced by 'dialog dismiss prompt'
	// rootCmd.AddCommand(newCreateStepDeleteCookieCmd()) // Replaced by 'data cookie delete'
	// rootCmd.AddCommand(newCreateStepMouseDownCmd())    // Replaced by 'mouse down'
	// rootCmd.AddCommand(newCreateStepMouseUpCmd())      // Replaced by 'mouse up'
	// rootCmd.AddCommand(newCreateStepMouseMoveCmd())    // Replaced by 'mouse move'
	// rootCmd.AddCommand(newCreateStepMouseEnterCmd())   // Replaced by 'mouse enter'
	// rootCmd.AddCommand(newCreateStepScrollPositionCmd())  // Replaced by 'navigate scroll-to'
	// rootCmd.AddCommand(newCreateStepSwitchIframeCmd())      // Replaced by 'window switch iframe'
	// rootCmd.AddCommand(newCreateStepSwitchNextTabCmd())     // Replaced by 'window switch tab next'
	// rootCmd.AddCommand(newCreateStepSwitchParentFrameCmd()) // Replaced by 'window switch parent-frame'
	// rootCmd.AddCommand(newCreateStepSwitchPrevTabCmd())     // Replaced by 'window switch tab prev'

	// ===== VERSION B ENHANCED COMMANDS =====
	// Cookie management commands (Version B) - Replaced by 'data cookie' commands
	// rootCmd.AddCommand(newCreateStepCookieCreateCmd())  // Replaced by 'data cookie create'
	// rootCmd.AddCommand(newCreateStepCookieWipeAllCmd()) // Replaced by 'data cookie clear-all'

	// Upload and dismiss commands (Version B)
	// rootCmd.AddCommand(newCreateStepUploadURLCmd())  // Replaced by 'file upload-url'
	// rootCmd.AddCommand(newCreateStepDismissPromptWithTextCmd())  // Replaced by 'dialog dismiss prompt-with-text'

	// Execute script command (Version B)
	// rootCmd.AddCommand(newCreateStepExecuteScriptCmd())  // Replaced by 'misc execute'

	// Enhanced mouse commands (Version B)
	// rootCmd.AddCommand(newCreateStepMouseMoveToCmd())  // Replaced by 'mouse move-to'
	// rootCmd.AddCommand(newCreateStepMouseMoveByCmd())  // Replaced by 'mouse move-by'

	// Enhanced pick commands (Version B) - Replaced by 'select' commands
	// rootCmd.AddCommand(newCreateStepPickIndexCmd())  // Replaced by 'select index'
	// rootCmd.AddCommand(newCreateStepPickLastCmd())   // Replaced by 'select last'

	// Enhanced wait commands (Version B) - Replaced by 'wait' command
	// rootCmd.AddCommand(newCreateStepWaitForElementTimeoutCmd())  // Replaced by 'wait element --timeout'
	// rootCmd.AddCommand(newCreateStepWaitForElementDefaultCmd())  // Replaced by 'wait element'

	// Enhanced store commands (Version B) - Replaced by 'data store' commands
	// rootCmd.AddCommand(newCreateStepStoreElementTextCmd())  // Replaced by 'data store element-text'
	// rootCmd.AddCommand(newCreateStepStoreLiteralValueCmd()) // Replaced by 'data store literal'

	// Scroll commands (Version B)
	// Note: These are already registered above as:
	// - newCreateStepScrollPositionCmd()
	// - newCreateStepScrollTopCmd()
	// - newCreateStepScrollBottomCmd()
	// - newCreateStepScrollElementCmd()

	// Window resize command (Version B)
	// rootCmd.AddCommand(newCreateStepWindowResizeCmd())  // Replaced by 'window resize'

	// ========================================
	// EXECUTION AND ANALYSIS COMMANDS
	// ========================================
	rootCmd.AddCommand(newExecuteGoalCmd())
	rootCmd.AddCommand(newMonitorExecutionCmd())
	rootCmd.AddCommand(newGetExecutionAnalysisCmd())
	rootCmd.AddCommand(newManageTestDataCmd())
	rootCmd.AddCommand(newCreateEnvironmentCmd())
}

// registerLegacyCommands registers all legacy commands with deprecation warnings
func registerLegacyCommands(rootCmd *cobra.Command) {
	// Get all legacy commands
	legacyCommands := GetLegacyCommands()

	// Add each legacy command to root
	for _, cmd := range legacyCommands {
		rootCmd.AddCommand(cmd)
	}
}
