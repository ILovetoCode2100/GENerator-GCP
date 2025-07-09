package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepPickValueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-pick-value CHECKPOINT_ID VALUE ELEMENT POSITION",
		Short: "Create a pick value step at a specific position in a checkpoint",
		Long: `Create a pick value step that selects a dropdown option by value in a specific element at the specified position in the checkpoint.
		
Example:
  api-cli create-step-pick-value 1678318 "option1" "Country dropdown" 1
  api-cli create-step-pick-value 1678318 "us" "#country-select" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			value := args[1]
			element := args[2]
			positionStr := args[3]
			
			// Convert IDs to int
			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}
			
			position, err := strconv.Atoi(positionStr)
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			
			// Validate value and element
			if value == "" {
				return fmt.Errorf("value cannot be empty")
			}
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create pick value step using the enhanced client
			stepID, err := client.CreatePickValueStep(checkpointID, value, element, position)
			if err != nil {
				return fmt.Errorf("failed to create pick value step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "PICK_VALUE",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"value":         value,
					"element":       element,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Pick value \"%s\" in %s", value, element),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: PICK_VALUE\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("value: %s\n", value)
				fmt.Printf("element: %s\n", element)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Pick value \"%s\" in %s\n", value, element)
			case "ai":
				fmt.Printf("Successfully created pick value step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: PICK_VALUE\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Value: %s\n", value)
				fmt.Printf("- Element: %s\n", element)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Pick value \"%s\" in %s\n", value, element)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created pick value step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Value: %s\n", value)
				fmt.Printf("   Element: %s\n", element)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}