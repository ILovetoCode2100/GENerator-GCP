package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// ========================================
// Utility Step Commands
// ========================================
// This file consolidates utility step commands including:
// - File operations (upload from URL)
// - Miscellaneous operations (comments, scripts)

// ========================================
// File Upload Commands
// ========================================

// StepFileCmd creates the file command with subcommands
func StepFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step-file",
		Short: "Handle file upload operations",
		Long: `Manage file upload operations including uploading local files and files from URLs.

This command consolidates all file upload related actions.`,
	}

	// Add subcommands
	cmd.AddCommand(uploadSubCmd())
	cmd.AddCommand(uploadURLSubCmd())

	return cmd
}

// uploadSubCmd creates the upload subcommand for file URLs
func uploadSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload [checkpoint-id] <selector> <file-url> [position]",
		Short: "Upload a file from URL",
		Long: `Upload a file from a URL to a specified element.

The URL must be accessible and point to a valid file. The selector identifies
the upload element (typically an input[type="file"] element).

Note: Local file paths are not supported. Files must be accessible via HTTP/HTTPS URLs.

Examples:
  # Using session context (modern)
  api-cli file upload "#file-input" "https://example.com/document.pdf"
  api-cli file upload "Upload Resume" "https://example.com/resume.docx" 5

  # Using explicit checkpoint (legacy)
  api-cli file upload cp_12345 "#file-input" "https://example.com/document.pdf" 1
  api-cli file upload 12345 "Upload File" "https://example.com/image.jpg" 2`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("requires exactly 2 arguments: <selector> <url>\n\nExamples:\n  api-cli step-file upload \"#file-input\" \"https://example.com/file.pdf\"\n  api-cli step-file upload \"input[type=file]\" \"https://example.com/document.docx\"")
			}
			if len(args) > 4 {
				return fmt.Errorf("too many arguments provided. Expected: <selector> <url> [position]")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileCommand(cmd, args, "upload", nil)
		},
	}

	return cmd
}

// uploadURLSubCmd creates the upload-url subcommand
func uploadURLSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload-url [checkpoint-id] <selector> <url> [position]",
		Short: "Upload a file from URL",
		Long: `Upload a file from a URL to a specified element.

The URL must be accessible and point to a valid file. The selector
identifies the upload element (typically an input[type="file"] element).

Examples:
  # Using session context (modern)
  api-cli file upload-url "#file-input" "https://example.com/document.pdf"
  api-cli file upload-url "Upload Resume" "https://example.com/resume.docx" 5

  # Using explicit checkpoint (legacy)
  api-cli file upload-url cp_12345 "#file-input" "https://example.com/document.pdf" 1
  api-cli file upload-url 12345 "Upload File" "https://example.com/image.jpg" 2`,
		Aliases: []string{"url"},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("requires exactly 2 arguments: <selector> <url>\n\nExamples:\n  api-cli step-file upload-url \"#file-input\" \"https://example.com/file.pdf\"\n  api-cli step-file upload-url \"input[type=file]\" \"https://example.com/document.docx\"")
			}
			if len(args) > 4 {
				return fmt.Errorf("too many arguments provided. Expected: <selector> <url> [position]")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileCommand(cmd, args, "upload-url", nil)
		},
	}

	return cmd
}

// runFileCommand executes a file command
func runFileCommand(cmd *cobra.Command, args []string, action string, options map[string]interface{}) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return fmt.Errorf("failed to initialize %s command: %w", action, err)
	}

	// Resolve checkpoint and position
	// Both upload and upload-url require 2 args: selector and file-path/url
	args, err := base.ResolveCheckpointAndPosition(args, 2)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate we have enough arguments
	if len(args) < 2 {
		return fmt.Errorf("upload requires both selector and URL arguments. Usage: api-cli step-file upload <selector> <url>")
	}

	// Extract arguments
	selector := args[0]
	filePathOrURL := args[1]

	// Validate selector
	if err := ValidateSelector(selector); err != nil {
		return err
	}

	// Validate URL format early
	if err := ValidateURL(filePathOrURL); err != nil {
		return fmt.Errorf("invalid URL: %w\n\nFile upload requires a valid HTTP/HTTPS URL. Local file paths are not supported.\n\nExample: https://example.com/document.pdf", err)
	}

	// Convert checkpoint ID to int
	checkpointID, err := strconv.Atoi(base.CheckpointID)
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	// Execute the appropriate action
	var stepID int
	switch action {
	case "upload":
		stepID, err = executeUploadAction(base.Client, checkpointID, selector, filePathOrURL, base.Position)
	case "upload-url":
		stepID, err = executeUploadURLAction(base.Client, checkpointID, selector, filePathOrURL, base.Position)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", action, err)
	}

	// Format and output the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: base.CheckpointID,
		Type:         "UPLOAD",
		Position:     base.Position,
		Selector:     selector,
		Value:        filePathOrURL,
		Description:  fmt.Sprintf("Upload %s to %s", filePathOrURL, selector),
		Meta: map[string]interface{}{
			"action": action,
		},
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// executeUploadAction executes a file upload from URL
func executeUploadAction(c *client.Client, checkpointID int, selector, fileURL string, position int) (int, error) {
	// Validate URL format
	if err := ValidateURL(fileURL); err != nil {
		return 0, err
	}

	// Use the appropriate client method based on the existing pattern
	// The client has multiple upload methods, using CreateStepUpload which matches the pattern
	return c.CreateStepUpload(checkpointID, selector, fileURL, position)
}

// executeUploadURLAction executes a URL file upload
func executeUploadURLAction(c *client.Client, checkpointID int, selector, url string, position int) (int, error) {
	// Validate URL format
	if err := ValidateURL(url); err != nil {
		return 0, err
	}

	// Use the CreateStepUploadURL method from the client
	return c.CreateStepUploadURL(checkpointID, url, selector, position)
}

// ========================================
// Miscellaneous Commands
// ========================================

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
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("comment text is required\n\nExamples:\n  api-cli step-misc comment \"This is a test comment\"\n  api-cli step-misc comment \"TODO: Add validation\"")
			}
			return nil
		},
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
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("JavaScript code is required\n\nExamples:\n  api-cli step-misc execute \"console.log('Hello')\"\n  api-cli step-misc execute \"document.title = 'New Title'\"")
			}
			return nil
		},
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
		return fmt.Errorf("failed to initialize %s command: %w", operation, err)
	}

	// Resolve checkpoint and position
	remainingArgs, err := base.ResolveCheckpointAndPosition(args, 1)
	if err != nil {
		return fmt.Errorf("failed to resolve checkpoint and position for %s: %w", operation, err)
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
		checkpointID, err := parseCheckpointID(base.CheckpointID)
		if err != nil {
			return fmt.Errorf("invalid checkpoint ID: %w", err)
		}
		stepID, err = base.Client.CreateStepComment(checkpointID, mainArg, base.Position)
		if err != nil {
			return fmt.Errorf("failed to create comment step: %w", err)
		}

	case "execute":
		stepType = "EXECUTE_SCRIPT"
		checkpointID, err := parseCheckpointID(base.CheckpointID)
		if err != nil {
			return fmt.Errorf("invalid checkpoint ID: %w", err)
		}
		stepID, err = base.Client.CreateStepExecuteScript(checkpointID, mainArg, base.Position)
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
		return fmt.Errorf("failed to format %s output: %w", operation, err)
	}
	fmt.Print(output)

	return nil
}

// ========================================
// Utility Functions
// ========================================

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
