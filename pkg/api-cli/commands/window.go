package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// windowOperation represents the type of window operation
type windowOperation string

const (
	windowSwitchIframe      windowOperation = "switch-iframe"
	windowSwitchTabNext     windowOperation = "switch-tab-next"
	windowSwitchTabPrev     windowOperation = "switch-tab-prev"
	windowSwitchTabByIndex  windowOperation = "switch-tab-index"
	windowSwitchParentFrame windowOperation = "switch-parent-frame"
	windowResize            windowOperation = "resize"
	windowMaximize          windowOperation = "maximize"
)

// windowCommandInfo contains metadata about each window operation
type windowCommandInfo struct {
	stepType    string
	description string
	usage       string
	examples    []string
	argsCount   []int // Valid argument counts (excluding position)
	parseStep   func(args []string) string
}

// windowCommands maps window operations to their metadata
var windowCommands = map[windowOperation]windowCommandInfo{
	windowSwitchIframe: {
		stepType:    "SWITCH",
		description: "Switch to iframe by element selector",
		usage:       "window switch iframe SELECTOR [POSITION]",
		examples: []string{
			`api-cli window switch iframe "#payment-frame" 1`,
			`api-cli window switch iframe "iframe[name='content']"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("switch to iframe \"%s\"", args[0])
		},
	},
	windowSwitchTabNext: {
		stepType:    "SWITCH",
		description: "Switch to next browser tab",
		usage:       "window switch tab next [POSITION]",
		examples: []string{
			`api-cli window switch tab next 1`,
			`api-cli window switch tab next  # Auto-increment position`,
		},
		argsCount: []int{0},
		parseStep: func(args []string) string {
			return "switch to next tab"
		},
	},
	windowSwitchTabPrev: {
		stepType:    "SWITCH",
		description: "Switch to previous browser tab",
		usage:       "window switch tab prev [POSITION]",
		examples: []string{
			`api-cli window switch tab prev 1`,
			`api-cli window switch tab prev  # Auto-increment position`,
		},
		argsCount: []int{0},
		parseStep: func(args []string) string {
			return "switch to previous tab"
		},
	},
	windowSwitchParentFrame: {
		stepType:    "SWITCH",
		description: "Switch to parent frame",
		usage:       "window switch parent-frame [POSITION]",
		examples: []string{
			`api-cli window switch parent-frame 1`,
			`api-cli window switch parent-frame  # Auto-increment position`,
		},
		argsCount: []int{0},
		parseStep: func(args []string) string {
			return "switch to parent frame"
		},
	},
	windowSwitchTabByIndex: {
		stepType:    "SWITCH",
		description: "Switch to browser tab by index (0-based)",
		usage:       "window switch tab INDEX [POSITION]",
		examples: []string{
			`api-cli window switch tab 0 1  # Switch to first tab`,
			`api-cli window switch tab 2    # Switch to third tab, auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("switch to tab index %s", args[0])
		},
	},
	windowResize: {
		stepType:    "RESIZE",
		description: "Resize browser window",
		usage:       "window resize WIDTHxHEIGHT [POSITION]",
		examples: []string{
			`api-cli window resize 1024x768 1`,
			`api-cli window resize 800x600  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("resize window to %s", args[0])
		},
	},
	windowMaximize: {
		stepType:    "WINDOW",
		description: "Maximize browser window",
		usage:       "window maximize [POSITION]",
		examples: []string{
			`api-cli window maximize 1`,
			`api-cli window maximize  # Auto-increment position`,
		},
		argsCount: []int{0},
		parseStep: func(args []string) string {
			return "maximize window"
		},
	},
}

