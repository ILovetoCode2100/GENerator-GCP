package client

import (
	"testing"
)

func TestResponseHandler_ParseStepID(t *testing.T) {
	handler := NewResponseHandler(false)

	tests := []struct {
		name     string
		response string
		wantID   int
		wantErr  bool
		errType  error
	}{
		{
			name:     "direct numeric ID",
			response: `{"id": 12345}`,
			wantID:   12345,
		},
		{
			name:     "direct string ID",
			response: `{"id": "12345"}`,
			wantID:   12345,
		},
		{
			name:     "wrapped in item with numeric ID",
			response: `{"item": {"id": 67890}}`,
			wantID:   67890,
		},
		{
			name:     "wrapped in item with string ID",
			response: `{"item": {"id": "67890"}}`,
			wantID:   67890,
		},
		{
			name:     "wrapped in testStep",
			response: `{"testStep": {"id": 11111}}`,
			wantID:   11111,
		},
		{
			name:     "wrapped in data",
			response: `{"data": {"id": 22222}}`,
			wantID:   22222,
		},
		{
			name:     "stepId field",
			response: `{"stepId": 33333}`,
			wantID:   33333,
		},
		{
			name:     "testStepId field",
			response: `{"testStepId": "44444"}`,
			wantID:   44444,
		},
		{
			name:     "placeholder ID warning",
			response: `{"id": 1}`,
			wantID:   1,
			wantErr:  false, // Should return ID but validate will warn
		},
		{
			name:     "zero ID",
			response: `{"id": 0}`,
			wantID:   0,
			wantErr:  true,
		},
		{
			name:     "success without ID",
			response: `{"success": true, "message": "Step created"}`,
			wantID:   0,
			wantErr:  true,
			errType:  &NoIDButSuccessError{},
		},
		{
			name:     "no ID found",
			response: `{"error": "some error"}`,
			wantID:   0,
			wantErr:  true,
		},
		{
			name:     "nested path item.stepId",
			response: `{"item": {"stepId": 55555}}`,
			wantID:   55555,
		},
		{
			name:     "float ID conversion",
			response: `{"id": 12345.0}`,
			wantID:   12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := handler.ParseStepID([]byte(tt.response))

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStepID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if id != tt.wantID {
				t.Errorf("ParseStepID() = %v, want %v", id, tt.wantID)
			}

			if tt.errType != nil && err != nil {
				switch tt.errType.(type) {
				case *NoIDButSuccessError:
					if !IsNoIDButSuccessError(err) {
						t.Errorf("Expected NoIDButSuccessError, got %T", err)
					}
				}
			}
		})
	}
}

func TestResponseHandler_ParseExecutionID(t *testing.T) {
	handler := NewResponseHandler(false)

	tests := []struct {
		name     string
		response string
		wantID   string
		wantErr  bool
	}{
		{
			name:     "string execution ID",
			response: `{"id": "exec_12345"}`,
			wantID:   "exec_12345",
		},
		{
			name:     "numeric execution ID",
			response: `{"id": 12345}`,
			wantID:   "12345",
		},
		{
			name:     "float execution ID",
			response: `{"id": 12345.0}`,
			wantID:   "12345",
		},
		{
			name:     "wrapped in item",
			response: `{"item": {"id": "exec_67890"}}`,
			wantID:   "exec_67890",
		},
		{
			name:     "executionId field",
			response: `{"executionId": "exec_11111"}`,
			wantID:   "exec_11111",
		},
		{
			name:     "numeric in item",
			response: `{"item": {"id": 22222}}`,
			wantID:   "22222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := handler.ParseExecutionID([]byte(tt.response))

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExecutionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if id != tt.wantID {
				t.Errorf("ParseExecutionID() = %v, want %v", id, tt.wantID)
			}
		})
	}
}

