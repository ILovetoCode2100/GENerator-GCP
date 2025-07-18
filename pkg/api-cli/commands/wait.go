package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// waitType represents the type of wait operation
type waitType string

const (
	waitElement           waitType = "element"
	waitElementNotVisible waitType = "element-not-visible"
	waitTime              waitType = "time"
)

// waitCommandInfo contains metadata about each wait type
type waitCommandInfo struct {
	stepType    string
	description string
	usage       string
	examples    []string
	argsCount   []int // Valid argument counts (excluding position)
	parseStep   func(args []string, timeout int) string
	hasTimeout  bool
}

// waitCommands maps wait types to their metadata
var waitCommands = map[waitType]waitCommandInfo{
	waitElement: {
		stepType:    "WAIT_ELEMENT",
		description: "Wait for an element to appear",
		usage:       "wait element SELECTOR [POSITION]",
		examples: []string{
			`api-cli wait element "Login button" 1`,
			`api-cli wait element "Success message"  # Auto-increment position`,
			`api-cli wait element "Loading" --timeout 5000  # Wait up to 5 seconds`,
			`api-cli wait element "#submit-btn" 2 --timeout 10000  # 10 second timeout`,
		},
		argsCount: []int{1},
		parseStep: func(args []string, timeout int) string {
			if timeout > 0 {
				return fmt.Sprintf("wait until %s appears (timeout: %dms)", args[0], timeout)
			}
			return fmt.Sprintf("wait until %s appears", args[0])
		},
		hasTimeout: true,
	},
	waitElementNotVisible: {
		stepType:    "WAIT_ELEMENT_NOT_VISIBLE",
		description: "Wait for an element to disappear",
		usage:       "wait element-not-visible SELECTOR [POSITION]",
		examples: []string{
			`api-cli wait element-not-visible "Loading spinner" 1`,
			`api-cli wait element-not-visible "#loader"  # Auto-increment position`,
			`api-cli wait element-not-visible "Modal overlay" --timeout 5000`,
			`api-cli wait element-not-visible ".progress-bar" 2 --timeout 10000`,
		},
		argsCount: []int{1},
		parseStep: func(args []string, timeout int) string {
			if timeout > 0 {
				return fmt.Sprintf("wait until %s disappears (timeout: %dms)", args[0], timeout)
			}
			return fmt.Sprintf("wait until %s disappears", args[0])
		},
		hasTimeout: true,
	},
	waitTime: {
		stepType:    "WAIT_TIME",
		description: "Wait for a specified time in milliseconds",
		usage:       "wait time MILLISECONDS [POSITION]",
		examples: []string{
			`api-cli wait time 1000 1  # Wait 1 second at position 1`,
			`api-cli wait time 500  # Wait 500ms, auto-increment position`,
			`api-cli wait time 3000 2  # Wait 3 seconds at position 2`,
		},
		argsCount: []int{1},
		parseStep: func(args []string, timeout int) string {
			ms, _ := strconv.Atoi(args[0])
			if ms >= 1000 {
				seconds := float64(ms) / 1000.0
				return fmt.Sprintf("wait %.1f seconds", seconds)
			}
			return fmt.Sprintf("wait %d milliseconds", ms)
		},
		hasTimeout: false,
	},
}

// newWaitCmd creates the consolidated wait command with subcommands
func newWaitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Create wait steps in checkpoints",
		Long: `Create various types of wait steps in checkpoints.

This command consolidates all wait operations for controlling test execution timing.

Available wait types:
  - element: Wait for an element to appear (with optional timeout)
  - element-not-visible: Wait for an element to disappear (with optional timeout)
  - time: Wait for a specified time in milliseconds`,
		Example: `  # Wait for element to appear
  api-cli wait element "Login button" 1

  # Wait for element with custom timeout
  api-cli wait element "Success message" --timeout 5000

  # Wait for 2 seconds
  api-cli wait time 2000`,
	}

	// Add subcommands for each wait type
	for wType, info := range waitCommands {
		cmd.AddCommand(newWaitSubCmd(wType, info))
	}

	return cmd
}

// extractWaitArgsFromUsage extracts the arguments part from the usage string
func extractWaitArgsFromUsage(usage string) string {
	parts := strings.Fields(usage)
	if len(parts) > 2 {
		return strings.Join(parts[2:], " ")
	}
	return ""
}

