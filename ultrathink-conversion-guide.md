# ULTRATHINK Command Conversion Guide

## Pattern Changes Required

### 1. Update Command Signature
```go
// OLD
Use: "create-step-wait-time CHECKPOINT_ID SECONDS POSITION"
Args: cobra.ExactArgs(3)

// NEW
Use: "create-step-wait-time SECONDS [POSITION]"
Args: cobra.RangeArgs(1, 2) // or custom validation for legacy support
```

### 2. Add Checkpoint Flag
```go
var checkpointFlag int
// In command creation
addCheckpointFlag(cmd, &checkpointFlag)
```

### 3. Use resolveStepContext
```go
// OLD
checkpointID, err := strconv.Atoi(args[0])
position, err := strconv.Atoi(args[2])

// NEW
ctx, err := resolveStepContext(args, checkpointFlag, 1)
checkpointID := ctx.CheckpointID
position := ctx.Position
```

### 4. Save Context After Success
```go
saveStepContext(ctx)
```

### 5. Use outputStepResult
```go
// OLD - custom output formatting
// NEW
output := &StepOutput{
    Status:       "success",
    StepType:     "WAIT_TIME",
    CheckpointID: ctx.CheckpointID,
    StepID:       stepID,
    Position:     ctx.Position,
    ParsedStep:   fmt.Sprintf("Wait %d seconds", seconds),
    UsingContext: ctx.UsingContext,
    AutoPosition: ctx.AutoPosition,
    Extra:        map[string]interface{}{"seconds": seconds},
}
return outputStepResult(output)
```

## Commands to Update

### Navigation (3 commands)
- wait-time: SECONDS [POSITION]
- wait-element: ELEMENT [POSITION]
- window: WIDTH HEIGHT [POSITION]

### Mouse (7 commands)
- double-click: ELEMENT [POSITION]
- right-click: ELEMENT [POSITION]
- hover: ELEMENT [POSITION]
- mouse-down: ELEMENT [POSITION]
- mouse-up: ELEMENT [POSITION]
- mouse-move: X Y [POSITION] (or ELEMENT [POSITION])
- mouse-enter: ELEMENT [POSITION]

### Input (5 commands)
- key: KEY [POSITION]
- pick: ELEMENT INDEX [POSITION]
- pick-value: ELEMENT VALUE [POSITION]
- pick-text: ELEMENT TEXT [POSITION]
- upload: ELEMENT FILE_PATH [POSITION]

### Scroll (4 commands)
- scroll-top: [POSITION]
- scroll-bottom: [POSITION]
- scroll-element: ELEMENT [POSITION]
- scroll-position: Y_POSITION [POSITION]

### Data (3 commands)
- store: ELEMENT VARIABLE_NAME [POSITION]
- store-value: ELEMENT VARIABLE_NAME [POSITION]
- execute-js: JAVASCRIPT [VARIABLE_NAME] [POSITION]

### Environment (3 commands)
- add-cookie: NAME VALUE DOMAIN PATH [POSITION]
- delete-cookie: NAME [POSITION]
- clear-cookies: [POSITION]

### Dialog (3 commands)
- dismiss-alert: [POSITION]
- dismiss-confirm: ACCEPT [POSITION]
- dismiss-prompt: TEXT [POSITION]

### Frame/Tab (4 commands)
- switch-iframe: ELEMENT [POSITION]
- switch-next-tab: [POSITION]
- switch-prev-tab: [POSITION]
- switch-parent-frame: [POSITION]

### Utility (1 command)
- comment: COMMENT [POSITION]
