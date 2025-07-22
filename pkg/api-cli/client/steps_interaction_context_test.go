package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
)

// TestInteractionStepsWithContext tests all interaction step methods with context
func TestInteractionStepsWithContext(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"item":{"id":123}}`))
	}))
	defer server.Close()

	// Create client
	cfg := &config.VirtuosoConfig{
		API: config.APIConfig{
			BaseURL:   server.URL,
			AuthToken: "test-token",
		},
		HTTP: config.HTTPConfig{
			Timeout:   30,
			Retries:   0,
			RetryWait: 1,
		},
		Headers: config.HeadersConfig{
			ClientID:   "test-client",
			ClientName: "Test Client",
		},
	}
	client := NewClient(cfg)
	ctx := context.Background()

	// Test all methods
	tests := []struct {
		name string
		fn   func() (int, error)
	}{
		{
			name: "CreateStepClickWithContext",
			fn: func() (int, error) {
				return client.CreateStepClickWithContext(ctx, 1, "button", 1)
			},
		},
		{
			name: "CreateStepWriteWithContext",
			fn: func() (int, error) {
				return client.CreateStepWriteWithContext(ctx, 1, "input", "text", 1)
			},
		},
		{
			name: "CreateStepClickWithVariableWithContext",
			fn: func() (int, error) {
				return client.CreateStepClickWithVariableWithContext(ctx, 1, "var1", 1)
			},
		},
		{
			name: "CreateStepClickWithDetailsWithContext",
			fn: func() (int, error) {
				return client.CreateStepClickWithDetailsWithContext(ctx, 1, "button", "first", "button", 1)
			},
		},
		{
			name: "CreateStepWriteWithVariableWithContext",
			fn: func() (int, error) {
				return client.CreateStepWriteWithVariableWithContext(ctx, 1, "input", "text", "var1", 1)
			},
		},
		{
			name: "CreateStepDoubleClickWithContext",
			fn: func() (int, error) {
				return client.CreateStepDoubleClickWithContext(ctx, 1, "button", 1)
			},
		},
		{
			name: "CreateStepRightClickWithContext",
			fn: func() (int, error) {
				return client.CreateStepRightClickWithContext(ctx, 1, "button", 1)
			},
		},
		{
			name: "CreateStepHoverWithContext",
			fn: func() (int, error) {
				return client.CreateStepHoverWithContext(ctx, 1, "button", 1)
			},
		},
		{
			name: "CreateStepKeyGlobalWithContext",
			fn: func() (int, error) {
				return client.CreateStepKeyGlobalWithContext(ctx, 1, "Enter", 1)
			},
		},
		{
			name: "CreateStepKeyTargetedWithContext",
			fn: func() (int, error) {
				return client.CreateStepKeyTargetedWithContext(ctx, 1, "input", "Enter", 1)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, err := test.fn()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if id != 123 {
				t.Errorf("expected id 123, got %d", id)
			}
		})
	}
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	// Create a test server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block forever
		<-time.After(1 * time.Hour)
	}))
	defer server.Close()

	// Create client
	cfg := &config.VirtuosoConfig{
		API: config.APIConfig{
			BaseURL:   server.URL,
			AuthToken: "test-token",
		},
		HTTP: config.HTTPConfig{
			Timeout:   30,
			Retries:   0,
			RetryWait: 1,
		},
		Headers: config.HeadersConfig{
			ClientID:   "test-client",
			ClientName: "Test Client",
		},
	}
	client := NewClient(cfg)

	// Test with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.CreateStepClickWithContext(ctx, 1, "button", 1)
	if err == nil {
		t.Error("expected error for canceled context")
	}

	clientErr, ok := err.(*ClientError)
	if !ok {
		t.Errorf("expected ClientError, got %T", err)
	}
	if clientErr.Kind != KindContextCanceled {
		t.Errorf("expected KindContextCanceled, got %s", clientErr.Kind)
	}
}

// TestBackwardCompatibility ensures non-context methods still work
func TestBackwardCompatibility(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"item":{"id":789}}`))
	}))
	defer server.Close()

	// Create client
	cfg := &config.VirtuosoConfig{
		API: config.APIConfig{
			BaseURL:   server.URL,
			AuthToken: "test-token",
		},
		HTTP: config.HTTPConfig{
			Timeout:   30,
			Retries:   0,
			RetryWait: 1,
		},
		Headers: config.HeadersConfig{
			ClientID:   "test-client",
			ClientName: "Test Client",
		},
	}
	client := NewClient(cfg)

	// Test that original methods still work
	id, err := client.CreateStepClick(1, "button", 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 789 {
		t.Errorf("expected id 789, got %d", id)
	}
}
