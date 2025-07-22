package commands

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/client"
	"github.com/spf13/cobra"
)

const (
	// Default timeout for API operations
	DefaultAPITimeout = 30 * time.Second

	// Extended timeout for long-running operations
	ExtendedAPITimeout = 5 * time.Minute

	// Timeout for execution operations
	ExecutionTimeout = 30 * time.Minute
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
