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

// newCreateStepPickIndexCmd creates the command for picking dropdown option by index
func newCreateStepPickIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-pick-index CHECKPOINT_ID SELECTOR INDEX POSITION",
		Short: "Pick dropdown option by index",
		Long: `Creates a step that picks a dropdown option by its index position.
Index is 0-based, so first option is 0, second is 1, etc.

Example:
  api-cli create-step-pick-index 1678318 "Country dropdown" 2 1`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepPickIndex,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepPickIndex(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	selector := args[1]
	
	index, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid index: %w", err)
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

	// Create the pick index step
	stepID, err := client.CreateStepPickIndex(checkpointID, selector, index, position)
	if err != nil {
		return fmt.Errorf("failed to create pick index step: %w", err)
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
		fmt.Printf("Created pick index step with ID %d for checkpoint %d. Selector: %s, index: %d, position: %d\n", 
			stepID, checkpointID, selector, index, position)
	default: // human
		fmt.Printf("Pick index step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Selector: %s\n", selector)
		fmt.Printf("Index: %d\n", index)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
