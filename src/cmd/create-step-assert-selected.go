package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertSelectedCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-assert-selected ELEMENT [POSITION]",
		Short: "Create an assert selected step at a specific position in a checkpoint",
		Long: `Create an assert selected step that verifies an option is selected at the specified position in the checkpoint.

Session Context:
  Uses current checkpoint from session context. Set with 'api-cli set-checkpoint CHECKPOINT_ID'
  Position auto-increments if not specified.

Examples:
  # Using session context (recommended)
  api-cli set-checkpoint 1678318
  api-cli create-step-assert-selected "Country dropdown"          # Auto-increment position
  api-cli create-step-assert-selected "Option 2" 2                # Explicit position
  
  # Override checkpoint for specific step
  api-cli create-step-assert-selected "Country dropdown" 1 --checkpoint 1678319`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			element := args[0]
			
			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Resolve checkpoint and position using session context
			ctx, err := resolveStepContext(args, checkpointFlag, 1)
			if err != nil {
				return err
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create assert selected step using the enhanced client
			stepID, err := client.CreateAssertSelectedStep(ctx.CheckpointID, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create assert selected step: %w", err)
			}
			
			// Save session state if position was auto-incremented
			saveStepContext(ctx)
			
			// Prepare output
			output := &StepOutput{
				Status:       "success",
				StepType:     "ASSERT_SELECTED",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("see %s is selected", element),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"element": element},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	return cmd
}
