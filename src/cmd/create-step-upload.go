package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepUploadCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-upload ELEMENT FILE_PATH [POSITION]",
		Short: "Create a file upload step at a specific position in a checkpoint",
		Long: `Create a file upload step that uploads a file to a specific element at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-upload "file upload" "document.pdf" 1
  api-cli create-step-upload "#file-input" "image.jpg"  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli create-step-upload "file upload" "document.pdf" 1 --checkpoint 1678318
  
  # Legacy syntax (still supported)
  api-cli create-step-upload 1678318 "document.pdf" "file upload" 1`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var element, filename string
			var ctx *StepContext
			var err error
			
			// Detect legacy syntax (first arg is numeric checkpoint ID)
			if len(args) == 4 {
				// Try to parse first arg as checkpoint ID
				if checkpointID, parseErr := strconv.Atoi(args[0]); parseErr == nil {
					// Legacy format: CHECKPOINT_ID FILENAME ELEMENT POSITION
					filename = args[1]
					element = args[2]
					position, posErr := strconv.Atoi(args[3])
					if posErr != nil {
						return fmt.Errorf("invalid position: %w", posErr)
					}
					ctx = &StepContext{
						CheckpointID: checkpointID,
						Position:     position,
						UsingContext: false,
						AutoPosition: false,
					}
				} else {
					// Not legacy format, treat as modern format with all args
					element = args[0]
					filename = args[1]
					ctx, err = resolveStepContext(args[2:], checkpointFlag, 0)
					if err != nil {
						return err
					}
				}
			} else {
				// Modern format: ELEMENT FILE_PATH [POSITION]
				element = args[0]
				filename = args[1]
				ctx, err = resolveStepContext(args[2:], checkpointFlag, 0)
				if err != nil {
					return err
				}
			}
			
			// Validate inputs
			if filename == "" {
				return fmt.Errorf("filename cannot be empty")
			}
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create upload step using the enhanced client
			stepID, err := client.CreateUploadStep(ctx.CheckpointID, filename, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create upload step: %w", err)
			}
			
			// Save config if position was auto-incremented
			saveStepContext(ctx)
			
			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "UPLOAD",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("upload \"%s\" to %s", filename, element),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"filename": filename,
					"element":  element,
				},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}