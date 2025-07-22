package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// WaitCommand implements the wait command group using BaseCommand pattern
type WaitCommand struct {
	*BaseCommand
	waitType string
}

// waitConfig contains configuration for each wait type
type waitConfig struct {
	stepType     string
	description  string
	usage        string
	examples     []string
	requiredArgs int
	hasTimeout   bool
	buildMeta    func(args []string, timeout int) map[string]interface{}
	parseStep    func(args []string, timeout int) string
}

// waitConfigs maps wait types to their configurations
var waitConfigs = map[string]waitConfig{
	"element": {
		stepType:    "WAIT_ELEMENT",
		description: "Wait for an element to be visible",
		usage:       "wait element [checkpoint-id] <selector> [position] [--timeout ms]",
		examples: []string{
			`api-cli wait element cp_12345 "Login button" 1`,
			`api-cli wait element "Login button"  # Uses session context`,
			`api-cli wait element "Success message" --timeout 5000  # Wait up to 5 seconds`,
			`api-cli wait element cp_12345 "#submit-btn" 2 --timeout 10000  # 10 second timeout`,
		},
		requiredArgs: 1,
		hasTimeout:   true,
		buildMeta: func(args []string, timeout int) map[string]interface{} {
			meta := map[string]interface{}{
				"selector": args[0],
			}
			if timeout > 0 {
				meta["timeout_ms"] = timeout
			}
			return meta
		},
		parseStep: func(args []string, timeout int) string {
			if timeout > 0 {
				return fmt.Sprintf("wait until %s appears (timeout: %dms)", args[0], timeout)
			}
			return fmt.Sprintf("wait until %s appears", args[0])
		},
	},
	"element-not-visible": {
		stepType:    "WAIT_ELEMENT_NOT_VISIBLE",
		description: "Wait for an element to disappear",
		usage:       "wait element-not-visible [checkpoint-id] <selector> [position] [--timeout ms]",
		examples: []string{
			`api-cli wait element-not-visible cp_12345 "Loading spinner" 1`,
			`api-cli wait element-not-visible "Loading spinner"  # Uses session context`,
			`api-cli wait element-not-visible "Modal overlay" --timeout 5000`,
			`api-cli wait element-not-visible cp_12345 ".progress-bar" 2 --timeout 10000`,
		},
		requiredArgs: 1,
		hasTimeout:   true,
		buildMeta: func(args []string, timeout int) map[string]interface{} {
			meta := map[string]interface{}{
				"selector": args[0],
			}
			if timeout > 0 {
				meta["timeout_ms"] = timeout
			}
			return meta
		},
		parseStep: func(args []string, timeout int) string {
			if timeout > 0 {
				return fmt.Sprintf("wait until %s disappears (timeout: %dms)", args[0], timeout)
			}
			return fmt.Sprintf("wait until %s disappears", args[0])
		},
	},
	"time": {
		stepType:    "WAIT_TIME",
		description: "Wait for a specified time in milliseconds",
		usage:       "wait time [checkpoint-id] <milliseconds> [position]",
		examples: []string{
			`api-cli wait time cp_12345 1000 1  # Wait 1 second at position 1`,
			`api-cli wait time 1000  # Wait 1 second, uses session context`,
			`api-cli wait time 500  # Wait 500ms, auto-increment position`,
			`api-cli wait time cp_12345 3000 2  # Wait 3 seconds at position 2`,
		},
		requiredArgs: 1,
		hasTimeout:   false,
		buildMeta: func(args []string, timeout int) map[string]interface{} {
			ms, _ := strconv.Atoi(args[0])
			return map[string]interface{}{
				"milliseconds": ms,
				"seconds":      float64(ms) / 1000.0,
			}
		},
		parseStep: func(args []string, timeout int) string {
			ms, _ := strconv.Atoi(args[0])
			if ms >= 1000 {
				seconds := float64(ms) / 1000.0
				return fmt.Sprintf("wait %.1f seconds", seconds)
			}
			return fmt.Sprintf("wait %d milliseconds", ms)
		},
	},
}

// newWaitCmd creates the new wait command using BaseCommand pattern
func newWaitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Create wait steps in checkpoints",
		Long: `Create various types of wait steps in checkpoints.

This command uses the standardized positional argument pattern:
- Optional checkpoint ID as first argument (falls back to session context)
- Required wait arguments
- Optional position as last argument (auto-increments if not specified)

Available wait types:
  - element: Wait for element to be visible (with optional timeout)
  - element-not-visible: Wait for element to disappear (with optional timeout)
  - time: Wait for specified time in milliseconds`,
		Example: `  # Wait for element to appear (with explicit checkpoint)
  api-cli wait element cp_12345 "Login button" 1

  # Wait for element (using session context)
  api-cli wait element "Login button"

  # Wait for element with custom timeout
  api-cli wait element "Success message" --timeout 5000

  # Wait for element to disappear
  api-cli wait element-not-visible "Loading spinner"

  # Wait for 2 seconds
  api-cli wait time 2000`,
	}

	// Add subcommands for each wait type
	for waitType, config := range waitConfigs {
		cmd.AddCommand(newWaitV2SubCmd(waitType, config))
	}

	return cmd
}

