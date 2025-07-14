package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// LegacyCommandMapping maps old command names to new consolidated commands
var LegacyCommandMapping = map[string]struct {
	NewCommand   string
	SubCommand   string
	ArgTransform func(args []string) []string
}{
	// Assert commands
	"create-step-assert-equals":                {"assert", "equals", nil},
	"create-step-assert-not-equals":            {"assert", "not-equals", nil},
	"create-step-assert-exists":                {"assert", "exists", nil},
	"create-step-assert-not-exists":            {"assert", "not-exists", nil},
	"create-step-assert-checked":               {"assert", "checked", nil},
	"create-step-assert-selected":              {"assert", "selected", nil},
	"create-step-assert-variable":              {"assert", "variable", nil},
	"create-step-assert-greater-than":          {"assert", "gt", nil},
	"create-step-assert-greater-than-or-equal": {"assert", "gte", nil},
	"create-step-assert-less-than":             {"assert", "lt", nil},
	"create-step-assert-less-than-or-equal":    {"assert", "lte", nil},
	"create-step-assert-matches":               {"assert", "matches", nil},

	// Interact commands
	"create-step-click":        {"interact", "click", nil},
	"create-step-double-click": {"interact", "double-click", nil},
	"create-step-right-click":  {"interact", "right-click", nil},
	"create-step-hover":        {"interact", "hover", nil},
	"create-step-write":        {"interact", "write", nil},
	"create-step-key":          {"interact", "key", nil},

	// Navigate commands
	"create-step-navigate":        {"navigate", "url", nil},
	"create-step-scroll-position": {"navigate", "scroll-to", nil},
	"create-step-scroll-top":      {"navigate", "scroll-top", nil},
	"create-step-scroll-bottom":   {"navigate", "scroll-bottom", nil},
	"create-step-scroll-element":  {"navigate", "scroll-element", nil},

	// Window commands
	"create-step-window-resize":       {"window", "resize", nil},
	"create-step-switch-next-tab":     {"window", "switch-tab", transformTabArgs("next")},
	"create-step-switch-prev-tab":     {"window", "switch-tab", transformTabArgs("prev")},
	"create-step-switch-iframe":       {"window", "switch-frame", nil},
	"create-step-switch-parent-frame": {"window", "switch-frame", transformParentFrameArgs},

	// Mouse commands
	"create-step-mouse-move-to": {"mouse", "move-to", nil},
	"create-step-mouse-move-by": {"mouse", "move-by", nil},
	"create-step-mouse-move":    {"mouse", "move", nil},
	"create-step-mouse-down":    {"mouse", "down", nil},
	"create-step-mouse-up":      {"mouse", "up", nil},
	"create-step-mouse-enter":   {"mouse", "enter", nil},

	// Data commands
	"create-step-store-element-text":  {"data", "store-text", nil},
	"create-step-store-literal-value": {"data", "store-value", nil},
	"create-step-cookie-create":       {"data", "cookie-create", nil},
	"create-step-delete-cookie":       {"data", "cookie-delete", nil},
	"create-step-cookie-wipe-all":     {"data", "cookie-clear", nil},

	// Dialog commands
	"create-step-dismiss-alert":            {"dialog", "dismiss-alert", nil},
	"create-step-dismiss-confirm":          {"dialog", "dismiss-confirm", nil},
	"create-step-dismiss-prompt":           {"dialog", "dismiss-prompt", nil},
	"create-step-dismiss-prompt-with-text": {"dialog", "dismiss-prompt", nil},

	// Wait commands
	"create-step-wait-element":             {"wait", "element", nil},
	"create-step-wait-for-element-default": {"wait", "element", nil},
	"create-step-wait-for-element-timeout": {"wait", "element", nil},
	"create-step-wait-time":                {"wait", "time", nil},

	// File commands
	"create-step-upload":     {"file", "upload", nil},
	"create-step-upload-url": {"file", "upload-url", nil},

	// Select commands
	"create-step-pick":       {"select", "option", nil},
	"create-step-pick-index": {"select", "index", nil},
	"create-step-pick-last":  {"select", "last", nil},

	// Misc commands
	"create-step-comment":        {"misc", "comment", nil},
	"create-step-execute-script": {"misc", "execute", nil},
}

// transformTabArgs creates a function that adds the tab direction as first argument
func transformTabArgs(direction string) func([]string) []string {
	return func(args []string) []string {
		return append([]string{direction}, args...)
	}
}

// transformParentFrameArgs transforms parent frame arguments
func transformParentFrameArgs(args []string) []string {
	return append([]string{"parent"}, args...)
}

// LegacyUsageTracker tracks usage of legacy commands for migration insights
type LegacyUsageTracker struct {
	CommandUsage map[string]int
	LastSaved    time.Time
	FilePath     string
}

var usageTracker = &LegacyUsageTracker{
	CommandUsage: make(map[string]int),
	FilePath:     os.ExpandEnv("$HOME/.api-cli/legacy-usage.log"),
}

