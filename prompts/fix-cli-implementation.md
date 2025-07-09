# Fix Virtuoso CLI Implementation to Match Working API

## Context
The Virtuoso CLI has incorrect implementations for several step creation methods. We have working CURL examples that show the correct API format, but the Go code is using wrong action names and missing required meta fields.

## Task
Update the following methods in `/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/pkg/virtuoso/client.go` to match the working CURL examples:

### 1. Fix CreateHoverStep
**Current (WRONG):**
```go
func (c *Client) CreateHoverStep(checkpointID int, element string, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "HOVER", // ❌ Wrong
        "target": map[string]interface{}{
            "selectors": []map[string]interface{}{
                {
                    "type":  "GUESS",
                    "value": fmt.Sprintf(`{"clue":"%s"}`, element),
                },
            },
        },
        "value": "",
        "meta":  map[string]interface{}{}, // ❌ Missing required fields
    }
    return c.addStep(checkpointID, position, parsedStep)
}
```

**Should be (based on working CURL):**
```go
func (c *Client) CreateHoverStep(checkpointID int, element string, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "MOUSE", // ✅ Correct
        "value": "",
        "target": map[string]interface{}{
            "selectors": []map[string]interface{}{
                {
                    "type":  "GUESS",
                    "value": fmt.Sprintf(`{"clue":"%s"}`, element),
                },
            },
        },
        "meta": map[string]interface{}{
            "kind": "MOUSE",
            "action": "OVER", // ✅ Required meta fields
        },
    }
    return c.addStep(checkpointID, position, parsedStep)
}
```

### 2. Fix CreateDoubleClickStep
Change from `"action": "DOUBLE_CLICK"` to `"action": "MOUSE"` with `meta: {kind: "MOUSE", action: "DOUBLE_CLICK"}`

### 3. Fix CreateRightClickStep  
Change from `"action": "RIGHT_CLICK"` to `"action": "MOUSE"` with `meta: {kind: "MOUSE", action: "RIGHT_CLICK"}`

### 4. Fix AddFillStep (if it exists)
Change from `"action": "FILL"` to `"action": "WRITE"` with `meta: {kind: "WRITE", append: false}`

### 5. Add CreateAddCookieStep (create-step-add-cookie.go)
```go
func (c *Client) CreateAddCookieStep(checkpointID int, name, value string, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "ENVIRONMENT",
        "value": value,
        "meta": map[string]interface{}{
            "kind": "ENVIRONMENT",
            "name": name,
            "type": "ADD",
        },
    }
    return c.addStep(checkpointID, position, parsedStep)
}
```

### 6. Add CreateCommentStep (create-step-comment.go)
```go
func (c *Client) CreateCommentStep(checkpointID int, comment string, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "COMMENT",
        "value": comment,
        "meta": map[string]interface{}{
            "kind": "COMMENT",
        },
    }
    return c.addStep(checkpointID, position, parsedStep)
}
```

### 7. Add CreateDismissAlertStep (create-step-dismiss-alert.go)
```go
func (c *Client) CreateDismissAlertStep(checkpointID int, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "DISMISS",
        "value": "",
        "meta": map[string]interface{}{
            "kind": "DISMISS",
            "type": "ALERT",
        },
    }
    return c.addStep(checkpointID, position, parsedStep)
}
```

## Working CURL Reference

Here are the working CURL examples that the Go code should match:

### Hover (MOUSE with OVER)
```json
{
    "checkpointId": "1678295",
    "stepIndex": 7,
    "parsedStep": {
        "action": "MOUSE",
        "value": "",
        "target": {
            "selectors": [
                {
                    "type": "GUESS",
                    "value": "{\"clue\":\"Tooltip Trigger\"}"
                }
            ]
        },
        "meta": {
            "kind": "MOUSE",
            "action": "OVER"
        }
    }
}
```

### Add Cookie (ENVIRONMENT)
```json
{
    "checkpointId": "1678295",
    "stepIndex": 27,
    "parsedStep": {
        "action": "ENVIRONMENT",
        "value": "username",
        "meta": {
            "kind": "ENVIRONMENT",
            "name": "login",
            "type": "ADD"
        }
    }
}
```

### Comment
```json
{
    "checkpointId": "1678295",
    "stepIndex": 33,
    "parsedStep": {
        "action": "COMMENT",
        "value": "This is a test comment",
        "meta": {
            "kind": "COMMENT"
        }
    }
}
```

### Dismiss Alert
```json
{
    "checkpointId": "1678295",
    "stepIndex": 30,
    "parsedStep": {
        "action": "DISMISS",
        "value": "",
        "meta": {
            "kind": "DISMISS",
            "type": "ALERT"
        }
    }
}
```

## Files to Update
1. `/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/pkg/virtuoso/client.go` - Fix all the Create*Step methods
2. The corresponding command files in `src/cmd/` may need updates to pass correct parameters

## Testing
After making these changes, test each command:
```bash
./bin/api-cli create-step-hover 1678295 "Menu Item" 1
./bin/api-cli create-step-double-click 1678295 "Element" 2
./bin/api-cli create-step-right-click 1678295 "Context Menu" 3
./bin/api-cli create-step-add-cookie 1678295 "session" "abc123" 4
./bin/api-cli create-step-comment 1678295 "Test comment" 5
./bin/api-cli create-step-dismiss-alert 1678295 6
```

The responses should match the structure shown in the CURL examples.