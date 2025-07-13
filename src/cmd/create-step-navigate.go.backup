package main

import (
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepNavigateCmd() *cobra.Command {
	var checkpointFlag int
	
	cmd := &cobra.Command{
		Use:   "create-step-navigate URL [POSITION]",
		Short: "Create a navigation step at a specific position in a checkpoint",
		Long: `Create a navigation step that goes to a specific URL at the specified position in the checkpoint.

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
  # Using current checkpoint context
  api-cli create-step-navigate "https://example.com" 1
  api-cli create-step-navigate "https://example.com"  # Auto-increment position
  
  # Override checkpoint explicitly
  api-cli create-step-navigate "https://example.com" 1 --checkpoint 1678318`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			
			// Validate URL
			if url == "" {
				return fmt.Errorf("URL cannot be empty")
			}
			
			// Resolve checkpoint and position
			ctx, err := resolveStepContext(args, checkpointFlag, 1)
			if err != nil {
				return err
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create navigation step using the enhanced client
			stepID, err := client.CreateNavigationStep(ctx.CheckpointID, url, ctx.Position)
			if err != nil {
				return fmt.Errorf("failed to create navigation step: %w", err)
			}
			
			// Save config if position was auto-incremented
			saveStepContext(ctx)
			
			// Output result
			output := &StepOutput{
				Status:       "success",
				StepType:     "NAVIGATE",
				CheckpointID: ctx.CheckpointID,
				StepID:       stepID,
				Position:     ctx.Position,
				ParsedStep:   fmt.Sprintf("Navigate to \"%s\"", url),
				UsingContext: ctx.UsingContext,
				AutoPosition: ctx.AutoPosition,
				Extra:        map[string]interface{}{"url": url},
			}
			
			return outputStepResult(output)
		},
	}
	
	addCheckpointFlag(cmd, &checkpointFlag)
	
	return cmd
}