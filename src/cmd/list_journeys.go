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

func newListJourneysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-journeys GOAL_ID SNAPSHOT_ID",
		Short: "List all journeys in a goal",
		Long: `List all journeys (testsuites) for a specific goal and snapshot.
		
This command retrieves all journeys associated with the specified goal
and snapshot, displaying their current status.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			goalIDStr := args[0]
			snapshotIDStr := args[1]
			
			// Convert IDs to int
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
			
			// List journeys
			journeys, err := client.ListJourneys(goalID, snapshotID)
			if err != nil {
				return fmt.Errorf("failed to list journeys: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":      "success",
					"goal_id":     goalID,
					"snapshot_id": snapshotID,
					"count":       len(journeys),
					"journeys":    journeys,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
				
			case "yaml":
				// Simple YAML output
				fmt.Printf("status: success\n")
				fmt.Printf("goal_id: %d\n", goalID)
				fmt.Printf("snapshot_id: %d\n", snapshotID)
				fmt.Printf("count: %d\n", len(journeys))
				fmt.Printf("journeys:\n")
				for _, j := range journeys {
					fmt.Printf("  - id: %d\n", j.ID)
					fmt.Printf("    name: %s\n", j.Name)
					fmt.Printf("    title: %s\n", j.Title)
					fmt.Printf("    archived: %t\n", j.Archived)
					fmt.Printf("    draft: %t\n", j.Draft)
				}
				
			case "ai":
				// AI-friendly output
				fmt.Printf("Found %d journeys in goal %d (snapshot %d):\n\n", len(journeys), goalID, snapshotID)
				for i, j := range journeys {
					fmt.Printf("%d. Journey: %s\n", i+1, j.Title)
					fmt.Printf("   - ID: %d\n", j.ID)
					fmt.Printf("   - Name: %s\n", j.Name)
					fmt.Printf("   - Status: ")
					if j.Archived {
						fmt.Printf("Archived")
					} else if j.Draft {
						fmt.Printf("Draft")
					} else {
						fmt.Printf("Active")
					}
					fmt.Printf("\n")
					fmt.Println()
				}
				if len(journeys) > 0 {
					fmt.Printf("Next steps:\n")
					fmt.Printf("1. Create checkpoint: api-cli create-checkpoint %d %d %d \"Checkpoint Name\"\n", 
						journeys[0].ID, goalID, snapshotID)
					fmt.Printf("2. Create another journey: api-cli create-journey %d %d \"New Journey\"\n", 
						goalID, snapshotID)
				}
				
			default: // human
				if len(journeys) == 0 {
					fmt.Printf("No journeys found for goal %d (snapshot %d)\n", goalID, snapshotID)
					return nil
				}
				
				// Create a table writer
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
				
				// Print header
				fmt.Fprintln(w, "ID\tNAME\tTITLE\tSTATUS")
				fmt.Fprintln(w, "──\t────\t─────\t──────")
				
				// Print journeys
				for _, j := range journeys {
					status := "Active"
					if j.Archived {
						status = "Archived"
					} else if j.Draft {
						status = "Draft"
					}
					
					title := j.Title
					if title == "" {
						title = j.Name
					}
					if len(title) > 30 {
						title = title[:27] + "..."
					}
					
					fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", j.ID, j.Name, title, status)
				}
				
				w.Flush()
				fmt.Printf("\nTotal: %d journeys\n", len(journeys))
			}
			
			return nil
		},
	}
	
	return cmd
}