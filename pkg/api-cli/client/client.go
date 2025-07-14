// Package virtuoso provides a client for interacting with the Virtuoso API.
// It includes methods for managing projects, goals, journeys, checkpoints, and test steps.
package client

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/marklovelady/api-cli-generator/pkg/api-cli/config"
	"github.com/sirupsen/logrus"
)

// Client wraps the Virtuoso API client
type Client struct {
	httpClient *resty.Client
	config     *config.VirtuosoConfig
	logger     *logrus.Logger
}

// NewClient creates a new Virtuoso API client
func NewClient(cfg *config.VirtuosoConfig) *Client {
	logger := logrus.New()
	if cfg.Output.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Create Resty client with all headers and config
	httpClient := resty.New().
		SetBaseURL(cfg.API.BaseURL).
		SetTimeout(time.Duration(cfg.HTTP.Timeout) * time.Second).
		SetRetryCount(cfg.HTTP.Retries).
		SetRetryWaitTime(time.Duration(cfg.HTTP.RetryWait) * time.Second).
		SetHeaders(cfg.GetHeaders()).
		OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
			logger.WithFields(logrus.Fields{
				"method": req.Method,
				"url":    req.URL,
			}).Debug("Making Virtuoso API request")
			return nil
		}).
		OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
			logger.WithFields(logrus.Fields{
				"status":   resp.StatusCode(),
				"duration": resp.Time(),
			}).Debug("Received Virtuoso API response")
			return nil
		})

	return &Client{
		httpClient: httpClient,
		config:     cfg,
		logger:     logger,
	}
}

// NewClientDirect creates a new Virtuoso API client with direct parameters (Version B compatibility)
func NewClientDirect(baseURL, token string) *Client {
	// Create a minimal config for Version B compatibility
	cfg := &config.VirtuosoConfig{
		API: config.APIConfig{
			BaseURL:   baseURL,
			AuthToken: token,
		},
		HTTP: config.HTTPConfig{
			Timeout:   30,
			Retries:   3,
			RetryWait: 1,
		},
		Output: config.OutputConfig{
			Verbose:       false,
			DefaultFormat: "human",
		},
	}

	// Set up headers
	headers := map[string]string{
		"Authorization":          "Bearer " + token,
		"Content-Type":           "application/json",
		"X-Virtuoso-Client-ID":   "api-cli-generator",
		"X-Virtuoso-Client-Name": "API CLI Generator",
	}

	logger := logrus.New()

	// Create Resty client
	httpClient := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetHeaders(headers).
		OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
			logger.WithFields(logrus.Fields{
				"method": req.Method,
				"url":    req.URL,
			}).Debug("Making Virtuoso API request")
			return nil
		}).
		OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
			logger.WithFields(logrus.Fields{
				"status":   resp.StatusCode(),
				"duration": resp.Time(),
			}).Debug("Received Virtuoso API response")
			return nil
		})

	return &Client{
		httpClient: httpClient,
		config:     cfg,
		logger:     logger,
	}
}

// APIResponse represents the standard Virtuoso API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Item    interface{} `json:"item,omitempty"`
	Items   interface{} `json:"items,omitempty"`
	Map     interface{} `json:"map,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Project represents a Virtuoso project
type Project struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	OrganizationID int       `json:"organizationId"`
	CreatedAt      time.Time `json:"createdAt,omitempty"`
}

// Goal represents a Virtuoso goal
type Goal struct {
	ID          int    `json:"id"`
	ProjectID   int    `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SnapshotID  string `json:"snapshotId,omitempty"`
	URL         string `json:"url,omitempty"`
}

