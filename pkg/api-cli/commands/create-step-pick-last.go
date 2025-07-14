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

// newCreateStepPickLastCmd creates the command for picking last dropdown option
func newCreateStepPickLastCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-pick-last CHECKPOINT_ID SELECTOR POSITION",
		Short: "Pick the last dropdown option",
		Long: `Creates a step that picks the last option in a dropdown.
This is equivalent to picking by index -1.

Example:
  api-cli create-step-pick-last 1678318 "Country dropdown" 1`,
		Args: cobra.ExactArgs(3),
		RunE: runCreateStepPickLast,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepPickLast(cmd *cobra.Command, args []string) error {
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
	client := client.NewClientDirect(baseURL, token)

	// Create the pick last step
	stepID, err := client.CreateStepPickLast(checkpointID, selector, position)
	if err != nil {
		return fmt.Errorf("failed to create pick last step: %w", err)
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
		fmt.Printf("Created pick last step with ID %d for checkpoint %d. Selector: %s, position: %d\n",
			stepID, checkpointID, selector, position)
	default: // human
		fmt.Printf("Pick last step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Selector: %s\n", selector)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
