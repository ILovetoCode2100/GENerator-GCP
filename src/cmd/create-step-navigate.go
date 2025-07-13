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

// newCreateStepNavigateCmd creates the command for creating a navigation step
func newCreateStepNavigateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-navigate CHECKPOINT_ID URL POSITION [flags]",
		Short: "Navigate to a URL",
		Long: `Creates a navigation step that navigates to the specified URL.

Examples:
  api-cli create-step-navigate 1678318 "https://example.com" 1
  api-cli create-step-navigate 1678318 "https://example.com" 1 --new-tab
  api-cli create-step-navigate 1678318 "https://example.com" 1 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: runCreateStepNavigate,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	cmd.Flags().Bool("new-tab", false, "Open URL in new tab")
	
	return cmd
}

func runCreateStepNavigate(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	url := args[1]
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	position, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	// Get flags
	useNewTab, _ := cmd.Flags().GetBool("new-tab")

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

	// Create the navigation step
	stepID, err := client.CreateStepNavigate(checkpointID, url, useNewTab, position)
	if err != nil {
		return fmt.Errorf("failed to create navigation step: %w", err)
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
		tabInfo := ""
		if useNewTab {
			tabInfo = " in new tab"
		}
		fmt.Printf("Created navigation step with ID %d for checkpoint %d. URL: %s%s, position: %d\n", 
			stepID, checkpointID, url, tabInfo, position)
	default: // human
		fmt.Printf("Navigation step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("URL: %s\n", url)
		if useNewTab {
			fmt.Printf("New Tab: Yes\n")
		}
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
