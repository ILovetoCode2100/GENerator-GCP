package client

import (
	"context"
	"fmt"
	"time"
)

// This file contains fixes for API response parsing issues
// These functions should replace the existing implementations in client.go

// ExecuteGoalFixed executes a goal and handles flexible response formats
func (c *Client) ExecuteGoalFixed(goalID, snapshotID int) (*Execution, error) {
	body := map[string]interface{}{
		"goalId":     goalID,
		"snapshotId": snapshotID,
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		Post(fmt.Sprintf("/goals/%d/snapshots/%d/execute", goalID, snapshotID))

	if err != nil {
		return nil, fmt.Errorf("execute goal request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("execute goal failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Use the new ResponseHandler for robust parsing
	handler := NewResponseHandler(c.config.Output.Verbose)
	var execution Execution

	if err := handler.ParseResponse(resp.Body(), &execution); err != nil {
		// Check if we have a partial success
		if IsNoIDButSuccessError(err) {
			c.logger.Warn("Execution may have been created but no ID returned")
			return &execution, nil
		}
		return nil, fmt.Errorf("failed to parse execution response: %w", err)
	}

	return &execution, nil
}

// addStepFixed is an improved version that handles various response formats
func (c *Client) addStepFixed(checkpointID int, stepIndex int, parsedStep map[string]interface{}) (int, error) {
	body := map[string]interface{}{
		"checkpointId": checkpointID,
		"stepIndex":    stepIndex,
		"parsedStep":   parsedStep,
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		Post("/teststeps?envelope=false")

	if err != nil {
		return 0, fmt.Errorf("add step request failed: %w", err)
	}

	if resp.IsError() {
		return 0, fmt.Errorf("add step failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Use the new ResponseHandler to extract step ID
	handler := NewResponseHandler(c.config.Output.Verbose)
	stepID, err := handler.ParseStepID(resp.Body())
	if err != nil {
		// Check different error types
		if IsPlaceholderError(err) {
			c.logger.Warnf("Step created with placeholder ID: %v", err)
			return stepID, nil // Return the ID even if it's a placeholder
		}
		if IsNoIDButSuccessError(err) {
			c.logger.Warn("Step appears to be created but no ID returned")
			return 0, nil
		}
		return 0, fmt.Errorf("failed to parse step response: %w", err)
	}

	return stepID, nil
}

// createStepWithCustomBodyFixed handles step creation with better response parsing
func (c *Client) createStepWithCustomBodyFixed(checkpointID int, stepData map[string]interface{}, position int) (int, error) {
	return c.addStepFixed(checkpointID, position, stepData)
}

// Helper function to create step with retry support
func (c *Client) createStepWithRetry(checkpointID int, stepData map[string]interface{}, position int) (int, error) {
	retryConfig := DefaultRetryConfig()
	retryConfig.MaxAttempts = 3
	retryConfig.InitialDelay = 500 * time.Millisecond

	var stepID int
	var lastErr error

	err := RetryWithBackoff(context.Background(), retryConfig, func() error {
		id, err := c.addStepFixed(checkpointID, position, stepData)
		if err != nil {
			lastErr = err
			return err
		}
		stepID = id
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("create step failed after retries: %w", lastErr)
	}

	return stepID, nil
}
