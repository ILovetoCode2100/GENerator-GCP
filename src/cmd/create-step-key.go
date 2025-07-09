package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-key CHECKPOINT_ID KEY POSITION",
		Short: "Create a keyboard press step at a specific position in a checkpoint",
		Long: `Create a keyboard press step that presses a specific key at the specified position in the checkpoint.
		
Example:
  api-cli create-step-key 1678318 "Enter" 1
  api-cli create-step-key 1678318 "Tab" 2 -o json
  api-cli create-step-key 1678318 "Escape" 3`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			key := args[1]
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
			
			// Validate key
			if key == "" {
				return fmt.Errorf("key cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create key step using the enhanced client
			stepID, err := client.CreateKeyStep(checkpointID, key, position)
			if err != nil {
				return fmt.Errorf("failed to create key step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "KEY",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"key":           key,
					"position":      position,
					"parsed_step":   fmt.Sprintf("press %s", key),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: KEY\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("key: %s\n", key)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: press %s\n", key)
			case "ai":
				fmt.Printf("Successfully created key step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: KEY\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Key: %s\n", key)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: press %s\n", key)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created key step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Key: %s\n", key)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}