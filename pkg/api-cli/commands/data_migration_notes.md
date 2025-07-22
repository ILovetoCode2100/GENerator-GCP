# Data Commands Migration to Context-Aware Client Methods

## Overview

The data commands in `data.go` have been successfully migrated to use the new context-aware client methods, providing better error handling, timeout management, and user-friendly error messages.

## Key Changes

### 1. Import Updates

Added required imports:

- `context` - For creating command contexts
- `errors` - For error type checking with `errors.As()`
- `client` package - For accessing error types

### 2. Method Updates in createDataStep()

#### Store Operations

- `CreateStoreStep` → `CreateStepStoreElementTextWithContext`
- `CreateStoreValueStep` → `CreateStepStoreLiteralValueWithContext`
- `CreateStepStoreAttribute` → `CreateStepStoreAttributeWithContext`

#### Cookie Operations

- `CreateStepCookieCreate` → `CreateStepCookieCreateWithContext`
- `CreateStepCookieCreateWithOptions` → `CreateStepCookieCreateWithOptionsWithContext`
- `CreateDeleteCookieStep` → `CreateStepCookieDeleteWithContext`
- `CreateClearCookiesStep` → `CreateStepCookieClearAllWithContext`

### 3. Context Management

```go
// Create context with timeout
ctx, cancel := dc.CommandContext()
defer cancel()
```

- Uses `CommandContext()` from BaseCommand (30-second timeout)
- Properly cancels context with defer

### 4. Enhanced Error Handling

#### API Errors

Provides specific messages for different HTTP status codes:

- 400: Invalid request details
- 401: Authentication failure hint
- 403: Permission denied explanation
- 404: Checkpoint not found
- 429: Rate limit with retry information
- 5xx: Server error with retry suggestion

#### Client Errors

Handles network and timeout scenarios:

- `KindTimeout`: Request timeout message
- `KindContextCanceled`: Operation canceled
- `KindConnectionFailed`: Network error with retry hint

## Benefits

1. **Better User Experience**

   - Clear, actionable error messages
   - Specific guidance for common issues
   - Rate limit retry information

2. **Improved Reliability**

   - 30-second timeout prevents hanging
   - Context cancellation support
   - Proper error propagation

3. **Consistent Error Handling**
   - Uses standard error types from client package
   - Follows same pattern as other migrated commands
   - Type-safe error checking with `errors.As()`

## Testing

Use the test script at `test-commands/test-data-context.sh` to verify:

- All store operations work correctly
- Cookie operations handle options properly
- Error messages are user-friendly
- Timeouts are respected

## Example Error Messages

Before:

```
Error: add step failed with status 404: {"error":"Not found"}
```

After:

```
Error: checkpoint not found: 12345
```

This migration ensures data commands are now more robust and provide a better user experience when errors occur.
