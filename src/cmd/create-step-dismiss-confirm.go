package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepDismissConfirmCmd() *cobra.Command {
	var checkpointFlag int
	var acceptFlag bool
	var cancelFlag bool

	cmd := &cobra.Command{
		Use:   "create-step-dismiss-confirm ACCEPT [POSITION]",
		Short: "Create a dismiss confirm dialog step at a specific position in a checkpoint",
		Long: `Create a dismiss confirm dialog step that handles a JavaScript confirm dialog at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

The ACCEPT parameter should be "true" to accept or "false" to cancel the dialog.
Alternatively, use --accept or --cancel flags.

Examples:
  # Using current checkpoint context
  api-cli create-step-dismiss-confirm true 1       # Accept dialog
  api-cli create-step-dismiss-confirm false        # Cancel dialog, auto-increment position
  api-cli create-step-dismiss-confirm --accept     # Accept using flag
  api-cli create-step-dismiss-confirm --cancel 2   # Cancel using flag at position 2

  # Override checkpoint explicitly
  api-cli create-step-dismiss-confirm true 1 --checkpoint 1678318

  # Legacy format (still supported)
  api-cli create-step-dismiss-confirm 1678318 1 --accept
  api-cli create-step-dismiss-confirm 1678318 2 --cancel`,
		Args: cobra.RangeArgs(0, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx *StepContext
			var accept bool
			var err error
			var acceptArgIndex int = -1

			// Determine accept value from flags first
			if acceptFlag {
				accept = true
			} else if cancelFlag {
				accept = false
			} else if len(args) > 0 {
				// Check if first argument is a boolean value
				acceptStr := strings.ToLower(args[0])
				if acceptStr == "true" || acceptStr == "false" {
					accept = (acceptStr == "true")
					acceptArgIndex = 0
				}
			}

			// Check for legacy format (first arg is checkpoint ID)
			if len(args) >= 2 {
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
					// In legacy format, accept is determined by flags only
					if !acceptFlag && !cancelFlag {
						// Default to cancel if no flag specified
						accept = false
					}
				} else if acceptArgIndex == 0 {
					// Modern format with ACCEPT argument
					positionIndex := 1
					ctx, err = resolveStepContext(args, checkpointFlag, positionIndex)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("invalid arguments: expected ACCEPT [POSITION] or CHECKPOINT_ID POSITION")
				}
			} else if acceptArgIndex == 0 {
				// Modern format with just ACCEPT argument
				ctx, err = resolveStepContext(args, checkpointFlag, 1)
				if err != nil {
					return err
				}
			} else if acceptFlag || cancelFlag {
				// Modern format with flags only
				ctx, err = resolveStepContext(args, checkpointFlag, 0)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("must specify accept/cancel action: use 'true'/'false' argument or --accept/--cancel flag")
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create dismiss confirm step using the enhanced client
			stepID, err := client.CreateDismissConfirmStep(ctx.CheckpointID, accept, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create dismiss confirm step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			action := "cancel"
			if accept {
				action = "accept"
			}

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "DISMISS_CONFIRM",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("dismiss confirm dialog (%s)", action),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"action": action},
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)
	cmd.Flags().BoolVar(&acceptFlag, "accept", false, "Accept the confirm dialog")
	cmd.Flags().BoolVar(&cancelFlag, "cancel", false, "Cancel the confirm dialog")

	return cmd
}
