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

// newCreateStepCookieWipeAllCmd creates the command for clearing all cookies
func newCreateStepCookieWipeAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-cookie-wipe-all CHECKPOINT_ID POSITION",
		Short: "Clear all cookies",
		Long: `Creates a step that clears all cookies in the browser.
This corresponds to the ENVIRONMENT action with CLEAR type.

Example:
  api-cli create-step-cookie-wipe-all 1678318 1`,
		Args: cobra.ExactArgs(2),
		RunE: runCreateStepCookieWipeAll,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepCookieWipeAll(cmd *cobra.Command, args []string) error {
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

	// Create the cookie wipe step
	stepID, err := client.CreateStepCookieWipeAll(checkpointID, position)
	if err != nil {
		return fmt.Errorf("failed to create cookie wipe step: %w", err)
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
		fmt.Printf("Created cookie wipe step with ID %d for checkpoint %d at position %d\n",
			stepID, checkpointID, position)
	default: // human
		fmt.Printf("Cookie wipe step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Position: %d\n", position)
		fmt.Printf("Effect: All cookies will be cleared\n")
	}

	return nil
}
