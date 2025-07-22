package client

import (
	"context"
	"fmt"
)

// Response types for the context methods

// StepResponse represents the response from creating a step
type StepResponse struct {
	ID           string                 `json:"id"`
	CheckpointID string                 `json:"checkpoint_id"`
	Position     int                    `json:"position"`
	Extension    map[string]interface{} `json:"extension"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// LibraryCheckpointResponse represents the response from library checkpoint operations
type LibraryCheckpointResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	CheckpointID string `json:"checkpoint_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// AttachLibraryResponse represents the response from attaching a library checkpoint
type AttachLibraryResponse struct {
	JourneyID           string `json:"journey_id"`
	LibraryCheckpointID string `json:"library_checkpoint_id"`
	Position            int    `json:"position"`
	AttachedAt          string `json:"attached_at"`
}

// Window operations with context

// CreateStepResizeWindowWithContext creates a step to resize the browser window with context support
func (c *Client) CreateStepResizeWindowWithContext(ctx context.Context, checkpointID string, width, height int, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "resizeWindow",
			"context": map[string]interface{}{
				"width":  width,
				"height": height,
			},
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create resize window step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepMaximizeWithContext creates a step to maximize the browser window with context support
func (c *Client) CreateStepMaximizeWithContext(ctx context.Context, checkpointID string, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "maximizeWindow",
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create maximize window step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepWindowSwitchWithContext creates a step to switch windows with context support
func (c *Client) CreateStepWindowSwitchWithContext(ctx context.Context, checkpointID string, action string, value interface{}, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "switchWindow",
			"context": map[string]interface{}{
				"action": action,
				"value":  value,
			},
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create window switch step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepSwitchTabWithContext creates a step to switch browser tabs with context support
func (c *Client) CreateStepSwitchTabWithContext(ctx context.Context, checkpointID string, tabID string, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "switchTab",
			"context": map[string]interface{}{
				"tabId": tabID,
			},
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create switch tab step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepSwitchIframeWithContext creates a step to switch to an iframe with context support
func (c *Client) CreateStepSwitchIframeWithContext(ctx context.Context, checkpointID string, iframeSelector string, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "switchIframe",
			"context": map[string]interface{}{
				"selector": iframeSelector,
			},
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create switch iframe step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepSwitchParentFrameWithContext creates a step to switch to parent frame with context support
func (c *Client) CreateStepSwitchParentFrameWithContext(ctx context.Context, checkpointID string, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "switchParentFrame",
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create switch parent frame step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepSetViewportWithContext creates a step to set viewport size with context support
func (c *Client) CreateStepSetViewportWithContext(ctx context.Context, checkpointID string, width, height int, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "setViewport",
			"context": map[string]interface{}{
				"width":  width,
				"height": height,
			},
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create set viewport step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// Miscellaneous operations with context

// CreateStepCommentWithContext creates a comment step with context support
func (c *Client) CreateStepCommentWithContext(ctx context.Context, checkpointID string, comment string, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "comment",
			"context": map[string]interface{}{
				"text": comment,
			},
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create comment step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepExecuteScriptWithContext creates a step to execute JavaScript with context support
func (c *Client) CreateStepExecuteScriptWithContext(ctx context.Context, checkpointID string, script string, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name": "executeScript",
			"context": map[string]interface{}{
				"script": script,
			},
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create execute script step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepEnvironmentWithContext creates an environment configuration step with context support
func (c *Client) CreateStepEnvironmentWithContext(ctx context.Context, checkpointID string, envConfig map[string]interface{}, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name":    "environment",
			"context": envConfig,
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create environment step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// CreateStepConfigureWithContext creates a configuration step with context support
func (c *Client) CreateStepConfigureWithContext(ctx context.Context, checkpointID string, config map[string]interface{}, position int) (*StepResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"position":      position,
		"extension": map[string]interface{}{
			"name":    "configure",
			"context": config,
		},
	}

	var result StepResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to create configure step: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// Session/checkpoint operations with context

// SetCheckpointWithContext sets the current checkpoint with context support
func (c *Client) SetCheckpointWithContext(ctx context.Context, checkpointID string) error {
	// This is a client-side operation, check context first
	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	c.currentCheckpoint = checkpointID
	return nil
}

// AddCheckpointToLibraryWithContext adds a checkpoint to the library with context support
func (c *Client) AddCheckpointToLibraryWithContext(ctx context.Context, checkpointID, name, description string) (*LibraryCheckpointResponse, error) {
	payload := map[string]interface{}{
		"checkpoint_id": checkpointID,
		"name":          name,
		"description":   description,
	}

	var result LibraryCheckpointResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/library/checkpoints")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to add checkpoint to library: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// GetLibraryCheckpointWithContext retrieves a library checkpoint with context support
func (c *Client) GetLibraryCheckpointWithContext(ctx context.Context, libraryCheckpointID string) (*LibraryCheckpointResponse, error) {
	var result LibraryCheckpointResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/library/checkpoints/%s", libraryCheckpointID))

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to get library checkpoint: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}

// AttachLibraryToJourneyWithContext attaches a library checkpoint to a journey with context support
func (c *Client) AttachLibraryToJourneyWithContext(ctx context.Context, journeyID, libraryCheckpointID string, position int) (*AttachLibraryResponse, error) {
	payload := map[string]interface{}{
		"journey_id":            journeyID,
		"library_checkpoint_id": libraryCheckpointID,
		"position":              position,
	}

	var result AttachLibraryResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/journeys/attach-library")

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to attach library to journey: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return &result, nil
}
