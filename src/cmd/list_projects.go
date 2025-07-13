package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newListProjectsCmd() *cobra.Command {
	var (
		limit  int
		offset int
	)

	cmd := &cobra.Command{
		Use:   "list-projects",
		Short: "List all projects in the organization",
		Long: `List all projects in your Virtuoso organization.

This command retrieves all projects accessible to your organization
and displays them in a table format by default. Supports pagination
for large result sets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// List projects with pagination
			projects, err := client.ListProjectsWithOptions(offset, limit)
			if err != nil {
				return fmt.Errorf("failed to list projects: %w", err)
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":   "success",
					"count":    len(projects),
					"projects": projects,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}

			case "yaml":
				// Simple YAML output
				fmt.Printf("status: success\n")
				fmt.Printf("count: %d\n", len(projects))
				fmt.Printf("projects:\n")
				for _, p := range projects {
					fmt.Printf("  - id: %d\n", p.ID)
					fmt.Printf("    name: %s\n", p.Name)
					if p.Description != "" {
						fmt.Printf("    description: %s\n", p.Description)
					}
					fmt.Printf("    organization_id: %d\n", p.OrganizationID)
					if !p.CreatedAt.IsZero() {
						fmt.Printf("    created_at: %s\n", p.CreatedAt.Format(time.RFC3339))
					}
				}

			case "ai":
				// AI-friendly output
				fmt.Printf("Found %d projects in organization %s:\n\n", len(projects), cfg.Org.ID)
				for i, p := range projects {
					fmt.Printf("%d. Project: %s\n", i+1, p.Name)
					fmt.Printf("   - ID: %d\n", p.ID)
					if p.Description != "" {
						fmt.Printf("   - Description: %s\n", p.Description)
					}
					if !p.CreatedAt.IsZero() {
						fmt.Printf("   - Created: %s\n", p.CreatedAt.Format("Jan 02, 2006"))
					}
					fmt.Println()
				}
				if len(projects) > 0 {
					fmt.Printf("Next steps:\n")
					fmt.Printf("1. List goals for a project: api-cli list-goals %d\n", projects[0].ID)
					fmt.Printf("2. Create a new goal: api-cli create-goal %d \"Goal Name\"\n", projects[0].ID)
				}

			default: // human
				if len(projects) == 0 {
					fmt.Println("No projects found in organization")
					return nil
				}

				// Create a table writer
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

				// Print header
				fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tCREATED")
				fmt.Fprintln(w, "──\t────\t───────────\t───────")

				// Print projects
				for _, p := range projects {
					created := "N/A"
					if !p.CreatedAt.IsZero() {
						created = p.CreatedAt.Format("2006-01-02")
					}

					description := p.Description
					if description == "" {
						description = "-"
					}
					if len(description) > 40 {
						description = description[:37] + "..."
					}

					fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", p.ID, p.Name, description, created)
				}

				w.Flush()
				fmt.Printf("\nTotal: %d projects\n", len(projects))
			}

			return nil
		},
	}

	// Add pagination flags
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of projects to return")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of projects to skip")

	return cmd
}
