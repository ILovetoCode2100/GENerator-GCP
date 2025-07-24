package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// CommandValidator validates and auto-corrects command syntax
type CommandValidator struct {
	// Command corrections
	corrections map[string]CommandCorrection

	// Deprecated commands
	deprecated map[string]string

	// Removed commands
	removed map[string]string

	// Flag validations
	flagValidations map[string][]string // command -> valid flags
}

// CommandCorrection defines how to fix a command
type CommandCorrection struct {
	OldSyntax    string
	NewSyntax    string
	ArgTransform func([]string) []string
}

// NewCommandValidator creates a new command validator
func NewCommandValidator() *CommandValidator {
	cv := &CommandValidator{
		corrections:     make(map[string]CommandCorrection),
		deprecated:      make(map[string]string),
		removed:         make(map[string]string),
		flagValidations: make(map[string][]string),
	}

	// Initialize corrections
	cv.initializeCorrections()
	cv.initializeDeprecated()
	cv.initializeRemoved()
	cv.initializeFlagValidations()

	return cv
}

// ValidateAndCorrect validates a command and returns corrected version
func (cv *CommandValidator) ValidateAndCorrect(cmd *cobra.Command, args []string) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}

	// Get full command path
	cmdPath := getCommandPath(cmd)

	// Check if command is removed
	if suggestion, exists := cv.removed[cmdPath]; exists {
		return nil, fmt.Errorf("command '%s' has been removed. %s", cmdPath, suggestion)
	}

	// Check if command is deprecated
	if newCmd, exists := cv.deprecated[cmdPath]; exists {
		fmt.Printf("Warning: '%s' is deprecated. Use '%s' instead.\n", cmdPath, newCmd)
		cmdPath = newCmd
	}

	// Apply corrections
	correctedArgs := cv.applyCorrections(cmdPath, args)

	// Validate flags
	if err := cv.validateFlags(cmd); err != nil {
		return nil, err
	}

	return correctedArgs, nil
}

// initializeCorrections sets up command corrections
func (cv *CommandValidator) initializeCorrections() {
	// Scroll commands - add hyphenation
	scrollCommands := []string{"top", "bottom", "element", "position", "by", "up", "down"}
	for _, scrollCmd := range scrollCommands {
		oldPath := fmt.Sprintf("step-navigate scroll %s", scrollCmd)
		newPath := fmt.Sprintf("step-navigate scroll-%s", scrollCmd)
		cv.corrections[oldPath] = CommandCorrection{
			OldSyntax: oldPath,
			NewSyntax: newPath,
			ArgTransform: func(args []string) []string {
				// Remove "scroll" and the subcommand from args
				if len(args) >= 2 && args[0] == "scroll" {
					return args[2:] // Keep remaining args
				}
				return args
			},
		}
	}

	// Switch tab command - fix argument order
	cv.corrections["step-window switch tab"] = CommandCorrection{
		OldSyntax: "step-window switch tab [checkpoint-id] next",
		NewSyntax: "step-window switch tab next [checkpoint-id]",
		ArgTransform: func(args []string) []string {
			// If args are: [checkpoint-id, "next", position]
			// Transform to: ["next", checkpoint-id, position]
			if len(args) >= 2 {
				// Check if second arg is "next" or "prev"
				if args[1] == "next" || args[1] == "prev" || args[1] == "previous" {
					// Swap first two args
					newArgs := make([]string, len(args))
					newArgs[0] = args[1]
					newArgs[1] = args[0]
					copy(newArgs[2:], args[2:])
					return newArgs
				}
			}
			return args
		},
	}

	// Store commands - simplify syntax
	cv.corrections["step-data store element-text"] = CommandCorrection{
		OldSyntax:    "step-data store element-text",
		NewSyntax:    "step-data store text",
		ArgTransform: nil, // No arg transformation needed
	}

	cv.corrections["step-data store element-attribute"] = CommandCorrection{
		OldSyntax:    "step-data store element-attribute",
		NewSyntax:    "step-data store attribute",
		ArgTransform: nil,
	}

	// Dialog commands - use hyphenated names
	dialogMappings := map[string]string{
		"alert accept":   "dismiss-alert",
		"alert dismiss":  "dismiss-alert",
		"confirm accept": "dismiss-confirm --accept",
		"confirm reject": "dismiss-confirm --reject",
		"prompt accept":  "dismiss-prompt --accept",
		"prompt reject":  "dismiss-prompt --reject",
	}

	for old, new := range dialogMappings {
		cv.corrections["step-dialog "+old] = CommandCorrection{
			OldSyntax: "step-dialog " + old,
			NewSyntax: "step-dialog " + new,
			ArgTransform: func(args []string) []string {
				// Remove the old subcommands
				if len(args) >= 2 && (args[0] == "alert" || args[0] == "confirm" || args[0] == "prompt") {
					return args[2:]
				}
				return args
			},
		}
	}

	// Mouse coordinate format
	cv.corrections["step-interact mouse move-by"] = CommandCorrection{
		OldSyntax: "step-interact mouse move-by [x y]",
		NewSyntax: "step-interact mouse move-by [x,y]",
		ArgTransform: func(args []string) []string {
			// Convert "100 200" to "100,200"
			if len(args) >= 2 {
				// Check if we have two numeric args
				if _, err1 := strconv.Atoi(args[0]); err1 == nil {
					if _, err2 := strconv.Atoi(args[1]); err2 == nil {
						// Combine into single coordinate
						newArgs := make([]string, 0, len(args)-1)
						newArgs = append(newArgs, fmt.Sprintf("%s,%s", args[0], args[1]))
						newArgs = append(newArgs, args[2:]...)
						return newArgs
					}
				}
			}
			return args
		},
	}

	// Wait time - convert seconds to milliseconds
	cv.corrections["step-wait time"] = CommandCorrection{
		OldSyntax: "step-wait time [seconds]",
		NewSyntax: "step-wait time [milliseconds]",
		ArgTransform: func(args []string) []string {
			if len(args) > 0 {
				// Check if the value is a small number (likely seconds)
				if val, err := strconv.Atoi(args[0]); err == nil && val < 1000 {
					// Convert to milliseconds
					args[0] = strconv.Itoa(val * 1000)
				}
			}
			return args
		},
	}
}

