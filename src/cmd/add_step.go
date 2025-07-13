package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newAddStepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-step",
		Short: "Add test steps to a checkpoint",
		Long: `Add test steps to a checkpoint. Use subcommands for different step types.

Available step types:
  navigate - Navigate to a URL
  click    - Click on an element
  wait     - Wait for an element`,
	}

	// Add subcommands
	cmd.AddCommand(newAddNavigateStepCmd())
	cmd.AddCommand(newAddClickStepCmd())
	cmd.AddCommand(newAddWaitStepCmd())

	return cmd
}

// Navigate step command
func newAddNavigateStepCmd() *cobra.Command {
	var url string

	cmd := &cobra.Command{
		Use:   "navigate CHECKPOINT_ID",
		Short: "Add a navigate step to a checkpoint",
		Long: `Add a navigation step to go to a specific URL.

Example:
  api-cli add-step navigate 1678318 --url "https://example.com"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]

			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}

			client := virtuoso.NewClient(cfg)

			stepID, err := client.AddNavigateStep(checkpointID, url)
			if err != nil {
				return fmt.Errorf("failed to add navigate step: %w", err)
			}

			formatStepOutput("NAVIGATE", checkpointID, stepID)
			return nil
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "URL to navigate to (required)")
	cmd.MarkFlagRequired("url")

	return cmd
}

// Click step command
func newAddClickStepCmd() *cobra.Command {
	var selector string

	cmd := &cobra.Command{
		Use:   "click CHECKPOINT_ID",
		Short: "Add a click step to a checkpoint",
		Long: `Add a click step to click on an element.

Example:
  api-cli add-step click 1678318 --selector "Submit Button"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]

			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}

			client := virtuoso.NewClient(cfg)

			stepID, err := client.AddClickStep(checkpointID, selector)
			if err != nil {
				return fmt.Errorf("failed to add click step: %w", err)
			}

			formatStepOutput("CLICK", checkpointID, stepID)
			return nil
		},
	}

	cmd.Flags().StringVar(&selector, "selector", "", "Element selector/clue (required)")
	_ = cmd.MarkFlagRequired("selector")

	return cmd
}

// Wait step command
func newAddWaitStepCmd() *cobra.Command {
	var selector string
	var timeout int

	cmd := &cobra.Command{
		Use:   "wait CHECKPOINT_ID",
		Short: "Add a wait step to a checkpoint",
		Long: `Add a wait step to wait for an element to appear.

Example:
  api-cli add-step wait 1678318 --selector "Loading Complete" --timeout 5000`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]

			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}

			client := virtuoso.NewClient(cfg)

			stepID, err := client.AddWaitStep(checkpointID, selector, timeout)
			if err != nil {
				return fmt.Errorf("failed to add wait step: %w", err)
			}

			formatStepOutput("WAIT", checkpointID, stepID)
			return nil
		},
	}

	cmd.Flags().StringVar(&selector, "selector", "", "Element selector/clue (required)")
	cmd.Flags().IntVar(&timeout, "timeout", 20000, "Timeout in milliseconds")
	_ = cmd.MarkFlagRequired("selector")

	return cmd
}

// Helper function to format output
func formatStepOutput(stepType string, checkpointID, stepID int) {
	switch cfg.Output.DefaultFormat {
	case "json":
		output := map[string]interface{}{
			"status":        "success",
			"step_type":     stepType,
			"checkpoint_id": checkpointID,
			"step_id":       stepID,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(output)
	case "yaml":
		fmt.Printf("status: success\n")
		fmt.Printf("step_type: %s\n", stepType)
		fmt.Printf("checkpoint_id: %d\n", checkpointID)
		fmt.Printf("step_id: %d\n", stepID)
	case "ai":
		fmt.Printf("Successfully added %s step:\n", stepType)
		fmt.Printf("- Step ID: %d\n", stepID)
		fmt.Printf("- Step Type: %s\n", stepType)
		fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("1. Add another step: api-cli add-step <type> %d <options>\n", checkpointID)
		fmt.Printf("2. Execute the test journey\n")
	default: // human
		fmt.Printf("✅ Added %s step to checkpoint %d\n", stepType, checkpointID)
	}
}
