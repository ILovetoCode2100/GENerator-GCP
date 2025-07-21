package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// WindowCommand implements the window command group using BaseCommand pattern
type WindowCommand struct {
	*BaseCommand
	operation string
	subtype   string
}

// windowConfig contains configuration for each window operation
type windowConfig struct {
	stepType     string
	description  string
	usage        string
	examples     []string
	requiredArgs int
	buildMeta    func(args []string) map[string]interface{}
}

// windowConfigs maps window operations to their configurations
var windowConfigs = map[string]windowConfig{
	"resize": {
		stepType:    "RESIZE",
		description: "Resize browser window",
		usage:       "window resize [checkpoint-id] <WIDTHxHEIGHT> [position]",
		examples: []string{
			`api-cli window resize cp_12345 1024x768 1`,
			`api-cli window resize 800x600  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			parts := strings.Split(args[0], "x")
			width, _ := strconv.Atoi(parts[0])
			height, _ := strconv.Atoi(parts[1])
			return map[string]interface{}{
				"size":   args[0],
				"width":  width,
				"height": height,
			}
		},
	},
	"maximize": {
		stepType:    "WINDOW",
		description: "Maximize browser window",
		usage:       "window maximize [checkpoint-id] [position]",
		examples: []string{
			`api-cli window maximize cp_12345 1`,
			`api-cli window maximize  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"window_type": "MAXIMIZE",
			}
		},
	},
	"switch.tab.next": {
		stepType:    "SWITCH",
		description: "Switch to next browser tab",
		usage:       "window switch tab next [checkpoint-id] [position]",
		examples: []string{
			`api-cli window switch tab next cp_12345 1`,
			`api-cli window switch tab next  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"direction": "next",
				"tab_type":  "NEXT_TAB",
			}
		},
	},
	"switch.tab.prev": {
		stepType:    "SWITCH",
		description: "Switch to previous browser tab",
		usage:       "window switch tab prev [checkpoint-id] [position]",
		examples: []string{
			`api-cli window switch tab prev cp_12345 1`,
			`api-cli window switch tab prev  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"direction": "previous",
				"tab_type":  "PREVIOUS_TAB",
			}
		},
	},
	"switch.tab.index": {
		stepType:    "SWITCH",
		description: "Switch to browser tab by index (0-based)",
		usage:       "window switch tab index [checkpoint-id] <index> [position]",
		examples: []string{
			`api-cli window switch tab index cp_12345 0 1  # Switch to first tab`,
			`api-cli window switch tab index 2  # Switch to third tab, uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			index, _ := strconv.Atoi(args[0])
			return map[string]interface{}{
				"index":    index,
				"tab_type": "TAB_BY_INDEX",
			}
		},
	},
	"switch.iframe": {
		stepType:    "SWITCH",
		description: "Switch to iframe by element selector",
		usage:       "window switch iframe [checkpoint-id] <selector> [position]",
		examples: []string{
			`api-cli window switch iframe cp_12345 "#payment-frame" 1`,
			`api-cli window switch iframe "iframe[name='content']"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector":   args[0],
				"frame_type": "FRAME_BY_ELEMENT",
			}
		},
	},
	"switch.parent-frame": {
		stepType:    "SWITCH",
		description: "Switch to parent frame",
		usage:       "window switch parent-frame [checkpoint-id] [position]",
		examples: []string{
			`api-cli window switch parent-frame cp_12345 1`,
			`api-cli window switch parent-frame  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"frame_type": "PARENT_FRAME",
			}
		},
	},
}

// newWindowV2Cmd creates the new window command using BaseCommand pattern
func newWindowV2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "window",
		Short: "Manage browser windows, tabs, and frames",
		Long: `Perform various window operations including tab switching, frame navigation, and window resizing.

This command uses the standardized positional argument pattern:
- Optional checkpoint ID as first argument (falls back to session context)
- Required operation arguments
- Optional position as last argument (auto-increments if not specified)

Available operations:
  - resize: Resize browser window to specific dimensions
  - maximize: Maximize browser window
  - switch tab next/prev/index: Navigate between browser tabs
  - switch iframe: Switch context to an iframe
  - switch parent-frame: Switch back to parent frame`,
		Example: `  # Resize window (with explicit checkpoint)
  api-cli window resize cp_12345 1024x768 1

  # Resize window (using session context)
  api-cli window resize 800x600

  # Maximize window
  api-cli window maximize

  # Switch to next tab
  api-cli window switch tab next

  # Switch to iframe
  api-cli window switch iframe "#payment-frame"

  # Switch to tab by index
  api-cli window switch tab index 2`,
	}

	// Add resize subcommand
	cmd.AddCommand(newWindowV2SubCmd("resize", windowConfigs["resize"]))

	// Add maximize subcommand
	cmd.AddCommand(newWindowV2SubCmd("maximize", windowConfigs["maximize"]))

	// Add switch subcommand with nested subcommands
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch between frames and tabs",
		Long:  "Switch context to different frames or browser tabs",
	}

	// Add tab subcommand with its own subcommands
	tabCmd := &cobra.Command{
		Use:   "tab",
		Short: "Switch between browser tabs",
		Long:  "Navigate to next or previous browser tab, or switch to specific tab by index",
	}
	tabCmd.AddCommand(newWindowV2SubCmd("next", windowConfigs["switch.tab.next"]))
	tabCmd.AddCommand(newWindowV2SubCmd("prev", windowConfigs["switch.tab.prev"]))
	tabCmd.AddCommand(newWindowV2SubCmd("index", windowConfigs["switch.tab.index"]))
	switchCmd.AddCommand(tabCmd)

	// Add iframe subcommand
	switchCmd.AddCommand(newWindowV2SubCmd("iframe", windowConfigs["switch.iframe"]))

	// Add parent-frame subcommand
	switchCmd.AddCommand(newWindowV2SubCmd("parent-frame", windowConfigs["switch.parent-frame"]))

	cmd.AddCommand(switchCmd)

	return cmd
}

// newWindowV2SubCmd creates a subcommand for a specific window operation
func newWindowV2SubCmd(operation string, config windowConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   operation + " " + extractWindowUsageArgs(config.usage),
		Short: config.description,
		Long: fmt.Sprintf(`%s

%s

Examples:
%s`, config.description, config.usage, strings.Join(config.examples, "\n")),
		RunE: func(cmd *cobra.Command, args []string) error {
			wc := &WindowCommand{
				BaseCommand: NewBaseCommand(),
				operation:   operation,
			}
			// Determine the full operation key for nested commands
			fullOp := operation
			if parentCmd := cmd.Parent(); parentCmd != nil {
				if parentCmd.Name() == "tab" {
					fullOp = "switch.tab." + operation
				} else if parentCmd.Name() == "switch" && operation == "iframe" {
					fullOp = "switch.iframe"
				} else if parentCmd.Name() == "switch" && operation == "parent-frame" {
					fullOp = "switch.parent-frame"
				}
			}
			return wc.Execute(cmd, args, windowConfigs[fullOp])
		},
	}

	// Validate args based on operation
	if config.requiredArgs > 0 {
		cmd.Args = cobra.RangeArgs(config.requiredArgs, config.requiredArgs+2) // +2 for optional checkpoint and position
	} else {
		cmd.Args = cobra.MaximumNArgs(2) // Only optional checkpoint and position
	}

	return cmd
}

// extractWindowUsageArgs extracts the arguments portion from the usage string
func extractWindowUsageArgs(usage string) string {
	parts := strings.Fields(usage)
	// Find the start of arguments (after "window" and operation keywords)
	startIdx := 0
	for i, part := range parts {
		if strings.HasPrefix(part, "[checkpoint-id]") || strings.HasPrefix(part, "<") {
			startIdx = i
			break
		}
	}
	if startIdx > 0 {
		return strings.Join(parts[startIdx:], " ")
	}
	return ""
}

// Execute runs the window command
func (wc *WindowCommand) Execute(cmd *cobra.Command, args []string, config windowConfig) error {
	// Initialize base command
	if err := wc.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	remainingArgs, err := wc.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate we have the required number of arguments
	if len(remainingArgs) != config.requiredArgs {
		return fmt.Errorf("expected %d arguments, got %d", config.requiredArgs, len(remainingArgs))
	}

	// Additional validation for specific operations
	if err := wc.validateWindowArgs(wc.operation, remainingArgs); err != nil {
		return err
	}

	// Build request metadata
	meta := config.buildMeta(remainingArgs)

	// Create the step
	stepResult, err := wc.createWindowStep(config.stepType, meta, remainingArgs)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", config.stepType, err)
	}

	// Format and output the result
	output, err := wc.FormatOutput(stepResult, wc.OutputFormat)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// validateWindowArgs validates arguments for specific window operations
func (wc *WindowCommand) validateWindowArgs(operation string, args []string) error {
	switch operation {
	case "resize":
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
	case "index":
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("index cannot be empty")
		}
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("index must be a number")
		}
	case "iframe":
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
	}
	return nil
}

// createWindowStep creates a window step via the API
func (wc *WindowCommand) createWindowStep(stepType string, meta map[string]interface{}, args []string) (*StepResult, error) {
	// Convert checkpoint ID from string to int
	checkpointID, err := strconv.Atoi(wc.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint ID: %s", wc.CheckpointID)
	}

	// Call the appropriate API method based on meta information
	var stepID int

	// Determine the specific operation from meta
	switch {
	case stepType == "RESIZE":
		width := meta["width"].(int)
		height := meta["height"].(int)
		stepID, err = wc.Client.CreateStepWindowResize(checkpointID, width, height, wc.Position)

	case stepType == "WINDOW" && meta["window_type"] == "MAXIMIZE":
		stepID, err = wc.Client.CreateStepWindowMaximize(checkpointID, wc.Position)

	case stepType == "SWITCH" && meta["tab_type"] == "NEXT_TAB":
		stepID, err = wc.Client.CreateStepSwitchNextTab(checkpointID, wc.Position)

	case stepType == "SWITCH" && meta["tab_type"] == "PREVIOUS_TAB":
		stepID, err = wc.Client.CreateStepSwitchPrevTab(checkpointID, wc.Position)

	case stepType == "SWITCH" && meta["tab_type"] == "TAB_BY_INDEX":
		index := meta["index"].(int)
		stepID, err = wc.Client.CreateStepSwitchTabByIndex(checkpointID, index, wc.Position)

	case stepType == "SWITCH" && meta["frame_type"] == "FRAME_BY_ELEMENT":
		selector := meta["selector"].(string)
		stepID, err = wc.Client.CreateStepSwitchIframe(checkpointID, selector, wc.Position)

	case stepType == "SWITCH" && meta["frame_type"] == "PARENT_FRAME":
		stepID, err = wc.Client.CreateStepSwitchParentFrame(checkpointID, wc.Position)

	default:
		return nil, fmt.Errorf("unknown window operation: %s with meta %v", stepType, meta)
	}

	if err != nil {
		return nil, err
	}

	// Build the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: wc.CheckpointID,
		Type:         stepType,
		Position:     wc.Position,
		Description:  wc.buildDescription(stepType, meta, args),
		Selector:     wc.extractSelector(meta),
		Meta:         meta,
	}

	// Save session state if position was auto-incremented
	if wc.Position == -1 && cfg.Session.AutoIncrementPos {
		if err := cfg.SaveConfig(); err != nil {
			// Don't fail the command, just warn
			// Note: In production, this warning would be sent to stderr
		}
	}

	return result, nil
}

// buildDescription builds a human-readable description for the step
func (wc *WindowCommand) buildDescription(stepType string, meta map[string]interface{}, args []string) string {
	switch {
	case stepType == "RESIZE":
		return fmt.Sprintf("resize window to %s", meta["size"])
	case stepType == "WINDOW" && meta["window_type"] == "MAXIMIZE":
		return "maximize window"
	case stepType == "SWITCH" && meta["tab_type"] == "NEXT_TAB":
		return "switch to next tab"
	case stepType == "SWITCH" && meta["tab_type"] == "PREVIOUS_TAB":
		return "switch to previous tab"
	case stepType == "SWITCH" && meta["tab_type"] == "TAB_BY_INDEX":
		return fmt.Sprintf("switch to tab index %d", meta["index"])
	case stepType == "SWITCH" && meta["frame_type"] == "FRAME_BY_ELEMENT":
		return fmt.Sprintf("switch to iframe \"%s\"", meta["selector"])
	case stepType == "SWITCH" && meta["frame_type"] == "PARENT_FRAME":
		return "switch to parent frame"
	default:
		return fmt.Sprintf("%s operation", stepType)
	}
}

// extractSelector extracts the selector from metadata if present
func (wc *WindowCommand) extractSelector(meta map[string]interface{}) string {
	if selector, ok := meta["selector"].(string); ok {
		return selector
	}
	return ""
}
