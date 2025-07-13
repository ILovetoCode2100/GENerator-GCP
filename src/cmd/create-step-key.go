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

// newCreateStepKeyCmd creates the command for creating a key press step
func newCreateStepKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-key CHECKPOINT_ID KEY POSITION [flags]",
		Short: "Press a key or key combination",
		Long: `Creates a key press step that presses the specified key or key combination.

Examples:
  api-cli create-step-key 1678318 "CTRL_a" 1
  api-cli create-step-key 1678318 "RETURN" 1 --target "Search"
  api-cli create-step-key 1678318 "F1" 1 --target "body" -o json

Common keys:
  - Single keys: RETURN, ESCAPE, TAB, SPACE, DELETE, BACKSPACE
  - Function keys: F1, F2, F3, etc.
  - Modifiers: CTRL_a, CTRL_c, CTRL_v, ALT_F4, SHIFT_TAB
  - Arrow keys: UP, DOWN, LEFT, RIGHT`,
		Args: cobra.ExactArgs(3),
		RunE: runCreateStepKey,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	cmd.Flags().String("target", "", "Target element selector (if not specified, key is pressed globally)")
	
	return cmd
}

func runCreateStepKey(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	key := args[1]
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	position, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	// Get flags
	target, _ := cmd.Flags().GetString("target")

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

	// Create the key step
	var stepID int
	if target != "" {
		stepID, err = client.CreateStepKeyTargeted(checkpointID, target, key, position)
	} else {
		stepID, err = client.CreateStepKeyGlobal(checkpointID, key, position)
	}

	if err != nil {
		return fmt.Errorf("failed to create key step: %w", err)
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
		targetInfo := ""
		if target != "" {
			targetInfo = fmt.Sprintf(" on target '%s'", target)
		} else {
			targetInfo = " globally"
		}
		fmt.Printf("Created key press step with ID %d for checkpoint %d. Key: %s%s, position: %d\n", 
			stepID, checkpointID, key, targetInfo, position)
	default: // human
		fmt.Printf("Key press step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		fmt.Printf("Key: %s\n", key)
		if target != "" {
			fmt.Printf("Target: %s\n", target)
		} else {
			fmt.Printf("Target: Global\n")
		}
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
