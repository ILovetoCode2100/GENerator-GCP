package commands

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newCreateStepWaitTimeCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-wait-time SECONDS [POSITION]",
		Short: "Create a wait time step at a specific position in a checkpoint",
		Long: `Create a wait time step that waits for a specified number of seconds at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-wait-time 5 1
  api-cli create-step-wait-time 10  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-wait-time 5 1 --checkpoint 1678318

  # Legacy syntax (deprecated but still supported)
  api-cli create-step-wait-time 1678318 5 1`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Support both modern and legacy syntax
			if len(args) == 3 {
				// Legacy: CHECKPOINT_ID SECONDS POSITION
				// Check if first arg is a checkpoint ID (all digits)
				if _, err := strconv.Atoi(args[0]); err == nil {
					return nil // Legacy syntax
				}
			}
			// Modern: SECONDS [POSITION]
			if len(args) >= 1 && len(args) <= 2 {
				return nil
			}
			return fmt.Errorf("accepts 1-2 args (modern) or 3 args (legacy), received %d", len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var seconds int
			var err error
			var ctx *StepContext

			// Handle legacy syntax
			if len(args) == 3 {
				if checkpointID, err := strconv.Atoi(args[0]); err == nil {
					// Legacy syntax detected
					seconds, err = strconv.Atoi(args[1])
					if err != nil {
						return fmt.Errorf("invalid seconds: %w", err)
					}

					position, err := strconv.Atoi(args[2])
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
				seconds, err = strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid seconds: %w", err)
				}

				// Resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 1)
				if err != nil {
					return err
				}
			}

			// Validate seconds
			if seconds <= 0 {
				return fmt.Errorf("seconds must be greater than 0")
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Create wait time step using the enhanced client
			stepID, err := client.CreateWaitTimeStep(ctx.CheckpointID, seconds, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create wait time step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "WAIT_TIME",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Wait %d seconds", seconds),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"seconds": seconds},
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
