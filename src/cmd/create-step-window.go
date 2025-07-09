package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepWindowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-window CHECKPOINT_ID WIDTH HEIGHT POSITION",
		Short: "Create a window resize step at a specific position in a checkpoint",
		Long: `Create a window resize step that sets the browser window size to the specified dimensions at the specified position in the checkpoint.
		
Example:
  api-cli create-step-window 1678318 1920 1080 1
  api-cli create-step-window 1678318 1280 800 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			widthStr := args[1]
			heightStr := args[2]
			positionStr := args[3]
			
			// Convert IDs to int
			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}
			
			width, err := strconv.Atoi(widthStr)
			if err != nil {
				return fmt.Errorf("invalid width: %w", err)
			}
			
			height, err := strconv.Atoi(heightStr)
			if err != nil {
				return fmt.Errorf("invalid height: %w", err)
			}
			
			position, err := strconv.Atoi(positionStr)
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			
			// Validate dimensions
			if width <= 0 || height <= 0 {
				return fmt.Errorf("width and height must be greater than 0")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create window resize step using the enhanced client
			stepID, err := client.CreateWindowResizeStep(checkpointID, width, height, position)
			if err != nil {
				return fmt.Errorf("failed to create window resize step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "WINDOW_RESIZE",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"width":         width,
					"height":        height,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Set browser window size to %dx%d", width, height),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: WINDOW_RESIZE\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("width: %d\n", width)
				fmt.Printf("height: %d\n", height)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Set browser window size to %dx%d\n", width, height)
			case "ai":
				fmt.Printf("Successfully created window resize step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: WINDOW_RESIZE\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Width: %d\n", width)
				fmt.Printf("- Height: %d\n", height)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Set browser window size to %dx%d\n", width, height)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created window resize step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Dimensions: %dx%d\n", width, height)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}