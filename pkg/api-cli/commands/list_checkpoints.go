package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

func newListCheckpointsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-checkpoints JOURNEY_ID",
		Short: "List all checkpoints in a Virtuoso journey",
		Long: `List all checkpoints (testcases) in a Virtuoso journey.

This command shows the checkpoints in order, identifies which checkpoint contains
the shared navigation step, and displays the step count for each checkpoint.

Example:
  api-cli list-checkpoints 608048
  api-cli list-checkpoints 608048 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			journeyIDStr := args[0]

			// Convert journey ID to int
			journeyID, err := strconv.Atoi(journeyIDStr)
			if err != nil {
				return fmt.Errorf("invalid journey ID: %w", err)
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Get journey with checkpoints
			journey, err := client.ListCheckpoints(journeyID)
			if err != nil {
				return fmt.Errorf("failed to list checkpoints: %w", err)
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"journey_id":   journey.ID,
					"journey_name": journey.Name,
					"checkpoints": func() []map[string]interface{} {
						checkpoints := make([]map[string]interface{}, len(journey.Cases))
						for i, cp := range journey.Cases {
							checkpoints[i] = map[string]interface{}{
								"id":            cp.ID,
								"position":      cp.Position,
								"title":         cp.Title,
								"step_count":    len(cp.Steps),
								"is_navigation": cp.Position == 1, // First checkpoint has navigation
							}
						}
						return checkpoints
					}(),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				// Simple YAML output
				fmt.Printf("journey_id: %d\n", journey.ID)
				fmt.Printf("journey_name: %s\n", journey.Name)
				fmt.Printf("checkpoints:\n")
				for _, cp := range journey.Cases {
					fmt.Printf("  - id: %d\n", cp.ID)
					fmt.Printf("    position: %d\n", cp.Position)
					fmt.Printf("    title: %s\n", cp.Title)
					fmt.Printf("    step_count: %d\n", len(cp.Steps))
					fmt.Printf("    is_navigation: %t\n", cp.Position == 1)
				}
			case "ai":
				// AI-friendly output
				fmt.Printf("Virtuoso Journey Checkpoints:\n")
				fmt.Printf("Journey: %s (ID: %d)\n", journey.Name, journey.ID)
				fmt.Printf("Total Checkpoints: %d\n\n", len(journey.Cases))

				for _, cp := range journey.Cases {
					fmt.Printf("Checkpoint %d:\n", cp.Position)
					fmt.Printf("- ID: %d\n", cp.ID)
					fmt.Printf("- Title: %s\n", cp.Title)
					fmt.Printf("- Steps: %d\n", len(cp.Steps))
					if cp.Position == 1 {
						fmt.Printf("- Note: Contains shared navigation step\n")
					}
					fmt.Printf("\n")
				}

				fmt.Printf("Usage:\n")
				fmt.Printf("- To add steps to a checkpoint: api-cli add-step <checkpoint-id> <step-type>\n")
				fmt.Printf("- The first checkpoint typically contains the navigation step shared by all tests\n")
			default: // human
				fmt.Printf("Journey: %s (ID: %d)\n", journey.Name, journey.ID)
				fmt.Printf("Checkpoints:\n")
				for _, cp := range journey.Cases {
					navigationMarker := ""
					if cp.Position == 1 {
						navigationMarker = " [Navigation]"
					}
					stepWord := "step"
					if len(cp.Steps) != 1 {
						stepWord = "steps"
					}
					fmt.Printf("%d. %s (ID: %d)%s - %d %s\n",
						cp.Position, cp.Title, cp.ID, navigationMarker, len(cp.Steps), stepWord)
				}
			}

			return nil
		},
	}

	return cmd
}