func TestResponseHandler_ParseExecutionResponse(t *testing.T) {
	handler := NewResponseHandler(false)

	tests := []struct {
		name     string
		response string
		validate func(*testing.T, *Execution)
		wantErr  bool
	}{
		{
			name: "execution with string ID",
			response: `{
				"id": "exec_123",
				"goalId": 456,
				"snapshotId": 789,
				"status": "running",
				"startTime": "2024-01-15T10:00:00Z",
				"duration": 1500,
				"resultsUrl": "https://example.com/results",
				"reportUrl": "https://example.com/report"
			}`,
			validate: func(t *testing.T, exec *Execution) {
				if exec.ID != "exec_123" {
					t.Errorf("ID = %v, want exec_123", exec.ID)
				}
				if exec.GoalID != 456 {
					t.Errorf("GoalID = %v, want 456", exec.GoalID)
				}
				if exec.Status != "running" {
					t.Errorf("Status = %v, want running", exec.Status)
				}
			},
		},
		{
			name: "execution with numeric ID wrapped in item",
			response: `{
				"item": {
					"id": 12345,
					"goalId": "456",
					"snapshotId": 789.0,
					"status": "completed",
					"startTime": "2024-01-15T10:00:00Z",
					"endTime": "2024-01-15T10:05:00Z",
					"duration": 300
				}
			}`,
			validate: func(t *testing.T, exec *Execution) {
				if exec.ID != "12345" {
					t.Errorf("ID = %v, want 12345", exec.ID)
				}
				if exec.GoalID != 456 {
					t.Errorf("GoalID = %v, want 456", exec.GoalID)
				}
				if exec.EndTime == nil || exec.EndTime.IsZero() {
					t.Error("EndTime should not be nil or zero")
				}
			},
		},
		{
			name: "execution with unix timestamp",
			response: `{
				"id": "exec_789",
				"goalId": 123,
				"snapshotId": 456,
				"status": "failed",
				"startTime": 1705316400,
				"duration": "600"
			}`,
			validate: func(t *testing.T, exec *Execution) {
				if exec.StartTime.IsZero() {
					t.Error("StartTime should not be zero")
				}
				if exec.Duration != 600 {
					t.Errorf("Duration = %v, want 600", exec.Duration)
				}
			},
		},
		{
			name: "execution with progress",
			response: `{
				"id": "exec_999",
				"goalId": 111,
				"snapshotId": 222,
				"status": "running",
				"startTime": "2024-01-15T10:00:00Z",
				"progress": {
					"current": 5,
					"total": 10,
					"percentage": 50
				}
			}`,
			validate: func(t *testing.T, exec *Execution) {
				if exec.Progress == nil {
					t.Error("Progress should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var exec Execution
			err := handler.ParseResponse([]byte(tt.response), &exec)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, &exec)
			}
		})
	}
}

func TestResponseHandler_ValidateID(t *testing.T) {
	handler := NewResponseHandler(false)

	tests := []struct {
		name    string
		id      int
		wantErr bool
		errType error
	}{
		{
			name:    "valid ID",
			id:      12345,
			wantErr: false,
		},
		{
			name:    "zero ID",
			id:      0,
			wantErr: true,
			errType: &InvalidIDError{},
		},
		{
			name:    "placeholder ID",
			id:      1,
			wantErr: true,
			errType: &PlaceholderIDError{},
		},
		{
			name:    "large valid ID",
			id:      999999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateID(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errType != nil && err != nil {
				switch tt.errType.(type) {
				case *InvalidIDError:
					if _, ok := err.(*InvalidIDError); !ok {
						t.Errorf("Expected InvalidIDError, got %T", err)
					}
				case *PlaceholderIDError:
					if !IsPlaceholderError(err) {
						t.Errorf("Expected PlaceholderIDError, got %T", err)
					}
				}
			}
		})
	}
}

