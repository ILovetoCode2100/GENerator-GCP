# Virtuoso API CLI - Implementation Guide for Error Fixes

## Overview

This guide provides step-by-step instructions for implementing the comprehensive error fixes documented in `ERROR_ANALYSIS_AND_SOLUTIONS.md`. The fixes address critical issues in API response parsing, command reliability, type system compatibility, and integration problems.

## Implementation Priority

### Phase 1: Critical Fixes (Immediate)

These fixes address blocking issues that prevent the CLI from functioning correctly.

#### 1.1 Fix ExecuteGoal Unmarshal Error

**Files to modify:**

- `pkg/api-cli/client/client.go`

**Implementation steps:**

1. Replace the existing `ExecuteGoal` function with the implementation from `client_fixes.go`:

```go
// In client.go, replace ExecuteGoal with:
func (c *Client) ExecuteGoal(goalID, snapshotID int) (*Execution, error) {
    return c.ExecuteGoalFixed(goalID, snapshotID)
}
```

2. Add the response parser import:

```go
import (
    // ... existing imports
    "encoding/json"
)
```

#### 1.2 Implement Response Parser

**Files to add:**

- `pkg/api-cli/client/response_parser.go` (already created)
- `pkg/api-cli/client/response_parser_test.go` (already created)

**Implementation steps:**

1. Integrate the response parser into all step creation methods:

```go
// In client.go, update addStep method:
func (c *Client) addStep(checkpointID int, stepIndex int, parsedStep map[string]interface{}) (int, error) {
    return c.addStepFixed(checkpointID, stepIndex, parsedStep)
}
```

#### 1.3 Fix YAML Type Normalization

**Files to add:**

- `pkg/api-cli/commands/yaml_normalizer.go` (already created)

**Files to modify:**

- `pkg/api-cli/commands/run_test_cmd.go`
- `pkg/api-cli/commands/test_templates.go`

**Implementation steps:**

1. Update YAML loading in `run_test_cmd.go`:

```go
// Replace existing YAML parsing with:
parser := NewYAMLParser(false)
data, err := parser.Parse(yamlContent)
if err != nil {
    return fmt.Errorf("failed to parse YAML: %w", err)
}

// Detect and convert format
format := DetectYAMLFormat(data)
standardData, err := ConvertToStandardFormat(data, format)
if err != nil {
    return fmt.Errorf("failed to convert YAML format: %w", err)
}
```

### Phase 2: Reliability Improvements (High Priority)

#### 2.1 Add Retry Mechanism

**Files to add:**

- `pkg/api-cli/client/retry.go` (already created)

**Files to modify:**

- `pkg/api-cli/client/client.go`

**Implementation steps:**

1. Create retryable client wrapper in commands that need it:

```go
// In commands that make API calls:
retryClient := client.NewRetryableClient(apiClient, nil)
execution, err := retryClient.ExecuteGoalWithRetry(ctx, goalID, snapshotID)
```

#### 2.2 Implement Command Validator

**Files to add:**

- `pkg/api-cli/commands/command_validator.go` (already created)

**Files to modify:**

- `pkg/api-cli/commands/base.go`

**Implementation steps:**

1. Add validator to BaseCommand:

```go
type BaseCommand struct {
    // ... existing fields
    validator *CommandValidator
}

func NewBaseCommand() *BaseCommand {
    return &BaseCommand{
        // ... existing initialization
        validator: NewCommandValidator(),
    }
}
```

2. Add validation in command execution:

```go
func (bc *BaseCommand) Execute(cmd string, args []string) error {
    // Validate and correct command
    cmd, args, err := bc.validator.ValidateAndCorrect(cmd, args)
    if err != nil {
        return fmt.Errorf("command validation failed: %w", err)
    }

    // ... rest of execution
}
```

### Phase 3: Integration Fixes (Medium Priority)

#### 3.1 Fix Pre-commit Compatibility

**Files to create:**

- `.pre-commit-config.yaml`

**Content:**

```yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.55.2
    hooks:
      - id: golangci-lint
        args: ["--fix"]

  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: go fmt ./...
        language: system
        types: [go]
        pass_filenames: false
```

#### 3.2 Update Deprecated Endpoints

**Files to modify:**

- `pkg/api-cli/client/client.go`

**Implementation steps:**

1. Update endpoint URLs:

```go
// Replace deprecated endpoints
const (
    // Old: "/checkpoints/{id}/steps"
    EndpointTestSteps = "/teststeps"

    // Old: "/goals/{id}/execute"
    EndpointExecutions = "/executions"
)
```

## Testing Strategy

### Unit Tests

Run all unit tests to ensure fixes don't break existing functionality:

```bash
# Run all tests
make test

# Run specific test files
go test ./pkg/api-cli/client -run TestParseStepResponse
go test ./pkg/api-cli/client -run TestParseExecutionResponse
go test ./pkg/api-cli/commands -run TestYAMLNormalization
```

### Integration Tests

Create a test script to verify all fixes:

```bash
#!/bin/bash
# test-fixes.sh

echo "Testing ExecuteGoal with numeric ID..."
./bin/api-cli execute-goal 1234 5678

echo "Testing step creation with response parsing..."
./bin/api-cli step-navigate to "https://example.com" 1

echo "Testing YAML loading..."
./bin/api-cli run-test test.yaml

echo "Testing command validation..."
./bin/api-cli step-navigate scroll to top  # Should auto-correct to scroll-to-top
```

### Manual Testing Checklist

- [ ] Execute goal command works with both string and numeric IDs
- [ ] Step creation commands don't fail with "no step ID" errors
- [ ] YAML files load correctly regardless of format
- [ ] Commands with syntax issues are auto-corrected
- [ ] Retry mechanism works for transient failures
- [ ] Pre-commit hooks pass

## Rollback Plan

If issues arise during implementation:

1. **Git branch strategy:**

   ```bash
   git checkout -b fix/api-response-issues
   # Make changes
   # If issues arise:
   git checkout main
   ```

2. **Feature flags:**

   ```go
   // Add to config
   type Config struct {
       // ... existing fields
       UseResponseParser bool `yaml:"use_response_parser"`
       EnableRetry       bool `yaml:"enable_retry"`
   }
   ```

3. **Gradual rollout:**
   - Start with non-critical commands
   - Monitor error rates
   - Expand to all commands once stable

## Monitoring

### Add Logging

```go
// In response_parser.go
if p.debug {
    log.Printf("Response format detected: %s", format.name)
    log.Printf("Extracted ID: %d", id)
}
```

### Metrics Collection

```go
// Add metrics for monitoring
type Metrics struct {
    ResponseParseSuccess int64
    ResponseParseFailed  int64
    RetryAttempts       int64
    CommandValidations  int64
}
```

## Deployment Steps

1. **Test in development:**

   ```bash
   make build
   ./bin/api-cli version
   ```

2. **Run test suite:**

   ```bash
   ./test-commands/test-unified-commands.sh
   ```

3. **Create release:**

   ```bash
   git tag -a v4.2.0 -m "Fix API response parsing and YAML compatibility"
   git push origin v4.2.0
   ```

4. **Update documentation:**
   - Update CLAUDE.md with fix notes
   - Update README.md with any new requirements

## Post-Implementation Tasks

1. **Monitor error logs** for any new issues
2. **Collect metrics** on response parsing success rates
3. **Document any new error patterns** discovered
4. **Update test cases** based on real-world usage

## Success Criteria

- Zero "cannot unmarshal" errors in execute-goal
- 95%+ success rate for step ID extraction
- 100% YAML file compatibility
- Zero command syntax errors after validation
- Retry mechanism reduces transient failures by 80%+

## Support

If you encounter issues during implementation:

1. Check error logs for detailed messages
2. Run with DEBUG=true for verbose output
3. Refer to test files for usage examples
4. Create issues in the repository with reproduction steps