// newWindowCmd creates the consolidated window command with subcommands
func newWindowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "window",
		Short: "Manage browser windows, tabs, and frames",
		Long: `Perform various window operations including tab switching, frame navigation, and window resizing.

This command consolidates all window-related operations:
  - Switch between iframes and parent frames
  - Navigate between browser tabs
  - Resize the browser window`,
		Example: `  # Switch to iframe
  api-cli window switch iframe "#payment-frame" 1

  # Switch to next tab
  api-cli window switch tab next

  # Switch to parent frame
  api-cli window switch parent-frame

  # Resize window
  api-cli window resize 1024x768`,
	}

	// Add switch subcommand
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch between frames and tabs",
		Long:  "Switch context to different frames or browser tabs",
	}

	// Add iframe switch
	switchCmd.AddCommand(newWindowSwitchSubCmd("iframe", windowSwitchIframe, windowCommands[windowSwitchIframe]))

	// Add tab switches
	tabCmd := &cobra.Command{
		Use:   "tab",
		Short: "Switch between browser tabs",
		Long:  "Navigate to next or previous browser tab, or switch to specific tab by index",
	}
	tabCmd.AddCommand(newWindowSwitchSubCmd("next", windowSwitchTabNext, windowCommands[windowSwitchTabNext]))
	tabCmd.AddCommand(newWindowSwitchSubCmd("prev", windowSwitchTabPrev, windowCommands[windowSwitchTabPrev]))
	tabCmd.AddCommand(newWindowSwitchTabIndexCmd()) // Add tab index as a subcommand of tab
	switchCmd.AddCommand(tabCmd)

	// Add parent frame switch
	switchCmd.AddCommand(newWindowSwitchSubCmd("parent-frame", windowSwitchParentFrame, windowCommands[windowSwitchParentFrame]))

	cmd.AddCommand(switchCmd)

	// Add resize subcommand
	cmd.AddCommand(newWindowResizeCmd())

	// Add maximize subcommand
	cmd.AddCommand(newWindowMaximizeCmd())

	return cmd
}

