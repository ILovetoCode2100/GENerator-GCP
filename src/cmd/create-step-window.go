package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepWindowCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-window WIDTH HEIGHT [POSITION]",
		Short: "Create a window resize step at a specific position in a checkpoint",
		Long: `Create a window resize step that sets the browser window size to the specified dimensions at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-window 1920 1080 1
  api-cli create-step-window 1280 800  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-window 1920 1080 1 --checkpoint 1678318

  # Legacy syntax (deprecated but still supported)
  api-cli create-step-window 1678318 1920 1080 1`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Support both modern and legacy syntax
			if len(args) == 4 {
				// Legacy: CHECKPOINT_ID WIDTH HEIGHT POSITION
				// Check if first arg is a checkpoint ID (all digits)
				if _, err := strconv.Atoi(args[0]); err == nil {
					return nil // Legacy syntax
				}
			}
			// Modern: WIDTH HEIGHT [POSITION]
			if len(args) >= 2 && len(args) <= 3 {
				return nil
			}
			return fmt.Errorf("accepts 2-3 args (modern) or 4 args (legacy), received %d", len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var width, height int
			var err error
			var ctx *StepContext

			// Handle legacy syntax
			if len(args) == 4 {
				if checkpointID, err := strconv.Atoi(args[0]); err == nil {
					// Legacy syntax detected
					width, err = strconv.Atoi(args[1])
					if err != nil {
						return fmt.Errorf("invalid width: %w", err)
					}

					height, err = strconv.Atoi(args[2])
					if err != nil {
						return fmt.Errorf("invalid height: %w", err)
					}

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
				}
			}

			// Modern syntax
			if ctx == nil {
				width, err = strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid width: %w", err)
				}

				height, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid height: %w", err)
				}

				// Resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 2)
				if err != nil {
					return err
				}
			}

			// Validate dimensions
			if width <= 0 || height <= 0 {
				return fmt.Errorf("width and height must be greater than 0")
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create window resize step using the enhanced client
			stepID, err := client.CreateWindowResizeStep(ctx.CheckpointID, width, height, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create window resize step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "WINDOW_RESIZE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Set browser window size to %dx%d", width, height),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"width":  width,
					"height": height,
				},
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
