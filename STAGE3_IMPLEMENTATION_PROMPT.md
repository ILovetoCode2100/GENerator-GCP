# Virtuoso API CLI - Stage 3 Implementation Guide

## Context

You are implementing the final stage of the Virtuoso API CLI, a Go-based command-line tool for test automation. Stages 1 and 2 have already been completed:

**Stage 1 Completed:**

- Navigation commands (back, forward, refresh)
- Window management (maximize, close, switch-tab-index)
- Data storage (store-attribute)

**Stage 2 Completed:**

- Click position enums (TOP_LEFT, CENTER, etc.)
- Wait element-not-visible
- Scroll-by with X,Y offsets
- Cookie operations with domain/path/expiry

This is Stage 3, focusing on:

1. Library checkpoint step management (move/remove)
2. Multi-key combinations and complex keyboard shortcuts
3. Advanced frame/iframe operations
4. Browser navigation history manipulation
5. Any remaining missing functionality

## Project Structure

```
virtuoso-GENerator/
├── pkg/api-cli/
│   ├── client/client.go      # API client methods
│   ├── commands/
│   │   ├── base.go          # Shared command infrastructure
│   │   ├── library.go       # Library operations (needs enhancement)
│   │   ├── interact.go      # User interactions (needs keyboard enhancements)
│   │   ├── navigate.go      # Navigation (needs history manipulation)
│   │   └── window.go        # Window management (needs frame enhancements)
│   └── config/config.go     # Configuration management
└── bin/api-cli              # Compiled binary
```

## Stage 3 Implementation Tasks

### 1. Library Checkpoint Step Management

The library commands already support basic operations but need step manipulation features.

**Files to modify:**

- `/pkg/api-cli/client/client.go` - Already has methods, ensure they work correctly
- `/pkg/api-cli/commands/library.go` - Already has commands, verify implementation

**Verify these existing methods work correctly:**

```go
// These should already exist in client.go:
func (c *Client) MoveLibraryCheckpointStep(libraryCheckpointID, testStepID, position int) error
func (c *Client) RemoveLibraryCheckpointStep(libraryCheckpointID, testStepID int) error
```

**Test the existing library commands:**

```bash
# Move a step within a library checkpoint
api-cli library move-step 7023 19660498 2

# Remove a step from a library checkpoint
api-cli library remove-step 7023 19660498
```

### 2. Multi-Key Combinations and Complex Keyboard Shortcuts

**Files to modify:**

- `/pkg/api-cli/client/client.go` - Enhance key methods to support modifiers
- `/pkg/api-cli/commands/interact.go` - The key command already accepts modifiers but needs proper implementation

**Update client.go to support key combinations:**

```go
// Enhanced CreateStepKeyGlobal with modifier support
func (c *Client) CreateStepKeyGlobalWithModifiers(checkpointID int, key string, modifiers []string, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "KEY",
        "value":  key,
        "meta": map[string]interface{}{
            "modifiers": modifiers, // ["ctrl", "shift", "alt", "meta"]
        },
    }

    return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// Enhanced CreateStepKeyTargeted with modifier support
func (c *Client) CreateStepKeyTargetedWithModifiers(checkpointID int, selector, key string, modifiers []string, position int) (int, error) {
    clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

    parsedStep := map[string]interface{}{
        "action": "KEY",
        "target": map[string]interface{}{
            "selectors": []map[string]interface{}{
                {
                    "type":  "GUESS",
                    "value": clueJSON,
                },
            },
        },
        "value": key,
        "meta": map[string]interface{}{
            "modifiers": modifiers,
        },
    }

    return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}
```

**Update interact.go executeKeyAction:**

```go
func executeKeyAction(c *client.Client, checkpointID int, key string, position int, options map[string]interface{}) (int, error) {
    target, _ := options["target"].(string)
    modifiers, _ := options["modifiers"].([]string)

    // Handle modifier keys
    if len(modifiers) > 0 {
        if target != "" {
            if err := ValidateSelector(target); err != nil {
                return 0, err
            }
            return c.CreateStepKeyTargetedWithModifiers(checkpointID, target, key, modifiers, position)
        } else {
            return c.CreateStepKeyGlobalWithModifiers(checkpointID, key, modifiers, position)
        }
    }

    // Original implementation for simple keys
    if target != "" {
        if err := ValidateSelector(target); err != nil {
            return 0, err
        }
        return c.CreateStepKeyTargeted(checkpointID, target, key, position)
    } else {
        return c.CreateStepKeyGlobal(checkpointID, key, position)
    }
}
```

