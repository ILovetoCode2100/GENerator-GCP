package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
)

// newCreateStepSwitchIframeCmd creates the command for switching to iframe by element selector
func newCreateStepSwitchIframeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-switch-iframe CHECKPOINT_ID SELECTOR POSITION",
		Short: "Switch to iframe by element selector",
		Long: `Creates a step that switches to an iframe identified by an element selector.
This corresponds to the SWITCH action with FRAME_BY_ELEMENT type.

Example:
  api-cli create-step-switch-iframe 1678318 "#myframe" 1`,
		Args: cobra.ExactArgs(3),
		RunE: runCreateStepSwitchIframe,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepSwitchIframe(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	selector := args[1]

	position, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	// Get API token from environment
	token := os.Getenv("VIRTUOSO_API_TOKEN")
	if token == "" {
		return fmt.Errorf("VIRTUOSO_API_TOKEN environment variable is required")
	}

	// Get API base URL from environment
	baseURL := os.Getenv("VIRTUOSO_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api-app2.virtuoso.qa/api"
	}

	// Create client
	client := virtuoso.NewClientDirect(baseURL, token)

	// Create the switch iframe step
	stepID, err := client.CreateStepSwitchIframe(checkpointID, selector, position)
	if err != nil {
		return fmt.Errorf("failed to create switch iframe step: %w", err)
	}

	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output")

	// Format output
	switch outputFormat {
	case "json":
		output, err := json.MarshalIndent(map[string]interface{}{"stepId": stepID, "checkpointId": checkpointID}, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
	case "yaml":
		output, err := yaml.Marshal(map[string]interface{}{"stepId": stepID, "checkpointId": checkpointID})
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Print(string(output))
	case "ai":
		fmt.Printf("Created switch iframe step with ID %d for checkpoint %d. Selector: %s, position: %d\n",
			stepID, checkpointID, selector, position)
	default: // human
		fmt.Printf("Switch iframe step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Iframe Selector: %s\n", selector)
		fmt.Printf("Position: %d\n", position)
		fmt.Printf("Effect: Browser will switch to iframe matching the selector\n")
	}

	return nil
}
