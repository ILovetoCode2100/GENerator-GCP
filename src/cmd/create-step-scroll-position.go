package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepScrollPositionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-scroll-position CHECKPOINT_ID X Y POSITION",
		Short: "Create a scroll position step at a specific position in a checkpoint",
		Long: `Create a scroll position step that scrolls to specific X and Y coordinates at the specified position in the checkpoint.
		
X and Y coordinates can be negative. Use -- before negative values to avoid flag parsing issues.
		
Example:
  api-cli create-step-scroll-position 1678318 100 200 1
  api-cli create-step-scroll-position 1678318 0 500 2 -o json
  api-cli create-step-scroll-position 1678318 -- -10 -20 3  # Use -- for negative coordinates`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			xStr := args[1]
			yStr := args[2]
			positionStr := args[3]
			
			// Convert IDs to int using helper function
			checkpointID, err := parseIntArg(checkpointIDStr, "checkpoint ID")
			if err != nil {
				return err
			}
			
			x, err := parseIntArg(xStr, "X coordinate")
			if err != nil {
				return err
			}
			
			y, err := parseIntArg(yStr, "Y coordinate")
			if err != nil {
				return err
			}
			
			position, err := parseIntArg(positionStr, "position")
			if err != nil {
				return err
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create scroll position step using the enhanced client
			stepID, err := client.CreateScrollPositionStep(checkpointID, x, y, position)
			if err != nil {
				return fmt.Errorf("failed to create scroll position step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "SCROLL_POSITION",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"x":             x,
					"y":             y,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Scroll to position (%d, %d)", x, y),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: SCROLL_POSITION\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("x: %d\n", x)
				fmt.Printf("y: %d\n", y)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Scroll to position (%d, %d)\n", x, y)
			case "ai":
				fmt.Printf("Successfully created scroll position step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: SCROLL_POSITION\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- X Coordinate: %d\n", x)
				fmt.Printf("- Y Coordinate: %d\n", y)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Scroll to position (%d, %d)\n", x, y)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created scroll position step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   X: %d, Y: %d\n", x, y)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}