# Scroll Action Step Commands

This document describes the three new scroll action step commands that have been added to the API CLI tool.

## Commands Overview

### 1. `create-step-scroll-top`
**Purpose**: Creates a scroll to top step that scrolls to the top of the page.

**Usage**: 
```bash
api-cli create-step-scroll-top CHECKPOINT_ID POSITION
```

**Parameters**:
- `CHECKPOINT_ID`: The ID of the checkpoint to add the step to
- `POSITION`: The position in the checkpoint where the step should be inserted

**Examples**:
```bash
# Create a scroll to top step at position 1
api-cli create-step-scroll-top 1678318 1

# Create with JSON output
api-cli create-step-scroll-top 1678318 2 -o json
```

**Generated Step**: `"scroll to top of page"`

### 2. `create-step-scroll-bottom`
**Purpose**: Creates a scroll to bottom step that scrolls to the bottom of the page.

**Usage**: 
```bash
api-cli create-step-scroll-bottom CHECKPOINT_ID POSITION
```

**Parameters**:
- `CHECKPOINT_ID`: The ID of the checkpoint to add the step to
- `POSITION`: The position in the checkpoint where the step should be inserted

**Examples**:
```bash
# Create a scroll to bottom step at position 1
api-cli create-step-scroll-bottom 1678318 1

# Create with YAML output
api-cli create-step-scroll-bottom 1678318 3 -o yaml
```

**Generated Step**: `"scroll to bottom of page"`

### 3. `create-step-scroll-element`
**Purpose**: Creates a scroll to element step that scrolls to a specific element on the page.

**Usage**: 
```bash
api-cli create-step-scroll-element CHECKPOINT_ID ELEMENT POSITION
```

**Parameters**:
- `CHECKPOINT_ID`: The ID of the checkpoint to add the step to
- `ELEMENT`: The element to scroll to (can be a selector or descriptive text)
- `POSITION`: The position in the checkpoint where the step should be inserted

**Examples**:
```bash
# Scroll to an element using descriptive text
api-cli create-step-scroll-element 1678318 "Contact form" 1

# Scroll to an element using CSS selector
api-cli create-step-scroll-element 1678318 "#footer" 2

# Create with AI-friendly output
api-cli create-step-scroll-element 1678318 ".main-content" 3 -o ai
```

**Generated Step**: `"scroll to [ELEMENT]"` (where [ELEMENT] is the provided element)

## Output Formats

All commands support the same output formats as other step creation commands:

- **human** (default): Human-readable format with checkmarks and clear messaging
- **json**: Structured JSON output for programmatic use
- **yaml**: YAML format for configuration files
- **ai**: AI-friendly format with detailed information and suggested next steps

## API Implementation

These commands use the Virtuoso API's `/teststeps` endpoint with the following action types:

- `SCROLL` action for all scroll operations
- Different `target.selectors` configurations:
  - Scroll top: `{"type": "SCROLL", "value": "{\"direction\":\"top\"}"}`
  - Scroll bottom: `{"type": "SCROLL", "value": "{\"direction\":\"bottom\"}"}`
  - Scroll to element: `{"type": "GUESS", "value": "{\"clue\":\"[ELEMENT]\"}"}`

## Error Handling

All commands include proper validation:
- Checkpoint ID must be a valid integer
- Position must be a valid integer
- Element (for scroll-element) cannot be empty
- Proper error messages are displayed for invalid inputs

## Integration

These commands are fully integrated into the existing CLI structure:
- Registered in `main.go` alongside other step creation commands
- Follow the same patterns as existing commands
- Include proper cobra.Command structure with help text and examples
- Use the same Virtuoso client methods for API communication