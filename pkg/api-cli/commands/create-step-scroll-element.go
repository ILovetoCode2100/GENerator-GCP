package commands

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newCreateStepScrollElementCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-scroll-element ELEMENT [POSITION]",
		Short: "Create a scroll to element step at a specific position in a checkpoint",
		Long: `Create a scroll to element step that scrolls to a specific element on the page at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-scroll-element "Contact form" 1
  api-cli create-step-scroll-element "#footer"  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-scroll-element "Contact form" 1 --checkpoint 1678318

  # Legacy format (deprecated but still supported)
  api-cli create-step-scroll-element 1678318 "Contact form" 1`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle legacy format (CHECKPOINT_ID ELEMENT POSITION)
			if len(args) == 3 {
				// Legacy format: first arg is checkpoint ID
				checkpointID, err := parseIntArg(args[0], "checkpoint ID")
				if err != nil {
					return err
				}
				element := args[1]
				position, err := parseIntArg(args[2], "position")
				if err != nil {
					return err
				}

				// Validate element
				if element == "" {
					return fmt.Errorf("element cannot be empty")
				}

				// Create Virtuoso client
				client := client.NewClient(cfg)

				// Create scroll to element step using the enhanced client
				stepID, err := client.CreateScrollElementStep(checkpointID, element, position)
				if err != nil {
					return fmt.Errorf("failed to create scroll to element step: %w", err)
				}

				// Output result using legacy context flags
				output := &StepOutput{
					Status:       "success",
					StepType:     "SCROLL_ELEMENT",
					CheckpointID: checkpointID,
					StepID:       stepID,
					Position:     position,
					ParsedStep:   fmt.Sprintf("scroll to %s", element),
					UsingContext: false,
					AutoPosition: false,
					Extra:        map[string]interface{}{"element": element},
				}

				return outputStepResult(output)
			}

			// Modern format: ELEMENT [POSITION]
			element := args[0]

			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}

			// Resolve checkpoint and position
			ctx, err := resolveStepContext(args, checkpointFlag, 1)
			if err != nil {
				return err
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Create scroll to element step using the enhanced client
			stepID, err := client.CreateScrollElementStep(ctx.CheckpointID, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create scroll to element step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "SCROLL_ELEMENT",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("scroll to %s", element),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"element": element},
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
