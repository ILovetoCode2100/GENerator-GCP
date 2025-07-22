# Assert Command Context Migration Example

This document demonstrates the migration pattern for updating assert commands to use context-aware client methods.

## Key Changes Made

### 1. Import Context Package

```go
import (
    "context"
    // ... other imports
    "virtuoso-cli/pkg/api-cli/client"
)
```

### 2. Update createAssertStep Method

**Before:**

```go
stepID, err = ac.Client.CreateAssertExistsStep(checkpointID, meta["selector"].(string), ac.Position)
```

**After:**

```go
// Create a context with timeout for the API operation
ctx, cancel := ac.CommandContext()
defer cancel()

// Use context-aware method
stepID, err = ac.Client.CreateAssertExistsStepWithContext(ctx, checkpointID, meta["selector"].(string), ac.Position)
```

### 3. Enhanced Error Handling

```go
if err != nil {
    // Handle context-specific errors
    if err == context.DeadlineExceeded {
        return nil, fmt.Errorf("request timed out while creating %s step", stepType)
    }
    if err == context.Canceled {
        return nil, fmt.Errorf("request was canceled while creating %s step", stepType)
    }

    // Check for specific API error types
    if client.IsNotFound(err) {
        return nil, fmt.Errorf("checkpoint %d not found", checkpointID)
    }
    if client.IsUnauthorized(err) {
        return nil, fmt.Errorf("unauthorized: please check your API token")
    }
    if client.IsRateLimited(err) {
        return nil, fmt.Errorf("rate limited: please try again later")
    }
    if client.IsTimeout(err) {
        return nil, fmt.Errorf("API request timed out")
    }

    // For API errors, provide more context
    if apiErr, ok := err.(*client.APIError); ok {
        return nil, fmt.Errorf("API error creating %s step: %v", stepType, apiErr)
    }

    // Generic error
    return nil, fmt.Errorf("failed to create %s step: %w", stepType, err)
}
```

## Benefits

1. **Timeout Control**: Commands now respect a 30-second timeout for API operations
2. **Better Error Messages**: Users get specific, actionable error messages
3. **Context Propagation**: Supports cancellation and deadline propagation
4. **Consistent Pattern**: Same pattern can be applied to all command types

## Migration Checklist

When migrating other commands, follow these steps:

- [ ] Import context package
- [ ] Import client package for error types
- [ ] Create context using `CommandContext()` at the start of API operations
- [ ] Always defer cancel() to clean up resources
- [ ] Replace client method calls with WithContext versions
- [ ] Add comprehensive error handling for different error types
- [ ] Test timeout scenarios
- [ ] Test error scenarios (404, 401, 429, etc.)

## Example Usage

```bash
# Command will timeout after 30 seconds
api-cli assert exists "Login button"

# Error examples:
# "request timed out while creating ASSERT_EXISTS step"
# "checkpoint 12345 not found"
# "unauthorized: please check your API token"
# "rate limited: please try again later"
```