**Test commands:**

```bash
# Ctrl+A (select all)
api-cli interact key "a" --modifiers ctrl

# Ctrl+Shift+Tab (previous tab)
api-cli interact key "Tab" --modifiers ctrl,shift

# Alt+F4 (close window)
api-cli interact key "F4" --modifiers alt

# Cmd+S on Mac
api-cli interact key "s" --modifiers meta
```

### 3. Advanced Frame/Iframe Operations

**Files to modify:**

- `/pkg/api-cli/client/client.go` - Add frame navigation by index/name
- `/pkg/api-cli/commands/window.go` - Add new frame commands

**Add to client.go:**

```go
// Switch to frame by index (0-based)
func (c *Client) CreateStepSwitchFrameByIndex(checkpointID int, index int, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "SWITCH",
        "meta": map[string]interface{}{
            "type": "FRAME_INDEX",
            "index": index,
        },
    }

    return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// Switch to frame by name attribute
func (c *Client) CreateStepSwitchFrameByName(checkpointID int, name string, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "SWITCH",
        "meta": map[string]interface{}{
            "type": "FRAME_NAME",
            "name": name,
        },
    }

    return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// Switch to main/default content (exit all frames)
func (c *Client) CreateStepSwitchToMainContent(checkpointID int, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "SWITCH",
        "meta": map[string]interface{}{
            "type": "DEFAULT_CONTENT",
        },
    }

    return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}
```

**Add to window.go windowCommands map:**

```go
windowSwitchFrameIndex: {
    stepType:    "SWITCH",
    description: "Switch to frame by index (0-based)",
    usage:       "window switch frame-index INDEX [POSITION]",
    examples: []string{
        `api-cli window switch frame-index 0 1  # Switch to first frame`,
        `api-cli window switch frame-index 2    # Switch to third frame`,
    },
    argsCount: []int{1},
    parseStep: func(args []string) string {
        return fmt.Sprintf("switch to frame index %s", args[0])
    },
},
windowSwitchFrameName: {
    stepType:    "SWITCH",
    description: "Switch to frame by name attribute",
    usage:       "window switch frame-name NAME [POSITION]",
    examples: []string{
        `api-cli window switch frame-name "content" 1`,
        `api-cli window switch frame-name "paymentFrame"`,
    },
    argsCount: []int{1},
    parseStep: func(args []string) string {
        return fmt.Sprintf("switch to frame named \"%s\"", args[0])
    },
},
windowSwitchMainContent: {
    stepType:    "SWITCH",
    description: "Switch to main content (exit all frames)",
    usage:       "window switch main-content [POSITION]",
    examples: []string{
        `api-cli window switch main-content 1`,
        `api-cli window switch main-content`,
    },
    argsCount: []int{0},
    parseStep: func(args []string) string {
        return "switch to main content"
    },
},
```

**Update window.go to handle new frame operations in the switch statement.**

### 4. Browser Navigation History Manipulation

**Files to modify:**

- `/pkg/api-cli/client/client.go` - Add history navigation methods
- `/pkg/api-cli/commands/navigate.go` - Add history commands

**Add to client.go:**

```go
// Go back N steps in browser history
func (c *Client) CreateStepNavigateBackN(checkpointID int, steps int, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "NAVIGATE",
        "meta": map[string]interface{}{
            "kind": "BACK",
            "steps": steps,
        },
    }

    return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// Go forward N steps in browser history
func (c *Client) CreateStepNavigateForwardN(checkpointID int, steps int, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "NAVIGATE",
        "meta": map[string]interface{}{
            "kind": "FORWARD",
            "steps": steps,
        },
    }

    return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}

