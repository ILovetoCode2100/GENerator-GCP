package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newCreateJourneyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-journey GOAL_ID SNAPSHOT_ID NAME",
		Short: "Create a new Virtuoso journey (testsuite)",
		Long: `Create a new journey (testsuite) in Virtuoso for the specified goal.

Example:
  api-cli create-journey 13776 43802 "My Test Journey"
  api-cli create-journey 13776 43802 "My Test Journey" -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			goalIDStr := args[0]
			snapshotIDStr := args[1]
			journeyName := args[2]

			// Convert goal ID to int
			goalID, err := strconv.Atoi(goalIDStr)
			if err != nil {
				return fmt.Errorf("invalid goal ID: %w", err)
			}

			// Convert snapshot ID to int
			snapshotID, err := strconv.Atoi(snapshotIDStr)
			if err != nil {
				return fmt.Errorf("invalid snapshot ID: %w", err)
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Create the journey
			journey, err := client.CreateJourney(goalID, snapshotID, journeyName)
			if err != nil {
				return fmt.Errorf("failed to create journey: %w", err)
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":     "success",
					"journey_id": journey.ID,
					"name":       journey.Name,
					"goal_id":    goalID,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				// Simple YAML output
				fmt.Printf("status: success\n")
				fmt.Printf("journey_id: %d\n", journey.ID)
				fmt.Printf("name: %s\n", journey.Name)
				fmt.Printf("goal_id: %d\n", goalID)
			case "ai":
				// AI-friendly output
				fmt.Printf("Successfully created Virtuoso journey:\n")
				fmt.Printf("- Journey ID: %d\n", journey.ID)
				fmt.Printf("- Journey Name: %s\n", journey.Name)
				fmt.Printf("- Goal ID: %d\n", goalID)
				fmt.Printf("- Snapshot ID: %d\n", snapshotID)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Create a checkpoint: api-cli create-checkpoint %d \"Checkpoint Name\"\n", journey.ID)
				fmt.Printf("2. Attach checkpoint to journey: api-cli attach-checkpoint %d <checkpoint-id>\n", journey.ID)
				fmt.Printf("3. Add steps to checkpoint: api-cli add-step <checkpoint-id> <step-type>\n")
			default: // human
				fmt.Printf("✅ Created journey '%s' with ID: %d\n", journey.Name, journey.ID)
			}

			return nil
		},
	}

	return cmd
}
