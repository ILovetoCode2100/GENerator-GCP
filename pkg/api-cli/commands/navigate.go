package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// NavigateCmd creates the navigate command with subcommands
func NavigateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "navigate",
		Short: "Navigate and scroll within the browser",
		Long: `Navigate to URLs and scroll within pages using various methods.

This command consolidates URL navigation and scrolling actions.`,
	}

	// Add subcommands
	cmd.AddCommand(navigateToSubCmd())
	cmd.AddCommand(scrollTopSubCmd())
	cmd.AddCommand(scrollBottomSubCmd())
	cmd.AddCommand(scrollElementSubCmd())
	cmd.AddCommand(scrollPositionSubCmd())
	cmd.AddCommand(scrollBySubCmd())
	cmd.AddCommand(scrollUpSubCmd())
	cmd.AddCommand(scrollDownSubCmd())

	return cmd
}

// navigateToSubCmd creates the navigate-to subcommand
func navigateToSubCmd() *cobra.Command {
	var (
		newTab      bool
		waitForLoad bool
	)

	cmd := &cobra.Command{
		Use:   "to [checkpoint-id] <url> [position]",
		Short: "Navigate to a URL",
		Long: `Navigate the browser to a specified URL.

Examples:
  # Using session context (modern)
  api-cli navigate to "https://example.com"
  api-cli navigate to "https://example.com/page" --new-tab

  # Using explicit checkpoint (legacy)
  api-cli navigate to cp_12345 "https://example.com" 1`,
		Aliases: []string{"url", "goto"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigation(cmd, args, "navigate-to", map[string]interface{}{
				"newTab":      newTab,
				"waitForLoad": waitForLoad,
			})
		},
	}

	cmd.Flags().BoolVar(&newTab, "new-tab", false, "Open URL in new tab")
	cmd.Flags().BoolVar(&waitForLoad, "wait", true, "Wait for page to load")

	return cmd
}

// scrollTopSubCmd creates the scroll-top subcommand
func scrollTopSubCmd() *cobra.Command {
	var smooth bool

	cmd := &cobra.Command{
		Use:   "scroll-top [checkpoint-id] [position]",
		Short: "Scroll to the top of the page",
		Long: `Scroll the page to the very top.

Examples:
  # Using session context
  api-cli navigate scroll-top
  api-cli navigate scroll-top --smooth

  # Using explicit checkpoint
  api-cli navigate scroll-top cp_12345 1`,
		Aliases: []string{"top"},
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigation(cmd, args, "scroll-top", map[string]interface{}{
				"smooth": smooth,
			})
		},
	}

	cmd.Flags().BoolVar(&smooth, "smooth", false, "Use smooth scrolling")

	return cmd
}

// scrollBottomSubCmd creates the scroll-bottom subcommand
func scrollBottomSubCmd() *cobra.Command {
	var smooth bool

	cmd := &cobra.Command{
		Use:   "scroll-bottom [checkpoint-id] [position]",
		Short: "Scroll to the bottom of the page",
		Long: `Scroll the page to the very bottom.

Examples:
  # Using session context
  api-cli navigate scroll-bottom
  api-cli navigate scroll-bottom --smooth

  # Using explicit checkpoint
  api-cli navigate scroll-bottom cp_12345 1`,
		Aliases: []string{"bottom"},
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigation(cmd, args, "scroll-bottom", map[string]interface{}{
				"smooth": smooth,
			})
		},
	}

	cmd.Flags().BoolVar(&smooth, "smooth", false, "Use smooth scrolling")

	return cmd
}

// scrollElementSubCmd creates the scroll-element subcommand
func scrollElementSubCmd() *cobra.Command {
	var (
		smooth       bool
		scrollToView bool
		block        string
		inline       string
	)

	cmd := &cobra.Command{
		Use:   "scroll-element [checkpoint-id] <selector> [position]",
		Short: "Scroll an element into view",
		Long: `Scroll to make a specific element visible in the viewport.

Examples:
  # Using session context
  api-cli navigate scroll-element "#target-section"
  api-cli navigate scroll-element ".important-content" --smooth
  api-cli navigate scroll-element "#form-submit" --block center

  # Using explicit checkpoint
  api-cli navigate scroll-element cp_12345 "#target-section" 1`,
		Aliases: []string{"element", "to-element"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigation(cmd, args, "scroll-element", map[string]interface{}{
				"smooth":       smooth,
				"scrollToView": scrollToView,
				"block":        block,
				"inline":       inline,
			})
		},
	}

	cmd.Flags().BoolVar(&smooth, "smooth", false, "Use smooth scrolling")
	cmd.Flags().BoolVar(&scrollToView, "into-view", true, "Scroll element into view")
	cmd.Flags().StringVar(&block, "block", "start", "Vertical alignment (start, center, end, nearest)")
	cmd.Flags().StringVar(&inline, "inline", "nearest", "Horizontal alignment (start, center, end, nearest)")

	return cmd
}

