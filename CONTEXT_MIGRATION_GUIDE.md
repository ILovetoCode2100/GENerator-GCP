# Context and Error Handling Migration Guide

## Overview

This guide explains how to migrate all CLI commands to use context.Context support and structured error handling.

## Migration Steps

### 1. Add Context Support to Client Methods

For each client method, create a context-aware version:

```go
// Original method
func (c *Client) CreateClickStep(checkpointID int, selector string, position int) (*StepResponse, error) {
    // existing implementation
}

// New context-aware method
func (c *Client) CreateClickStepWithContext(ctx context.Context, checkpointID int, selector string, position int) (*StepResponse, error) {
    // Same implementation, but add SetContext(ctx) to resty request:
    resp, err := c.httpClient.R().
        SetContext(ctx).  // Add this line
        SetBody(body).
        SetResult(&response).
        Post("/step")

    // Handle context errors
    if err != nil {
        if errors.Is(err, context.Canceled) {
            return nil, NewClientError("CreateClickStep", KindContextCanceled, "operation canceled", err)
        }
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, NewClientError("CreateClickStep", KindTimeout, "operation timed out", err)
        }
        return nil, err
    }

    // Handle API errors using handleErrorResponse
    if resp.IsError() {
        return nil, c.handleErrorResponse(resp, "CreateClickStep")
    }

    return &response, nil
}
```

### 2. Update Command Handlers

For each command handler in the commands package:

#### A. Import Required Packages

```go
import (
    "context"
    "fmt"
    // ... other imports
)
```

#### B. Create Context at Start of RunE

```go
RunE: func(cmd *cobra.Command, args []string) error {
    // Create context with timeout
    ctx, cancel := CommandContext()
    defer cancel()

    // ... rest of command logic
}
```

#### C. Call Context-Aware Client Methods

```go
// Before
result, err := apiClient.CreateClickStep(checkpointID, selector, position)

// After
result, err := apiClient.CreateClickStepWithContext(ctx, checkpointID, selector, position)
```

#### D. Handle Errors with Better Messages

```go
if err != nil {
    // Check for specific error types
    if apiErr, ok := err.(*client.APIError); ok {
        switch apiErr.Status {
        case 401:
            return fmt.Errorf("authentication failed - please check your API token")
        case 404:
            return fmt.Errorf("checkpoint %d not found", checkpointID)
        case 429:
            if apiErr.RetryAfter > 0 {
                return fmt.Errorf("rate limit exceeded - retry after %d seconds", apiErr.RetryAfter)
            }
            return fmt.Errorf("rate limit exceeded - please try again later")
        default:
            return fmt.Errorf("API error: %s", apiErr.Message)
        }
    }

    // Check for context errors
    if errors.Is(err, context.Canceled) {
        return fmt.Errorf("operation canceled by user")
    }
    if errors.Is(err, context.DeadlineExceeded) {
        return fmt.Errorf("operation timed out")
    }

    return fmt.Errorf("failed to create step: %w", err)
}
```

### 3. Update Base Command Structure

In `base.go`, add context handling to the BaseCommand pattern:

```go
type BaseCommand struct {
    client *client.Client
    ctx    context.Context
    cancel context.CancelFunc
}

func (b *BaseCommand) Setup() error {
    // Create context
    b.ctx, b.cancel = CommandContext()

    // ... existing setup code
}

func (b *BaseCommand) Cleanup() {
    if b.cancel != nil {
        b.cancel()
    }
}
```

### 4. Exit Code Handling

In the main command runner or root command:

```go
// In cmd/api-cli/main.go or root command
if err := rootCmd.Execute(); err != nil {
    SetExitCode(rootCmd, err)
}
```

## Command Categories to Migrate

### 1. Browser Commands (`browser_commands.go`)

- [ ] Navigate commands (8 types)
- [ ] Scroll commands (7 types)
- [ ] Window commands (5 types)

### 2. Interaction Commands (`interaction_commands.go`)

- [ ] Click variants (3 types)
- [ ] Hover commands (1 type)
- [ ] Write command (1 type)
- [ ] Key command (1 type)
- [ ] Mouse commands (6 types)
- [ ] Select commands (3 types)

### 3. Assert Commands (`assert.go`)

- [ ] All 12 assertion types

### 4. Data Commands (`data.go`)

- [ ] Store commands (3 types)
- [ ] Cookie commands (3 types)

### 5. Wait Commands (`wait.go`)

- [ ] Wait element (2 types)
- [ ] Wait time (1 type)

### 6. Dialog Commands (`dialog.go`)

- [ ] Alert/confirm/prompt commands (6 types)

### 7. File Commands (`file.go`)

- [ ] Upload commands (2 types)

### 8. Execution Commands (`execution_management.go`)

- [ ] Execute goal
- [ ] Get status
- [ ] Get analysis

### 9. Library Commands (`library.go`)

- [ ] All 6 library operations

## Testing the Migration

### 1. Unit Tests

Update tests to provide context:

```go
func TestCreateClickStep(t *testing.T) {
    ctx := context.Background()

    // Test with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    result, err := client.CreateClickStepWithContext(ctx, checkpointID, selector, position)
    // ... assertions
}
```

### 2. Integration Tests

Test timeout and cancellation:

```go
// Test timeout
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
defer cancel()

_, err := client.CreateClickStepWithContext(ctx, checkpointID, selector, position)
assert.True(t, errors.Is(err, context.DeadlineExceeded))

// Test cancellation
ctx, cancel = context.WithCancel(context.Background())
go func() {
    time.Sleep(10 * time.Millisecond)
    cancel()
}()

_, err = client.LongRunningOperationWithContext(ctx)
assert.True(t, errors.Is(err, context.Canceled))
```

### 3. End-to-End Tests

Use the fixed test script with timeouts:

```bash
# Set custom timeout via environment
export VIRTUOSO_API_TIMEOUT=60s
./test-all-commands-fixed.sh
```

## Benefits After Migration

1. **Timeout Control**: Operations won't hang indefinitely
2. **Graceful Interruption**: Users can cancel with Ctrl+C
3. **Better Error Messages**: Clear, actionable error information
4. **Consistent Exit Codes**: Scripts can rely on exit codes
5. **Future-Ready**: Support for distributed tracing and metrics

## Backward Compatibility

To maintain backward compatibility during migration:

1. Keep original methods temporarily
2. Have original methods call context versions with background context:

```go
func (c *Client) CreateClickStep(checkpointID int, selector string, position int) (*StepResponse, error) {
    return c.CreateClickStepWithContext(context.Background(), checkpointID, selector, position)
}
```

3. Mark original methods as deprecated
4. Remove after grace period

## Timeline

1. **Phase 1**: Implement context support in client package (20% complete)
2. **Phase 2**: Update all command handlers (10% complete)
3. **Phase 3**: Add comprehensive tests
4. **Phase 4**: Update documentation
5. **Phase 5**: Deprecate non-context methods
6. **Phase 6**: Remove deprecated methods (future release)
