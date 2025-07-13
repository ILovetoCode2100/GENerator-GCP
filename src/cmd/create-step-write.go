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

// newCreateStepWriteCmd creates the command for creating a write step
func newCreateStepWriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-write CHECKPOINT_ID SELECTOR VALUE POSITION [flags]",
		Short: "Write text to an input element",
		Long: `Creates a write step that inputs text into the specified element.

Examples:
  api-cli create-step-write 1678318 "First Name" "John" 1
  api-cli create-step-write 1678318 "Message" "hello world" 1 --variable "message"
  api-cli create-step-write 1678318 "Age" "24" 1 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepWrite,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	cmd.Flags().String("variable", "", "Store the value in a variable")

	return cmd
}

func runCreateStepWrite(cmd *cobra.Command, args []string) error {
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

	// Get flags
	variable, _ := cmd.Flags().GetString("variable")

	// Get API configuration
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

	// Create the write step
	var stepID int
	if variable != "" {
		stepID, err = client.CreateStepWriteWithVariable(checkpointID, selector, value, variable, position)
	} else {
		stepID, err = client.CreateStepWrite(checkpointID, selector, value, position)
	}

	if err != nil {
		return fmt.Errorf("failed to create write step: %w", err)
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
		varInfo := ""
		if variable != "" {
			varInfo = fmt.Sprintf(" (stored in variable '%s')", variable)
		}
		fmt.Printf("Created write step with ID %d for checkpoint %d. Selector: %s, value: %s%s, position: %d\n",
			stepID, checkpointID, selector, value, varInfo, position)
	default: // human
		fmt.Printf("Write step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Selector: %s\n", selector)
		fmt.Printf("Value: %s\n", value)
		if variable != "" {
			fmt.Printf("Variable: %s\n", variable)
		}
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
