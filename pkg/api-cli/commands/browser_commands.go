package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// BrowserCommand implements browser-related commands (navigate and window) using BaseCommand pattern
type BrowserCommand struct {
	*BaseCommand
	commandType string // "navigate" or "window"
	operation   string
}

// browserConfig contains configuration for each browser operation
type browserConfig struct {
	commandType  string // "navigate" or "window"
	stepType     string
	description  string
	usage        string
	examples     []string
	requiredArgs int
	buildMeta    func(args []string, flags map[string]interface{}) map[string]interface{}
	flags        []browserFlagConfig
}

// browserFlagConfig defines a command flag for browser commands
type browserFlagConfig struct {
	name         string
	shorthand    string
	defaultValue interface{}
	description  string
}

// browserConfigs maps browser operations to their configurations
var browserConfigs = map[string]browserConfig{
	// Navigate commands
	"navigate.to": {
		commandType: "navigate",
		stepType:    "NAVIGATE",
		description: "Navigate to a URL",
		usage:       "navigate to [checkpoint-id] <url> [position]",
		examples: []string{
			`api-cli navigate to "https://example.com"`,
			`api-cli navigate to "https://example.com/page" --new-tab`,
			`api-cli navigate to cp_12345 "https://example.com" 1`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"url":         args[0],
				"newTab":      flags["new-tab"],
				"waitForLoad": flags["wait"],
			}
		},
		flags: []browserFlagConfig{
			{name: "new-tab", shorthand: "", defaultValue: false, description: "Open URL in new tab"},
			{name: "wait", shorthand: "", defaultValue: true, description: "Wait for page to load"},
		},
	},
	"navigate.scroll-top": {
		commandType: "navigate",
		stepType:    "SCROLL_TOP",
		description: "Scroll to the top of the page",
		usage:       "navigate scroll-top [checkpoint-id] [position]",
		examples: []string{
			`api-cli navigate scroll-top`,
			`api-cli navigate scroll-top --smooth`,
			`api-cli navigate scroll-top cp_12345 1`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"smooth": flags["smooth"],
			}
		},
		flags: []browserFlagConfig{
			{name: "smooth", shorthand: "", defaultValue: false, description: "Use smooth scrolling"},
		},
	},
	"navigate.scroll-bottom": {
		commandType: "navigate",
		stepType:    "SCROLL_BOTTOM",
		description: "Scroll to the bottom of the page",
		usage:       "navigate scroll-bottom [checkpoint-id] [position]",
		examples: []string{
			`api-cli navigate scroll-bottom`,
			`api-cli navigate scroll-bottom --smooth`,
			`api-cli navigate scroll-bottom cp_12345 1`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"smooth": flags["smooth"],
			}
		},
		flags: []browserFlagConfig{
			{name: "smooth", shorthand: "", defaultValue: false, description: "Use smooth scrolling"},
		},
	},
	"navigate.scroll-element": {
		commandType: "navigate",
		stepType:    "SCROLL_ELEMENT",
		description: "Scroll an element into view",
		usage:       "navigate scroll-element [checkpoint-id] <selector> [position]",
		examples: []string{
			`api-cli navigate scroll-element "#target-section"`,
			`api-cli navigate scroll-element ".important-content" --smooth`,
			`api-cli navigate scroll-element "#form-submit" --block center`,
			`api-cli navigate scroll-element cp_12345 "#target-section" 1`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"selector":     args[0],
				"smooth":       flags["smooth"],
				"scrollToView": flags["into-view"],
				"block":        flags["block"],
				"inline":       flags["inline"],
			}
		},
		flags: []browserFlagConfig{
			{name: "smooth", shorthand: "", defaultValue: false, description: "Use smooth scrolling"},
			{name: "into-view", shorthand: "", defaultValue: true, description: "Scroll element into view"},
			{name: "block", shorthand: "", defaultValue: "start", description: "Vertical alignment (start, center, end, nearest)"},
			{name: "inline", shorthand: "", defaultValue: "nearest", description: "Horizontal alignment (start, center, end, nearest)"},
		},
	},
	"navigate.scroll-position": {
		commandType: "navigate",
		stepType:    "SCROLL_POSITION",
		description: "Scroll to specific coordinates",
		usage:       "navigate scroll-position [checkpoint-id] <x,y> [position]",
		examples: []string{
			`api-cli navigate scroll-position "0,500"`,
			`api-cli navigate scroll-position "100,1000" --smooth`,
			`api-cli navigate scroll-position --x 0 --y 500`,
			`api-cli navigate scroll-position cp_12345 "0,500" 1`,
		},
		requiredArgs: 0, // Can use flags or positional arg
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"x":      flags["x"],
				"y":      flags["y"],
				"smooth": flags["smooth"],
			}
		},
		flags: []browserFlagConfig{
			{name: "x", shorthand: "", defaultValue: 0, description: "X coordinate"},
			{name: "y", shorthand: "", defaultValue: 0, description: "Y coordinate"},
			{name: "smooth", shorthand: "", defaultValue: false, description: "Use smooth scrolling"},
		},
	},
	"navigate.scroll-by": {
		commandType: "navigate",
		stepType:    "SCROLL_BY",
		description: "Scroll by relative offset",
		usage:       "navigate scroll-by [checkpoint-id] <x,y> [position]",
		examples: []string{
			`api-cli navigate scroll-by "0,500"    # Scroll down 500px`,
			`api-cli navigate scroll-by "-100,0"  # Scroll left 100px`,
			`api-cli navigate scroll-by --x 0 --y -500  # Scroll up 500px`,
			`api-cli navigate scroll-by cp_12345 "0,500" 1`,
		},
		requiredArgs: 0, // Can use flags or positional arg
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"x":      flags["x"],
				"y":      flags["y"],
				"smooth": flags["smooth"],
			}
		},
		flags: []browserFlagConfig{
			{name: "x", shorthand: "", defaultValue: 0, description: "X offset (negative for left)"},
			{name: "y", shorthand: "", defaultValue: 0, description: "Y offset (negative for up)"},
			{name: "smooth", shorthand: "", defaultValue: false, description: "Use smooth scrolling"},
		},
	},
	"navigate.scroll-up": {
		commandType: "navigate",
		stepType:    "SCROLL_UP",
		description: "Scroll up by one viewport height",
		usage:       "navigate scroll-up [checkpoint-id] [position]",
		examples: []string{
			`api-cli navigate scroll-up`,
			`api-cli navigate scroll-up --smooth`,
			`api-cli navigate scroll-up cp_12345 1`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"smooth": flags["smooth"],
			}
		},
		flags: []browserFlagConfig{
			{name: "smooth", shorthand: "", defaultValue: false, description: "Use smooth scrolling"},
		},
	},
	"navigate.scroll-down": {
		commandType: "navigate",
		stepType:    "SCROLL_DOWN",
		description: "Scroll down by one viewport height",
		usage:       "navigate scroll-down [checkpoint-id] [position]",
		examples: []string{
			`api-cli navigate scroll-down`,
			`api-cli navigate scroll-down --smooth`,
			`api-cli navigate scroll-down cp_12345 1`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"smooth": flags["smooth"],
			}
		},
		flags: []browserFlagConfig{
			{name: "smooth", shorthand: "", defaultValue: false, description: "Use smooth scrolling"},
		},
	},

	// Window commands
	"window.resize": {
		commandType: "window",
		stepType:    "RESIZE",
		description: "Resize browser window",
		usage:       "window resize [checkpoint-id] <WIDTHxHEIGHT> [position]",
		examples: []string{
			`api-cli window resize cp_12345 1024x768 1`,
			`api-cli window resize 800x600  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			parts := strings.Split(args[0], "x")
			width, _ := strconv.Atoi(parts[0])
			height, _ := strconv.Atoi(parts[1])
			return map[string]interface{}{
				"size":   args[0],
				"width":  width,
				"height": height,
			}
		},
		flags: []browserFlagConfig{},
	},
	"window.maximize": {
		commandType: "window",
		stepType:    "WINDOW",
		description: "Maximize browser window",
		usage:       "window maximize [checkpoint-id] [position]",
		examples: []string{
			`api-cli window maximize cp_12345 1`,
			`api-cli window maximize  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"window_type": "MAXIMIZE",
			}
		},
		flags: []browserFlagConfig{},
	},
	"window.switch.tab.next": {
		commandType: "window",
		stepType:    "SWITCH",
		description: "Switch to next browser tab",
		usage:       "window switch tab next [checkpoint-id] [position]",
		examples: []string{
			`api-cli window switch tab next cp_12345 1`,
			`api-cli window switch tab next  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"direction": "next",
				"tab_type":  "NEXT_TAB",
			}
		},
		flags: []browserFlagConfig{},
	},
	"window.switch.tab.prev": {
		commandType: "window",
		stepType:    "SWITCH",
		description: "Switch to previous browser tab",
		usage:       "window switch tab prev [checkpoint-id] [position]",
		examples: []string{
			`api-cli window switch tab prev cp_12345 1`,
			`api-cli window switch tab prev  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"direction": "previous",
				"tab_type":  "PREVIOUS_TAB",
			}
		},
		flags: []browserFlagConfig{},
	},
	"window.switch.tab.index": {
		commandType: "window",
		stepType:    "SWITCH",
		description: "Switch to browser tab by index (0-based)",
		usage:       "window switch tab index [checkpoint-id] <index> [position]",
		examples: []string{
			`api-cli window switch tab index cp_12345 0 1  # Switch to first tab`,
			`api-cli window switch tab index 2  # Switch to third tab, uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			index, _ := strconv.Atoi(args[0])
			return map[string]interface{}{
				"index":    index,
				"tab_type": "TAB_BY_INDEX",
			}
		},
		flags: []browserFlagConfig{},
	},
	"window.switch.iframe": {
		commandType: "window",
		stepType:    "SWITCH",
		description: "Switch to iframe by element selector",
		usage:       "window switch iframe [checkpoint-id] <selector> [position]",
		examples: []string{
			`api-cli window switch iframe cp_12345 "#payment-frame" 1`,
			`api-cli window switch iframe "iframe[name='content']"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"selector":   args[0],
				"frame_type": "FRAME_BY_ELEMENT",
			}
		},
		flags: []browserFlagConfig{},
	},
	"window.switch.parent-frame": {
		commandType: "window",
		stepType:    "SWITCH",
		description: "Switch to parent frame",
		usage:       "window switch parent-frame [checkpoint-id] [position]",
		examples: []string{
			`api-cli window switch parent-frame cp_12345 1`,
			`api-cli window switch parent-frame  # Uses session context`,
		},
		requiredArgs: 0,
		buildMeta: func(args []string, flags map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"frame_type": "PARENT_FRAME",
			}
		},
		flags: []browserFlagConfig{},
	},
}

