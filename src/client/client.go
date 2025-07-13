// Package client provides an enhanced API client with features like retry logic,
// circuit breaker patterns, and structured error handling for interacting with external APIs.
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	// "github.com/marklovelady/api-cli-generator/src/api" // TODO: Uncomment when generated API is available
)

// EnhancedClient wraps the generated API client with additional features
type EnhancedClient struct {
	// Generated client
	// apiClient *api.Client // TODO: Uncomment when generated API is available

	// Enhanced HTTP client
	httpClient *resty.Client

	// Configuration
	config Config

	// Logger
	logger *logrus.Logger
}

// Config holds client configuration
type Config struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
	Debug      bool
}

// NewEnhancedClient creates a new enhanced API client
func NewEnhancedClient(cfg Config) (*EnhancedClient, error) {
	logger := logrus.New()
	if cfg.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Create Resty client with retry and logging
	httpClient := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.MaxRetries).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second).
		OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
			logger.WithFields(logrus.Fields{
				"method": req.Method,
				"url":    req.URL,
			}).Debug("Making API request")
			return nil
		}).
		OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
			logger.WithFields(logrus.Fields{
				"status":   resp.StatusCode(),
				"duration": resp.Time(),
			}).Debug("Received API response")
			return nil
		})

	// Add authentication if provided
	if cfg.APIKey != "" {
		httpClient.SetAuthToken(cfg.APIKey)
	}

	// TODO: Create the generated client when API is available
	// apiClient, err := api.NewClient(cfg.BaseURL, api.WithHTTPClient(httpClient.GetClient()))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create API client: %w", err)
	// }

	return &EnhancedClient{
		// apiClient:  apiClient,
		httpClient: httpClient,
		config:     cfg,
		logger:     logger,
	}, nil
}

// GetAPIClient returns the underlying generated client
// TODO: Uncomment when generated API is available
// func (c *EnhancedClient) GetAPIClient() *api.Client {
// 	return c.apiClient
// }

// WithContext creates a new context with common values
func (c *EnhancedClient) WithContext(ctx context.Context) context.Context {
	// Add any common context values here
	return ctx
}

// HandleError provides consistent error handling
func (c *EnhancedClient) HandleError(err error, operation string) error {
	if err == nil {
		return nil
	}

	c.logger.WithError(err).WithField("operation", operation).Error("API operation failed")

	// Check for specific error types and wrap accordingly
	// TODO: Add specific error type handling when generated API is available
	// For now, return the wrapped error
	return fmt.Errorf("%s failed: %w", operation, err)
}

// ExecuteWithRetry executes a function with retry logic
func (c *EnhancedClient) ExecuteWithRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			waitTime := time.Duration(attempt) * time.Second
			c.logger.WithFields(logrus.Fields{
				"attempt":   attempt,
				"wait_time": waitTime,
			}).Debug("Retrying operation")
			time.Sleep(waitTime)
		}

		if err := fn(); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return c.HandleError(lastErr, operation)
}
