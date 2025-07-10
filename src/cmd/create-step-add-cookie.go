package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAddCookieCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-add-cookie NAME VALUE DOMAIN PATH [POSITION]",
		Short: "Create an add cookie step at a specific position in a checkpoint",
		Long: `Create an add cookie step that adds a cookie with the specified name and value at the specified position in the checkpoint.

Modern usage (with session context):
  api-cli set-checkpoint 1678318
  api-cli create-step-add-cookie "session_id" "abc123" "example.com" "/"
  api-cli create-step-add-cookie "user_preference" "dark_mode" "example.com" "/" 2
  api-cli create-step-add-cookie "theme" "light" "example.com" "/" --checkpoint 1678319

Legacy usage (backward compatible):
  api-cli create-step-add-cookie 1678318 "session_id" "abc123" 1
  api-cli create-step-add-cookie 1678318 "user_preference" "dark_mode" 2 -o json`,
		Args: cobra.RangeArgs(4, 6),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine if using legacy or modern syntax
			var name, value, domain, path string
			var ctx *StepContext
			var err error
			
			// Check for legacy syntax (first arg is numeric checkpoint ID)
			if _, parseErr := strconv.Atoi(args[0]); parseErr == nil && len(args) >= 4 {
				// Legacy syntax: CHECKPOINT_ID NAME VALUE POSITION (or CHECKPOINT_ID NAME VALUE DOMAIN PATH POSITION)
				checkpointID, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid checkpoint ID: %w", err)
				}
				
				if len(args) == 4 {
					// Legacy: CHECKPOINT_ID NAME VALUE POSITION
					name = args[1]
					value = args[2]
					domain = ""
					path = ""
					position, err := strconv.Atoi(args[3])
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
					// Legacy: CHECKPOINT_ID NAME VALUE DOMAIN PATH POSITION
					name = args[1]
					value = args[2]
					domain = args[3]
					path = args[4]
					position, err := strconv.Atoi(args[5])
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
			} else {
				// Modern syntax: NAME VALUE DOMAIN PATH [POSITION]
				name = args[0]
				value = args[1]
				domain = args[2]
				path = args[3]
				
				// Determine position index for resolveStepContext
				positionIndex := 4
				ctx, err = resolveStepContext(args, checkpointFlag, positionIndex)
				if err != nil {
					return err
				}
			}
			
			// Validate name and value
			if name == "" {
				return fmt.Errorf("cookie name cannot be empty")
			}
			if value == "" {
				return fmt.Errorf("cookie value cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create add cookie step using the enhanced client
			stepID, err := client.CreateAddCookieStep(ctx.CheckpointID, name, value, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create add cookie step: %w", err)
			}
			
			// Save session context if position was auto-incremented
			saveStepContext(ctx)
			
			// Build parsed step description
			parsedStep := fmt.Sprintf("add cookie \"%s\" with value \"%s\"", name, value)
			if domain != "" {
				parsedStep += fmt.Sprintf(" for domain \"%s\"", domain)
			}
			if path != "" {
				parsedStep += fmt.Sprintf(" with path \"%s\"", path)
			}
			
			// Prepare extra data for output
			extraData := map[string]interface{}{
				"name":  name,
				"value": value,
			}
			if domain != "" {
				extraData["domain"] = domain
			}
			if path != "" {
				extraData["path"] = path
			}
			
			// Create step output
			output := &StepOutput{
				Status:       "success",
				StepType:     "ADD_COOKIE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   parsedStep,
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        extraData,
			}
			
			// Output the result
			return outputStepResult(output)
		},
	}
	
	// Add checkpoint flag
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}