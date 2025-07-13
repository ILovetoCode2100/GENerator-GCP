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

// newCreateStepStoreLiteralValueCmd creates the command for storing literal value in variable
func newCreateStepStoreLiteralValueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-store-literal-value CHECKPOINT_ID VALUE VARIABLE_NAME POSITION",
		Short: "Store literal value in variable",
		Long: `Creates a step that stores a literal value in a variable.
The variable can be used in subsequent steps.

Example:
  api-cli create-step-store-literal-value 1678318 "Hello World" "greeting" 1`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepStoreLiteralValue,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepStoreLiteralValue(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	value := args[1]
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

	// Create the store literal value step
	stepID, err := client.CreateStepStoreLiteralValue(checkpointID, value, variableName, position)
	if err != nil {
		return fmt.Errorf("failed to create store literal value step: %w", err)
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
		fmt.Printf("Created store literal value step with ID %d for checkpoint %d. Value: %s, variable: %s, position: %d\n", 
			stepID, checkpointID, value, variableName, position)
	default: // human
		fmt.Printf("Store literal value step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Value: %s\n", value)
		fmt.Printf("Variable: %s\n", variableName)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
