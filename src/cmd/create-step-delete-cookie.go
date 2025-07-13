package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepDeleteCookieCmd() *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   "create-step-delete-cookie NAME [POSITION]",
		Short: "Create a delete cookie step at a specific position in a checkpoint",
		Long: `Create a delete cookie step that removes a specific cookie by name at the specified position in the checkpoint.

Modern usage (with session context):
  api-cli set-checkpoint 1678318
  api-cli create-step-delete-cookie "session_id"
  api-cli create-step-delete-cookie "auth_token" 2
  api-cli create-step-delete-cookie "user_pref" --checkpoint 1678319

Legacy usage (backward compatible):
  api-cli create-step-delete-cookie 1678318 "session_id" 1
  api-cli create-step-delete-cookie 1678318 "auth_token" 2 -o json`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine if using legacy or modern syntax
			var name string
			var ctx *StepContext
			var err error

			// Check for legacy syntax (first arg is numeric checkpoint ID)
			if _, parseErr := strconv.Atoi(args[0]); parseErr == nil && len(args) >= 3 {
				// Legacy syntax: CHECKPOINT_ID NAME POSITION
				checkpointID, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid checkpoint ID: %w", err)
				}

				name = args[1]
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
			} else {
				// Modern syntax: NAME [POSITION]
				name = args[0]

				// Determine position index for resolveStepContext
				positionIndex := 1
				ctx, err = resolveStepContext(args, checkpointFlag, positionIndex)
				if err != nil {
					return err
				}
			}

			// Validate name
			if name == "" {
				return fmt.Errorf("cookie name cannot be empty")
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create delete cookie step using the enhanced client
			stepID, err := client.CreateDeleteCookieStep(ctx.CheckpointID, name, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create delete cookie step: %w", err)
			}

			// Save session context if position was auto-incremented
			saveStepContext(ctx)

			// Create step output
			output := &StepOutput{
				Status:       "success",
				StepType:     "DELETE_COOKIE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("delete cookie: %s", name),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"name": name},
			}

			// Output the result
			return outputStepResult(output)
		},
	}

	// Add checkpoint flag
	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}
