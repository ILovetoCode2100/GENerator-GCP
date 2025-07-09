package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertCheckedCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-assert-checked ELEMENT [POSITION]",
		Short: "Create an assertion step that verifies a checkbox or radio button is checked at a specific position",
		Long: `Create an assertion step that verifies a checkbox or radio button is checked at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-assert-checked "Terms and conditions checkbox" 1
  api-cli create-step-assert-checked "Terms and conditions checkbox"  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli create-step-assert-checked "Terms and conditions checkbox" 1 --checkpoint 1678318`,
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
			client := virtuoso.NewClient(cfg)
			
			// Create assert checked step using the client
			stepID, err := client.CreateAssertCheckedStep(ctx.CheckpointID, element, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create assert checked step: %w", err)
			}
			
			// Save config if position was auto-incremented
			saveStepContext(ctx)
			
			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "ASSERT_CHECKED",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("see %s is checked", element),
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