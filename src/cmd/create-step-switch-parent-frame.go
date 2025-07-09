package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepSwitchParentFrameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-switch-parent-frame CHECKPOINT_ID POSITION",
		Short: "Create a switch to parent frame step at a specific position in a checkpoint",
		Long: `Create a switch to parent frame step that switches back to the parent frame at the specified position in the checkpoint.
		
Example:
  api-cli create-step-switch-parent-frame 1678318 1
  api-cli create-step-switch-parent-frame 1678318 2 -o json`,
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
			
			// Create switch parent frame step using the enhanced client
			stepID, err := client.CreateSwitchParentFrameStep(checkpointID, position)
			if err != nil {
				return fmt.Errorf("failed to create switch parent frame step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "SWITCH",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"position":      position,
					"parsed_step":   "switch to parent frame",
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: SWITCH\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: switch to parent frame\n")
			case "ai":
				fmt.Printf("Successfully created switch parent frame step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: SWITCH (parent frame)\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: switch to parent frame\n")
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Switch to iframe: api-cli create-step-switch-iframe %d <selector> <position>\n", checkpointID)
				fmt.Printf("3. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created switch parent frame step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}