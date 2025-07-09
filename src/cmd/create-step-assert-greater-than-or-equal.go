package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertGreaterThanOrEqualCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-assert-greater-than-or-equal CHECKPOINT_ID ELEMENT VALUE POSITION",
		Short: "Create an assertion step that verifies an element is greater than or equal to a specific value at a specific position",
		Long: `Create an assertion step that verifies an element is greater than or equal to a specific value at the specified position in the checkpoint.
		
Example:
  api-cli create-step-assert-greater-than-or-equal 1678318 "Score" "75" 1
  api-cli create-step-assert-greater-than-or-equal 1678318 "#score-value" "80" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			element := args[1]
			value := args[2]
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
			
			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Validate value
			if value == "" {
				return fmt.Errorf("value cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create assert greater than or equal step using the client
			stepID, err := client.CreateAssertGreaterThanOrEqualStep(checkpointID, element, value, position)
			if err != nil {
				return fmt.Errorf("failed to create assert greater than or equal step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "ASSERT_GREATER_THAN_OR_EQUAL",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"element":       element,
					"value":         value,
					"position":      position,
					"parsed_step":   fmt.Sprintf("expect %s to be greater than or equal to \"%s\"", element, value),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: ASSERT_GREATER_THAN_OR_EQUAL\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("element: %s\n", element)
				fmt.Printf("value: %s\n", value)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: expect %s to be greater than or equal to \"%s\"\n", element, value)
			case "ai":
				fmt.Printf("Successfully created assert greater than or equal step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: ASSERT_GREATER_THAN_OR_EQUAL\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Element: %s\n", element)
				fmt.Printf("- Value: %s\n", value)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: expect %s to be greater than or equal to \"%s\"\n", element, value)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created assert greater than or equal step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Element: %s\n", element)
				fmt.Printf("   Value: %s\n", value)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}