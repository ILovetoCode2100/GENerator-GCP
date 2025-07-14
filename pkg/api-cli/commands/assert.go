package commands

import (
	"fmt"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// assertType represents the type of assertion being performed
type assertType string

const (
	assertExists             assertType = "exists"
	assertNotExists          assertType = "not-exists"
	assertEquals             assertType = "equals"
	assertNotEquals          assertType = "not-equals"
	assertChecked            assertType = "checked"
	assertSelected           assertType = "selected"
	assertGreaterThan        assertType = "gt"
	assertGreaterThanOrEqual assertType = "gte"
	assertLessThan           assertType = "lt"
	assertLessThanOrEqual    assertType = "lte"
	assertMatches            assertType = "matches"
	assertVariable           assertType = "variable"
)

// assertCommandInfo contains metadata about each assertion type
type assertCommandInfo struct {
	stepType    string
	description string
	usage       string
	examples    []string
	argsCount   []int // Valid argument counts (excluding position)
	parseStep   func(args []string) string
}

// assertCommands maps assertion types to their metadata
var assertCommands = map[assertType]assertCommandInfo{
	assertExists: {
		stepType:    "ASSERT_EXISTS",
		description: "Assert that an element exists",
		usage:       "assert exists ELEMENT [POSITION]",
		examples: []string{
			`api-cli assert exists "Login button" 1`,
			`api-cli assert exists "Success message"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("see \"%s\"", args[0])
		},
	},
	assertNotExists: {
		stepType:    "ASSERT_NOT_EXISTS",
		description: "Assert that an element does not exist",
		usage:       "assert not-exists ELEMENT [POSITION]",
		examples: []string{
			`api-cli assert not-exists "Error message" 1`,
			`api-cli assert not-exists "Loading spinner"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("not see \"%s\"", args[0])
		},
	},
	assertEquals: {
		stepType:    "ASSERT_EQUALS",
		description: "Assert that an element has a specific text value",
		usage:       "assert equals ELEMENT VALUE [POSITION]",
		examples: []string{
			`api-cli assert equals "Username field" "john@example.com" 1`,
			`api-cli assert equals "Total price" "$99.99"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect %s to have text \"%s\"", args[0], args[1])
		},
	},
	assertNotEquals: {
		stepType:    "ASSERT_NOT_EQUALS",
		description: "Assert that an element does not have a specific text value",
		usage:       "assert not-equals ELEMENT VALUE [POSITION]",
		examples: []string{
			`api-cli assert not-equals "Status" "Error" 1`,
			`api-cli assert not-equals "Username" "admin"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect %s to not have text \"%s\"", args[0], args[1])
		},
	},
	assertChecked: {
		stepType:    "ASSERT_CHECKED",
		description: "Assert that a checkbox is checked",
		usage:       "assert checked ELEMENT [POSITION]",
		examples: []string{
			`api-cli assert checked "Terms checkbox" 1`,
			`api-cli assert checked "Remember me"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect \"%s\" to be checked", args[0])
		},
	},
	assertSelected: {
		stepType:    "ASSERT_SELECTED",
		description: "Assert that an option is selected",
		usage:       "assert selected ELEMENT [POSITION]",
		examples: []string{
			`api-cli assert selected "Country dropdown" 1`,
			`api-cli assert selected "Payment method"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect \"%s\" to be selected", args[0])
		},
	},
	assertGreaterThan: {
		stepType:    "ASSERT_GREATER_THAN",
		description: "Assert that an element's value is greater than a number",
		usage:       "assert gt ELEMENT VALUE [POSITION]",
		examples: []string{
			`api-cli assert gt "Price" "10" 1`,
			`api-cli assert gt "Quantity" "0"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect %s to be greater than %s", args[0], args[1])
		},
	},
	assertGreaterThanOrEqual: {
		stepType:    "ASSERT_GREATER_THAN_OR_EQUAL",
		description: "Assert that an element's value is greater than or equal to a number",
		usage:       "assert gte ELEMENT VALUE [POSITION]",
		examples: []string{
			`api-cli assert gte "Score" "50" 1`,
			`api-cli assert gte "Balance" "0"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect %s to be greater than or equal to %s", args[0], args[1])
		},
	},
	assertLessThan: {
		stepType:    "ASSERT_LESS_THAN",
		description: "Assert that an element's value is less than a number",
		usage:       "assert lt ELEMENT VALUE [POSITION]",
		examples: []string{
			`api-cli assert lt "Error count" "5" 1`,
			`api-cli assert lt "Processing time" "1000"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect %s to be less than %s", args[0], args[1])
		},
	},
	assertLessThanOrEqual: {
		stepType:    "ASSERT_LESS_THAN_OR_EQUAL",
		description: "Assert that an element's value is less than or equal to a number",
		usage:       "assert lte ELEMENT VALUE [POSITION]",
		examples: []string{
			`api-cli assert lte "Discount" "100" 1`,
			`api-cli assert lte "Items" "10"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect %s to be less than or equal to %s", args[0], args[1])
		},
	},
	assertMatches: {
		stepType:    "ASSERT_MATCHES",
		description: "Assert that an element matches a regular expression",
		usage:       "assert matches ELEMENT PATTERN [POSITION]",
		examples: []string{
			`api-cli assert matches "Email field" "^[\\w.-]+@[\\w.-]+\\.\\w+$" 1`,
			`api-cli assert matches "Phone" "^\\d{3}-\\d{3}-\\d{4}$"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect %s to match pattern \"%s\"", args[0], args[1])
		},
	},
	assertVariable: {
		stepType:    "ASSERT_VARIABLE",
		description: "Assert that a stored variable has the expected value",
		usage:       "assert variable VARIABLE_NAME EXPECTED_VALUE [POSITION]",
		examples: []string{
			`api-cli assert variable "orderId" "12345" 1`,
			`api-cli assert variable "username" "john.doe"  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("expect $%s to equal \"%s\"", args[0], args[1])
		},
	},
}

