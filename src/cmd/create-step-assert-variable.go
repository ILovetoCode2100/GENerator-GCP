package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertVariableCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-assert-variable VARIABLE_NAME EXPECTED_VALUE [POSITION]",
		Short: "Create an assert variable step at a specific position in a checkpoint",
		Long: `Create an assert variable step that verifies a stored variable has the expected value at the specified position in the checkpoint.

Session Context:
  Uses current checkpoint from session context. Set with 'api-cli set-checkpoint CHECKPOINT_ID'
  Position auto-increments if not specified.

Examples:
  # Using session context (recommended)
  api-cli set-checkpoint 1678318
  api-cli create-step-assert-variable "orderId" "12345"          # Auto-increment position
  api-cli create-step-assert-variable "username" "john.doe" 2     # Explicit position
  
  # Override checkpoint for specific step
  api-cli create-step-assert-variable "orderId" "12345" 1 --checkpoint 1678319`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			variableName := args[0]
			expectedValue := args[1]
			
			// Validate inputs
			if variableName == "" {
				return fmt.Errorf("variable name cannot be empty")
			}
			if expectedValue == "" {
				return fmt.Errorf("expected value cannot be empty")
			}
			
			// Resolve checkpoint and position using session context
			ctx, err := resolveStepContext(args, checkpointFlag, 2)
			if err != nil {
				return err
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create assert variable step using the enhanced client
			stepID, err := client.CreateAssertVariableStep(ctx.CheckpointID, variableName, expectedValue, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create assert variable step: %w", err)
			}
			
			// Save session state if position was auto-incremented
			saveStepContext(ctx)
			
			// Prepare output
			output := &StepOutput{
				Status:       "success",
				StepType:     "ASSERT_VARIABLE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("expect $%s to equal \"%s\"", variableName, expectedValue),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"variable_name": variableName, "expected_value": expectedValue},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	return cmd
}
