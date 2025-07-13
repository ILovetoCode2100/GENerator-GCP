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

func newCreateStepExecuteScriptCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "create-step-execute-script CHECKPOINT_ID SCRIPT_NAME POSITION",
		Short: "Create a step to execute a custom script",
		Long: `Create a step to execute a custom script in a checkpoint.

This command creates a step that will execute a custom script by name.
The script must be available in the Virtuoso environment.

Examples:
  # Create an execute script step at position 1
  api-cli create-step-execute-script 1678318 "my-login-script" 1

  # Create with JSON output
  api-cli create-step-execute-script 1678318 "validation-script" 2 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %v", err)
			}

			scriptName := args[1]
			if scriptName == "" {
				return fmt.Errorf("script name cannot be empty")
			}

			position, err := strconv.Atoi(args[2])
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
			stepID, err := client.CreateStepExecuteScript(checkpointID, scriptName, position)
			if err != nil {
				return fmt.Errorf("failed to create execute script step: %v", err)
			}

			// Output the response in the requested format
			return outputResponse(stepID, checkpointID, scriptName, position, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "human", "Output format (human, json, yaml, ai)")
	return cmd
}

func outputResponse(stepID, checkpointID int, scriptName string, position int, format string) error {
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
		fmt.Printf("Created execute script step with ID %d for checkpoint %d. Script: %s, position: %d\n",
			stepID, checkpointID, scriptName, position)

	default: // human
		fmt.Printf("Execute script step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Script Name: %s\n", scriptName)
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
