package commands

import (
	"fmt"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// InteractCmd creates the interact command with subcommands
func InteractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interact",
		Short: "Interact with elements (click, write, hover, etc.)",
		Long: `Interact with page elements through various actions like clicking, writing text, hovering, and keyboard input.

This command consolidates multiple interaction types into a single interface.`,
	}

	// Add subcommands
	cmd.AddCommand(clickSubCmd())
	cmd.AddCommand(doubleClickSubCmd())
	cmd.AddCommand(rightClickSubCmd())
	cmd.AddCommand(hoverSubCmd())
	cmd.AddCommand(writeSubCmd())
	cmd.AddCommand(keySubCmd())

	return cmd
}

// clickSubCmd creates the click subcommand
func clickSubCmd() *cobra.Command {
	var (
		variable    string
		position    string
		elementType string
	)

	cmd := &cobra.Command{
		Use:   "click [checkpoint-id] <selector> [position]",
		Short: "Click on an element",
		Long: `Click on an element identified by a CSS selector.

Examples:
  # Using session context (modern)
  api-cli interact click "button.submit"
  api-cli interact click "#login-btn" --position TOP_RIGHT
  api-cli interact click "a.nav-link" --variable "linkText"

  # Using explicit checkpoint (legacy)
  api-cli interact click cp_12345 "button.submit" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteraction(cmd, args, "click", map[string]interface{}{
				"variable":    variable,
				"position":    position,
				"elementType": elementType,
			})
		},
	}

	cmd.Flags().StringVar(&variable, "variable", "", "Store element text in variable")
	cmd.Flags().StringVar(&position, "position", "CENTER", "Click position on element")
	cmd.Flags().StringVar(&elementType, "element-type", "", "Type of element (BUTTON, LINK, etc.)")

	return cmd
}

// doubleClickSubCmd creates the double-click subcommand
func doubleClickSubCmd() *cobra.Command {
	var position string

	cmd := &cobra.Command{
		Use:   "double-click [checkpoint-id] <selector> [position]",
		Short: "Double-click on an element",
		Long: `Double-click on an element identified by a CSS selector.

Examples:
  # Using session context
  api-cli interact double-click ".item-card"
  api-cli interact double-click "#file-icon" --position CENTER

  # Using explicit checkpoint
  api-cli interact double-click cp_12345 ".item-card" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteraction(cmd, args, "double-click", map[string]interface{}{
				"position": position,
			})
		},
	}

	cmd.Flags().StringVar(&position, "position", "CENTER", "Click position on element")

	return cmd
}

// rightClickSubCmd creates the right-click subcommand
func rightClickSubCmd() *cobra.Command {
	var position string

	cmd := &cobra.Command{
		Use:   "right-click [checkpoint-id] <selector> [position]",
		Short: "Right-click on an element",
		Long: `Right-click on an element to open context menu.

Examples:
  # Using session context
  api-cli interact right-click ".data-row"
  api-cli interact right-click "#context-target" --position TOP_LEFT

  # Using explicit checkpoint
  api-cli interact right-click cp_12345 ".data-row" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteraction(cmd, args, "right-click", map[string]interface{}{
				"position": position,
			})
		},
	}

	cmd.Flags().StringVar(&position, "position", "CENTER", "Click position on element")

	return cmd
}

// hoverSubCmd creates the hover subcommand
func hoverSubCmd() *cobra.Command {
	var (
		position string
		duration int
	)

	cmd := &cobra.Command{
		Use:   "hover [checkpoint-id] <selector> [position]",
		Short: "Hover over an element",
		Long: `Hover the mouse over an element, optionally for a specific duration.

Examples:
  # Using session context
  api-cli interact hover ".menu-item"
  api-cli interact hover "#tooltip-trigger" --duration 2000

  # Using explicit checkpoint
  api-cli interact hover cp_12345 ".menu-item" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteraction(cmd, args, "hover", map[string]interface{}{
				"position": position,
				"duration": duration,
			})
		},
	}

	cmd.Flags().StringVar(&position, "position", "CENTER", "Hover position on element")
	cmd.Flags().IntVar(&duration, "duration", 0, "Hover duration in milliseconds")

	return cmd
}

// writeSubCmd creates the write subcommand
func writeSubCmd() *cobra.Command {
	var (
		variable string
		clear    bool
		delay    int
	)

	cmd := &cobra.Command{
		Use:   "write [checkpoint-id] <selector> <text> [position]",
		Short: "Write text to an input element",
		Long: `Write or type text into an input field, textarea, or contenteditable element.

Examples:
  # Using session context
  api-cli interact write "input#username" "john.doe@example.com"
  api-cli interact write "textarea.comment" "This is a comment" --clear
  api-cli interact write "#search" "{{searchTerm}}" --variable searchTerm

  # Using explicit checkpoint
  api-cli interact write cp_12345 "input#username" "john.doe@example.com" 1`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteraction(cmd, args, "write", map[string]interface{}{
				"variable": variable,
				"clear":    clear,
				"delay":    delay,
			})
		},
	}

	cmd.Flags().StringVar(&variable, "variable", "", "Use variable value for text")
	cmd.Flags().BoolVar(&clear, "clear", false, "Clear field before writing")
	cmd.Flags().IntVar(&delay, "delay", 0, "Delay between keystrokes in ms")

	return cmd
}

