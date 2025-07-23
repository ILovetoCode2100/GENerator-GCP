package client

import (
	"context"
	"strconv"
)

// Window operation wrappers that call context-aware versions

// CreateStepResizeWindowV2 creates a step to resize the browser window (new version)
func (c *Client) CreateStepResizeWindowV2(checkpointID string, width, height int, position int) (*StepResponse, error) {
	return c.CreateStepResizeWindowWithContext(context.Background(), checkpointID, width, height, position)
}

// CreateStepMaximizeV2 creates a step to maximize the browser window (new version)
func (c *Client) CreateStepMaximizeV2(checkpointID string, position int) (*StepResponse, error) {
	return c.CreateStepMaximizeWithContext(context.Background(), checkpointID, position)
}

// CreateStepWindowSwitchV2 creates a step to switch windows (new version)
func (c *Client) CreateStepWindowSwitchV2(checkpointID string, action string, value interface{}, position int) (*StepResponse, error) {
	return c.CreateStepWindowSwitchWithContext(context.Background(), checkpointID, action, value, position)
}

// CreateStepSwitchTabV2 creates a step to switch browser tabs (new version)
func (c *Client) CreateStepSwitchTabV2(checkpointID string, tabID string, position int) (*StepResponse, error) {
	return c.CreateStepSwitchTabWithContext(context.Background(), checkpointID, tabID, position)
}

// CreateStepSwitchIframeV2 creates a step to switch to an iframe (new version)
func (c *Client) CreateStepSwitchIframeV2(checkpointID string, iframeSelector string, position int) (*StepResponse, error) {
	return c.CreateStepSwitchIframeWithContext(context.Background(), checkpointID, iframeSelector, position)
}

// CreateStepSetViewportV2 creates a step to set viewport size (new version)
func (c *Client) CreateStepSetViewportV2(checkpointID string, width, height int, position int) (*StepResponse, error) {
	return c.CreateStepSetViewportWithContext(context.Background(), checkpointID, width, height, position)
}

// Miscellaneous operation wrappers

// CreateStepEnvironmentV2 creates an environment configuration step (new version)
func (c *Client) CreateStepEnvironmentV2(checkpointID string, envConfig map[string]interface{}, position int) (*StepResponse, error) {
	return c.CreateStepEnvironmentWithContext(context.Background(), checkpointID, envConfig, position)
}

// CreateStepConfigureV2 creates a configuration step (new version)
func (c *Client) CreateStepConfigureV2(checkpointID string, config map[string]interface{}, position int) (*StepResponse, error) {
	return c.CreateStepConfigureWithContext(context.Background(), checkpointID, config, position)
}

// Session/checkpoint operation wrappers

// SetCheckpointV2 sets the current checkpoint (new version)
func (c *Client) SetCheckpointV2(checkpointID string) error {
	return c.SetCheckpointWithContext(context.Background(), checkpointID)
}

// AddCheckpointToLibraryV2 adds a checkpoint to the library (new version)
func (c *Client) AddCheckpointToLibraryV2(checkpointID, name, description string) (*LibraryCheckpointResponse, error) {
	return c.AddCheckpointToLibraryWithContext(context.Background(), checkpointID, name, description)
}

// GetLibraryCheckpointV2 retrieves a library checkpoint (new version)
func (c *Client) GetLibraryCheckpointV2(libraryCheckpointID string) (*LibraryCheckpointResponse, error) {
	return c.GetLibraryCheckpointWithContext(context.Background(), libraryCheckpointID)
}

// AttachLibraryToJourneyV2 attaches a library checkpoint to a journey (new version)
func (c *Client) AttachLibraryToJourneyV2(journeyID, libraryCheckpointID string, position int) (*AttachLibraryResponse, error) {
	return c.AttachLibraryToJourneyWithContext(context.Background(), journeyID, libraryCheckpointID, position)
}

// Backward compatibility wrappers for existing methods

// CreateStepSwitchParentFrame creates a step to switch to parent frame (backward compatibility)
func (c *Client) CreateStepSwitchParentFrame(checkpointID int, position int) (int, error) {
	_, err := c.CreateStepSwitchParentFrameWithContext(context.Background(), strconv.Itoa(checkpointID), position)
	if err != nil {
		return 0, err
	}
	// Return position as step ID for backward compatibility
	return position, nil
}

// CreateStepComment creates a comment step (backward compatibility)
func (c *Client) CreateStepComment(checkpointID int, comment string, position int) (int, error) {
	resp, err := c.CreateStepCommentWithContext(context.Background(), strconv.Itoa(checkpointID), comment, position)
	if err != nil {
		return 0, err
	}
	// Try to parse step ID from response
	if resp != nil && resp.ID != "" {
		stepID, _ := strconv.Atoi(resp.ID)
		if stepID > 0 {
			return stepID, nil
		}
	}
	// Fallback to position for backward compatibility
	return position, nil
}

// CreateStepExecuteScript creates a step to execute a custom script (backward compatibility)
func (c *Client) CreateStepExecuteScript(checkpointID int, scriptName string, position int) (int, error) {
	// For backward compatibility, scriptName is treated as the script content
	resp, err := c.CreateStepExecuteScriptWithContext(context.Background(), strconv.Itoa(checkpointID), scriptName, position)
	if err != nil {
		return 0, err
	}
	// Try to parse step ID from response
	if resp != nil && resp.ID != "" {
		stepID, _ := strconv.Atoi(resp.ID)
		if stepID > 0 {
			return stepID, nil
		}
	}
	// Fallback to position for backward compatibility
	return position, nil
}
