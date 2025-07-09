# Step Type Mapping Guide

This guide shows how to map user-friendly step names to the correct Virtuoso API actions.

## Navigation & Control

| User-Friendly | API Action | Meta Fields | Notes |
|--------------|------------|-------------|-------|
| navigate | NAVIGATE | `kind: "NAVIGATE", type: "URL"` | Navigate to URL |
| wait | WAIT | `kind: "WAIT", type: "ELEMENT"` | Wait for element |
| wait_time | WAIT | `kind: "WAIT", type: "TIME"` | Wait for seconds |
| wait_element | WAIT | `kind: "WAIT", type: "ELEMENT"` | Same as wait |
| window | WINDOW | `kind: "WINDOW", type: "RESIZE"` | Resize window |

## Mouse Actions

| User-Friendly | API Action | Meta Fields | Notes |
|--------------|------------|-------------|-------|
| click | CLICK | `kind: "CLICK", type: "NATURAL"` | Regular click |
| double_click | MOUSE | `kind: "MOUSE", action: "DOUBLE_CLICK"` | Double click |
| hover | MOUSE | `kind: "MOUSE", action: "OVER"` | Mouse over |
| right_click | MOUSE | `kind: "MOUSE", action: "RIGHT_CLICK"` | Right click |

## Input & Forms

| User-Friendly | API Action | Meta Fields | Notes |
|--------------|------------|-------------|-------|
| write | WRITE | `kind: "WRITE", append: false` | Enter text |
| fill | WRITE | `kind: "WRITE", append: false` | Same as write |
| type | WRITE | `kind: "WRITE", append: false` | Same as write |
| key | KEY | `kind: "KEYBOARD"` | Press key |
| press | KEY | `kind: "KEYBOARD"` | Same as key |
| pick | PICK | `kind: "PICK", type: "VISIBLE_TEXT"` | Select dropdown |
| select | PICK | `kind: "PICK", type: "VISIBLE_TEXT"` | Same as pick |
| upload | UPLOAD | `kind: "UPLOAD"` | Upload file |

## Scrolling

| User-Friendly | API Action | Meta Fields | Notes |
|--------------|------------|-------------|-------|
| scroll_to_top | SCROLL | `kind: "SCROLL", type: "TOP"` | Scroll to top |
| scroll_to_bottom | SCROLL | `kind: "SCROLL", type: "BOTTOM"` | Scroll to bottom |
| scroll_to_element | SCROLL | `type: "ELEMENT"` | Scroll to element |
| scroll | SCROLL | `kind: "SCROLL", type: "POSITION"` | Scroll to x,y |

## Assertions

| User-Friendly | API Action | Meta Fields | Notes |
|--------------|------------|-------------|-------|
| assert_exists | ASSERT_EXISTS | `kind: "ASSERT"` | Element exists |
| assert_not_exists | ASSERT_NOT_EXISTS | `kind: "ASSERT"` | Element not exists |
| assert_equals | ASSERT_EQUALS | `kind: "ASSERT"` | Value equals |
| assert_checked | ASSERT_CHECKED | `kind: "ASSERT"` | Checkbox checked |
| assert_selected | ASSERT_SELECTED | `kind: "ASSERT"` | Dropdown selected |
| assert_variable | ASSERT_VARIABLE | `kind: "ASSERT_VARIABLE", type: "EQUALS"` | Variable equals |

## Data Operations

| User-Friendly | API Action | Meta Fields | Notes |
|--------------|------------|-------------|-------|
| store | STORE | `kind: "STORE"` | Store from element |
| store_value | STORE | `kind: "STORE"` | Store literal value |
| execute_js | EXECUTE | `explicit: true` | Execute JavaScript |

## Not Supported

These step types are NOT available in the Virtuoso API:
- drag_drop
- check/uncheck/choose (use click instead)
- assert_contains/assert_not_contains
- assert_enabled/assert_disabled
- assert_visible/assert_hidden
- add_cookie/clear_cookies
- refresh/go_back/go_forward
- accept_alert/dismiss_alert
- switch_to_frame/switch_to_default
- new_tab/close_tab/switch_tab
- screenshot
- comment
- assert_url/assert_title
- count_elements

## Implementation Notes

1. The CLI should accept user-friendly names and convert them to API actions
2. The YAML parser should handle this mapping automatically
3. Error messages should suggest alternatives for unsupported steps
4. Documentation should show both friendly names and API actions

## Example Conversion

```yaml
# User writes:
steps:
  - type: hover
    selector: ".menu"
  - type: fill
    selector: "#email"
    value: "test@example.com"

# Converts to API:
{
  "parsedStep": {
    "action": "MOUSE",
    "target": { ... },
    "meta": {
      "kind": "MOUSE",
      "action": "OVER"
    }
  }
}
{
  "parsedStep": {
    "action": "WRITE",
    "value": "test@example.com",
    "target": { ... },
    "meta": {
      "kind": "WRITE",
      "append": false
    }
  }
}
```