func TestResponseHandler_ComplexResponses(t *testing.T) {
	handler := NewResponseHandler(false)

	// Test dialog command response format
	dialogResponse := `{
		"success": true,
		"item": {
			"id": 1,
			"type": "dismiss-alert",
			"message": "Alert dismissed"
		}
	}`

	id, err := handler.ParseStepID([]byte(dialogResponse))
	if err != nil {
		t.Errorf("Failed to parse dialog response: %v", err)
	}
	if id != 1 {
		t.Errorf("Dialog response ID = %v, want 1", id)
	}

	// Test mouse command response format
	mouseResponse := `{
		"testStep": {
			"id": "1",
			"action": "move-to",
			"coordinates": "100,200"
		}
	}`

	id, err = handler.ParseStepID([]byte(mouseResponse))
	if err != nil {
		t.Errorf("Failed to parse mouse response: %v", err)
	}
	if id != 1 {
		t.Errorf("Mouse response ID = %v, want 1", id)
	}

	// Test execution response with mixed types
	execResponse := `{
		"success": true,
		"item": {
			"id": 67890,
			"goalId": "123",
			"snapshotId": 456.0,
			"status": "queued",
			"startTime": null,
			"endTime": null,
			"duration": 0,
			"resultsUrl": "",
			"reportUrl": ""
		}
	}`

	var exec Execution
	err = handler.ParseResponse([]byte(execResponse), &exec)
	if err != nil {
		t.Errorf("Failed to parse execution response: %v", err)
	}
	if exec.ID != "67890" {
		t.Errorf("Execution ID = %v, want 67890", exec.ID)
	}
	if exec.GoalID != 123 {
		t.Errorf("Execution GoalID = %v, want 123", exec.GoalID)
	}
}

func TestResponseHandler_EdgeCases(t *testing.T) {
	handler := NewResponseHandler(false)

	// Empty response
	_, err := handler.ParseStepID([]byte(""))
	if err == nil {
		t.Error("Expected error for empty response")
	}

	// Invalid JSON
	_, err = handler.ParseStepID([]byte("not json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Null values
	nullResponse := `{"id": null, "success": true}`
	_, err = handler.ParseStepID([]byte(nullResponse))
	if !IsNoIDButSuccessError(err) {
		t.Errorf("Expected NoIDButSuccessError for null ID with success, got %v", err)
	}

	// Very large number
	largeNumResponse := `{"id": 9223372036854775807}`
	id, err := handler.ParseStepID([]byte(largeNumResponse))
	if err != nil {
		t.Errorf("Failed to parse large number: %v", err)
	}
	if id <= 0 {
		t.Error("Large number should be parsed successfully")
	}
}

func TestResponseHandler_TimeFormats(t *testing.T) {
	handler := NewResponseHandler(false)

	timeFormats := []struct {
		name    string
		input   interface{}
		notZero bool
	}{
		{"RFC3339", "2024-01-15T10:00:00Z", true},
		{"RFC3339Nano", "2024-01-15T10:00:00.123456789Z", true},
		{"Custom1", "2024-01-15T10:00:00.000Z", true},
		{"Custom2", "2024-01-15 10:00:00", true},
		{"Unix timestamp", float64(1705316400), true},
		{"Invalid", "not a time", false},
		{"Null", nil, false},
	}

	for _, tf := range timeFormats {
		t.Run(tf.name, func(t *testing.T) {
			result := handler.parseTime(tf.input)
			if tf.notZero && result.IsZero() {
				t.Errorf("Expected non-zero time for %v", tf.input)
			}
			if !tf.notZero && !result.IsZero() {
				t.Errorf("Expected zero time for %v", tf.input)
			}
		})
	}
}

// Benchmark tests
func BenchmarkResponseHandler_ParseStepID(b *testing.B) {
	handler := NewResponseHandler(false)
	response := []byte(`{"item": {"id": 12345}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.ParseStepID(response)
	}
}

func BenchmarkResponseHandler_ParseExecutionResponse(b *testing.B) {
	handler := NewResponseHandler(false)
	response := []byte(`{
		"item": {
			"id": 12345,
			"goalId": 456,
			"snapshotId": 789,
			"status": "running",
			"startTime": "2024-01-15T10:00:00Z"
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var exec Execution
		_ = handler.ParseResponse(response, &exec)
	}
}
