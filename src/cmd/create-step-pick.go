package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepPickCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-pick ELEMENT INDEX [POSITION]",
		Short: "Create a dropdown selection step at a specific position in a checkpoint",
		Long: `Create a dropdown selection step that picks a specific option by index from a dropdown element at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-pick "country dropdown" 2 1
  api-cli create-step-pick "#size-select" 0  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-pick "country dropdown" 2 1 --checkpoint 1678318

  # Legacy format (still supported)
  api-cli create-step-pick 1678318 "Option 1" "country dropdown" 1`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var element string
			var index string
			var ctx *StepContext
			var err error

			// Check if using legacy format (first arg is numeric)
			if len(args) == 4 {
				if _, err := strconv.Atoi(args[0]); err == nil {
					// Legacy format: CHECKPOINT_ID VALUE ELEMENT POSITION
					// Note: In legacy format, "VALUE" was actually the text to pick, not an index
					checkpointID, _ := strconv.Atoi(args[0])
					index = args[1] // In legacy, this was the value/text
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
				// Modern format: ELEMENT INDEX [POSITION]
				element = args[0]
				index = args[1]

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
			if index == "" {
				return fmt.Errorf("index/value cannot be empty")
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create pick step using the enhanced client
			stepID, err := client.CreatePickStep(ctx.CheckpointID, index, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create pick step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "PICK",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("pick \"%s\" from %s", index, element),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"element": element,
					"value":   index,
				},
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
