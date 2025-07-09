package virtuoso

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
	
	"github.com/go-resty/resty/v2"
	"github.com/marklovelady/api-cli-generator/pkg/config"
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
	ID          int                `json:"id"`
	GoalID      int                `json:"goalId"`
	SnapshotID  int                `json:"snapshotId"`
	Name        string             `json:"name"`
	Title       string             `json:"title"`
	Archived    bool               `json:"archived"`
	Draft       bool               `json:"draft"`
	Cases       []CheckpointDetail `json:"cases"`
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
		"projectId":        projectID,
		"name":             name,
		"environmentId":    nil,
		"url":              url,
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
		Success bool `json:"success"`
		Item    Goal `json:"item"`
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
	ID             int                    `json:"id"`
	CanonicalID    string                 `json:"canonicalId"`
	CheckpointID   int                    `json:"checkpointId"`
	StepIndex      int                    `json:"stepIndex"`
	Action         string                 `json:"action"`
	Value          string                 `json:"value"`
	Optional       bool                   `json:"optional"`
	IgnoreOutcome  bool                   `json:"ignoreOutcome"`
	Skip           bool                   `json:"skip"`
	Meta           map[string]interface{} `json:"meta"`
	Target         map[string]interface{} `json:"target"`
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
	var response struct {
		Success bool      `json:"success"`
		Items   []Project `json:"items"`
		Error   string    `json:"error,omitempty"`
	}
	
	resp, err := c.httpClient.R().
		SetQueryParam("organizationId", c.config.Org.ID).
		SetResult(&response).
		Get("/projects")
	
	if err != nil {
		return nil, fmt.Errorf("list projects request failed: %w", err)
	}
	
	if resp.IsError() {
		if response.Error != "" {
			return nil, fmt.Errorf("list projects failed: %s", response.Error)
		}
		return nil, fmt.Errorf("list projects failed with status %d: %s", resp.StatusCode(), resp.String())
	}
	
	// Convert to slice of pointers
	projects := make([]*Project, len(response.Items))
	for i := range response.Items {
		projects[i] = &response.Items[i]
	}
	
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
		Success bool                     `json:"success"`
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
		"action": "FILL",
		"target": map[string]interface{}{
			"selectors": []map[string]interface{}{
				{
					"type":  "GUESS",
					"value": fmt.Sprintf(`{"clue":"%s"}`, selector),
				},
			},
		},
		"value": value,
		"meta":  map[string]interface{}{},
	}
	
	return c.addStep(checkpointID, 999, parsedStep)
}