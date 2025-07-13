#!/bin/bash

# Script to help integrate Version B command registrations into Version A's main.go

echo "Creating command registration patch..."

cat > /tmp/command-registrations.txt << 'EOF'
// Add these command registrations to the init() function in main.go

	// === VERSION B COMMAND REGISTRATIONS ===
	
	// Cookie management commands
	rootCmd.AddCommand(newCreateStepCookieCreateCmd())
	rootCmd.AddCommand(newCreateStepCookieWipeAllCmd())
	
	// Upload and dismiss commands
	rootCmd.AddCommand(newCreateStepUploadURLCmd())
	rootCmd.AddCommand(newCreateStepDismissPromptWithTextCmd())
	
	// Execute script command
	rootCmd.AddCommand(newCreateStepExecuteScriptCmd())
	
	// Enhanced mouse commands
	rootCmd.AddCommand(newCreateStepMouseMoveToCmd())
	rootCmd.AddCommand(newCreateStepMouseMoveByCmd())
	
	// Enhanced pick commands
	rootCmd.AddCommand(newCreateStepPickIndexCmd())
	rootCmd.AddCommand(newCreateStepPickLastCmd())
	
	// Enhanced wait commands
	rootCmd.AddCommand(newCreateStepWaitForElementTimeoutCmd())
	rootCmd.AddCommand(newCreateStepWaitForElementDefaultCmd())
	
	// Enhanced store commands
	rootCmd.AddCommand(newCreateStepStoreElementTextCmd())
	rootCmd.AddCommand(newCreateStepStoreLiteralValueCmd())
	
	// Switch/frame commands (if not already present)
	rootCmd.AddCommand(newCreateStepSwitchIframeCmd())
	rootCmd.AddCommand(newCreateStepSwitchNextTabCmd())
	rootCmd.AddCommand(newCreateStepSwitchParentFrameCmd())
	rootCmd.AddCommand(newCreateStepSwitchPrevTabCmd())
	
	// Assertion commands (enhanced versions)
	rootCmd.AddCommand(newCreateStepAssertNotEqualsCmd())
	rootCmd.AddCommand(newCreateStepAssertGreaterThanCmd())
	rootCmd.AddCommand(newCreateStepAssertGreaterThanOrEqualCmd())
	rootCmd.AddCommand(newCreateStepAssertMatchesCmd())
	
	// New navigation commands (enhanced version)
	rootCmd.AddCommand(newCreateStepNavigateCmd())
	
	// New interaction commands (enhanced versions)
	rootCmd.AddCommand(newCreateStepClickCmd())
	rootCmd.AddCommand(newCreateStepWriteCmd())
	
	// New scroll commands
	rootCmd.AddCommand(newCreateStepScrollToPositionCmd())
	rootCmd.AddCommand(newCreateStepScrollByOffsetCmd())
	rootCmd.AddCommand(newCreateStepScrollToTopCmd())
	
	// New window commands
	rootCmd.AddCommand(newCreateStepWindowResizeCmd())
	
	// New keyboard commands (enhanced version)
	rootCmd.AddCommand(newCreateStepKeyCmd())
	
	// New documentation commands
	rootCmd.AddCommand(newCreateStepCommentCmd())
EOF

echo "Command registrations saved to /tmp/command-registrations.txt"
echo ""
echo "Next steps:"
echo "1. Open /Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/main.go"
echo "2. Find the init() function"
echo "3. Add the command registrations from /tmp/command-registrations.txt"
echo "4. Make sure to avoid duplicates if some commands already exist"
echo ""
echo "Reference file available at:"
echo "  /Users/marklovelady/_dev/virtuoso-api-cli-generator/merge-helpers/main-version-b.go"