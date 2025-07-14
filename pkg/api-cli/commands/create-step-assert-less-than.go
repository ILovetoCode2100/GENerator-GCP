package commands

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newCreateStepAssertLessThanCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-assert-less-than ELEMENT VALUE [POSITION]",
		Short: "Create an assert less than step at a specific position in a checkpoint",
		Long: `Create an assert less than step that verifies an element's value is less than the specified value at the specified position in the checkpoint.

Session Context:
  Uses current checkpoint from session context. Set with 'api-cli set-checkpoint CHECKPOINT_ID'
  Position auto-increments if not specified.

Examples:
  # Using session context (recommended)
  api-cli set-checkpoint 1678318
  api-cli create-step-assert-less-than "Price field" "100"          # Auto-increment position
  api-cli create-step-assert-less-than "Count display" "50" 2         # Explicit position

  # Override checkpoint for specific step
  api-cli create-step-assert-less-than "Price field" "100" 1 --checkpoint 1678319`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			element := args[0]
			value := args[1]

			// Validate element and value
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			if value == "" {
				return fmt.Errorf("value cannot be empty")
			}

			// Resolve checkpoint and position using session context
			ctx, err := resolveStepContext(args, checkpointFlag, 2)
			if err != nil {
				return err
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Create assert less than step using the enhanced client
			stepID, err := client.CreateAssertLessThanStep(ctx.CheckpointID, element, value, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create assert less than step: %w", err)
			}

			// Save session state if position was auto-incremented
			saveStepContext(ctx)

			// Prepare output
			output := &StepOutput{
				Status:       "success",
				StepType:     "ASSERT_LESS_THAN",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("expect %s to be less than \"%s\"", element, value),
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
