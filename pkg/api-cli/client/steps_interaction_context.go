package client

import (
	"context"
	"errors"
	"fmt"
)

// createStepWithCustomBodyContext is a context-aware version of createStepWithCustomBody
func (c *Client) createStepWithCustomBodyContext(ctx context.Context, checkpointID int, parsedStepBody map[string]interface{}, position int) (int, error) {
	body := map[string]interface{}{
		"checkpointId": checkpointID,
		"stepIndex":    position,
		"parsedStep":   parsedStepBody,
	}

	var response struct {
		Item struct {
			ID int `json:"id"`
		} `json:"item"`
		Error string `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&response).
		Post("/teststeps?envelope=false")

	if err != nil {
		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return 0, NewClientError("createStepWithCustomBodyContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return 0, NewClientError("createStepWithCustomBodyContext", KindTimeout, "request timeout", err)
		}
		return 0, NewClientError("createStepWithCustomBodyContext", KindConnectionFailed, "create step request failed", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return 0, NewAPIError(resp.StatusCode(), ErrCodeBadRequest, response.Error)
		}
		return 0, c.handleErrorResponse("createStepWithCustomBodyContext", resp)
	}

	return response.Item.ID, nil
}

// CreateStepClickWithContext creates a click step with context support
func (c *Client) CreateStepClickWithContext(ctx context.Context, checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "CLICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepWriteWithContext creates a write/input step with context support
func (c *Client) CreateStepWriteWithContext(ctx context.Context, checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "WRITE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": value,
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepClickWithVariableWithContext creates a click step with variable target with context support
func (c *Client) CreateStepClickWithVariableWithContext(ctx context.Context, checkpointID int, variable string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"","variable":"%s"}`, variable)

	parsedStep := map[string]interface{}{
		"action": "CLICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepClickWithDetailsWithContext creates a click step with position and element type with context support
func (c *Client) CreateStepClickWithDetailsWithContext(ctx context.Context, checkpointID int, selector, positionType, elementType string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s","position":"%s","elementType":"%s"}`, selector, positionType, elementType)

	parsedStep := map[string]interface{}{
		"action": "CLICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepWriteWithVariableWithContext creates a write step with variable storage with context support
func (c *Client) CreateStepWriteWithVariableWithContext(ctx context.Context, checkpointID int, selector, value, variable string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "WRITE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value":    value,
		"meta":     map[string]interface{}{},
		"variable": variable,
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepDoubleClickWithContext creates a double-click step with context support
func (c *Client) CreateStepDoubleClickWithContext(ctx context.Context, checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta": map[string]interface{}{
			"action": "DOUBLE_CLICK",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepRightClickWithContext creates a right-click step with context support
func (c *Client) CreateStepRightClickWithContext(ctx context.Context, checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta": map[string]interface{}{
			"action": "RIGHT_CLICK",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepHoverWithContext creates a hover step with context support
func (c *Client) CreateStepHoverWithContext(ctx context.Context, checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta": map[string]interface{}{
			"action": "OVER",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepKeyGlobalWithContext creates a global key press step with context support
func (c *Client) CreateStepKeyGlobalWithContext(ctx context.Context, checkpointID int, key string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "KEY",
		"value":  key,
		"meta":   map[string]interface{}{},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepKeyTargetedWithContext creates a targeted key press step with context support
func (c *Client) CreateStepKeyTargetedWithContext(ctx context.Context, checkpointID int, selector, key string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "KEY",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": key,
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}
