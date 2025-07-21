package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CompatibilityTransformer handles backward compatibility for old --checkpoint flag syntax
type CompatibilityTransformer struct {
	// Map of command patterns that need transformation
	transformPatterns map[string]TransformFunc
}

// TransformFunc defines how to transform arguments for a specific command pattern
type TransformFunc func(args []string, checkpointID string, position string) ([]string, error)

// NewCompatibilityTransformer creates a new compatibility transformer
func NewCompatibilityTransformer() *CompatibilityTransformer {
	ct := &CompatibilityTransformer{
		transformPatterns: make(map[string]TransformFunc),
	}
	ct.registerPatterns()
	return ct
}

// registerPatterns registers all command transformation patterns
func (ct *CompatibilityTransformer) registerPatterns() {
	// Assert commands: assert [type] [selector] --checkpoint ID
	// New format: assert [type] [checkpoint] [selector] [position]
	assertTransform := func(args []string, checkpointID string, position string) ([]string, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("assert command requires type and selector")
		}
		// Insert checkpoint ID after command type
		newArgs := []string{args[0], checkpointID}
		newArgs = append(newArgs, args[1:]...)
		if position != "" {
			newArgs = append(newArgs, position)
		}
		return newArgs, nil
	}

	// Register assert command patterns
	for _, cmd := range []string{"assert exists", "assert not-exists", "assert equals",
		"assert not-equals", "assert checked", "assert selected", "assert variable",
		"assert gt", "assert gte", "assert lt", "assert lte", "assert matches"} {
		ct.transformPatterns[cmd] = assertTransform
	}

	// Wait commands: wait [type] [args] --checkpoint ID
	// New format: wait [type] [checkpoint] [args] [position]
	waitTransform := func(args []string, checkpointID string, position string) ([]string, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("wait command requires type")
		}
		// Insert checkpoint ID after command type
		newArgs := []string{args[0], checkpointID}
		if len(args) > 1 {
			newArgs = append(newArgs, args[1:]...)
		}
		if position != "" {
			newArgs = append(newArgs, position)
		}
		return newArgs, nil
	}

	ct.transformPatterns["wait element"] = waitTransform
	ct.transformPatterns["wait element-not-visible"] = waitTransform
	ct.transformPatterns["wait time"] = waitTransform

	// Mouse commands: mouse [action] [args] --checkpoint ID
	// New format: mouse [action] [checkpoint] [args] [position]
	mouseTransform := func(args []string, checkpointID string, position string) ([]string, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("mouse command requires action")
		}
		// Insert checkpoint ID after action
		newArgs := []string{args[0], checkpointID}
		if len(args) > 1 {
			newArgs = append(newArgs, args[1:]...)
		}
		if position != "" {
			newArgs = append(newArgs, position)
		}
		return newArgs, nil
	}

	for _, cmd := range []string{"mouse move-to", "mouse move-by", "mouse move",
		"mouse down", "mouse up", "mouse enter"} {
		ct.transformPatterns[cmd] = mouseTransform
	}

	// Data commands: data [action] [args] [position] --checkpoint ID
	// New format: data [action] [checkpoint] [args] [position]
	dataTransform := func(args []string, checkpointID string, position string) ([]string, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("data command requires action and arguments")
		}

		// Find where position might be in the args
		positionIndex := -1
		for i := 1; i < len(args); i++ {
			if _, err := strconv.Atoi(args[i]); err == nil {
				// This could be the position
				positionIndex = i
			}
		}

		// Build new args with checkpoint after action
		newArgs := []string{args[0], checkpointID}

		if positionIndex > 0 {
			// Add args before position
			newArgs = append(newArgs, args[1:positionIndex]...)
			// Add position
			newArgs = append(newArgs, args[positionIndex])
		} else {
			// No position found in original args
			newArgs = append(newArgs, args[1:]...)
			if position != "" {
				newArgs = append(newArgs, position)
			}
		}

		return newArgs, nil
	}

	ct.transformPatterns["data store-text"] = dataTransform
	ct.transformPatterns["data store-value"] = dataTransform
	ct.transformPatterns["data store-attribute"] = dataTransform
	ct.transformPatterns["data cookie-create"] = dataTransform
	ct.transformPatterns["data cookie-delete"] = dataTransform
	ct.transformPatterns["data cookie-clear"] = dataTransform

	// Window commands: window [action] [args] [position] --checkpoint ID
	// New format: window [action] [checkpoint] [args] [position]
	windowTransform := func(args []string, checkpointID string, position string) ([]string, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("window command requires action")
		}

		// Find position in args if present
		positionIndex := -1
		for i := 1; i < len(args); i++ {
			if _, err := strconv.Atoi(args[i]); err == nil {
				positionIndex = i
				break
			}
		}

		// Build new args
		newArgs := []string{args[0], checkpointID}

		if positionIndex > 0 {
			// Add args before position
			newArgs = append(newArgs, args[1:positionIndex]...)
			// Add position
			newArgs = append(newArgs, args[positionIndex])
		} else {
			// No position in original
			if len(args) > 1 {
				newArgs = append(newArgs, args[1:]...)
			}
			if position != "" {
				newArgs = append(newArgs, position)
			}
		}

		return newArgs, nil
	}

	ct.transformPatterns["window resize"] = windowTransform
	ct.transformPatterns["window maximize"] = windowTransform
	ct.transformPatterns["window switch-tab"] = windowTransform
	ct.transformPatterns["window switch-iframe"] = windowTransform
	ct.transformPatterns["window switch-parent-frame"] = windowTransform
}

