package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateGoalCmd() *cobra.Command {
	var url string

	cmd := &cobra.Command{
		Use:   "create-goal PROJECT_ID NAME",
		Short: "Create a new Virtuoso goal with automatic initial journey",
		Long: `Create a new goal in Virtuoso for the specified project.
This command automatically creates an initial journey and retrieves the snapshot ID.

Example:
  api-cli create-goal 123 "My Test Goal"
  api-cli create-goal 123 "My Test Goal" --url "https://example.com"
  api-cli create-goal 123 "My Test Goal" -o json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectIDStr := args[0]
			goalName := args[1]

			// Convert project ID to int
			projectID, err := strconv.Atoi(projectIDStr)
			if err != nil {
				return fmt.Errorf("invalid project ID: %w", err)
			}

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create the goal
			goal, err := client.CreateGoal(projectID, goalName, url)
			if err != nil {
				return fmt.Errorf("failed to create goal: %w", err)
			}

			// Get the snapshot ID
			snapshotID, err := client.GetGoalSnapshot(goal.ID)
			if err != nil {
				return fmt.Errorf("failed to get snapshot ID: %w", err)
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":      "success",
					"goal_id":     goal.ID,
					"snapshot_id": snapshotID,
					"name":        goal.Name,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				// Simple YAML output
				fmt.Printf("status: success\n")
				fmt.Printf("goal_id: %d\n", goal.ID)
				fmt.Printf("snapshot_id: %s\n", snapshotID)
				fmt.Printf("name: %s\n", goal.Name)
			case "ai":
				// AI-friendly output
				fmt.Printf("Successfully created Virtuoso goal:\n")
				fmt.Printf("- Goal ID: %d\n", goal.ID)
				fmt.Printf("- Goal Name: %s\n", goal.Name)
				fmt.Printf("- Snapshot ID: %s\n", snapshotID)
				fmt.Printf("- URL: %s\n", url)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Create a journey: api-cli create-journey %d %s \"Journey Name\"\n", goal.ID, snapshotID)
				fmt.Printf("2. View goal details: api-cli get-goal %d\n", goal.ID)
			default: // human
				fmt.Printf("✅ Created goal '%s' with ID: %d, Snapshot ID: %s\n", goal.Name, goal.ID, snapshotID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&url, "url", "https://www.example.com", "URL for the goal")

	return cmd
}
