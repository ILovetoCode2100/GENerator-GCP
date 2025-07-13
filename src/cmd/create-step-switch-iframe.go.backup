package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepSwitchIFrameCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-switch-iframe SELECTOR [POSITION]",
		Short: "Create a switch to iframe step at a specific position in a checkpoint",
		Long: `Create a switch to iframe step that switches to an iframe by element selector at the specified position in the checkpoint.

Modern usage (with session context):
  api-cli set-checkpoint 1678318
  api-cli create-step-switch-iframe "#content-frame"
  api-cli create-step-switch-iframe "search iframe" 2
  api-cli create-step-switch-iframe "#frame" --checkpoint 1678319

Legacy usage (backward compatible):
  api-cli create-step-switch-iframe 1678318 "#content-frame" 1
  api-cli create-step-switch-iframe 1678318 "search iframe" 2 -o json`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var selector string
			var ctx *StepContext
			var err error
			
			// Handle both modern and legacy argument patterns
			if len(args) == 3 {
				// Legacy: CHECKPOINT_ID SELECTOR POSITION
				checkpointID, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid checkpoint ID: %w", err)
				}
				selector = args[1]
				position, err := strconv.Atoi(args[2])
				if err != nil {
					return fmt.Errorf("invalid position: %w", err)
				}
				
				ctx = &StepContext{
					CheckpointID: checkpointID,
					Position:     position,
					UsingContext: false,
					AutoPosition: false,
				}
			} else {
				// Modern: SELECTOR [POSITION]
				selector = args[0]
				
				// Use helper to resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 1) // position at index 1
				if err != nil {
					return err
				}
			}
			
			// Validate selector
			if selector == "" {
				return fmt.Errorf("selector cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create switch iframe step using the enhanced client
			stepID, err := client.CreateSwitchIFrameStep(ctx.CheckpointID, selector, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create switch iframe step: %w", err)
			}
			
			// Save session context if needed
			saveStepContext(ctx)
			
			// Output result using the unified format
			output := &StepOutput{
				Status:       "success",
				StepType:     "SWITCH",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("switch to iframe by element: %s", selector),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"selector": selector,
				},
			}
			
			return outputStepResult(output)
		},
	}
	
	// Add the checkpoint flag
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}