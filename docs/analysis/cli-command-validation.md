# CLI Command Validation Report

## ❌ INVALID CLI Commands

Based on the Postman collection analysis, these CLI commands are using **incorrect** API actions:

### Mouse Actions (Incorrect)
| CLI Command | Current Action | Should Be | Required Meta |
|------------|----------------|-----------|---------------|
| `create-step-hover` | `HOVER` | `MOUSE` | `meta: {kind: "MOUSE", action: "OVER"}` |
| `create-step-double-click` | `DOUBLE_CLICK` | `MOUSE` | `meta: {kind: "MOUSE", action: "DOUBLE_CLICK"}` |
| `create-step-right-click` | `RIGHT_CLICK` | `MOUSE` | `meta: {kind: "MOUSE", action: "RIGHT_CLICK"}` |

### Unsupported Commands
These commands create steps that don't exist in the Virtuoso API:
- `create-step-add-cookie` - No cookie management in API
- `create-step-comment` - No comment support in API
- `create-step-dismiss-alert` - No alert handling in API

## ✅ VALID CLI Commands

These commands appear to use the correct API actions:

### Navigation & Control
- `create-step-navigate` → `NAVIGATE` ✅
- `create-step-wait-time` → `WAIT` with time ✅
- `create-step-wait-element` → `WAIT` with element ✅
- `create-step-window` → `WINDOW` ✅

### Input & Forms
- `create-step-write` → `WRITE` ✅
- `create-step-key` → `KEY` ✅
- `create-step-pick` → `PICK` ✅
- `create-step-upload` → `UPLOAD` ✅

### Assertions
- `create-step-assert-exists` → `ASSERT_EXISTS` ✅
- `create-step-assert-not-exists` → `ASSERT_NOT_EXISTS` ✅
- `create-step-assert-equals` → `ASSERT_EQUALS` ✅
- `create-step-assert-checked` → `ASSERT_CHECKED` ✅

### Data Operations
- `create-step-store` → `STORE` ✅
- `create-step-execute-js` → `EXECUTE` ✅

### Basic Actions
- `create-step-click` → `CLICK` ✅

### Scrolling
- `create-step-scroll-top` → `SCROLL` with TOP ✅
- `create-step-scroll-bottom` → `SCROLL` with BOTTOM ✅
- `create-step-scroll-element` → `SCROLL` with ELEMENT ✅

## 🔧 Required Fixes

### 1. Fix Mouse Actions in client.go

```go
// WRONG - Current implementation
func (c *Client) CreateHoverStep(checkpointID int, element string, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "HOVER", // ❌ Wrong!
        // ...
    }
}

// CORRECT - Should be
func (c *Client) CreateHoverStep(checkpointID int, element string, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "MOUSE", // ✅ Correct
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
            "kind": "MOUSE",
            "action": "OVER", // ✅ This specifies hover
        },
    }
    return c.addStep(checkpointID, position, parsedStep)
}
```

### 2. Remove Unsupported Commands

These files should be deleted as they create invalid steps:
- `src/cmd/create-step-add-cookie.go`
- `src/cmd/create-step-comment.go`
- `src/cmd/create-step-dismiss-alert.go`

### 3. Update Documentation

The COMMANDS.md file needs to be updated to remove references to unsupported step types.

## Summary

- **3 commands** need their API action fixed (hover, double-click, right-click)
- **3 commands** should be removed (add-cookie, comment, dismiss-alert)
- **~20 commands** are working correctly
- The structure creation commands handle the mapping correctly in YAML but individual CLI commands don't