// newWaitV2SubCmd creates a subcommand for a specific wait type
func newWaitV2SubCmd(waitType string, config waitConfig) *cobra.Command {
	var timeoutFlag int

	cmd := &cobra.Command{
		Use:   waitType + " " + extractWaitUsageArgs(config.usage),
		Short: config.description,
		Long: fmt.Sprintf(`%s

%s

Examples:
%s`, config.description, config.usage, strings.Join(config.examples, "\n")),
		RunE: func(cmd *cobra.Command, args []string) error {
			wc := &WaitCommand{
				BaseCommand: NewBaseCommand(),
				waitType:    waitType,
			}
			return wc.Execute(cmd, args, config, timeoutFlag)
		},
	}

	// Add timeout flag for commands that support it
	if config.hasTimeout {
		cmd.Flags().IntVar(&timeoutFlag, "timeout", 0, "Custom timeout in milliseconds (default uses Virtuoso's default)")
	}

	return cmd
}

// extractWaitUsageArgs extracts the arguments portion from the usage string
func extractWaitUsageArgs(usage string) string {
	parts := strings.Fields(usage)
	if len(parts) > 2 {
		// Skip "wait" and subcommand
		argsStart := 2
		result := []string{}
		for i := argsStart; i < len(parts); i++ {
			// Skip flags
			if strings.HasPrefix(parts[i], "[--") {
				continue
			}
			result = append(result, parts[i])
		}
		return strings.Join(result, " ")
	}
	return ""
}

// Execute runs the wait command
func (wc *WaitCommand) Execute(cmd *cobra.Command, args []string, config waitConfig, timeoutFlag int) error {
	// Initialize base command
	if err := wc.Init(cmd); err != nil {
		return err
	}

	// Resolve checkpoint and position
	remainingArgs, err := wc.ResolveCheckpointAndPosition(args, config.requiredArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve arguments: %w", err)
	}

	// Validate we have the required number of arguments
	if len(remainingArgs) != config.requiredArgs {
		return fmt.Errorf("expected %d arguments, got %d", config.requiredArgs, len(remainingArgs))
	}

	// Validate arguments based on wait type
	if err := wc.validateArgs(wc.waitType, remainingArgs, timeoutFlag); err != nil {
		return err
	}

	// Build request metadata
	meta := config.buildMeta(remainingArgs, timeoutFlag)

	// Create the step
	stepResult, err := wc.createWaitStep(config.stepType, remainingArgs, timeoutFlag)
	if err != nil {
		// Provide more specific error messages based on the error type
		switch {
		case strings.Contains(err.Error(), "timed out"):
			return fmt.Errorf("API request timed out while creating %s step. The operation took longer than 30 seconds", config.stepType)
		case strings.Contains(err.Error(), "invalid checkpoint ID"):
			return fmt.Errorf("invalid checkpoint ID '%s'. Please ensure you're using a valid numeric checkpoint ID", wc.CheckpointID)
		case strings.Contains(err.Error(), "connection refused"):
			return fmt.Errorf("cannot connect to Virtuoso API. Please check your network connection and API URL configuration")
		case strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized"):
			return fmt.Errorf("authentication failed. Please check your API token in the configuration")
		case strings.Contains(err.Error(), "404"):
			return fmt.Errorf("checkpoint not found. Please verify checkpoint ID '%s' exists", wc.CheckpointID)
		default:
			return fmt.Errorf("failed to create %s step: %w", config.stepType, err)
		}
	}

	// Add parsed step description
	stepResult.Description = config.parseStep(remainingArgs, timeoutFlag)
	stepResult.Meta = meta

	// Save config if position was auto-incremented
	if wc.Position == -1 && cfg != nil && cfg.Session.AutoIncrementPos {
		if cfg.Session.NextPosition == 0 {
			cfg.Session.NextPosition = 1
		} else {
			cfg.Session.NextPosition++
		}
		cfg.SaveConfig()
	}

	// Format and output the result
	output, err := wc.FormatOutput(stepResult, wc.OutputFormat)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

// validateArgs validates arguments for a specific wait type
func (wc *WaitCommand) validateArgs(waitType string, args []string, timeout int) error {
	switch waitType {
	case "element", "element-not-visible":
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("selector cannot be empty")
		}
		if timeout < 0 {
			return fmt.Errorf("timeout cannot be negative")
		}
		if timeout > 60000 {
			return fmt.Errorf("timeout cannot exceed 60000ms (60 seconds)")
		}
	case "time":
		if len(args) < 1 || args[0] == "" {
			return fmt.Errorf("time cannot be empty")
		}
		ms, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid time value: %w", err)
		}
		if ms < 0 {
			return fmt.Errorf("time cannot be negative")
		}
		if ms > 60000 {
			return fmt.Errorf("time cannot exceed 60000ms (60 seconds)")
		}
	}
	return nil
}

