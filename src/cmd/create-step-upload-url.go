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

// newCreateStepUploadURLCmd creates the create-step-upload-url command
func newCreateStepUploadURLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-upload-url CHECKPOINT_ID URL SELECTOR POSITION",
		Short: "Create a step to upload a file from URL",
		Long: `Create a step to upload a file from URL to a specified element.

The command creates a step that uploads a file from the provided URL to the element
identified by the selector. The selector is used as a clue for the GUESS target type.

Arguments:
  CHECKPOINT_ID  The ID of the checkpoint to add the step to
  URL           The URL of the file to upload
  SELECTOR      The selector clue for the upload element (e.g., "Résumé:")
  POSITION      The position of the step in the checkpoint

Examples:
  # Upload a PDF from URL to a resume upload field
  api-cli create-step-upload-url 1234 https://example.com/resume.pdf "Résumé:" 1

  # Upload with JSON output
  api-cli create-step-upload-url 1234 https://example.com/doc.pdf "Upload Document" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepUploadURL,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepUploadURL(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	url := args[1]
	selector := args[2]

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

	// Create the step
	stepID, err := client.CreateStepUploadURL(checkpointID, url, selector, position)
	if err != nil {
		return fmt.Errorf("failed to create upload URL step: %w", err)
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
		fmt.Printf("Created upload URL step with ID %d for checkpoint %d. URL: %s, selector: %s, position: %d\n",
			stepID, checkpointID, url, selector, position)
	default: // human
		fmt.Printf("Upload URL step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("File URL: %s\n", url)
		fmt.Printf("Upload Selector: %s\n", selector)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