// NavigateCmd creates the navigate command with subcommands
func NavigateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "navigate",
		Short: "Navigate and scroll within the browser",
		Long: `Navigate to URLs and scroll within pages using various methods.

This command consolidates URL navigation and scrolling actions.`,
	}

	// Add subcommands
	cmd.AddCommand(createBrowserSubCmd("to", browserConfigs["navigate.to"]))
	cmd.AddCommand(createBrowserSubCmd("scroll-top", browserConfigs["navigate.scroll-top"]))
	cmd.AddCommand(createBrowserSubCmd("scroll-bottom", browserConfigs["navigate.scroll-bottom"]))
	cmd.AddCommand(createBrowserSubCmd("scroll-element", browserConfigs["navigate.scroll-element"]))
	cmd.AddCommand(createBrowserSubCmd("scroll-position", browserConfigs["navigate.scroll-position"]))
	cmd.AddCommand(createBrowserSubCmd("scroll-by", browserConfigs["navigate.scroll-by"]))
	cmd.AddCommand(createBrowserSubCmd("scroll-up", browserConfigs["navigate.scroll-up"]))
	cmd.AddCommand(createBrowserSubCmd("scroll-down", browserConfigs["navigate.scroll-down"]))

	return cmd
}

// newWindowCmd creates the window command with subcommands
func newWindowCmd() *cobra.Command {
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
	cmd.AddCommand(createBrowserSubCmd("resize", browserConfigs["window.resize"]))

	// Add maximize subcommand
	cmd.AddCommand(createBrowserSubCmd("maximize", browserConfigs["window.maximize"]))

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
	tabCmd.AddCommand(createBrowserSubCmd("next", browserConfigs["window.switch.tab.next"]))
	tabCmd.AddCommand(createBrowserSubCmd("prev", browserConfigs["window.switch.tab.prev"]))
	tabCmd.AddCommand(createBrowserSubCmd("index", browserConfigs["window.switch.tab.index"]))
	switchCmd.AddCommand(tabCmd)

	// Add iframe subcommand
	switchCmd.AddCommand(createBrowserSubCmd("iframe", browserConfigs["window.switch.iframe"]))

	// Add parent-frame subcommand
	switchCmd.AddCommand(createBrowserSubCmd("parent-frame", browserConfigs["window.switch.parent-frame"]))

	cmd.AddCommand(switchCmd)

	return cmd
}

