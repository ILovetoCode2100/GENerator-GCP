package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// ================================================================================
// MAIN INTERACTION COMMAND
// ================================================================================

// InteractionCmd creates the consolidated interaction command with all subcommands
func InteractionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interact",
		Short: "Interact with page elements (click, hover, type, select, etc.)",
		Long: `Interact with page elements through various actions including:
  - Click actions: click, double-click, right-click
  - Text input: write, key presses
  - Mouse actions: hover, move, drag
  - Dropdown selection: select by value, index, or last option

This command consolidates all user interaction types into a single interface.`,
	}

	// Click-based interactions
	cmd.AddCommand(clickSubCmd())
	cmd.AddCommand(doubleClickSubCmd())
	cmd.AddCommand(rightClickSubCmd())

	// Text and keyboard interactions
	cmd.AddCommand(writeSubCmd())
	cmd.AddCommand(keySubCmd())

	// Mouse interactions
	cmd.AddCommand(hoverSubCmd())
	cmd.AddCommand(mouseSubCmd())

	// Dropdown selection
	cmd.AddCommand(selectSubCmd())

	return cmd
}

// ================================================================================
// CLICK-BASED INTERACTIONS
// ================================================================================

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

// ================================================================================
// TEXT AND KEYBOARD INTERACTIONS
// ================================================================================

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

// ================================================================================
// MOUSE INTERACTIONS
// ================================================================================

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

// mouseSubCmd creates the mouse movement and advanced operations subcommand
func mouseSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mouse",
		Short: "Advanced mouse operations",
		Long: `Perform advanced mouse operations including movement, drag and drop.

Available actions:
  - move-to: Move mouse to element center
  - move-by: Move mouse by relative offset
  - move: Move mouse to absolute coordinates
  - down: Press mouse button down
  - up: Release mouse button
  - enter: Move mouse into element`,
	}

	// Add mouse subcommands
	cmd.AddCommand(mouseMoveToSubCmd())
	cmd.AddCommand(mouseMoveBySubCmd())
	cmd.AddCommand(mouseMoveSubCmd())
	cmd.AddCommand(mouseDownSubCmd())
	cmd.AddCommand(mouseUpSubCmd())
	cmd.AddCommand(mouseEnterSubCmd())

	return cmd
}

// mouseMoveToSubCmd - move mouse to element
func mouseMoveToSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move-to [checkpoint-id] <selector> [position]",
		Short: "Move mouse to element center",
		Long: `Move the mouse cursor to the center of an element.

Examples:
  api-cli interact mouse move-to "button.submit"
  api-cli interact mouse move-to cp_12345 "#target" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMouseAction(cmd, args, "move-to")
		},
	}
	return cmd
}

// mouseMoveBySubCmd - move mouse by offset
func mouseMoveBySubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move-by [checkpoint-id] <x,y> [position]",
		Short: "Move mouse by relative offset",
		Long: `Move the mouse cursor by a relative offset from its current position.

Examples:
  api-cli interact mouse move-by "100,50"
  api-cli interact mouse move-by "-50,25"
  api-cli interact mouse move-by cp_12345 "100,-50" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMouseAction(cmd, args, "move-by")
		},
	}
	return cmd
}

// mouseMoveSubCmd - move mouse to coordinates
func mouseMoveSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move [checkpoint-id] <x,y> [position]",
		Short: "Move mouse to absolute coordinates",
		Long: `Move the mouse cursor to absolute screen coordinates.

Examples:
  api-cli interact mouse move "500,300"
  api-cli interact mouse move cp_12345 "100,200" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMouseAction(cmd, args, "move")
		},
	}
	return cmd
}

// mouseDownSubCmd - press mouse button
func mouseDownSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down [checkpoint-id] <selector> [position]",
		Short: "Press mouse button down",
		Long: `Press and hold the mouse button on an element.

Examples:
  api-cli interact mouse down "#drag-handle"
  api-cli interact mouse down cp_12345 "button" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMouseAction(cmd, args, "down")
		},
	}
	return cmd
}

