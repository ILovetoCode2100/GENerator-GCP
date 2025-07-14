package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

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
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			stepIDStr := args[0]
			canonicalID := args[1]

			// Validate that URL flag was provided
			if urlFlag == "" {
				return fmt.Errorf("--url flag is required")
			}

			// Convert step ID to int
			stepID, err := strconv.Atoi(stepIDStr)
			if err != nil {
				return fmt.Errorf("invalid step ID: %w", err)
			}

			// Validate canonical ID
			if canonicalID == "" {
				return fmt.Errorf("canonical ID cannot be empty")
			}

			// Validate URL format
			parsedURL, err := url.Parse(urlFlag)
			if err != nil {
				return fmt.Errorf("invalid URL format: %w", err)
			}
			if parsedURL.Scheme == "" || parsedURL.Host == "" {
				return fmt.Errorf("URL must include scheme (http/https) and host")
			}

			// Create Virtuoso client
			apiClient := client.NewClient(cfg)

			// Get current step details for comparison (optional, for human output)
			var originalStep *client.Step
			if cfg.Output.DefaultFormat == "human" || cfg.Output.DefaultFormat == "" {
				originalStep, _ = apiClient.GetStep(stepID)
				// Non-fatal if we can't get the original
			}

			// Update the navigation step
			step, err := apiClient.UpdateNavigationStep(stepID, canonicalID, urlFlag, newTab)
			if err != nil {
				return fmt.Errorf("failed to update navigation step: %w", err)
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_id":       step.ID,
					"canonical_id":  step.CanonicalID,
					"action":        step.Action,
					"url":           step.Value,
					"new_tab":       newTab,
					"checkpoint_id": step.CheckpointID,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				// Simple YAML output
				fmt.Printf("status: success\n")
				fmt.Printf("step_id: %d\n", step.ID)
				fmt.Printf("canonical_id: %s\n", step.CanonicalID)
				fmt.Printf("action: %s\n", step.Action)
				fmt.Printf("url: %s\n", step.Value)
				fmt.Printf("new_tab: %t\n", newTab)
				fmt.Printf("checkpoint_id: %d\n", step.CheckpointID)
			case "ai":
				// AI-friendly output
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

	// Add flags
	cmd.Flags().StringVar(&urlFlag, "url", "", "New URL for the navigation step (required)")
	cmd.Flags().BoolVar(&newTab, "new-tab", false, "Open URL in a new tab")
	cmd.MarkFlagRequired("url")

	return cmd
}
