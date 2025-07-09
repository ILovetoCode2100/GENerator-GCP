package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-upload CHECKPOINT_ID FILENAME ELEMENT POSITION",
		Short: "Create a file upload step at a specific position in a checkpoint",
		Long: `Create a file upload step that uploads a file to a specific element at the specified position in the checkpoint.
		
Example:
  api-cli create-step-upload 1678318 "document.pdf" "file upload" 1
  api-cli create-step-upload 1678318 "image.jpg" "#file-input" 2 -o json
  api-cli create-step-upload 1678318 "data.csv" "input[type='file']" 3`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			filename := args[1]
			element := args[2]
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
			
			// Validate inputs
			if filename == "" {
				return fmt.Errorf("filename cannot be empty")
			}
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create upload step using the enhanced client
			stepID, err := client.CreateUploadStep(checkpointID, filename, element, position)
			if err != nil {
				return fmt.Errorf("failed to create upload step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "UPLOAD",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"filename":      filename,
					"element":       element,
					"position":      position,
					"parsed_step":   fmt.Sprintf("upload \"%s\" to %s", filename, element),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: UPLOAD\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("filename: %s\n", filename)
				fmt.Printf("element: %s\n", element)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: upload \"%s\" to %s\n", filename, element)
			case "ai":
				fmt.Printf("Successfully created upload step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: UPLOAD\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Filename: %s\n", filename)
				fmt.Printf("- Element: %s\n", element)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: upload \"%s\" to %s\n", filename, element)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created upload step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Filename: %s\n", filename)
				fmt.Printf("   Element: %s\n", element)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}