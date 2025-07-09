# Step Commands Documentation

This document provides comprehensive documentation for all step creation commands in the API CLI Generator.

## Overview

All step commands use the same base endpoint (`POST /teststeps`) and follow a consistent pattern:
- Each command creates a specific type of test step
- All commands require a checkpoint ID and position
- All commands support multiple output formats (human, json, yaml, ai)
- The `parsedStep` field uses natural language syntax that the Virtuoso API interprets

## Navigation and Control

### 1. create-step-navigate
Creates a navigation step that navigates to a URL.

```bash
api-cli create-step-navigate CHECKPOINT_ID URL POSITION
```

**Example:**
```bash
api-cli create-step-navigate 1678318 "https://example.com" 1
```

**Parsed Step:** `"Navigate to \"[URL]\""`

---

### 2. create-step-wait-time
Creates a wait time step that pauses execution for a specified number of seconds.

```bash
api-cli create-step-wait-time CHECKPOINT_ID SECONDS POSITION
```

**Example:**
```bash
api-cli create-step-wait-time 1678318 5 2
```

**Parsed Step:** `"Wait [SECONDS] seconds"`

---

### 3. create-step-wait-element
Creates a wait step that waits for an element to appear.

```bash
api-cli create-step-wait-element CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-wait-element 1678318 "Loading Complete" 3
```

**Parsed Step:** `"wait until [ELEMENT] appears"`

---

### 4. create-step-window
Creates a window resize step.

```bash
api-cli create-step-window CHECKPOINT_ID WIDTH HEIGHT POSITION
```

**Example:**
```bash
api-cli create-step-window 1678318 1920 1080 1
```

**Parsed Step:** `"Set browser window size to [WIDTH]x[HEIGHT]"`

## Mouse Actions

### 5. create-step-click
Creates a click step on an element.

```bash
api-cli create-step-click CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-click 1678318 "Sign in" 5
```

**Parsed Step:** `"click on [ELEMENT]"`

---

### 6. create-step-double-click
Creates a double-click step on an element.

```bash
api-cli create-step-double-click CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-double-click 1678318 "Element" 6
```

**Parsed Step:** `"double-click on [ELEMENT]"`

---

### 7. create-step-hover
Creates a hover step on an element.

```bash
api-cli create-step-hover CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-hover 1678318 "Menu Item" 7
```

**Parsed Step:** `"hover on [ELEMENT]"`

---

### 8. create-step-right-click
Creates a right-click step on an element.

```bash
api-cli create-step-right-click CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-right-click 1678318 "Element" 8
```

**Parsed Step:** `"right-click on [ELEMENT]"`

## Input and Forms

### 9. create-step-write
Creates a text input step.

```bash
api-cli create-step-write CHECKPOINT_ID TEXT ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-write 1678318 "username@example.com" "Email field" 9
```

**Parsed Step:** `"type \"[TEXT]\" in [ELEMENT]"`

---

### 10. create-step-key
Creates a keyboard press step.

```bash
api-cli create-step-key CHECKPOINT_ID KEY POSITION
```

**Example:**
```bash
api-cli create-step-key 1678318 "Enter" 10
```

**Parsed Step:** `"press [KEY]"`

---

### 11. create-step-pick
Creates a dropdown selection step.

```bash
api-cli create-step-pick CHECKPOINT_ID VALUE ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-pick 1678318 "United States" "Country dropdown" 11
```

**Parsed Step:** `"pick \"[VALUE]\" from [ELEMENT]"`

---

### 12. create-step-upload
Creates a file upload step.

```bash
api-cli create-step-upload CHECKPOINT_ID FILENAME ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-upload 1678318 "test-file.pdf" "File Input" 12
```

**Parsed Step:** `"upload \"[FILENAME]\" to [ELEMENT]"`

## Scroll Actions

### 13. create-step-scroll-top
Creates a scroll to top step.

```bash
api-cli create-step-scroll-top CHECKPOINT_ID POSITION
```

**Example:**
```bash
api-cli create-step-scroll-top 1678318 13
```

**Parsed Step:** `"scroll to top of page"`

---

### 14. create-step-scroll-bottom
Creates a scroll to bottom step.

```bash
api-cli create-step-scroll-bottom CHECKPOINT_ID POSITION
```

**Example:**
```bash
api-cli create-step-scroll-bottom 1678318 14
```

**Parsed Step:** `"scroll to bottom of page"`

---

### 15. create-step-scroll-element
Creates a scroll to element step.

```bash
api-cli create-step-scroll-element CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-scroll-element 1678318 "Submit Button" 15
```

**Parsed Step:** `"scroll to [ELEMENT]"`

## Assertions

### 16. create-step-assert-exists
Creates an assertion that verifies an element exists.

```bash
api-cli create-step-assert-exists CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-assert-exists 1678318 "Welcome Message" 16
```

**Parsed Step:** `"see \"[ELEMENT]\""`

---

### 17. create-step-assert-not-exists
Creates an assertion that verifies an element does not exist.