// CreateLegacyWrapper creates a wrapper command that translates old format to new
func CreateLegacyWrapper(oldCommand string) *cobra.Command {
	mapping, exists := LegacyCommandMapping[oldCommand]
	if !exists {
		return nil
	}

	// Extract the base description from the new command
	baseDescription := fmt.Sprintf("(DEPRECATED) %s - Use 'api-cli %s %s' instead",
		oldCommand, mapping.NewCommand, mapping.SubCommand)

	cmd := &cobra.Command{
		Use:        oldCommand,
		Short:      baseDescription,
		Hidden:     false, // Keep visible for backward compatibility
		Deprecated: fmt.Sprintf("use 'api-cli %s %s' instead", mapping.NewCommand, mapping.SubCommand),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Track usage
			trackLegacyUsage(oldCommand)

			// Show deprecation warning
			showDeprecationWarning(oldCommand, mapping.NewCommand, mapping.SubCommand)

			// Transform arguments if needed
			if mapping.ArgTransform != nil {
				args = mapping.ArgTransform(args)
			}

			// Get the root command
			root := cmd.Root()

			// Find the new consolidated command
			newCmd, _, err := root.Find([]string{mapping.NewCommand})
			if err != nil {
				return fmt.Errorf("failed to find new command '%s': %w", mapping.NewCommand, err)
			}

			// Build new command arguments
			newArgs := []string{mapping.SubCommand}
			newArgs = append(newArgs, args...)

			// Set up the new command with arguments
			newCmd.SetArgs(newArgs)

			// Execute the new command
			return newCmd.Execute()
		},
	}

	// Copy flags from the original command if they exist
	// This ensures backward compatibility with command-specific flags
	addLegacyFlags(cmd, oldCommand)

	return cmd
}

// showDeprecationWarning displays a formatted deprecation warning
func showDeprecationWarning(oldCommand, newCommand, subCommand string) {
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan)

	fmt.Fprintln(os.Stderr, "")
	yellow.Fprintf(os.Stderr, "⚠️  DEPRECATION WARNING\n")
	fmt.Fprintf(os.Stderr, "The command '%s' is deprecated and will be removed in a future version.\n", oldCommand)
	cyan.Fprintf(os.Stderr, "Please use: api-cli %s %s\n", newCommand, subCommand)
	fmt.Fprintln(os.Stderr, "")

	// Optional: Show migration command
	if os.Getenv("API_CLI_SHOW_MIGRATION_HELP") == "1" {
		fmt.Fprintf(os.Stderr, "To automatically update your scripts, run:\n")
		cyan.Fprintf(os.Stderr, "  ./scripts/migrate-commands.sh -a your-script.sh\n")
		fmt.Fprintln(os.Stderr, "")
	}
}

// trackLegacyUsage records usage of legacy commands
func trackLegacyUsage(command string) {
	usageTracker.CommandUsage[command]++

	// Save to file periodically (every 10 uses or 5 minutes)
	totalUses := 0
	for _, count := range usageTracker.CommandUsage {
		totalUses += count
	}

	if totalUses%10 == 0 || time.Since(usageTracker.LastSaved) > 5*time.Minute {
		saveLegacyUsage()
	}
}

// saveLegacyUsage persists usage statistics to file
func saveLegacyUsage() {
	// Create directory if it doesn't exist
	dir := os.ExpandEnv("$HOME/.api-cli")
	os.MkdirAll(dir, 0755)

	// Append to log file
	file, err := os.OpenFile(usageTracker.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return // Silently fail - don't interrupt user's work
	}
	defer file.Close()

	// Write usage data
	timestamp := time.Now().Format(time.RFC3339)
	for cmd, count := range usageTracker.CommandUsage {
		fmt.Fprintf(file, "%s,%s,%d\n", timestamp, cmd, count)
	}

	// Reset counters
	usageTracker.CommandUsage = make(map[string]int)
	usageTracker.LastSaved = time.Now()
}

// addLegacyFlags adds command-specific flags for backward compatibility
func addLegacyFlags(cmd *cobra.Command, oldCommand string) {
	// Add common flags that might be used by legacy commands
	switch {
	case strings.Contains(oldCommand, "assert"):
		// Assert commands typically don't have special flags

	case strings.Contains(oldCommand, "click"):
		cmd.Flags().String("variable", "", "Store result in variable")
		cmd.Flags().String("position", "", "Click position (e.g., TOP_LEFT)")
		cmd.Flags().String("element-type", "", "Element type")

	case strings.Contains(oldCommand, "navigate"):
		cmd.Flags().Bool("new-tab", false, "Open in new tab")

	case strings.Contains(oldCommand, "write"):
		cmd.Flags().String("variable", "", "Use variable value")

	case strings.Contains(oldCommand, "key"):
		cmd.Flags().String("target", "", "Target element selector")
	}

	// Add global flags
	cmd.Flags().Int("checkpoint", 0, "Checkpoint ID (overrides session)")
	cmd.Flags().String("output", "human", "Output format: human, json, yaml, ai")
}

// GetLegacyCommands returns all legacy commands that should be registered
func GetLegacyCommands() []*cobra.Command {
	var commands []*cobra.Command

	for oldCommand := range LegacyCommandMapping {
		if cmd := CreateLegacyWrapper(oldCommand); cmd != nil {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// GenerateMigrationReport generates a report of legacy command usage
func GenerateMigrationReport() string {
	report := strings.Builder{}
	report.WriteString("=== Legacy Command Usage Report ===\n\n")

	// Read usage log
	logData, err := os.ReadFile(usageTracker.FilePath)
	if err != nil {
		report.WriteString("No usage data available.\n")
		return report.String()
	}

	// Parse and aggregate usage data
	usage := make(map[string]int)
	lines := strings.Split(string(logData), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			cmd := parts[1]
			count, _ := fmt.Sscanf(parts[2], "%d", new(int))
			usage[cmd] += count
		}
	}

	// Generate report
	if len(usage) == 0 {
		report.WriteString("No legacy commands have been used.\n")
	} else {
		report.WriteString("Legacy Command Usage:\n")
		for cmd, count := range usage {
			if mapping, exists := LegacyCommandMapping[cmd]; exists {
				report.WriteString(fmt.Sprintf("  %-40s: %d uses → use '%s %s'\n",
					cmd, count, mapping.NewCommand, mapping.SubCommand))
			}
		}
	}

	return report.String()
}