// scrollPositionSubCmd creates the scroll-position subcommand
func scrollPositionSubCmd() *cobra.Command {
	var (
		x      int
		y      int
		smooth bool
	)

	cmd := &cobra.Command{
		Use:   "scroll-position [checkpoint-id] <x,y> [position]",
		Short: "Scroll to specific coordinates",
		Long: `Scroll the page to specific X,Y coordinates.

Examples:
  # Using session context
  api-cli navigate scroll-position "0,500"
  api-cli navigate scroll-position "100,1000" --smooth
  api-cli navigate scroll-position --x 0 --y 500

  # Using explicit checkpoint
  api-cli navigate scroll-position cp_12345 "0,500" 1`,
		Aliases: []string{"position", "to-position", "xy"},
		Args:    cobra.RangeArgs(0, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse coordinates if provided as argument
			if len(args) > 0 && strings.Contains(args[0], ",") {
				coords := strings.Split(args[0], ",")
				if len(coords) == 2 {
					if xVal, err := strconv.Atoi(strings.TrimSpace(coords[0])); err == nil {
						x = xVal
					}
					if yVal, err := strconv.Atoi(strings.TrimSpace(coords[1])); err == nil {
						y = yVal
					}
					// Remove the coordinate argument
					args = append(args[:0], args[1:]...)
				}
			}

			return runNavigation(cmd, args, "scroll-position", map[string]interface{}{
				"x":      x,
				"y":      y,
				"smooth": smooth,
			})
		},
	}

	cmd.Flags().IntVar(&x, "x", 0, "X coordinate")
	cmd.Flags().IntVar(&y, "y", 0, "Y coordinate")
	cmd.Flags().BoolVar(&smooth, "smooth", false, "Use smooth scrolling")

	return cmd
}

// scrollBySubCmd creates the scroll-by subcommand
func scrollBySubCmd() *cobra.Command {
	var (
		x      int
		y      int
		smooth bool
	)

	cmd := &cobra.Command{
		Use:   "scroll-by [checkpoint-id] <x,y> [position]",
		Short: "Scroll by relative offset",
		Long: `Scroll the page by a relative X,Y offset from current position.

Examples:
  # Using session context
  api-cli navigate scroll-by "0,500"    # Scroll down 500px
  api-cli navigate scroll-by "-100,0"  # Scroll left 100px
  api-cli navigate scroll-by --x 0 --y -500  # Scroll up 500px

  # Using explicit checkpoint
  api-cli navigate scroll-by cp_12345 "0,500" 1`,
		Aliases: []string{"by", "offset"},
		Args:    cobra.RangeArgs(0, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse coordinates if provided as argument
			if len(args) > 0 && strings.Contains(args[0], ",") {
				coords := strings.Split(args[0], ",")
				if len(coords) == 2 {
					if xVal, err := strconv.Atoi(strings.TrimSpace(coords[0])); err == nil {
						x = xVal
					}
					if yVal, err := strconv.Atoi(strings.TrimSpace(coords[1])); err == nil {
						y = yVal
					}
					// Remove the coordinate argument
					args = append(args[:0], args[1:]...)
				}
			}

			return runNavigation(cmd, args, "scroll-by", map[string]interface{}{
				"x":      x,
				"y":      y,
				"smooth": smooth,
			})
		},
	}

	cmd.Flags().IntVar(&x, "x", 0, "X offset (negative for left)")
	cmd.Flags().IntVar(&y, "y", 0, "Y offset (negative for up)")
	cmd.Flags().BoolVar(&smooth, "smooth", false, "Use smooth scrolling")

	return cmd
}

// scrollUpSubCmd creates the scroll-up subcommand
func scrollUpSubCmd() *cobra.Command {
	var smooth bool

	cmd := &cobra.Command{
		Use:   "scroll-up [checkpoint-id] [position]",
		Short: "Scroll up by one viewport height",
		Long: `Scroll the page up by one viewport height.

Examples:
  # Using session context
  api-cli navigate scroll-up
  api-cli navigate scroll-up --smooth

  # Using explicit checkpoint
  api-cli navigate scroll-up cp_12345 1`,
		Aliases: []string{"up"},
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigation(cmd, args, "scroll-up", map[string]interface{}{
				"smooth": smooth,
			})
		},
	}

	cmd.Flags().BoolVar(&smooth, "smooth", false, "Use smooth scrolling")

	return cmd
}