// mouseUpSubCmd - release mouse button
func mouseUpSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up [checkpoint-id] <selector> [position]",
		Short: "Release mouse button",
		Long: `Release the mouse button on an element.

Examples:
  api-cli interact mouse up "#drop-zone"
  api-cli interact mouse up cp_12345 "button" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMouseAction(cmd, args, "up")
		},
	}
	return cmd
}

// mouseEnterSubCmd - move mouse into element
func mouseEnterSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enter [checkpoint-id] <selector> [position]",
		Short: "Move mouse into element",
		Long: `Move the mouse cursor into an element's boundaries.

Examples:
  api-cli interact mouse enter "#tooltip-target"
  api-cli interact mouse enter cp_12345 ".dropdown-trigger" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMouseAction(cmd, args, "enter")
		},
	}
	return cmd
}

// ================================================================================
// DROPDOWN SELECTION
// ================================================================================

// selectSubCmd creates the select dropdown subcommand
func selectSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select options from dropdowns",
		Long: `Select options from dropdown elements using various methods.

Available methods:
  - option: Select by value or text
  - index: Select by index (0-based)
  - last: Select the last option`,
	}

	// Add selection subcommands
	cmd.AddCommand(selectOptionSubCmd())
	cmd.AddCommand(selectIndexSubCmd())
	cmd.AddCommand(selectLastSubCmd())

	return cmd
}

// selectOptionSubCmd - select by value/text
func selectOptionSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "option [checkpoint-id] <selector> <value> [position]",
		Short: "Select dropdown option by value or text",
		Long: `Select a dropdown option by its value attribute or visible text.

Examples:
  api-cli interact select option "#country" "United States"
  api-cli interact select option cp_12345 "select[name='country']" "US" 1`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelectAction(cmd, args, "option")
		},
	}
	return cmd
}

// selectIndexSubCmd - select by index
func selectIndexSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index [checkpoint-id] <selector> <index> [position]",
		Short: "Select dropdown option by index (0-based)",
		Long: `Select a dropdown option by its index position (0-based).

Examples:
  api-cli interact select index "#country" 0
  api-cli interact select index cp_12345 ".dropdown" 3 1`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelectAction(cmd, args, "index")
		},
	}
	return cmd
}

// selectLastSubCmd - select last option
func selectLastSubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last [checkpoint-id] <selector> [position]",
		Short: "Select the last dropdown option",
		Long: `Select the last option in a dropdown.

Examples:
  api-cli interact select last "#country"
  api-cli interact select last cp_12345 ".dropdown" 1`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelectAction(cmd, args, "last")
		},
	}
	return cmd
}

// ================================================================================
// SHARED EXECUTION FUNCTIONS
// ================================================================================

// runInteraction executes a basic interaction command with context support
func runInteraction(cmd *cobra.Command, args []string, action string, options map[string]interface{}) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Create context for this operation
	ctx, cancel := base.CommandContext()
	defer cancel()

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
		stepID, err = executeClickActionWithContext(ctx, base.Client, checkpointID, args[0], base.Position, options)
	case "double-click":
		stepID, err = executeDoubleClickActionWithContext(ctx, base.Client, checkpointID, args[0], base.Position, options)
	case "right-click":
		stepID, err = executeRightClickActionWithContext(ctx, base.Client, checkpointID, args[0], base.Position, options)
	case "hover":
		stepID, err = executeHoverActionWithContext(ctx, base.Client, checkpointID, args[0], base.Position, options)
	case "write":
		stepID, err = executeWriteActionWithContext(ctx, base.Client, checkpointID, args[0], args[1], base.Position, options)
	case "key":
		stepID, err = executeKeyActionWithContext(ctx, base.Client, checkpointID, args[0], base.Position, options)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		// Enhanced error handling with context-aware messages
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("operation timed out after 30 seconds while creating %s step", action)
		} else if ctx.Err() == context.Canceled {
			return fmt.Errorf("operation was canceled while creating %s step", action)
		}

		// Check for common API errors and provide helpful messages
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "404"):
			return fmt.Errorf("checkpoint not found (ID: %d). Please verify the checkpoint exists", checkpointID)
		case strings.Contains(errMsg, "401"):
			return fmt.Errorf("authentication failed. Please check your API token in the configuration")
		case strings.Contains(errMsg, "403"):
			return fmt.Errorf("access denied. You don't have permission to modify this checkpoint")
		case strings.Contains(errMsg, "400"):
			return fmt.Errorf("invalid request for %s step: %w", action, err)
		case strings.Contains(errMsg, "500"):
			return fmt.Errorf("server error while creating %s step. Please try again later", action)
		default:
			return fmt.Errorf("failed to create %s step: %w", action, err)
		}
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

