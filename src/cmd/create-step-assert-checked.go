package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertCheckedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-assert-checked CHECKPOINT_ID ELEMENT POSITION",
		Short: "Create an assertion step that verifies a checkbox or radio button is checked at a specific position",
		Long: `Create an assertion step that verifies a checkbox or radio button is checked at the specified position in the checkpoint.
		
Example:
  api-cli create-step-assert-checked 1678318 "Terms and conditions checkbox" 1
  api-cli create-step-assert-checked 1678318 "#terms-checkbox" 2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			element := args[1]
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
			
			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create assert checked step using the client
			stepID, err := client.CreateAssertCheckedStep(checkpointID, element, position)
			if err != nil {
				return fmt.Errorf("failed to create assert checked step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "ASSERT_CHECKED",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"element":       element,
					"position":      position,
					"parsed_step":   fmt.Sprintf("see %s is checked", element),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: ASSERT_CHECKED\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("element: %s\n", element)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: see %s is checked\n", element)
			case "ai":
				fmt.Printf("Successfully created assert checked step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: ASSERT_CHECKED\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Element: %s\n", element)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: see %s is checked\n", element)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created assert checked step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Element: %s\n", element)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}