// newAssertCmd creates the consolidated assert command with subcommands
func newAssertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assert",
		Short: "Create assertion steps in checkpoints",
		Long: `Create various types of assertion steps in checkpoints.

This command consolidates all assertion operations into a single command with subcommands for each assertion type.

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
		Example: `  # Assert element exists
  api-cli assert exists "Login button" 1

  # Assert element text equals value
  api-cli assert equals "Username" "john@example.com"

  # Assert numeric comparison
  api-cli assert gt "Price" "10"

  # Assert pattern match
  api-cli assert matches "Email" "^[\\w.-]+@[\\w.-]+\\.\\w+$"`,
	}

	// Add subcommands for each assertion type
	for aType, info := range assertCommands {
		cmd.AddCommand(newAssertSubCmd(aType, info))
	}

	return cmd
}

// extractArgsFromUsage extracts the arguments part from the usage string
func extractArgsFromUsage(usage string) string {
	parts := strings.Fields(usage)
	if len(parts) > 2 {
		return strings.Join(parts[2:], " ")
	}
	return ""
}

// newAssertSubCmd creates a subcommand for a specific assertion type
func newAssertSubCmd(aType assertType, info assertCommandInfo) *cobra.Command {
	var checkpointFlag int

	cmd := &cobra.Command{
		Use:   string(aType) + " " + extractArgsFromUsage(info.usage),
		Short: info.description,
		Long: fmt.Sprintf(`%s

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
%s`, info.description, strings.Join(info.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			// Validate argument count
			validCounts := info.argsCount
			for _, count := range validCounts {
				if len(args) == count || len(args) == count+1 {
					return nil
				}
			}

			// Generate expected count message
			expectedCounts := []string{}
			for _, count := range validCounts {
				expectedCounts = append(expectedCounts, fmt.Sprintf("%d", count))
				expectedCounts = append(expectedCounts, fmt.Sprintf("%d", count+1))
			}

			return fmt.Errorf("accepts %s args, received %d", strings.Join(expectedCounts, " or "), len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAssertCommand(aType, info, args, checkpointFlag)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// runAssertCommand executes the assertion command logic
func runAssertCommand(aType assertType, info assertCommandInfo, args []string, checkpointFlag int) error {
	// Validate arguments based on assertion type
	if err := validateAssertArgs(aType, args); err != nil {
		return err
	}

	// Resolve checkpoint and position
	positionIndex := info.argsCount[0] // Position comes after required args
	ctx, err := resolveStepContext(args, checkpointFlag, positionIndex)
	if err != nil {
		return err
	}

	// Create Virtuoso client
	apiClient := client.NewClient(cfg)

	// Call the appropriate API method based on assertion type
	stepID, err := callAssertAPI(apiClient, aType, ctx, args)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", info.stepType, err)
	}

	// Save config if position was auto-incremented
	saveStepContext(ctx)

	// Build extra data for output
	extra := buildExtraData(aType, args)

	// Output result
	output := &StepOutput{
		Status:       "success",
		StepType:     info.stepType,
		CheckpointID: ctx.CheckpointID,
		StepID:       stepID,
		Position:     ctx.Position,
		ParsedStep:   info.parseStep(args),
		UsingContext: ctx.UsingContext,
		AutoPosition: ctx.AutoPosition,
		Extra:        extra,
	}

	return outputStepResult(output)
}

// validateAssertArgs validates arguments for a specific assertion type
func validateAssertArgs(aType assertType, args []string) error {
	switch aType {
	case assertExists, assertNotExists, assertChecked, assertSelected:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("element cannot be empty")
		}
	case assertEquals, assertNotEquals, assertGreaterThan, assertGreaterThanOrEqual,
		assertLessThan, assertLessThanOrEqual, assertMatches:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("element cannot be empty")
		}
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("value cannot be empty")
		}
	case assertVariable:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("variable name cannot be empty")
		}
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("expected value cannot be empty")
		}
	}
	return nil
}

// callAssertAPI calls the appropriate client API method for the assertion type
func callAssertAPI(apiClient *client.Client, aType assertType, ctx *StepContext, args []string) (int, error) {
	switch aType {
	case assertExists:
		return apiClient.CreateAssertExistsStep(ctx.CheckpointID, args[0], ctx.Position)
	case assertNotExists:
		return apiClient.CreateAssertNotExistsStep(ctx.CheckpointID, args[0], ctx.Position)
	case assertEquals:
		return apiClient.CreateAssertEqualsStep(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case assertNotEquals:
		return apiClient.CreateAssertNotEqualsStep(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case assertChecked:
		return apiClient.CreateAssertCheckedStep(ctx.CheckpointID, args[0], ctx.Position)
	case assertSelected:
		return apiClient.CreateAssertSelectedStep(ctx.CheckpointID, args[0], ctx.Position)
	case assertGreaterThan:
		return apiClient.CreateAssertGreaterThanStep(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case assertGreaterThanOrEqual:
		return apiClient.CreateAssertGreaterThanOrEqualStep(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case assertLessThan:
		return apiClient.CreateAssertLessThanStep(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case assertLessThanOrEqual:
		return apiClient.CreateAssertLessThanOrEqualStep(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case assertMatches:
		return apiClient.CreateAssertMatchesStep(ctx.CheckpointID, args[0], args[1], ctx.Position)
	case assertVariable:
		return apiClient.CreateAssertVariableStep(ctx.CheckpointID, args[0], args[1], ctx.Position)
	default:
		return 0, fmt.Errorf("unsupported assertion type: %s", aType)
	}
}

// buildExtraData builds the extra data map for output based on assertion type
func buildExtraData(aType assertType, args []string) map[string]interface{} {
	extra := make(map[string]interface{})

	switch aType {
	case assertExists, assertNotExists, assertChecked, assertSelected:
		extra["element"] = args[0]
	case assertEquals, assertNotEquals, assertGreaterThan, assertGreaterThanOrEqual,
		assertLessThan, assertLessThanOrEqual, assertMatches:
		extra["element"] = args[0]
		extra["value"] = args[1]
	case assertVariable:
		extra["variable_name"] = args[0]
		extra["expected_value"] = args[1]
	}

	return extra
}
