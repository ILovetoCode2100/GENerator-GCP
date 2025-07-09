package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepExecuteJsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-execute-js CHECKPOINT_ID JAVASCRIPT POSITION",
		Short: "Create an execute JavaScript step at a specific position in a checkpoint",
		Long: `Create an execute JavaScript step that runs custom JavaScript code at the specified position in the checkpoint.
		
Example:
  api-cli create-step-execute-js 1678318 "window.scrollTo(0, 0)" 1
  api-cli create-step-execute-js 1678318 "document.querySelector('#modal').style.display = 'none'" 2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			javascript := args[1]
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
			
			// Validate JavaScript
			if javascript == "" {
				return fmt.Errorf("javascript cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create execute JavaScript step using the enhanced client
			stepID, err := client.CreateExecuteJsStep(checkpointID, javascript, position)
			if err != nil {
				return fmt.Errorf("failed to create execute JavaScript step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "EXECUTE_JS",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"javascript":    javascript,
					"position":      position,
					"parsed_step":   fmt.Sprintf("execute JS \"%s\"", javascript),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: EXECUTE_JS\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("javascript: %s\n", javascript)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: execute JS \"%s\"\n", javascript)
			case "ai":
				fmt.Printf("Successfully created execute JavaScript step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: EXECUTE_JS\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- JavaScript: %s\n", javascript)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: execute JS \"%s\"\n", javascript)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created execute JavaScript step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   JavaScript: %s\n", javascript)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}