// initializeDeprecated sets up deprecated command mappings
func (cv *CommandValidator) initializeDeprecated() {
	cv.deprecated["step-misc add-comment"] = "step-misc comment"
	cv.deprecated["step-interact input"] = "step-interact write"
	cv.deprecated["step-interact type"] = "step-interact write"
}

// initializeRemoved sets up removed commands with suggestions
func (cv *CommandValidator) initializeRemoved() {
	cv.removed["step-navigate scroll-right"] = "Horizontal scrolling is not supported. Use scroll-by with x,y coordinates instead."
	cv.removed["step-navigate scroll-left"] = "Horizontal scrolling is not supported. Use scroll-by with negative x coordinate instead."
	cv.removed["step-navigate back"] = "Browser back navigation is not supported by the API."
	cv.removed["step-navigate forward"] = "Browser forward navigation is not supported by the API."
	cv.removed["step-navigate refresh"] = "Browser refresh is not supported by the API."
}

// initializeFlagValidations sets up valid flags per command
func (cv *CommandValidator) initializeFlagValidations() {
	// Click command doesn't support offset flags
	cv.flagValidations["step-interact click"] = []string{"clear", "delay"}

	// Write command supports these flags
	cv.flagValidations["step-interact write"] = []string{"clear", "delay", "append"}

	// Navigate supports these flags
	cv.flagValidations["step-navigate to"] = []string{"new-tab", "wait"}
}

// applyCorrections applies corrections to command arguments
func (cv *CommandValidator) applyCorrections(cmdPath string, args []string) []string {
	// Check for exact match
	if correction, exists := cv.corrections[cmdPath]; exists {
		if correction.ArgTransform != nil {
			return correction.ArgTransform(args)
		}
		return args
	}

	// Check for partial matches (for subcommands)
	for path, correction := range cv.corrections {
		if strings.HasPrefix(cmdPath+" "+strings.Join(args, " "), path) {
			if correction.ArgTransform != nil {
				return correction.ArgTransform(args)
			}
			return args
		}
	}

	return args
}

// validateFlags checks if command flags are valid
func (cv *CommandValidator) validateFlags(cmd *cobra.Command) error {
	cmdPath := getCommandPath(cmd)
	validFlags, hasValidation := cv.flagValidations[cmdPath]

	if !hasValidation {
		return nil // No validation rules for this command
	}

	// Check each flag that was set
	invalidFlags := []string{}
	cmd.Flags().Visit(func(flag *cobra.Flag) {
		if flag.Name == "help" || flag.Name == "config" || flag.Name == "output" || flag.Name == "verbose" {
			return // Skip global flags
		}

		found := false
		for _, valid := range validFlags {
			if flag.Name == valid {
				found = true
				break
			}
		}

		if !found {
			invalidFlags = append(invalidFlags, flag.Name)
		}
	})

	if len(invalidFlags) > 0 {
		return fmt.Errorf("invalid flags for %s: %s. Valid flags are: %s",
			cmdPath, strings.Join(invalidFlags, ", "), strings.Join(validFlags, ", "))
	}

	return nil
}

// getCommandPath returns the full command path
func getCommandPath(cmd *cobra.Command) string {
	parts := []string{}
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() != "" {
			parts = append([]string{c.Name()}, parts...)
		}
	}
	return strings.Join(parts, " ")
}

// ValidatorMiddleware creates a middleware function for cobra commands
func ValidatorMiddleware(validator *CommandValidator) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		correctedArgs, err := validator.ValidateAndCorrect(cmd, args)
		if err != nil {
			return err
		}

		// Update args if corrections were made
		if !equalStringSlices(args, correctedArgs) {
			fmt.Printf("Auto-corrected: %s\n", strings.Join(correctedArgs, " "))
			// Note: In actual implementation, you'd need to update the command's args
			// This might require modifying the cobra command structure
		}

		return nil
	}
}

// equalStringSlices checks if two string slices are equal
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// GetCorrectionSummary returns a summary of all corrections
func (cv *CommandValidator) GetCorrectionSummary() string {
	var summary strings.Builder

	summary.WriteString("Command Syntax Corrections:\n")
	summary.WriteString("==========================\n\n")

	for _, correction := range cv.corrections {
		summary.WriteString(fmt.Sprintf("OLD: %s\nNEW: %s\n\n", correction.OldSyntax, correction.NewSyntax))
	}

	summary.WriteString("\nDeprecated Commands:\n")
	summary.WriteString("===================\n\n")
	for old, new := range cv.deprecated {
		summary.WriteString(fmt.Sprintf("%s → %s\n", old, new))
	}

	summary.WriteString("\nRemoved Commands:\n")
	summary.WriteString("=================\n\n")
	for cmd, reason := range cv.removed {
		summary.WriteString(fmt.Sprintf("%s: %s\n", cmd, reason))
	}

	return summary.String()
}