// Journey represents a Virtuoso journey (testsuite)
type Journey struct {
	ID          int      `json:"id"`
	GoalID      int      `json:"goalId"`
	SnapshotID  int      `json:"snapshotId"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	CanonicalID string   `json:"canonicalId"`
	Archived    bool     `json:"archived"`
	Draft       bool     `json:"draft"`
	Tags        []string `json:"tags"`
}

// Checkpoint represents a Virtuoso checkpoint (testcase)
type Checkpoint struct {
	ID         int    `json:"id"`
	GoalID     int    `json:"goalId"`
	SnapshotID int    `json:"snapshotId"`
	Title      string `json:"title"`
}

// CheckpointDetail represents a checkpoint with additional details
type CheckpointDetail struct {
	ID         int                      `json:"id"`
	GoalID     int                      `json:"goalId"`
	SnapshotID int                      `json:"snapshotId"`
	Title      string                   `json:"title"`
	Position   int                      `json:"position"`
	Steps      []map[string]interface{} `json:"steps,omitempty"`
}

// JourneyWithCheckpoints represents a journey with its checkpoints
type JourneyWithCheckpoints struct {
	ID         int                `json:"id"`
	GoalID     int                `json:"goalId"`
	SnapshotID int                `json:"snapshotId"`
	Name       string             `json:"name"`
	Title      string             `json:"title"`
	Archived   bool               `json:"archived"`
	Draft      bool               `json:"draft"`
	Cases      []CheckpointDetail `json:"cases"`
}

// CreateProject creates a new project
func (c *Client) CreateProject(name, description string) (*Project, error) {
	// Convert org ID to int
	orgID, err := strconv.Atoi(c.config.Org.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	body := map[string]interface{}{
		"name":           name,
		"organizationId": orgID,
	}

	// Only add description if provided
	if description != "" {
		body["description"] = description
	}

	var response struct {
		Success bool    `json:"success"`
		Item    Project `json:"item"`
		Error   string  `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post("/projects")

	if err != nil {
		return nil, fmt.Errorf("create project request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("create project failed: %s", response.Error)
		}
		return nil, fmt.Errorf("create project failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("create project failed: API returned success=false")
	}

	return &response.Item, nil
}

// CreateGoal creates a new goal (with auto-created journey)
func (c *Client) CreateGoal(projectID int, name, url string) (*Goal, error) {
	body := map[string]interface{}{
		"projectId":     projectID,
		"name":          name,
		"environmentId": nil,
		"url":           url,
		"deviceSize": map[string]int{
			"width":  1280,
			"height": 800,
		},
		"meta": map[string]interface{}{
			"onlyManagersCanPublishJourneys": nil,
			"disableSameOrigin":              false,
			"disableEventLogs":               false,
			"bridgeClientUuid":               nil,
			"defaultMaxParallelism":          nil,
			"maxParallelism":                 nil,
			"legacyElementSelection":         nil,
			"clientSideLibraryShimsEnabled":  false,
			"popupAutoDismiss":               true,
			"waitForNetworkTraffic":          nil,
			"failScriptOnNavigate":           nil,
		},
		"createFirstJourney": true,
	}

	var response struct {
		Success bool   `json:"success"`
		Item    Goal   `json:"item"`
		Error   string `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post("/goals")

	if err != nil {
		return nil, fmt.Errorf("create goal request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("create goal failed: %s", response.Error)
		}
		return nil, fmt.Errorf("create goal failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("create goal failed: API returned success=false")
	}

	return &response.Item, nil
}

// GetGoalSnapshot gets the snapshot ID for a goal
func (c *Client) GetGoalSnapshot(goalID int) (string, error) {
	var response struct {
		Success bool `json:"success"`
		Item    struct {
			Snapshots []struct {
				SnapshotID int `json:"snapshotId"`
			} `json:"snapshots"`
		} `json:"item"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("/goals/%d/versions", goalID))

	if err != nil {
		return "", fmt.Errorf("get snapshot request failed: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("get snapshot failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return "", fmt.Errorf("get snapshot failed: API returned success=false")
	}

	if len(response.Item.Snapshots) == 0 {
		return "", fmt.Errorf("no snapshots found for goal %d", goalID)
	}

	// Return the first snapshot ID as string
	return fmt.Sprintf("%d", response.Item.Snapshots[0].SnapshotID), nil
}

// CreateJourney creates a new journey (testsuite)
func (c *Client) CreateJourney(goalID, snapshotID int, name string) (*Journey, error) {
	body := map[string]interface{}{
		"goalId":     goalID,
		"snapshotId": snapshotID,
		"name":       name,
		"title":      name, // Use same value for title
		"archived":   false,
		"draft":      true,
	}

	var response struct {
		Success bool    `json:"success"`
		Item    Journey `json:"item"`
		Error   string  `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post("/testsuites")

	if err != nil {
		return nil, fmt.Errorf("create journey request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("create journey failed: %s", response.Error)
		}
		return nil, fmt.Errorf("create journey failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Check if we have the journey in the wrapped response
	if response.Item.ID != 0 {
		return &response.Item, nil
	}

	// Try parsing as direct journey (in case API doesn't wrap)
	var journey Journey
	if err := json.Unmarshal(resp.Body(), &journey); err == nil && journey.ID != 0 {
		return &journey, nil
	}

	return nil, fmt.Errorf("could not parse journey response")
}

// CreateCheckpoint creates a new checkpoint (testcase)
func (c *Client) CreateCheckpoint(goalID, snapshotID int, title string) (*Checkpoint, error) {
	body := map[string]interface{}{
		"goalId":     goalID,
		"snapshotId": snapshotID,
		"title":      title,
	}

	var response struct {
		Success bool       `json:"success"`
		Item    Checkpoint `json:"item"`
		Error   string     `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post("/testcases")

	if err != nil {
		return nil, fmt.Errorf("create checkpoint request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("create checkpoint failed: %s", response.Error)
		}
		return nil, fmt.Errorf("create checkpoint failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Try direct response first (in case envelope=false is implied)
	if response.Item.ID != 0 {
		return &response.Item, nil
	}

	// Try parsing as direct checkpoint
	var checkpoint Checkpoint
	if err := json.Unmarshal(resp.Body(), &checkpoint); err == nil && checkpoint.ID != 0 {
		return &checkpoint, nil
	}

	return nil, fmt.Errorf("could not parse checkpoint response")
}

// AttachCheckpoint attaches a checkpoint to a journey
func (c *Client) AttachCheckpoint(journeyID, checkpointID, position int) error {
	body := map[string]interface{}{
		"checkpointId": checkpointID,
		"position":     position,
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		Post(fmt.Sprintf("/testsuites/%d/checkpoints/attach", journeyID))

	if err != nil {
		return fmt.Errorf("attach checkpoint request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("attach checkpoint failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// Step represents a test step
type Step struct {
	ID            int                    `json:"id"`
	CanonicalID   string                 `json:"canonicalId"`
	CheckpointID  int                    `json:"checkpointId"`
	StepIndex     int                    `json:"stepIndex"`
	Action        string                 `json:"action"`
	Value         string                 `json:"value"`
	Optional      bool                   `json:"optional"`
	IgnoreOutcome bool                   `json:"ignoreOutcome"`
	Skip          bool                   `json:"skip"`
	Meta          map[string]interface{} `json:"meta"`
	Target        map[string]interface{} `json:"target"`
}

// addStep is a helper method to add a step to a checkpoint
func (c *Client) addStep(checkpointID int, stepIndex int, parsedStep map[string]interface{}) (int, error) {
	body := map[string]interface{}{
		"checkpointId": checkpointID,
		"stepIndex":    stepIndex,
		"parsedStep":   parsedStep,
	}

	var response interface{}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post("/teststeps")

	if err != nil {
		return 0, fmt.Errorf("add step request failed: %w", err)
	}

	if resp.IsError() {
		return 0, fmt.Errorf("add step failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// The API returns 200 with the created step details
	// Try to extract the ID from the response
	if responseMap, ok := response.(map[string]interface{}); ok {
		if id, ok := responseMap["id"].(float64); ok {
			return int(id), nil
		}
	}

	// If we can't find an ID, return a success indicator
	// The API might not return an ID for steps
	return 1, nil
}

// AddNavigateStep adds a navigation step to a checkpoint
func (c *Client) AddNavigateStep(checkpointID int, url string) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "NAVIGATE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, url),
				},
			},
		},
		"value": url,
		"meta":  map[string]interface{}{},
	}

	// For now, we'll append at the end by using a high stepIndex
	// In a real implementation, we'd query existing steps first
	return c.addStep(checkpointID, 999, parsedStep)
}

// AddClickStep adds a click step to a checkpoint
func (c *Client) AddClickStep(checkpointID int, selector string) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "CLICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, selector),
				},
			},
		},
		"value": "",
		"meta":  map[string]interface{}{},
	}

	// For now, we'll append at the end by using a high stepIndex
	// In a real implementation, we'd query existing steps first
	return c.addStep(checkpointID, 999, parsedStep)
}

// AddWaitStep adds a wait step to a checkpoint
func (c *Client) AddWaitStep(checkpointID int, selector string, timeout int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, selector),
				},
			},
		},
		"value": fmt.Sprintf("%d", timeout),
		"meta":  map[string]interface{}{},
	}

	// For now, we'll append at the end by using a high stepIndex
	// In a real implementation, we'd query existing steps first
	return c.addStep(checkpointID, 999, parsedStep)
}

// TestConnection tests the API connection and authentication
func (c *Client) TestConnection() (string, error) {
	// Try to list projects with organizationId to test auth
	start := time.Now()

	resp, err := c.httpClient.R().
		SetQueryParam("organizationId", c.config.Org.ID).
		Get("/projects")

	duration := time.Since(start)

	if err != nil {
		return "", fmt.Errorf("connection test failed: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 401 {
			return "", fmt.Errorf("authentication failed: invalid API token")
		}
		return "", fmt.Errorf("API returned error status %d", resp.StatusCode())
	}

	return duration.String(), nil
}

// ListProjects lists all projects for the organization
func (c *Client) ListProjects() ([]*Project, error) {
	return c.ListProjectsWithOptions(0, 100)
}

