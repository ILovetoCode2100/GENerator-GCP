package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateCheckpointCmd() *cobra.Command {
	var position int

	cmd := &cobra.Command{
		Use:   "create-checkpoint JOURNEY_ID GOAL_ID SNAPSHOT_ID NAME",
		Short: "Create and attach a checkpoint to a journey",
		Long: `Create a new checkpoint (testcase) and automatically attach it to the specified journey.
This command performs both operations in sequence to ensure the checkpoint is properly configured.

Example:
  api-cli create-checkpoint 608038 13776 43802 "Login Test"
  api-cli create-checkpoint 608038 13776 43802 "Checkout Test" --position 3
  api-cli create-checkpoint 608038 13776 43802 "Payment Test" -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			journeyIDStr := args[0]
			goalIDStr := args[1]
			snapshotIDStr := args[2]
			checkpointName := args[3]

			// Convert IDs to int
			journeyID, err := strconv.Atoi(journeyIDStr)
			if err != nil {
				return fmt.Errorf("invalid journey ID: %w", err)
			}

			goalID, err := strconv.Atoi(goalIDStr)
			if err != nil {
				return fmt.Errorf("invalid goal ID: %w", err)
			}

			snapshotID, err := strconv.Atoi(snapshotIDStr)
			if err != nil {
				return fmt.Errorf("invalid snapshot ID: %w", err)
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create the checkpoint
			checkpoint, err := client.CreateCheckpoint(goalID, snapshotID, checkpointName)
			if err != nil {
				return fmt.Errorf("failed to create checkpoint: %w", err)
			}

			// Attach the checkpoint to the journey
			err = client.AttachCheckpoint(journeyID, checkpoint.ID, position)
			if err != nil {
				// Checkpoint was created but not attached
				switch cfg.Output.DefaultFormat {
				case "json":
					output := map[string]interface{}{
						"status":        "partial",
						"checkpoint_id": checkpoint.ID,
						"name":          checkpoint.Title,
						"error":         fmt.Sprintf("Checkpoint created but not attached: %v", err),
					}
					encoder := json.NewEncoder(os.Stdout)
					encoder.SetIndent("", "  ")
					encoder.Encode(output)
				default:
					fmt.Printf("⚠️  Created checkpoint '%s' with ID: %d, but failed to attach to journey: %v\n",
						checkpoint.Title, checkpoint.ID, err)
				}
				return err
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"checkpoint_id": checkpoint.ID,
					"journey_id":    journeyID,
					"position":      position,
					"name":          checkpoint.Title,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				// Simple YAML output
				fmt.Printf("status: success\n")
				fmt.Printf("checkpoint_id: %d\n", checkpoint.ID)
				fmt.Printf("journey_id: %d\n", journeyID)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("name: %s\n", checkpoint.Title)
			case "ai":
				// AI-friendly output
				fmt.Printf("Successfully created and attached checkpoint:\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpoint.ID)
				fmt.Printf("- Checkpoint Name: %s\n", checkpoint.Title)
				fmt.Printf("- Journey ID: %d\n", journeyID)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add steps to checkpoint: api-cli add-step %d <step-type> <step-details>\n", checkpoint.ID)
				fmt.Printf("2. Create another checkpoint: api-cli create-checkpoint %d %d %d \"Next Test\"\n",
					journeyID, goalID, snapshotID)
			default: // human
				fmt.Printf("✅ Created and attached checkpoint '%s' with ID: %d to journey %d at position %d\n",
					checkpoint.Title, checkpoint.ID, journeyID, position)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&position, "position", 2, "Position in journey (must be 2 or greater)")

	return cmd
}
