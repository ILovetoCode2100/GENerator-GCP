package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

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
// This structure supports AI-friendly output formats and test structure generation
type BaseCommand struct {
	Client       *client.Client
	CheckpointID string
	Position     int
	OutputFormat string // Supports: human, json, yaml, ai (with contextual test structure)
}

// NewBaseCommand creates a new base command instance
func NewBaseCommand() *BaseCommand {
	return &BaseCommand{
		OutputFormat: "human",
	}
}

// CommandContext creates a context with timeout for API operations
func (bc *BaseCommand) CommandContext() (context.Context, context.CancelFunc) {
	// Default timeout of 30 seconds for API operations
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// Init initializes the base command with client and config
func (bc *BaseCommand) Init(cmd *cobra.Command) error {
	// Check if global config is set
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Check if API token is configured
	if cfg.API.AuthToken == "" {
		return fmt.Errorf("API token not configured. Please set api.auth_token in config or VIRTUOSO_API_TOKEN environment variable")
	}

	// Create client using config
	bc.Client = client.NewClientDirect(cfg.API.BaseURL, cfg.API.AuthToken)

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
	// Check if we have a checkpoint ID from session
	var sessionCheckpoint string
	if cfg != nil && cfg.Session.CurrentCheckpointID != nil {
		sessionCheckpoint = strconv.Itoa(*cfg.Session.CurrentCheckpointID)
	}

	if len(args) < requiresArgs {
		return nil, fmt.Errorf("insufficient arguments")
	}

	// Try to parse position from last argument
	var isPosition bool
	if len(args) > 0 {
		lastArg := args[len(args)-1]
		pos, isPos := ParsePosition(lastArg)
		if isPos {
			bc.Position = pos
			isPosition = true
			args = args[:len(args)-1] // Remove position from args
		}
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
	if len(args) > 0 && (strings.HasPrefix(args[0], "cp_") || IsNumeric(args[0])) {
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
// Supports multiple formats for different use cases:
// - json: Structured data for programmatic parsing
// - yaml: Configuration-friendly format for test definitions
// - ai: Enhanced output with context, next steps, and test structure
// - human: Default readable format for manual inspection
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
		// AI-friendly format with contextual information for test generation
		// Includes: command result, test context, suggested next steps, and journey structure
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

// FormatAI formats a step result for AI consumption with enhanced context
// Returns a structured format that includes:
// - Command execution result
// - Test context (checkpoint, journey, position)
// - Suggested next steps based on the command type
// - Current test structure information
func (bc *BaseCommand) FormatAI(result *StepResult) string {
	// Create AI-optimized output structure
	aiOutput := map[string]interface{}{
		"command": result.Type,
		"result":  "success",
		"message": fmt.Sprintf("Step created: %s", result.Description),
		"step_details": map[string]interface{}{
			"id":       result.ID,
			"type":     result.Type,
			"position": result.Position,
			"selector": result.Selector,
			"meta":     result.Meta,
		},
		"context": map[string]interface{}{
			"checkpoint_id": bc.CheckpointID,
			"position":      result.Position,
		},
		"next_steps": bc.suggestNextSteps(result.Type),
		"test_structure": map[string]interface{}{
			"current_position": result.Position,
		},
	}

	// Convert to JSON for structured output
	data, _ := json.MarshalIndent(aiOutput, "", "  ")
	return string(data)
}

// suggestNextSteps provides intelligent next step suggestions based on command type
func (bc *BaseCommand) suggestNextSteps(stepType string) []string {
	// AI-driven suggestions for test flow continuity
	suggestions := map[string][]string{
		"NAVIGATE": {
			"wait element 'body'",
			"assert exists '.main-content'",
			"interact click 'first visible button'",
		},
		"CLICK": {
			"wait element '.loading' --not-exists",
			"assert exists '.success-message'",
			"assert not-exists '.error-message'",
		},
		"WRITE": {
			"interact key 'Tab'",
			"interact click 'Submit'",
			"assert equals 'input value' 'expected value'",
		},
		"ASSERT_EXISTS": {
			"interact click 'element'",
			"data store element-text 'element' 'variableName'",
			"assert equals 'element' 'expected text'",
		},
	}

	if steps, ok := suggestions[stepType]; ok {
		return steps
	}

	// Default suggestions
	return []string{
		"wait time 1000",
		"assert exists '.next-element'",
		"interact click '.continue-button'",
	}
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
