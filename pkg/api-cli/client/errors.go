package client

import (
	"fmt"
	"net/http"
)

// APIError represents a structured error from the Virtuoso API
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Status     int    `json:"status"`
	Details    string `json:"details,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	RetryAfter int    `json:"retry_after,omitempty"` // For rate limiting
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error %s (status %d): %s", e.Code, e.Status, e.Message)
	}
	return fmt.Sprintf("API error (status %d): %s", e.Status, e.Message)
}

// IsRetryable returns true if the error is temporary and the request can be retried
func (e *APIError) IsRetryable() bool {
	switch e.Status {
	case http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusBadGateway:
		return true
	case http.StatusInternalServerError:
		// Some 500 errors might be retryable
		return e.Code == "TEMPORARY_ERROR" || e.Code == "TIMEOUT"
	default:
		return false
	}
}

// Common error codes
const (
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeForbidden         = "FORBIDDEN"
	ErrCodeBadRequest        = "BAD_REQUEST"
	ErrCodeConflict          = "CONFLICT"
	ErrCodeRateLimited       = "RATE_LIMITED"
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeTimeout           = "TIMEOUT"
	ErrCodeValidation        = "VALIDATION_ERROR"
	ErrCodeDependencyMissing = "DEPENDENCY_MISSING"
)

// ClientError represents errors that occur on the client side
type ClientError struct {
	Op      string // Operation being performed
	Kind    string // Kind of error
	Message string
	Err     error // Underlying error
}

// Error implements the error interface
func (e *ClientError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %s (%v)", e.Op, e.Kind, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s: %s", e.Op, e.Kind, e.Message)
}

// Unwrap returns the underlying error
func (e *ClientError) Unwrap() error {
	return e.Err
}

// Client error kinds
const (
	KindInvalidInput     = "invalid_input"
	KindConnectionFailed = "connection_failed"
	KindTimeout          = "timeout"
	KindContextCanceled  = "context_canceled"
	KindSerialization    = "serialization_error"
)

// Helper functions for creating common errors

// NewAPIError creates a new APIError from an HTTP response
func NewAPIError(status int, code, message string) *APIError {
	return &APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// NewClientError creates a new ClientError
func NewClientError(op, kind, message string, err error) *ClientError {
	return &ClientError{
		Op:      op,
		Kind:    kind,
		Message: message,
		Err:     err,
	}
}

// IsNotFound returns true if the error represents a not found error
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Status == http.StatusNotFound || apiErr.Code == ErrCodeNotFound
	}
	return false
}

// IsUnauthorized returns true if the error represents an unauthorized error
func IsUnauthorized(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Status == http.StatusUnauthorized || apiErr.Code == ErrCodeUnauthorized
	}
	return false
}

// IsRateLimited returns true if the error represents a rate limit error
func IsRateLimited(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Status == http.StatusTooManyRequests || apiErr.Code == ErrCodeRateLimited
	}
	return false
}

// IsTimeout returns true if the error represents a timeout
func IsTimeout(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code == ErrCodeTimeout
	}
	if clientErr, ok := err.(*ClientError); ok {
		return clientErr.Kind == KindTimeout
	}
	return false
}

// IsContextCanceled returns true if the error represents a canceled context
func IsContextCanceled(err error) bool {
	if clientErr, ok := err.(*ClientError); ok {
		return clientErr.Kind == KindContextCanceled
	}
	return false
}