// TransformCommand checks if a command needs transformation and applies it
func (ct *CompatibilityTransformer) TransformCommand(cmd string, args []string) ([]string, bool, error) {
	// Look for --checkpoint flag
	checkpointID, newArgs, position := extractCheckpointFlag(args)
	if checkpointID == "" {
		// No --checkpoint flag found, no transformation needed
		return args, false, nil
	}

	// Build full command for pattern matching
	fullCmd := cmd
	if len(newArgs) > 0 && isSubcommand(newArgs[0]) {
		fullCmd = cmd + " " + newArgs[0]
		newArgs = newArgs[1:]
	}

	// Find transformation pattern
	transform, exists := ct.transformPatterns[fullCmd]
	if !exists {
		// No specific transformation pattern, use generic
		transform = genericTransform
	}

	// Apply transformation
	transformedArgs, err := transform(newArgs, checkpointID, position)
	if err != nil {
		return nil, false, err
	}

	// Add subcommand back if it was extracted
	if strings.Contains(fullCmd, " ") {
		parts := strings.SplitN(fullCmd, " ", 2)
		transformedArgs = append([]string{parts[1]}, transformedArgs...)
	}

	return transformedArgs, true, nil
}

// extractCheckpointFlag extracts --checkpoint flag and its value from args
func extractCheckpointFlag(args []string) (string, []string, string) {
	var checkpointID string
	var position string
	var newArgs []string

	i := 0
	for i < len(args) {
		if args[i] == "--checkpoint" && i+1 < len(args) {
			checkpointID = args[i+1]
			i += 2 // Skip flag and value
		} else if args[i] == "--position" && i+1 < len(args) {
			position = args[i+1]
			i += 2 // Skip flag and value
		} else if strings.HasPrefix(args[i], "--checkpoint=") {
			checkpointID = strings.TrimPrefix(args[i], "--checkpoint=")
			i++
		} else if strings.HasPrefix(args[i], "--position=") {
			position = strings.TrimPrefix(args[i], "--position=")
			i++
		} else {
			newArgs = append(newArgs, args[i])
			i++
		}
	}

	return checkpointID, newArgs, position
}

// isSubcommand checks if a string is likely a subcommand
func isSubcommand(s string) bool {
	// List of known subcommands
	subcommands := []string{
		"exists", "not-exists", "equals", "not-equals", "checked", "selected",
		"variable", "gt", "gte", "lt", "lte", "matches",
		"element", "element-not-visible", "time",
		"move-to", "move-by", "move", "down", "up", "enter",
		"store-text", "store-value", "store-attribute",
		"cookie-create", "cookie-delete", "cookie-clear",
		"resize", "maximize", "switch-tab", "switch-iframe", "switch-parent-frame",
	}

	for _, sub := range subcommands {
		if s == sub {
			return true
		}
	}
	return false
}

// genericTransform provides a generic transformation for commands without specific patterns
func genericTransform(args []string, checkpointID string, position string) ([]string, error) {
	// Generic pattern: insert checkpoint as first argument
	newArgs := []string{checkpointID}
	newArgs = append(newArgs, args...)
	if position != "" {
		newArgs = append(newArgs, position)
	}
	return newArgs, nil
}

// IsUsingOldSyntax checks if command arguments use old --checkpoint syntax
func IsUsingOldSyntax(args []string) bool {
	for i, arg := range args {
		if arg == "--checkpoint" && i+1 < len(args) {
			return true
		}
		if strings.HasPrefix(arg, "--checkpoint=") {
			return true
		}
	}
	return false
}

// ShowDeprecationWarning prints a deprecation warning for old syntax usage
func ShowDeprecationWarning(cmd string, oldArgs []string, newArgs []string) {
	fmt.Fprintf(os.Stderr, "\n⚠️  DEPRECATION WARNING: You are using the old --checkpoint flag syntax.\n")
	fmt.Fprintf(os.Stderr, "   Old syntax: %s %s\n", cmd, strings.Join(oldArgs, " "))
	fmt.Fprintf(os.Stderr, "   New syntax: %s %s\n", cmd, strings.Join(newArgs, " "))
	fmt.Fprintf(os.Stderr, "   The --checkpoint flag syntax will be removed in v4.0 (July 2025).\n")
	fmt.Fprintf(os.Stderr, "   See MIGRATION_GUIDE.md for migration instructions.\n\n")
}
