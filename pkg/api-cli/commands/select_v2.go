package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// SelectCommand implements the select command group using BaseCommand pattern
type SelectCommand struct {
	*BaseCommand
	selectType string
}

// selectConfig contains configuration for each select operation
type selectConfig struct {
	stepType     string
	description  string
	usage        string
	examples     []string
	requiredArgs int
	buildMeta    func(args []string) map[string]interface{}
}

// selectConfigs maps select operations to their configurations
var selectConfigs = map[string]selectConfig{
	"option": {
		stepType:    "PICK",
		description: "Select dropdown option by value or text",
		usage:       "select option [checkpoint-id] <selector> <value> [position]",
		examples: []string{
			`api-cli select option cp_12345 "Country dropdown" "United States" 1`,
			`api-cli select option "#country-select" "USA"  # Uses session context`,
			`api-cli select option "select[name='country']" "US"  # Select by value`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"value":    args[1],
				"method":   "value",
			}
		},
	},
	"index": {
		stepType:    "PICK",
		description: "Select dropdown option by index (0-based)",
		usage:       "select index [checkpoint-id] <selector> <index> [position]",
		examples: []string{
			`api-cli select index cp_12345 "Country dropdown" 2 1`,
			`api-cli select index "#country-select" 0  # First option, uses session context`,
			`api-cli select index ".dropdown" 3  # Fourth option (0-based index)`,
		},
		requiredArgs: 2,
		buildMeta: func(args []string) map[string]interface{} {
			index, _ := strconv.Atoi(args[1])
			return map[string]interface{}{
				"selector": args[0],
				"index":    index,
				"method":   "index",
			}
		},
	},
	"last": {
		stepType:    "PICK",
		description: "Select the last dropdown option",
		usage:       "select last [checkpoint-id] <selector> [position]",
		examples: []string{
			`api-cli select last cp_12345 "Country dropdown" 1`,
			`api-cli select last "#country-select"  # Uses session context`,
			`api-cli select last ".dropdown"  # Select last option`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"index":    -1, // Last option is represented as -1
				"method":   "last",
			}
		},
	},
}

// newSelectV2Cmd creates the select command group
func newSelectV2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select options from dropdowns",
		Long: `Select options from dropdown elements using various methods.

This command consolidates dropdown selection operations:
  - Select by value or text
  - Select by index (0-based)
  - Select last option

The command supports both modern (session-based) and legacy (explicit checkpoint) formats:
  - Modern: select <type> <selector> <value> [position]
  - Legacy: select <type> <checkpoint-id> <selector> <value> [position]`,
		Example: `  # Select option by value/text
  api-cli select option "Country dropdown" "United States"
  api-cli select option cp_12345 "#country" "USA" 1

  # Select option by index (0-based)
  api-cli select index "Country dropdown" 2
  api-cli select index cp_12345 ".country-select" 0 5

  # Select last option
  api-cli select last "#country-select"
  api-cli select last cp_12345 "select[name='country']" 3`,
	}

	// Add subcommands for each select type
	for selectType, config := range selectConfigs {
		cmd.AddCommand(newSelectSubcommand(selectType, config))
	}

	return cmd
}

