package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// selectAction represents the type of select action
type selectAction string

const (
	selectIndex selectAction = "index"
	selectLast  selectAction = "last"
)

// selectCommandInfo contains metadata about each select action
type selectCommandInfo struct {
	stepType    string
	description string
	usage       string
	examples    []string
	argsCount   []int // Valid argument counts (excluding position)
	parseStep   func(args []string) string
}

// selectCommands maps select actions to their metadata
var selectCommands = map[selectAction]selectCommandInfo{
	selectIndex: {
		stepType:    "PICK",
		description: "Pick dropdown option by index",
		usage:       "select index SELECTOR INDEX [POSITION]",
		examples: []string{
			`api-cli select index "Country dropdown" 2 1`,
			`api-cli select index "#country-select" 0  # First option, auto-increment position`,
			`api-cli select index ".dropdown" 3  # Fourth option (0-based index)`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("pick option at index %s from \"%s\"", args[1], args[0])
		},
	},
	selectLast: {
		stepType:    "PICK",
		description: "Pick the last dropdown option",
		usage:       "select last SELECTOR [POSITION]",
		examples: []string{
			`api-cli select last "Country dropdown" 1`,
			`api-cli select last "#country-select"  # Auto-increment position`,
			`api-cli select last ".dropdown"  # Pick last option`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("pick last option from \"%s\"", args[0])
		},
	},
}

// newSelectCmd creates the consolidated select command with subcommands
func newSelectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select options from dropdowns",
		Long: `Select options from dropdown elements using various methods.

This command consolidates dropdown selection operations:
  - Select by index (0-based)
  - Select last option

Index is 0-based, so first option is 0, second is 1, etc.`,
		Example: `  # Select option by index (0-based)
  api-cli select index "Country dropdown" 2

  # Select last option
  api-cli select last "#country-select"

  # With explicit position
  api-cli select index ".dropdown" 0 5`,
	}

	// Add subcommands for each select action
	for action, info := range selectCommands {
		cmd.AddCommand(newSelectSubCmd(action, info))
	}

	return cmd
}

// newSelectSubCmd creates a subcommand for a specific select action
func newSelectSubCmd(action selectAction, info selectCommandInfo) *cobra.Command {
	var checkpointFlag int

	// Extract command name and args from usage
	usageParts := strings.Fields(info.usage)
	cmdName := string(action)
	argsUsage := ""
	if len(usageParts) > 2 {
		argsUsage = strings.Join(usageParts[2:], " ")
	}

	cmd := &cobra.Command{
		Use:   cmdName + " " + argsUsage,
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
			return runSelectCommand(action, info, args, checkpointFlag)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// runSelectCommand executes the select command logic
func runSelectCommand(action selectAction, info selectCommandInfo, args []string, checkpointFlag int) error {
	// Validate arguments based on action type
	if err := validateSelectArgs(action, args); err != nil {
		return err
	}

	// Resolve checkpoint and position
	positionIndex := len(info.argsCount) // Position comes after required args
	ctx, err := resolveStepContext(args, checkpointFlag, positionIndex)
	if err != nil {
		return err
	}

	// Create Virtuoso client
	apiClient := client.NewClient(cfg)

	// Call the appropriate API method based on action type
	stepID, err := callSelectAPI(apiClient, action, ctx, args)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", info.stepType, err)
	}

	// Save config if position was auto-incremented
	saveStepContext(ctx)

	// Build extra data for output
	extra := buildSelectExtraData(action, args)

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

// validateSelectArgs validates arguments for a specific select action
func validateSelectArgs(action selectAction, args []string) error {
	switch action {
	case selectIndex:
		if len(args) < 2 {
			return fmt.Errorf("selector and index are required")
		}
		if args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
		// Validate index is a non-negative integer
		index, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("index must be a valid integer: %w", err)
		}
		if index < 0 {
			return fmt.Errorf("index must be 0 or greater (got %d)", index)
		}
	case selectLast:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
	}
	return nil
}

// callSelectAPI calls the appropriate client API method for the select action
func callSelectAPI(apiClient *client.Client, action selectAction, ctx *StepContext, args []string) (int, error) {
	switch action {
	case selectIndex:
		index, _ := strconv.Atoi(args[1])
		return apiClient.CreateStepPickIndex(ctx.CheckpointID, args[0], index, ctx.Position)
	case selectLast:
		return apiClient.CreateStepPickLast(ctx.CheckpointID, args[0], ctx.Position)
	default:
		return 0, fmt.Errorf("unsupported select action: %s", action)
	}
}

// buildSelectExtraData builds the extra data map for output based on select action
func buildSelectExtraData(action selectAction, args []string) map[string]interface{} {
	extra := make(map[string]interface{})
	extra["selector"] = args[0]

	switch action {
	case selectIndex:
		index, _ := strconv.Atoi(args[1])
		extra["index"] = index
		extra["method"] = "index"
	case selectLast:
		extra["method"] = "last"
		extra["index"] = -1 // Last option is represented as -1 in API
	}

	return extra
}
