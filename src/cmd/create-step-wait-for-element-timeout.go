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

// newCreateStepWaitForElementTimeoutCmd creates the command for waiting for element with custom timeout
func newCreateStepWaitForElementTimeoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-wait-for-element-timeout CHECKPOINT_ID SELECTOR TIMEOUT_MS POSITION",
		Short: "Wait for element with custom timeout",
		Long: `Creates a step that waits for an element to appear with a custom timeout.
The timeout is specified in milliseconds.

Example:
  api-cli create-step-wait-for-element-timeout 1678318 "Loading spinner" 5000 1`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepWaitForElementTimeout,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepWaitForElementTimeout(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	selector := args[1]
	
	timeoutMs, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

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

	// Create the wait for element timeout step
	stepID, err := client.CreateStepWaitForElementTimeout(checkpointID, selector, timeoutMs, position)
	if err != nil {
		return fmt.Errorf("failed to create wait for element timeout step: %w", err)
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
		fmt.Printf("Created wait for element timeout step with ID %d for checkpoint %d. Selector: %s, timeout: %dms, position: %d\n", 
			stepID, checkpointID, selector, timeoutMs, position)
	default: // human
		fmt.Printf("Wait for element timeout step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Selector: %s\n", selector)
		fmt.Printf("Timeout: %dms\n", timeoutMs)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
