package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

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
			stepIDStr := args[0]

			// Convert step ID to int
			stepID, err := strconv.Atoi(stepIDStr)
			if err != nil {
				return fmt.Errorf("invalid step ID: %w", err)
			}

			// Create Virtuoso client
			client := client.NewClient(cfg)

			// Get the step
			step, err := client.GetStep(stepID)
			if err != nil {
				return fmt.Errorf("failed to get step: %w", err)
			}

			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
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
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				// Simple YAML output
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
				// AI-friendly output
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
