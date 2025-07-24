package client

import (
	"fmt"
)

// This file shows how to integrate ResponseHandler into existing methods
// These are drop-in replacements for problematic methods

// ExecuteGoalRobust is a robust version of ExecuteGoal that handles all response variations
func (c *Client) ExecuteGoalRobust(goalID, snapshotID int) (*Execution, error) {
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

	// Use ResponseHandler for robust parsing
	handler := NewResponseHandler(c.config.Output.Verbose)
	var execution Execution

	if err := handler.ParseResponse(resp.Body(), &execution); err != nil {
		// Still return execution if we have partial data
		if IsNoIDButSuccessError(err) {
			c.logger.Warn("Goal execution started but no ID returned")
			return &execution, nil
		}
		return nil, fmt.Errorf("failed to parse execution response: %w", err)
	}

	// Extract execution ID if needed for compatibility
	if execution.ID == "" {
		if id, err := handler.ParseExecutionID(resp.Body()); err == nil {
			execution.ID = id
		}
	}

	return &execution, nil
}

// CreateStepRobust handles all step creation with robust response parsing
func (c *Client) CreateStepRobust(checkpointID int, stepType string, stepData map[string]interface{}, position int) (int, error) {
	// Ensure we have the basic structure
	if stepData == nil {
		stepData = make(map[string]interface{})
	}
	stepData["type"] = stepType

	body := map[string]interface{}{
		"checkpointId": checkpointID,
		"stepIndex":    position,
		"parsedStep":   stepData,
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		Post("/teststeps?envelope=false")

	if err != nil {
		return 0, fmt.Errorf("create step request failed: %w", err)
	}

	if resp.IsError() {
		return 0, fmt.Errorf("create step failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Parse step ID with full error handling
	handler := NewResponseHandler(c.config.Output.Verbose)
	stepID, err := handler.ParseStepID(resp.Body())

	if err != nil {
		// Handle known response quirks
		switch {
		case IsPlaceholderError(err):
			// Common for dialog and mouse commands
			c.logger.Debugf("Step created with placeholder ID for %s step", stepType)
			if stepType == "dismiss-alert" || stepType == "mouse-move" {
				// Expected for these types
				return 0, nil
			}
			return stepID, nil // Return ID for other types

		case IsNoIDButSuccessError(err):
			// Step created but no ID in response
			c.logger.Infof("Step of type %s created successfully", stepType)
			return 0, nil

		default:
			// Actual error
			return 0, fmt.Errorf("failed to parse step response: %w", err)
		}
	}

	// Validate non-placeholder IDs
	if err := ValidateStepResponse(stepID, nil); err != nil {
		c.logger.Warnf("Step ID validation warning: %v", err)
	}

	return stepID, nil
}

// Convenience methods for specific step types that handle their quirks

// CreateDialogStepRobust handles dialog steps which often return placeholder IDs
func (c *Client) CreateDialogStepRobust(checkpointID int, dialogType string, acceptReject bool, text string, position int) (int, error) {
	stepData := map[string]interface{}{
		"type": dialogType,
	}

	if dialogType == "dismiss-confirm" || dialogType == "dismiss-prompt" {
		stepData["acceptReject"] = acceptReject
	}

	if dialogType == "dismiss-prompt-with-text" && text != "" {
		stepData["text"] = text
	}

	// Dialog steps typically return placeholder IDs
	_, err := c.CreateStepRobust(checkpointID, dialogType, stepData, position)
	if err != nil {
		return 0, err
	}

	// Don't return placeholder IDs for dialog steps
	return 0, nil
}

// CreateMouseStepRobust handles mouse steps which may return placeholder IDs
func (c *Client) CreateMouseStepRobust(checkpointID int, action string, target interface{}, position int) (int, error) {
	stepData := map[string]interface{}{
		"action": action,
	}

	switch action {
	case "move-to":
		stepData["selector"] = target
	case "move-by":
		stepData["coordinates"] = target
	}

	// Mouse steps may return placeholder IDs
	stepID, err := c.CreateStepRobust(checkpointID, "mouse", stepData, position)
	if err != nil {
		return 0, err
	}

	// Return 0 if placeholder, actual ID otherwise
	if stepID == 1 {
		return 0, nil
	}

	return stepID, nil
}

// BatchCreateStepsRobust handles creating multiple steps with robust error handling
func (c *Client) BatchCreateStepsRobust(checkpointID int, steps []map[string]interface{}) ([]int, error) {
	var createdIDs []int
	handler := NewResponseHandler(c.config.Output.Verbose)

	for i, step := range steps {
		position := i + 1 // 1-based positioning

		resp, err := c.httpClient.R().
			SetBody(map[string]interface{}{
				"checkpointId": checkpointID,
				"stepIndex":    position,
				"parsedStep":   step,
			}).
			Post("/teststeps?envelope=false")

		if err != nil {
			c.logger.Errorf("Failed to create step %d: %v", position, err)
			continue
		}

		if resp.IsError() {
			c.logger.Errorf("Step %d creation failed with status %d", position, resp.StatusCode())
			continue
		}

		// Try to get step ID
		if stepID, err := handler.ParseStepID(resp.Body()); err == nil {
			createdIDs = append(createdIDs, stepID)
		} else if IsNoIDButSuccessError(err) {
			c.logger.Debugf("Step %d created but no ID returned", position)
			createdIDs = append(createdIDs, 0)
		} else if IsPlaceholderError(err) {
			c.logger.Debugf("Step %d created with placeholder ID", position)
			createdIDs = append(createdIDs, 0)
		}
	}

	if len(createdIDs) == 0 {
		return nil, fmt.Errorf("failed to create any steps")
	}

	return createdIDs, nil
}
