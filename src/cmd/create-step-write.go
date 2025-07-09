package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepWriteCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-write TEXT ELEMENT [POSITION]",
		Short: "Create a text input step at a specific position in a checkpoint",
		Long: `Create a text input step that types text into a specified element at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-write "john@example.com" "email field" 1
  api-cli create-step-write "password123" "#password-input"  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli create-step-write "john@example.com" "email field" 1 --checkpoint 1678318`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			element := args[1]
			
			// Validate inputs
			if text == "" {
				return fmt.Errorf("text cannot be empty")
			}
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Resolve checkpoint and position
			ctx, err := resolveStepContext(args, checkpointFlag, 2)
			if err != nil {
				return err
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create write step using the enhanced client
			stepID, err := client.CreateWriteStep(ctx.CheckpointID, text, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create write step: %w", err)
			}
			
			// Save config if position was auto-incremented
			saveStepContext(ctx)
			
			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "WRITE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("type \"%s\" in %s", text, element),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"text": text, "element": element},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}