package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertEqualsCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-assert-equals ELEMENT VALUE [POSITION]",
		Short: "Create an assertion step that verifies an element has a specific text value at a specific position",
		Long: `Create an assertion step that verifies an element has a specific text value at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-assert-equals "Username field" "john@example.com" 1
  api-cli create-step-assert-equals "Username field" "john@example.com"  # Auto-increment position

  # Override checkpoint explicitly
  api-cli create-step-assert-equals "Username field" "john@example.com" 1 --checkpoint 1678318`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			element := args[0]
			value := args[1]

			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}

			// Validate value
			if value == "" {
				return fmt.Errorf("value cannot be empty")
			}

			// Resolve checkpoint and position
			ctx, err := resolveStepContext(args, checkpointFlag, 2)
			if err != nil {
				return err
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create assert equals step using the client
			stepID, err := client.CreateAssertEqualsStep(ctx.CheckpointID, element, value, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create assert equals step: %w", err)
			}

			// Save config if position was auto-incremented
			saveStepContext(ctx)

			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "ASSERT_EQUALS",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("expect %s to have text \"%s\"", element, value),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"element": element, "value": value},
			}

			return outputStepResult(output)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
