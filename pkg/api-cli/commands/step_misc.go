package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// newStepMiscCmd creates the misc command with subcommands
func newStepMiscCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step-misc",
		Short: "Miscellaneous operations (comments, scripts)",
		Long: `Miscellaneous operations including adding comments and executing scripts.

This command consolidates various utility operations that don't fit into other categories.`,
	}

	// Add subcommands
	cmd.AddCommand(commentSubCmd())
	cmd.AddCommand(executeSubCmd())

	return cmd
}

// commentSubCmd creates the comment subcommand
func commentSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment [checkpoint-id] <text> [position]",
		Short: "Add a comment to the test",
		Long: `Add a comment to document test steps, mark TODOs, or provide context.

Examples:
  # Using session context (modern)
  api-cli misc comment "This is a test comment"
  api-cli misc comment "TODO: Add validation for edge cases"
  api-cli misc comment "FIXME: Handle timeout errors"

  # Using explicit checkpoint (legacy)
  api-cli misc comment cp_12345 "Login validation step" 1
  api-cli misc comment 1678318 "Check password requirements" 2

Comments support multi-word text - all arguments after the checkpoint/position are joined.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMiscOperation(cmd, args, "comment", nil)
		},
	}

	return cmd
}

// executeSubCmd creates the execute script subcommand
func executeSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute [checkpoint-id] <script> [position]",
		Short: "Execute JavaScript code",
		Long: `Execute custom JavaScript code in the browser context.

Examples:
  # Using session context (modern)
  api-cli misc execute "console.log('Hello World')"
  api-cli misc execute "document.title = 'New Title'"
  api-cli misc execute "localStorage.setItem('key', 'value')"

  # Using explicit checkpoint (legacy)
  api-cli misc execute cp_12345 "alert('Test')" 1
  api-cli misc execute 1678318 "window.scrollTo(0, 0)" 2

The script can be any valid JavaScript expression or statement.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMiscOperation(cmd, args, "execute", nil)
		},
	}

	return cmd
}

// runMiscOperation handles the execution of misc operations
func runMiscOperation(cmd *cobra.Command, args []string, operation string, options map[string]interface{}) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	remainingArgs, err := base.ResolveCheckpointAndPosition(args, 1)
	if err != nil {
		return err
	}

	// Extract the main argument based on operation
	var mainArg string
	switch operation {
	case "comment":
		// Join all remaining args for multi-word comments
		mainArg = strings.Join(remainingArgs, " ")
		if mainArg == "" {
			return fmt.Errorf("comment text cannot be empty")
		}

	case "execute":
		// Join all remaining args for multi-line scripts
		mainArg = strings.Join(remainingArgs, " ")
		if mainArg == "" {
			return fmt.Errorf("script cannot be empty")
		}
		// Basic validation for JavaScript
		if err := validateJavaScript(mainArg); err != nil {
			return fmt.Errorf("invalid JavaScript: %w", err)
		}

	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}

	// Call appropriate client method based on operation
	var stepID int
	var stepType string
	switch operation {
	case "comment":
		stepType = "COMMENT"
		stepID, err = base.Client.CreateStepComment(parseCheckpointID(base.CheckpointID), mainArg, base.Position)
		if err != nil {
			return fmt.Errorf("failed to create comment step: %w", err)
		}

	case "execute":
		stepType = "EXECUTE_SCRIPT"
		stepID, err = base.Client.CreateStepExecuteScript(parseCheckpointID(base.CheckpointID), mainArg, base.Position)
		if err != nil {
			return fmt.Errorf("failed to create execute script step: %w", err)
		}
	}

	// Create result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		Type:         stepType,
		Position:     base.Position,
		CheckpointID: base.CheckpointID,
		Description:  getOperationDescription(operation, mainArg),
		Meta: map[string]interface{}{
			"operation": operation,
			"content":   mainArg,
		},
	}

	// Format and output result
	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return err
	}
	fmt.Print(output)

	return nil
}

// validateJavaScript performs basic JavaScript validation
func validateJavaScript(script string) error {
	// Basic validation - ensure it's not empty and doesn't contain obvious issues
	script = strings.TrimSpace(script)
	if script == "" {
		return fmt.Errorf("script cannot be empty")
	}

	// Check for common syntax issues
	openBraces := strings.Count(script, "{")
	closeBraces := strings.Count(script, "}")
	if openBraces != closeBraces {
		return fmt.Errorf("mismatched braces: %d open, %d close", openBraces, closeBraces)
	}

	openParens := strings.Count(script, "(")
	closeParens := strings.Count(script, ")")
	if openParens != closeParens {
		return fmt.Errorf("mismatched parentheses: %d open, %d close", openParens, closeParens)
	}

	// Check for unclosed strings (basic check)
	if strings.Count(script, "'")%2 != 0 {
		return fmt.Errorf("unclosed single quote")
	}
	if strings.Count(script, "\"")%2 != 0 {
		return fmt.Errorf("unclosed double quote")
	}

	return nil
}

// getOperationDescription returns a human-readable description for the operation
func getOperationDescription(operation, content string) string {
	switch operation {
	case "comment":
		// Truncate long comments for description
		if len(content) > 50 {
			return fmt.Sprintf("Comment: %s...", content[:47])
		}
		return fmt.Sprintf("Comment: %s", content)

	case "execute":
		// Truncate long scripts for description
		if len(content) > 50 {
			return fmt.Sprintf("Execute JavaScript: %s...", content[:47])
		}
		return fmt.Sprintf("Execute JavaScript: %s", content)

	default:
		return fmt.Sprintf("%s operation", operation)
	}
}

// parseCheckpointID converts a string checkpoint ID to an integer
func parseCheckpointID(id string) int {
	// Handle cp_ prefix
	if strings.HasPrefix(id, "cp_") {
		id = strings.TrimPrefix(id, "cp_")
	}

	// Parse to int
	checkpointID, _ := strconv.Atoi(id)
	return checkpointID
}
