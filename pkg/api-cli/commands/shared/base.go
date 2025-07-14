package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// StepCommand is the interface that all step commands must implement
type StepCommand interface {
	// GetStepType returns the Virtuoso API step type (e.g., "NAVIGATE", "CLICK")
	GetStepType() string

	// GetDescription returns a human-readable description of the step
	GetDescription() string

	// ValidateArgs validates the command arguments
	ValidateArgs(args []string) error

	// BuildRequest builds the API request from arguments
	BuildRequest(args []string, options map[string]interface{}) (*StepRequest, error)

	// FormatResult formats the API response for output
	FormatResult(result *StepResult, format string) (string, error)
}

// BaseCommand provides common functionality for all step commands
type BaseCommand struct {
	Client       *client.Client
	CheckpointID string
	Position     int
	OutputFormat string
}

// NewBaseCommand creates a new base command instance
func NewBaseCommand() *BaseCommand {
	return &BaseCommand{
		OutputFormat: "human",
	}
}

// Init initializes the base command with client and config
func (bc *BaseCommand) Init(cmd *cobra.Command) error {
	// Get API configuration from environment
	token := os.Getenv("VIRTUOSO_API_TOKEN")
	if token == "" {
		return fmt.Errorf("VIRTUOSO_API_TOKEN environment variable is required")
	}

	// Get API base URL from environment
	baseURL := os.Getenv("VIRTUOSO_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api-app2.virtuoso.qa/api"
	}

	// Create client
	bc.Client = client.NewClientDirect(baseURL, token)

	// Get output format flag
	format, _ := cmd.Flags().GetString("output")
	if format != "" {
		bc.OutputFormat = format
	}

	return nil
}

// ResolveCheckpointAndPosition resolves checkpoint ID and position from arguments
// Supports both modern (session-based) and legacy (explicit checkpoint) formats
func (bc *BaseCommand) ResolveCheckpointAndPosition(args []string, requiresArgs int) ([]string, error) {
	// Check if we have a checkpoint ID from session (environment variable)
	sessionCheckpoint := os.Getenv("VIRTUOSO_SESSION_CHECKPOINT")

	if len(args) < requiresArgs {
		return nil, fmt.Errorf("insufficient arguments")
	}

	// Try to parse position from last argument
	lastArg := args[len(args)-1]
	position, isPosition := ParsePosition(lastArg)

	if isPosition {
		bc.Position = position
		args = args[:len(args)-1] // Remove position from args
	}

	// Now check if we have enough arguments after removing position
	if len(args) < requiresArgs {
		// Try modern format with session checkpoint
		if sessionCheckpoint != "" {
			bc.CheckpointID = sessionCheckpoint
			if !isPosition {
				bc.Position = -1 // Default position
			}
			return args, nil
		}
		return nil, fmt.Errorf("insufficient arguments: need checkpoint ID or active session")
	}

	// Check if first argument looks like a checkpoint ID (legacy format)
	if strings.HasPrefix(args[0], "cp_") || IsNumeric(args[0]) {
		bc.CheckpointID = args[0]
		args = args[1:] // Remove checkpoint ID from args

		// If we didn't find position earlier and have one more arg, check it
		if !isPosition && len(args) > requiresArgs-1 {
			if pos, ok := ParsePosition(args[len(args)-1]); ok {
				bc.Position = pos
				args = args[:len(args)-1]
			}
		}
	} else if sessionCheckpoint != "" {
		// Modern format - use session checkpoint
		bc.CheckpointID = sessionCheckpoint
	} else {
		return nil, fmt.Errorf("no checkpoint ID provided and no active session")
	}

	if !isPosition {
		bc.Position = -1 // Default position
	}

	return args, nil
}

// ParsePosition parses a position argument
func ParsePosition(arg string) (int, bool) {
	// Handle numeric positions
	if IsNumeric(arg) {
		pos := 0
		fmt.Sscanf(arg, "%d", &pos)
		return pos, true
	}

	// Handle "last" keyword
	if strings.ToLower(arg) == "last" {
		return -1, true
	}

	return 0, false
}

// IsNumeric checks if a string is numeric
func IsNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// FormatOutput formats the output based on the specified format
func (bc *BaseCommand) FormatOutput(result interface{}, format string) (string, error) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "yaml":
		// Using yaml.v3 for YAML formatting
		yamlData, err := yaml.Marshal(result)
		if err != nil {
			return "", err
		}
		return string(yamlData), nil

	case "ai":
		// AI-friendly format
		if stepResult, ok := result.(*StepResult); ok {
			return bc.FormatAI(stepResult), nil
		}
		// Fallback to JSON for other types
		return bc.FormatOutput(result, "json")

	case "human":
		fallthrough
	default:
		if stepResult, ok := result.(*StepResult); ok {
			return bc.FormatHuman(stepResult), nil
		}
		// Fallback to string representation
		return fmt.Sprintf("%+v", result), nil
	}
}

// FormatHuman formats a step result for human reading
func (bc *BaseCommand) FormatHuman(result *StepResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Step created successfully!\n"))
	sb.WriteString(fmt.Sprintf("ID: %s\n", result.ID))
	sb.WriteString(fmt.Sprintf("Type: %s\n", result.Type))
	sb.WriteString(fmt.Sprintf("Position: %d\n", result.Position))

	if result.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", result.Description))
	}

	if result.Selector != "" {
		sb.WriteString(fmt.Sprintf("Selector: %s\n", result.Selector))
	}

	if len(result.Meta) > 0 {
		sb.WriteString("Meta:\n")
		for k, v := range result.Meta {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	return sb.String()
}

// FormatAI formats a step result for AI consumption
func (bc *BaseCommand) FormatAI(result *StepResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("STEP_CREATED type=%s id=%s position=%d",
		result.Type, result.ID, result.Position))

	if result.Selector != "" {
		sb.WriteString(fmt.Sprintf(" selector=%q", result.Selector))
	}

	if result.Description != "" {
		sb.WriteString(fmt.Sprintf(" description=%q", result.Description))
	}

	for k, v := range result.Meta {
		sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
	}

	return sb.String()
}

// ValidateSelector validates a CSS selector
func ValidateSelector(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector cannot be empty")
	}

	// Basic validation - could be expanded
	if strings.ContainsAny(selector, "\n\r\t") {
		return fmt.Errorf("selector contains invalid whitespace characters")
	}

	return nil
}

// ValidateURL validates a URL
func ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	return nil
}

// ParseKeyValue parses a key=value string
func ParseKeyValue(s string) (string, string, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format, expected key=value")
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

// MergeMaps merges two maps, with values from m2 overriding m1
func MergeMaps(m1, m2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range m1 {
		result[k] = v
	}

	for k, v := range m2 {
		result[k] = v
	}

	return result
}
