package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepStoreCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-store ELEMENT VARIABLE_NAME [POSITION]",
		Short: "Create a store step at a specific position in a checkpoint",
		Long: `Create a store step that stores a value from an element into a variable at the specified position in the checkpoint.

Modern usage (with session context):
  api-cli set-checkpoint 1678318
  api-cli create-step-store "#user-id" "userId"
  api-cli create-step-store "User name field" "userName" 2
  api-cli create-step-store "#email" "userEmail" --checkpoint 1678319

Legacy usage:
  api-cli create-step-store 1678318 "#user-id" "userId" 1
  api-cli create-step-store 1678318 "User name field" "userName" 2 -o json`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var element, variableName string
			var ctx *StepContext
			var err error

			// Handle both modern and legacy patterns
			if len(args) == 4 {
				// Legacy: CHECKPOINT_ID ELEMENT VARIABLE_NAME POSITION
				checkpointID, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid checkpoint ID: %w", err)
				}
				position, err := strconv.Atoi(args[3])
				if err != nil {
					return fmt.Errorf("invalid position: %w", err)
				}
				element = args[1]
				variableName = args[2]
				ctx = &StepContext{
					CheckpointID: checkpointID,
					Position:     position,
					UsingContext: false,
					AutoPosition: false,
				}
			} else {
				// Modern: ELEMENT VARIABLE_NAME [POSITION]
				element = args[0]
				variableName = args[1]
				ctx, err = resolveStepContext(args, checkpointFlag, 2)
				if err != nil {
					return err
				}
			}

			// Validate element and variable name
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			if variableName == "" {
				return fmt.Errorf("variable name cannot be empty")
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create store step using the enhanced client
			stepID, err := client.CreateStoreStep(ctx.CheckpointID, element, variableName, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create store step: %w", err)
			}

			// Save session context if position was auto-incremented
			saveStepContext(ctx)

			// Output the result
			output := &StepOutput{
				Status:       "success",
				StepType:     "STORE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("store value from %s as $%s", element, variableName),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"element":       element,
					"variable_name": variableName,
				},
			}

			return outputStepResult(output)
		},
	}

	// Add the --checkpoint flag
	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
