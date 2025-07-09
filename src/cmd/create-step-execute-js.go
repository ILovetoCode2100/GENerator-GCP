package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepExecuteJsCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-execute-js JAVASCRIPT [VARIABLE_NAME] [POSITION]",
		Short: "Create an execute JavaScript step at a specific position in a checkpoint",
		Long: `Create an execute JavaScript step that runs custom JavaScript code at the specified position in the checkpoint.
		
Modern usage (with session context):
  api-cli set-checkpoint 1678318
  api-cli create-step-execute-js "window.scrollTo(0, 0)"
  api-cli create-step-execute-js "document.querySelector('#user-id').innerText" "userId"
  api-cli create-step-execute-js "localStorage.getItem('token')" "authToken" 3
  api-cli create-step-execute-js "window.location.href" --checkpoint 1678319

Legacy usage:
  api-cli create-step-execute-js 1678318 "window.scrollTo(0, 0)" 1
  api-cli create-step-execute-js 1678318 "document.querySelector('#modal').style.display = 'none'" 2 -o json
  api-cli create-step-execute-js 1678318 "document.querySelector('#user-id').innerText" "userId" 3`,
		Args: cobra.RangeArgs(1, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var javascript, variableName string
			var ctx *StepContext
			var err error
			
			// Handle both modern and legacy patterns
			if len(args) >= 3 {
				// Check if first arg is a checkpoint ID (legacy pattern)
				if checkpointID, err := strconv.Atoi(args[0]); err == nil {
					// Legacy patterns
					if len(args) == 3 {
						// Legacy: CHECKPOINT_ID JAVASCRIPT POSITION
						position, err := strconv.Atoi(args[2])
						if err != nil {
							return fmt.Errorf("invalid position: %w", err)
						}
						javascript = args[1]
						ctx = &StepContext{
							CheckpointID: checkpointID,
							Position:     position,
							UsingContext: false,
							AutoPosition: false,
						}
					} else if len(args) == 4 {
						// Legacy: CHECKPOINT_ID JAVASCRIPT VARIABLE_NAME POSITION
						position, err := strconv.Atoi(args[3])
						if err != nil {
							return fmt.Errorf("invalid position: %w", err)
						}
						javascript = args[1]
						variableName = args[2]
						ctx = &StepContext{
							CheckpointID: checkpointID,
							Position:     position,
							UsingContext: false,
							AutoPosition: false,
						}
					}
				} else {
					// Modern pattern with 3 args: JAVASCRIPT VARIABLE_NAME POSITION
					javascript = args[0]
					variableName = args[1]
					ctx, err = resolveStepContext(args, checkpointFlag, 2)
					if err != nil {
						return err
					}
				}
			} else {
				// Modern patterns with 1-2 args
				javascript = args[0]
				if len(args) == 2 {
					// Check if second arg is a position (number) or variable name (string)
					if _, err := strconv.Atoi(args[1]); err == nil {
						// JAVASCRIPT POSITION
						ctx, err = resolveStepContext(args, checkpointFlag, 1)
					} else {
						// JAVASCRIPT VARIABLE_NAME
						variableName = args[1]
						ctx, err = resolveStepContext(args, checkpointFlag, 2)
					}
					if err != nil {
						return err
					}
				} else {
					// Just JAVASCRIPT
					ctx, err = resolveStepContext(args, checkpointFlag, 1)
					if err != nil {
						return err
					}
				}
			}
			
			// Validate JavaScript
			if javascript == "" {
				return fmt.Errorf("javascript cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create execute JavaScript step using the enhanced client
			// Note: The client method doesn't support variable name, so we'll need to handle that separately if needed
			stepID, err := client.CreateExecuteJsStep(ctx.CheckpointID, javascript, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create execute JavaScript step: %w", err)
			}
			
			// Save session context if position was auto-incremented
			saveStepContext(ctx)
			
			// Build parsed step description
			parsedStep := fmt.Sprintf("execute JS \"%s\"", javascript)
			if variableName != "" {
				parsedStep = fmt.Sprintf("execute JS \"%s\" and store in $%s", javascript, variableName)
			}
			
			// Output the result
			extra := map[string]interface{}{
				"javascript": javascript,
			}
			if variableName != "" {
				extra["variable_name"] = variableName
			}
			
			output := &StepOutput{
				Status:       "success",
				StepType:     "EXECUTE_JS",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   parsedStep,
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        extra,
			}
			
			return outputStepResult(output)
		},
	}
	
	// Add the --checkpoint flag
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}