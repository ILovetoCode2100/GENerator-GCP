# Virtuoso API Step Validation Report

Based on the Postman collection analysis, here are the **correct** step types and formats:

## ✅ Valid Step Types from Postman Collection

### Navigation and Control
- **NAVIGATE** - Navigate to URL
- **WAIT** - Wait for time or element (uses `meta.type` to distinguish)
- **WINDOW** - Resize window

### Mouse Actions
- **CLICK** - Regular click
- **MOUSE** - Other mouse actions (uses `meta.action` for type):
  - DOUBLE_CLICK
  - OVER (hover)
  - RIGHT_CLICK
  - MOVE
  - ENTER
  - DOWN/UP

### Input and Forms
- **WRITE** - Enter text in field
- **KEY** - Press keyboard key
- **PICK** - Select from dropdown
- **UPLOAD** - Upload file

### Scroll Actions
- **SCROLL** - All scroll actions (uses `meta.type`):
  - TOP
  - BOTTOM
  - POSITION
  - ELEMENT

### Assertions
- **ASSERT_EXISTS**
- **ASSERT_NOT_EXISTS**
- **ASSERT_EQUALS**
- **ASSERT_VARIABLE**
- **ASSERT_LESS_THAN_OR_EQUAL**
- **ASSERT_SELECTED**
- **ASSERT_CHECKED**

### Data Operations
- **STORE** - Store variable from element or value
- **EXECUTE** - Execute JavaScript

## ❌ Invalid Step Types in My YAML

These step types I used are **NOT** in the Postman collection and need to be corrected:

### Incorrect Step Types:
- `fill` → Use `WRITE`
- `type` → Use `WRITE`
- `press` → Use `KEY`
- `double_click` → Use `MOUSE` with `meta.action: "DOUBLE_CLICK"`
- `hover` → Use `MOUSE` with `meta.action: "OVER"`
- `right_click` → Use `MOUSE` with `meta.action: "RIGHT_CLICK"`
- `drag_drop` → Not found (might need MOUSE DOWN/UP sequence)
- `select` → Use `PICK`
- `check`/`uncheck`/`choose` → Not found (might use CLICK)
- `wait_element` → Use `WAIT` with element
- `wait_time` → Use `WAIT` with `meta.type: "TIME"`
- `scroll_to_top` → Use `SCROLL` with `meta.type: "TOP"`
- `scroll_to_bottom` → Use `SCROLL` with `meta.type: "BOTTOM"`
- `scroll_to_element` → Use `SCROLL` with `meta.type: "ELEMENT"`
- `assert_contains`/`assert_not_contains` → Not found
- `assert_not_checked` → Not found
- `assert_enabled`/`assert_disabled` → Not found
- `assert_visible`/`assert_hidden` → Not found
- `execute_js`/`execute_script` → Use `EXECUTE`
- `store_value` → Use `STORE`
- `add_cookie`/`clear_cookies` → Not found
- `refresh`/`go_back`/`go_forward` → Not found
- `accept_alert`/`dismiss_alert`/`alert_text` → Not found
- `switch_to_frame`/`switch_to_default` → Not found
- `new_tab`/`close_tab`/`switch_tab` → Not found
- `screenshot` → Not found
- `comment` → Not found
- `assert_url`/`assert_title` → Not found
- `count_elements` → Not found

## 📝 Correct Step Format Examples

### WRITE (not fill/type)
```json
{
  "action": "WRITE",
  "value": "test@example.com",
  "target": {
    "selectors": [{
      "type": "GUESS",
      "value": "{\"clue\":\"Email Field\"}"
    }]
  },
  "meta": {
    "kind": "WRITE",
    "append": false
  }
}
```

### MOUSE for hover (not hover)
```json
{
  "action": "MOUSE",
  "value": "",
  "target": {
    "selectors": [{
      "type": "GUESS",
      "value": "{\"clue\":\"Menu Item\"}"
    }]
  },
  "meta": {
    "kind": "MOUSE",
    "action": "OVER"
  }
}
```

### WAIT for element (not wait_element)
```json
{
  "action": "WAIT",
  "value": "5000",
  "element": {
    "target": {
      "selectors": [{
        "type": "GUESS",
        "value": "{\"clue\":\"Loading Complete\"}"
      }]
    }
  },
  "meta": {
    "duration": 5000,
    "kind": "WAIT",
    "poll": 100,
    "type": "ELEMENT"
  }
}
```

### SCROLL to element (not scroll_to_element)
```json
{
  "action": "SCROLL",
  "target": {
    "selectors": [{
      "type": "GUESS",
      "value": "{\"clue\":\"Submit Button\"}"
    }]
  },
  "value": "",
  "meta": {
    "type": "ELEMENT"
  }
}
```

## 🔧 CLI Commands Need Updates

The CLI commands need to map to these correct action types:
- `create-step-write` ✅ (correct)
- `create-step-fill` → Should use WRITE action
- `create-step-hover` → Should use MOUSE action with OVER
- `create-step-wait-element` → Should use WAIT action with element
- etc.

## 📋 Next Steps

1. Update the CLI implementation to use correct action types
2. Fix the mapping in the Go code
3. Update the YAML parser to convert friendly names to API actions
4. Update documentation with correct step types