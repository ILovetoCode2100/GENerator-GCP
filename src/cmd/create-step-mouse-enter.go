package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepMouseEnterCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-mouse-enter ELEMENT [POSITION]",
		Short: "Create a mouse enter step at a specific position in a checkpoint",
		Long: `Create a mouse enter step that triggers a mouse enter event on a specific element at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-mouse-enter "Hover menu" 1
  api-cli create-step-mouse-enter "#hover-trigger"  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-mouse-enter "Hover menu" 1 --checkpoint 1678318

  # Legacy syntax (still supported)
  api-cli create-step-mouse-enter 1678318 "Hover menu" 1`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var element string
			var ctx *StepContext
			var err error

			// Handle both modern and legacy syntax
			if len(args) == 3 {
				// Legacy syntax: CHECKPOINT_ID ELEMENT POSITION
				checkpointID, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid checkpoint ID: %w", err)
				}
				checkpointFlag = checkpointID
				element = args[1]
				// Shift args to match modern pattern
				args = args[1:]
			} else {
				// Modern syntax: ELEMENT [POSITION]
				element = args[0]
			}

			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}

			// Resolve checkpoint and position
			ctx, err = resolveStepContext(args, checkpointFlag, 1)
			if err != nil {
				return err
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create mouse enter step using the enhanced client
			stepID, err := client.CreateMouseEnterStep(ctx.CheckpointID, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create mouse enter step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "MOUSE_ENTER",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Mouse enter on %s", element),
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
