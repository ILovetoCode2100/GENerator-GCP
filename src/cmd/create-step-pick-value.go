package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepPickValueCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-pick-value ELEMENT VALUE [POSITION]",
		Short: "Create a pick value step at a specific position in a checkpoint",
		Long: `Create a pick value step that selects a dropdown option by value in a specific element at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-pick-value "Country dropdown" "option1" 1
  api-cli create-step-pick-value "#country-select" "us"  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-pick-value "Country dropdown" "option1" 1 --checkpoint 1678318

  # Legacy format (still supported)
  api-cli create-step-pick-value 1678318 "option1" "Country dropdown" 1`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var element string
			var value string
			var ctx *StepContext
			var err error

			// Check if using legacy format (first arg is numeric)
			if len(args) == 4 {
				if _, err := strconv.Atoi(args[0]); err == nil {
					// Legacy format: CHECKPOINT_ID VALUE ELEMENT POSITION
					checkpointID, _ := strconv.Atoi(args[0])
					value = args[1]
					element = args[2]
					position, err := strconv.Atoi(args[3])
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
					// Modern format with 4 args is invalid
					return fmt.Errorf("invalid arguments: when providing 4 arguments, first must be checkpoint ID (number)")
				}
			} else {
				// Modern format: ELEMENT VALUE [POSITION]
				element = args[0]
				value = args[1]

				// Resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 2)
				if err != nil {
					return err
				}
			}

			// Validate inputs
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			if value == "" {
				return fmt.Errorf("value cannot be empty")
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create pick value step using the enhanced client
			stepID, err := client.CreatePickValueStep(ctx.CheckpointID, value, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create pick value step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "PICK_VALUE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Pick value \"%s\" in %s", value, element),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"element": element,
					"value":   value,
				},
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
