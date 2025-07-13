package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepKeyCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-key KEY [POSITION]",
		Short: "Create a keyboard press step at a specific position in a checkpoint",
		Long: `Create a keyboard press step that presses a specific key at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-key "Enter" 1
  api-cli create-step-key "Tab"  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli create-step-key "Enter" 1 --checkpoint 1678318
  
  # Legacy format (still supported)
  api-cli create-step-key 1678318 "Enter" 1`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var key string
			var ctx *StepContext
			var err error
			
			// Check if using legacy format (first arg is numeric)
			if len(args) >= 3 {
				if _, err := strconv.Atoi(args[0]); err == nil {
					// Legacy format: CHECKPOINT_ID KEY POSITION
					checkpointID, _ := strconv.Atoi(args[0])
					key = args[1]
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
					// Modern format with 3 args is invalid
					return fmt.Errorf("invalid arguments: when providing 3 arguments, first must be checkpoint ID (number)")
				}
			} else {
				// Modern format: KEY [POSITION]
				key = args[0]
				
				// Resolve checkpoint and position
				ctx, err = resolveStepContext(args, checkpointFlag, 1)
				if err != nil {
					return err
				}
			}
			
			// Validate key
			if key == "" {
				return fmt.Errorf("key cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create key step using the enhanced client
			stepID, err := client.CreateKeyStep(ctx.CheckpointID, key, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create key step: %w", err)
			}
			
			// Save config if position was auto-incremented
			saveStepContext(ctx)
			
			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "KEY",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("press %s", key),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"key": key},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}