// runMouseAction executes a mouse action command
func runMouseAction(cmd *cobra.Command, args []string, action string) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Determine required args based on action
	requiredArgs := 1

	// Resolve checkpoint and position
	args, err := base.ResolveCheckpointAndPosition(args, requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate arguments
	if err := validateMouseArgs(args, action); err != nil {
		return err
	}

	// Convert checkpoint ID to int
	checkpointID, err := strconv.Atoi(base.CheckpointID)
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	// Execute the appropriate mouse action
	var stepID int
	switch action {
	case "move-to":
		stepID, err = base.Client.CreateStepMouseMove(checkpointID, args[0], base.Position)
	case "move-by":
		coords := parseCoordinates(args[0])
		stepID, err = base.Client.CreateStepMouseMoveBy(checkpointID, coords[0], coords[1], base.Position)
	case "move":
		coords := parseCoordinates(args[0])
		stepID, err = base.Client.CreateStepMouseMoveTo(checkpointID, coords[0], coords[1], base.Position)
	case "down":
		stepID, err = base.Client.CreateStepMouseDown(checkpointID, args[0], base.Position)
	case "up":
		stepID, err = base.Client.CreateStepMouseUp(checkpointID, args[0], base.Position)
	case "enter":
		stepID, err = base.Client.CreateStepMouseEnter(checkpointID, args[0], base.Position)
	default:
		return fmt.Errorf("unknown mouse action: %s", action)
	}

	if err != nil {
		return fmt.Errorf("failed to create mouse %s step: %w", action, err)
	}

	// Format and output the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: base.CheckpointID,
		Type:         "mouse-" + action,
		Position:     base.Position,
		Description:  buildMouseDescription(action, args),
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// runSelectAction executes a select action command
func runSelectAction(cmd *cobra.Command, args []string, selectType string) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Determine required args
	requiredArgs := 1 // selector for "last"
	if selectType == "option" || selectType == "index" {
		requiredArgs = 2 // selector and value/index
	}

	// Resolve checkpoint and position
	args, err := base.ResolveCheckpointAndPosition(args, requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate arguments
	if err := validateSelectArgs(args, selectType); err != nil {
		return err
	}

	// Convert checkpoint ID to int
	checkpointID, err := strconv.Atoi(base.CheckpointID)
	if err != nil {
		return fmt.Errorf("invalid checkpoint ID: %w", err)
	}

	// Execute the appropriate select action
	var stepID int
	switch selectType {
	case "option":
		stepID, err = base.Client.CreateStepPick(checkpointID, args[0], args[1], base.Position)
	case "index":
		index, _ := strconv.Atoi(args[1])
		stepID, err = base.Client.CreateStepPickIndex(checkpointID, args[0], index, base.Position)
	case "last":
		stepID, err = base.Client.CreateStepPickLast(checkpointID, args[0], base.Position)
	default:
		return fmt.Errorf("unknown select type: %s", selectType)
	}

	if err != nil {
		return fmt.Errorf("failed to create select %s step: %w", selectType, err)
	}

	// Format and output the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: base.CheckpointID,
		Type:         "select-" + selectType,
		Position:     base.Position,
		Description:  buildSelectDescription(selectType, args),
		Selector:     args[0],
	}

	output, err := base.FormatOutput(result, base.OutputFormat)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Println(output)
	return nil
}