// scrollDownSubCmd creates the scroll-down subcommand
func scrollDownSubCmd() *cobra.Command {
	var smooth bool

	cmd := &cobra.Command{
		Use:   "scroll-down [checkpoint-id] [position]",
		Short: "Scroll down by one viewport height",
		Long: `Scroll the page down by one viewport height.

Examples:
  # Using session context
  api-cli navigate scroll-down
  api-cli navigate scroll-down --smooth

  # Using explicit checkpoint
  api-cli navigate scroll-down cp_12345 1`,
		Aliases: []string{"down"},
		Args:    cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigation(cmd, args, "scroll-down", map[string]interface{}{
				"smooth": smooth,
			})
		},
	}

	cmd.Flags().BoolVar(&smooth, "smooth", false, "Use smooth scrolling")

	return cmd
}

// runNavigation executes a navigation command
func runNavigation(cmd *cobra.Command, args []string, action string, options map[string]interface{}) error {
	base := NewBaseCommand()
	if err := base.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	var err error
	requiredArgs := 0 // Most scroll commands don't require args
	if action == "navigate-to" {
		requiredArgs = 1 // URL required
	} else if action == "scroll-element" {
		requiredArgs = 1 // Selector required
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
	case "navigate-to":
		stepID, err = executeNavigateToAction(base.Client, checkpointID, args[0], base.Position, options)
	case "scroll-top":
		stepID, err = executeScrollTopAction(base.Client, checkpointID, base.Position, options)
	case "scroll-bottom":
		stepID, err = executeScrollBottomAction(base.Client, checkpointID, base.Position, options)
	case "scroll-element":
		stepID, err = executeScrollElementAction(base.Client, checkpointID, args[0], base.Position, options)
	case "scroll-position":
		stepID, err = executeScrollPositionAction(base.Client, checkpointID, base.Position, options)
	case "scroll-by":
		stepID, err = executeScrollByAction(base.Client, checkpointID, base.Position, options)
	case "scroll-up":
		stepID, err = executeScrollUpAction(base.Client, checkpointID, base.Position, options)
	case "scroll-down":
		stepID, err = executeScrollDownAction(base.Client, checkpointID, base.Position, options)
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

// executeNavigateToAction executes a navigate action using the client
func executeNavigateToAction(c *client.Client, checkpointID int, url string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateURL(url); err != nil {
		return 0, err
	}

	newTab, _ := options["newTab"].(bool)
	return c.CreateStepNavigate(checkpointID, url, newTab, position)
}

// executeScrollTopAction executes a scroll-to-top action using the client
func executeScrollTopAction(c *client.Client, checkpointID int, position int, options map[string]interface{}) (int, error) {
	return c.CreateStepScrollToTop(checkpointID, position)
}

// executeScrollBottomAction executes a scroll-to-bottom action using the client
func executeScrollBottomAction(c *client.Client, checkpointID int, position int, options map[string]interface{}) (int, error) {
	// Use the dedicated scroll-to-bottom method
	return c.CreateStepScrollBottom(checkpointID, position)
}

// executeScrollElementAction executes a scroll-to-element action using the client
func executeScrollElementAction(c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	return c.CreateStepScrollElement(checkpointID, selector, position)
}

// executeScrollPositionAction executes a scroll-to-position action using the client
func executeScrollPositionAction(c *client.Client, checkpointID int, position int, options map[string]interface{}) (int, error) {
	x, _ := options["x"].(int)
	y, _ := options["y"].(int)

	return c.CreateStepScrollToPosition(checkpointID, x, y, position)
}

// executeScrollByAction executes a scroll-by-offset action using the client
func executeScrollByAction(c *client.Client, checkpointID int, position int, options map[string]interface{}) (int, error) {
	x, _ := options["x"].(int)
	y, _ := options["y"].(int)

	return c.CreateStepScrollByOffset(checkpointID, x, y, position)
}

// executeScrollUpAction executes a scroll-up action using the client
func executeScrollUpAction(c *client.Client, checkpointID int, position int, options map[string]interface{}) (int, error) {
	// Scroll up by one viewport height (negative Y value)
	return c.CreateStepScrollByOffset(checkpointID, 0, -1000, position) // Using -1000 as approx viewport height
}

// executeScrollDownAction executes a scroll-down action using the client
func executeScrollDownAction(c *client.Client, checkpointID int, position int, options map[string]interface{}) (int, error) {
	// Scroll down by one viewport height (positive Y value)
	return c.CreateStepScrollByOffset(checkpointID, 0, 1000, position) // Using 1000 as approx viewport height
}
