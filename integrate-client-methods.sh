#!/bin/bash

# Script to help integrate Version B client methods into Version A

echo "Starting client.go integration..."

# Create a temporary file with the methods to add
cat > /tmp/client-methods-to-add.txt << 'EOF'

// ============= VERSION B METHODS TO ADD =============
// Add these methods to Version A's client.go

// createStepWithCustomBody is a helper method for complex step creation
func (c *Client) createStepWithCustomBody(checkpointID int, parsedStepBody map[string]interface{}, position int) (*StepResponse, error) {
	url := fmt.Sprintf("%s/teststeps?envelope=false", c.baseURL)
	
	requestBody := map[string]interface{}{
		"checkpointId": checkpointID,
		"stepIndex":    position,
		"parsedStep":   parsedStepBody,
	}
	
	resp, err := c.client.R().
		SetHeader("Authorization", "Bearer "+c.token).
		SetHeader("X-Virtuoso-Client-ID", c.clientID).
		SetHeader("X-Virtuoso-Client-Name", c.clientName).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create step: %w", err)
	}
	
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode(), resp.String())
	}
	
	var response StepResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	response.CheckpointID = checkpointID
	return &response, nil
}

// CreateStepCookieCreate creates a cookie with the specified name and value
func (c *Client) CreateStepCookieCreate(checkpointID int, name, value string, position int) (*StepResponse, error) {
	request := StepRequest{
		Action: "ENVIRONMENT",
		Value:  value,
		Meta: map[string]interface{}{
			"type": "ADD",
			"name": name,
		},
	}
	return c.createStep(checkpointID, request, position)
}

// CreateStepCookieWipeAll clears all cookies
func (c *Client) CreateStepCookieWipeAll(checkpointID int, position int) (*StepResponse, error) {
	request := StepRequest{
		Action: "ENVIRONMENT",
		Meta: map[string]interface{}{
			"type": "CLEAR",
		},
	}
	return c.createStep(checkpointID, request, position)
}

// CreateStepExecuteScript creates a step to execute a custom script
func (c *Client) CreateStepExecuteScript(checkpointID int, scriptName string, position int) (*StepResponse, error) {
	request := StepRequest{
		Action: "EXECUTE",
		Value:  scriptName,
		Meta: map[string]interface{}{
			"explicit": true,
			"script":   scriptName,
		},
	}
	return c.createStep(checkpointID, request, position)
}

// CreateStepMouseMoveTo creates a mouse move to absolute coordinates step
func (c *Client) CreateStepMouseMoveTo(checkpointID int, x, y, position int) (*StepResponse, error) {
	request := StepRequest{
		Action: "MOUSE",
		Meta: map[string]interface{}{
			"action": "MOVE",
			"x":      x,
			"y":      y,
		},
	}
	return c.createStep(checkpointID, request, position)
}

// CreateStepMouseMoveBy creates a mouse move by relative offset step
func (c *Client) CreateStepMouseMoveBy(checkpointID int, x, y, position int) (*StepResponse, error) {
	request := StepRequest{
		Action: "MOUSE",
		Meta: map[string]interface{}{
			"action": "OFFSET",
			"x":      x,
			"y":      y,
		},
	}
	return c.createStep(checkpointID, request, position)
}

EOF

echo "Methods to add have been saved to /tmp/client-methods-to-add.txt"
echo ""
echo "Next steps:"
echo "1. Open /Users/marklovelady/_dev/virtuoso-api-cli-generator/pkg/virtuoso/client.go"
echo "2. Add the methods from /tmp/client-methods-to-add.txt"
echo "3. Or copy all methods from merge-helpers/client-version-b.go that don't exist in Version A"
echo ""
echo "Note: The full Version B client.go is available at:"
echo "  /Users/marklovelady/_dev/virtuoso-api-cli-generator/merge-helpers/client-version-b.go"