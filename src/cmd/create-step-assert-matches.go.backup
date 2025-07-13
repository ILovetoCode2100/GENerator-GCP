package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertMatchesCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-assert-matches ELEMENT REGEX_PATTERN [POSITION]",
		Short: "Create an assertion step that verifies an element matches a regex pattern at a specific position",
		Long: `Create an assertion step that verifies an element matches a regex pattern at the specified position in the checkpoint.

Session Context:
  Uses current checkpoint from session context. Set with 'api-cli set-checkpoint CHECKPOINT_ID'
  Position auto-increments if not specified.

Examples:
  # Using session context (recommended)
  api-cli set-checkpoint 1678318
  api-cli create-step-assert-matches "Email" "/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/"          # Auto-increment position
  api-cli create-step-assert-matches "#email-field" "/.*@example\.com/" 2         # Explicit position
  
  # Override checkpoint for specific step
  api-cli create-step-assert-matches "Email" "/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/" 1 --checkpoint 1678319`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			element := args[0]
			regexPattern := args[1]
			
			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Validate regex pattern
			if regexPattern == "" {
				return fmt.Errorf("regex pattern cannot be empty")
			}
			
			// Resolve checkpoint and position using session context
			ctx, err := resolveStepContext(args, checkpointFlag, 2)
			if err != nil {
				return err
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create assert matches step using the client
			stepID, err := client.CreateAssertMatchesStep(ctx.CheckpointID, element, regexPattern, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create assert matches step: %w", err)
			}
			
			// Save session state if position was auto-incremented
			saveStepContext(ctx)
			
			// Prepare output
			output := &StepOutput{
				Status:       "success",
				StepType:     "ASSERT_MATCHES",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("expect %s to match pattern \"%s\"", element, regexPattern),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"element": element, "regex_pattern": regexPattern},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	return cmd
}