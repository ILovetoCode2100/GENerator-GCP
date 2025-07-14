package commands

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newCreateStepHoverCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-hover ELEMENT [POSITION]",
		Short: "Create a hover step at a specific position in a checkpoint",
		Long: `Create a hover step that hovers over a specific element at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-hover "Menu item" 1
  api-cli create-step-hover ".dropdown-trigger"  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-hover "Menu item" 1 --checkpoint 1678318

  # Legacy syntax (deprecated but still supported)
  api-cli create-step-hover 1678318 "Menu item" 1`,
		Args: func(cmd *cobra.Command, args []string) error {
			// Support both modern and legacy syntax
			if len(args) == 3 {
				// Legacy: CHECKPOINT_ID ELEMENT POSITION
				// Check if first arg is a checkpoint ID (all digits)
				if _, err := strconv.Atoi(args[0]); err == nil {
					return nil // Legacy syntax
				}
			}
			// Modern: ELEMENT [POSITION]
			if len(args) >= 1 && len(args) <= 2 {
				return nil
			}
			return fmt.Errorf("accepts 1-2 args (modern) or 3 args (legacy), received %d", len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var element string
			var ctx *StepContext
			var err error

			// Handle legacy syntax
			if len(args) == 3 {
				if checkpointID, err := strconv.Atoi(args[0]); err == nil {
					// Legacy syntax detected
					element = args[1]

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
				element = args[0]

				// Resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 1)
				if err != nil {
					return err
				}
			}

			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Create hover step using the enhanced client
			stepID, err := client.CreateHoverStep(ctx.CheckpointID, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create hover step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "HOVER",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("hover on %s", element),
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
