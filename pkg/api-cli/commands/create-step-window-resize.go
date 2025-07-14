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

// newCreateStepWindowResizeCmd creates the command for creating a window resize step
func newCreateStepWindowResizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-window-resize CHECKPOINT_ID WIDTH HEIGHT POSITION",
		Short: "Resize the browser window",
		Long: `Creates a window resize step that resizes the browser window to the specified dimensions.

Examples:
  api-cli create-step-window-resize 1678318 1024 768 1
  api-cli create-step-window-resize 1678318 1920 1080 1 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepWindowResize,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

func runCreateStepWindowResize(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	width, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid width: %w", err)
	}

	height, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid height: %w", err)
	}

	position, err := strconv.Atoi(args[3])
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

	// Create the window resize step
	stepID, err := client.CreateStepWindowResize(checkpointID, width, height, position)
	if err != nil {
		return fmt.Errorf("failed to create window resize step: %w", err)
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
		fmt.Printf("Created window resize step with ID %d for checkpoint %d. Size: %dx%d, position: %d\n",
			stepID, checkpointID, width, height, position)
	default: // human
		fmt.Printf("Window resize step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Width: %d\n", width)
		fmt.Printf("Height: %d\n", height)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
