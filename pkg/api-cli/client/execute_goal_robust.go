package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ExecutionRobust is a more flexible version of Execution that handles various ID formats
type ExecutionRobust struct {
	ID          interface{} `json:"id"`          // Can be string or number
	ExecutionID interface{} `json:"executionId"` // Alternative field name
	Status      string      `json:"status"`
	Progress    int         `json:"progress"`
	StartTime   interface{} `json:"startTime"` // Can be string or timestamp
	EndTime     interface{} `json:"endTime"`
	GoalID      int         `json:"goalId"`
	SnapshotID  string      `json:"snapshotId"`
	ResultsURL  string      `json:"resultsUrl"`
	Error       string      `json:"error,omitempty"`
}

// ToExecution converts ExecutionRobust to standard Execution
func (er *ExecutionRobust) ToExecution() (*Execution, error) {
	handler := NewResponseHandler()

	// Normalize ID
	idStr := ""
	if er.ID != nil {
		idStr = handler.normalizeExecutionID(er.ID)
	} else if er.ExecutionID != nil {
		idStr = handler.normalizeExecutionID(er.ExecutionID)
	}

	// Parse times
	var startTime, endTime time.Time
	var err error

	if er.StartTime != nil {
		startTime, err = handler.ParseTimeFormat(er.StartTime)
		if err != nil {
			// Use zero time if parsing fails
			startTime = time.Time{}
		}
	}

	if er.EndTime != nil {
		endTime, err = handler.ParseTimeFormat(er.EndTime)
		if err != nil {
			endTime = time.Time{}
		}
	}

	return &Execution{
		ID:          idStr,
		ExecutionID: idStr,
		Status:      er.Status,
		Progress:    er.Progress,
		StartTime:   startTime,
		EndTime:     endTime,
		GoalID:      er.GoalID,
		SnapshotID:  er.SnapshotID,
		ResultsURL:  er.ResultsURL,
	}, nil
}

// ExecuteGoalRobust is a more robust version of ExecuteGoal that handles various response formats
func (c *Client) ExecuteGoalRobust(goalID int, snapshotID string) (*Execution, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return c.ExecuteGoalRobustWithContext(ctx, goalID, snapshotID)
}

// ExecuteGoalRobustWithContext is the context-aware version
func (c *Client) ExecuteGoalRobustWithContext(ctx context.Context, goalID int, snapshotID string) (*Execution, error) {
	body := map[string]interface{}{
		"goalId": goalID,
	}

	if snapshotID != "" {
		body["snapshotId"] = snapshotID
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(body).
		Post("/executions")

	if err != nil {
		return nil, fmt.Errorf("execute goal request failed: %w", err)
	}

	if resp.IsError() {
		return nil, c.handleErrorResponse("ExecuteGoalRobust", resp)
	}

	// Use the ResponseHandler to parse the execution
	handler := NewResponseHandler()
	execution, err := handler.ParseExecutionResponse(resp.Body())
	if err != nil {
		// Try alternative parsing
		var robustResp struct {
			Item      *ExecutionRobust `json:"item"`
			Execution *ExecutionRobust `json:"execution"`
			Data      *ExecutionRobust `json:"data"`
			Success   bool             `json:"success"`
			Error     string           `json:"error"`
		}

		if parseErr := json.Unmarshal(resp.Body(), &robustResp); parseErr != nil {
			return nil, fmt.Errorf("failed to parse execution response: %w (original: %w)", parseErr, err)
		}

		// Find the execution in various locations
		var execRobust *ExecutionRobust
		if robustResp.Item != nil {
			execRobust = robustResp.Item
		} else if robustResp.Execution != nil {
			execRobust = robustResp.Execution
		} else if robustResp.Data != nil {
			execRobust = robustResp.Data
		}

		if execRobust == nil {
			// Try direct parsing
			var directExec ExecutionRobust
			if directErr := json.Unmarshal(resp.Body(), &directExec); directErr == nil {
				execRobust = &directExec
			}
		}

		if execRobust != nil {
			return execRobust.ToExecution()
		}

		return nil, fmt.Errorf("no execution data found in response: %s", string(resp.Body()))
	}

	return execution, nil
}

// CreateStepRobust is a universal step creation method that handles all response variations
func (c *Client) CreateStepRobust(ctx context.Context, checkpointID int, stepType string, stepData map[string]interface{}, position int) (int, error) {
	// Use the existing createStepWithCustomBodyContext which already has some robustness
	stepID, err := c.createStepWithCustomBodyContext(ctx, checkpointID, stepData, position)

	if err != nil {
		// Check if it's a placeholder ID error
		if _, ok := err.(*PlaceholderIDError); ok {
			// Log warning but return the ID
			if c.config != nil && c.config.Output.Verbose {
				fmt.Printf("Warning: Placeholder ID returned for %s step\n", stepType)
			}
			return stepID, nil
		}

		// Check if it's a success without ID
		if _, ok := err.(*NoIDButSuccessError); ok {
			// Generate a temporary ID or return a special value
			if c.config != nil && c.config.Output.Verbose {
				fmt.Printf("Info: Step created successfully but no ID returned\n")
			}
			return 0, nil // Or return a special success value
		}
	}

	return stepID, err
}

// BatchCreateStepsRobust creates multiple steps with robust error handling
func (c *Client) BatchCreateStepsRobust(ctx context.Context, checkpointID int, steps []StepDefinition) ([]int, []error) {
	ids := make([]int, len(steps))
	errors := make([]error, len(steps))

	for i, step := range steps {
		stepData := map[string]interface{}{
			"action": step.Action,
			"target": step.Target,
			"value":  step.Value,
			"meta":   step.Meta,
		}

		id, err := c.CreateStepRobust(ctx, checkpointID, step.Action, stepData, step.Position)
		ids[i] = id
		errors[i] = err

		// Continue even if one step fails
		if err != nil && c.config != nil && c.config.Output.Verbose {
			fmt.Printf("Warning: Step %d failed: %v\n", i, err)
		}
	}

	return ids, errors
}

// StepDefinition defines a step to be created
type StepDefinition struct {
	Action   string                 `json:"action"`
	Target   map[string]interface{} `json:"target,omitempty"`
	Value    string                 `json:"value,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
	Position int                    `json:"position"`
}
