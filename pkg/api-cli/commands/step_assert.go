package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// AssertCommand implements the assert command group using BaseCommand pattern
type AssertCommand struct {
	*BaseCommand
	assertType string
}

// assertConfig contains configuration for each assertion type
type assertConfig struct {
	stepType     string
	description  string
	usage        string
	examples     []string
	requiredArgs int
	buildMeta    func(args []string) map[string]interface{}
}

// assertConfigs maps assertion types to their configurations
var assertConfigs = map[string]assertConfig{
	"exists": {
		stepType:    "ASSERT_EXISTS",
		description: "Assert that an element exists",
		usage:       "assert exists [checkpoint-id] <element> [position]",
		examples: []string{
			`api-cli step-assert exists cp_12345 "Login button" 1`,
			`api-cli step-assert exists "Login button" --checkpoint 12345`,
			`api-cli step-assert exists "Login button"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
			}
		},
	},
	"not-exists": {
		stepType:    "ASSERT_NOT_EXISTS",
		description: "Assert that an element does not exist",
		usage:       "assert not-exists [checkpoint-id] <element> [position]",
		examples: []string{
			`api-cli step-assert not-exists cp_12345 "Error message" 1`,
			`api-cli step-assert not-exists "Error message" --checkpoint 12345`,
			`api-cli step-assert not-exists "Error message"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
			}
		},
	},
	"equals": {
		stepType:    "ASSERT_EQUALS",
		description: "Assert that an element has a specific text value",
		usage:       "assert equals [checkpoint-id] <element> <value> [position]",
		examples: []string{
			`api-cli step-assert equals cp_12345 "Username field" "john@example.com" 1`,
			`api-cli step-assert equals "Total price" "$99.99"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"value":    args[1],
			}
		},
	},
	"not-equals": {
		stepType:    "ASSERT_NOT_EQUALS",
		description: "Assert that an element does not have a specific text value",
		usage:       "assert not-equals [checkpoint-id] <element> <value> [position]",
		examples: []string{
			`api-cli step-assert not-equals cp_12345 "Status" "Error" 1`,
			`api-cli step-assert not-equals "Username" "admin"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"value":    args[1],
			}
		},
	},
	"checked": {
		stepType:    "ASSERT_CHECKED",
		description: "Assert that a checkbox is checked",
		usage:       "assert checked [checkpoint-id] <element> [position]",
		examples: []string{
			`api-cli step-assert checked cp_12345 "Terms checkbox" 1`,
			`api-cli step-assert checked "Remember me"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
			}
		},
	},
	"selected": {
		stepType:    "ASSERT_SELECTED",
		description: "Assert that an option is selected",
		usage:       "assert selected [checkpoint-id] <element> [position]",
		examples: []string{
			`api-cli step-assert selected cp_12345 "Country dropdown" 1`,
			`api-cli step-assert selected "Language selector"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
			}
		},
	},
	"gt": {
		stepType:    "ASSERT_GREATER_THAN",
		description: "Assert that a value is greater than another",
		usage:       "assert gt [checkpoint-id] <element> <value> [position]",
		examples: []string{
			`api-cli step-assert gt cp_12345 "Price" "10" 1`,
			`api-cli step-assert gt "Score" "100"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"value":    args[1],
			}
		},
	},
	"gte": {
		stepType:    "ASSERT_GREATER_THAN_OR_EQUAL",
		description: "Assert that a value is greater than or equal to another",
		usage:       "assert gte [checkpoint-id] <element> <value> [position]",
		examples: []string{
			`api-cli step-assert gte cp_12345 "Age" "18" 1`,
			`api-cli step-assert gte "Count" "0"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"value":    args[1],
			}
		},
	},
	"lt": {
		stepType:    "ASSERT_LESS_THAN",
		description: "Assert that a value is less than another",
		usage:       "assert lt [checkpoint-id] <element> <value> [position]",
		examples: []string{
			`api-cli step-assert lt cp_12345 "Error count" "5" 1`,
			`api-cli step-assert lt "Temperature" "32"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"value":    args[1],
			}
		},
	},
	"lte": {
		stepType:    "ASSERT_LESS_THAN_OR_EQUAL",
		description: "Assert that a value is less than or equal to another",
		usage:       "assert lte [checkpoint-id] <element> <value> [position]",
		examples: []string{
			`api-cli step-assert lte cp_12345 "Stock" "100" 1`,
			`api-cli step-assert lte "Discount" "50"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"value":    args[1],
			}
		},
	},
	"matches": {
		stepType:    "ASSERT_MATCHES",
		description: "Assert that an element matches a regex pattern",
		usage:       "assert matches [checkpoint-id] <element> <pattern> [position]",
		examples: []string{
			`api-cli step-assert matches cp_12345 "Email" "^[\\w.-]+@[\\w.-]+\\.\\w+$" 1`,
			`api-cli step-assert matches "Phone" "^\\d{3}-\\d{3}-\\d{4}$"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"pattern":  args[1],
			}
		},
	},
	"variable": {
		stepType:    "ASSERT_VARIABLE",
		description: "Assert that a variable equals a value",
		usage:       "assert variable [checkpoint-id] <variable> <value> [position]",
		examples: []string{
			`api-cli step-assert variable cp_12345 "userRole" "admin" 1`,
			`api-cli step-assert variable "loginStatus" "success"  # Uses session context`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"variable": args[0],
				"value":    args[1],
			}
		},
	},
}