// keySubCmd creates the key press subcommand
func keySubCmd() *cobra.Command {
	var (
		target    string
		modifiers []string
		repeat    int
	)

	cmd := &cobra.Command{
		Use:   "key [checkpoint-id] <key> [position]",
		Short: "Press keyboard keys",
		Long: `Press one or more keyboard keys, optionally with modifiers.

Examples:
  # Using session context
  api-cli interact key "Enter"
  api-cli interact key "Escape"
  api-cli interact key "a" --modifiers ctrl
  api-cli interact key "Tab" --target "input#username"

  # Using explicit checkpoint
  api-cli interact key cp_12345 "Enter" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteraction(cmd, args, "key", map[string]interface{}{
				"target":    target,
				"modifiers": modifiers,
				"repeat":    repeat,
			})
		},
	}

	cmd.Flags().StringVar(&target, "target", "", "Target element selector")
	cmd.Flags().StringSliceVar(&modifiers, "modifiers", []string{}, "Key modifiers (ctrl, shift, alt, meta)")
	cmd.Flags().IntVar(&repeat, "repeat", 1, "Number of times to press key")

	return cmd
}

// runInteraction executes an interaction command
func runInteraction(cmd *cobra.Command, args []string, action string, options map[string]interface{}) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	var err error
	requiredArgs := 1 // selector
	if action == "write" {
		requiredArgs = 2 // selector and text
	}

	args, err = base.ResolveCheckpointAndPosition(args, requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Convert checkpoint ID to int
	checkpointID, err := strconv.Atoi(base.CheckpointID)
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	// Build and execute based on action
	var stepID int
	switch action {
	case "click":
		stepID, err = executeClickAction(base.Client, checkpointID, args[0], base.Position, options)
	case "double-click":
		stepID, err = executeDoubleClickAction(base.Client, checkpointID, args[0], base.Position, options)
	case "right-click":
		stepID, err = executeRightClickAction(base.Client, checkpointID, args[0], base.Position, options)
	case "hover":
		stepID, err = executeHoverAction(base.Client, checkpointID, args[0], base.Position, options)
	case "write":
		stepID, err = executeWriteAction(base.Client, checkpointID, args[0], args[1], base.Position, options)
	case "key":
		stepID, err = executeKeyAction(base.Client, checkpointID, args[0], base.Position, options)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", action, err)
	}

	// Format and output the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: base.CheckpointID,
		Type:         action,
		Position:     base.Position,
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// isValidClickPosition validates if the given position is a valid click position enum
func isValidClickPosition(position string) bool {
	validPositions := []string{
		"TOP_LEFT", "TOP_CENTER", "TOP_RIGHT",
		"CENTER_LEFT", "CENTER", "CENTER_RIGHT",
		"BOTTOM_LEFT", "BOTTOM_CENTER", "BOTTOM_RIGHT",
	}
	for _, valid := range validPositions {
		if position == valid {
			return true
		}
	}
	return false
}

// executeClickAction executes a click action using the client
func executeClickAction(c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	variable, _ := options["variable"].(string)
	positionType, _ := options["position"].(string)
	elementType, _ := options["elementType"].(string)

	// Validate position enum if provided
	if positionType != "" && positionType != "CENTER" && !isValidClickPosition(positionType) {
		return 0, fmt.Errorf("invalid position: %s. Valid positions: TOP_LEFT, TOP_CENTER, TOP_RIGHT, CENTER_LEFT, CENTER, CENTER_RIGHT, BOTTOM_LEFT, BOTTOM_CENTER, BOTTOM_RIGHT", positionType)
	}

	if variable != "" {
		return c.CreateStepClickWithVariable(checkpointID, variable, position)
	} else if positionType != "" && elementType != "" {
		return c.CreateStepClickWithDetails(checkpointID, selector, positionType, elementType, position)
	} else {
		return c.CreateStepClick(checkpointID, selector, position)
	}
}

// executeDoubleClickAction executes a double-click action using the client
func executeDoubleClickAction(c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	return c.CreateStepDoubleClick(checkpointID, selector, position)
}

// executeRightClickAction executes a right-click action using the client
func executeRightClickAction(c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	return c.CreateStepRightClick(checkpointID, selector, position)
}

// executeHoverAction executes a hover action using the client
func executeHoverAction(c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	return c.CreateStepHover(checkpointID, selector, position)
}

// executeWriteAction executes a write action using the client
func executeWriteAction(c *client.Client, checkpointID int, selector string, text string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	variable, _ := options["variable"].(string)

	if variable != "" {
		return c.CreateStepWriteWithVariable(checkpointID, selector, text, variable, position)
	} else {
		return c.CreateStepWrite(checkpointID, selector, text, position)
	}
}

// executeKeyAction executes a key press action using the client
func executeKeyAction(c *client.Client, checkpointID int, key string, position int, options map[string]interface{}) (int, error) {
	target, _ := options["target"].(string)

	if target != "" {
		if err := ValidateSelector(target); err != nil {
			return 0, err
		}
		return c.CreateStepKeyTargeted(checkpointID, target, key, position)
	} else {
		return c.CreateStepKeyGlobal(checkpointID, key, position)
	}
}
