package client

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

// ResponseHandler provides robust handling of API responses with various formats
type ResponseHandler struct {
	DebugMode bool
}

// NewResponseHandler creates a new response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{
		DebugMode: false,
	}
}

// UniversalResponse can handle various response formats from the API
type UniversalResponse struct {
	// Direct ID at root level
	ID interface{} `json:"id,omitempty"`

	// Item wrapper (used in many responses)
	Item struct {
		ID interface{} `json:"id,omitempty"`
	} `json:"item,omitempty"`

	// TestStep wrapper (used in step creation)
	TestStep struct {
		ID interface{} `json:"id,omitempty"`
	} `json:"testStep,omitempty"`

	// Data wrapper (used in some responses)
	Data struct {
		ID interface{} `json:"id,omitempty"`
	} `json:"data,omitempty"`

	// Execution wrapper (for execute-goal)
	Execution struct {
		ID interface{} `json:"id,omitempty"`
	} `json:"execution,omitempty"`

	// Generic result wrapper
	Result struct {
		ID interface{} `json:"id,omitempty"`
	} `json:"result,omitempty"`

	// Success indicator (some APIs return this without ID)
	Success bool   `json:"success,omitempty"`
	Status  string `json:"status,omitempty"`

	// Error information
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`

	// Raw response for debugging
	Raw json.RawMessage `json:"-"`
}

// ExtractID attempts to extract an ID from various response formats
func (h *ResponseHandler) ExtractID(response []byte) (interface{}, error) {
	var universal UniversalResponse
	universal.Raw = response

	if err := json.Unmarshal(response, &universal); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for explicit error
	if universal.Error != "" || universal.Message != "" {
		return nil, fmt.Errorf("API error: %s %s", universal.Error, universal.Message)
	}

	// Try various locations for ID
	idCandidates := []interface{}{
		universal.ID,
		universal.Item.ID,
		universal.TestStep.ID,
		universal.Data.ID,
		universal.Execution.ID,
		universal.Result.ID,
	}

	for _, candidate := range idCandidates {
		if id := h.normalizeID(candidate); id != nil {
			// Check for placeholder IDs
			if h.isPlaceholderID(id) {
				return id, &PlaceholderIDError{ID: id}
			}
			return id, nil
		}
	}

	// If no ID but success indicated, it might be a valid response
	if universal.Success || universal.Status == "success" || universal.Status == "ok" {
		return nil, &NoIDButSuccessError{}
	}

	// Try to find ID in raw JSON using common patterns
	if id := h.findIDInRawJSON(response); id != nil {
		return id, nil
	}

	if h.DebugMode {
		fmt.Printf("DEBUG: Could not find ID in response: %s\n", string(response))
	}

	return nil, &InvalidIDError{Response: string(response)}
}

// ExtractIntID extracts an ID and converts it to int
func (h *ResponseHandler) ExtractIntID(response []byte) (int, error) {
	id, err := h.ExtractID(response)
	if err != nil {
		return 0, err
	}

	return h.toInt(id)
}

// ExtractStringID extracts an ID and converts it to string
func (h *ResponseHandler) ExtractStringID(response []byte) (string, error) {
	id, err := h.ExtractID(response)
	if err != nil {
		return "", err
	}

	return h.toString(id)
}

// normalizeID converts various ID formats to a consistent type
func (h *ResponseHandler) normalizeID(id interface{}) interface{} {
	if id == nil {
		return nil
	}

	// Handle different types
	switch v := id.(type) {
	case float64:
		if v == 0 {
			return nil
		}
		return int(v)
	case int64:
		if v == 0 {
			return nil
		}
		return int(v)
	case int:
		if v == 0 {
			return nil
		}
		return v
	case string:
		if v == "" || v == "0" {
			return nil
		}
		// Try to parse as int
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
		return v
	case json.Number:
		if intVal, err := v.Int64(); err == nil {
			if intVal == 0 {
				return nil
			}
			return int(intVal)
		}
		return v.String()
	default:
		// Use reflection for unknown types
		rv := reflect.ValueOf(id)
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			return nil
		}
		return id
	}
}

// toInt converts an interface{} to int
func (h *ResponseHandler) toInt(id interface{}) (int, error) {
	switch v := id.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case json.Number:
		i, err := v.Int64()
		return int(i), err
	default:
		return 0, fmt.Errorf("cannot convert %T to int", id)
	}
}

// toString converts an interface{} to string
func (h *ResponseHandler) toString(id interface{}) (string, error) {
	switch v := id.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case json.Number:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", id), nil
	}
}

// isPlaceholderID checks if the ID is a placeholder (like 1 for dialog/mouse commands)
func (h *ResponseHandler) isPlaceholderID(id interface{}) bool {
	intID, err := h.toInt(id)
	if err != nil {
		return false
	}
	return intID == 1 || intID == 0
}

// findIDInRawJSON attempts to find an ID using regex patterns
func (h *ResponseHandler) findIDInRawJSON(response []byte) interface{} {
	responseStr := string(response)

	// Common patterns for ID fields
	patterns := []string{
		`"id"\s*:\s*(\d+)`,
		`"id"\s*:\s*"(\d+)"`,
		`"stepId"\s*:\s*(\d+)`,
		`"executionId"\s*:\s*(\d+)`,
		`"itemId"\s*:\s*(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(responseStr); len(matches) > 1 {
			if id, err := strconv.Atoi(matches[1]); err == nil && id > 0 {
				return id
			}
		}
	}

	return nil
}

