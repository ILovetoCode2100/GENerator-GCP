package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-comment CHECKPOINT_ID COMMENT POSITION",
		Short: "Create a comment step at a specific position in a checkpoint",
		Long: `Create a comment step that adds documentation or notes at the specified position in the checkpoint.
		
Example:
  api-cli create-step-comment 1678318 "This step logs in the user" 1
  api-cli create-step-comment 1678318 "Validate the dashboard loads correctly" 2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			comment := args[1]
			positionStr := args[2]
			
			// Convert IDs to int
			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}
			
			position, err := strconv.Atoi(positionStr)
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			
			// Validate comment
			if comment == "" {
				return fmt.Errorf("comment cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create comment step using the enhanced client
			stepID, err := client.CreateCommentStep(checkpointID, comment, position)
			if err != nil {
				return fmt.Errorf("failed to create comment step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "COMMENT",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"comment":       comment,
					"position":      position,
					"parsed_step":   fmt.Sprintf("# %s", comment),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: COMMENT\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("comment: %s\n", comment)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: # %s\n", comment)
			case "ai":
				fmt.Printf("Successfully created comment step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: COMMENT\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Comment: %s\n", comment)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: # %s\n", comment)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created comment step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Comment: %s\n", comment)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}