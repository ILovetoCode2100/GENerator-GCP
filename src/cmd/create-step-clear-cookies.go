package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepClearCookiesCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-clear-cookies [POSITION]",
		Short: "Create a clear all cookies step at a specific position in a checkpoint",
		Long: `Create a clear all cookies step that removes all cookies at the specified position in the checkpoint.

Modern usage (with session context):
  api-cli set-checkpoint 1678318
  api-cli create-step-clear-cookies
  api-cli create-step-clear-cookies 2
  api-cli create-step-clear-cookies --checkpoint 1678319

Legacy usage (backward compatible):
  api-cli create-step-clear-cookies 1678318 1
  api-cli create-step-clear-cookies 1678318 2 -o json`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine if using legacy or modern syntax
			var ctx *StepContext
			var err error
			
			// Check for legacy syntax (first arg is numeric checkpoint ID)
			if len(args) >= 2 {
				if _, parseErr := strconv.Atoi(args[0]); parseErr == nil {
					// Legacy syntax: CHECKPOINT_ID POSITION
					checkpointID, err := strconv.Atoi(args[0])
					if err != nil {
						return fmt.Errorf("invalid checkpoint ID: %w", err)
					}
					
					position, err := strconv.Atoi(args[1])
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
					// Modern syntax with explicit position: [POSITION]
					positionIndex := 0
					ctx, err = resolveStepContext(args, checkpointFlag, positionIndex)
					if err != nil {
						return err
					}
				}
			} else {
				// Modern syntax: no args or just position
				positionIndex := 0
				ctx, err = resolveStepContext(args, checkpointFlag, positionIndex)
				if err != nil {
					return err
				}
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create clear cookies step using the enhanced client
			stepID, err := client.CreateClearCookiesStep(ctx.CheckpointID, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create clear cookies step: %w", err)
			}
			
			// Save session context if position was auto-incremented
			saveStepContext(ctx)
			
			// Create step output
			output := &StepOutput{
				Status:       "success",
				StepType:     "CLEAR_COOKIES",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   "clear all cookies",
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        nil,
			}
			
			// Output the result
			return outputStepResult(output)
		},
	}
	
	// Add checkpoint flag
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}