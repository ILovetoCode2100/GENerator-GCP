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

// newCreateStepWaitForElementDefaultCmd creates the command for waiting for element with default timeout
func newCreateStepWaitForElementDefaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-wait-for-element-default CHECKPOINT_ID SELECTOR POSITION",
		Short: "Wait for element with default timeout",
		Long: `Creates a step that waits for an element to appear with the default timeout of 20 seconds.
This is equivalent to waiting with a timeout of 20000ms.

Example:
  api-cli create-step-wait-for-element-default 1678318 "Loading complete" 1`,
		Args: cobra.ExactArgs(3),
		RunE: runCreateStepWaitForElementDefault,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepWaitForElementDefault(cmd *cobra.Command, args []string) error {
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

	// Create the wait for element default step
	stepID, err := client.CreateStepWaitForElementDefault(checkpointID, selector, position)
	if err != nil {
		return fmt.Errorf("failed to create wait for element default step: %w", err)
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
		fmt.Printf("Created wait for element default step with ID %d for checkpoint %d. Selector: %s, timeout: 20000ms, position: %d\n", 
			stepID, checkpointID, selector, position)
	default: // human
		fmt.Printf("Wait for element default step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Selector: %s\n", selector)
		fmt.Printf("Timeout: 20000ms (default)\n")
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
