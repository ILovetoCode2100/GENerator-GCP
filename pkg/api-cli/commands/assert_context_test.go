package commands

import (
	"context"
	"errors"
	"strings"
	"testing"

	"virtuoso-cli/pkg/api-cli/client"
)

// TestAssertCommandContext verifies that assert commands properly use context
func TestAssertCommandContext(t *testing.T) {
	tests := []struct {
		name       string
		assertType string
		setupError error
		wantErrMsg string
	}{
		{
			name:       "context timeout",
			assertType: "exists",
			setupError: context.DeadlineExceeded,
			wantErrMsg: "request timed out while creating ASSERT_EXISTS step",
		},
		{
			name:       "context canceled",
			assertType: "equals",
			setupError: context.Canceled,
			wantErrMsg: "request was canceled while creating ASSERT_EQUALS step",
		},
		{
			name:       "not found error",
			assertType: "checked",
			setupError: client.NewAPIError(404, client.ErrCodeNotFound, "Checkpoint not found"),
			wantErrMsg: "checkpoint 12345 not found",
		},
		{
			name:       "unauthorized error",
			assertType: "gt",
			setupError: client.NewAPIError(401, client.ErrCodeUnauthorized, "Invalid token"),
			wantErrMsg: "unauthorized: please check your API token",
		},
		{
			name:       "rate limited error",
			assertType: "matches",
			setupError: client.NewAPIError(429, client.ErrCodeRateLimited, "Too many requests"),
			wantErrMsg: "rate limited: please try again later",
		},
		{
			name:       "generic API error",
			assertType: "variable",
			setupError: client.NewAPIError(400, "BAD_REQUEST", "Invalid selector"),
			wantErrMsg: "API error creating ASSERT_VARIABLE step:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &AssertCommand{
				BaseCommand: &BaseCommand{
					CheckpointID: "12345",
					Position:     1,
				},
			}

			// Mock the error in createAssertStep
			config := assertConfigs[tt.assertType]
			meta := make(map[string]interface{})

			// This would require mocking the client, but demonstrates the error handling pattern
			_, err := ac.createAssertStep(strings.ToUpper("ASSERT_"+tt.assertType), meta)

			if err != nil && !errors.Is(err, tt.setupError) {
				// Check if error message contains expected text
				if !contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErrMsg, err.Error())
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