// newStepAssertCmd creates the new assert command using BaseCommand pattern
func newStepAssertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step-assert",
		Short: "Create assertion steps in checkpoints",
		Long: `Create various types of assertion steps in checkpoints.

This command uses the standardized positional argument pattern:
- Optional checkpoint ID as first argument (falls back to session context)
- Required assertion arguments
- Optional position as last argument (auto-increments if not specified)

Available assertion types:
  - exists: Assert element exists
  - not-exists: Assert element does not exist
  - equals: Assert element text equals value
  - not-equals: Assert element text does not equal value
  - checked: Assert checkbox is checked
  - selected: Assert option is selected
  - gt: Assert value is greater than
  - gte: Assert value is greater than or equal to
  - lt: Assert value is less than
  - lte: Assert value is less than or equal to
  - matches: Assert element matches regex pattern
  - variable: Assert variable equals value`,
		Example: `  # Assert element exists (with explicit checkpoint)
  api-cli step-assert exists cp_12345 "Login button" 1

  # Assert element exists (using session context)
  api-cli step-assert exists "Login button"

  # Assert element text equals value
  api-cli step-assert equals "Username" "john@example.com"

  # Assert numeric comparison
  api-cli step-assert gt "Price" "10"

  # Assert pattern match
  api-cli step-assert matches "Email" "^[\\w.-]+@[\\w.-]+\\.\\w+$"`,
	}

	// Add subcommands for each assertion type
	for assertType, config := range assertConfigs {
		cmd.AddCommand(newAssertV2SubCmd(assertType, config))
	}

	return cmd
}

// newAssertV2SubCmd creates a subcommand for a specific assertion type
func newAssertV2SubCmd(assertType string, config assertConfig) *cobra.Command {
	var checkpointFlag string

	cmd := &cobra.Command{
		Use:   assertType + " " + extractUsageArgs(config.usage),
		Short: config.description,
		Long: fmt.Sprintf(`%s

%s

Examples:
%s`, config.description, config.usage, strings.Join(config.examples, "\n")),
		RunE: func(cmd *cobra.Command, args []string) error {
			ac := &AssertCommand{
				BaseCommand: NewBaseCommand(),
				assertType:  assertType,
			}

			// If --checkpoint flag is set, prepend it to args
			if checkpointFlag != "" {
				args = append([]string{checkpointFlag}, args...)
			}

			return ac.Execute(cmd, args, config)
		},
	}

	// Add --checkpoint flag
	cmd.Flags().StringVar(&checkpointFlag, "checkpoint", "", "Checkpoint ID (alternative to positional argument)")

	return cmd
}