// createBrowserSubCmd creates a subcommand for a specific browser operation
func createBrowserSubCmd(operation string, config browserConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   operation + " " + extractBrowserUsageArgs(config.usage),
		Short: config.description,
		Long: fmt.Sprintf(`%s

%s

Examples:
%s`, config.description, config.usage, strings.Join(config.examples, "\n")),
		RunE: func(cmd *cobra.Command, args []string) error {
			bc := &BrowserCommand{
				BaseCommand: NewBaseCommand(),
				commandType: config.commandType,
				operation:   operation,
			}

			// Determine the full operation key for nested commands
			fullOp := config.commandType + "." + operation
			if parentCmd := cmd.Parent(); parentCmd != nil {
				if parentCmd.Name() == "tab" {
					fullOp = "window.switch.tab." + operation
				} else if parentCmd.Name() == "switch" && operation == "iframe" {
					fullOp = "window.switch.iframe"
				} else if parentCmd.Name() == "switch" && operation == "parent-frame" {
					fullOp = "window.switch.parent-frame"
				}
			}

			// Collect flag values
			flags := make(map[string]interface{})
			for _, flagCfg := range config.flags {
				switch flagCfg.defaultValue.(type) {
				case bool:
					flags[flagCfg.name], _ = cmd.Flags().GetBool(flagCfg.name)
				case int:
					flags[flagCfg.name], _ = cmd.Flags().GetInt(flagCfg.name)
				case string:
					flags[flagCfg.name], _ = cmd.Flags().GetString(flagCfg.name)
				}
			}

			return bc.Execute(cmd, args, browserConfigs[fullOp], flags)
		},
	}

	// Add command aliases
	aliases := getCommandAliases(operation)
	if len(aliases) > 0 {
		cmd.Aliases = aliases
	}

	// Define flags
	for _, flagCfg := range config.flags {
		switch v := flagCfg.defaultValue.(type) {
		case bool:
			cmd.Flags().BoolVar(new(bool), flagCfg.name, v, flagCfg.description)
		case int:
			cmd.Flags().IntVar(new(int), flagCfg.name, v, flagCfg.description)
		case string:
			cmd.Flags().StringVar(new(string), flagCfg.name, v, flagCfg.description)
		}
	}

	// Validate args based on operation
	if config.requiredArgs > 0 {
		cmd.Args = cobra.RangeArgs(config.requiredArgs, config.requiredArgs+2) // +2 for optional checkpoint and position
	} else {
		cmd.Args = cobra.MaximumNArgs(2) // Only optional checkpoint and position
	}

	// Special handling for scroll commands that accept coordinate arguments
	if operation == "scroll-position" || operation == "scroll-by" {
		cmd.Args = cobra.RangeArgs(0, 3) // Can have coordinate arg + checkpoint + position
	}

	return cmd
}

