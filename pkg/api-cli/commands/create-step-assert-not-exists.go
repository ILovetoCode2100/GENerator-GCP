package commands

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newCreateStepAssertNotExistsCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-assert-not-exists ELEMENT [POSITION]",
		Short: "Create an assertion step that verifies an element does not exist at a specific position",
		Long: `Create an assertion step that verifies an element does not exist at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-assert-not-exists "Error message" 1
  api-cli create-step-assert-not-exists "Error message"  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-assert-not-exists "Error message" 1 --checkpoint 1678318`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Create assert not exists step using the client
			stepID, err := client.CreateAssertNotExistsStep(ctx.CheckpointID, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create assert not exists step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "ASSERT_NOT_EXISTS",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("do not see \"%s\"", element),
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
