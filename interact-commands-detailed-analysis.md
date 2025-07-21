# Interact Commands Detailed Analysis

## Overview

The `interact` command group consolidates 6 types of user interactions with page elements. All interact commands use the same API endpoint (`POST /teststeps?envelope=false`) but vary in their action types and metadata.

## Command Structure

### Base Pattern

- **Modern usage (session context)**: `api-cli interact <subcommand> <selector> [text] [--flags]`
- **Legacy usage (explicit checkpoint)**: `api-cli interact <subcommand> <checkpoint-id> <selector> [text] <position>`

### API Request Body Structure

```json
{
  "checkpointId": 12345,
  "stepIndex": 1,
  "parsedStep": {
    "action": "ACTION_TYPE",
    "target": {
      /* optional */
    },
    "value": "...",
    "meta": {
      /* optional */
    }
  }
}
```

## Command Variations

### 1. Click Command

**Variations**: 3 different client methods based on parameters

#### Basic Click

```bash
api-cli interact click "button.submit"
```

- Client method: `CreateStepClick`
- API action: `CLICK`
- Target: `{"clue": "button.submit"}`

#### Click with Variable

```bash
api-cli interact click "a.link" --variable "linkText"
```

- Client method: `CreateStepClickWithVariable`
- API action: `CLICK`
- Target: `{"clue": "", "variable": "linkText"}`

#### Click with Position and Element Type

```bash
api-cli interact click "#button" --position TOP_LEFT --element-type BUTTON
```

- Client method: `CreateStepClickWithDetails`
- API action: `CLICK`
- Target: `{"clue": "#button", "position": "TOP_LEFT", "elementType": "BUTTON"}`

**Position Enums**: TOP_LEFT, TOP_CENTER, TOP_RIGHT, CENTER_LEFT, CENTER, CENTER_RIGHT, BOTTOM_LEFT, BOTTOM_CENTER, BOTTOM_RIGHT

### 2. Double-Click Command

```bash
api-cli interact double-click ".item-card" --position CENTER
```

- Client method: `CreateStepDoubleClick`
- API action: `MOUSE`
- Meta: `{"action": "DOUBLE_CLICK"}`
- No variations

### 3. Right-Click Command

```bash
api-cli interact right-click ".data-row" --position TOP_LEFT
```

- Client method: `CreateStepRightClick`
- API action: `MOUSE`
- Meta: `{"action": "RIGHT_CLICK"}`
- No variations

### 4. Hover Command

```bash
api-cli interact hover ".menu-item" --duration 2000 --position CENTER
```

- Client method: `CreateStepHover`
- API action: `MOUSE`
- Meta: `{"action": "OVER"}`
- Note: Duration parameter is accepted but not currently used in the client implementation

### 5. Write Command

**Variations**: 2 different client methods

#### Basic Write

```bash
api-cli interact write "input#username" "john.doe@example.com"
```

- Client method: `CreateStepWrite`
- API action: `WRITE`
- Value: `"john.doe@example.com"`

#### Write with Variable

```bash
api-cli interact write "#search" "{{searchTerm}}" --variable searchTerm
```

- Client method: `CreateStepWriteWithVariable`
- API action: `WRITE`
- Note: `--clear` and `--delay` flags are accepted but not used in current implementation

### 6. Key Command

**Variations**: 4 different client methods based on target and modifiers

#### Global Key Press

```bash
api-cli interact key "Enter"
```

- Client method: `CreateStepKeyGlobal`
- API action: `KEY`
- Value: `"Enter"`
- Meta: `{}`

#### Targeted Key Press

```bash
api-cli interact key "Tab" --target "input#username"
```

- Client method: `CreateStepKeyTargeted`
- API action: `KEY`
- Target: `{"clue": "input#username"}`

#### Global Key with Modifiers

```bash
api-cli interact key "a" --modifiers ctrl,shift
```

- Client method: `CreateStepKeyGlobalWithModifiers`
- API action: `KEY`
- Value: `"a"`
- Meta: `{"modifiers": ["ctrl", "shift"]}`

#### Targeted Key with Modifiers

```bash
api-cli interact key "s" --target "#editor" --modifiers ctrl
```

- Client method: `CreateStepKeyTargetedWithModifiers`
- API action: `KEY`
- Target: `{"clue": "#editor"}`
- Meta: `{"modifiers": ["ctrl"]}`

**Modifier Options**: ctrl, shift, alt, meta
**Note**: `--repeat` flag is accepted but not used in current implementation

## Parameter Order and Resolution

1. Commands use `ResolveCheckpointAndPosition` to handle both modern and legacy syntax
2. Required arguments vary:
   - Most commands: 1 (selector)
   - Write command: 2 (selector and text)
3. Position is auto-resolved from session context or explicit parameter
4. Checkpoint ID must be numeric and is validated

## Validation

- All selectors are validated using `ValidateSelector` function
- Click position enums are validated against a fixed list
- Checkpoint IDs must be convertible to integers

## Implementation Notes

1. **Mouse Actions**: double-click, right-click, and hover all use the `MOUSE` action with different meta actions
2. **Unused Parameters**: Several flags are accepted in the CLI but not implemented in the client:
   - hover: `--duration`
   - write: `--clear`, `--delay`
   - key: `--repeat`
3. **Variable Support**: Only click and write commands support variable storage/usage
4. **Target Specifications**: All commands except key use the GUESS selector type with clue JSON