// ================================================================================
// ACTION EXECUTION FUNCTIONS
// ================================================================================

// executeClickActionWithContext executes a click action using the client with context support
func executeClickActionWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("operation cancelled before execution: %w", ctx.Err())
	default:
	}

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

	// Use the native context-aware client methods
	if variable != "" {
		return c.CreateStepClickWithVariableWithContext(ctx, checkpointID, variable, position)
	} else if positionType != "" && elementType != "" {
		return c.CreateStepClickWithDetailsWithContext(ctx, checkpointID, selector, positionType, elementType, position)
	} else {
		return c.CreateStepClickWithContext(ctx, checkpointID, selector, position)
	}
}

// executeDoubleClickActionWithContext executes a double-click action using the client with context support
func executeDoubleClickActionWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	return c.CreateStepDoubleClickWithContext(ctx, checkpointID, selector, position)
}

// executeRightClickActionWithContext executes a right-click action using the client with context support
func executeRightClickActionWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	return c.CreateStepRightClickWithContext(ctx, checkpointID, selector, position)
}

// executeHoverActionWithContext executes a hover action using the client with context support
func executeHoverActionWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	return c.CreateStepHoverWithContext(ctx, checkpointID, selector, position)
}

// executeWriteActionWithContext executes a write action using the client with context support
func executeWriteActionWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, text string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	variable, _ := options["variable"].(string)

	if variable != "" {
		return c.CreateStepWriteWithVariableWithContext(ctx, checkpointID, selector, text, variable, position)
	} else {
		return c.CreateStepWriteWithContext(ctx, checkpointID, selector, text, position)
	}
}

// executeKeyActionWithContext executes a key press action using the client with context support
func executeKeyActionWithContext(ctx context.Context, c *client.Client, checkpointID int, key string, position int, options map[string]interface{}) (int, error) {
	target, _ := options["target"].(string)
	modifiers, _ := options["modifiers"].([]string)

	// Handle modifier keys
	if len(modifiers) > 0 {
		if target != "" {
			if err := ValidateSelector(target); err != nil {
				return 0, err
			}
			// TODO: Add context support for CreateStepKeyTargetedWithModifiers when available
			return c.CreateStepKeyTargetedWithModifiers(checkpointID, target, key, modifiers, position)
		} else {
			// TODO: Add context support for CreateStepKeyGlobalWithModifiers when available
			return c.CreateStepKeyGlobalWithModifiers(checkpointID, key, modifiers, position)
		}
	}

	// Original implementation for simple keys
	if target != "" {
		if err := ValidateSelector(target); err != nil {
			return 0, err
		}
		return c.CreateStepKeyTargetedWithContext(ctx, checkpointID, target, key, position)
	} else {
		return c.CreateStepKeyGlobalWithContext(ctx, checkpointID, key, position)
	}
}

// ================================================================================
// MOUSE ACTION EXECUTION FUNCTIONS
// ================================================================================

// executeMouseMoveToWithContext executes a mouse move-to action with context support
func executeMouseMoveToWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int) (int, error) {
	// TODO: Add native context support when available in client
	// For now, use the non-context version
	return c.CreateStepMouseMove(checkpointID, selector, position)
}

// executeMouseMoveByWithContext executes a mouse move-by action with context support
func executeMouseMoveByWithContext(ctx context.Context, c *client.Client, checkpointID int, x, y, position int) (int, error) {
	// TODO: Add native context support when available in client
	return c.CreateStepMouseMoveBy(checkpointID, x, y, position)
}

// executeMouseMoveWithContext executes a mouse move action with context support
func executeMouseMoveWithContext(ctx context.Context, c *client.Client, checkpointID int, x, y, position int) (int, error) {
	// TODO: Add native context support when available in client
	return c.CreateStepMouseMoveTo(checkpointID, x, y, position)
}

// executeMouseDownWithContext executes a mouse down action with context support
func executeMouseDownWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int) (int, error) {
	// TODO: Add native context support when available in client
	return c.CreateStepMouseDown(checkpointID, selector, position)
}

