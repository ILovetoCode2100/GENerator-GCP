package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

// mouseAction represents the type of mouse action
type mouseAction string

const (
	mouseDown   mouseAction = "down"
	mouseUp     mouseAction = "up"
	mouseEnter  mouseAction = "enter"
	mouseMove   mouseAction = "move"
	mouseMoveBy mouseAction = "move-by"
	mouseMoveTo mouseAction = "move-to"
)

// mouseCommandInfo contains metadata about each mouse action
type mouseCommandInfo struct {
	stepType    string
	description string
	usage       string
	examples    []string
	argsCount   []int // Valid argument counts (excluding position)
	parseStep   func(args []string) string
}

// mouseCommands maps mouse actions to their metadata
var mouseCommands = map[mouseAction]mouseCommandInfo{
	mouseDown: {
		stepType:    "MOUSE_DOWN",
		description: "Press mouse button down on element",
		usage:       "mouse down SELECTOR [POSITION]",
		examples: []string{
			`api-cli mouse down "button.submit" 1`,
			`api-cli mouse down "#drag-handle"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("mouse down on \"%s\"", args[0])
		},
	},
	mouseUp: {
		stepType:    "MOUSE_UP",
		description: "Release mouse button on element",
		usage:       "mouse up SELECTOR [POSITION]",
		examples: []string{
			`api-cli mouse up "button.submit" 1`,
			`api-cli mouse up "#drop-zone"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("mouse up on \"%s\"", args[0])
		},
	},
	mouseEnter: {
		stepType:    "MOUSE_ENTER",
		description: "Move mouse to hover over element",
		usage:       "mouse enter SELECTOR [POSITION]",
		examples: []string{
			`api-cli mouse enter ".dropdown-trigger" 1`,
			`api-cli mouse enter "#tooltip-target"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("mouse enter \"%s\"", args[0])
		},
	},
	mouseMove: {
		stepType:    "MOUSE_MOVE",
		description: "Move mouse to element",
		usage:       "mouse move SELECTOR [POSITION]",
		examples: []string{
			`api-cli mouse move ".hover-menu" 1`,
			`api-cli mouse move "#target-element"  # Auto-increment position`,
		},
		argsCount: []int{1},
		parseStep: func(args []string) string {
			return fmt.Sprintf("move mouse to \"%s\"", args[0])
		},
	},
	mouseMoveBy: {
		stepType:    "MOUSE_MOVE_BY",
		description: "Move mouse by relative offset",
		usage:       "mouse move-by DX DY [POSITION]",
		examples: []string{
			`api-cli mouse move-by 100 50 1`,
			`api-cli mouse move-by -50 -25  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("move mouse by (%s, %s)", args[0], args[1])
		},
	},
	mouseMoveTo: {
		stepType:    "MOUSE_MOVE_TO",
		description: "Move mouse to absolute coordinates",
		usage:       "mouse move-to X Y [POSITION]",
		examples: []string{
			`api-cli mouse move-to 500 300 1`,
			`api-cli mouse move-to 100 200  # Auto-increment position`,
		},
		argsCount: []int{2},
		parseStep: func(args []string) string {
			return fmt.Sprintf("move mouse to (%s, %s)", args[0], args[1])
		},
	},
}

// newMouseCmd creates the consolidated mouse command with subcommands
func newMouseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mouse",
		Short: "Perform mouse actions",
		Long: `Perform various mouse operations including clicking, moving, and hovering.

This command consolidates all mouse-related operations:
  - Mouse button actions (down, up)
  - Mouse movement (absolute, relative, to element)
  - Mouse hover/enter actions`,
		Example: `  # Press mouse button down
  api-cli mouse down "#drag-handle" 1

  # Release mouse button
  api-cli mouse up "#drop-zone"

  # Move mouse to element
  api-cli mouse move ".hover-menu"

  # Move mouse by offset
  api-cli mouse move-by 100 50

  # Move mouse to coordinates
  api-cli mouse move-to 500 300`,
	}

	// Add subcommands for each mouse action
	for action, info := range mouseCommands {
		cmd.AddCommand(newMouseSubCmd(action, info))
	}

	return cmd
}

// newMouseSubCmd creates a subcommand for a specific mouse action
func newMouseSubCmd(action mouseAction, info mouseCommandInfo) *cobra.Command {
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
			return runMouseCommand(action, info, args, checkpointFlag)
		},
	}

	addCheckpointFlag(cmd, &checkpointFlag)

	return cmd
}

// runMouseCommand executes the mouse command logic
func runMouseCommand(action mouseAction, info mouseCommandInfo, args []string, checkpointFlag int) error {
	// Validate arguments based on action type
	if err := validateMouseArgs(action, args); err != nil {
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
	stepID, err := callMouseAPI(apiClient, action, ctx, args)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", info.stepType, err)
	}

	// Save config if position was auto-incremented
	saveStepContext(ctx)

	// Build extra data for output
	extra := buildMouseExtraData(action, args)

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

// validateMouseArgs validates arguments for a specific mouse action
func validateMouseArgs(action mouseAction, args []string) error {
	switch action {
	case mouseDown, mouseUp, mouseEnter, mouseMove:
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
	case mouseMoveBy, mouseMoveTo:
		if len(args) < 2 {
			return fmt.Errorf("both X and Y coordinates are required")
		}
		// Validate numeric values
		if _, err := strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("X coordinate must be a number")
		}
		if _, err := strconv.Atoi(args[1]); err != nil {
			return fmt.Errorf("Y coordinate must be a number")
		}
	}
	return nil
}

// callMouseAPI calls the appropriate client API method for the mouse action
func callMouseAPI(apiClient *client.Client, action mouseAction, ctx *StepContext, args []string) (int, error) {
	switch action {
	case mouseDown:
		return apiClient.CreateStepMouseDown(ctx.CheckpointID, args[0], ctx.Position)
	case mouseUp:
		return apiClient.CreateStepMouseUp(ctx.CheckpointID, args[0], ctx.Position)
	case mouseEnter:
		return apiClient.CreateStepMouseEnter(ctx.CheckpointID, args[0], ctx.Position)
	case mouseMove:
		return apiClient.CreateStepMouseMove(ctx.CheckpointID, args[0], ctx.Position)
	case mouseMoveBy:
		x, _ := strconv.Atoi(args[0])
		y, _ := strconv.Atoi(args[1])
		return apiClient.CreateStepMouseMoveBy(ctx.CheckpointID, x, y, ctx.Position)
	case mouseMoveTo:
		x, _ := strconv.Atoi(args[0])
		y, _ := strconv.Atoi(args[1])
		return apiClient.CreateStepMouseMoveTo(ctx.CheckpointID, x, y, ctx.Position)
	default:
		return 0, fmt.Errorf("unsupported mouse action: %s", action)
	}
}

// buildMouseExtraData builds the extra data map for output based on mouse action
func buildMouseExtraData(action mouseAction, args []string) map[string]interface{} {
	extra := make(map[string]interface{})

	switch action {
	case mouseDown, mouseUp, mouseEnter, mouseMove:
		extra["selector"] = args[0]
	case mouseMoveBy:
		x, _ := strconv.Atoi(args[0])
		y, _ := strconv.Atoi(args[1])
		extra["dx"] = x
		extra["dy"] = y
		extra["relative"] = true
	case mouseMoveTo:
		x, _ := strconv.Atoi(args[0])
		y, _ := strconv.Atoi(args[1])
		extra["x"] = x
		extra["y"] = y
		extra["absolute"] = true
	}

	return extra
}
