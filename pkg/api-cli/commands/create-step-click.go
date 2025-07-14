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

// newCreateStepClickCmd creates the command for creating a click step
func newCreateStepClickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-click CHECKPOINT_ID SELECTOR POSITION [flags]",
		Short: "Click on an element",
		Long: `Creates a click step that clicks on the specified element.

Examples:
  api-cli create-step-click 1678318 "Submit" 1
  api-cli create-step-click 1678318 "Login" 1 --position TOP_RIGHT --element-type BUTTON
  api-cli create-step-click 1678318 "" 1 --variable "variableTarget"
  api-cli create-step-click 1678318 "Submit" 1 -o json`,
		Args: cobra.ExactArgs(3),
		RunE: runCreateStepClick,
	}

	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")
	cmd.Flags().String("variable", "", "Use variable as target instead of selector")
	cmd.Flags().String("position", "", "Element position (e.g., TOP_RIGHT, BOTTOM_LEFT)")
	cmd.Flags().String("element-type", "", "Element type (e.g., BUTTON, INPUT, LINK)")

	return cmd
}

func runCreateStepClick(cmd *cobra.Command, args []string) error {
	// Parse arguments
	checkpointID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	selector := args[1]

	position, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	// Get flags
	variable, _ := cmd.Flags().GetString("variable")
	positionType, _ := cmd.Flags().GetString("position")
	elementType, _ := cmd.Flags().GetString("element-type")

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

	// Create the click step based on parameters
	var stepID int
	if variable != "" {
		stepID, err = client.CreateStepClickWithVariable(checkpointID, variable, position)
	} else if positionType != "" && elementType != "" {
		stepID, err = client.CreateStepClickWithDetails(checkpointID, selector, positionType, elementType, position)
	} else {
		stepID, err = client.CreateStepClick(checkpointID, selector, position)
	}

	if err != nil {
		return fmt.Errorf("failed to create click step: %w", err)
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
		target := selector
		if variable != "" {
			target = fmt.Sprintf("variable '%s'", variable)
		}
		fmt.Printf("Created click step with ID %d for checkpoint %d. Target: %s, position: %d\n",
			stepID, checkpointID, target, position)
	default: // human
		fmt.Printf("Click step created successfully!\n")
		fmt.Printf("Step ID: %d\n", stepID)
		fmt.Printf("Checkpoint ID: %d\n", checkpointID)
		if variable != "" {
			fmt.Printf("Variable Target: %s\n", variable)
		} else {
			fmt.Printf("Selector: %s\n", selector)
		}
		if positionType != "" {
			fmt.Printf("Position Type: %s\n", positionType)
		}
		if elementType != "" {
			fmt.Printf("Element Type: %s\n", elementType)
		}
		fmt.Printf("Position: %d\n", position)
	}

	return nil
}
