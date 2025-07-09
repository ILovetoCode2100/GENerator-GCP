package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepPickTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-pick-text CHECKPOINT_ID TEXT ELEMENT POSITION",
		Short: "Create a pick text step at a specific position in a checkpoint",
		Long: `Create a pick text step that selects a dropdown option by visible text in a specific element at the specified position in the checkpoint.
		
Example:
  api-cli create-step-pick-text 1678318 "United States" "Country dropdown" 1
  api-cli create-step-pick-text 1678318 "Premium Plan" "#subscription-select" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			text := args[1]
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
			
			// Validate text and element
			if text == "" {
				return fmt.Errorf("text cannot be empty")
			}
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create pick text step using the enhanced client
			stepID, err := client.CreatePickTextStep(checkpointID, text, element, position)
			if err != nil {
				return fmt.Errorf("failed to create pick text step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "PICK_TEXT",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"text":          text,
					"element":       element,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Pick text \"%s\" in %s", text, element),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: PICK_TEXT\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("text: %s\n", text)
				fmt.Printf("element: %s\n", element)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Pick text \"%s\" in %s\n", text, element)
			case "ai":
				fmt.Printf("Successfully created pick text step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: PICK_TEXT\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Text: %s\n", text)
				fmt.Printf("- Element: %s\n", element)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Pick text \"%s\" in %s\n", text, element)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created pick text step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Text: %s\n", text)
				fmt.Printf("   Element: %s\n", element)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}