package client

import (
	"context"
	"fmt"
)

// CreateAssertExistsStepWithContext creates an assertion step that verifies an element exists with context support
func (c *Client) CreateAssertExistsStepWithContext(ctx context.Context, checkpointID int, element string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertNotExistsStepWithContext creates an assertion step that verifies an element does not exist with context support
func (c *Client) CreateAssertNotExistsStepWithContext(ctx context.Context, checkpointID int, element string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertEqualsStepWithContext creates an assertion step that verifies an element has a specific text value with context support
func (c *Client) CreateAssertEqualsStepWithContext(ctx context.Context, checkpointID int, element string, value string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertNotEqualsStepWithContext creates an assertion step that verifies an element does not equal a value with context support
func (c *Client) CreateAssertNotEqualsStepWithContext(ctx context.Context, checkpointID int, element string, value string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertCheckedStepWithContext creates an assertion step that verifies a checkbox or radio button is checked with context support
func (c *Client) CreateAssertCheckedStepWithContext(ctx context.Context, checkpointID int, element string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertSelectedStepWithContext creates an assertion step that verifies an option is selected with context support
func (c *Client) CreateAssertSelectedStepWithContext(ctx context.Context, checkpointID int, element string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertGreaterThanStepWithContext creates an assertion step that verifies an element is greater than a value with context support
func (c *Client) CreateAssertGreaterThanStepWithContext(ctx context.Context, checkpointID int, element string, value string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertGreaterThanOrEqualStepWithContext creates an assertion step that verifies an element is greater than or equal to a value with context support
func (c *Client) CreateAssertGreaterThanOrEqualStepWithContext(ctx context.Context, checkpointID int, element string, value string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertLessThanStepWithContext creates an assertion step that verifies a value is less than with context support
func (c *Client) CreateAssertLessThanStepWithContext(ctx context.Context, checkpointID int, element string, value string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertLessThanOrEqualStepWithContext creates an assertion step that verifies a value is less than or equal with context support
func (c *Client) CreateAssertLessThanOrEqualStepWithContext(ctx context.Context, checkpointID int, element string, value string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertMatchesStepWithContext creates an assertion step that verifies an element matches a regex pattern with context support
func (c *Client) CreateAssertMatchesStepWithContext(ctx context.Context, checkpointID int, element string, regexPattern string, position int) (int, error) {
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
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}

// CreateAssertVariableStepWithContext creates an assertion step that verifies a variable value with context support
func (c *Client) CreateAssertVariableStepWithContext(ctx context.Context, checkpointID int, variableName string, expectedValue string, position int) (int, error) {
	parsedStep := map[string]interface{}{
		"action":   "ASSERT_VARIABLE",
		"variable": variableName,
		"value":    expectedValue,
		"meta": map[string]interface{}{
			"kind": "ASSERT_VARIABLE",
			"type": "EQUALS",
		},
	}
	return c.createStepWithCustomBodyContext(ctx, checkpointID, parsedStep, position)
}
