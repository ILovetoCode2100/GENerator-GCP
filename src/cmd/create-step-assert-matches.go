package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
)

func newCreateStepAssertMatchesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-step-assert-matches CHECKPOINT_ID ELEMENT REGEX_PATTERN POSITION",
		Short: "Create an assertion step that verifies an element matches a regex pattern at a specific position",
		Long: `Create an assertion step that verifies an element matches a regex pattern at the specified position in the checkpoint.
		
Example:
  api-cli create-step-assert-matches 1678318 "Email" "/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/" 1
  api-cli create-step-assert-matches 1678318 "#email-field" "/.*@example\.com/" 2 -o json`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkpointIDStr := args[0]
			element := args[1]
			regexPattern := args[2]
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
			
			// Validate element
			if element == "" {
				return fmt.Errorf("element cannot be empty")
			}
			
			// Validate regex pattern
			if regexPattern == "" {
				return fmt.Errorf("regex pattern cannot be empty")
			}
			
			// Create Virtuoso client
			client := virtuoso.NewClient(cfg)
			
			// Create assert matches step using the client
			stepID, err := client.CreateAssertMatchesStep(checkpointID, element, regexPattern, position)
			if err != nil {
				return fmt.Errorf("failed to create assert matches step: %w", err)
			}
			
			// Format output based on the format flag
			switch cfg.Output.DefaultFormat {
			case "json":
				output := map[string]interface{}{
					"status":        "success",
					"step_type":     "ASSERT_MATCHES",
					"checkpoint_id": checkpointID,
					"step_id":       stepID,
					"element":       element,
					"regex_pattern": regexPattern,
					"position":      position,
					"parsed_step":   fmt.Sprintf("expect %s to match pattern \"%s\"", element, regexPattern),
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(output); err != nil {
					return fmt.Errorf("failed to encode JSON output: %w", err)
				}
			case "yaml":
				fmt.Printf("status: success\n")
				fmt.Printf("step_type: ASSERT_MATCHES\n")
				fmt.Printf("checkpoint_id: %d\n", checkpointID)
				fmt.Printf("step_id: %d\n", stepID)
				fmt.Printf("element: %s\n", element)
				fmt.Printf("regex_pattern: %s\n", regexPattern)
				fmt.Printf("position: %d\n", position)
				fmt.Printf("parsed_step: expect %s to match pattern \"%s\"\n", element, regexPattern)
			case "ai":
				fmt.Printf("Successfully created assert matches step:\n")
				fmt.Printf("- Step ID: %d\n", stepID)
				fmt.Printf("- Step Type: ASSERT_MATCHES\n")
				fmt.Printf("- Checkpoint ID: %d\n", checkpointID)
				fmt.Printf("- Element: %s\n", element)
				fmt.Printf("- Regex Pattern: %s\n", regexPattern)
				fmt.Printf("- Position: %d\n", position)
				fmt.Printf("- Parsed Step: expect %s to match pattern \"%s\"\n", element, regexPattern)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. Add another step: api-cli create-step-* %d <options>\n", checkpointID)
				fmt.Printf("2. Execute the test journey\n")
			default: // human
				fmt.Printf("✅ Created assert matches step at position %d in checkpoint %d\n", position, checkpointID)
				fmt.Printf("   Element: %s\n", element)
				fmt.Printf("   Regex Pattern: %s\n", regexPattern)
				fmt.Printf("   Step ID: %d\n", stepID)
			}
			
			return nil
		},
	}
	
	return cmd
}