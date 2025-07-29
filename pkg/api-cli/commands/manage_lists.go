package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// ListConfig defines the configuration for a list operation
type ListConfig struct {
	ResourceType string
	Description  string
	LongDesc     string
	Headers      []string
	FormatFunc   func(item interface{}) []string
	ListFunc     func(ctx context.Context, c *client.Client, args []string, limit, offset int) ([]interface{}, error)
	AIHelpFunc   func(items []interface{}) string
}

// Generic list command builder
func newListCommand(config ListConfig) *cobra.Command {
	var (
		limit  int
		offset int
	)

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("list-%ss", config.ResourceType),
		Short: config.Description,
		Long:  config.LongDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout and signal handling
			ctx, cancel := CommandContext()
			defer cancel()

			// Create Virtuoso client
			c := client.NewClient(cfg)

			// Call the appropriate list function
			items, err := config.ListFunc(ctx, c, args, limit, offset)
			if err != nil {
				// Handle specific error types
				if err == context.Canceled {
					return fmt.Errorf("operation canceled by user")
				}
				if err == context.DeadlineExceeded {
					return fmt.Errorf("operation timed out while listing %ss", config.ResourceType)
				}
				if apiErr, ok := err.(*client.APIError); ok {
					switch apiErr.Status {
					case 401:
						return fmt.Errorf("authentication failed: please check your API token")
					case 403:
						return fmt.Errorf("permission denied: you don't have access to list %ss", config.ResourceType)
					case 429:
						return fmt.Errorf("rate limit exceeded: please try again later")
					default:
						return fmt.Errorf("API error listing %ss: %s", config.ResourceType, apiErr.Message)
					}
				}
				return fmt.Errorf("failed to list %ss: %w", config.ResourceType, err)
			}

			// Format output based on the format flag
			return formatListOutput(items, config, cfg.Output.DefaultFormat)
		},
	}

	// Add pagination flags
	cmd.Flags().IntVar(&limit, "limit", 50, fmt.Sprintf("Maximum number of %ss to return", config.ResourceType))
	cmd.Flags().IntVar(&offset, "offset", 0, fmt.Sprintf("Number of %ss to skip", config.ResourceType))

	return cmd
}

// formatListOutput handles all output formatting for list commands
func formatListOutput(items []interface{}, config ListConfig, format string) error {
	switch format {
	case "json":
		output := map[string]interface{}{
			"status":                  "success",
			"count":                   len(items),
			config.ResourceType + "s": items,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)

	case "yaml":
		fmt.Printf("status: success\n")
		fmt.Printf("count: %d\n", len(items))
		fmt.Printf("%ss:\n", config.ResourceType)
		for _, item := range items {
			fields := config.FormatFunc(item)
			for i, header := range config.Headers {
				if i < len(fields) && fields[i] != "" && fields[i] != "-" {
					fmt.Printf("  - %s: %s\n", header, fields[i])
				}
			}
		}

	case "ai":
		fmt.Printf("Found %d %ss:\n\n", len(items), config.ResourceType)
		for i, item := range items {
			fields := config.FormatFunc(item)
			fmt.Printf("%d. %s: %s\n", i+1, config.ResourceType, fields[1]) // Name is usually second field
			for j, header := range config.Headers {
				if j > 0 && j < len(fields) && fields[j] != "" && fields[j] != "-" {
					fmt.Printf("   - %s: %s\n", header, fields[j])
				}
			}
			fmt.Println()
		}
		if config.AIHelpFunc != nil {
			fmt.Print(config.AIHelpFunc(items))
		}

	default: // human
		if len(items) == 0 {
			fmt.Printf("No %ss found\n", config.ResourceType)
			return nil
		}

		// Create a table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		// Print header
		fmt.Fprintln(w, joinWithTabs(config.Headers...))
		fmt.Fprintln(w, createSeparator(len(config.Headers)))

		// Print items
		for _, item := range items {
			fields := config.FormatFunc(item)
			fmt.Fprintln(w, joinWithTabs(fields...))
		}

		w.Flush()
		fmt.Printf("\nTotal: %d %ss\n", len(items), config.ResourceType)
	}

	return nil
}

// Helper functions
func joinWithTabs(fields ...string) string {
	result := ""
	for i, field := range fields {
		if i > 0 {
			result += "\t"
		}
		result += field
	}
	return result
}

