package client

import (
	"context"
	"errors"
	"fmt"
)

// addStepContext is a context-aware version of addStep
func (c *Client) addStepContext(ctx context.Context, checkpointID int, stepIndex int, parsedStep map[string]interface{}) (int, error) {
	body := map[string]interface{}{
		"checkpointId": checkpointID,
		"stepIndex":    stepIndex,
		"parsedStep":   parsedStep,
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
			return 0, NewClientError("addStepContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return 0, NewClientError("addStepContext", KindTimeout, "request timeout", err)
		}
		return 0, NewClientError("addStepContext", KindConnectionFailed, "add step request failed", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return 0, NewAPIError(resp.StatusCode(), ErrCodeBadRequest, response.Error)
		}
		return 0, c.handleErrorResponse("addStepContext", resp)
	}

	return response.Item.ID, nil
}

// Data Storage Methods with Context

// CreateStepStoreWithContext creates a general store step with context
func (c *Client) CreateStepStoreWithContext(ctx context.Context, checkpointID int, selector, variableName string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "STORE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value":    "",
		"variable": variableName,
		"meta":     map[string]interface{}{},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepStoreElementTextWithContext creates a step to store element text in variable with context
func (c *Client) CreateStepStoreElementTextWithContext(ctx context.Context, checkpointID int, selector, variableName string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "STORE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value":    "",
		"meta":     map[string]interface{}{},
		"variable": variableName,
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepStoreLiteralValueWithContext creates a step to store literal value in variable with context
func (c *Client) CreateStepStoreLiteralValueWithContext(ctx context.Context, checkpointID int, value, variableName string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action":   "STORE",
		"value":    value,
		"meta":     map[string]interface{}{},
		"variable": variableName,
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepStoreAttributeWithContext creates a step to store element attribute value with context
func (c *Client) CreateStepStoreAttributeWithContext(ctx context.Context, checkpointID int, selector string, attribute string, variable string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "STORE",
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
			"attribute": attribute,
		},
		"variable": variable,
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepStoreValueWithContext creates a step to store element value with context
func (c *Client) CreateStepStoreValueWithContext(ctx context.Context, checkpointID int, selector, variableName string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "STORE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value":    "",
		"variable": variableName,
		"meta": map[string]interface{}{
			"type": "VALUE",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// Cookie Methods with Context

// CreateStepCookieCreateWithContext creates a cookie with the specified name and value with context
func (c *Client) CreateStepCookieCreateWithContext(ctx context.Context, checkpointID int, name, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"value":  value,
		"meta": map[string]interface{}{
			"type": "ADD",
			"name": name,
		},
	}

	return c.addStepContext(ctx, checkpointID, position, parsedStep)
}

// CreateStepCookieCreateWithOptionsWithContext creates a cookie with specified options with context
func (c *Client) CreateStepCookieCreateWithOptionsWithContext(ctx context.Context, checkpointID int, name, value string, options map[string]interface{}, position int) (int, error) {
	meta := map[string]interface{}{
		"type": "ADD",
		"name": name,
	}

	// Add optional fields if provided
	if domain, ok := options["domain"].(string); ok && domain != "" {
		meta["domain"] = domain
	}
	if path, ok := options["path"].(string); ok && path != "" {
		meta["path"] = path
	}
	if secure, ok := options["secure"].(bool); ok && secure {
		meta["secure"] = true
	}
	if httpOnly, ok := options["httpOnly"].(bool); ok && httpOnly {
		meta["httpOnly"] = true
	}

	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"value":  value,
		"meta":   meta,
	}

	return c.addStepContext(ctx, checkpointID, position, parsedStep)
}

// CreateStepCookieDeleteWithContext creates a step to delete a specific cookie with context
func (c *Client) CreateStepCookieDeleteWithContext(ctx context.Context, checkpointID int, name string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"meta": map[string]interface{}{
			"type": "DELETE",
			"name": name,
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepCookieClearAllWithContext clears all cookies with context
func (c *Client) CreateStepCookieClearAllWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"meta": map[string]interface{}{
			"type": "CLEAR",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// Dialog Methods with Context

// CreateStepAlertAcceptWithContext creates a step to accept an alert dialog with context
func (c *Client) CreateStepAlertAcceptWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"type":   "ALERT",
			"action": "OK",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepAlertDismissWithContext creates a step to dismiss an alert dialog with context
func (c *Client) CreateStepAlertDismissWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "ALERT",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepConfirmAcceptWithContext creates a step to accept a confirm dialog with context
func (c *Client) CreateStepConfirmAcceptWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"type":   "CONFIRM",
			"action": "OK",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepConfirmDismissWithContext creates a step to dismiss (cancel) a confirm dialog with context
func (c *Client) CreateStepConfirmDismissWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"type":   "CONFIRM",
			"action": "CANCEL",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepPromptDismissWithContext creates a step to dismiss a prompt dialog with context
func (c *Client) CreateStepPromptDismissWithContext(ctx context.Context, checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"type":   "PROMPT",
			"action": "OK",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepPromptDismissWithTextWithContext creates a step to dismiss a prompt with response text with context
func (c *Client) CreateStepPromptDismissWithTextWithContext(ctx context.Context, checkpointID int, text string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  text,
		"meta": map[string]interface{}{
			"type":   "PROMPT",
			"action": "OK",
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// File Upload Methods with Context

// CreateStepFileUploadWithContext creates a file upload step with context
func (c *Client) CreateStepFileUploadWithContext(ctx context.Context, checkpointID int, selector, filePath string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "UPLOAD",
		"value":  filePath,
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"meta": map[string]interface{}{},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateStepFileUploadByURLWithContext creates a step to upload a file from URL with context
func (c *Client) CreateStepFileUploadByURLWithContext(ctx context.Context, checkpointID int, url, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "UPLOAD",
		"value":  "",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"meta": map[string]interface{}{
			"url": url,
		},
	}

	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}
