package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
)

// newCreateStepSwitchParentFrameCmd creates the command for switching to parent frame
func newCreateStepSwitchParentFrameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-switch-parent-frame CHECKPOINT_ID POSITION",
		Short: "Switch to parent frame",
		Long: `Creates a step that switches to the parent frame in the browser.
This corresponds to the SWITCH action with PARENT_FRAME type.

Example:
  api-cli create-step-switch-parent-frame 1678318 1`,
		Args: cobra.ExactArgs(2),
		RunE: runCreateStepSwitchParentFrame,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepSwitchParentFrame(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	position, err := strconv.Atoi(args[1])
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
	client := client.NewClientDirect(baseURL, token)

	// Create the switch parent frame step
	stepID, err := client.CreateStepSwitchParentFrame(checkpointID, position)
	if err != nil {
		return fmt.Errorf("failed to create switch parent frame step: %w", err)
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
		fmt.Printf("Created switch parent frame step with ID %d for checkpoint %d at position %d\n",
			stepID, checkpointID, position)
	default: // human
		fmt.Printf("Switch parent frame step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Position: %d\n", position)
		fmt.Printf("Effect: Browser will switch to the parent frame\n")
	}

	return nil
}
