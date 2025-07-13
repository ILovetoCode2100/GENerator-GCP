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

// newCreateStepDismissPromptWithTextCmd creates the create-step-dismiss-prompt-with-text command
func newCreateStepDismissPromptWithTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-dismiss-prompt-with-text CHECKPOINT_ID TEXT POSITION",
		Short: "Create a step to dismiss a prompt with response text",
		Long: `Create a step to dismiss a browser prompt dialog with specified response text.

The command creates a step that dismisses a prompt dialog by clicking OK and 
providing the specified text as the response. This is useful for handling 
JavaScript prompt() dialogs that require user input.

Arguments:
  CHECKPOINT_ID  The ID of the checkpoint to add the step to
  TEXT          The text to enter in the prompt dialog
  POSITION      The position of the step in the checkpoint

Examples:
  # Dismiss prompt with user input
  api-cli create-step-dismiss-prompt-with-text 1234 "John Doe" 1
  
  # Dismiss prompt with JSON output
  api-cli create-step-dismiss-prompt-with-text 1234 "user@example.com" 2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: runCreateStepDismissPromptWithText,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepDismissPromptWithText(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	text := args[1]
	
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

	// Create the step
	stepID, err := client.CreateStepDismissPromptWithText(checkpointID, text, position)
	if err != nil {
		return fmt.Errorf("failed to create dismiss prompt step: %w", err)
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
		fmt.Printf("Created dismiss prompt step with ID %d for checkpoint %d. Response text: %s, position: %d\n", 
			stepID, checkpointID, text, position)
	default: // human
		fmt.Printf("Dismiss prompt step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Response Text: %s\n", text)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