// extractUsageArgs extracts the arguments portion from the usage string
func extractUsageArgs(usage string) string {
	parts := strings.Fields(usage)
	if len(parts) > 2 {
		// Skip "assert" and subcommand
		return strings.Join(parts[2:], " ")
	}
	return ""
}

// Execute runs the assert command
func (ac *AssertCommand) Execute(cmd *cobra.Command, args []string, config assertConfig) error {
	// Initialize base command
	if err := ac.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	remainingArgs, err := ac.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate we have the required number of arguments
	if len(remainingArgs) != config.requiredArgs {
		return fmt.Errorf("expected %d arguments, got %d", config.requiredArgs, len(remainingArgs))
	}

	// Build request metadata
	meta := config.buildMeta(remainingArgs)

	// Create the step
	stepResult, err := ac.createAssertStep(config.stepType, meta)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", config.stepType, err)
	}

	// Format and output the result
	output, err := ac.FormatOutput(stepResult, ac.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format %s step output: %w", config.stepType, err)
	}

	fmt.Print(output)
	return nil
}

// createAssertStep creates an assertion step via the API
func (ac *AssertCommand) createAssertStep(stepType string, meta map[string]interface{}) (*StepResult, error) {
	// Convert checkpoint ID from string to int
	checkpointID, err := strconv.Atoi(ac.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint ID: %s", ac.CheckpointID)
	}

	// Create a context with timeout for the API operation
	ctx, cancel := ac.CommandContext()
	defer cancel()

	// Build the request based on step type
	var stepID int

	// Use the client to create the appropriate step with context
	switch stepType {
	case "ASSERT_EXISTS":
		stepID, err = ac.Client.CreateAssertExistsStepWithContext(ctx, checkpointID, meta["selector"].(string), ac.Position)
	case "ASSERT_NOT_EXISTS":
		stepID, err = ac.Client.CreateAssertNotExistsStepWithContext(ctx, checkpointID, meta["selector"].(string), ac.Position)
	case "ASSERT_EQUALS":
		stepID, err = ac.Client.CreateAssertEqualsStepWithContext(ctx, checkpointID, meta["selector"].(string), meta["value"].(string), ac.Position)
	case "ASSERT_NOT_EQUALS":
		stepID, err = ac.Client.CreateAssertNotEqualsStepWithContext(ctx, checkpointID, meta["selector"].(string), meta["value"].(string), ac.Position)
	case "ASSERT_CHECKED":
		stepID, err = ac.Client.CreateAssertCheckedStepWithContext(ctx, checkpointID, meta["selector"].(string), ac.Position)
	case "ASSERT_SELECTED":
		stepID, err = ac.Client.CreateAssertSelectedStepWithContext(ctx, checkpointID, meta["selector"].(string), ac.Position)
	case "ASSERT_GREATER_THAN":
		stepID, err = ac.Client.CreateAssertGreaterThanStepWithContext(ctx, checkpointID, meta["selector"].(string), meta["value"].(string), ac.Position)
	case "ASSERT_GREATER_THAN_OR_EQUAL":
		stepID, err = ac.Client.CreateAssertGreaterThanOrEqualStepWithContext(ctx, checkpointID, meta["selector"].(string), meta["value"].(string), ac.Position)
	case "ASSERT_LESS_THAN":
		stepID, err = ac.Client.CreateAssertLessThanStepWithContext(ctx, checkpointID, meta["selector"].(string), meta["value"].(string), ac.Position)
	case "ASSERT_LESS_THAN_OR_EQUAL":
		stepID, err = ac.Client.CreateAssertLessThanOrEqualStepWithContext(ctx, checkpointID, meta["selector"].(string), meta["value"].(string), ac.Position)
	case "ASSERT_MATCHES":
		stepID, err = ac.Client.CreateAssertMatchesStepWithContext(ctx, checkpointID, meta["selector"].(string), meta["pattern"].(string), ac.Position)
	case "ASSERT_VARIABLE":
		stepID, err = ac.Client.CreateAssertVariableStepWithContext(ctx, checkpointID, meta["variable"].(string), meta["value"].(string), ac.Position)
	default:
		return nil, fmt.Errorf("unknown assertion type: %s", stepType)
	}

	if err != nil {
		// Handle different error types with specific messages
		if err == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timed out while creating %s step", stepType)
		}
		if err == context.Canceled {
			return nil, fmt.Errorf("request was canceled while creating %s step", stepType)
		}

		// Check for specific API error types
		if client.IsNotFound(err) {
			return nil, fmt.Errorf("checkpoint %d not found", checkpointID)
		}
		if client.IsUnauthorized(err) {
			return nil, fmt.Errorf("unauthorized: please check your API token")
		}
		if client.IsRateLimited(err) {
			return nil, fmt.Errorf("rate limited: please try again later")
		}
		if client.IsTimeout(err) {
			return nil, fmt.Errorf("API request timed out")
		}

		// For API errors, provide more context
		if apiErr, ok := err.(*client.APIError); ok {
			return nil, fmt.Errorf("API error creating %s step: %v", stepType, apiErr)
		}

		// Generic error
		return nil, fmt.Errorf("failed to create %s step: %w", stepType, err)
	}

	// Build the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: ac.CheckpointID,
		Type:         stepType,
		Position:     ac.Position,
		Description:  ac.buildDescription(stepType, meta),
		Selector:     ac.extractSelector(meta),
		Meta:         meta,
	}

	// Save session state if position was auto-incremented
	if ac.Position == -1 && cfg.Session.AutoIncrementPos {
		if err := cfg.SaveConfig(); err != nil {
			// Don't fail the command, just warn
			// Note: In production, this warning would be sent to stderr
		}
	}

	return result, nil
}

