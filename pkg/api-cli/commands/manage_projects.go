package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// Common output format helper
func formatOutput(format string, data interface{}) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	case "yaml":
		// Convert to YAML format manually for simplicity
		if m, ok := data.(map[string]interface{}); ok {
			for k, v := range m {
				fmt.Printf("%s: %v\n", k, v)
			}
		}
		return nil
	default:
		// Human format is handled by the caller
		return nil
	}
}

// Common API error handling helper
func handleAPIError(err error, operation string) error {
	if apiErr, ok := err.(*client.APIError); ok {
		switch apiErr.Status {
		case 401:
			return fmt.Errorf("authentication failed: please check your API token")
		case 403:
			return fmt.Errorf("permission denied: you don't have access to %s", operation)
		case 404:
			return fmt.Errorf("not found: the requested resource does not exist")
		case 409:
			return fmt.Errorf("conflict: a resource with this name may already exist")
		case 429:
			return fmt.Errorf("rate limit exceeded: please try again later")
		default:
			return fmt.Errorf("API error during %s: %s", operation, apiErr.Message)
		}
	}
	if err == context.Canceled {
		return fmt.Errorf("operation canceled by user")
	}
	if err == context.DeadlineExceeded {
		return fmt.Errorf("operation timed out")
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// CREATE PROJECT command
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
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("project name is required\n\nExample:\n  api-cli create-project \"My Test Project\"")
			}
			if strings.TrimSpace(args[0]) == "" {
				return fmt.Errorf("project name cannot be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := strings.TrimSpace(args[0])

			ctx, cancel := CommandContext()
			defer cancel()

			apiClient := client.NewClient(cfg)

			project, err := callWithContext(ctx, func() (*client.Project, error) {
				return apiClient.CreateProject(projectName, description)
			})
			if err != nil {
				return handleAPIError(err, "create project")
			}

			// Format output
			switch cfg.Output.DefaultFormat {
			case "json":
				return formatOutput("json", map[string]interface{}{
					"status":     "success",
					"project_id": project.ID,
					"name":       project.Name,
				})
			case "yaml":
				return formatOutput("yaml", map[string]interface{}{
					"status":     "success",
					"project_id": project.ID,
					"name":       project.Name,
				})
			case "ai":
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

// CREATE GOAL command
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
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("requires exactly 2 arguments: PROJECT_ID NAME\n\nExample:\n  api-cli create-goal 123 \"My Test Goal\"")
			}
			if strings.TrimSpace(args[1]) == "" {
				return fmt.Errorf("goal name cannot be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := parseID(args[0], "project")
			if err != nil {
				return fmt.Errorf("failed to parse project ID '%s': %w", args[0], err)
			}
			goalName := strings.TrimSpace(args[1])

			// Validate URL if provided
			if url != "" {
				if err := ValidateURL(url); err != nil {
					return fmt.Errorf("invalid URL: %w\n\nURL must start with http:// or https://", err)
				}
			}

			apiClient := client.NewClient(cfg)

			goal, err := apiClient.CreateGoal(projectID, goalName, url)
			if err != nil {
				return handleAPIError(err, "create goal")
			}

			snapshotID, err := apiClient.GetGoalSnapshot(goal.ID)
			if err != nil {
				return handleAPIError(err, "get snapshot ID")
			}

			// Format output
			switch cfg.Output.DefaultFormat {
			case "json":
				return formatOutput("json", map[string]interface{}{
					"status":      "success",
					"goal_id":     goal.ID,
					"snapshot_id": snapshotID,
					"name":        goal.Name,
				})
			case "yaml":
				return formatOutput("yaml", map[string]interface{}{
					"status":      "success",
					"goal_id":     goal.ID,
					"snapshot_id": snapshotID,
					"name":        goal.Name,
				})
			case "ai":
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

// CREATE JOURNEY command
func newCreateJourneyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-journey GOAL_ID SNAPSHOT_ID NAME",
		Short: "Create a new Virtuoso journey (testsuite)",
		Long: `Create a new journey (testsuite) in Virtuoso for the specified goal.

Example:
  api-cli create-journey 13776 43802 "My Test Journey"
  api-cli create-journey 13776 43802 "My Test Journey" -o json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return fmt.Errorf("requires exactly 3 arguments: GOAL_ID SNAPSHOT_ID NAME\n\nExample:\n  api-cli create-journey 13776 43802 \"My Test Journey\"")
			}
			if strings.TrimSpace(args[2]) == "" {
				return fmt.Errorf("journey name cannot be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			goalID, err := parseID(args[0], "goal")
			if err != nil {
				return fmt.Errorf("failed to parse goal ID '%s': %w", args[0], err)
			}
			snapshotID, err := parseID(args[1], "snapshot")
			if err != nil {
				return fmt.Errorf("failed to parse snapshot ID '%s': %w", args[1], err)
			}
			journeyName := strings.TrimSpace(args[2])

			apiClient := client.NewClient(cfg)

			journey, err := apiClient.CreateJourney(goalID, snapshotID, journeyName)
			if err != nil {
				return handleAPIError(err, "create journey")
			}

			// Format output
			switch cfg.Output.DefaultFormat {
			case "json":
				return formatOutput("json", map[string]interface{}{
					"status":     "success",
					"journey_id": journey.ID,
					"name":       journey.Name,
					"goal_id":    goalID,
				})
			case "yaml":
				return formatOutput("yaml", map[string]interface{}{
					"status":     "success",
					"journey_id": journey.ID,
					"name":       journey.Name,
					"goal_id":    goalID,
				})
			case "ai":
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

// CREATE CHECKPOINT command
func newCreateCheckpointCmd() *cobra.Command {
	var position int

	cmd := &cobra.Command{
		Use:   "create-checkpoint JOURNEY_ID GOAL_ID SNAPSHOT_ID NAME",
		Short: "Create and attach a checkpoint to a journey",
		Long: `Create a new checkpoint (testcase) and automatically attach it to the specified journey.
This command performs both operations in sequence to ensure the checkpoint is properly configured.

Example:
  api-cli create-checkpoint 608038 13776 43802 "Login Test"
  api-cli create-checkpoint 608038 13776 43802 "Checkout Test" --position 3
  api-cli create-checkpoint 608038 13776 43802 "Payment Test" -o json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 4 {
				return fmt.Errorf("requires exactly 4 arguments: JOURNEY_ID GOAL_ID SNAPSHOT_ID NAME\n\nExample:\n  api-cli create-checkpoint 608038 13776 43802 \"Login Test\"")
			}
			if strings.TrimSpace(args[3]) == "" {
				return fmt.Errorf("checkpoint name cannot be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			journeyID, err := parseID(args[0], "journey")
			if err != nil {
				return fmt.Errorf("failed to parse journey ID '%s': %w", args[0], err)
			}
			goalID, err := parseID(args[1], "goal")
			if err != nil {
				return fmt.Errorf("failed to parse goal ID '%s': %w", args[1], err)
			}
			snapshotID, err := parseID(args[2], "snapshot")
			if err != nil {
				return fmt.Errorf("failed to parse snapshot ID '%s': %w", args[2], err)
			}
			checkpointName := args[3]

			apiClient := client.NewClient(cfg)

			checkpoint, err := apiClient.CreateCheckpoint(goalID, snapshotID, checkpointName)
			if err != nil {
				return handleAPIError(err, "create checkpoint")
			}

			// Attach the checkpoint to the journey
			err = apiClient.AttachCheckpoint(journeyID, checkpoint.ID, position)
			if err != nil {
				// Checkpoint was created but not attached
				switch cfg.Output.DefaultFormat {
				case "json":
					output := map[string]interface{}{
						"status":        "partial",
						"checkpoint_id": checkpoint.ID,
						"name":          checkpoint.Title,
						"error":         fmt.Sprintf("Checkpoint created but not attached: %v", err),
					}
					formatOutput("json", output)
				default:
					fmt.Printf("⚠️  Created checkpoint '%s' with ID: %d, but failed to attach to journey: %v\n",
						checkpoint.Title, checkpoint.ID, err)
				}
				return err
			}

			// Format output
			switch cfg.Output.DefaultFormat {
			case "json":
				return formatOutput("json", map[string]interface{}{
					"status":        "success",
					"checkpoint_id": checkpoint.ID,
					"journey_id":    journeyID,
					"position":      position,
					"name":          checkpoint.Title,
				})
			case "yaml":
				return formatOutput("yaml", map[string]interface{}{
					"status":        "success",
					"checkpoint_id": checkpoint.ID,
					"journey_id":    journeyID,
					"position":      position,
					"name":          checkpoint.Title,
				})
			case "ai":
				fmt.Printf("Successfully created and attached checkpoint:\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpoint.ID)
				fmt.Printf("- Checkpoint Name: %s\n", checkpoint.Title)
				fmt.Printf("- Journey ID: %d\n", journeyID)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add steps to checkpoint: api-cli add-step %d <step-type> <step-details>\n", checkpoint.ID)
				fmt.Printf("2. Create another checkpoint: api-cli create-checkpoint %d %d %d \"Next Test\"\n",
					journeyID, goalID, snapshotID)
			default: // human
				fmt.Printf("✅ Created and attached checkpoint '%s' with ID: %d to journey %d at position %d\n",
					checkpoint.Title, checkpoint.ID, journeyID, position)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&position, "position", 2, "Position in journey (must be 2 or greater)")
	return cmd
}

// UPDATE JOURNEY command
func newUpdateJourneyCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update-journey JOURNEY_ID",
		Short: "Update a Virtuoso journey (testsuite) name",
		Long: `Update the name of an existing journey (testsuite) in Virtuoso.

Example:
  api-cli update-journey 12345 --name "Updated Journey Name"
  api-cli update-journey 12345 --name "New Name" -o json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly 1 argument: JOURNEY_ID\n\nExample:\n  api-cli update-journey 12345 --name \"Updated Journey Name\"")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate name is provided and not empty
			if name == "" {
				return fmt.Errorf("--name flag is required")
			}
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("journey name cannot be empty")
			}

			journeyID, err := parseID(args[0], "journey")
			if err != nil {
				return fmt.Errorf("failed to parse journey ID '%s': %w", args[0], err)
			}

			apiClient := client.NewClient(cfg)

			// Get current journey details for human output
			var originalJourney *client.Journey
			if cfg.Output.DefaultFormat == "human" || cfg.Output.DefaultFormat == "" {
				originalJourney, err = apiClient.GetJourney(journeyID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Could not fetch current journey details: %v\n", err)
				}
			}

			journey, err := apiClient.UpdateJourney(journeyID, name)
			if err != nil {
				return handleAPIError(err, "update journey")
			}

			// Format output
			switch cfg.Output.DefaultFormat {
			case "json":
				return formatOutput("json", map[string]interface{}{
					"status":     "success",
					"journey_id": journey.ID,
					"name":       journey.Name,
					"goal_id":    journey.GoalID,
				})
			case "yaml":
				return formatOutput("yaml", map[string]interface{}{
					"status":     "success",
					"journey_id": journey.ID,
					"name":       journey.Name,
					"goal_id":    journey.GoalID,
				})
			case "ai":
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

	cmd.Flags().StringVar(&name, "name", "", "New name for the journey (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

// UPDATE NAVIGATION command
func newUpdateNavigationCmd() *cobra.Command {
	var urlFlag string
	var newTab bool

	cmd := &cobra.Command{
		Use:   "update-navigation STEP_ID CANONICAL_ID",
		Short: "Update a navigation step URL in Virtuoso",
		Long: `Update the URL of an existing navigation step in Virtuoso.

This command requires both the step ID and canonical ID (obtained from get-step command).
The canonical ID ensures you're updating the correct version of the step.

Example:
  # First get the step details
  api-cli get-step 12345

  # Then update the navigation URL
  api-cli update-navigation 12345 "abc-def-123" --url "https://example.com"
  api-cli update-navigation 12345 "abc-def-123" --url "https://example.com" --new-tab`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("requires exactly 2 arguments: STEP_ID CANONICAL_ID\n\nExample:\n  api-cli update-navigation 12345 \"abc-def-123\" --url \"https://example.com\"")
			}
			if strings.TrimSpace(args[1]) == "" {
				return fmt.Errorf("canonical ID cannot be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if urlFlag == "" {
				return fmt.Errorf("--url flag is required")
			}

			// Validate URL format
			if err := ValidateURL(urlFlag); err != nil {
				return fmt.Errorf("invalid URL: %w\n\nURL must start with http:// or https://", err)
			}

			stepID, err := parseID(args[0], "step")
			if err != nil {
				return err
			}
			canonicalID := args[1]

			apiClient := client.NewClient(cfg)

			// Get current step details for comparison (optional, for human output)
			var originalStep *client.Step
			if cfg.Output.DefaultFormat == "human" || cfg.Output.DefaultFormat == "" {
				originalStep, _ = apiClient.GetStep(stepID)
			}

			step, err := apiClient.UpdateNavigationStep(stepID, canonicalID, urlFlag, newTab)
			if err != nil {
				return handleAPIError(err, "update navigation step")
			}

			// Format output
			switch cfg.Output.DefaultFormat {
			case "json":
				return formatOutput("json", map[string]interface{}{
					"status":        "success",
					"step_id":       step.ID,
					"canonical_id":  step.CanonicalID,
					"action":        step.Action,
					"url":           step.Value,
					"new_tab":       newTab,
					"checkpoint_id": step.CheckpointID,
				})
			case "yaml":
				return formatOutput("yaml", map[string]interface{}{
					"status":        "success",
					"step_id":       step.ID,
					"canonical_id":  step.CanonicalID,
					"action":        step.Action,
					"url":           step.Value,
					"new_tab":       newTab,
					"checkpoint_id": step.CheckpointID,
				})
			case "ai":
				fmt.Printf("Successfully updated navigation step:\n")
				fmt.Printf("- Step ID: %d\n", step.ID)
				fmt.Printf("- Canonical ID: %s\n", step.CanonicalID)
				fmt.Printf("- New URL: %s\n", step.Value)
				if originalStep != nil && originalStep.Value != step.Value {
					fmt.Printf("- Previous URL: %s\n", originalStep.Value)
				}
				fmt.Printf("- Open in New Tab: %t\n", newTab)
				fmt.Printf("- Checkpoint ID: %d\n", step.CheckpointID)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Run the test: api-cli run-journey <journey-id>\n")
				fmt.Printf("2. Get step details: api-cli get-step %d\n", step.ID)
			default: // human
				fmt.Printf("✅ Updated navigation step %d\n", step.ID)
				if originalStep != nil && originalStep.Value != step.Value {
					fmt.Printf("   From: %s\n", originalStep.Value)
					fmt.Printf("   To:   %s\n", step.Value)
				} else {
					fmt.Printf("   URL: %s\n", step.Value)
				}
				if newTab {
					fmt.Printf("   Opens in: New tab\n")
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&urlFlag, "url", "", "New URL for the navigation step (required)")
	cmd.Flags().BoolVar(&newTab, "new-tab", false, "Open URL in a new tab")
	cmd.MarkFlagRequired("url")
	return cmd
}

// GET STEP command
func newGetStepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-step STEP_ID",
		Short: "Get details of a Virtuoso test step",
		Long: `Retrieve detailed information about a test step in Virtuoso.

This command is particularly useful for getting the canonicalId, which is required
for updating steps.

Example:
  api-cli get-step 12345
  api-cli get-step 12345 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stepID, err := parseID(args[0], "step")
			if err != nil {
				return fmt.Errorf("failed to parse step ID '%s': %w", args[0], err)
			}

			apiClient := client.NewClient(cfg)

			step, err := apiClient.GetStep(stepID)
			if err != nil {
				return handleAPIError(err, "get step")
			}

			// Format output
			switch cfg.Output.DefaultFormat {
			case "json":
				return formatOutput("json", map[string]interface{}{
					"id":             step.ID,
					"canonical_id":   step.CanonicalID,
					"checkpoint_id":  step.CheckpointID,
					"step_index":     step.StepIndex,
					"action":         step.Action,
					"value":          step.Value,
					"optional":       step.Optional,
					"ignore_outcome": step.IgnoreOutcome,
					"skip":           step.Skip,
					"meta":           step.Meta,
					"target":         step.Target,
				})
			case "yaml":
				fmt.Printf("id: %d\n", step.ID)
				fmt.Printf("canonical_id: %s\n", step.CanonicalID)
				fmt.Printf("checkpoint_id: %d\n", step.CheckpointID)
				fmt.Printf("step_index: %d\n", step.StepIndex)
				fmt.Printf("action: %s\n", step.Action)
				fmt.Printf("value: %s\n", step.Value)
				fmt.Printf("optional: %t\n", step.Optional)
				fmt.Printf("ignore_outcome: %t\n", step.IgnoreOutcome)
				fmt.Printf("skip: %t\n", step.Skip)
				if step.Meta != nil && len(step.Meta) > 0 {
					fmt.Printf("meta:\n")
					for k, v := range step.Meta {
						fmt.Printf("  %s: %v\n", k, v)
					}
				}
			case "ai":
				fmt.Printf("Virtuoso Test Step Details:\n")
				fmt.Printf("- Step ID: %d\n", step.ID)
				fmt.Printf("- Canonical ID: %s (required for updates)\n", step.CanonicalID)
				fmt.Printf("- Checkpoint ID: %d\n", step.CheckpointID)
				fmt.Printf("- Step Index: %d\n", step.StepIndex)
				fmt.Printf("- Action: %s\n", step.Action)
				if step.Value != "" {
					fmt.Printf("- Value: %s\n", step.Value)
				}
				fmt.Printf("\nStep Flags:\n")
				fmt.Printf("- Optional: %t\n", step.Optional)
				fmt.Printf("- Ignore Outcome: %t\n", step.IgnoreOutcome)
				fmt.Printf("- Skip: %t\n", step.Skip)
				if step.Meta != nil && len(step.Meta) > 0 {
					fmt.Printf("\nMeta Properties:\n")
					for k, v := range step.Meta {
						fmt.Printf("- %s: %v\n", k, v)
					}
				}
				fmt.Printf("\nUsage:\n")
				fmt.Printf("To update this step, use: api-cli update-step %s\n", step.CanonicalID)
			default: // human
				fmt.Printf("Step ID: %d\n", step.ID)
				fmt.Printf("Canonical ID: %s ← Use this for updates\n", step.CanonicalID)
				fmt.Printf("Action: %s\n", step.Action)
				if step.Value != "" {
					fmt.Printf("Value: %s\n", step.Value)
				}
				fmt.Printf("Checkpoint: %d (Index: %d)\n", step.CheckpointID, step.StepIndex)
				if step.Optional || step.IgnoreOutcome || step.Skip {
					fmt.Printf("Flags:")
					if step.Optional {
						fmt.Printf(" [Optional]")
					}
					if step.IgnoreOutcome {
						fmt.Printf(" [IgnoreOutcome]")
					}
					if step.Skip {
						fmt.Printf(" [Skip]")
					}
					fmt.Printf("\n")
				}
				if step.Meta != nil && len(step.Meta) > 0 {
					fmt.Printf("Meta:\n")
					for k, v := range step.Meta {
						fmt.Printf("  %s: %v\n", k, v)
					}
				}
			}
			return nil
		},
	}

	return cmd
}
