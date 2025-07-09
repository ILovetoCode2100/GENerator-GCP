package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepDismissPromptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-dismiss-prompt CHECKPOINT_ID TEXT POSITION",
		Short: "Create a dismiss prompt step at a specific position in a checkpoint",
		Long: `Create a dismiss prompt step that dismisses a JavaScript prompt dialog with the specified text at the specified position in the checkpoint.
		
Example:
  api-cli create-step-dismiss-prompt 1678318 "John Doe" 1
  api-cli create-step-dismiss-prompt 1678318 "user@example.com" 2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			text := args[1]
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
			
			// Validate text
			if text == "" {
				return fmt.Errorf("text cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create dismiss prompt step using the enhanced client
			stepID, err := client.CreateDismissPromptStep(checkpointID, text, position)
			if err != nil {
				return fmt.Errorf("failed to create dismiss prompt step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "DISMISS_PROMPT",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"text":          text,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Dismiss prompt with text: %s", text),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: DISMISS_PROMPT\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("text: %s\n", text)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Dismiss prompt with text: %s\n", text)
			case "ai":
				fmt.Printf("Successfully created dismiss prompt step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: DISMISS_PROMPT\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Text: %s\n", text)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Dismiss prompt with text: %s\n", text)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created dismiss prompt step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Text: %s\n", text)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}