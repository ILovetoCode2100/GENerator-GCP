package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertNotEqualsCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-assert-not-equals ELEMENT VALUE [POSITION]",
		Short: "Create an assertion step that verifies an element does not equal a specific value at a specific position",
		Long: `Create an assertion step that verifies an element does not equal a specific value at the specified position in the checkpoint.

Session Context:
  Uses current checkpoint from session context. Set with 'api-cli set-checkpoint CHECKPOINT_ID'
  Position auto-increments if not specified.

Examples:
  # Using session context (recommended)
  api-cli set-checkpoint 1678318
  api-cli create-step-assert-not-equals "Status" "Failed"          # Auto-increment position
  api-cli create-step-assert-not-equals "#status" "Error" 2         # Explicit position
  
  # Override checkpoint for specific step
  api-cli create-step-assert-not-equals "Status" "Failed" 1 --checkpoint 1678319`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			element := args[0]
			value := args[1]
			
			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Validate value
			if value == "" {
				return fmt.Errorf("value cannot be empty")
			}
			
			// Resolve checkpoint and position using session context
			ctx, err := resolveStepContext(args, checkpointFlag, 2)
			if err != nil {
				return err
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create assert not equals step using the client
			stepID, err := client.CreateAssertNotEqualsStep(ctx.CheckpointID, element, value, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create assert not equals step: %w", err)
			}
			
			// Save session state if position was auto-incremented
			saveStepContext(ctx)
			
			// Prepare output
			output := &StepOutput{
				Status:       "success",
				StepType:     "ASSERT_NOT_EQUALS",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("expect %s to not equal \"%s\"", element, value),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"element": element, "value": value},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	return cmd
}