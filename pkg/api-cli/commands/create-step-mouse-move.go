package commands

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newCreateStepMouseMoveCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-mouse-move X Y [POSITION]",
		Short: "Create a mouse move step to absolute coordinates at a specific position in a checkpoint",
		Long: `Create a mouse move step that moves the mouse to specific X,Y coordinates at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-mouse-move 100 200 1
  api-cli create-step-mouse-move 100 200  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-mouse-move 100 200 1 --checkpoint 1678318

  # Legacy syntax (still supported)
  api-cli create-step-mouse-move 1678318 100 200 1`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var x, y int
			var ctx *StepContext
			var err error

			// Handle both modern and legacy syntax
			if len(args) == 4 {
				// Legacy syntax: CHECKPOINT_ID X Y POSITION
				checkpointID, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid checkpoint ID: %w", err)
				}
				checkpointFlag = checkpointID
				x, err = parseIntArg(args[1], "X coordinate")
				if err != nil {
					return err
				}
				y, err = parseIntArg(args[2], "Y coordinate")
				if err != nil {
					return err
				}
				// Shift args to match modern pattern
				args = args[2:]
			} else {
				// Modern syntax: X Y [POSITION]
				x, err = parseIntArg(args[0], "X coordinate")
				if err != nil {
					return err
				}
				y, err = parseIntArg(args[1], "Y coordinate")
				if err != nil {
					return err
				}
			}

			// Resolve checkpoint and position
			ctx, err = resolveStepContext(args, checkpointFlag, 2)
			if err != nil {
				return err
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Create mouse move step using the enhanced client
			stepID, err := client.CreateMouseMoveToStep(ctx.CheckpointID, x, y, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create mouse move step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "MOUSE_MOVE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Move mouse to (%d, %d)", x, y),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"x": x, "y": y},
			}

			return outputStepResult(output)
		},
	}

	// Enable negative numbers for coordinate values
	enableNegativeNumbers(cmd)
	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
