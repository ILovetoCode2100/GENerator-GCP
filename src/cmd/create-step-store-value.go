package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepStoreValueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-store-value CHECKPOINT_ID VALUE VARIABLE_NAME POSITION",
		Short: "Create a store value step at a specific position in a checkpoint",
		Long: `Create a store value step that stores a specific value in a variable at the specified position in the checkpoint.
		
Example:
  api-cli create-step-store-value 1678318 "test@example.com" "email" 1
  api-cli create-step-store-value 1678318 "12345" "user_id" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			value := args[1]
			variableName := args[2]
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
			
			// Validate value and variable name
			if value == "" {
				return fmt.Errorf("value cannot be empty")
			}
			if variableName == "" {
				return fmt.Errorf("variable name cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create store value step using the enhanced client
			stepID, err := client.CreateStoreValueStep(checkpointID, value, variableName, position)
			if err != nil {
				return fmt.Errorf("failed to create store value step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "STORE_VALUE",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"value":         value,
					"variable_name": variableName,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Store value \"%s\" in variable \"%s\"", value, variableName),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: STORE_VALUE\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("value: %s\n", value)
				fmt.Printf("variable_name: %s\n", variableName)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Store value \"%s\" in variable \"%s\"\n", value, variableName)
			case "ai":
				fmt.Printf("Successfully created store value step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: STORE_VALUE\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Value: %s\n", value)
				fmt.Printf("- Variable Name: %s\n", variableName)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Store value \"%s\" in variable \"%s\"\n", value, variableName)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created store value step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Value: %s\n", value)
				fmt.Printf("   Variable Name: %s\n", variableName)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}