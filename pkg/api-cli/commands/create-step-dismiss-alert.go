package commands

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newCreateStepDismissAlertCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-dismiss-alert [POSITION]",
		Short: "Create a dismiss alert step at a specific position in a checkpoint",
		Long: `Create a dismiss alert step that dismisses JavaScript alerts, confirms, or prompts at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-dismiss-alert 1
  api-cli create-step-dismiss-alert  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-dismiss-alert 1 --checkpoint 1678318

  # Legacy format (still supported)
  api-cli create-step-dismiss-alert 1678318 1`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx *StepContext
			var err error

			// Check for legacy format (2 args where first is checkpoint ID)
			if len(args) == 2 {
				// Try to parse first arg as checkpoint ID
				if checkpointID, parseErr := strconv.Atoi(args[0]); parseErr == nil {
					// Legacy format detected
					position, posErr := strconv.Atoi(args[1])
					if posErr != nil {
						return fmt.Errorf("invalid position: %w", posErr)
					}
					ctx = &StepContext{
						CheckpointID: checkpointID,
						Position:     position,
						UsingContext: false,
						AutoPosition: false,
					}
				} else {
					// Not legacy format, treat as modern format error
					return fmt.Errorf("invalid arguments: expected [POSITION] or CHECKPOINT_ID POSITION")
				}
			} else {
				// Modern format - resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 0)
				if err != nil {
					return err
				}
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Create dismiss alert step using the enhanced client
			stepID, err := client.CreateDismissAlertStep(ctx.CheckpointID, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create dismiss alert step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "DISMISS_ALERT",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   "dismiss alert",
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
