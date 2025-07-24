package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/marklovelady/api-cli-generator/pkg/api-cli/constants"
)

// Navigation methods with context

// CreateStepNavigateWithContext creates a navigation step with context support (Version B)
func (c *Client) CreateStepNavigateWithContext(ctx context.Context, checkpointID int, url string, useNewTab bool, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "NAVIGATE",
		"value":  url,
		"meta": map[string]interface{}{
			"useNewTab": useNewTab,
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateNavigationStepWithContext creates a navigation step at a specific position with context support
func (c *Client) CreateNavigationStepWithContext(ctx context.Context, checkpointID int, url string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "NAVIGATE",
		"target": map[string]interface{}{
			"url": url,
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// AddNavigateStepWithContext adds a navigation step to a checkpoint with context support
func (c *Client) AddNavigateStepWithContext(ctx context.Context, checkpointID int, url string) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "NAVIGATE",
		"target": map[string]interface{}{
			"url": url,
		},
	}
	// Get the next position
	position, err := c.getNextPositionContext(ctx, checkpointID)
	if err != nil {
		return 0, err
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// Wait methods with context

// CreateStepWaitForElementTimeoutWithContext creates a step to wait for element with custom timeout and context support (Version B)
func (c *Client) CreateStepWaitForElementTimeoutWithContext(ctx context.Context, checkpointID int, selector string, timeoutMs int, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)
	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"kind":  "AI",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta": map[string]interface{}{
			"type":      "ELEMENT_VISIBLE",
			"timeoutMs": timeoutMs,
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepWaitForElementNotVisibleWithContext creates a step to wait for element to disappear with context support (Version B)
func (c *Client) CreateStepWaitForElementNotVisibleWithContext(ctx context.Context, checkpointID int, selector string, timeoutMs int, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)
	// Use default timeout if none specified
	if timeoutMs <= 0 {
		timeoutMs = constants.DefaultTimeoutMs
	}
	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"kind":  "AI",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta": map[string]interface{}{
			"type":      "ELEMENT_NOT_VISIBLE",
			"timeoutMs": timeoutMs,
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateWaitTimeStepWithContext creates a wait time step at a specific position with context support
func (c *Client) CreateWaitTimeStepWithContext(ctx context.Context, checkpointID int, seconds int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"value":  fmt.Sprintf("%d", seconds*constants.MillisecondsPerSecond), // Convert seconds to milliseconds
		"meta": map[string]interface{}{
			"type": "TIME",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepWaitForElementWithContext creates a step to wait for element with default timeout and context support
func (c *Client) CreateStepWaitForElementWithContext(ctx context.Context, checkpointID int, selector string, position int) (int, error) {
	return c.CreateStepWaitForElementTimeoutWithContext(ctx, checkpointID, selector, constants.DefaultTimeoutMs, position)
}

// Scroll methods with context

// CreateScrollStepWithContext creates a generic scroll step with context support
func (c *Client) CreateScrollStepWithContext(ctx context.Context, checkpointID int, scrollType string, element string, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "SCROLL",
		},
	}

	// Add scroll type specific configuration
	switch scrollType {
	case "TOP":
		parsedStep["meta"].(map[string]interface{})["type"] = "TOP"
	case "BOTTOM":
		parsedStep["meta"].(map[string]interface{})["type"] = "BOTTOM"
	case "ELEMENT":
		clueJSON := fmt.Sprintf(`{"clue":"%s"}`, element)
		parsedStep["target"] = map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"kind":  "AI",
					"value": clueJSON,
				},
			},
		}
		parsedStep["meta"].(map[string]interface{})["type"] = "ELEMENT"
	case "POSITION":
		parsedStep["meta"].(map[string]interface{})["type"] = "POSITION"
		parsedStep["meta"].(map[string]interface{})["x"] = x
		parsedStep["meta"].(map[string]interface{})["y"] = y
	case "BY":
		parsedStep["meta"].(map[string]interface{})["type"] = "BY"
		parsedStep["meta"].(map[string]interface{})["x"] = x
		parsedStep["meta"].(map[string]interface{})["y"] = y
	default:
		return 0, fmt.Errorf("unknown scroll type: %s", scrollType)
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// AddScrollToBottomStepWithContext adds a scroll to bottom step with context support
func (c *Client) AddScrollToBottomStepWithContext(ctx context.Context, checkpointID int) (int, error) {
	// Get the next position
	position, err := c.getNextPositionContext(ctx, checkpointID)
	if err != nil {
		return 0, err
	}
	return c.CreateScrollStepWithContext(ctx, checkpointID, "BOTTOM", "", 0, 0, position)
}

// AddScrollToTopStepWithContext adds a scroll to top step with context support
func (c *Client) AddScrollToTopStepWithContext(ctx context.Context, checkpointID int) (int, error) {
	// Get the next position
	position, err := c.getNextPositionContext(ctx, checkpointID)
	if err != nil {
		return 0, err
	}
	return c.CreateScrollStepWithContext(ctx, checkpointID, "TOP", "", 0, 0, position)
}

// CreateScrollTopStepWithContext creates a scroll to top step at a specific position with context support
func (c *Client) CreateScrollTopStepWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "SCROLL",
			"type": "TOP",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateScrollBottomStepWithContext creates a scroll to bottom step at a specific position with context support
func (c *Client) CreateScrollBottomStepWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "SCROLL",
			"type": "BOTTOM",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateScrollElementStepWithContext creates a scroll to element step at a specific position with context support
func (c *Client) CreateScrollElementStepWithContext(ctx context.Context, checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": "",
		"meta": map[string]interface{}{
			"type": "ELEMENT",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateScrollPositionStepWithContext creates a scroll to position step with context support
func (c *Client) CreateScrollPositionStepWithContext(ctx context.Context, checkpointID int, x int, y int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "SCROLL",
			"type": "POSITION",
			"x":    x,
			"y":    y,
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// Version B scroll methods with context

// CreateStepScrollToPositionWithContext creates a scroll to position step with context support (Version B)
func (c *Client) CreateStepScrollToPositionWithContext(ctx context.Context, checkpointID int, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"x":    x,
			"y":    y,
			"type": "POSITION",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepScrollByOffsetWithContext creates a scroll by offset step with context support (Version B)
func (c *Client) CreateStepScrollByOffsetWithContext(ctx context.Context, checkpointID int, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"x":    x,
			"y":    y,
			"type": "OFFSET",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepScrollToTopWithContext creates a scroll to top step with context support (Version B)
func (c *Client) CreateStepScrollToTopWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "TOP",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepScrollBottomWithContext creates a scroll to bottom step with context support
func (c *Client) CreateStepScrollBottomWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "BOTTOM",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepScrollElementWithContext creates a scroll to element step with context support
func (c *Client) CreateStepScrollElementWithContext(ctx context.Context, checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"kind":  "AI",
					"value": clueJSON,
				},
			},
		},
		"value": "",
		"meta": map[string]interface{}{
			"type": "ELEMENT",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepScrollPositionWithContext creates a scroll to position step with context support
func (c *Client) CreateStepScrollPositionWithContext(ctx context.Context, checkpointID int, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"x":    x,
			"y":    y,
			"type": "POSITION",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepScrollTopWithContext creates a scroll to top step with context support
func (c *Client) CreateStepScrollTopWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "TOP",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// getNextPositionContext is a context-aware version of getNextPosition
func (c *Client) getNextPositionContext(ctx context.Context, checkpointID int) (int, error) {
	steps, err := c.ListCheckpointStepsWithContext(ctx, checkpointID)
	if err != nil {
		return 0, fmt.Errorf("failed to get current steps: %w", err)
	}
	return len(steps) + 1, nil
}

// ListCheckpointStepsWithContext is a context-aware version of ListCheckpointSteps
func (c *Client) ListCheckpointStepsWithContext(ctx context.Context, checkpointID int) ([]Step, error) {
	var response struct {
		Items []Step `json:"items"`
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		Get(fmt.Sprintf("/checkpoints/%d/teststeps", checkpointID))

	if err != nil {
		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return nil, NewClientError("ListCheckpointStepsWithContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewClientError("ListCheckpointStepsWithContext", KindTimeout, "request timeout", err)
		}
		return nil, NewClientError("ListCheckpointStepsWithContext", KindConnectionFailed, "list checkpoint steps request failed", err)
	}

	if resp.IsError() {
		return nil, c.handleErrorResponse("ListCheckpointStepsWithContext", resp)
	}

	return response.Items, nil
}