// getCommandAliases returns aliases for specific commands
func getCommandAliases(operation string) []string {
	aliases := map[string][]string{
		"to":              {"url", "goto"},
		"scroll-top":      {"top"},
		"scroll-bottom":   {"bottom"},
		"scroll-element":  {"element", "to-element"},
		"scroll-position": {"position", "to-position", "xy"},
		"scroll-by":       {"by", "offset"},
		"scroll-up":       {"up"},
		"scroll-down":     {"down"},
	}
	return aliases[operation]
}

// extractBrowserUsageArgs extracts the arguments portion from the usage string
func extractBrowserUsageArgs(usage string) string {
	parts := strings.Fields(usage)
	// Find the start of arguments (after command keywords)
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

// Execute runs the browser command
func (bc *BrowserCommand) Execute(cmd *cobra.Command, args []string, config browserConfig, flags map[string]interface{}) error {
	// Initialize base command
	if err := bc.Init(cmd); err != nil {
		return fmt.Errorf("failed to initialize command: %w", err)
	}

	// Special handling for scroll commands with coordinate arguments
	if bc.operation == "scroll-position" || bc.operation == "scroll-by" {
		args = bc.parseCoordinateArgs(args, flags)
	}

	// Resolve checkpoint and position
	remainingArgs, err := bc.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve checkpoint and position: %w", err)
	}

	// Validate we have the required number of arguments
	if len(remainingArgs) < config.requiredArgs {
		return fmt.Errorf("expected %d arguments, got %d - usage: %s", config.requiredArgs, len(remainingArgs), config.usage)
	}

	// Additional validation for specific operations
	if err := bc.validateBrowserArgs(bc.operation, remainingArgs); err != nil {
		return err
	}

	// Build request metadata
	meta := config.buildMeta(remainingArgs, flags)

	// Create the step
	stepResult, err := bc.createBrowserStep(config.stepType, meta, remainingArgs)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", config.stepType, err)
	}

	// Format and output the result
	output, err := bc.FormatOutput(stepResult, bc.OutputFormat)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// parseCoordinateArgs handles coordinate parsing for scroll-position and scroll-by commands
