package main

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepCommentCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-comment COMMENT [POSITION]",
		Short: "Create a comment step at a specific position in a checkpoint",
		Long: `Create a comment step that adds documentation or notes at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context (modern syntax)
  api-cli create-step-comment "This step logs in the user" 1
  api-cli create-step-comment "Validate the dashboard loads correctly"  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli create-step-comment "This is a comment" 1 --checkpoint 1678318
  
  # Legacy syntax (backward compatibility)
  api-cli create-step-comment 1678318 "This step logs in the user" 1`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var comment string
			var ctx *StepContext
			var err error
			
			// Detect legacy syntax: 3 args where first is numeric
			if len(args) == 3 {
				if _, err := strconv.Atoi(args[0]); err == nil {
					// Legacy syntax: CHECKPOINT_ID COMMENT POSITION
					checkpointID, _ := strconv.Atoi(args[0])
					comment = args[1]
					position, err := strconv.Atoi(args[2])
					if err != nil {
						return fmt.Errorf("invalid position: %w", err)
					}
					
					// Create context manually for legacy syntax
					ctx = &StepContext{
						CheckpointID: checkpointID,
						Position:     position,
						UsingContext: false,
						AutoPosition: false,
					}
				} else {
					// Modern syntax with 3 args (shouldn't happen with valid input)
					return fmt.Errorf("invalid arguments: when providing 3 arguments, first must be checkpoint ID (number)")
				}
			} else {
				// Modern syntax: COMMENT [POSITION]
				comment = args[0]
				
				// Resolve checkpoint and position using helper
				ctx, err = resolveStepContext(args, checkpointFlag, 1)
				if err != nil {
					return err
				}
			}
			
			// Validate comment
			if comment == "" {
				return fmt.Errorf("comment cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create comment step using the enhanced client
			stepID, err := client.CreateCommentStep(ctx.CheckpointID, comment, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create comment step: %w", err)
			}
			
			// Save context if position was auto-incremented
			saveStepContext(ctx)
			
			// Format output using the helper
			output := &StepOutput{
				Status:       "success",
				StepType:     "COMMENT",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("# %s", comment),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra: map[string]interface{}{
					"comment": comment,
				},
			}
			
			return outputStepResult(output)
		},
	}
	
	// Add checkpoint flag
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}