// Navigate to specific history entry by offset (negative for back, positive for forward)
func (c *Client) CreateStepNavigateHistory(checkpointID int, offset int, position int) (int, error) {
    parsedStep := map[string]interface{}{
        "action": "NAVIGATE",
        "meta": map[string]interface{}{
            "kind": "HISTORY",
            "offset": offset,
        },
    }

    return c.createStepWithCustomBody(checkpointID, parsedStep, position)
}
```

**Update navigate.go to add options to back/forward commands:**

```go
// In navigateBackSubCmd, add:
var steps int
cmd.Flags().IntVar(&steps, "steps", 1, "Number of steps to go back")

// In navigateForwardSubCmd, add:
var steps int
cmd.Flags().IntVar(&steps, "steps", 1, "Number of steps to go forward")

// Update executeNavigateBackAction:
func executeNavigateBackAction(c *client.Client, checkpointID int, position int, options map[string]interface{}) (int, error) {
    steps, _ := options["steps"].(int)
    if steps <= 0 {
        steps = 1
    }

    if steps == 1 {
        return c.CreateStepNavigateBack(checkpointID, position)
    }
    return c.CreateStepNavigateBackN(checkpointID, steps, position)
}
```

### 5. Additional Missing Functionality

**A. Simple Directional Scrolling**

Add to navigate.go:

```go
// Add scroll-up and scroll-down commands
scrollUpSubCmd: {
    Use:   "scroll-up [checkpoint-id] [position]",
    Short: "Scroll up by one viewport height",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation: scroll by -window.innerHeight
    },
},
scrollDownSubCmd: {
    Use:   "scroll-down [checkpoint-id] [position]",
    Short: "Scroll down by one viewport height",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation: scroll by +window.innerHeight
    },
},
```

**B. Enhanced Click with Options**

The click command already has position and elementType flags but ensure they work properly with the API.

**C. Store Element Attribute**

This is already implemented in Stage 1 as part of the data commands. Verify it works:

```bash
api-cli data store attribute "a.link" "href" "linkUrl"
```

## Testing Requirements

Create a comprehensive test script `test-stage3-features.sh`:

```bash
#!/bin/bash

# Test Stage 3 Features

# 1. Library Step Management
echo "Testing library step management..."
./bin/api-cli library move-step 7023 19660498 3
./bin/api-cli library remove-step 7023 19660499

# 2. Multi-key Combinations
echo "Testing keyboard combinations..."
./bin/api-cli interact key "a" --modifiers ctrl
./bin/api-cli interact key "Tab" --modifiers ctrl,shift
./bin/api-cli interact key "s" --modifiers meta --target "#editor"

# 3. Advanced Frame Operations
echo "Testing frame operations..."
./bin/api-cli window switch frame-index 0
./bin/api-cli window switch frame-name "payment"
./bin/api-cli window switch main-content

# 4. Browser History
echo "Testing browser history..."
./bin/api-cli navigate back --steps 2
./bin/api-cli navigate forward --steps 1

# 5. Additional Features
echo "Testing additional features..."
./bin/api-cli navigate scroll-up
./bin/api-cli navigate scroll-down
./bin/api-cli data store attribute "img" "src" "imageUrl"
```

## Implementation Order

1. **Start with keyboard modifiers** - Update interact.go and client methods
2. **Add frame operations** - Extend window commands
3. **Enhance navigation history** - Add steps parameter to back/forward
4. **Verify library commands** - Test existing move-step and remove-step
5. **Add simple scroll commands** - scroll-up/scroll-down
6. **Comprehensive testing** - Run all test scripts

## API Endpoints Reference

- Library operations: `/library-checkpoints/{id}/steps/{stepId}` (PUT for move, DELETE for remove)
- Frame switching uses the standard step creation endpoint with appropriate meta fields
- History navigation uses standard NAVIGATE action with different meta.kind values

## Success Criteria

- All 73 original commands have equivalent functionality in the consolidated structure
- Multi-key combinations work correctly (Ctrl+A, Cmd+S, etc.)
- Frame navigation supports selector, index, name, and main content
- Browser history can go back/forward multiple steps
- Library step management works reliably
- All commands maintain backward compatibility through legacy wrappers
- Comprehensive test coverage with >95% success rate

## Notes

- Maintain the consolidated command structure - don't add new top-level commands
- Use existing patterns from base.go for consistency
- All commands should support --dry-run, session context, and all output formats
- Update help text and examples for clarity
- Test with real Virtuoso API to ensure compatibility
