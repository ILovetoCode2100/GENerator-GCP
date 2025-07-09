package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepStoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-store CHECKPOINT_ID ELEMENT VARIABLE_NAME POSITION",
		Short: "Create a store step at a specific position in a checkpoint",
		Long: `Create a store step that stores a value from an element into a variable at the specified position in the checkpoint.
		
Example:
  api-cli create-step-store 1678318 "#user-id" "userId" 1
  api-cli create-step-store 1678318 "User name field" "userName" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			element := args[1]
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
			
			// Validate element and variable name
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			if variableName == "" {
				return fmt.Errorf("variable name cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create store step using the enhanced client
			stepID, err := client.CreateStoreStep(checkpointID, element, variableName, position)
			if err != nil {
				return fmt.Errorf("failed to create store step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "STORE",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"element":       element,
					"variable_name": variableName,
					"position":      position,
					"parsed_step":   fmt.Sprintf("store value from %s as $%s", element, variableName),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: STORE\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("element: %s\n", element)
				fmt.Printf("variable_name: %s\n", variableName)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: store value from %s as $%s\n", element, variableName)
			case "ai":
				fmt.Printf("Successfully created store step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: STORE\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Element: %s\n", element)
				fmt.Printf("- Variable Name: %s\n", variableName)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: store value from %s as $%s\n", element, variableName)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Use the variable in another step: $%s\n", variableName)
				fmt.Printf("3. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created store step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Element: %s\n", element)
				fmt.Printf("   Variable: $%s\n", variableName)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}