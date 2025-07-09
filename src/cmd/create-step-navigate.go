package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepNavigateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-navigate CHECKPOINT_ID URL POSITION",
		Short: "Create a navigation step at a specific position in a checkpoint",
		Long: `Create a navigation step that goes to a specific URL at the specified position in the checkpoint.
		
Example:
  api-cli create-step-navigate 1678318 "https://example.com" 1
  api-cli create-step-navigate 1678318 "https://dashboard.example.com" 2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			url := args[1]
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
			
			// Validate URL
			if url == "" {
				return fmt.Errorf("URL cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create navigation step using the enhanced client
			stepID, err := client.CreateNavigationStep(checkpointID, url, position)
			if err != nil {
				return fmt.Errorf("failed to create navigation step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "NAVIGATE",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"url":           url,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Navigate to \"%s\"", url),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: NAVIGATE\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("url: %s\n", url)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Navigate to \"%s\"\n", url)
			case "ai":
				fmt.Printf("Successfully created navigation step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: NAVIGATE\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- URL: %s\n", url)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Navigate to \"%s\"\n", url)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created navigation step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   URL: %s\n", url)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}