// createWaitStep creates a wait step via the API
func (wc *WaitCommand) createWaitStep(stepType string, args []string, timeout int) (*StepResult, error) {
	// Create context with timeout for API operation
	ctx, cancel := wc.CommandContext()
	defer cancel()

	// Convert checkpoint ID from string to int
	checkpointID, err := strconv.Atoi(wc.CheckpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid checkpoint ID: %s", wc.CheckpointID)
	}

	// If position is -1, use nil to let API auto-assign
	position := wc.Position
	if position == -1 {
		// For wait commands, we'll use a high position number to append at end
		position = 9999
	}

	// Call the appropriate API method based on wait type
	var stepID int
	switch stepType {
	case "WAIT_ELEMENT":
		if timeout > 0 {
			// Use custom timeout version
			stepID, err = wc.createWaitElementWithTimeout(ctx, checkpointID, args[0], timeout, position)
			if err != nil {
				return nil, fmt.Errorf("failed to create wait element step with timeout: %w", err)
			}
		} else {
			// Use default timeout version
			stepID, err = wc.createWaitElement(ctx, checkpointID, args[0], position)
			if err != nil {
				return nil, fmt.Errorf("failed to create wait element step: %w", err)
			}
		}
	case "WAIT_ELEMENT_NOT_VISIBLE":
		stepID, err = wc.createWaitElementNotVisible(ctx, checkpointID, args[0], timeout, position)
		if err != nil {
			return nil, fmt.Errorf("failed to create wait element not visible step: %w", err)
		}
	case "WAIT_TIME":
		// Convert milliseconds to seconds for the API (if it expects seconds)
		ms, _ := strconv.Atoi(args[0])
		seconds := ms / 1000
		if ms%1000 != 0 {
			// If there are remaining milliseconds, round up to next second
			seconds++
		}
		stepID, err = wc.createWaitTime(ctx, checkpointID, seconds, position)
		if err != nil {
			return nil, fmt.Errorf("failed to create wait time step: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported wait type: %s", stepType)
	}

	// Check if context was cancelled
	if ctx.Err() != nil {
		return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
	}

	// Build the result
	result := &StepResult{
		ID:           strconv.Itoa(stepID),
		CheckpointID: wc.CheckpointID,
		Type:         stepType,
		Position:     position,
	}

	// Add selector for element-based waits
	if stepType == "WAIT_ELEMENT" || stepType == "WAIT_ELEMENT_NOT_VISIBLE" {
		result.Selector = args[0]
	}

	// Add value for time wait
	if stepType == "WAIT_TIME" {
		result.Value = args[0]
	}

	return result, nil
}

// createWaitElement creates a wait element step with context
func (wc *WaitCommand) createWaitElement(ctx context.Context, checkpointID int, element string, position int) (int, error) {
	// Since the client doesn't have context-aware methods yet, we'll call the regular method
	// In the future, this should call wc.Client.CreateWaitElementStepWithContext
	done := make(chan struct{})
	var stepID int
	var err error

	go func() {
		stepID, err = wc.Client.CreateWaitElementStep(checkpointID, element, position)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("wait element creation timed out")
	case <-done:
		return stepID, err
	}
}

// createWaitElementWithTimeout creates a wait element step with custom timeout and context
func (wc *WaitCommand) createWaitElementWithTimeout(ctx context.Context, checkpointID int, element string, timeout int, position int) (int, error) {
	// Since the client doesn't have context-aware methods yet, we'll call the regular method
	// In the future, this should call wc.Client.CreateStepWaitForElementTimeoutWithContext
	done := make(chan struct{})
	var stepID int
	var err error

	go func() {
		stepID, err = wc.Client.CreateStepWaitForElementTimeout(checkpointID, element, timeout, position)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("wait element with timeout creation timed out")
	case <-done:
		return stepID, err
	}
}

// createWaitElementNotVisible creates a wait element not visible step with context
func (wc *WaitCommand) createWaitElementNotVisible(ctx context.Context, checkpointID int, element string, timeout int, position int) (int, error) {
	// Since the client doesn't have context-aware methods yet, we'll call the regular method
	// In the future, this should call wc.Client.CreateStepWaitForElementNotVisibleWithContext
	done := make(chan struct{})
	var stepID int
	var err error

	go func() {
		stepID, err = wc.Client.CreateStepWaitForElementNotVisible(checkpointID, element, timeout, position)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("wait element not visible creation timed out")
	case <-done:
		return stepID, err
	}
}

// createWaitTime creates a wait time step with context
func (wc *WaitCommand) createWaitTime(ctx context.Context, checkpointID int, seconds int, position int) (int, error) {
	// Since the client doesn't have context-aware methods yet, we'll call the regular method
	// In the future, this should call wc.Client.CreateWaitTimeStepWithContext
	done := make(chan struct{})
	var stepID int
	var err error

	go func() {
		stepID, err = wc.Client.CreateWaitTimeStep(checkpointID, seconds, position)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("wait time creation timed out")
	case <-done:
		return stepID, err
	}
}