// newWindowSwitchSubCmd creates a subcommand for window switch operations
func newWindowSwitchSubCmd(name string, op windowOperation, info windowCommandInfo) *cobra.Command {
	var checkpointFlag int

	// Extract args from usage
	usageParts := strings.Fields(info.usage)
	argsUsage := ""
	if len(usageParts) > 3 {
		argsUsage = strings.Join(usageParts[3:], " ")
	}

	cmd := &cobra.Command{
		Use:   name + " " + argsUsage,
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
			return runWindowCommand(op, info, args, checkpointFlag)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// newWindowSwitchTabIndexCmd creates the window switch tab by index subcommand
func newWindowSwitchTabIndexCmd() *cobra.Command {
	var checkpointFlag int
	info := windowCommands[windowSwitchTabByIndex]

	cmd := &cobra.Command{
		Use:     "INDEX",
		Short:   info.description,
		Aliases: []string{},
		Long: fmt.Sprintf(`%s

The index is 0-based (0 for first tab, 1 for second, etc.).

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
%s`, info.description, strings.Join(info.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 2 {
				return fmt.Errorf("accepts 1 or 2 args, received %d", len(args))
			}
			// Validate index is a number
			if _, err := strconv.Atoi(args[0]); err != nil {
				return fmt.Errorf("index must be a number")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWindowCommand(windowSwitchTabByIndex, info, args, checkpointFlag)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// newWindowResizeCmd creates the window resize subcommand
func newWindowResizeCmd() *cobra.Command {
	var checkpointFlag int
	info := windowCommands[windowResize]

	cmd := &cobra.Command{
		Use:   "resize WIDTHxHEIGHT [POSITION]",
		Short: info.description,
		Long: fmt.Sprintf(`%s

The size should be specified as WIDTHxHEIGHT (e.g., 1024x768).

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
%s`, info.description, strings.Join(info.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 || len(args) > 2 {
				return fmt.Errorf("accepts 1 or 2 args, received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWindowCommand(windowResize, info, args, checkpointFlag)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// newWindowMaximizeCmd creates the window maximize subcommand
func newWindowMaximizeCmd() *cobra.Command {
	var checkpointFlag int
	info := windowCommands[windowMaximize]

	cmd := &cobra.Command{
		Use:   "maximize [POSITION]",
		Short: info.description,
		Long: fmt.Sprintf(`%s

Uses the current checkpoint from session context by default. Override with --checkpoint flag.
Position is auto-incremented if not specified and auto-increment is enabled.

Examples:
%s`, info.description, strings.Join(info.examples, "\n")),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("accepts 0 or 1 args, received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWindowCommand(windowMaximize, info, args, checkpointFlag)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// runWindowCommand executes the window command logic
func runWindowCommand(op windowOperation, info windowCommandInfo, args []string, checkpointFlag int) error {
	// Validate arguments based on operation type
	if err := validateWindowArgs(op, args); err != nil {
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

	// Call the appropriate API method based on operation type
	stepID, err := callWindowAPI(apiClient, op, ctx, args)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", info.stepType, err)
	}

	// Save config if position was auto-incremented
	saveStepContext(ctx)

	// Build extra data for output
	extra := buildWindowExtraData(op, args)

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

// validateWindowArgs validates arguments for a specific window operation
func validateWindowArgs(op windowOperation, args []string) error {
	switch op {
	case windowSwitchIframe:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
	case windowResize:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("size cannot be empty")
		}
		// Validate size format
		if !strings.Contains(args[0], "x") {
			return fmt.Errorf("size must be in format WIDTHxHEIGHT (e.g., 1024x768)")
		}
		parts := strings.Split(args[0], "x")
		if len(parts) != 2 {
			return fmt.Errorf("size must be in format WIDTHxHEIGHT (e.g., 1024x768)")
		}
		if _, err := strconv.Atoi(parts[0]); err != nil {
			return fmt.Errorf("width must be a number")
		}
		if _, err := strconv.Atoi(parts[1]); err != nil {
			return fmt.Errorf("height must be a number")
		}
	case windowSwitchTabByIndex:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("index cannot be empty")
		}
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("index must be a number")
		}
	case windowSwitchTabNext, windowSwitchTabPrev, windowSwitchParentFrame, windowMaximize:
		// No arguments needed
	}
	return nil
}

// callWindowAPI calls the appropriate client API method for the window operation
func callWindowAPI(apiClient *client.Client, op windowOperation, ctx *StepContext, args []string) (int, error) {
	switch op {
	case windowSwitchIframe:
		return apiClient.CreateStepSwitchIframe(ctx.CheckpointID, args[0], ctx.Position)
	case windowSwitchTabNext:
		return apiClient.CreateStepSwitchNextTab(ctx.CheckpointID, ctx.Position)
	case windowSwitchTabPrev:
		return apiClient.CreateStepSwitchPrevTab(ctx.CheckpointID, ctx.Position)
	case windowSwitchParentFrame:
		return apiClient.CreateStepSwitchParentFrame(ctx.CheckpointID, ctx.Position)
	case windowSwitchTabByIndex:
		index, _ := strconv.Atoi(args[0])
		return apiClient.CreateStepSwitchTabByIndex(ctx.CheckpointID, index, ctx.Position)
	case windowResize:
		// Parse width and height
		parts := strings.Split(args[0], "x")
		width, _ := strconv.Atoi(parts[0])
		height, _ := strconv.Atoi(parts[1])
		return apiClient.CreateStepWindowResize(ctx.CheckpointID, width, height, ctx.Position)
	case windowMaximize:
		return apiClient.CreateStepWindowMaximize(ctx.CheckpointID, ctx.Position)
	default:
		return 0, fmt.Errorf("unsupported window operation: %s", op)
	}
}

// buildWindowExtraData builds the extra data map for output based on window operation
func buildWindowExtraData(op windowOperation, args []string) map[string]interface{} {
	extra := make(map[string]interface{})

	switch op {
	case windowSwitchIframe:
		extra["selector"] = args[0]
		extra["frame_type"] = "FRAME_BY_ELEMENT"
	case windowSwitchTabNext:
		extra["direction"] = "next"
		extra["tab_type"] = "NEXT_TAB"
	case windowSwitchTabPrev:
		extra["direction"] = "previous"
		extra["tab_type"] = "PREVIOUS_TAB"
	case windowSwitchParentFrame:
		extra["frame_type"] = "PARENT_FRAME"
	case windowSwitchTabByIndex:
		index, _ := strconv.Atoi(args[0])
		extra["index"] = index
		extra["tab_type"] = "TAB_BY_INDEX"
	case windowResize:
		parts := strings.Split(args[0], "x")
		width, _ := strconv.Atoi(parts[0])
		height, _ := strconv.Atoi(parts[1])
		extra["size"] = args[0]
		extra["width"] = width
		extra["height"] = height
	case windowMaximize:
		extra["window_type"] = "MAXIMIZE"
	}

	return extra
}
