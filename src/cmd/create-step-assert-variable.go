package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertVariableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-assert-variable CHECKPOINT_ID VARIABLE_NAME EXPECTED_VALUE POSITION",
		Short: "Create an assert variable step at a specific position in a checkpoint",
		Long: `Create an assert variable step that verifies a stored variable has the expected value at the specified position in the checkpoint.
		
Example:
  api-cli create-step-assert-variable 1678318 "orderId" "12345" 1
  api-cli create-step-assert-variable 1678318 "username" "john.doe" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			variableName := args[1]
			expectedValue := args[2]
			positionStr := args[3]
			
			// Convert IDs to int
			checkpointID, err := strconv.Atoi(checkpointIDStr)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}
			
			position, err := strconv.Atoi(positionStr)
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			
			// Validate inputs
			if variableName == "" {
				return fmt.Errorf("variable name cannot be empty")
			}
			if expectedValue == "" {
				return fmt.Errorf("expected value cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create assert variable step using the enhanced client
			stepID, err := client.CreateAssertVariableStep(checkpointID, variableName, expectedValue, position)
			if err != nil {
				return fmt.Errorf("failed to create assert variable step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":         "success",
					"step_type":      "ASSERT_VARIABLE",
					"checkpoint_id":  checkpointID,
					"step_id":        stepID,
					"variable_name":  variableName,
					"expected_value": expectedValue,
					"position":       position,
					"parsed_step":    fmt.Sprintf("expect $%s to equal \"%s\"", variableName, expectedValue),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: ASSERT_VARIABLE\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("variable_name: %s\n", variableName)
				fmt.Printf("expected_value: %s\n", expectedValue)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: expect $%s to equal \"%s\"\n", variableName, expectedValue)
			case "ai":
				fmt.Printf("Successfully created assert variable step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: ASSERT_VARIABLE\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Variable Name: %s\n", variableName)
				fmt.Printf("- Expected Value: %s\n", expectedValue)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: expect $%s to equal \"%s\"\n", variableName, expectedValue)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created assert variable step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Variable: $%s\n", variableName)
				fmt.Printf("   Expected Value: %s\n", expectedValue)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}
