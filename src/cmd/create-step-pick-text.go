package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepPickTextCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-pick-text ELEMENT TEXT [POSITION]",
		Short: "Create a pick text step at a specific position in a checkpoint",
		Long: `Create a pick text step that selects a dropdown option by visible text in a specific element at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-pick-text "Country dropdown" "United States" 1
  api-cli create-step-pick-text "#subscription-select" "Premium Plan"  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-pick-text "Country dropdown" "United States" 1 --checkpoint 1678318

  # Legacy syntax (still supported)
  api-cli create-step-pick-text 1678318 "United States" "Country dropdown" 1`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var element, text string
			var ctx *StepContext
			var err error

			// Detect legacy syntax (first arg is numeric checkpoint ID)
			if len(args) == 4 {
				// Try to parse first arg as checkpoint ID
				if checkpointID, parseErr := strconv.Atoi(args[0]); parseErr == nil {
					// Legacy format: CHECKPOINT_ID TEXT ELEMENT POSITION
					text = args[1]
					element = args[2]
					position, posErr := strconv.Atoi(args[3])
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
					// Not legacy format, treat as modern format with all args
					element = args[0]
					text = args[1]
					ctx, err = resolveStepContext(args[2:], checkpointFlag, 0)
					if err != nil {
						return err
					}
				}
			} else {
				// Modern format: ELEMENT TEXT [POSITION]
				element = args[0]
				text = args[1]
				ctx, err = resolveStepContext(args[2:], checkpointFlag, 0)
				if err != nil {
					return err
				}
			}

			// Validate text and element
			if text == "" {
				return fmt.Errorf("text cannot be empty")
			}
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create pick text step using the enhanced client
			stepID, err := client.CreatePickTextStep(ctx.CheckpointID, text, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create pick text step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "PICK_TEXT",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Pick text \"%s\" in %s", text, element),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"text":    text,
					"element": element,
				},
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
