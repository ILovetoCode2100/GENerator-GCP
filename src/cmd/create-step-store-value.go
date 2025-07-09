package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepStoreValueCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-store-value VALUE VARIABLE_NAME [POSITION]",
		Short: "Create a store value step at a specific position in a checkpoint",
		Long: `Create a store value step that stores a specific value in a variable at the specified position in the checkpoint.
		
Modern usage (with session context):
  api-cli set-checkpoint 1678318
  api-cli create-step-store-value "test@example.com" "email"
  api-cli create-step-store-value "12345" "user_id" 2
  api-cli create-step-store-value "admin" "userRole" --checkpoint 1678319

Legacy usage:
  api-cli create-step-store-value 1678318 "test@example.com" "email" 1
  api-cli create-step-store-value 1678318 "12345" "user_id" 2 -o json`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var value, variableName string
			var ctx *StepContext
			var err error
			
			// Handle both modern and legacy patterns
			if len(args) == 4 {
				// Legacy: CHECKPOINT_ID VALUE VARIABLE_NAME POSITION
				checkpointID, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid checkpoint ID: %w", err)
				}
				position, err := strconv.Atoi(args[3])
				if err != nil {
					return fmt.Errorf("invalid position: %w", err)
				}
				value = args[1]
				variableName = args[2]
				ctx = &StepContext{
					CheckpointID: checkpointID,
					Position:     position,
					UsingContext: false,
					AutoPosition: false,
				}
			} else {
				// Modern: VALUE VARIABLE_NAME [POSITION]
				value = args[0]
				variableName = args[1]
				ctx, err = resolveStepContext(args, checkpointFlag, 2)
				if err != nil {
					return err
				}
			}
			
			// Validate value and variable name
			if value == "" {
				return fmt.Errorf("value cannot be empty")
			}
			if variableName == "" {
				return fmt.Errorf("variable name cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create store value step using the enhanced client
			stepID, err := client.CreateStoreValueStep(ctx.CheckpointID, value, variableName, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create store value step: %w", err)
			}
			
			// Save session context if position was auto-incremented
			saveStepContext(ctx)
			
			// Output the result
			output := &StepOutput{
				Status:       "success",
				StepType:     "STORE_VALUE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Store value \"%s\" in variable \"%s\"", value, variableName),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"value":         value,
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