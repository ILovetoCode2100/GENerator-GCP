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

func newCreateStepMouseMoveByCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "create-step-mouse-move-by CHECKPOINT_ID X Y POSITION",
		Short: "Create a mouse move by relative offset step",
		Long: `Create a mouse move by relative offset step in a checkpoint.

This command creates a step that will move the mouse by the specified X,Y offset.
The coordinates are relative to the current mouse position.

Examples:
  # Create a mouse move by step at position 1
  api-cli create-step-mouse-move-by 1678318 -10 40 1

  # Create with JSON output
  api-cli create-step-mouse-move-by 1678318 50 -20 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %v", err)
			}

			x, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid X offset: %v", err)
			}

			y, err := strconv.Atoi(args[2])
			if err != nil {
				return fmt.Errorf("invalid Y offset: %v", err)
			}

			position, err := strconv.Atoi(args[3])
			if err != nil {
				return fmt.Errorf("invalid position: %v", err)
			}

			// Get API configuration
			token := os.Getenv("VIRTUOSO_API_TOKEN")
			if token == "" {
				return fmt.Errorf("VIRTUOSO_API_TOKEN environment variable is required")
			}

			baseURL := os.Getenv("VIRTUOSO_API_BASE_URL")
			if baseURL == "" {
				baseURL = "https://api-app2.virtuoso.qa/api"
			}

			// Create client
			client := virtuoso.NewClientDirect(baseURL, token)

			// Create the step
			stepID, err := client.CreateStepMouseMoveBy(checkpointID, x, y, position)
			if err != nil {
				return fmt.Errorf("failed to create mouse move by step: %v", err)
			}

			// Output the response in the requested format
			return outputMouseMoveByResponse(stepID, checkpointID, position, outputFormat, x, y)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, yaml, ai)")
	return cmd
}

func outputMouseMoveByResponse(stepID, checkpointID, position int, format string, x, y int) error {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(map[string]interface{}{"stepId": stepID, "checkpointId": checkpointID}, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(jsonData))

	case "yaml":
		yamlData, err := yaml.Marshal(map[string]interface{}{"stepId": stepID, "checkpointId": checkpointID})
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %v", err)
		}
		fmt.Print(string(yamlData))

	case "ai":
		fmt.Printf("Mouse move by step created successfully:\n")
		fmt.Printf("- Step ID: %d\n", stepID)
		fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("- X Offset: %d\n", x)
		fmt.Printf("- Y Offset: %d\n", y)
		fmt.Printf("- Position: %d\n", position)

	default: // human
		fmt.Printf("Mouse move by step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Offset: (%d, %d)\n", x, y)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
