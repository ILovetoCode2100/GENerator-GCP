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

// newCreateStepAssertNotEqualsCmd creates the command for creating an assert not equals step
func newCreateStepAssertNotEqualsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-assert-not-equals CHECKPOINT_ID SELECTOR VALUE POSITION",
		Short: "Create an assertion step that checks element does not equal value",
		Long: `Creates an assertion step that verifies the selected element does not equal the specified value.
This corresponds to the ASSERT_NOT_EQUALS action.

Example:
  api-cli create-step-assert-not-equals 1678318 "Status" "Failed" 1`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepAssertNotEquals,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepAssertNotEquals(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	selector := args[1]
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

	// Create the assertion step
	stepID, err := client.CreateStepAssertNotEquals(checkpointID, selector, value, position)
	if err != nil {
		return fmt.Errorf("failed to create assert not equals step: %w", err)
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
		fmt.Printf("Created assert not equals step with ID %d for checkpoint %d. Element '%s' should not equal '%s', position: %d\n",
			stepID, checkpointID, selector, value, position)
	default: // human
		fmt.Printf("Assert not equals step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Selector: %s\n", selector)
		fmt.Printf("Expected not equal to: %s\n", value)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
