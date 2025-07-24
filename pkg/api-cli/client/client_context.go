package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/constants"
)

// CreateProjectWithContext creates a new project with context support
func (c *Client) CreateProjectWithContext(ctx context.Context, name, description string) (*Project, error) {
	// Convert org ID to int
	orgID, err := strconv.Atoi(c.config.Org.ID)
	if err != nil {
		return nil, NewClientError("CreateProjectWithContext", KindInvalidInput, "invalid organization ID", err)
	}

	body := map[string]interface{}{
		"name":           name,
		"description":    description,
		"organizationId": orgID,
	}

	var result Project
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader(constants.HeaderAuthorization, constants.AuthorizationHeaderPrefix+c.config.API.AuthToken).
		SetBody(body).
		SetResult(&result).
		Post("/projects")

	if err != nil {
		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return nil, NewClientError("CreateProjectWithContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewClientError("CreateProjectWithContext", KindTimeout, "request timeout", err)
		}
		return nil, NewClientError("CreateProjectWithContext", KindConnectionFailed, "failed to create project", err)
	}

	if !resp.IsSuccess() {
		return nil, c.handleErrorResponse("CreateProjectWithContext", resp)
	}

	return &result, nil
}

// CreateGoalWithContext creates a new goal with context support
func (c *Client) CreateGoalWithContext(ctx context.Context, projectID int, name, url string) (*Goal, error) {
	body := map[string]interface{}{
		"projectId":     projectID,
		"name":          name,
		"environmentId": nil,
		"url":           url,
	}

	var result struct {
		Goal     Goal `json:"goal"`
		Snapshot struct {
			ID int `json:"id"`
		} `json:"snapshot"`
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader(constants.HeaderAuthorization, constants.AuthorizationHeaderPrefix+c.config.API.AuthToken).
		SetBody(body).
		SetResult(&result).
		Post("/goals")

	if err != nil {
		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return nil, NewClientError("CreateGoalWithContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewClientError("CreateGoalWithContext", KindTimeout, "request timeout", err)
		}
		return nil, NewClientError("CreateGoalWithContext", KindConnectionFailed, "failed to create goal", err)
	}

	if !resp.IsSuccess() {
		return nil, c.handleErrorResponse("CreateGoalWithContext", resp)
	}

	// Store snapshot ID in the goal object for convenience
	result.Goal.SnapshotID = strconv.Itoa(result.Snapshot.ID)
	return &result.Goal, nil
}

// CreateJourneyWithContext creates a new journey with context support
func (c *Client) CreateJourneyWithContext(ctx context.Context, goalID, snapshotID int, name string) (*Journey, error) {
	body := map[string]interface{}{
		"goalId":     goalID,
		"snapshotId": snapshotID,
		"name":       name,
		"title":      name, // Use same value for title
	}

	var result Journey
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader(constants.HeaderAuthorization, constants.AuthorizationHeaderPrefix+c.config.API.AuthToken).
		SetBody(body).
		SetResult(&result).
		Post("/journeys")

	if err != nil {
		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return nil, NewClientError("CreateJourneyWithContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewClientError("CreateJourneyWithContext", KindTimeout, "request timeout", err)
		}
		return nil, NewClientError("CreateJourneyWithContext", KindConnectionFailed, "failed to create journey", err)
	}

	if !resp.IsSuccess() {
		return nil, c.handleErrorResponse("CreateJourneyWithContext", resp)
	}

	return &result, nil
}

// CreateCheckpointWithContext creates a new checkpoint with context support
func (c *Client) CreateCheckpointWithContext(ctx context.Context, goalID, snapshotID int, title string) (*Checkpoint, error) {
	body := map[string]interface{}{
		"goalId":     goalID,
		"snapshotId": snapshotID,
		"title":      title,
	}

	var result Checkpoint
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader(constants.HeaderAuthorization, constants.AuthorizationHeaderPrefix+c.config.API.AuthToken).
		SetBody(body).
		SetResult(&result).
		Post("/checkpoints")

	if err != nil {
		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return nil, NewClientError("CreateCheckpointWithContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewClientError("CreateCheckpointWithContext", KindTimeout, "request timeout", err)
		}
		return nil, NewClientError("CreateCheckpointWithContext", KindConnectionFailed, "failed to create checkpoint", err)
	}

	if !resp.IsSuccess() {
		return nil, c.handleErrorResponse("CreateCheckpointWithContext", resp)
	}

	return &result, nil
}

// CreateClickStepWithContext creates a click step with context support
func (c *Client) CreateClickStepWithContext(ctx context.Context, checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "CLICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"xpath": "",
					"nlp":   element,
				},
			},
		},
		"position":     position,
		"optional":     false,
		"description":  fmt.Sprintf("Click on %s", element),
		"checkpointId": checkpointID,
	}

	body := map[string]interface{}{
		"checkpointId": checkpointID,
		"parsedStep":   parsedStep,
	}

	var result struct {
		ID int `json:"id"`
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader(constants.HeaderAuthorization, constants.AuthorizationHeaderPrefix+c.config.API.AuthToken).
		SetBody(body).
		SetResult(&result).
		Post("/steps")

	if err != nil {
		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return 0, NewClientError("CreateClickStepWithContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return 0, NewClientError("CreateClickStepWithContext", KindTimeout, "request timeout", err)
		}
		return 0, NewClientError("CreateClickStepWithContext", KindConnectionFailed, "failed to create click step", err)
	}

	if !resp.IsSuccess() {
		return 0, c.handleErrorResponse("CreateClickStepWithContext", resp)
	}

	return result.ID, nil
}

// ExecuteGoalWithContext executes a goal with context support
func (c *Client) ExecuteGoalWithContext(ctx context.Context, goalID, snapshotID int) (*Execution, error) {
	body := map[string]interface{}{
		"goalId":     goalID,
		"snapshotId": snapshotID,
	}

	var result Execution
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader(constants.HeaderAuthorization, constants.AuthorizationHeaderPrefix+c.config.API.AuthToken).
		SetBody(body).
		SetResult(&result).
		Post("/executions")

	if err != nil {
		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return nil, NewClientError("ExecuteGoalWithContext", KindContextCanceled, "request canceled", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewClientError("ExecuteGoalWithContext", KindTimeout, "request timeout", err)
		}
		return nil, NewClientError("ExecuteGoalWithContext", KindConnectionFailed, "failed to execute goal", err)
	}

	if !resp.IsSuccess() {
		return nil, c.handleErrorResponse("ExecuteGoalWithContext", resp)
	}

	return &result, nil
}

// Helper method to handle error responses
func (c *Client) handleErrorResponse(operation string, resp *resty.Response) error {
	// Try to parse API error from response
	var apiErr APIError
	if err := json.Unmarshal(resp.Body(), &apiErr); err == nil && apiErr.Message != "" {
		apiErr.Status = resp.StatusCode()
		return &apiErr
	}

	// Fallback to generic error
	message := string(resp.Body())
	if message == "" {
		message = fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	// Map common HTTP status codes to error codes
	code := ErrCodeInternalError
	switch resp.StatusCode() {
	case 400:
		code = ErrCodeBadRequest
	case 401:
		code = ErrCodeUnauthorized
	case 403:
		code = ErrCodeForbidden
	case 404:
		code = ErrCodeNotFound
	case 409:
		code = ErrCodeConflict
	case 429:
		code = ErrCodeRateLimited
	}

	return NewAPIError(resp.StatusCode(), code, message)
}
