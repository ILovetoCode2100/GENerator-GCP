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

// newCreateStepStoreElementTextCmd creates the command for storing element text in variable
func newCreateStepStoreElementTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-store-element-text CHECKPOINT_ID SELECTOR VARIABLE_NAME POSITION",
		Short: "Store element text in variable",
		Long: `Creates a step that stores the text content of an element in a variable.
The variable can be used in subsequent steps.

Example:
  api-cli create-step-store-element-text 1678318 "Username field" "current_user" 1`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepStoreElementText,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepStoreElementText(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	selector := args[1]
	variableName := args[2]

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

	// Create the store element text step
	stepID, err := client.CreateStepStoreElementText(checkpointID, selector, variableName, position)
	if err != nil {
		return fmt.Errorf("failed to create store element text step: %w", err)
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
		fmt.Printf("Created store element text step with ID %d for checkpoint %d. Selector: %s, variable: %s, position: %d\n", 
			stepID, checkpointID, selector, variableName, position)
	default: // human
		fmt.Printf("Store element text step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Selector: %s\n", selector)
		fmt.Printf("Variable: %s\n", variableName)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