// ListProjectsWithOptions lists projects with pagination options
func (c *Client) ListProjectsWithOptions(offset, limit int) ([]*Project, error) {
	var response struct {
		Success bool               `json:"success"`
		Map     map[string]Project `json:"map"`
		Total   int                `json:"total"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	req := c.httpClient.R().
		SetQueryParam("organizationId", c.config.Org.ID).
		SetResult(&response)

	// Add pagination parameters if specified
	if limit > 0 {
		req.SetQueryParam("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		req.SetQueryParam("offset", strconv.Itoa(offset))
	}

	resp, err := req.Get("/projects")

	if err != nil {
		return nil, fmt.Errorf("list projects request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != nil {
			return nil, fmt.Errorf("list projects failed: %s", response.Error.Message)
		}
		return nil, fmt.Errorf("list projects failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Convert map to slice of pointers
	projects := make([]*Project, 0, len(response.Map))
	for _, project := range response.Map {
		p := project // Create a copy to avoid pointer issues
		projects = append(projects, &p)
	}

	// Sort by ID for consistent ordering
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ID < projects[j].ID
	})

	return projects, nil
}

// ListGoals lists all goals for a project
func (c *Client) ListGoals(projectID int) ([]*Goal, error) {
	var response struct {
		Success bool   `json:"success"`
		Items   []Goal `json:"items"`
		Error   string `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetQueryParam("archived", "false").
		SetResult(&response).
		Get(fmt.Sprintf("/projects/%d/goals", projectID))

	if err != nil {
		return nil, fmt.Errorf("list goals request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("list goals failed: %s", response.Error)
		}
		return nil, fmt.Errorf("list goals failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Convert to slice of pointers
	goals := make([]*Goal, len(response.Items))
	for i := range response.Items {
		goals[i] = &response.Items[i]
	}

	return goals, nil
}

// ListJourneys lists all journeys for a goal
func (c *Client) ListJourneys(goalID, snapshotID int) ([]*Journey, error) {
	// Define the response structure to match the actual API
	type JourneyEntry struct {
		Journey    Journey `json:"journey"`
		LastChange struct {
			Date string `json:"date"`
		} `json:"lastChange"`
	}

	var response struct {
		Success bool                    `json:"success"`
		Map     map[string]JourneyEntry `json:"map"`
		Error   struct {
			Code    string `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetQueryParam("snapshotId", fmt.Sprintf("%d", snapshotID)).
		SetQueryParam("goalId", fmt.Sprintf("%d", goalID)).
		SetQueryParam("includeSequencesDetails", "true").
		SetResult(&response).
		Get("/testsuites/latest_status")

	if err != nil {
		return nil, fmt.Errorf("list journeys request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error.Message != "" {
			return nil, fmt.Errorf("list journeys failed: %s", response.Error.Message)
		}
		return nil, fmt.Errorf("list journeys failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Convert map to slice of journey pointers
	journeys := make([]*Journey, 0, len(response.Map))
	for _, entry := range response.Map {
		journey := entry.Journey
		journeys = append(journeys, &journey)
	}

	// Sort by ID to ensure consistent ordering (auto-created journey first)
	sort.Slice(journeys, func(i, j int) bool {
		return journeys[i].ID < journeys[j].ID
	})

	return journeys, nil
}

// UpdateJourney updates an existing journey (testsuite)
func (c *Client) UpdateJourney(journeyID int, name string) (*Journey, error) {
	body := map[string]interface{}{
		"title": name,
	}

	var response struct {
		Success bool    `json:"success"`
		Item    Journey `json:"item"`
		Error   string  `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Put(fmt.Sprintf("/testsuites/%d", journeyID))

	if err != nil {
		return nil, fmt.Errorf("update journey request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("update journey failed: %s", response.Error)
		}
		return nil, fmt.Errorf("update journey failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Check if we have the journey in the wrapped response
	if response.Item.ID != 0 {
		return &response.Item, nil
	}

	// Try parsing as direct journey (in case API doesn't wrap)
	var journey Journey
	if err := json.Unmarshal(resp.Body(), &journey); err == nil && journey.ID != 0 {
		return &journey, nil
	}

	return nil, fmt.Errorf("could not parse journey response")
}

// GetJourney retrieves a single journey by ID
func (c *Client) GetJourney(journeyID int) (*Journey, error) {
	var response struct {
		Success bool    `json:"success"`
		Item    Journey `json:"item"`
		Error   string  `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("/testsuites/%d", journeyID))

	if err != nil {
		return nil, fmt.Errorf("get journey request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("get journey failed: %s", response.Error)
		}
		return nil, fmt.Errorf("get journey failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Check if we have the journey in the wrapped response
	if response.Item.ID != 0 {
		return &response.Item, nil
	}

	// Try parsing as direct journey (in case API doesn't wrap)
	var journey Journey
	if err := json.Unmarshal(resp.Body(), &journey); err == nil && journey.ID != 0 {
		return &journey, nil
	}

	return nil, fmt.Errorf("could not parse journey response")
}

// GetStep retrieves a single test step by ID
func (c *Client) GetStep(stepID int) (*Step, error) {
	var response struct {
		Success bool   `json:"success"`
		Item    Step   `json:"item"`
		Error   string `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("/teststeps/%d", stepID))

	if err != nil {
		return nil, fmt.Errorf("get step request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("get step failed: %s", response.Error)
		}
		return nil, fmt.Errorf("get step failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Check if we have the step in the wrapped response
	if response.Item.ID != 0 {
		return &response.Item, nil
	}

	// Try parsing as direct step (in case API doesn't wrap)
	var step Step
	if err := json.Unmarshal(resp.Body(), &step); err == nil && step.ID != 0 {
		return &step, nil
	}

	return nil, fmt.Errorf("could not parse step response")
}

// UpdateNavigationStep updates a navigation step with a new URL
func (c *Client) UpdateNavigationStep(stepID int, canonicalID, url string, useNewTab bool) (*Step, error) {
	// Validate URL
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	body := map[string]interface{}{
		"id":          stepID,
		"canonicalId": canonicalID,
		"action":      "NAVIGATE",
		"value":       url,
		"meta": map[string]interface{}{
			"kind":      "NAVIGATE",
			"type":      "URL",
			"value":     url,
			"url":       url,
			"useNewTab": useNewTab,
		},
		"optional":      false,
		"ignoreOutcome": false,
		"skip":          false,
	}

	var response struct {
		Success bool   `json:"success"`
		Item    Step   `json:"item"`
		Error   string `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Put(fmt.Sprintf("/teststeps/%d/properties", stepID))

	if err != nil {
		return nil, fmt.Errorf("update navigation step request failed: %w", err)
	}

	if resp.IsError() {
		// Check for canonical ID mismatch
		if resp.StatusCode() == 400 || resp.StatusCode() == 409 {
			return nil, fmt.Errorf("update failed - canonical ID mismatch or invalid request")
		}
		if response.Error != "" {
			return nil, fmt.Errorf("update navigation step failed: %s", response.Error)
		}
		return nil, fmt.Errorf("update navigation step failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Check if we have the step in the wrapped response
	if response.Item.ID != 0 {
		return &response.Item, nil
	}

	// Try parsing as direct step (in case API doesn't wrap)
	var step Step
	if err := json.Unmarshal(resp.Body(), &step); err == nil && step.ID != 0 {
		return &step, nil
	}

	return nil, fmt.Errorf("could not parse step response")
}

// ListCheckpoints retrieves all checkpoints for a journey
func (c *Client) ListCheckpoints(journeyID int) (*JourneyWithCheckpoints, error) {
	var response struct {
		Success bool                   `json:"success"`
		Item    JourneyWithCheckpoints `json:"item"`
		Error   string                 `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("/testsuites/%d", journeyID))

	if err != nil {
		return nil, fmt.Errorf("list checkpoints request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("list checkpoints failed: %s", response.Error)
		}
		return nil, fmt.Errorf("list checkpoints failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	// Check if we have the journey in the wrapped response
	if response.Item.ID != 0 {
		// Process checkpoints to add position numbers
		for i := range response.Item.Cases {
			response.Item.Cases[i].Position = i + 1
		}
		return &response.Item, nil
	}

	// Try parsing as direct journey (in case API doesn't wrap)
	var journey JourneyWithCheckpoints
	if err := json.Unmarshal(resp.Body(), &journey); err == nil && journey.ID != 0 {
		// Process checkpoints to add position numbers
		for i := range journey.Cases {
			journey.Cases[i].Position = i + 1
		}
		return &journey, nil
	}

	return nil, fmt.Errorf("could not parse journey response")
}

// GetFirstJourney retrieves the first (auto-created) journey for a goal
func (c *Client) GetFirstJourney(goalID, snapshotID int) (*Journey, error) {
	journeys, err := c.ListJourneys(goalID, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to list journeys: %w", err)
	}

	if len(journeys) == 0 {
		return nil, fmt.Errorf("no journeys found for goal %d", goalID)
	}

	// Return the first journey (usually the auto-created one)
	return journeys[0], nil
}

// GetFirstCheckpoint retrieves the first checkpoint of a journey (contains navigation)
func (c *Client) GetFirstCheckpoint(journeyID int) (*CheckpointDetail, error) {
	journey, err := c.ListCheckpoints(journeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}

	if len(journey.Cases) == 0 {
		return nil, fmt.Errorf("no checkpoints found for journey %d", journeyID)
	}

	// Return the first checkpoint
	return &journey.Cases[0], nil
}

// GetNavigationStep retrieves the navigation step from the first checkpoint
func (c *Client) GetNavigationStep(checkpointID int) (*Step, error) {
	// In Virtuoso, the navigation step is typically the first step
	// We need to query the checkpoint's steps
	// For now, we'll use a placeholder - in production you'd implement a proper endpoint
	return nil, fmt.Errorf("GetNavigationStep not yet implemented - use checkpoint details from ListCheckpoints")
}

// AddFillStep adds a fill (type) step to a checkpoint
func (c *Client) AddFillStep(checkpointID int, selector, value string) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WRITE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, selector),
				},
			},
		},
		"value": value,
		"meta": map[string]interface{}{
			"kind":   "WRITE",
			"append": false,
		},
	}

	return c.addStep(checkpointID, 999, parsedStep)
}

// CreateNavigationStep creates a navigation step at a specific position
func (c *Client) CreateNavigationStep(checkpointID int, url string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "NAVIGATE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, url),
				},
			},
		},
		"value": url,
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateWaitTimeStep creates a wait time step at a specific position
func (c *Client) CreateWaitTimeStep(checkpointID int, seconds int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"value":  fmt.Sprintf("%d", seconds*1000), // Convert seconds to milliseconds
		"meta": map[string]interface{}{
			"kind":     "WAIT",
			"type":     "TIME",
			"duration": seconds * 1000,
			"poll":     100,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateWaitElementStep creates a wait for element step at a specific position
func (c *Client) CreateWaitElementStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": "20000", // Default 20 second timeout
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateWindowResizeStep creates a window resize step at a specific position
func (c *Client) CreateWindowResizeStep(checkpointID int, width int, height int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WINDOW",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "WINDOW",
			"type": "RESIZE",
			"dimension": map[string]interface{}{
				"width":  width,
				"height": height,
			},
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateClickStep creates a click step at a specific position
func (c *Client) CreateClickStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "CLICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": "",
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateDoubleClickStep creates a double-click step at a specific position
func (c *Client) CreateDoubleClickStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind":   "MOUSE",
			"action": "DOUBLE_CLICK",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateHoverStep creates a hover step at a specific position
func (c *Client) CreateHoverStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind":   "MOUSE",
			"action": "OVER",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateRightClickStep creates a right-click step at a specific position
func (c *Client) CreateRightClickStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind":   "MOUSE",
			"action": "RIGHT_CLICK",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateWriteStep creates a text input step at a specific position
func (c *Client) CreateWriteStep(checkpointID int, text string, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WRITE",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": text,
		"meta": map[string]interface{}{
			"kind":   "WRITE",
			"append": false,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateKeyStep creates a keyboard press step at a specific position
func (c *Client) CreateKeyStep(checkpointID int, key string, position int) (int, error) {
	// Note: Based on the CSV, key press seems to use the same format as wait time
	// This might need adjustment based on actual API requirements
	parsedStep := map[string]interface{}{
		"action": "KEY",
		"value":  key,
		"meta": map[string]interface{}{
			"kind": "KEY",
			"key":  key,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreatePickStep creates a dropdown selection step at a specific position
func (c *Client) CreatePickStep(checkpointID int, value string, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "PICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": value,
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateUploadStep creates a file upload step at a specific position
func (c *Client) CreateUploadStep(checkpointID int, filename string, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "UPLOAD",
		"value":  filename,
		"element": map[string]interface{}{
			"target": map[string]interface{}{
				"selectors": []map[string]interface{}{
					{
						"type":  "GUESS",
						"value": fmt.Sprintf(`{"clue":"%s"}`, element),
					},
				},
			},
		},
		"meta": map[string]interface{}{
			"kind": "UPLOAD",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateScrollTopStep creates a scroll to top step at a specific position
func (c *Client) CreateScrollTopStep(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "SCROLL",
			"type": "TOP",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateScrollBottomStep creates a scroll to bottom step at a specific position
func (c *Client) CreateScrollBottomStep(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "SCROLL",
			"type": "BOTTOM",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateScrollElementStep creates a scroll to element step at a specific position
func (c *Client) CreateScrollElementStep(checkpointID int, element string, position int) (int, error) {
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

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertExistsStep creates an assertion step that verifies an element exists
func (c *Client) CreateAssertExistsStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_EXISTS",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": fmt.Sprintf("see \"%s\"", element),
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertNotExistsStep creates an assertion step that verifies an element does not exist
func (c *Client) CreateAssertNotExistsStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_NOT_EXISTS",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": fmt.Sprintf("do not see \"%s\"", element),
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertEqualsStep creates an assertion step that verifies an element has a specific text value
func (c *Client) CreateAssertEqualsStep(checkpointID int, element string, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_EQUALS",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": fmt.Sprintf("expect %s to have text \"%s\"", element, value),
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertCheckedStep creates an assertion step that verifies a checkbox or radio button is checked
func (c *Client) CreateAssertCheckedStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_CHECKED",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": fmt.Sprintf("see %s is checked", element),
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateStoreStep creates a store step that saves an element value to a variable
func (c *Client) CreateStoreStep(checkpointID int, element string, variableName string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action":   "STORE",
		"variable": variableName,
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind": "STORE",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateExecuteJsStep creates an execute JavaScript step
func (c *Client) CreateExecuteJsStep(checkpointID int, javascript string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "EXECUTE",
		"value":  javascript,
		"meta": map[string]interface{}{
			"explicit": true,
			"script":   javascript,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAddCookieStep creates an add cookie step
func (c *Client) CreateAddCookieStep(checkpointID int, name string, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"value":  value,
		"meta": map[string]interface{}{
			"kind": "ENVIRONMENT",
			"name": name,
			"type": "ADD",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateDismissAlertStep creates a dismiss alert step
func (c *Client) CreateDismissAlertStep(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "DISMISS",
			"type": "ALERT",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateCommentStep creates a comment step
func (c *Client) CreateCommentStep(checkpointID int, comment string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "COMMENT",
		"value":  comment,
		"meta": map[string]interface{}{
			"kind": "COMMENT",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertLessThanOrEqualStep creates an assertion step that verifies a value is less than or equal
func (c *Client) CreateAssertLessThanOrEqualStep(checkpointID int, element string, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_LESS_THAN_OR_EQUAL",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": value,
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertLessThanStep creates an assertion step that verifies a value is less than
func (c *Client) CreateAssertLessThanStep(checkpointID int, element string, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_LESS_THAN",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": value,
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertSelectedStep creates an assertion step that verifies an option is selected
func (c *Client) CreateAssertSelectedStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_SELECTED",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": "",
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertVariableStep creates an assertion step that verifies a variable value
func (c *Client) CreateAssertVariableStep(checkpointID int, variableName string, expectedValue string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action":   "ASSERT_VARIABLE",
		"variable": variableName,
		"value":    expectedValue,
		"meta": map[string]interface{}{
			"kind": "ASSERT_VARIABLE",
			"type": "EQUALS",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateDismissConfirmStep creates a dismiss confirm dialog step
func (c *Client) CreateDismissConfirmStep(checkpointID int, accept bool, position int) (int, error) {
	action := "CANCEL"
	if accept {
		action = "OK"
	}
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"kind":   "DISMISS",
			"type":   "CONFIRM",
			"action": action,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateDismissPromptStep creates a dismiss prompt dialog step
func (c *Client) CreateDismissPromptStep(checkpointID int, text string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  text,
		"meta": map[string]interface{}{
			"kind":   "DISMISS",
			"type":   "PROMPT",
			"action": "OK",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateClearCookiesStep creates a clear all cookies step
func (c *Client) CreateClearCookiesStep(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "ENVIRONMENT",
			"type": "CLEAR",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateDeleteCookieStep creates a delete cookie step
func (c *Client) CreateDeleteCookieStep(checkpointID int, name string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"value":  "",
		"meta": map[string]interface{}{
			"kind": "ENVIRONMENT",
			"name": name,
			"type": "DELETE",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateMouseDownStep creates a mouse down step
func (c *Client) CreateMouseDownStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind":   "MOUSE",
			"action": "DOWN",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateMouseUpStep creates a mouse up step
func (c *Client) CreateMouseUpStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind":   "MOUSE",
			"action": "UP",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateMouseMoveStep creates a mouse move step
func (c *Client) CreateMouseMoveStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind":   "MOUSE",
			"action": "MOVE",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateMouseEnterStep creates a mouse enter step
func (c *Client) CreateMouseEnterStep(checkpointID int, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind":   "MOUSE",
			"action": "ENTER",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreatePickValueStep creates a pick by value step
func (c *Client) CreatePickValueStep(checkpointID int, value string, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "PICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": value,
		"meta": map[string]interface{}{
			"kind": "PICK",
			"type": "VALUE",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreatePickTextStep creates a pick by visible text step
func (c *Client) CreatePickTextStep(checkpointID int, text string, element string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "PICK",
		"value":  text,
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"meta": map[string]interface{}{
			"kind": "PICK",
			"type": "VISIBLE_TEXT",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateScrollPositionStep creates a scroll to position step
func (c *Client) CreateScrollPositionStep(checkpointID int, x int, y int, position int) (int, error) {
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

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateStoreValueStep creates a store value step (stores a literal value)
func (c *Client) CreateStoreValueStep(checkpointID int, value string, variableName string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action":   "STORE",
		"variable": variableName,
		"value":    value,
		"meta": map[string]interface{}{
			"kind": "STORE",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateMouseMoveToStep creates a mouse move to absolute coordinates step
func (c *Client) CreateMouseMoveToStep(checkpointID int, x int, y int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"meta": map[string]interface{}{
			"action": "MOVE",
			"x":      x,
			"y":      y,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateMouseMoveByStep creates a mouse move by relative offset step
func (c *Client) CreateMouseMoveByStep(checkpointID int, x int, y int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"value":  "",
		"meta": map[string]interface{}{
			"action": "OFFSET",
			"x":      x,
			"y":      y,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateSwitchIFrameStep creates a switch to iframe step by element selector
func (c *Client) CreateSwitchIFrameStep(checkpointID int, selector string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SWITCH",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, selector),
				},
			},
		},
		"meta": map[string]interface{}{
			"type": "FRAME_BY_ELEMENT",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateSwitchNextTabStep creates a switch to next tab step
func (c *Client) CreateSwitchNextTabStep(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SWITCH",
		"meta": map[string]interface{}{
			"type": "NEXT_TAB",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateSwitchParentFrameStep creates a switch to parent frame step
func (c *Client) CreateSwitchParentFrameStep(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SWITCH",
		"meta": map[string]interface{}{
			"type": "PARENT_FRAME",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateSwitchPrevTabStep creates a switch to previous tab step
func (c *Client) CreateSwitchPrevTabStep(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SWITCH",
		"meta": map[string]interface{}{
			"type": "PREV_TAB",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertNotEqualsStep creates an assertion step that verifies an element does not equal a value
func (c *Client) CreateAssertNotEqualsStep(checkpointID int, element string, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_NOT_EQUALS",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": value,
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertGreaterThanStep creates an assertion step that verifies an element is greater than a value
func (c *Client) CreateAssertGreaterThanStep(checkpointID int, element string, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_GREATER_THAN",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": value,
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertGreaterThanOrEqualStep creates an assertion step that verifies an element is greater than or equal to a value
func (c *Client) CreateAssertGreaterThanOrEqualStep(checkpointID int, element string, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_GREATER_THAN_OR_EQUAL",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": value,
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateAssertMatchesStep creates an assertion step that verifies an element matches a regex pattern
func (c *Client) CreateAssertMatchesStep(checkpointID int, element string, regexPattern string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ASSERT_MATCHES",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, element),
				},
			},
		},
		"value": regexPattern,
		"meta":  map[string]interface{}{},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// ValidateCheckpoint validates that a checkpoint exists and is accessible
func (c *Client) ValidateCheckpoint(checkpointID int) error {
	// Try to get the checkpoint details via the testcases endpoint
	var response struct {
		Success bool       `json:"success"`
		Item    Checkpoint `json:"item"`
		Error   string     `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("/testcases/%d", checkpointID))

	if err != nil {
		return fmt.Errorf("checkpoint validation request failed: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 404 {
			return fmt.Errorf("checkpoint %d not found", checkpointID)
		}
		if response.Error != "" {
			return fmt.Errorf("checkpoint validation failed: %s", response.Error)
		}
		return fmt.Errorf("checkpoint validation failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return fmt.Errorf("checkpoint validation failed: API returned success=false")
	}

	// Verify we got a valid checkpoint back
	if response.Item.ID != checkpointID {
		return fmt.Errorf("checkpoint validation failed: returned checkpoint ID %d does not match requested ID %d",
			response.Item.ID, checkpointID)
	}

	return nil
}

// Execution represents a goal execution
type Execution struct {
	ID         string                 `json:"id"`
	GoalID     int                    `json:"goalId"`
	SnapshotID int                    `json:"snapshotId"`
	Status     string                 `json:"status"`
	StartTime  time.Time              `json:"startTime"`
	EndTime    *time.Time             `json:"endTime,omitempty"`
	Duration   int                    `json:"duration,omitempty"`
	Progress   *ExecutionProgress     `json:"progress,omitempty"`
	ResultsURL string                 `json:"resultsUrl,omitempty"`
	ReportURL  string                 `json:"reportUrl,omitempty"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
}

// ExecutionProgress represents execution progress details
type ExecutionProgress struct {
	CompletedSteps    int     `json:"completedSteps"`
	TotalSteps        int     `json:"totalSteps"`
	CompletedJourneys int     `json:"completedJourneys"`
	TotalJourneys     int     `json:"totalJourneys"`
	CurrentJourney    string  `json:"currentJourney"`
	SuccessRate       float64 `json:"successRate"`
	FailedSteps       int     `json:"failedSteps"`
	PercentComplete   float64 `json:"percentComplete"`
}

// ExecutionAnalysis represents detailed execution analysis
type ExecutionAnalysis struct {
	ExecutionID string                 `json:"executionId"`
	Summary     *ExecutionSummary      `json:"summary"`
	Failures    []ExecutionFailure     `json:"failures"`
	Performance *ExecutionPerformance  `json:"performance,omitempty"`
	AIInsights  []string               `json:"aiInsights,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// ExecutionSummary represents execution summary statistics
type ExecutionSummary struct {
	TotalSteps     int     `json:"totalSteps"`
	PassedSteps    int     `json:"passedSteps"`
	FailedSteps    int     `json:"failedSteps"`
	SkippedSteps   int     `json:"skippedSteps"`
	SuccessRate    float64 `json:"successRate"`
	Duration       string  `json:"duration"`
	TotalJourneys  int     `json:"totalJourneys"`
	PassedJourneys int     `json:"passedJourneys"`
	FailedJourneys int     `json:"failedJourneys"`
}

// ExecutionFailure represents a failed step with details
type ExecutionFailure struct {
	StepID       int    `json:"stepId"`
	JourneyName  string `json:"journeyName"`
	CheckpointID int    `json:"checkpointId"`
	Action       string `json:"action"`
	Error        string `json:"error"`
	Screenshot   string `json:"screenshot,omitempty"`
	AISuggestion string `json:"aiSuggestion,omitempty"`
	Timestamp    string `json:"timestamp"`
}

// ExecutionPerformance represents performance metrics
type ExecutionPerformance struct {
	AverageStepTime  int `json:"averageStepTime"`
	SlowestStepTime  int `json:"slowestStepTime"`
	FastestStepTime  int `json:"fastestStepTime"`
	NetworkRequests  int `json:"networkRequests"`
	JavaScriptErrors int `json:"javascriptErrors"`
	PageLoadTime     int `json:"pageLoadTime"`
}

// ExecuteGoal executes a goal and returns execution details
func (c *Client) ExecuteGoal(goalID, snapshotID int) (*Execution, error) {
	body := map[string]interface{}{
		"goalId":     goalID,
		"snapshotId": snapshotID,
	}

	var response struct {
		Success bool      `json:"success"`
		Item    Execution `json:"item"`
		Error   string    `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post(fmt.Sprintf("/goals/%d/snapshots/%d/execute", goalID, snapshotID))

	if err != nil {
		return nil, fmt.Errorf("execute goal request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("execute goal failed: %s", response.Error)
		}
		return nil, fmt.Errorf("execute goal failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("execute goal failed: API returned success=false")
	}

	return &response.Item, nil
}

// GetExecutionStatus gets the current status of an execution
func (c *Client) GetExecutionStatus(executionID string) (*Execution, error) {
	var response struct {
		Success bool      `json:"success"`
		Item    Execution `json:"item"`
		Error   string    `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("/executions/%s", executionID))

	if err != nil {
		return nil, fmt.Errorf("get execution status request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("get execution status failed: %s", response.Error)
		}
		return nil, fmt.Errorf("get execution status failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("get execution status failed: API returned success=false")
	}

	return &response.Item, nil
}

// GetExecutionAnalysis gets detailed analysis of an execution
func (c *Client) GetExecutionAnalysis(executionID string, includeAI bool) (*ExecutionAnalysis, error) {
	var response struct {
		Success bool              `json:"success"`
		Item    ExecutionAnalysis `json:"item"`
		Error   string            `json:"error,omitempty"`
	}

	req := c.httpClient.R().SetResult(&response)

	if includeAI {
		req = req.SetQueryParam("includeAI", "true")
	}

	resp, err := req.Get(fmt.Sprintf("/executions/analysis/%s", executionID))

	if err != nil {
		return nil, fmt.Errorf("get execution analysis request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("get execution analysis failed: %s", response.Error)
		}
		return nil, fmt.Errorf("get execution analysis failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("get execution analysis failed: API returned success=false")
	}

	return &response.Item, nil
}

// TestData represents a test data table
type TestData struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Columns     []string               `json:"columns"`
	RowCount    int                    `json:"rowCount"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// TestDataRow represents a single row of test data
type TestDataRow struct {
	ID      string                 `json:"id"`
	TableID string                 `json:"tableId"`
	Data    map[string]interface{} `json:"data"`
}

// CreateTestDataTable creates a new test data table
func (c *Client) CreateTestDataTable(name, description string, columns []string) (*TestData, error) {
	body := map[string]interface{}{
		"name":        name,
		"description": description,
		"columns":     columns,
	}

	var response struct {
		Success bool     `json:"success"`
		Item    TestData `json:"item"`
		Error   string   `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post("/testdata/tables/create")

	if err != nil {
		return nil, fmt.Errorf("create test data table request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("create test data table failed: %s", response.Error)
		}
		return nil, fmt.Errorf("create test data table failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("create test data table failed: API returned success=false")
	}

	return &response.Item, nil
}

// GetTestDataTable gets details of a test data table
func (c *Client) GetTestDataTable(tableID string) (*TestData, error) {
	var response struct {
		Success bool     `json:"success"`
		Item    TestData `json:"item"`
		Error   string   `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("/testdata/tables/%s", tableID))

	if err != nil {
		return nil, fmt.Errorf("get test data table request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("get test data table failed: %s", response.Error)
		}
		return nil, fmt.Errorf("get test data table failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("get test data table failed: API returned success=false")
	}

	return &response.Item, nil
}

// ImportTestDataFromCSV imports test data from a CSV file
func (c *Client) ImportTestDataFromCSV(tableID string, csvData [][]string) error {
	body := map[string]interface{}{
		"tableId": tableID,
		"data":    csvData,
		"format":  "csv",
	}

	var response struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post(fmt.Sprintf("/testdata/tables/%s/import", tableID))

	if err != nil {
		return fmt.Errorf("import test data request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return fmt.Errorf("import test data failed: %s", response.Error)
		}
		return fmt.Errorf("import test data failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return fmt.Errorf("import test data failed: API returned success=false")
	}

	return nil
}

// Environment represents a test environment
type Environment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Variables   map[string]interface{} `json:"variables"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// CreateEnvironment creates a new test environment
func (c *Client) CreateEnvironment(name, description string, variables map[string]interface{}) (*Environment, error) {
	body := map[string]interface{}{
		"name":        name,
		"description": description,
		"variables":   variables,
	}

	var response struct {
		Success bool        `json:"success"`
		Item    Environment `json:"item"`
		Error   string      `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post("/environments")

	if err != nil {
		return nil, fmt.Errorf("create environment request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("create environment failed: %s", response.Error)
		}
		return nil, fmt.Errorf("create environment failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("create environment failed: API returned success=false")
	}

	return &response.Item, nil
}

// ============= VERSION B ENHANCED METHODS =============
// These methods provide enhanced functionality from Version B

// createStepWithCustomBody is a helper method for complex step creation (Version B enhancement)
func (c *Client) createStepWithCustomBody(checkpointID int, parsedStepBody map[string]interface{}, position int) (int, error) {
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
		SetBody(body).
		SetResult(&response).
		Post("/teststeps?envelope=false")

	if err != nil {
		return 0, fmt.Errorf("create step request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error != "" {
			return 0, fmt.Errorf("create step failed: %s", response.Error)
		}
		return 0, fmt.Errorf("create step failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return response.Item.ID, nil
}

// CreateStep creates a generic step using a structured request
func (c *Client) CreateStep(request interface{}) (interface{}, error) {
	// Convert the request to the format expected by createStepWithCustomBody
	var checkpointID int
	var position int
	var parsedStep map[string]interface{}

	// Handle different request types
	switch req := request.(type) {
	case map[string]interface{}:
		// Extract fields from map
		if id, ok := req["checkpointId"].(string); ok {
			checkpointID, _ = strconv.Atoi(id)
		}
		if pos, ok := req["position"].(int); ok {
			position = pos
		}

		// Build parsedStep from the request
		parsedStep = make(map[string]interface{})
		if stepType, ok := req["type"].(string); ok {
			parsedStep["action"] = stepType
		}
		if value, ok := req["value"].(string); ok {
			parsedStep["value"] = value
		}
		if selector, ok := req["selector"].(string); ok && selector != "" {
			clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)
			parsedStep["target"] = map[string]interface{}{
				"selectors": []map[string]interface{}{
					{
						"type":  "GUESS",
						"value": clueJSON,
					},
				},
			}
		}
		if meta, ok := req["meta"].(map[string]interface{}); ok {
			parsedStep["meta"] = meta
		} else {
			parsedStep["meta"] = map[string]interface{}{}
		}

	default:
		return nil, fmt.Errorf("unsupported request type: %T", request)
	}

	// Create the step
	stepID, err := c.createStepWithCustomBody(checkpointID, parsedStep, position)
	if err != nil {
		return nil, err
	}

	// Return a response similar to StepResult
	result := map[string]interface{}{
		"id":           stepID,
		"checkpointId": checkpointID,
		"type":         parsedStep["action"],
		"position":     position,
	}

	if selector, ok := parsedStep["target"].(map[string]interface{}); ok {
		result["selector"] = selector
	}

	if value, ok := parsedStep["value"].(string); ok {
		result["value"] = value
	}

	if meta, ok := parsedStep["meta"].(map[string]interface{}); ok {
		result["meta"] = meta
	}

	return result, nil
}

// CreateStepCookieCreate creates a cookie with the specified name and value (Version B)
func (c *Client) CreateStepCookieCreate(checkpointID int, name, value string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"value":  value,
		"meta": map[string]interface{}{
			"type": "ADD",
			"name": name,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateStepCookieWipeAll clears all cookies (Version B)
func (c *Client) CreateStepCookieWipeAll(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"meta": map[string]interface{}{
			"type": "CLEAR",
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateStepExecuteScript creates a step to execute a custom script (Version B)
func (c *Client) CreateStepExecuteScript(checkpointID int, scriptName string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "EXECUTE",
		"value":  scriptName,
		"meta": map[string]interface{}{
			"explicit": true,
			"script":   scriptName,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateStepUploadURL creates a step to upload a file from URL (Version B)
func (c *Client) CreateStepUploadURL(checkpointID int, url, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "UPLOAD",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": url,
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepDismissPromptWithText creates a step to dismiss a prompt with response text (Version B)
func (c *Client) CreateStepDismissPromptWithText(checkpointID int, text string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  text,
		"meta": map[string]interface{}{
			"type":   "PROMPT",
			"action": "OK",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepPickIndex creates a step to pick dropdown option by index (Version B)
func (c *Client) CreateStepPickIndex(checkpointID int, selector string, index int, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "PICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": fmt.Sprintf("%d", index),
		"meta": map[string]interface{}{
			"type": "INDEX",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepPickLast creates a step to pick last dropdown option (Version B)
func (c *Client) CreateStepPickLast(checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "PICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": "-1",
		"meta": map[string]interface{}{
			"type": "INDEX",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepWaitForElementTimeout creates a step to wait for element with custom timeout (Version B)
func (c *Client) CreateStepWaitForElementTimeout(checkpointID int, selector string, timeoutMs int, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": fmt.Sprintf("%d", timeoutMs),
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepWaitForElementDefault creates a step to wait for element with default timeout (Version B)
func (c *Client) CreateStepWaitForElementDefault(checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": "20000",
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepStoreElementText creates a step to store element text in variable (Version B)
func (c *Client) CreateStepStoreElementText(checkpointID int, selector, variableName string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepStoreLiteralValue creates a step to store literal value in variable (Version B)
func (c *Client) CreateStepStoreLiteralValue(checkpointID int, value, variableName string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action":   "STORE",
		"value":    value,
		"meta":     map[string]interface{}{},
		"variable": variableName,
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepMouseMoveTo creates a mouse move to absolute coordinates step (Version B)
func (c *Client) CreateStepMouseMoveTo(checkpointID int, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"meta": map[string]interface{}{
			"action": "MOVE",
			"x":      x,
			"y":      y,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateStepMouseMoveBy creates a mouse move by relative offset step (Version B)
func (c *Client) CreateStepMouseMoveBy(checkpointID int, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "MOUSE",
		"meta": map[string]interface{}{
			"action": "OFFSET",
			"x":      x,
			"y":      y,
		},
	}

	return c.addStep(checkpointID, position, parsedStep)
}

// CreateStepSwitchIframe creates a step to switch to iframe by element selector (Version B)
func (c *Client) CreateStepSwitchIframe(checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "SWITCH",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"meta": map[string]interface{}{
			"type": "FRAME_BY_ELEMENT",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepSwitchNextTab creates a step to switch to next tab (Version B)
func (c *Client) CreateStepSwitchNextTab(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SWITCH",
		"meta": map[string]interface{}{
			"type": "NEXT_TAB",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepSwitchParentFrame creates a step to switch to parent frame (Version B)
func (c *Client) CreateStepSwitchParentFrame(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SWITCH",
		"meta": map[string]interface{}{
			"type": "PARENT_FRAME",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepSwitchPrevTab creates a step to switch to previous tab (Version B)
func (c *Client) CreateStepSwitchPrevTab(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SWITCH",
		"meta": map[string]interface{}{
			"type": "PREV_TAB",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertNotEquals creates a step to assert element does not equal value (Version B)
func (c *Client) CreateStepAssertNotEquals(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_NOT_EQUALS",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertGreaterThan creates a step to assert element is greater than value (Version B)
func (c *Client) CreateStepAssertGreaterThan(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_GREATER_THAN",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertGreaterThanOrEqual creates a step to assert element is greater than or equal to value (Version B)
func (c *Client) CreateStepAssertGreaterThanOrEqual(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_GREATER_THAN_OR_EQUAL",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertMatches creates a step to assert element matches regex pattern (Version B)
func (c *Client) CreateStepAssertMatches(checkpointID int, selector, pattern string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_MATCHES",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": pattern,
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepNavigate creates a navigation step (Version B)
func (c *Client) CreateStepNavigate(checkpointID int, url string, useNewTab bool, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "NAVIGATE",
		"value":  url,
		"meta":   map[string]interface{}{},
	}

	if useNewTab {
		parsedStep["meta"].(map[string]interface{})["useNewTab"] = true
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepClick creates a click step (Version B)
func (c *Client) CreateStepClick(checkpointID int, selector string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepClickWithVariable creates a click step with variable target (Version B)
func (c *Client) CreateStepClickWithVariable(checkpointID int, variable string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepClickWithDetails creates a click step with position and element type (Version B)
func (c *Client) CreateStepClickWithDetails(checkpointID int, selector, positionType, elementType string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepWrite creates a write/input step (Version B)
func (c *Client) CreateStepWrite(checkpointID int, selector, value string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepWriteWithVariable creates a write step with variable storage (Version B)
func (c *Client) CreateStepWriteWithVariable(checkpointID int, selector, value, variable string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepScrollToPosition creates a scroll to position step (Version B)
func (c *Client) CreateStepScrollToPosition(checkpointID int, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"x":    x,
			"y":    y,
			"type": "POSITION",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepScrollByOffset creates a scroll by offset step (Version B)
func (c *Client) CreateStepScrollByOffset(checkpointID int, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"x":    x,
			"y":    y,
			"type": "OFFSET",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepScrollToTop creates a scroll to top step (Version B)
func (c *Client) CreateStepScrollToTop(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "TOP",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepWindowResize creates a window resize step (Version B)
func (c *Client) CreateStepWindowResize(checkpointID int, width, height, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WINDOW",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "RESIZE",
			"dimension": map[string]interface{}{
				"width":  width,
				"height": height,
			},
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepKeyGlobal creates a global key press step (Version B)
func (c *Client) CreateStepKeyGlobal(checkpointID int, key string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "KEY",
		"value":  key,
		"meta":   map[string]interface{}{},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepKeyTargeted creates a targeted key press step (Version B)
func (c *Client) CreateStepKeyTargeted(checkpointID int, selector, key string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepComment creates a comment step (Version B)
func (c *Client) CreateStepComment(checkpointID int, comment string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "COMMENT",
		"value":  comment,
		"meta":   map[string]interface{}{},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertLessThan creates a step to assert element is less than value
func (c *Client) CreateStepAssertLessThan(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_LESS_THAN",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertLessThanOrEqual creates a step to assert element is less than or equal to value
func (c *Client) CreateStepAssertLessThanOrEqual(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_LESS_THAN_OR_EQUAL",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertSelected creates a step to assert an option is selected in a dropdown
func (c *Client) CreateStepAssertSelected(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_SELECTED",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertVariable creates a step to assert a variable has a specific value
func (c *Client) CreateStepAssertVariable(checkpointID int, variableName, expectedValue string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action":   "ASSERT_VARIABLE",
		"variable": variableName,
		"value":    expectedValue,
		"meta": map[string]interface{}{
			"kind": "ASSERT_VARIABLE",
			"type": "EQUALS",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertEquals creates a step to assert element equals value
func (c *Client) CreateStepAssertEquals(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_EQUALS",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertExists creates a step to assert element exists
func (c *Client) CreateStepAssertExists(checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_EXISTS",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertNotExists creates a step to assert element does not exist
func (c *Client) CreateStepAssertNotExists(checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_NOT_EXISTS",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAssertChecked creates a step to assert checkbox/radio is checked
func (c *Client) CreateStepAssertChecked(checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "ASSERT_CHECKED",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepAddCookie creates a step to add a cookie
func (c *Client) CreateStepAddCookie(checkpointID int, name, value, domain string, position int) (int, error) {
	// Note: domain parameter is included for compatibility but may not be used by API
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"value":  value,
		"meta": map[string]interface{}{
			"type": "ADD",
			"name": name,
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepDeleteCookie creates a step to delete a specific cookie
func (c *Client) CreateStepDeleteCookie(checkpointID int, name string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"meta": map[string]interface{}{
			"type": "DELETE",
			"name": name,
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepClearCookies creates a step to clear all cookies
func (c *Client) CreateStepClearCookies(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "ENVIRONMENT",
		"meta": map[string]interface{}{
			"type": "CLEAR",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepDismissAlert creates a step to dismiss an alert dialog
func (c *Client) CreateStepDismissAlert(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "ALERT",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepDismissConfirm creates a step to dismiss a confirm dialog
func (c *Client) CreateStepDismissConfirm(checkpointID int, accept bool, position int) (int, error) {
	action := "CANCEL"
	if accept {
		action = "OK"
	}
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  "",
		"meta": map[string]interface{}{
			"type":   "CONFIRM",
			"action": action,
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepDismissPrompt creates a step to dismiss a prompt dialog
func (c *Client) CreateStepDismissPrompt(checkpointID int, text string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "DISMISS",
		"value":  text,
		"meta": map[string]interface{}{
			"type":   "PROMPT",
			"action": "OK",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepMouseDown creates a mouse down step
func (c *Client) CreateStepMouseDown(checkpointID int, selector string, position int) (int, error) {
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
			"action": "DOWN",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepMouseUp creates a mouse up step
func (c *Client) CreateStepMouseUp(checkpointID int, selector string, position int) (int, error) {
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
			"action": "UP",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepMouseEnter creates a mouse enter step
func (c *Client) CreateStepMouseEnter(checkpointID int, selector string, position int) (int, error) {
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
			"action": "ENTER",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepMouseMove creates a mouse move step
func (c *Client) CreateStepMouseMove(checkpointID int, selector string, position int) (int, error) {
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
			"action": "MOVE",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepPick creates a general pick dropdown option step
func (c *Client) CreateStepPick(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "PICK",
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepPickText creates a pick by text step
func (c *Client) CreateStepPickText(checkpointID int, selector, text string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "PICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": text,
		"meta": map[string]interface{}{
			"type": "TEXT",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepPickValue creates a pick by value step
func (c *Client) CreateStepPickValue(checkpointID int, selector, value string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "PICK",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": value,
		"meta": map[string]interface{}{
			"type": "VALUE",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepScrollBottom creates a scroll to bottom step
func (c *Client) CreateStepScrollBottom(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "BOTTOM",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepScrollElement creates a scroll to element step
func (c *Client) CreateStepScrollElement(checkpointID int, selector string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "SCROLL",
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
			"type": "ELEMENT",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepScrollPosition creates a scroll to position step
func (c *Client) CreateStepScrollPosition(checkpointID int, x, y, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "POSITION",
			"x":    x,
			"y":    y,
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepScrollTop creates a scroll to top step
func (c *Client) CreateStepScrollTop(checkpointID int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "SCROLL",
		"value":  "",
		"meta": map[string]interface{}{
			"type": "TOP",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepStore creates a general store step
func (c *Client) CreateStepStore(checkpointID int, selector, variableName string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepStoreValue creates a store value step
func (c *Client) CreateStepStoreValue(checkpointID int, selector, variableName string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepUpload creates a file upload step
func (c *Client) CreateStepUpload(checkpointID int, selector, filePath string, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "UPLOAD",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": filePath,
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepWaitElement creates a wait for element step
func (c *Client) CreateStepWaitElement(checkpointID int, selector string, timeout int, position int) (int, error) {
	clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": clueJSON,
				},
			},
		},
		"value": fmt.Sprintf("%d", timeout),
		"meta":  map[string]interface{}{},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepWaitTime creates a wait for time step
func (c *Client) CreateStepWaitTime(checkpointID int, milliseconds int, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WAIT",
		"value":  fmt.Sprintf("%d", milliseconds),
		"meta": map[string]interface{}{
			"type": "TIME",
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepWindow creates a window operation step
func (c *Client) CreateStepWindow(checkpointID int, operation string, params map[string]interface{}, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "WINDOW",
		"value":  "",
		"meta": map[string]interface{}{
			"type": strings.ToUpper(operation),
		},
	}

	// Handle resize operation specifically
	if strings.ToUpper(operation) == "RESIZE" {
		if size, ok := params["size"].(string); ok && size != "" {
			// Parse size in format "WIDTHxHEIGHT"
			parts := strings.Split(size, "x")
			if len(parts) == 2 {
				if width, err := strconv.Atoi(parts[0]); err == nil {
					if height, err := strconv.Atoi(parts[1]); err == nil {
						parsedStep["meta"].(map[string]interface{})["dimension"] = map[string]interface{}{
							"width":  width,
							"height": height,
						}
					}
				}
			}
		}
	} else {
		// Add any additional params to meta for non-resize operations
		for k, v := range params {
			if k != "size" { // Skip size since we handled it above
				parsedStep["meta"].(map[string]interface{})[k] = v
			}
		}
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepExecuteJs creates an execute JavaScript step
func (c *Client) CreateStepExecuteJs(checkpointID int, javascript string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action": "EXECUTE",
		"value":  javascript,
		"meta": map[string]interface{}{
			"explicit": true,
			"script":   javascript,
		},
	}

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepDoubleClick creates a double-click step
func (c *Client) CreateStepDoubleClick(checkpointID int, selector string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepHover creates a hover step
func (c *Client) CreateStepHover(checkpointID int, selector string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// CreateStepRightClick creates a right-click step
func (c *Client) CreateStepRightClick(checkpointID int, selector string, position int) (int, error) {
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

	return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// ============= LIBRARY CHECKPOINT METHODS =============
// These methods handle library checkpoint operations

// LibraryCheckpoint represents a library checkpoint
type LibraryCheckpoint struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Steps       []map[string]interface{} `json:"steps,omitempty"`
	CreatedAt   string                   `json:"createdAt,omitempty"`
	UpdatedAt   string                   `json:"updatedAt,omitempty"`
}

// AddCheckpointToLibrary converts a checkpoint to a library checkpoint
func (c *Client) AddCheckpointToLibrary(checkpointID int) (*LibraryCheckpoint, error) {
	var response struct {
		Success bool              `json:"success"`
		Item    LibraryCheckpoint `json:"item"`
		Error   struct {
			Code    string `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Post(fmt.Sprintf("/testcases/%d/add-to-library", checkpointID))

	if err != nil {
		return nil, fmt.Errorf("add checkpoint to library request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error.Message != "" {
			return nil, fmt.Errorf("add checkpoint to library failed: %s", response.Error.Message)
		}
		return nil, fmt.Errorf("add checkpoint to library failed with status %d: %s",
			resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("add checkpoint to library failed: API returned success=false")
	}

	return &response.Item, nil
}

// GetLibraryCheckpoint retrieves details of a library checkpoint
func (c *Client) GetLibraryCheckpoint(libraryCheckpointID int) (*LibraryCheckpoint, error) {
	var response struct {
		Success bool              `json:"success"`
		Item    LibraryCheckpoint `json:"item"`
		Error   struct {
			Code    string `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetResult(&response).
		Get(fmt.Sprintf("/library/checkpoints/%d", libraryCheckpointID))

	if err != nil {
		return nil, fmt.Errorf("get library checkpoint request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error.Message != "" {
			return nil, fmt.Errorf("get library checkpoint failed: %s", response.Error.Message)
		}
		return nil, fmt.Errorf("get library checkpoint failed with status %d: %s",
			resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("get library checkpoint failed: API returned success=false")
	}

	return &response.Item, nil
}

// AttachLibraryCheckpoint attaches a library checkpoint to a journey at a specific position
func (c *Client) AttachLibraryCheckpoint(journeyID, libraryCheckpointID, position int) (*Checkpoint, error) {
	body := map[string]interface{}{
		"libraryCheckpointId": libraryCheckpointID,
		"position":            position,
	}

	var response struct {
		Success bool       `json:"success"`
		Item    Checkpoint `json:"item"`
		Error   struct {
			Code    string `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"error,omitempty"`
	}

	resp, err := c.httpClient.R().
		SetBody(body).
		SetResult(&response).
		Post(fmt.Sprintf("/testsuites/%d/checkpoints/attach", journeyID))

	if err != nil {
		return nil, fmt.Errorf("attach library checkpoint request failed: %w", err)
	}

	if resp.IsError() {
		if response.Error.Message != "" {
			return nil, fmt.Errorf("attach library checkpoint failed: %s", response.Error.Message)
		}
		return nil, fmt.Errorf("attach library checkpoint failed with status %d: %s",
			resp.StatusCode(), resp.String())
	}

	if !response.Success {
		return nil, fmt.Errorf("attach library checkpoint failed: API returned success=false")
	}

	return &response.Item, nil
}