func (bc *BrowserCommand) parseCoordinateArgs(args []string, flags map[string]interface{}) []string {
	if len(args) > 0 && strings.Contains(args[0], ",") {
		coords := strings.Split(args[0], ",")
		if len(coords) == 2 {
			if xVal, err := strconv.Atoi(strings.TrimSpace(coords[0])); err == nil {
				flags["x"] = xVal
			}
			if yVal, err := strconv.Atoi(strings.TrimSpace(coords[1])); err == nil {
				flags["y"] = yVal
			}
			// Remove the coordinate argument
			return append(args[:0], args[1:]...)
		}
	}
	return args
}

// validateBrowserArgs validates arguments for specific browser operations
func (bc *BrowserCommand) validateBrowserArgs(operation string, args []string) error {
	switch operation {
	case "to":
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("URL cannot be empty")
		}
		return ValidateURL(args[0])
	case "scroll-element", "iframe":
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
		return ValidateSelector(args[0])
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
	}
	return nil
}

// createBrowserStep creates a browser step via the API
func (bc *BrowserCommand) createBrowserStep(stepType string, meta map[string]interface{}, args []string) (*StepResult, error) {
	// Convert checkpoint ID from string to int
	checkpointID, err := strconv.Atoi(bc.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint ID: %s - must be a numeric ID", bc.CheckpointID)
	}

	// Call the appropriate API method based on command type and operation
	var stepID int

	if bc.commandType == "navigate" {
		stepID, err = bc.executeNavigateAction(checkpointID, bc.operation, meta, args)
	} else {
		stepID, err = bc.executeWindowAction(checkpointID, stepType, meta)
	}

	if err != nil {
		// Add context about the operation that failed
		opDescription := bc.buildDescription(stepType, meta, args)
		return nil, fmt.Errorf("failed to create %s: %w", opDescription, err)
	}

	// Build the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: bc.CheckpointID,
		Type:         stepType,
		Position:     bc.Position,
		Description:  bc.buildDescription(stepType, meta, args),
		Selector:     bc.extractSelector(meta),
		Meta:         meta,
	}

	// Save session state if position was auto-incremented
	if bc.Position == -1 && cfg.Session.AutoIncrementPos {
		if err := cfg.SaveConfig(); err != nil {
			// Don't fail the command, just warn
			// Note: In production, this warning would be sent to stderr
		}
	}

	return result, nil
}

// executeNavigateAction executes a navigation command
func (bc *BrowserCommand) executeNavigateAction(checkpointID int, operation string, meta map[string]interface{}, args []string) (int, error) {
	c := bc.Client

	// Note: Context is handled internally by the client methods
	// They call WithContext versions with appropriate timeouts

	switch operation {
	case "to":
		newTab, _ := meta["newTab"].(bool)
		stepID, err := c.CreateStepNavigate(checkpointID, meta["url"].(string), newTab, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to navigate to URL: %w", err)
		}
		return stepID, nil
	case "scroll-top":
		stepID, err := c.CreateStepScrollToTop(checkpointID, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to scroll to top: %w", err)
		}
		return stepID, nil
	case "scroll-bottom":
		stepID, err := c.CreateStepScrollBottom(checkpointID, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to scroll to bottom: %w", err)
		}
		return stepID, nil
	case "scroll-element":
		selector := meta["selector"].(string)
		stepID, err := c.CreateStepScrollElement(checkpointID, selector, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to scroll to element '%s': %w", selector, err)
		}
		return stepID, nil
	case "scroll-position":
		x, _ := meta["x"].(int)
		y, _ := meta["y"].(int)
		stepID, err := c.CreateStepScrollToPosition(checkpointID, x, y, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to scroll to position (%d, %d): %w", x, y, err)
		}
		return stepID, nil
	case "scroll-by":
		x, _ := meta["x"].(int)
		y, _ := meta["y"].(int)
		stepID, err := c.CreateStepScrollByOffset(checkpointID, x, y, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to scroll by offset (%d, %d): %w", x, y, err)
		}
		return stepID, nil
	case "scroll-up":
		// Scroll up by one viewport height (negative Y value)
		stepID, err := c.CreateStepScrollByOffset(checkpointID, 0, -1000, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to scroll up: %w", err)
		}
		return stepID, nil
	case "scroll-down":
		// Scroll down by one viewport height (positive Y value)
		stepID, err := c.CreateStepScrollByOffset(checkpointID, 0, 1000, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to scroll down: %w", err)
		}
		return stepID, nil
	default:
		return 0, fmt.Errorf("unknown navigate operation: %s", operation)
	}
}