// newSelectSubcommand creates a subcommand for a specific select type
func newSelectSubcommand(selectType string, config selectConfig) *cobra.Command {
	selectCmd := &SelectCommand{
		BaseCommand: NewBaseCommand(),
		selectType:  selectType,
	}

	cmd := &cobra.Command{
		Use:   selectType + " " + strings.TrimPrefix(config.usage, "select "+selectType+" "),
		Short: config.description,
		Long: fmt.Sprintf(`%s

%s

Examples:
%s`, config.description, config.usage, strings.Join(config.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			// Flexible validation - could have checkpoint ID or not
			minArgs := config.requiredArgs
			maxArgs := config.requiredArgs + 2 // +1 for checkpoint, +1 for position

			if len(args) < minArgs || len(args) > maxArgs {
				return fmt.Errorf("incorrect number of arguments")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize base command
			if err := selectCmd.Init(cmd); err != nil {
				return err
			}

			// Validate and execute
			if err := selectCmd.ValidateArgs(args, config); err != nil {
				return err
			}

			return selectCmd.Execute(args, config)
		},
	}

	// Add common flags
	cmd.Flags().StringP("output", "o", "human", "Output format (human, json, yaml, ai)")

	return cmd
}

// ValidateArgs validates the arguments for the select command
func (sc *SelectCommand) ValidateArgs(args []string, config selectConfig) error {
	// Resolve checkpoint and position
	remainingArgs, err := sc.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return err
	}

	// Validate selector
	if len(remainingArgs) < 1 || remainingArgs[0] == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	// Type-specific validation
	switch sc.selectType {
	case "option":
		if len(remainingArgs) < 2 || remainingArgs[1] == "" {
			return fmt.Errorf("value cannot be empty")
		}
	case "index":
		if len(remainingArgs) < 2 {
			return fmt.Errorf("index is required")
		}
		// Validate index is a non-negative integer
		index, err := strconv.Atoi(remainingArgs[1])
		if err != nil {
			return fmt.Errorf("index must be a valid integer: %w", err)
		}
		if index < 0 {
			return fmt.Errorf("index must be 0 or greater (got %d)", index)
		}
	case "last":
		// No additional validation needed
	}

	return nil
}

// Execute runs the select command
func (sc *SelectCommand) Execute(args []string, config selectConfig) error {
	// Resolve checkpoint and position
	remainingArgs, err := sc.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return err
	}

	// Convert checkpoint ID to int
	checkpointID, err := strconv.Atoi(sc.CheckpointID)
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	// Auto-increment position if needed
	if sc.Position == -1 && cfg != nil && cfg.Session.AutoIncrementPos {
		sc.Position = cfg.Session.NextPosition
		if sc.Position == 0 {
			sc.Position = 1
		}
	}

	// Call the appropriate API method
	var stepID int
	switch sc.selectType {
	case "option":
		stepID, err = sc.Client.CreateStepPick(checkpointID, remainingArgs[0], remainingArgs[1], sc.Position)
	case "index":
		index, _ := strconv.Atoi(remainingArgs[1])
		stepID, err = sc.Client.CreateStepPickIndex(checkpointID, remainingArgs[0], index, sc.Position)
	case "last":
		stepID, err = sc.Client.CreateStepPickLast(checkpointID, remainingArgs[0], sc.Position)
	default:
		return fmt.Errorf("unsupported select type: %s", sc.selectType)
	}

	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", config.stepType, err)
	}

	// Update session position if auto-increment is enabled
	if cfg != nil && cfg.Session.AutoIncrementPos {
		if cfg.Session.NextPosition == 0 {
			cfg.Session.NextPosition = 1
		} else {
			cfg.Session.NextPosition++
		}
		if err := cfg.SaveConfig(); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to save position update: %v\n", err)
		}
	}

	// Build result
	result := &StepResult{
		ID:          fmt.Sprintf("%d", stepID),
		Type:        config.stepType,
		Position:    sc.Position,
		Description: sc.buildDescription(remainingArgs, config),
		Selector:    remainingArgs[0],
		Meta:        config.buildMeta(remainingArgs),
	}

	// Format and output result
	output, err := sc.FormatOutput(result, sc.OutputFormat)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

// buildDescription builds a human-readable description of the step
func (sc *SelectCommand) buildDescription(args []string, config selectConfig) string {
	switch sc.selectType {
	case "option":
		return fmt.Sprintf("Select option '%s' from '%s'", args[1], args[0])
	case "index":
		return fmt.Sprintf("Select option at index %s from '%s'", args[1], args[0])
	case "last":
		return fmt.Sprintf("Select last option from '%s'", args[0])
	default:
		return fmt.Sprintf("Select from '%s'", args[0])
	}
}
