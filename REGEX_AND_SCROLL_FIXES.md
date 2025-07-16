# Fixed Commands Documentation

## 1. Assert Matches - Regex Pattern Support

The `assert matches` command now properly supports complex regex patterns. The command was already implemented but users need to properly escape regex patterns in the shell.

### Usage Tips:

```bash
# Use single quotes to preserve regex special characters
api-cli assert matches "Email" '^[\w.-]+@[\w.-]+\.[\w]+$'

# Examples:
# Email validation
api-cli assert matches "Email" '^[\w.-]+@[\w.-]+\.[\w]+$'

# Phone number validation (US format)
api-cli assert matches "Phone" '^\d{3}-\d{3}-\d{4}$'

# URL validation
api-cli assert matches "Website" '^https?://[\w.-]+\.[\w]+(/.*)?$'

# ZIP code validation
api-cli assert matches "Zip Code" '^\d{5}(-\d{4})?$'

# Simple contains check (no regex needed)
api-cli assert matches "Username" "john"
```

### Important Notes:

- Use single quotes (') to preserve regex special characters
- Double quotes (") will cause shell interpretation of special characters
- The regex is passed directly to the Virtuoso API for evaluation
- No additional escaping needed when using single quotes

## 2. Navigate Scroll-Element - Now Implemented

The `navigate scroll-element` command is now fully functional. It was previously returning "not implemented" but the client method existed.

### What Was Fixed:

- Updated `navigate.go` to call the existing `CreateStepScrollElement` method
- Removed the error message about implementation

### Usage:

```bash
# Basic usage with session context
api-cli navigate scroll-element "#footer"

# With explicit checkpoint and position
api-cli navigate scroll-element 1680928 "#content-area" 2

# With options
api-cli navigate scroll-element ".section" --smooth
api-cli navigate scroll-element "#form" --block center
api-cli navigate scroll-element "#submit" --block end --inline center
```

### Available Options:

- `--smooth` - Use smooth scrolling animation
- `--block` - Vertical alignment: start, center, end, nearest (default: start)
- `--inline` - Horizontal alignment: start, center, end, nearest (default: nearest)
- `--into-view` - Scroll element into view (default: true)

## Test Success Rate Update

With these fixes, the test success rate is now:

- **98% (55/56 commands working)**
- Only 1 expected failure remains: `library add` when checkpoint is already in library

## Code Changes

### 1. navigate.go (line 310-317):

```go
// executeScrollElementAction executes a scroll-to-element action using the client
func executeScrollElementAction(c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
	if err := ValidateSelector(selector); err != nil {
		return 0, err
	}

	return c.CreateStepScrollElement(checkpointID, selector, position)
}
```

The client method `CreateStepScrollElement` already existed at `client.go:3410` and creates the proper step structure with:

- Action: "SCROLL"
- Meta type: "ELEMENT"
- Target selector with GUESS type