// executeWindowAction executes a window command
func (bc *BrowserCommand) executeWindowAction(checkpointID int, stepType string, meta map[string]interface{}) (int, error) {
	c := bc.Client

	// Note: Context is handled internally by the client methods
	// They call WithContext versions with appropriate timeouts

	switch {
	case stepType == "RESIZE":
		width := meta["width"].(int)
		height := meta["height"].(int)
		stepID, err := c.CreateStepWindowResize(checkpointID, width, height, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to resize window to %dx%d: %w", width, height, err)
		}
		return stepID, nil

	case stepType == "WINDOW" && meta["window_type"] == "MAXIMIZE":
		stepID, err := c.CreateStepWindowMaximize(checkpointID, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to maximize window: %w", err)
		}
		return stepID, nil

	case stepType == "SWITCH" && meta["tab_type"] == "NEXT_TAB":
		stepID, err := c.CreateStepSwitchNextTab(checkpointID, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to switch to next tab: %w", err)
		}
		return stepID, nil

	case stepType == "SWITCH" && meta["tab_type"] == "PREVIOUS_TAB":
		stepID, err := c.CreateStepSwitchPrevTab(checkpointID, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to switch to previous tab: %w", err)
		}
		return stepID, nil

	case stepType == "SWITCH" && meta["tab_type"] == "TAB_BY_INDEX":
		index := meta["index"].(int)
		stepID, err := c.CreateStepSwitchTabByIndex(checkpointID, index, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to switch to tab index %d: %w", index, err)
		}
		return stepID, nil

	case stepType == "SWITCH" && meta["frame_type"] == "FRAME_BY_ELEMENT":
		selector := meta["selector"].(string)
		stepID, err := c.CreateStepSwitchIframe(checkpointID, selector, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to switch to iframe '%s': %w", selector, err)
		}
		return stepID, nil

	case stepType == "SWITCH" && meta["frame_type"] == "PARENT_FRAME":
		stepID, err := c.CreateStepSwitchParentFrame(checkpointID, bc.Position)
		if err != nil {
			return 0, fmt.Errorf("failed to switch to parent frame: %w", err)
		}
		return stepID, nil

	default:
		return 0, fmt.Errorf("unknown window operation: %s with meta %v", stepType, meta)
	}
}

// buildDescription builds a human-readable description for the step
func (bc *BrowserCommand) buildDescription(stepType string, meta map[string]interface{}, args []string) string {
	if bc.commandType == "navigate" {
		switch bc.operation {
		case "to":
			desc := fmt.Sprintf("navigate to %s", meta["url"])
			if newTab, ok := meta["newTab"].(bool); ok && newTab {
				desc += " (new tab)"
			}
			return desc
		case "scroll-top":
			return "scroll to top"
		case "scroll-bottom":
			return "scroll to bottom"
		case "scroll-element":
			return fmt.Sprintf("scroll to element \"%s\"", meta["selector"])
		case "scroll-position":
			return fmt.Sprintf("scroll to position (%d, %d)", meta["x"], meta["y"])
		case "scroll-by":
			return fmt.Sprintf("scroll by (%d, %d)", meta["x"], meta["y"])
		case "scroll-up":
			return "scroll up"
		case "scroll-down":
			return "scroll down"
		}
	} else if bc.commandType == "window" {
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
		}
	}
	return fmt.Sprintf("%s operation", stepType)
}

// extractSelector extracts the selector from metadata if present
func (bc *BrowserCommand) extractSelector(meta map[string]interface{}) string {
	if selector, ok := meta["selector"].(string); ok {
		return selector
	}
	return ""
}
