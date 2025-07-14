package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newUpdateJourneyCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update-journey JOURNEY_ID",
		Short: "Update a Virtuoso journey (testsuite) name",
		Long: `Update the name of an existing journey (testsuite) in Virtuoso.

Example:
  api-cli update-journey 12345 --name "Updated Journey Name"
  api-cli update-journey 12345 --name "New Name" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			journeyIDStr := args[0]

			// Validate that name flag was provided
			if name == "" {
				return fmt.Errorf("--name flag is required")
			}

			// Convert journey ID to int
			journeyID, err := strconv.Atoi(journeyIDStr)
			if err != nil {
				return fmt.Errorf("invalid journey ID: %w", err)
			}

			// Create Virtuoso client
			apiClient := client.NewClient(cfg)

			// Get current journey details for human output
			var originalJourney *client.Journey
			if cfg.Output.DefaultFormat == "human" || cfg.Output.DefaultFormat == "" {
				originalJourney, err = apiClient.GetJourney(journeyID)
				if err != nil {
					// Non-fatal error, we can still try to update
					fmt.Fprintf(os.Stderr, "Warning: Could not fetch current journey details: %v\n", err)
				}
			}

			// Update the journey
			journey, err := apiClient.UpdateJourney(journeyID, name)
			if err != nil {
				return fmt.Errorf("failed to update journey: %w", err)
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":     "success",
					"journey_id": journey.ID,
					"name":       journey.Name,
					"goal_id":    journey.GoalID,
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
				fmt.Printf("goal_id: %d\n", journey.GoalID)
			case "ai":
				// AI-friendly output
				fmt.Printf("Successfully updated Virtuoso journey:\n")
				fmt.Printf("- Journey ID: %d\n", journey.ID)
				fmt.Printf("- New Name: %s\n", journey.Name)
				if originalJourney != nil && originalJourney.Name != journey.Name {
					fmt.Printf("- Previous Name: %s\n", originalJourney.Name)
				}
				fmt.Printf("- Goal ID: %d\n", journey.GoalID)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. List journeys: api-cli list-journeys %d <snapshot-id>\n", journey.GoalID)
				fmt.Printf("2. Create checkpoint: api-cli create-checkpoint %d \"Checkpoint Name\"\n", journey.ID)
				fmt.Printf("3. Attach checkpoint: api-cli attach-checkpoint %d <checkpoint-id>\n", journey.ID)
			default: // human
				if originalJourney != nil && originalJourney.Name != journey.Name {
					fmt.Printf("✅ Updated journey %d\n", journey.ID)
					fmt.Printf("   Before: %s\n", originalJourney.Name)
					fmt.Printf("   After:  %s\n", journey.Name)
				} else {
					fmt.Printf("✅ Updated journey '%s' (ID: %d)\n", journey.Name, journey.ID)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&name, "name", "", "New name for the journey (required)")
	cmd.MarkFlagRequired("name")

	return cmd
}
