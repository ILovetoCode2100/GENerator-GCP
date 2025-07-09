package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newListGoalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-goals PROJECT_ID",
		Short: "List all goals in a project",
		Long: `List all non-archived goals in a Virtuoso project.
		
This command retrieves all active goals for the specified project
and displays them in a table format by default.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectIDStr := args[0]
			
			// Convert project ID to int
			projectID, err := strconv.Atoi(projectIDStr)
			if err != nil {
				return fmt.Errorf("invalid project ID: %w", err)
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// List goals
			goals, err := client.ListGoals(projectID)
			if err != nil {
				return fmt.Errorf("failed to list goals: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":     "success",
					"project_id": projectID,
					"count":      len(goals),
					"goals":      goals,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
				
			case "yaml":
				// Simple YAML output
				fmt.Printf("status: success\n")
				fmt.Printf("project_id: %d\n", projectID)
				fmt.Printf("count: %d\n", len(goals))
				fmt.Printf("goals:\n")
				for _, g := range goals {
					fmt.Printf("  - id: %d\n", g.ID)
					fmt.Printf("    name: %s\n", g.Name)
					if g.URL != "" {
						fmt.Printf("    url: %s\n", g.URL)
					}
					if g.Description != "" {
						fmt.Printf("    description: %s\n", g.Description)
					}
					if g.SnapshotID != "" {
						fmt.Printf("    snapshot_id: %s\n", g.SnapshotID)
					}
				}
				
			case "ai":
				// AI-friendly output
				fmt.Printf("Found %d goals in project %d:\n\n", len(goals), projectID)
				for i, g := range goals {
					fmt.Printf("%d. Goal: %s\n", i+1, g.Name)
					fmt.Printf("   - ID: %d\n", g.ID)
					if g.URL != "" {
						fmt.Printf("   - URL: %s\n", g.URL)
					}
					if g.Description != "" {
						fmt.Printf("   - Description: %s\n", g.Description)
					}
					if g.SnapshotID != "" {
						fmt.Printf("   - Snapshot ID: %s\n", g.SnapshotID)
					}
					fmt.Println()
				}
				if len(goals) > 0 {
					fmt.Printf("Next steps:\n")
					// For listing journeys, we need to get the snapshot ID first
					if goals[0].SnapshotID != "" {
						snapshotID, _ := strconv.Atoi(goals[0].SnapshotID)
						fmt.Printf("1. List journeys: api-cli list-journeys %d %d\n", goals[0].ID, snapshotID)
					} else {
						fmt.Printf("1. Get goal snapshot: Use API to get snapshot ID for goal %d\n", goals[0].ID)
					}
					fmt.Printf("2. Create a new journey: api-cli create-journey %d <snapshot-id> \"Journey Name\"\n", goals[0].ID)
				}
				
			default: // human
				if len(goals) == 0 {
					fmt.Printf("No goals found in project %d\n", projectID)
					return nil
				}
				
				// Create a table writer
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
				
				// Print header
				fmt.Fprintln(w, "ID\tNAME\tURL\tSNAPSHOT ID")
				fmt.Fprintln(w, "──\t────\t───\t───────────")
				
				// Print goals
				for _, g := range goals {
					url := g.URL
					if url == "" {
						url = "-"
					}
					if len(url) > 40 {
						url = url[:37] + "..."
					}
					
					snapshotID := g.SnapshotID
					if snapshotID == "" {
						snapshotID = "-"
					}
					
					fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", g.ID, g.Name, url, snapshotID)
				}
				
				w.Flush()
				fmt.Printf("\nTotal: %d goals in project %d\n", len(goals), projectID)
			}
			
			return nil
		},
	}
	
	return cmd
}