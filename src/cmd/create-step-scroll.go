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

// newCreateStepScrollToPositionCmd creates the command for scrolling to position
func newCreateStepScrollToPositionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-scroll-to-position CHECKPOINT_ID X Y POSITION",
		Short: "Scroll to specific coordinates",
		Long: `Creates a scroll step that scrolls to the specified coordinates.

Examples:
  api-cli create-step-scroll-to-position 1678318 100 200 1
  api-cli create-step-scroll-to-position 1678318 0 500 1 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepScrollToPosition,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepScrollToPosition(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	x, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid x coordinate: %w", err)
	}

	y, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid y coordinate: %w", err)
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
	client := virtuoso.NewClientDirect(baseURL, token)

	// Create the scroll step
	stepID, err := client.CreateStepScrollToPosition(checkpointID, x, y, position)
	if err != nil {
		return fmt.Errorf("failed to create scroll to position step: %w", err)
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
		fmt.Printf("Created scroll to position step with ID %d for checkpoint %d. Coordinates: (%d, %d), position: %d\n", 
			stepID, checkpointID, x, y, position)
	default: // human
		fmt.Printf("Scroll to position step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Coordinates: (%d, %d)\n", x, y)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}

// newCreateStepScrollByOffsetCmd creates the command for scrolling by offset
func newCreateStepScrollByOffsetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-scroll-by-offset CHECKPOINT_ID X Y POSITION",
		Short: "Scroll by offset amount",
		Long: `Creates a scroll step that scrolls by the specified offset.

Examples:
  api-cli create-step-scroll-by-offset 1678318 0 500 1
  api-cli create-step-scroll-by-offset 1678318 100 -200 1 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: runCreateStepScrollByOffset,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepScrollByOffset(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	x, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid x offset: %w", err)
	}

	y, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid y offset: %w", err)
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
	client := virtuoso.NewClientDirect(baseURL, token)

	// Create the scroll step
	stepID, err := client.CreateStepScrollByOffset(checkpointID, x, y, position)
	if err != nil {
		return fmt.Errorf("failed to create scroll by offset step: %w", err)
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
		fmt.Printf("Created scroll by offset step with ID %d for checkpoint %d. Offset: (%d, %d), position: %d\n", 
			stepID, checkpointID, x, y, position)
	default: // human
		fmt.Printf("Scroll by offset step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Offset: (%d, %d)\n", x, y)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}

// newCreateStepScrollToTopCmd creates the command for scrolling to top
func newCreateStepScrollToTopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-scroll-to-top CHECKPOINT_ID POSITION",
		Short: "Scroll to top of page",
		Long: `Creates a scroll step that scrolls to the top of the page.

Examples:
  api-cli create-step-scroll-to-top 1678318 1
  api-cli create-step-scroll-to-top 1678318 1 -o json`,
		Args: cobra.ExactArgs(2),
		RunE: runCreateStepScrollToTop,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	
	return cmd
}

func runCreateStepScrollToTop(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	position, err := strconv.Atoi(args[1])
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
	client := virtuoso.NewClientDirect(baseURL, token)

	// Create the scroll step
	stepID, err := client.CreateStepScrollToTop(checkpointID, position)
	if err != nil {
		return fmt.Errorf("failed to create scroll to top step: %w", err)
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
		fmt.Printf("Created scroll to top step with ID %d for checkpoint %d. Position: %d\n", 
			stepID, checkpointID, position)
	default: // human
		fmt.Printf("Scroll to top step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
