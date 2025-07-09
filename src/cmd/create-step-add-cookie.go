package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAddCookieCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-add-cookie CHECKPOINT_ID NAME VALUE POSITION",
		Short: "Create an add cookie step at a specific position in a checkpoint",
		Long: `Create an add cookie step that adds a cookie with the specified name and value at the specified position in the checkpoint.
		
Example:
  api-cli create-step-add-cookie 1678318 "session_id" "abc123" 1
  api-cli create-step-add-cookie 1678318 "user_preference" "dark_mode" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			name := args[1]
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
			
			// Validate name and value
			if name == "" {
				return fmt.Errorf("cookie name cannot be empty")
			}
			if value == "" {
				return fmt.Errorf("cookie value cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create add cookie step using the enhanced client
			stepID, err := client.CreateAddCookieStep(checkpointID, name, value, position)
			if err != nil {
				return fmt.Errorf("failed to create add cookie step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "ADD_COOKIE",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"name":          name,
					"value":         value,
					"position":      position,
					"parsed_step":   fmt.Sprintf("add cookie \"%s\" with value \"%s\"", name, value),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: ADD_COOKIE\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("name: %s\n", name)
				fmt.Printf("value: %s\n", value)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: add cookie \"%s\" with value \"%s\"\n", name, value)
			case "ai":
				fmt.Printf("Successfully created add cookie step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: ADD_COOKIE\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Cookie Name: %s\n", name)
				fmt.Printf("- Cookie Value: %s\n", value)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: add cookie \"%s\" with value \"%s\"\n", name, value)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created add cookie step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Cookie Name: %s\n", name)
				fmt.Printf("   Cookie Value: %s\n", value)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}