// executeMouseUpWithContext executes a mouse up action with context support
func executeMouseUpWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int) (int, error) {
	// TODO: Add native context support when available in client
	return c.CreateStepMouseUp(checkpointID, selector, position)
}

// executeMouseEnterWithContext executes a mouse enter action with context support
func executeMouseEnterWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int) (int, error) {
	// TODO: Add native context support when available in client
	return c.CreateStepMouseEnter(checkpointID, selector, position)
}

// ================================================================================
// SELECT ACTION EXECUTION FUNCTIONS
// ================================================================================

// executeSelectOptionWithContext executes a select option action with context support
func executeSelectOptionWithContext(ctx context.Context, c *client.Client, checkpointID int, selector, value string, position int) (int, error) {
	// TODO: Add native context support when available in client
	return c.CreateStepPick(checkpointID, selector, value, position)
}

// executeSelectIndexWithContext executes a select by index action with context support
func executeSelectIndexWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, index, position int) (int, error) {
	// TODO: Add native context support when available in client
	return c.CreateStepPickIndex(checkpointID, selector, index, position)
}

// executeSelectLastWithContext executes a select last action with context support
func executeSelectLastWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int) (int, error) {
	// TODO: Add native context support when available in client
	return c.CreateStepPickLast(checkpointID, selector, position)
}

// ================================================================================
// VALIDATION FUNCTIONS
// ================================================================================

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

// validateMouseArgs validates arguments for mouse actions
func validateMouseArgs(args []string, action string) error {
	switch action {
	case "down", "up", "enter", "move-to":
		if args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
	case "move-by", "move":
		// Validate coordinate format
		coords := strings.Split(args[0], ",")
		if len(coords) != 2 {
			return fmt.Errorf("coordinates must be in format 'x,y'")
		}
		if _, err := strconv.Atoi(strings.TrimSpace(coords[0])); err != nil {
			return fmt.Errorf("X coordinate must be a number")
		}
		if _, err := strconv.Atoi(strings.TrimSpace(coords[1])); err != nil {
			return fmt.Errorf("Y coordinate must be a number")
		}
	}
	return nil
}

// validateSelectArgs validates arguments for select actions
func validateSelectArgs(args []string, selectType string) error {
	// Validate selector
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	// Type-specific validation
	switch selectType {
	case "option":
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("value cannot be empty")
		}
	case "index":
		if len(args) < 2 {
			return fmt.Errorf("index is required")
		}
		// Validate index is a non-negative integer
		index, err := strconv.Atoi(args[1])
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

// ================================================================================
// HELPER FUNCTIONS
// ================================================================================

// parseCoordinates parses x,y coordinate string into integers
func parseCoordinates(coordStr string) []int {
	coords := strings.Split(coordStr, ",")
	if len(coords) != 2 {
		return []int{0, 0}
	}

	x, _ := strconv.Atoi(strings.TrimSpace(coords[0]))
	y, _ := strconv.Atoi(strings.TrimSpace(coords[1]))

	return []int{x, y}
}

// buildMouseDescription creates a human-readable description for mouse actions
func buildMouseDescription(action string, args []string) string {
	switch action {
	case "move-to":
		return fmt.Sprintf("Move mouse to element '%s'", args[0])
	case "move-by":
		coords := parseCoordinates(args[0])
		return fmt.Sprintf("Move mouse by (%d, %d)", coords[0], coords[1])
	case "move":
		coords := parseCoordinates(args[0])
		return fmt.Sprintf("Move mouse to (%d, %d)", coords[0], coords[1])
	case "down":
		return fmt.Sprintf("Press mouse button on '%s'", args[0])
	case "up":
		return fmt.Sprintf("Release mouse button on '%s'", args[0])
	case "enter":
		return fmt.Sprintf("Move mouse into '%s'", args[0])
	default:
		return fmt.Sprintf("Mouse %s", action)
	}
}

// buildSelectDescription creates a human-readable description for select actions
func buildSelectDescription(selectType string, args []string) string {
	switch selectType {
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