// Custom error types
type PlaceholderIDError struct {
	ID interface{}
}

func (e *PlaceholderIDError) Error() string {
	return fmt.Sprintf("placeholder ID returned: %v (operation likely succeeded)", e.ID)
}

type NoIDButSuccessError struct{}

func (e *NoIDButSuccessError) Error() string {
	return "no ID in response but operation indicated success"
}

type InvalidIDError struct {
	Response string
}

func (e *InvalidIDError) Error() string {
	return fmt.Sprintf("no valid ID found in response: %s", e.Response)
}

// ParseExecutionResponse handles the specific execute-goal response format
func (h *ResponseHandler) ParseExecutionResponse(response []byte) (*Execution, error) {
	// First try the standard response format
	var resp struct {
		Item      *Execution `json:"item"`
		Execution *Execution `json:"execution"`
		Data      *Execution `json:"data"`
		Success   bool       `json:"success"`
		Error     string     `json:"error"`
	}

	if err := json.Unmarshal(response, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse execution response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("API error: %s", resp.Error)
	}

	// Try different response locations
	candidates := []*Execution{resp.Item, resp.Execution, resp.Data}
	for _, exec := range candidates {
		if exec != nil && (exec.ID != "" || exec.ExecutionID != "") {
			// Normalize the execution ID
			if exec.ID == "" && exec.ExecutionID != "" {
				exec.ID = exec.ExecutionID
			}
			// Handle both string and numeric IDs
			exec.ID = h.normalizeExecutionID(exec.ID)
			return exec, nil
		}
	}

	// Try parsing as direct execution object
	var directExec Execution
	if err := json.Unmarshal(response, &directExec); err == nil {
		if directExec.ID != "" || directExec.ExecutionID != "" {
			if directExec.ID == "" && directExec.ExecutionID != "" {
				directExec.ID = directExec.ExecutionID
			}
			directExec.ID = h.normalizeExecutionID(directExec.ID)
			return &directExec, nil
		}
	}

	return nil, fmt.Errorf("no execution data found in response")
}

// normalizeExecutionID handles both string and numeric execution IDs
func (h *ResponseHandler) normalizeExecutionID(id interface{}) string {
	if id == nil {
		return ""
	}

	switch v := id.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", id)
	}
}

// ParseTimeFormat handles various time formats from the API
func (h *ResponseHandler) ParseTimeFormat(timeStr interface{}) (time.Time, error) {
	if timeStr == nil {
		return time.Time{}, fmt.Errorf("nil time value")
	}

	switch v := timeStr.(type) {
	case string:
		// Try various time formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse time string: %s", v)

	case float64:
		// Assume Unix timestamp
		return time.Unix(int64(v), 0), nil

	case int64:
		// Unix timestamp
		return time.Unix(v, 0), nil

	default:
		return time.Time{}, fmt.Errorf("unsupported time type: %T", timeStr)
	}
}
