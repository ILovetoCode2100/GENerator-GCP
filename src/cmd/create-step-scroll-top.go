package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepScrollTopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-scroll-top CHECKPOINT_ID POSITION",
		Short: "Create a scroll to top step at a specific position in a checkpoint",
		Long: `Create a scroll to top step that scrolls to the top of the page at the specified position in the checkpoint.
		
Example:
  api-cli create-step-scroll-top 1678318 1
  api-cli create-step-scroll-top 1678318 2 -o json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			positionStr := args[1]
			
			// Convert IDs to int
			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}
			
			position, err := strconv.Atoi(positionStr)
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create scroll to top step using the enhanced client
			stepID, err := client.CreateScrollTopStep(checkpointID, position)
			if err != nil {
				return fmt.Errorf("failed to create scroll to top step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "SCROLL_TOP",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"position":      position,
					"parsed_step":   "scroll to top of page",
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: SCROLL_TOP\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: scroll to top of page\n")
			case "ai":
				fmt.Printf("Successfully created scroll to top step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: SCROLL_TOP\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: scroll to top of page\n")
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created scroll to top step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}