func createSeparator(count int) string {
	sep := ""
	for i := 0; i < count; i++ {
		if i > 0 {
			sep += "\t"
		}
		sep += "──"
	}
	return sep
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// List command implementations
func NewListProjectsCmd() *cobra.Command {
	return newListCommand(ListConfig{
		ResourceType: "project",
		Description:  "List all projects in the organization",
		LongDesc: `List all projects in your Virtuoso organization.

This command retrieves all projects accessible to your organization
and displays them in a table format by default. Supports pagination
for large result sets.`,
		Headers: []string{"ID", "NAME", "DESCRIPTION", "CREATED"},
		FormatFunc: func(item interface{}) []string {
			p := item.(*client.Project)
			created := "N/A"
			if !p.CreatedAt.IsZero() {
				created = p.CreatedAt.Format("2006-01-02")
			}
			description := p.Description
			if description == "" {
				description = "-"
			}
			return []string{
				fmt.Sprintf("%d", p.ID),
				p.Name,
				truncate(description, 40),
				created,
			}
		},
		ListFunc: func(ctx context.Context, c *client.Client, args []string, limit, offset int) ([]interface{}, error) {
			// Helper function to wrap ListProjectsWithOptions with context support
			// This can be replaced with c.ListProjectsWithOptionsContext(ctx, offset, limit) when available
			type result struct {
				projects []*client.Project
				err      error
			}
			resultChan := make(chan result, 1)

			go func() {
				projects, err := c.ListProjectsWithOptions(offset, limit)
				resultChan <- result{projects: projects, err: err}
			}()

			select {
			case res := <-resultChan:
				if res.err != nil {
					return nil, res.err
				}
				items := make([]interface{}, len(res.projects))
				for i, p := range res.projects {
					items[i] = p
				}
				return items, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
		AIHelpFunc: func(items []interface{}) string {
			if len(items) > 0 {
				p := items[0].(*client.Project)
				return fmt.Sprintf("\nNext steps:\n"+
					"1. List goals for a project: api-cli list-goals %d\n"+
					"2. Create a new goal: api-cli create-goal %d \"Goal Name\"\n", p.ID, p.ID)
			}
			return ""
		},
	})
}

func NewListGoalsCmd() *cobra.Command {
	return newListCommand(ListConfig{
		ResourceType: "goal",
		Description:  "List all goals for a project",
		LongDesc: `List all goals for a specific project.

This command retrieves all goals associated with a project
and displays them in a table format by default.`,
		Headers: []string{"ID", "NAME", "DESCRIPTION", "SNAPSHOT ID"},
		FormatFunc: func(item interface{}) []string {
			g := item.(*client.Goal)
			description := g.Description
			if description == "" {
				description = "-"
			}
			snapshotID := g.SnapshotID
			if snapshotID == "" && g.LatestSnapshotID != 0 {
				snapshotID = fmt.Sprintf("%d", g.LatestSnapshotID)
			}
			return []string{
				fmt.Sprintf("%d", g.ID),
				g.Name,
				truncate(description, 30),
				snapshotID,
			}
		},
		ListFunc: func(ctx context.Context, c *client.Client, args []string, limit, offset int) ([]interface{}, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("project ID is required")
			}
			projectID, err := strconv.Atoi(args[0])
			if err != nil {
				return nil, fmt.Errorf("invalid project ID: %w", err)
			}

			// Helper function to wrap ListGoals with context support
			// This can be replaced with c.ListGoalsContext(ctx, projectID) when available
			type result struct {
				goals []*client.Goal
				err   error
			}
			resultChan := make(chan result, 1)

			go func() {
				goals, err := c.ListGoals(projectID)
				resultChan <- result{goals: goals, err: err}
			}()

			select {
			case res := <-resultChan:
				if res.err != nil {
					return nil, res.err
				}
				items := make([]interface{}, len(res.goals))
				for i, g := range res.goals {
					items[i] = g
				}
				return items, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
		AIHelpFunc: func(items []interface{}) string {
			if len(items) > 0 {
				g := items[0].(*client.Goal)
				return fmt.Sprintf("\nNext steps:\n"+
					"1. List journeys for a goal: api-cli list-journeys %d %s\n"+
					"2. Create a new journey: api-cli create-journey %d %s \"Journey Name\"\n", g.ID, g.SnapshotID, g.ID, g.SnapshotID)
			}
			return ""
		},
	})
}

func NewListJourneysCmd() *cobra.Command {
	return newListCommand(ListConfig{
		ResourceType: "journey",
		Description:  "List all journeys for a goal",
		LongDesc: `List all journeys associated with a specific goal.

This command retrieves all journeys for a goal and displays
them in a table format by default.`,
		Headers: []string{"ID", "NAME", "TITLE", "STATUS"},
		FormatFunc: func(item interface{}) []string {
			j := item.(*client.Journey)
			status := "active"
			if j.Archived {
				status = "archived"
			} else if j.Draft {
				status = "draft"
			}
			title := j.Title
			if title == "" {
				title = "-"
			}
			return []string{
				fmt.Sprintf("%d", j.ID),
				j.Name,
				truncate(title, 30),
				status,
			}
		},
		ListFunc: func(ctx context.Context, c *client.Client, args []string, limit, offset int) ([]interface{}, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("goal ID and snapshot ID are required")
			}
			goalID, err := strconv.Atoi(args[0])
			if err != nil {
				return nil, fmt.Errorf("invalid goal ID: %w", err)
			}
			snapshotID, err := strconv.Atoi(args[1])
			if err != nil {
				return nil, fmt.Errorf("invalid snapshot ID: %w", err)
			}
			journeys, err := c.ListJourneys(goalID, snapshotID)
			if err != nil {
				return nil, fmt.Errorf("failed to list journeys for goal %d and snapshot %d: %w", goalID, snapshotID, err)
			}
			items := make([]interface{}, len(journeys))
			for i, j := range journeys {
				items[i] = j
			}
			return items, nil
		},
		AIHelpFunc: func(items []interface{}) string {
			if len(items) > 0 {
				j := items[0].(*client.Journey)
				return fmt.Sprintf("\nNext steps:\n"+
					"1. List checkpoints for a journey: api-cli list-checkpoints %d\n"+
					"2. Create a new checkpoint: api-cli create-checkpoint %d %d %d \"Checkpoint Title\"\n", j.ID, j.ID, j.GoalID, j.SnapshotID)
			}
			return ""
		},
	})
}

func NewListCheckpointsCmd() *cobra.Command {
	return newListCommand(ListConfig{
		ResourceType: "checkpoint",
		Description:  "List all checkpoints for a journey",
		LongDesc: `List all checkpoints associated with a specific journey.

This command retrieves all checkpoints for a journey and displays
them in a table format by default.`,
		Headers: []string{"ID", "TITLE", "STEPS", "POSITION"},
		FormatFunc: func(item interface{}) []string {
			c := item.(*client.CheckpointDetail)
			stepCount := 0
			if c.Steps != nil {
				stepCount = len(c.Steps)
			}
			return []string{
				fmt.Sprintf("%d", c.ID),
				truncate(c.Title, 40),
				fmt.Sprintf("%d", stepCount),
				fmt.Sprintf("%d", c.Position),
			}
		},
		ListFunc: func(ctx context.Context, c *client.Client, args []string, limit, offset int) ([]interface{}, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("journey ID is required")
			}
			journeyID, err := strconv.Atoi(args[0])
			if err != nil {
				return nil, fmt.Errorf("invalid journey ID: %w", err)
			}
			journeyWithCheckpoints, err := c.ListCheckpoints(journeyID)
			if err != nil {
				return nil, fmt.Errorf("failed to list checkpoints for journey %d: %w", journeyID, err)
			}
			// Extract checkpoints from the journey
			if journeyWithCheckpoints == nil || journeyWithCheckpoints.Cases == nil {
				return []interface{}{}, nil
			}
			items := make([]interface{}, len(journeyWithCheckpoints.Cases))
			for i, cp := range journeyWithCheckpoints.Cases {
				items[i] = &cp
			}
			return items, nil
		},
		AIHelpFunc: func(items []interface{}) string {
			if len(items) > 0 {
				c := items[0].(*client.CheckpointDetail)
				return fmt.Sprintf("\nNext steps:\n"+
					"1. Add a step to checkpoint: api-cli navigate to cp_%d \"https://example.com\"\n"+
					"2. Get checkpoint details: api-cli get-step %d\n", c.ID, c.ID)
			}
			return ""
		},
	})
}

func NewListCheckpointStepsCmd() *cobra.Command {
	return newListCommand(ListConfig{
		ResourceType: "checkpoint-step",
		Description:  "List all steps in a checkpoint",
		LongDesc: `List all steps in a specific checkpoint.

This command retrieves all steps for a checkpoint and displays
them in a table format by default, showing the step details and order.`,
		Headers: []string{"ID", "POS", "ACTION", "SELECTOR/VALUE", "OPTIONAL", "SKIP"},
		FormatFunc: func(item interface{}) []string {
			s := item.(*client.Step)
			value := s.Value
			if value == "" {
				value = "-"
			}
			// Truncate long values
			if len(value) > 40 {
				value = value[:37] + "..."
			}
			return []string{
				fmt.Sprintf("%d", s.ID),
				fmt.Sprintf("%d", s.StepIndex),
				s.Action,
				value,
				fmt.Sprintf("%t", s.Optional),
				fmt.Sprintf("%t", s.Skip),
			}
		},
		ListFunc: func(ctx context.Context, c *client.Client, args []string, limit, offset int) ([]interface{}, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("checkpoint ID is required")
			}

			// Handle both numeric and cp_ prefixed IDs
			checkpointIDStr := args[0]
			if len(checkpointIDStr) > 3 && checkpointIDStr[:3] == "cp_" {
				checkpointIDStr = checkpointIDStr[3:]
			}

			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid checkpoint ID: %w", err)
			}

			// Use the context-aware method
			steps, err := c.ListCheckpointStepsWithContext(ctx, checkpointID)
			if err != nil {
				return nil, err
			}

			items := make([]interface{}, len(steps))
			for i := range steps {
				items[i] = &steps[i]
			}
			return items, nil
		},
		AIHelpFunc: func(items []interface{}) string {
			if len(items) > 0 {
				s := items[0].(*client.Step)
				return fmt.Sprintf("\nCheckpoint has %d steps. You can:\n"+
					"1. Add more steps: api-cli step-navigate to %d \"https://example.com\"\n"+
					"2. Run the test: api-cli execute-goal <goal-id>\n", len(items), s.CheckpointID)
			}
			return "\nNo steps found in this checkpoint. Add steps using step commands."
		},
	})
}
