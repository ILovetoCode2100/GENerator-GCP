package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStepResponse(t *testing.T) {
	parser := NewResponseParser(false)

	testCases := []struct {
		name     string
		response string
		wantID   int
		wantErr  bool
	}{
		{
			name:     "direct ID format",
			response: `{"id": 12345, "status": "created"}`,
			wantID:   12345,
		},
		{
			name:     "item ID format",
			response: `{"success": true, "item": {"id": 67890}}`,
			wantID:   67890,
		},
		{
			name:     "testStep ID format",
			response: `{"success": true, "testStep": {"id": 11111}}`,
			wantID:   11111,
		},
		{
			name:     "data ID format",
			response: `{"data": {"id": 22222}}`,
			wantID:   22222,
		},
		{
			name:     "step ID format",
			response: `{"step": {"id": 33333}}`,
			wantID:   33333,
		},
		{
			name:     "response ID format",
			response: `{"response": {"id": 44444}}`,
			wantID:   44444,
		},
		{
			name:     "success without ID",
			response: `{"success": true, "message": "Step created"}`,
			wantID:   0,
			wantErr:  false, // Should not error, just return 0
		},
		{
			name:     "error response",
			response: `{"success": false, "error": "Invalid request"}`,
			wantErr:  true,
		},
		{
			name:     "empty response",
			response: `{}`,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := parser.ParseStepResponse([]byte(tc.response))

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantID, id)
			}
		})
	}
}

func TestParseExecutionResponse(t *testing.T) {
	testCases := []struct {
		name     string
		response string
		wantID   string
		wantErr  bool
	}{
		{
			name: "string ID format",
			response: `{
				"id": "exec_12345",
				"goalId": 1234,
				"snapshotId": 5678,
				"status": "RUNNING",
				"startTime": "2025-01-24T10:00:00Z"
			}`,
			wantID: "exec_12345",
		},
		{
			name: "numeric ID format",
			response: `{
				"id": 12345,
				"goalId": 1234,
				"snapshotId": 5678,
				"status": "RUNNING",
				"startTime": "2025-01-24T10:00:00Z"
			}`,
			wantID: "12345",
		},
		{
			name: "with progress",
			response: `{
				"id": 67890,
				"goalId": 1234,
				"snapshotId": 5678,
				"status": "RUNNING",
				"startTime": "2025-01-24T10:00:00Z",
				"progress": {
					"completed": 5,
					"total": 10,
					"percentage": 50
				}
			}`,
			wantID: "67890",
		},
		{
			name: "with end time",
			response: `{
				"id": "exec_99999",
				"goalId": 1234,
				"snapshotId": 5678,
				"status": "COMPLETED",
				"startTime": "2025-01-24T10:00:00Z",
				"endTime": "2025-01-24T10:05:00Z",
				"duration": 300000
			}`,
			wantID: "exec_99999",
		},
		{
			name:     "invalid JSON",
			response: `{invalid json}`,
			wantErr:  true,
		},
		{
			name:     "missing required fields",
			response: `{"id": 12345}`,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rawMsg json.RawMessage = []byte(tc.response)
			execution, err := ParseExecutionResponse(rawMsg)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, execution)
				assert.Equal(t, tc.wantID, execution.ID)
			}
		})
	}
}

func TestValidateStepResponse(t *testing.T) {
	testCases := []struct {
		name      string
		stepID    int
		wantError bool
		errorType string
	}{
		{
			name:      "valid step ID",
			stepID:    12345,
			wantError: false,
		},
		{
			name:      "placeholder ID 1",
			stepID:    1,
			wantError: true,
			errorType: "placeholder",
		},
		{
			name:      "zero ID",
			stepID:    0,
			wantError: true,
			errorType: "placeholder",
		},
		{
			name:      "negative ID",
			stepID:    -1,
			wantError: false, // Negative IDs are technically valid
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStepResponse(tc.stepID, nil)

			if tc.wantError {
				assert.Error(t, err)
				if tc.errorType == "placeholder" {
					assert.True(t, IsPlaceholderError(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTimeParsingFormats(t *testing.T) {
	testCases := []struct {
		name    string
		timeStr string
		wantErr bool
	}{
		{
			name:    "RFC3339",
			timeStr: "2025-01-24T10:00:00Z",
		},
		{
			name:    "RFC3339Nano",
			timeStr: "2025-01-24T10:00:00.123456789Z",
		},
		{
			name:    "Custom format 1",
			timeStr: "2025-01-24T10:00:00.000Z",
		},
		{
			name:    "Custom format 2",
			timeStr: "2025-01-24 10:00:00",
		},
		{
			name:    "Invalid format",
			timeStr: "24/01/2025 10:00",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseTime(tc.timeStr)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}
