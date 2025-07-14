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

// newCreateStepCommentCmd creates the command for creating a comment step
func newCreateStepCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-comment CHECKPOINT_ID COMMENT POSITION",
		Short: "Add a comment to the test",
		Long: `Creates a comment step that adds a comment to the test for documentation purposes.

Examples:
  api-cli create-step-comment 1678318 "This is a comment" 1
  api-cli create-step-comment 1678318 "TODO: Add login validation" 1
  api-cli create-step-comment 1678318 "FIXME: Check password error handling" 1 -o json

Comments are useful for:
  - Documenting test steps
  - Adding TODO items
  - Marking FIXME areas
  - Providing context for complex operations`,
		Args: cobra.ExactArgs(3),
		RunE: runCreateStepComment,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepComment(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	comment := args[1]
	if comment == "" {
		return fmt.Errorf("comment cannot be empty")
	}

	position, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

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
	client := client.NewClientDirect(baseURL, token)

	// Create the comment step
	stepID, err := client.CreateStepComment(checkpointID, comment, position)
	if err != nil {
		return fmt.Errorf("failed to create comment step: %w", err)
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
		fmt.Printf("Created comment step with ID %d for checkpoint %d. Comment: %s, position: %d\n",
			stepID, checkpointID, comment, position)
	default: // human
		fmt.Printf("Comment step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Comment: %s\n", comment)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
