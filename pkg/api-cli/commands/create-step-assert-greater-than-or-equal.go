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

// newCreateStepAssertGreaterThanOrEqualCmd creates the command for creating an assert greater than or equal step
func newCreateStepAssertGreaterThanOrEqualCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-assert-greater-than-or-equal CHECKPOINT_ID SELECTOR VALUE POSITION",
		Short: "Create an assertion step that checks element is greater than or equal to value",
		Long: `Creates an assertion step that verifies the selected element is greater than or equal to the specified value.
This corresponds to the ASSERT_GREATER_THAN_OR_EQUAL action.

Example:
  api-cli create-step-assert-greater-than-or-equal 1678318 "Score" "75" 1`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepAssertGreaterThanOrEqual,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepAssertGreaterThanOrEqual(cmd *cobra.Command, args []string) error {
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
	client := client.NewClientDirect(baseURL, token)

	// Create the assertion step
	stepID, err := client.CreateStepAssertGreaterThanOrEqual(checkpointID, selector, value, position)
	if err != nil {
		return fmt.Errorf("failed to create assert greater than or equal step: %w", err)
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
		fmt.Printf("Created assert greater than or equal step with ID %d for checkpoint %d. Element '%s' should be greater than or equal to '%s', position: %d\n",
			stepID, checkpointID, selector, value, position)
	default: // human
		fmt.Printf("Assert greater than or equal step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Selector: %s\n", selector)
		fmt.Printf("Expected greater than or equal to: %s\n", value)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
