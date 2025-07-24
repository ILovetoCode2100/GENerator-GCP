package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
	"github.com/spf13/cobra"
)

// ========================================
// Constants
// ========================================

const (
	// Default timeout for API operations
	DefaultAPITimeout = 30 * time.Second

	// Extended timeout for long-running operations
	ExtendedAPITimeout = 5 * time.Minute

	// Timeout for execution operations
	ExecutionTimeout = 30 * time.Minute

	// Default polling interval for status updates
	DefaultPollingInterval = 5 * time.Second

	// Default follow timeout for monitoring operations
	DefaultFollowTimeout = 300 * time.Second
)

// Exit codes for consistent CLI behavior
const (
	ExitSuccess         = 0
	ExitGeneralError    = 1
	ExitUsageError      = 2
	ExitAPIError        = 3
	ExitTimeout         = 4
	ExitCanceled        = 5
	ExitNotFound        = 6
	ExitUnauthorized    = 7
	ExitRateLimited     = 8
	ExitValidationError = 9
)

// ========================================
// Configuration Management
// ========================================

// Global config variable for the commands package
var cfg *config.VirtuosoConfig

// SetConfig sets the global config for the commands package
func SetConfig(c *config.VirtuosoConfig) {
	cfg = c
}

// validateConfig checks if the configuration is valid
func validateConfig() error {
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	if cfg.API.AuthToken == "" {
		return fmt.Errorf("API auth token not configured")
	}

	if cfg.Org.ID == "" {
		return fmt.Errorf("organization ID not configured")
	}

	return nil
}

// ========================================
// Context Helpers
// ========================================

// CommandContext creates a context for command execution with signal handling
func CommandContext() (context.Context, context.CancelFunc) {
	return CommandContextWithTimeout(DefaultAPITimeout)
}

// CommandContextWithTimeout creates a context with a specific timeout and signal handling
func CommandContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			// User interrupted, cancel context
			cancel()
		case <-ctx.Done():
			// Context finished normally
		}
		signal.Stop(sigChan)
	}()

	return ctx, cancel
}

// ExtendedCommandContext creates a context for long-running operations
func ExtendedCommandContext() (context.Context, context.CancelFunc) {
	return CommandContextWithTimeout(ExtendedAPITimeout)
}

// ExecutionCommandContext creates a context for execution operations
func ExecutionCommandContext() (context.Context, context.CancelFunc) {
	return CommandContextWithTimeout(ExecutionTimeout)
}

// InfiniteContext creates a context that only responds to cancellation, not timeout
func InfiniteContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			// User interrupted, cancel context
			cancel()
		case <-ctx.Done():
			// Context finished normally
		}
		signal.Stop(sigChan)
	}()

	return ctx, cancel
}

// callWithContext executes a function with context support
func callWithContext[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	type result struct {
		value T
		err   error
	}

	done := make(chan result, 1)

	go func() {
		val, err := fn()
		done <- result{value: val, err: err}
	}()

	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case res := <-done:
		return res.value, res.err
	}
}

// ========================================
// Environment and Session Helpers
// ========================================

// getSessionCheckpointID retrieves checkpoint ID from environment
func getSessionCheckpointID() string {
	return os.Getenv("VIRTUOSO_SESSION_ID")
}

// resolveCheckpointID resolves checkpoint ID from various sources
func resolveCheckpointID(args []string, flagValue string) (string, []string, error) {
	// Priority: flag > positional > session
	if flagValue != "" {
		return flagValue, args, nil
	}

	if len(args) > 0 && isCheckpointID(args[0]) {
		return args[0], args[1:], nil
	}

	sessionID := getSessionCheckpointID()
	if sessionID != "" {
		return sessionID, args, nil
	}

	return "", args, fmt.Errorf("checkpoint ID required: provide via argument, --checkpoint flag, or VIRTUOSO_SESSION_ID environment variable")
}

// isCheckpointID checks if a string looks like a checkpoint ID
func isCheckpointID(s string) bool {
	// Check if it's a number or starts with cp_
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return strings.HasPrefix(s, "cp_")
}

// parseCheckpointID extracts numeric ID from checkpoint string
func parseCheckpointID(id string) (int, error) {
	// Remove cp_ prefix if present
	id = strings.TrimPrefix(id, "cp_")
	return strconv.Atoi(id)
}

// ========================================
// Command Initialization
// ========================================

// initCommand performs common command initialization
func initCommand(cmd *cobra.Command) error {
	if err := validateConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Set output format from flag or config
	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "" {
		outputFormat = cfg.Output.DefaultFormat
	}

	return nil
}

// ========================================
// Utility Functions
// ========================================

// parseID converts string ID to int with validation
func parseID(idStr string, resourceType string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s ID: %s", resourceType, idStr)
	}
	if id <= 0 {
		return 0, fmt.Errorf("%s ID must be positive: %d", resourceType, id)
	}
	return id, nil
}

// extractPosition extracts position from args if present
func extractPosition(args []string) (int, []string) {
	if len(args) > 0 {
		if pos, err := strconv.Atoi(args[len(args)-1]); err == nil && pos > 0 {
			return pos, args[:len(args)-1]
		}
	}
	return -1, args
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", d.Seconds()*1000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// GetTimeoutFromEnv returns a timeout duration from environment variable or default
func GetTimeoutFromEnv(envVar string, defaultTimeout time.Duration) time.Duration {
	if timeoutStr := os.Getenv(envVar); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			return timeout
		}
	}
	return defaultTimeout
}

