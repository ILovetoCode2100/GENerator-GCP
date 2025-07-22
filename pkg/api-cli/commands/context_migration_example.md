# Context Migration Example: Click Command

This document demonstrates the migration pattern for adding context support to CLI commands using the click command as an example.

## Overview

The migration adds context support to enable:

- Proper timeout handling (30-second default)
- Graceful cancellation of operations
- Better error messages based on context state
- Foundation for future enhancements (tracing, metrics, etc.)

## Key Changes

### 1. Import Context Package

```go
import (
    "context"
    // ... other imports
)
```

### 2. Add CommandContext Method to BaseCommand

```go
// CommandContext creates a context with timeout for API operations
func (bc *BaseCommand) CommandContext() (context.Context, context.CancelFunc) {
    // Default timeout of 30 seconds for API operations
    return context.WithTimeout(context.Background(), 30*time.Second)
}
```

### 3. Update Command RunE Function

Before:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    return runInteraction(cmd, args, "click", map[string]interface{}{
        "variable":    variable,
        "position":    position,
        "elementType": elementType,
    })
}
```

After:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    // Create base command and initialize
    base := NewBaseCommand()
    if err := base.Init(cmd); err != nil {
        return fmt.Errorf("initialization failed: %w", err)
    }

    // Create context for this operation
    ctx, cancel := base.CommandContext()
    defer cancel()

    // Run the interaction with context
    return runInteractionWithContext(ctx, base, cmd, args, "click", map[string]interface{}{
        "variable":    variable,
        "position":    position,
        "elementType": elementType,
    })
}
```

### 4. Create Context-Aware Execution Function

```go
func runInteractionWithContext(ctx context.Context, base *BaseCommand, cmd *cobra.Command, args []string, action string, options map[string]interface{}) error {
    // ... argument resolution ...

    // Execute with context
    switch action {
    case "click":
        stepID, err = executeClickActionWithContext(ctx, base.Client, checkpointID, args[0], base.Position, options)
    // ... other actions ...
    }

    // Enhanced error handling
    if err != nil {
        // Check context errors first
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("operation timed out after 30 seconds while creating %s step", action)
        } else if ctx.Err() == context.Canceled {
            return fmt.Errorf("operation was canceled while creating %s step", action)
        }

        // Check for common API errors and provide helpful messages
        errMsg := err.Error()
        switch {
        case strings.Contains(errMsg, "404"):
            return fmt.Errorf("checkpoint not found (ID: %d). Please verify the checkpoint exists", checkpointID)
        case strings.Contains(errMsg, "401"):
            return fmt.Errorf("authentication failed. Please check your API token in the configuration")
        // ... more error cases ...
        }
    }
    // ... rest of function ...
}
```

### 5. Update Action Execution Function

```go
func executeClickActionWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int, options map[string]interface{}) (int, error) {
    // Check if context is already cancelled
    select {
    case <-ctx.Done():
        return 0, fmt.Errorf("operation cancelled before execution: %w", ctx.Err())
    default:
    }

    // ... validation ...

    // Use context-aware wrapper functions
    if variable != "" {
        return CreateStepClickWithVariableWithContext(ctx, c, checkpointID, variable, position)
    } else if positionType != "" && elementType != "" {
        return CreateStepClickWithDetailsWithContext(ctx, c, checkpointID, selector, positionType, elementType, position)
    } else {
        return CreateStepClickWithContext(ctx, c, checkpointID, selector, position)
    }
}
```

### 6. Temporary Context Wrappers

Until the client package supports context natively, we use wrapper functions:

```go
func CreateStepClickWithContext(ctx context.Context, c *client.Client, checkpointID int, selector string, position int) (int, error) {
    // Create a channel for the result
    type result struct {
        stepID int
        err    error
    }
    resultChan := make(chan result, 1)

    // Run the API call in a goroutine
    go func() {
        stepID, err := c.CreateStepClick(checkpointID, selector, position)
        resultChan <- result{stepID: stepID, err: err}
    }()

    // Wait for either result or context cancellation
    select {
    case res := <-resultChan:
        return res.stepID, res.err
    case <-ctx.Done():
        return 0, fmt.Errorf("operation cancelled: %w", ctx.Err())
    }
}
```

## Migration Benefits

1. **Timeout Protection**: Commands automatically timeout after 30 seconds
2. **Better Error Messages**: Users get specific error messages for different failure scenarios
3. **Cancellation Support**: Foundation for future CTRL+C handling
4. **Backward Compatibility**: Legacy functions remain for gradual migration
5. **Consistent Pattern**: Same approach can be applied to all commands

## Migration Steps for Other Commands

1. **Update the command's RunE function** to create context and call context-aware version
2. **Create context-aware execution function** (e.g., `runNavigationWithContext`)
3. **Update action execution functions** to accept context parameter
4. **Add temporary wrapper functions** for client methods that don't support context
5. **Keep legacy versions** for backward compatibility during migration
6. **Test thoroughly** to ensure error handling works correctly

## Future Improvements

Once the client package is updated with native context support:

1. Remove temporary wrapper functions
2. Update client calls to use context-aware methods directly
3. Add proper HTTP request cancellation
4. Consider adding request tracing and metrics

## Testing the Migration

```bash
# Test normal operation
api-cli interact click "button.submit"

# Test with invalid checkpoint (404 error)
api-cli interact click 99999 "button.submit"

# Test with invalid token (401 error)
VIRTUOSO_API_TOKEN=invalid api-cli interact click "button.submit"

# Test timeout (if API is slow)
# The operation will timeout after 30 seconds with a clear message
```

## Summary

This migration pattern provides a clean way to add context support while maintaining backward compatibility. The enhanced error handling improves the user experience by providing specific, actionable error messages. The pattern is consistent and can be applied to all commands in the codebase.
