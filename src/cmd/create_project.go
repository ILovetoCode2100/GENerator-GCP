package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateProjectCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create-project NAME",
		Short: "Create a new Virtuoso project",
		Long: `Create a new project in Virtuoso with the specified name.

Example:
  api-cli create-project "My Test Project"
  api-cli create-project "My Test Project" --description "Project for testing"
  api-cli create-project "My Test Project" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)

			// Create the project
			project, err := client.CreateProject(projectName, description)
			if err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":     "success",
					"project_id": project.ID,
					"name":       project.Name,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				// Simple YAML output
				fmt.Printf("status: success\n")
				fmt.Printf("project_id: %d\n", project.ID)
				fmt.Printf("name: %s\n", project.Name)
			case "ai":
				// AI-friendly output
				fmt.Printf("Successfully created Virtuoso project:\n")
				fmt.Printf("- Project ID: %d\n", project.ID)
				fmt.Printf("- Project Name: %s\n", project.Name)
				fmt.Printf("- Organization ID: %d\n", project.OrganizationID)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Create a goal: api-cli create-goal %d \"Goal Name\"\n", project.ID)
				fmt.Printf("2. View project details: api-cli get-project %d\n", project.ID)
			default: // human
				fmt.Printf("✅ Created project '%s' with ID: %d\n", project.Name, project.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Description for the project")

	return cmd
}
