package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepDismissPromptCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-dismiss-prompt TEXT [POSITION]",
		Short: "Create a dismiss prompt step at a specific position in a checkpoint",
		Long: `Create a dismiss prompt step that dismisses a JavaScript prompt dialog with the specified text at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-dismiss-prompt "John Doe" 1
  api-cli create-step-dismiss-prompt "user@example.com"  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli create-step-dismiss-prompt "John Doe" 1 --checkpoint 1678318
  
  # Legacy format (still supported)
  api-cli create-step-dismiss-prompt 1678318 "John Doe" 1`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx *StepContext
			var text string
			var err error
			
			// Check for legacy format (3 args where first is checkpoint ID)
			if len(args) == 3 {
				// Try to parse first arg as checkpoint ID
				if checkpointID, parseErr := strconv.Atoi(args[0]); parseErr == nil {
					// Legacy format detected
					text = args[1]
					position, posErr := strconv.Atoi(args[2])
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
					// Not legacy format, treat as modern format error
					return fmt.Errorf("invalid arguments: expected TEXT [POSITION] or CHECKPOINT_ID TEXT POSITION")
				}
			} else {
				// Modern format - text is first argument
				text = args[0]
				
				// Resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 1)
				if err != nil {
					return err
				}
			}
			
			// Validate text
			if text == "" {
				return fmt.Errorf("text cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create dismiss prompt step using the enhanced client
			stepID, err := client.CreateDismissPromptStep(ctx.CheckpointID, text, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create dismiss prompt step: %w", err)
			}
			
			// Save config if position was auto-incremented
			saveStepContext(ctx)
			
			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "DISMISS_PROMPT",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Dismiss prompt with text: %s", text),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"text": text},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}