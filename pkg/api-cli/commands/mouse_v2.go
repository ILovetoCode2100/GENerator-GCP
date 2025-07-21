package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// MouseCommand implements the mouse command group using BaseCommand pattern
type MouseCommand struct {
	*BaseCommand
	action string
}

// mouseConfig contains configuration for each mouse action
type mouseConfig struct {
	stepType     string
	description  string
	usage        string
	examples     []string
	requiredArgs int
	buildMeta    func(args []string) map[string]interface{}
}

// mouseConfigs maps mouse actions to their configurations
var mouseConfigs = map[string]mouseConfig{
	"move-to": {
		stepType:    "MOUSE_MOVE_TO",
		description: "Move mouse to element center",
		usage:       "mouse move-to [checkpoint-id] <selector> [position]",
		examples: []string{
			`api-cli mouse move-to cp_12345 "button.submit" 1`,
			`api-cli mouse move-to "#target-element"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
			}
		},
	},
	"move-by": {
		stepType:    "MOUSE_MOVE_BY",
		description: "Move mouse by relative offset",
		usage:       "mouse move-by [checkpoint-id] <x,y> [position]",
		examples: []string{
			`api-cli mouse move-by cp_12345 "100,50" 1`,
			`api-cli mouse move-by "-50,25"  # Uses session context`,
			`api-cli mouse move-by cp_12345 -- -50,-25 2  # Negative values`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			coords := parseCoordinates(args[0])
			return map[string]interface{}{
				"dx":       coords[0],
				"dy":       coords[1],
				"relative": true,
			}
		},
	},
	"move": {
		stepType:    "MOUSE_MOVE",
		description: "Move mouse to absolute coordinates",
		usage:       "mouse move [checkpoint-id] <x,y> [position]",
		examples: []string{
			`api-cli mouse move cp_12345 "500,300" 1`,
			`api-cli mouse move "100,200"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			coords := parseCoordinates(args[0])
			return map[string]interface{}{
				"x":        coords[0],
				"y":        coords[1],
				"absolute": true,
			}
		},
	},
	"down": {
		stepType:    "MOUSE_DOWN",
		description: "Press mouse button down",
		usage:       "mouse down [checkpoint-id] <selector> [position]",
		examples: []string{
			`api-cli mouse down cp_12345 "button.submit" 1`,
			`api-cli mouse down "#drag-handle"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"button":   "left",
			}
		},
	},
	"up": {
		stepType:    "MOUSE_UP",
		description: "Release mouse button",
		usage:       "mouse up [checkpoint-id] <selector> [position]",
		examples: []string{
			`api-cli mouse up cp_12345 "button.submit" 1`,
			`api-cli mouse up "#drop-zone"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
				"button":   "left",
			}
		},
	},
	"enter": {
		stepType:    "MOUSE_ENTER",
		description: "Move mouse into element",
		usage:       "mouse enter [checkpoint-id] <selector> [position]",
		examples: []string{
			`api-cli mouse enter cp_12345 ".dropdown-trigger" 1`,
			`api-cli mouse enter "#tooltip-target"  # Uses session context`,
		},
		requiredArgs: 1,
		buildMeta: func(args []string) map[string]interface{} {
			return map[string]interface{}{
				"selector": args[0],
			}
		},
	},
}

// newMouseV2Cmd creates the new mouse command using BaseCommand pattern
func newMouseV2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mouse",
		Short: "Perform mouse actions",
		Long: `Perform various mouse operations including clicking, moving, and hovering.

This command uses the standardized positional argument pattern:
- Optional checkpoint ID as first argument (falls back to session context)
- Required mouse action arguments
- Optional position as last argument (auto-increments if not specified)

Available mouse actions:
  - move-to: Move mouse to element center
  - move-by: Move mouse by relative offset
  - move: Move mouse to absolute coordinates
  - down: Press mouse button down
  - up: Release mouse button
  - enter: Move mouse into element`,
		Example: `  # Move mouse to element (with explicit checkpoint)
  api-cli mouse move-to cp_12345 "button.submit" 1

  # Move mouse to element (using session context)
  api-cli mouse move-to ".hover-menu"

  # Move mouse by offset
  api-cli mouse move-by "100,50"
  api-cli mouse move-by -- -50,-25  # Negative values

  # Move mouse to coordinates
  api-cli mouse move "500,300"

  # Press and release mouse button
  api-cli mouse down "#drag-handle"
  api-cli mouse up "#drop-zone"`,
	}

	// Add subcommands for each mouse action
	for action, config := range mouseConfigs {
		cmd.AddCommand(newMouseV2SubCmd(action, config))
	}

	return cmd
}

// newMouseV2SubCmd creates a subcommand for a specific mouse action
func newMouseV2SubCmd(action string, config mouseConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   action + " " + extractMouseUsageArgs(config.usage),
		Short: config.description,
		Long: fmt.Sprintf(`%s

%s

Examples:
%s`, config.description, config.usage, strings.Join(config.examples, "\n")),
		RunE: func(cmd *cobra.Command, args []string) error {
			mc := &MouseCommand{
				BaseCommand: NewBaseCommand(),
				action:      action,
			}
			return mc.Execute(cmd, args, config)
		},
	}

	return cmd
}

