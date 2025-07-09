package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepDeleteCookieCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-delete-cookie CHECKPOINT_ID NAME POSITION",
		Short: "Create a delete cookie step at a specific position in a checkpoint",
		Long: `Create a delete cookie step that removes a specific cookie by name at the specified position in the checkpoint.
		
Example:
  api-cli create-step-delete-cookie 1678318 "session_id" 1
  api-cli create-step-delete-cookie 1678318 "auth_token" 2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			name := args[1]
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
			
			// Validate name
			if name == "" {
				return fmt.Errorf("cookie name cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create delete cookie step using the enhanced client
			stepID, err := client.CreateDeleteCookieStep(checkpointID, name, position)
			if err != nil {
				return fmt.Errorf("failed to create delete cookie step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "DELETE_COOKIE",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"name":          name,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Delete cookie: %s", name),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: DELETE_COOKIE\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("name: %s\n", name)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Delete cookie: %s\n", name)
			case "ai":
				fmt.Printf("Successfully created delete cookie step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: DELETE_COOKIE\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Cookie Name: %s\n", name)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Delete cookie: %s\n", name)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created delete cookie step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Cookie Name: %s\n", name)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}