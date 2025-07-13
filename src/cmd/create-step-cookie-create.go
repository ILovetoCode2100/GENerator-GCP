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

// newCreateStepCookieCreateCmd creates the command for creating a cookie step
func newCreateStepCookieCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-cookie-create CHECKPOINT_ID NAME VALUE POSITION",
		Short: "Create a cookie with specified name and value",
		Long: `Creates a cookie step that adds a new cookie with the specified name and value.
This corresponds to the ENVIRONMENT action with ADD type.

Example:
  api-cli create-step-cookie-create 1678318 "session" "abc123" 1`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepCookieCreate,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepCookieCreate(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	name := args[1]
	value := args[2]

	position, err := strconv.Atoi(args[3])
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

	// Create the cookie step
	stepID, err := client.CreateStepCookieCreate(checkpointID, name, value, position)
	if err != nil {
		return fmt.Errorf("failed to create cookie step: %w", err)
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
		fmt.Printf("Created cookie step with ID %d for checkpoint %d. Cookie name: %s, value: %s, position: %d\n",
			stepID, checkpointID, name, value, position)
	default: // human
		fmt.Printf("Cookie step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Cookie Name: %s\n", name)
		fmt.Printf("Cookie Value: %s\n", value)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
