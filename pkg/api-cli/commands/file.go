package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/commands/shared"
	"github.com/spf13/cobra"
)

// FileCmd creates the file command with subcommands
func FileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Handle file upload operations",
		Long: `Manage file upload operations including uploading local files and files from URLs.

This command consolidates all file upload related actions.`,
	}

	// Add subcommands
	cmd.AddCommand(uploadSubCmd())
	cmd.AddCommand(uploadURLSubCmd())

	return cmd
}

// uploadSubCmd creates the upload subcommand for local files
func uploadSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload [checkpoint-id] <selector> <file-path> [position]",
		Short: "Upload a local file",
		Long: `Upload a local file to a specified element.

The file must exist on the local filesystem. The selector identifies
the upload element (typically an input[type="file"] element).

Examples:
  # Using session context (modern)
  api-cli file upload "#file-input" "/path/to/document.pdf"
  api-cli file upload "Upload Resume" "resume.docx" 5

  # Using explicit checkpoint (legacy)
  api-cli file upload cp_12345 "#file-input" "/path/to/document.pdf" 1
  api-cli file upload 12345 "Upload File" "image.jpg" 2`,
		Args: cobra.MinimumNArgs(2),
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
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileCommand(cmd, args, "upload-url", nil)
		},
	}

	return cmd
}

// runFileCommand executes a file command
func runFileCommand(cmd *cobra.Command, args []string, action string, options map[string]interface{}) error {
	base := shared.NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	// Both upload and upload-url require 2 args: selector and file-path/url
	args, err := base.ResolveCheckpointAndPosition(args, 2)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Extract arguments
	selector := args[0]
	filePathOrURL := args[1]

	// Validate selector
	if err := shared.ValidateSelector(selector); err != nil {
		return err
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
	result := &shared.StepResult{
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

// executeUploadAction executes a local file upload
func executeUploadAction(c *client.Client, checkpointID int, selector, filePath string, position int) (int, error) {
	// Validate file existence
	if err := validateFilePath(filePath); err != nil {
		return 0, err
	}

	// Use the appropriate client method based on the existing pattern
	// The client has multiple upload methods, using CreateStepUpload which matches the pattern
	return c.CreateStepUpload(checkpointID, selector, filePath, position)
}

// executeUploadURLAction executes a URL file upload
func executeUploadURLAction(c *client.Client, checkpointID int, selector, url string, position int) (int, error) {
	// Validate URL format
	if err := shared.ValidateURL(url); err != nil {
		return 0, err
	}

	// Use the CreateStepUploadURL method from the client
	return c.CreateStepUploadURL(checkpointID, url, selector, position)
}

// validateFilePath validates that a file exists
func validateFilePath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("error accessing file: %w", err)
	}

	// Check if it's a regular file (not a directory)
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	return nil
}
