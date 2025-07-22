package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// Helper function to wrap CreateProject with context support
// This can be removed once CreateProjectWithContext is available in the client
func createProjectWithContext(ctx context.Context, c *client.Client, name, description string) (*client.Project, error) {
	// Create a channel to receive the result
	type result struct {
		project *client.Project
		err     error
	}
	resultChan := make(chan result, 1)

	// Run the API call in a goroutine
	go func() {
		project, err := c.CreateProject(name, description)
		resultChan <- result{project: project, err: err}
	}()

	// Wait for either the result or context cancellation
	select {
	case res := <-resultChan:
		return res.project, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

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

			// Create context with timeout and signal handling
			ctx, cancel := CommandContext()
			defer cancel()

			// Create Virtuoso client
			apiClient := client.NewClient(cfg)

			// Create the project
			// Note: When CreateProjectWithContext is available, use it instead
			// project, err := apiClient.CreateProjectWithContext(ctx, projectName, description)
			project, err := createProjectWithContext(ctx, apiClient, projectName, description)
			if err != nil {
				// Handle different error types with appropriate exit codes
				if apiErr, ok := err.(*client.APIError); ok {
					switch apiErr.Status {
					case 401:
						return fmt.Errorf("authentication failed: please check your API token")
					case 403:
						return fmt.Errorf("permission denied: you don't have access to create projects")
					case 409:
						return fmt.Errorf("conflict: a project with this name may already exist")
					case 429:
						return fmt.Errorf("rate limit exceeded: please try again later")
					default:
						return fmt.Errorf("API error creating project: %s", apiErr.Message)
					}
				}
				if err == context.Canceled {
					return fmt.Errorf("operation canceled by user")
				}
				if err == context.DeadlineExceeded {
					return fmt.Errorf("operation timed out")
				}
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

	// Wrap the command to handle exit codes properly
	originalRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := originalRunE(cmd, args)
		if err != nil {
			// Log the error for debugging
			if cfg.Output.DefaultFormat == "human" {
				fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			}
		}
		// Note: SetExitCode would be called by the main command runner
		// For now, we just return the error and let cobra handle it
		return err
	}

	return cmd
}