// SetExitCode sets the appropriate exit code based on error type
func SetExitCode(cmd *cobra.Command, err error) {
	if err == nil {
		os.Exit(ExitSuccess)
		return
	}

	// Check for API errors
	if apiErr, ok := err.(*client.APIError); ok {
		switch apiErr.Status {
		case 404:
			os.Exit(ExitNotFound)
		case 401:
			os.Exit(ExitUnauthorized)
		case 429:
			os.Exit(ExitRateLimited)
		case 400:
			os.Exit(ExitValidationError)
		default:
			os.Exit(ExitAPIError)
		}
		return
	}

	// Check for client errors
	if clientErr, ok := err.(*client.ClientError); ok {
		switch clientErr.Kind {
		case client.KindTimeout:
			os.Exit(ExitTimeout)
		case client.KindContextCanceled:
			os.Exit(ExitCanceled)
		case client.KindInvalidInput:
			os.Exit(ExitValidationError)
		default:
			os.Exit(ExitGeneralError)
		}
		return
	}

	// Check for context errors
	if errors.Is(err, context.Canceled) {
		os.Exit(ExitCanceled)
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		os.Exit(ExitTimeout)
		return
	}

	// Default to general error
	os.Exit(ExitGeneralError)
}

// ========================================
// Validation Command
// ========================================

// newValidateConfigCmd creates a command to validate API configuration
func newValidateConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-config",
		Short: "Validate API configuration and connectivity",
		Long: `Validate that the configuration file is properly set up and that the API is accessible.

This command checks:
- Configuration file exists and is valid
- API endpoint is reachable
- Authentication credentials are valid
- Shows current organization details`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Step 1: Check configuration file
			switch cfg.Output.DefaultFormat {
			case "json":
				fmt.Fprintf(os.Stdout, `{"step": "checking_config", "status": "in_progress"}`+"\n")
			default:
				fmt.Println("🔍 Checking configuration file...")
			}

			if cfg == nil {
				return fmt.Errorf("configuration not loaded")
			}

			// Check if config file exists
			configPath := "./config/virtuoso-config.yaml"
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				switch cfg.Output.DefaultFormat {
				case "json":
					fmt.Fprintf(os.Stdout, `{"step": "config_file_check", "status": "warning", "message": "Config file not found, using defaults"}`+"\n")
				default:
					fmt.Printf("⚠️  Config file not found at %s, using defaults\n", configPath)
				}
			}

			// Validate required fields
			if cfg.API.BaseURL == "" {
				return fmt.Errorf("base URL not configured")
			}

			if cfg.API.AuthToken == "" {
				return fmt.Errorf("auth token not configured")
			}

			if cfg.Org.ID == "" {
				return fmt.Errorf("organization ID not configured")
			}

			// Step 2: Test API connectivity
			switch cfg.Output.DefaultFormat {
			case "json":
				fmt.Fprintf(os.Stdout, `{"step": "testing_api", "status": "in_progress"}`+"\n")
			default:
				fmt.Println("🌐 Testing API connectivity...")
			}

			apiClient := client.NewClient(cfg)

			// Step 3: Validate authentication by listing projects
			switch cfg.Output.DefaultFormat {
			case "json":
				fmt.Fprintf(os.Stdout, `{"step": "validating_auth", "status": "in_progress"}`+"\n")
			default:
				fmt.Println("🔐 Validating authentication...")
			}

			// We'll implement a simple health check by trying to list projects
			resp, err := apiClient.TestConnection()
			if err != nil {
				return fmt.Errorf("API connection failed: %w", err)
			}

			// Output results
			switch cfg.Output.DefaultFormat {
			case "json":
				result := map[string]interface{}{
					"status": "valid",
					"config": map[string]interface{}{
						"base_url":        cfg.API.BaseURL,
						"organization_id": cfg.Org.ID,
						"headers":         cfg.Headers,
					},
					"api_test": map[string]interface{}{
						"reachable":     true,
						"authenticated": true,
						"response_time": resp,
					},
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.Encode(result)

			case "yaml":
				fmt.Println("status: valid")
				fmt.Printf("base_url: %s\n", cfg.API.BaseURL)
				fmt.Printf("organization_id: %s\n", cfg.Org.ID)
				fmt.Println("api_reachable: true")
				fmt.Println("authenticated: true")

			case "ai":
				fmt.Println("Configuration validation successful!")
				fmt.Printf("\nConfiguration Details:\n")
				fmt.Printf("- Base URL: %s\n", cfg.API.BaseURL)
				fmt.Printf("- Organization ID: %s\n", cfg.Org.ID)
				fmt.Printf("- Auth Token: %s...%s (hidden)\n", cfg.API.AuthToken[:8], cfg.API.AuthToken[len(cfg.API.AuthToken)-4:])
				fmt.Printf("\nAPI Status:\n")
				fmt.Printf("- API is reachable\n")
				fmt.Printf("- Authentication is valid\n")
				fmt.Printf("- Response time: %s\n", resp)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. List projects: api-cli list-projects\n")
				fmt.Printf("2. Create a project: api-cli create-project \"My Project\"\n")

			default: // human
				fmt.Println("✅ Configuration is valid!")
				fmt.Printf("\n📋 Configuration:\n")
				fmt.Printf("   Base URL: %s\n", cfg.API.BaseURL)
				fmt.Printf("   Organization ID: %s\n", cfg.Org.ID)
				fmt.Printf("   Auth Token: ***...%s\n", cfg.API.AuthToken[len(cfg.API.AuthToken)-4:])
				fmt.Println("\n✅ API connection successful")
				fmt.Println("✅ Authentication valid")
			}

			return nil
		},
	}

	return cmd
}