```bash
api-cli create-step-assert-not-exists CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-assert-not-exists 1678318 "Error Message" 17
```

**Parsed Step:** `"do not see \"[ELEMENT]\""`

---

### 18. create-step-assert-equals
Creates an assertion that verifies an element has a specific text value.

```bash
api-cli create-step-assert-equals CHECKPOINT_ID ELEMENT VALUE POSITION
```

**Example:**
```bash
api-cli create-step-assert-equals 1678318 "Total Price" "$99.99" 18
```

**Parsed Step:** `"expect [ELEMENT] to have text \"[VALUE]\""`

---

### 19. create-step-assert-checked
Creates an assertion that verifies a checkbox or radio button is checked.

```bash
api-cli create-step-assert-checked CHECKPOINT_ID ELEMENT POSITION
```

**Example:**
```bash
api-cli create-step-assert-checked 1678318 "Terms Checkbox" 19
```

**Parsed Step:** `"see [ELEMENT] is checked"`

## Data Operations

### 20. create-step-store
Creates a step that stores an element's value in a variable.

```bash
api-cli create-step-store CHECKPOINT_ID ELEMENT VARIABLE_NAME POSITION
```

**Example:**
```bash
api-cli create-step-store 1678318 "Order ID" "orderId" 20
```

**Parsed Step:** `"store value from [ELEMENT] as $[VARIABLE_NAME]"`

---

### 21. create-step-execute-js
Creates a step that executes JavaScript code.

```bash
api-cli create-step-execute-js CHECKPOINT_ID JAVASCRIPT POSITION
```

**Example:**
```bash
api-cli create-step-execute-js 1678318 "console.log('Test');" 21
```

**Parsed Step:** `"execute JS \"[JAVASCRIPT]\""`

## Browser Management

### 22. create-step-add-cookie
Creates a step that adds a cookie.

```bash
api-cli create-step-add-cookie CHECKPOINT_ID NAME VALUE POSITION
```

**Example:**
```bash
api-cli create-step-add-cookie 1678318 "session" "abc123" 22
```

**Parsed Step:** `"add cookie \"[NAME]\" with value \"[VALUE]\""`

---

### 23. create-step-dismiss-alert
Creates a step that dismisses JavaScript alerts.

```bash
api-cli create-step-dismiss-alert CHECKPOINT_ID POSITION
```

**Example:**
```bash
api-cli create-step-dismiss-alert 1678318 23
```

**Parsed Step:** `"dismiss alert"`

---

### 24. create-step-comment
Creates a comment step for documentation.

```bash
api-cli create-step-comment CHECKPOINT_ID COMMENT POSITION
```

**Example:**
```bash
api-cli create-step-comment 1678318 "This is a comment for documentation" 24
```

**Parsed Step:** `"# [COMMENT]"`

## Common Options

All commands support the following global flags:

- `--config string`: Specify config file (default: ./config/virtuoso-config.yaml)
- `-o, --output string`: Output format (json, yaml, human, ai) (default: "human")
- `-v, --verbose`: Enable verbose output

## Output Formats

### Human (Default)
```
✅ Created navigation step with ID: 100001
```

### JSON
```json
{
  "status": "success",
  "step_id": 100001,
  "checkpoint_id": 1678318,
  "action": "NAVIGATE",
  "position": 1
}
```

### YAML
```yaml
status: success
step_id: 100001
checkpoint_id: 1678318
action: NAVIGATE
position: 1
```

### AI
```
Successfully created navigation step:
- Step ID: 100001
- Checkpoint ID: 1678318
- Action: NAVIGATE
- Position: 1

Next steps:
1. Create additional steps: api-cli create-step-click 1678318 "element" 2
2. List steps in checkpoint: api-cli list-steps 1678318
3. Execute the test journey
```

## API Response Structure

All step creation endpoints return responses with the following structure:

```json
{
  "id": 100001,
  "checkpointId": 44444,
  "action": "NAVIGATE",
  "target": "element-or-target",
  "value": "value-if-applicable",
  "position": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "meta": {
    "additional": "metadata"
  }
}
```

## Best Practices

1. **Position Management**: Always specify positions to control step execution order
2. **Element Selection**: Use specific, unique element identifiers when possible
3. **Wait Steps**: Add wait steps after navigation or actions that trigger page changes
4. **Assertions**: Place assertions after the elements they verify are expected to appear
5. **Comments**: Use comment steps to document complex test logic
6. **Variables**: Use store steps to capture dynamic values for later use

## Error Handling

All commands validate inputs and provide helpful error messages:

- Invalid checkpoint ID format
- Invalid position (must be positive)
- Empty required parameters
- API errors with detailed messages

## Integration with Test Structure

These step commands are designed to work within the Virtuoso test structure hierarchy:

```
Project → Goal → Journey → Checkpoint → Steps
```

Steps must be created within existing checkpoints. Use the checkpoint creation commands to create checkpoints before adding steps.