package client

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryWithBackoff retries operations with exponential backoff
func RetryWithBackoff(ctx context.Context, config *RetryConfig, operation RetryableFunc) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		// Check if error is retryable
		if !isRetryable(err) {
			return err
		}

		if attempt == config.MaxAttempts {
			return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, err)
		}

		// Add jitter if enabled
		actualDelay := delay
		if config.Jitter {
			jitter := time.Duration(rand.Float64() * float64(delay) * 0.3)
			actualDelay = delay + jitter
		}

		// Wait with backoff
		select {
		case <-time.After(actualDelay):
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return nil
}

// RetryWithCustomBackoff allows custom backoff strategies
func RetryWithCustomBackoff(ctx context.Context, operation RetryableFunc, backoffFunc BackoffFunc) error {
	attempt := 0
	maxAttempts := 10 // Safety limit

	for {
		attempt++
		err := operation()
		if err == nil {
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		if attempt >= maxAttempts {
			return fmt.Errorf("operation failed after %d attempts: %w", attempt, err)
		}

		delay := backoffFunc(attempt, err)
		if delay <= 0 {
			return err // Backoff function signals to stop retrying
		}

		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}
}

// BackoffFunc calculates delay for retry attempt
type BackoffFunc func(attempt int, err error) time.Duration

// ExponentialBackoff returns an exponential backoff function
func ExponentialBackoff(base time.Duration, multiplier float64, maxDelay time.Duration) BackoffFunc {
	return func(attempt int, err error) time.Duration {
		delay := base * time.Duration(math.Pow(multiplier, float64(attempt-1)))
		if delay > maxDelay {
			return maxDelay
		}
		return delay
	}
}

// LinearBackoff returns a linear backoff function
func LinearBackoff(increment time.Duration, maxDelay time.Duration) BackoffFunc {
	return func(attempt int, err error) time.Duration {
		delay := increment * time.Duration(attempt)
		if delay > maxDelay {
			return maxDelay
		}
		return delay
	}
}

// isRetryable determines if an error should trigger a retry
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types
	if apiErr, ok := err.(*APIError); ok {
		switch apiErr.Status {
		case 429: // Rate limit
			return true
		case 500, 502, 503, 504: // Server errors
			return true
		case 408: // Request timeout
			return true
		default:
			return false
		}
	}

	if clientErr, ok := err.(*ClientError); ok {
		switch clientErr.Kind {
		case KindTimeout, KindConnectionFailed:
			return true
		case KindContextCanceled:
			return false // Don't retry cancelled operations
		default:
			return false
		}
	}

	// Check for network-related errors
	errStr := err.Error()
	retryableStrings := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"timeout",
		"temporary failure",
		"too many requests",
		"service unavailable",
	}

	for _, s := range retryableStrings {
		if contains(errStr, s) {
			return true
		}
	}

	return false
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// RetryableClient wraps a client with retry logic
type RetryableClient struct {
	*Client
	retryConfig *RetryConfig
}

// NewRetryableClient creates a client with automatic retry
func NewRetryableClient(client *Client, config *RetryConfig) *RetryableClient {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &RetryableClient{
		Client:      client,
		retryConfig: config,
	}
}

// ExecuteGoalWithRetry executes a goal with automatic retry
func (rc *RetryableClient) ExecuteGoalWithRetry(ctx context.Context, goalID, snapshotID int) (*Execution, error) {
	var result *Execution
	var lastErr error

	err := RetryWithBackoff(ctx, rc.retryConfig, func() error {
		exec, err := rc.ExecuteGoalWithContext(ctx, goalID, snapshotID)
		if err != nil {
			lastErr = err
			return err
		}
		result = exec
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("execute goal failed: %w", lastErr)
	}

	return result, nil
}

// CreateStepWithRetry creates a step with automatic retry
func (rc *RetryableClient) CreateStepWithRetry(ctx context.Context, checkpointID int, stepData map[string]interface{}, position int) (int, error) {
	var stepID int
	var lastErr error

	err := RetryWithBackoff(ctx, rc.retryConfig, func() error {
		id, err := rc.createStepWithCustomBodyContext(ctx, checkpointID, stepData, position)
		if err != nil {
			lastErr = err
			return err
		}
		stepID = id
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("create step failed: %w", lastErr)
	}

	return stepID, nil
}

// CircuitBreaker prevents cascading failures
type CircuitBreaker struct {
	failureThreshold int
	resetTimeout     time.Duration
	failures         int
	lastFailure      time.Time
	state            CircuitState
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            CircuitClosed,
	}
}

// Call executes function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.failures = 0
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := fn()
	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.failureThreshold {
			cb.state = CircuitOpen
			return fmt.Errorf("circuit breaker opened after %d failures: %w", cb.failures, err)
		}

		return err
	}

	// Success - reset if in half-open state
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.failures = 0
	}

	return nil
}
