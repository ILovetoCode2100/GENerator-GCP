package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepWaitTimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-wait-time CHECKPOINT_ID SECONDS POSITION",
		Short: "Create a wait time step at a specific position in a checkpoint",
		Long: `Create a wait time step that waits for a specified number of seconds at the specified position in the checkpoint.
		
Example:
  api-cli create-step-wait-time 1678318 5 2
  api-cli create-step-wait-time 1678318 10 3 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			secondsStr := args[1]
			positionStr := args[2]
			
			// Convert IDs to int
			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}
			
			seconds, err := strconv.Atoi(secondsStr)
			if err != nil {
				return fmt.Errorf("invalid seconds: %w", err)
			}
			
			position, err := strconv.Atoi(positionStr)
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			
			// Validate seconds
			if seconds <= 0 {
				return fmt.Errorf("seconds must be greater than 0")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create wait time step using the enhanced client
			stepID, err := client.CreateWaitTimeStep(checkpointID, seconds, position)
			if err != nil {
				return fmt.Errorf("failed to create wait time step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "WAIT_TIME",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"seconds":       seconds,
					"position":      position,
					"parsed_step":   fmt.Sprintf("Wait %d seconds", seconds),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: WAIT_TIME\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("seconds: %d\n", seconds)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: Wait %d seconds\n", seconds)
			case "ai":
				fmt.Printf("Successfully created wait time step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: WAIT_TIME\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Seconds: %d\n", seconds)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: Wait %d seconds\n", seconds)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created wait time step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Wait duration: %d seconds\n", seconds)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}