# Response Handler Integration Guide

This guide shows how to integrate the new ResponseHandler into existing client methods to solve API response parsing issues.

## Common Issues Solved

1. **Mixed ID types**: APIs return IDs as strings, numbers, or floats
2. **Variable response formats**: IDs found at different paths (item.id, testStep.id, data.id, etc.)
3. **Placeholder IDs**: Dialog and mouse commands return ID: 1
4. **Missing IDs**: Some commands succeed but don't return IDs
5. **Type conversion errors**: "cannot unmarshal number into Go struct field"

## Integration Examples

### 1. Basic Step Creation

Replace existing step creation methods:

```go
// Before: Direct parsing that fails with type mismatches
var response struct {
    ID int `json:"id"`
}
if err := json.Unmarshal(resp.Body(), &response); err != nil {
    return 0, err
}

// After: Using ResponseHandler
handler := NewResponseHandler(c.config.Output.Verbose)
stepID, err := handler.ParseStepID(resp.Body())
if err != nil {
    if IsPlaceholderError(err) {
        c.logger.Warnf("Step created with placeholder ID: %v", err)
        return stepID, nil
    }
    if IsNoIDButSuccessError(err) {
        c.logger.Warn("Step created but no ID returned")
        return 0, nil
    }
    return 0, err
}
```

### 2. Execution Handling

For ExecuteGoal and similar commands:

```go
// Use ParseResponse for complex types
handler := NewResponseHandler(c.config.Output.Verbose)
var execution Execution
if err := handler.ParseResponse(resp.Body(), &execution); err != nil {
    if IsNoIDButSuccessError(err) {
        c.logger.Warn("Execution created but no ID returned")
        return &execution, nil
    }
    return nil, err
}
```

### 3. Dialog Commands

Handle placeholder IDs gracefully:

```go
func (c *Client) CreateDialogStep(checkpointID int, dialogType string, position int) (int, error) {
    // ... create step ...

    handler := NewResponseHandler(c.config.Output.Verbose)
    stepID, err := handler.ParseStepID(resp.Body())
    if err != nil {
        if IsPlaceholderError(err) {
            // Dialog steps often return ID: 1
            c.logger.Info("Dialog step created (placeholder ID returned)")
            return 0, nil // Don't propagate placeholder IDs
        }
        return 0, err
    }

    return stepID, nil
}
```

### 4. Batch Operations

For commands that process multiple items:

```go
func (c *Client) ProcessBatchResponse(resp []byte) ([]int, error) {
    handler := NewResponseHandler(c.config.Output.Verbose)

    var wrapper struct {
        Items []json.RawMessage `json:"items"`
    }

    if err := json.Unmarshal(resp, &wrapper); err != nil {
        return nil, err
    }

    var ids []int
    for i, item := range wrapper.Items {
        id, err := handler.ParseStepID(item)
        if err != nil {
            if IsNoIDButSuccessError(err) {
                c.logger.Warnf("Item %d created but no ID returned", i)
                continue
            }
            return nil, fmt.Errorf("failed to parse item %d: %w", i, err)
        }
        ids = append(ids, id)
    }

    return ids, nil
}
```

## Error Handling Patterns

### 1. Graceful Degradation

```go
stepID, err := handler.ParseStepID(resp.Body())
if err != nil {
    switch {
    case IsPlaceholderError(err):
        // Log but continue - step was created
        c.logger.Warn(err)
        return stepID, nil

    case IsNoIDButSuccessError(err):
        // Step created but no ID available
        c.logger.Info("Step created successfully")
        return 0, nil

    default:
        // Real error
        return 0, err
    }
}
```

### 2. Strict Validation

```go
stepID, err := handler.ParseStepID(resp.Body())
if err != nil {
    return 0, fmt.Errorf("invalid response: %w", err)
}

// Additional validation
if err := handler.ValidateStepResponse(stepID, resp.Body()); err != nil {
    return 0, fmt.Errorf("invalid step ID: %w", err)
}
```

## Migration Checklist

When updating a method to use ResponseHandler:

1. ✅ Replace direct JSON unmarshaling with handler.ParseResponse() or handler.ParseStepID()
2. ✅ Add error type checking for PlaceholderError and NoIDButSuccessError
3. ✅ Update logging to provide clear feedback about response issues
4. ✅ Test with various response formats (wrapped, unwrapped, numeric, string IDs)
5. ✅ Ensure backward compatibility with existing callers

## Testing

The response handler includes comprehensive tests. When integrating:

```go
// Test with various response formats
responses := []string{
    `{"id": 12345}`,
    `{"item": {"id": "12345"}}`,
    `{"testStep": {"id": 12345.0}}`,
    `{"success": true, "item": {"id": 1}}`, // Placeholder
}

for _, resp := range responses {
    handler := NewResponseHandler(true) // Enable debug
    id, err := handler.ParseStepID([]byte(resp))
    // Handle based on your requirements
}
```

## Benefits

1. **Robust parsing**: Handles all known response format variations
2. **Type safety**: Automatic conversion between numeric and string types
3. **Clear errors**: Distinguishes between real failures and API quirks
4. **Future-proof**: Easy to add new response formats as discovered
5. **Debugging**: Optional debug output shows where IDs were found

## Notes

- The handler attempts multiple parsing strategies before failing
- Placeholder IDs (like 1) are detected but still returned for logging
- Success responses without IDs are handled gracefully
- All time formats are supported including Unix timestamps
- Large numbers are handled correctly without overflow
