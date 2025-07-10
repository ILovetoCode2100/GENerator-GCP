# Command Signature Patterns

## Identified Patterns

### Modern Pattern (Session Context)
```
Args: cobra.MinimumNArgs(1) or similar
Usage: ELEMENT [POSITION] with optional --checkpoint flag
Key: Uses resolveStepContext() from step_helpers.go
```

### Legacy Pattern (Checkpoint Required)
```
Args: cobra.ExactArgs(3) or similar
Usage: CHECKPOINT_ID ELEMENT POSITION
Key: Direct checkpoint ID parsing
```

## Commands by Pattern
### Modern Commands:
- create-step-assert-checked
- create-step-assert-equals
- create-step-assert-exists
- create-step-assert-greater-than-or-equal
- create-step-assert-greater-than
- create-step-assert-less-than-or-equal
- create-step-assert-matches
- create-step-assert-not-equals
- create-step-assert-not-exists
- create-step-assert-selected
- create-step-assert-variable
- create-step-click
- create-step-navigate
- create-step-write