// buildDescription creates a human-readable description for the step
func (ac *AssertCommand) buildDescription(stepType string, meta map[string]interface{}) string {
	switch stepType {
	case "ASSERT_EXISTS":
		return fmt.Sprintf("see \"%s\"", meta["selector"])
	case "ASSERT_NOT_EXISTS":
		return fmt.Sprintf("not see \"%s\"", meta["selector"])
	case "ASSERT_EQUALS":
		return fmt.Sprintf("expect %s to have text \"%s\"", meta["selector"], meta["value"])
	case "ASSERT_NOT_EQUALS":
		return fmt.Sprintf("expect %s to not have text \"%s\"", meta["selector"], meta["value"])
	case "ASSERT_CHECKED":
		return fmt.Sprintf("expect %s to be checked", meta["selector"])
	case "ASSERT_SELECTED":
		return fmt.Sprintf("expect %s to be selected", meta["selector"])
	case "ASSERT_GREATER_THAN":
		return fmt.Sprintf("expect %s to be greater than %s", meta["selector"], meta["value"])
	case "ASSERT_GREATER_THAN_OR_EQUAL":
		return fmt.Sprintf("expect %s to be greater than or equal to %s", meta["selector"], meta["value"])
	case "ASSERT_LESS_THAN":
		return fmt.Sprintf("expect %s to be less than %s", meta["selector"], meta["value"])
	case "ASSERT_LESS_THAN_OR_EQUAL":
		return fmt.Sprintf("expect %s to be less than or equal to %s", meta["selector"], meta["value"])
	case "ASSERT_MATCHES":
		return fmt.Sprintf("expect %s to match pattern \"%s\"", meta["selector"], meta["pattern"])
	case "ASSERT_VARIABLE":
		return fmt.Sprintf("expect variable %s to equal \"%s\"", meta["variable"], meta["value"])
	default:
		return stepType
	}
}

// extractSelector extracts the selector from metadata if present
func (ac *AssertCommand) extractSelector(meta map[string]interface{}) string {
	if selector, ok := meta["selector"].(string); ok {
		return selector
	}
	return ""
}
