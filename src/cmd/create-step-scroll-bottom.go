package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepScrollBottomCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-scroll-bottom [POSITION]",
		Short: "Create a scroll to bottom step at a specific position in a checkpoint",
		Long: `Create a scroll to bottom step that scrolls to the bottom of the page at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-scroll-bottom 1
  api-cli create-step-scroll-bottom  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-scroll-bottom 1 --checkpoint 1678318

  # Legacy format (deprecated but still supported)
  api-cli create-step-scroll-bottom 1678318 1`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle legacy format (CHECKPOINT_ID POSITION)
			if len(args) == 2 {
				// Legacy format: first arg is checkpoint ID
				checkpointID, err := parseIntArg(args[0], "checkpoint ID")
				if err != nil {
					return err
				}
				position, err := parseIntArg(args[1], "position")
				if err != nil {
					return err
				}

				// Create Virtuoso client
				client := virtuoso.NewClient(cfg)

				// Create scroll to bottom step using the enhanced client
				stepID, err := client.CreateScrollBottomStep(checkpointID, position)
				if err != nil {
					return fmt.Errorf("failed to create scroll to bottom step: %w", err)
				}

				// Output result using legacy context flags
				output := &StepOutput{
					Status:       "success",
					StepType:     "SCROLL_BOTTOM",
					CheckpointID: checkpointID,
					StepID:       stepID,
					Position:     position,
					ParsedStep:   "scroll to bottom of page",
					UsingContext: false,
					AutoPosition: false,
				}

				return outputStepResult(output)
			}

			// Modern format: use session context
			ctx, err := resolveStepContext(args, checkpointFlag, 0)
			if err != nil {
				return err
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create scroll to bottom step using the enhanced client
			stepID, err := client.CreateScrollBottomStep(ctx.CheckpointID, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create scroll to bottom step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "SCROLL_BOTTOM",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   "scroll to bottom of page",
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