// newWaitSubCmd creates a subcommand for a specific wait type
func newWaitSubCmd(wType waitType, info waitCommandInfo) *cobra.Command {
	var checkpointFlag int
	var timeoutFlag int

	cmd := &cobra.Command{
		Use:   string(wType) + " " + extractWaitArgsFromUsage(info.usage),
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
			return runWaitCommand(wType, info, args, checkpointFlag, timeoutFlag)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	// Add timeout flag for commands that support it
	if info.hasTimeout {
		cmd.Flags().IntVar(&timeoutFlag, "timeout", 0, "Custom timeout in milliseconds (default uses Virtuoso's default)")
	}

	return cmd
}

// runWaitCommand executes the wait command logic
func runWaitCommand(wType waitType, info waitCommandInfo, args []string, checkpointFlag int, timeoutFlag int) error {
	// Validate arguments based on wait type
	if err := validateWaitArgs(wType, args, timeoutFlag); err != nil {
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

	// Call the appropriate API method based on wait type
	stepID, err := callWaitAPI(apiClient, wType, ctx, args, timeoutFlag)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", info.stepType, err)
	}

	// Save config if position was auto-incremented
	saveStepContext(ctx)

	// Build extra data for output
	extra := buildWaitExtraData(wType, args, timeoutFlag)

	// Output result
	output := &StepOutput{
		Status:       "success",
		StepType:     info.stepType,
		CheckpointID: ctx.CheckpointID,
		StepID:       stepID,
		Position:     ctx.Position,
		ParsedStep:   info.parseStep(args, timeoutFlag),
		UsingContext: ctx.UsingContext,
		AutoPosition: ctx.AutoPosition,
		Extra:        extra,
	}

	return outputStepResult(output)
}

// validateWaitArgs validates arguments for a specific wait type
func validateWaitArgs(wType waitType, args []string, timeout int) error {
	switch wType {
	case waitElement, waitElementNotVisible:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
		if timeout < 0 {
			return fmt.Errorf("timeout cannot be negative")
		}
		if timeout > 60000 {
			return fmt.Errorf("timeout cannot exceed 60000ms (60 seconds)")
		}
	case waitTime:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("time cannot be empty")
		}
		ms, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid time value: %w", err)
		}
		if ms < 0 {
			return fmt.Errorf("time cannot be negative")
		}
		if ms > 60000 {
			return fmt.Errorf("time cannot exceed 60000ms (60 seconds)")
		}
	}
	return nil
}

// callWaitAPI calls the appropriate client API method for the wait type
func callWaitAPI(apiClient *client.Client, wType waitType, ctx *StepContext, args []string, timeout int) (int, error) {
	switch wType {
	case waitElement:
		if timeout > 0 {
			// Use custom timeout version
			return apiClient.CreateStepWaitForElementTimeout(ctx.CheckpointID, args[0], timeout, ctx.Position)
		} else {
			// Use default timeout version (or the standard wait element)
			return apiClient.CreateWaitElementStep(ctx.CheckpointID, args[0], ctx.Position)
		}
	case waitElementNotVisible:
		return apiClient.CreateStepWaitForElementNotVisible(ctx.CheckpointID, args[0], timeout, ctx.Position)
	case waitTime:
		// Convert milliseconds to seconds for the API (if it expects seconds)
		ms, _ := strconv.Atoi(args[0])
		seconds := ms / 1000
		if ms%1000 != 0 {
			// If there are remaining milliseconds, round up to next second
			seconds++
		}
		return apiClient.CreateWaitTimeStep(ctx.CheckpointID, seconds, ctx.Position)
	default:
		return 0, fmt.Errorf("unsupported wait type: %s", wType)
	}
}

// buildWaitExtraData builds the extra data map for output based on wait type
func buildWaitExtraData(wType waitType, args []string, timeout int) map[string]interface{} {
	extra := make(map[string]interface{})

	switch wType {
	case waitElement, waitElementNotVisible:
		extra["selector"] = args[0]
		if timeout > 0 {
			extra["timeout_ms"] = timeout
		}
	case waitTime:
		ms, _ := strconv.Atoi(args[0])
		extra["milliseconds"] = ms
		extra["seconds"] = float64(ms) / 1000.0
	}

	return extra
}
