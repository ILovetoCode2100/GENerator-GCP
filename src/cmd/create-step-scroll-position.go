package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepScrollPositionCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-scroll-position X Y [POSITION]",
		Short: "Create a scroll position step at a specific position in a checkpoint",
		Long: `Create a scroll position step that scrolls to specific X and Y coordinates at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

X and Y coordinates can be negative. Use -- before negative values to avoid flag parsing issues.

Examples:
  # Using current checkpoint context
  api-cli create-step-scroll-position 100 200 1
  api-cli create-step-scroll-position 0 500  # Auto-increment position
  api-cli create-step-scroll-position -- -10 -20  # Negative coordinates with auto-position
  api-cli create-step-scroll-position -- -10 -20 3  # Negative coordinates with position
  
  # Override checkpoint explicitly
  api-cli create-step-scroll-position 100 200 1 --checkpoint 1678318
  
  # Legacy format (deprecated but still supported)
  api-cli create-step-scroll-position 1678318 100 200 1`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle legacy format (CHECKPOINT_ID X Y POSITION)
			if len(args) == 4 {
				// Legacy format: first arg is checkpoint ID
				checkpointID, err := parseIntArg(args[0], "checkpoint ID")
				if err != nil {
					return err
				}
				x, err := parseIntArg(args[1], "X coordinate")
				if err != nil {
					return err
				}
				y, err := parseIntArg(args[2], "Y coordinate")
				if err != nil {
					return err
				}
				position, err := parseIntArg(args[3], "position")
				if err != nil {
					return err
				}
				
				// Create Virtuoso client
				client := virtuoso.NewClient(cfg)
				
				// Create scroll position step using the enhanced client
				stepID, err := client.CreateScrollPositionStep(checkpointID, x, y, position)
				if err != nil {
					return fmt.Errorf("failed to create scroll position step: %w", err)
				}
				
				// Output result using legacy context flags
				output := &StepOutput{
					Status:       "success",
					StepType:     "SCROLL_POSITION",
					CheckpointID: checkpointID,
					StepID:       stepID,
					Position:     position,
					ParsedStep:   fmt.Sprintf("scroll to position (%d, %d)", x, y),
					UsingContext: false,
					AutoPosition: false,
					Extra:        map[string]interface{}{"x": x, "y": y},
				}
				
				return outputStepResult(output)
			}
			
			// Modern format: X Y [POSITION]
			x, err := parseIntArg(args[0], "X coordinate")
			if err != nil {
				return err
			}
			
			y, err := parseIntArg(args[1], "Y coordinate")
			if err != nil {
				return err
			}
			
			// Resolve checkpoint and position
			ctx, err := resolveStepContext(args, checkpointFlag, 2)
			if err != nil {
				return err
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create scroll position step using the enhanced client
			stepID, err := client.CreateScrollPositionStep(ctx.CheckpointID, x, y, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create scroll position step: %w", err)
			}
			
			// Save config if position was auto-incremented
			saveStepContext(ctx)
			
			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "SCROLL_POSITION",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("scroll to position (%d, %d)", x, y),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"x": x, "y": y},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	
	// Enable negative number support
	enableNegativeNumbers(cmd)
	
	return cmd
}