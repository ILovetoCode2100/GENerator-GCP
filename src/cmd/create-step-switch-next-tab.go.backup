package main

import (
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepSwitchNextTabCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-switch-next-tab [POSITION]",
		Short: "Create a switch to next tab step at a specific position in a checkpoint",
		Long: `Create a switch to next tab step that switches to the next browser tab at the specified position in the checkpoint.

Modern usage (with session context):
  api-cli set-checkpoint 1678318
  api-cli create-step-switch-next-tab
  api-cli create-step-switch-next-tab 2
  api-cli create-step-switch-next-tab --checkpoint 1678319

Legacy usage (backward compatible):
  api-cli create-step-switch-next-tab 1678318 1
  api-cli create-step-switch-next-tab 1678318 2 -o json`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx *StepContext
			var err error
			
			// Handle both modern and legacy argument patterns
			if len(args) == 2 {
				// Legacy: CHECKPOINT_ID POSITION
				checkpointID, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				position, err := strconv.Atoi(args[1])
				if err != nil {
					return err
				}
				
				ctx = &StepContext{
					CheckpointID: checkpointID,
					Position:     position,
					UsingContext: false,
					AutoPosition: false,
				}
			} else {
				// Modern: [POSITION]
				// Use helper to resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 0) // position at index 0
				if err != nil {
					return err
				}
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create switch next tab step using the enhanced client
			stepID, err := client.CreateSwitchNextTabStep(ctx.CheckpointID, ctx.Position)
			if err != nil {
				return err
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
				ParsedStep:   "switch to next tab",
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
			}
			
			return outputStepResult(output)
		},
	}
	
	// Add the checkpoint flag
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}