// extractMouseUsageArgs extracts the arguments portion from the usage string
func extractMouseUsageArgs(usage string) string {
	parts := strings.Fields(usage)
	if len(parts) > 2 {
		// Skip "mouse" and subcommand
		return strings.Join(parts[2:], " ")
	}
	return ""
}

// Execute runs the mouse command
func (mc *MouseCommand) Execute(cmd *cobra.Command, args []string, config mouseConfig) error {
	// Initialize base command
	if err := mc.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	remainingArgs, err := mc.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate we have the required number of arguments
	if len(remainingArgs) != config.requiredArgs {
		return fmt.Errorf("expected %d arguments, got %d", config.requiredArgs, len(remainingArgs))
	}

	// Validate arguments based on action type
	if err := mc.validateArgs(remainingArgs); err != nil {
		return err
	}

	// Build request metadata
	meta := config.buildMeta(remainingArgs)

	// Create the step
	stepResult, err := mc.createMouseStep(config.stepType, meta, remainingArgs)
	if err != nil {
		return fmt.Errorf("failed to create %s step: %w", config.stepType, err)
	}

	// Format and output the result
	output, err := mc.FormatOutput(stepResult, mc.OutputFormat)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// validateArgs validates arguments for the mouse action
func (mc *MouseCommand) validateArgs(args []string) error {
	switch mc.action {
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

// createMouseStep creates a mouse step via the API
func (mc *MouseCommand) createMouseStep(stepType string, meta map[string]interface{}, args []string) (*StepResult, error) {
	// Convert checkpoint ID from string to int
	checkpointID, err := strconv.Atoi(mc.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint ID: %s", mc.CheckpointID)
	}

	// Build the request based on step type
	var stepID int

	// Use the client to create the appropriate step
	switch stepType {
	case "MOUSE_DOWN":
		stepID, err = mc.Client.CreateStepMouseDown(checkpointID, meta["selector"].(string), mc.Position)
	case "MOUSE_UP":
		stepID, err = mc.Client.CreateStepMouseUp(checkpointID, meta["selector"].(string), mc.Position)
	case "MOUSE_ENTER":
		stepID, err = mc.Client.CreateStepMouseEnter(checkpointID, meta["selector"].(string), mc.Position)
	case "MOUSE_MOVE":
		// For move to element, using the existing mouse move method
		if selector, ok := meta["selector"].(string); ok {
			stepID, err = mc.Client.CreateStepMouseMove(checkpointID, selector, mc.Position)
		} else {
			// For absolute coordinates
			x := meta["x"].(int)
			y := meta["y"].(int)
			stepID, err = mc.Client.CreateStepMouseMoveTo(checkpointID, x, y, mc.Position)
		}
	case "MOUSE_MOVE_BY":
		dx := meta["dx"].(int)
		dy := meta["dy"].(int)
		stepID, err = mc.Client.CreateStepMouseMoveBy(checkpointID, dx, dy, mc.Position)
	case "MOUSE_MOVE_TO":
		// Move to element center
		stepID, err = mc.Client.CreateStepMouseMove(checkpointID, meta["selector"].(string), mc.Position)
	default:
		return nil, fmt.Errorf("unknown mouse action type: %s", stepType)
	}

	if err != nil {
		return nil, err
	}

	// Build the result
	result := &StepResult{
		ID:           fmt.Sprintf("%d", stepID),
		CheckpointID: mc.CheckpointID,
		Type:         stepType,
		Position:     mc.Position,
		Description:  mc.buildDescription(stepType, meta),
		Selector:     mc.extractSelector(meta),
		Meta:         meta,
	}

	// Save session state if position was auto-incremented
	if mc.Position == -1 && cfg.Session.AutoIncrementPos {
		if err := cfg.SaveConfig(); err != nil {
			// Don't fail the command, just warn
			// Note: In production, this warning would be sent to stderr
		}
	}

	return result, nil
}

// buildDescription creates a human-readable description for the step
func (mc *MouseCommand) buildDescription(stepType string, meta map[string]interface{}) string {
	switch stepType {
	case "MOUSE_DOWN":
		return fmt.Sprintf("mouse down on \"%s\"", meta["selector"])
	case "MOUSE_UP":
		return fmt.Sprintf("mouse up on \"%s\"", meta["selector"])
	case "MOUSE_ENTER":
		return fmt.Sprintf("mouse enter \"%s\"", meta["selector"])
	case "MOUSE_MOVE":
		if selector, ok := meta["selector"].(string); ok {
			return fmt.Sprintf("move mouse to \"%s\"", selector)
		}
		return fmt.Sprintf("move mouse to (%d, %d)", meta["x"], meta["y"])
	case "MOUSE_MOVE_BY":
		return fmt.Sprintf("move mouse by (%d, %d)", meta["dx"], meta["dy"])
	case "MOUSE_MOVE_TO":
		return fmt.Sprintf("move mouse to \"%s\"", meta["selector"])
	default:
		return stepType
	}
}

// extractSelector extracts the selector from metadata if present
func (mc *MouseCommand) extractSelector(meta map[string]interface{}) string {
	if selector, ok := meta["selector"].(string); ok {
		return selector
	}
	return ""
}
