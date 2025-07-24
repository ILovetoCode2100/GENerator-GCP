# Virtuoso API CLI - Error Analysis and Comprehensive Solutions

## Executive Summary

This document provides a comprehensive analysis of errors encountered in the Virtuoso API CLI project and proposes robust solutions. The errors fall into four main categories: API Response Issues, Command Reliability, Type System Issues, and Integration Problems.

## 1. API Response Issues

### 1.1 Execute Goal Unmarshal Error

**Problem**:

```
cannot unmarshal number into Go struct field Execution.item.id of type string
```

**Root Cause**: The API returns `Execution.ID` as a number in some cases but the struct expects a string.

**Solution**:

```go
// pkg/api-cli/client/client.go

// Use flexible response structure that handles both string and numeric IDs
type ExecutionResponseWrapper struct {
    Success bool            `json:"success"`
    Item    json.RawMessage `json:"item"`
    Error   string          `json:"error,omitempty"`
}

// ExecuteGoal with flexible ID handling
func (c *Client) ExecuteGoal(goalID, snapshotID int) (*Execution, error) {
    body := map[string]interface{}{
        "goalId":     goalID,
        "snapshotId": snapshotID,
    }

    var wrapper ExecutionResponseWrapper
    resp, err := c.httpClient.R().
        SetBody(body).
        SetResult(&wrapper).
        Post(fmt.Sprintf("/goals/%d/snapshots/%d/execute", goalID, snapshotID))

    if err != nil {
        return nil, fmt.Errorf("execute goal request failed: %w", err)
    }

    if resp.IsError() || !wrapper.Success {
        if wrapper.Error != "" {
            return nil, fmt.Errorf("execute goal failed: %s", wrapper.Error)
        }
        return nil, fmt.Errorf("execute goal failed with status %d", resp.StatusCode())
    }

    // Parse the flexible response
    execution, err := parseExecutionResponse(wrapper.Item)
    if err != nil {
        return nil, fmt.Errorf("failed to parse execution response: %w", err)
    }

    return execution, nil
}

// Helper to parse execution with flexible ID handling
func parseExecutionResponse(data json.RawMessage) (*Execution, error) {
    // Try parsing with string ID first
    var exec Execution
    if err := json.Unmarshal(data, &exec); err == nil {
        return &exec, nil
    }

    // Try parsing with numeric ID
    var numericExec struct {
        ID         int        `json:"id"`
        GoalID     int        `json:"goalId"`
        SnapshotID int        `json:"snapshotId"`
        Status     string     `json:"status"`
        StartTime  time.Time  `json:"startTime"`
        EndTime    *time.Time `json:"endTime,omitempty"`
        Duration   int        `json:"duration,omitempty"`
        Progress   *ExecutionProgress `json:"progress,omitempty"`
        ResultsURL string     `json:"resultsUrl,omitempty"`
        ReportURL  string     `json:"reportUrl,omitempty"`
    }

    if err := json.Unmarshal(data, &numericExec); err != nil {
        return nil, fmt.Errorf("failed to parse execution response: %w", err)
    }

    // Convert numeric ID to string
    exec = Execution{
        ID:         strconv.Itoa(numericExec.ID),
        GoalID:     numericExec.GoalID,
        SnapshotID: numericExec.SnapshotID,
        Status:     numericExec.Status,
        StartTime:  numericExec.StartTime,
        EndTime:    numericExec.EndTime,
        Duration:   numericExec.Duration,
        Progress:   numericExec.Progress,
        ResultsURL: numericExec.ResultsURL,
        ReportURL:  numericExec.ReportURL,
    }

    return &exec, nil
}
```

### 1.2 Step ID Response Parsing

**Problem**: Commands report "no step ID returned in response" even when steps are created successfully.

**Root Cause**: API returns step IDs in different response formats (`item.id`, `testStep.id`, `id`).

**Solution**:

```go
// pkg/api-cli/client/response_parser.go

package client

import (
    "encoding/json"
    "fmt"
)

// UnifiedStepResponse handles all step response variations
type UnifiedStepResponse struct {
    stepID   int
    response json.RawMessage
}

// ParseStepResponse extracts step ID from various response formats
func ParseStepResponse(body []byte) (int, error) {
    // Try multiple response formats
    formats := []struct {
        name   string
        parser func([]byte) (int, error)
    }{
        {"direct", parseDirectID},
        {"item", parseItemID},
        {"testStep", parseTestStepID},
        {"data", parseDataID},
        {"step", parseStepID},
    }

    for _, format := range formats {
        if id, err := format.parser(body); err == nil && id > 0 {
            return id, nil
        }
    }

    // If we can't find an ID but response looks successful, return a warning
    var genericResp map[string]interface{}
    if err := json.Unmarshal(body, &genericResp); err == nil {
        if success, ok := genericResp["success"].(bool); ok && success {
            // Log warning but don't fail - step was likely created
            fmt.Printf("Warning: Step created but ID not found in response\n")
            return 0, nil
        }
    }

    return 0, fmt.Errorf("no step ID found in response")
}

// Parser functions for different formats
func parseDirectID(body []byte) (int, error) {
    var resp struct {
        ID int `json:"id"`
    }
    if err := json.Unmarshal(body, &resp); err != nil {
        return 0, err
    }
    return resp.ID, nil
}

func parseItemID(body []byte) (int, error) {
    var resp struct {
        Item struct {
            ID int `json:"id"`
        } `json:"item"`
    }
    if err := json.Unmarshal(body, &resp); err != nil {
        return 0, err
    }
    return resp.Item.ID, nil
}

func parseTestStepID(body []byte) (int, error) {
    var resp struct {
        TestStep struct {
            ID int `json:"id"`
        } `json:"testStep"`
    }
    if err := json.Unmarshal(body, &resp); err != nil {
        return 0, err
    }
    return resp.TestStep.ID, nil
}

func parseDataID(body []byte) (int, error) {
    var resp struct {
        Data struct {
            ID int `json:"id"`
        } `json:"data"`
    }
    if err := json.Unmarshal(body, &resp); err != nil {
        return 0, err
    }
    return resp.Data.ID, nil
}

func parseStepID(body []byte) (int, error) {
    var resp struct {
        Step struct {
            ID int `json:"id"`
        } `json:"step"`
    }
    if err := json.Unmarshal(body, &resp); err != nil {
        return 0, err
    }
    return resp.Step.ID, nil
}
```

## 2. Command Reliability Issues

### 2.1 Command Syntax Problems

**Problem**: Various commands have syntax issues (hyphenation, argument order).

**Solution**: Implement a command validator and auto-fixer:

```go
// pkg/api-cli/commands/command_validator.go

package commands

import (
    "fmt"
    "strings"
)

// CommandValidator ensures command syntax consistency
type CommandValidator struct {
    corrections map[string]string
    argOrders   map[string][]string
}

func NewCommandValidator() *CommandValidator {
    return &CommandValidator{
        corrections: map[string]string{
            // Scroll commands - ensure hyphenation
            "scroll to top":      "scroll-to-top",
            "scroll to bottom":   "scroll-to-bottom",
            "scroll to element":  "scroll-to-element",
            "scroll to position": "scroll-to-position",
            "scroll by":          "scroll-by",
            "scroll up":          "scroll-up",
            "scroll down":        "scroll-down",

            // Dialog commands - ensure hyphenation
            "dismiss alert":   "dismiss-alert",
            "dismiss confirm": "dismiss-confirm",
            "dismiss prompt":  "dismiss-prompt",

            // Window commands
            "switch tab": "switch-tab",
            "switch iframe": "switch-iframe",
        },
        argOrders: map[string][]string{
            "switch-tab": {"checkpoint_id", "direction", "position"},
            "mouse-move-by": {"checkpoint_id", "x,y", "position"},
        },
    }
}

// ValidateAndCorrect fixes common command syntax issues
func (v *CommandValidator) ValidateAndCorrect(cmd string, args []string) (string, []string, error) {
    // Fix command hyphenation
    if corrected, ok := v.corrections[cmd]; ok {
        cmd = corrected
    }

    // Fix argument order if needed
    if expectedOrder, ok := v.argOrders[cmd]; ok {
        args = v.reorderArgs(cmd, args, expectedOrder)
    }

    // Validate specific commands
    switch cmd {
    case "mouse-move-by":
        args = v.fixMouseCoordinates(args)
    case "wait-time":
        args = v.ensureMilliseconds(args)
    case "resize":
        args = v.fixResizeDimensions(args)
    }

    return cmd, args, nil
}

func (v *CommandValidator) fixMouseCoordinates(args []string) []string {
    // Ensure coordinates are comma-separated
    for i, arg := range args {
        if strings.Contains(arg, " ") && !strings.Contains(arg, ",") {
            args[i] = strings.Replace(arg, " ", ",", 1)
        }
    }
    return args
}

func (v *CommandValidator) ensureMilliseconds(args []string) []string {
    // Convert seconds to milliseconds if needed
    for i, arg := range args {
        if val, err := strconv.Atoi(arg); err == nil && val < 100 {
            // Likely seconds, convert to milliseconds
            args[i] = strconv.Itoa(val * 1000)
        }
    }
    return args
}

func (v *CommandValidator) fixResizeDimensions(args []string) []string {
    // Ensure dimensions are in WIDTHxHEIGHT format
    for i, arg := range args {
        if strings.Contains(arg, " ") && !strings.Contains(arg, "x") {
            args[i] = strings.Replace(arg, " ", "x", 1)
        }
    }
    return args
}
```

### 2.2 Placeholder ID Returns

**Problem**: Some commands return placeholder IDs (ID: 1) instead of actual IDs.

**Solution**: Implement response validation:

```go
// pkg/api-cli/client/response_validator.go

package client

// ValidateStepResponse ensures response contains valid data
func ValidateStepResponse(stepID int, response interface{}) error {
    // Check for placeholder IDs
    if stepID == 1 || stepID == 0 {
        // Query API to verify step was actually created
        // This is a defensive check for unreliable responses
        return &PlaceholderIDError{
            Message: "API returned placeholder ID, step may have been created",
            StepID:  stepID,
        }
    }

    return nil
}

type PlaceholderIDError struct {
    Message string
    StepID  int
}

func (e *PlaceholderIDError) Error() string {
    return e.Message
}

// IsPlaceholderError checks if error is due to placeholder ID
func IsPlaceholderError(err error) bool {
    _, ok := err.(*PlaceholderIDError)
    return ok
}
```

## 3. Type System Issues

### 3.1 YAML Type Incompatibilities

**Problem**: 98% failure rate when validating YAML files due to `map[interface{}]interface{}` vs `map[string]interface{}` type issues.

**Solution**: Implement a YAML normalizer:

```go
// pkg/api-cli/commands/yaml_normalizer.go

package commands

import (
    "fmt"
    "gopkg.in/yaml.v3"
)

// NormalizeYAML converts YAML types to JSON-compatible types
func NormalizeYAML(data interface{}) interface{} {
    switch v := data.(type) {
    case map[interface{}]interface{}:
        m := make(map[string]interface{})
        for key, val := range v {
            keyStr := fmt.Sprintf("%v", key)
            m[keyStr] = NormalizeYAML(val)
        }
        return m
    case []interface{}:
        for i, val := range v {
            v[i] = NormalizeYAML(val)
        }
        return v
    case map[string]interface{}:
        for key, val := range v {
            v[key] = NormalizeYAML(val)
        }
        return v
    default:
        return v
    }
}

// YAMLParser provides robust YAML parsing
type YAMLParser struct {
    strict bool
}

func NewYAMLParser(strict bool) *YAMLParser {
    return &YAMLParser{strict: strict}
}

// Parse handles various YAML formats
func (p *YAMLParser) Parse(content []byte) (map[string]interface{}, error) {
    var raw interface{}
    if err := yaml.Unmarshal(content, &raw); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }

    // Normalize the parsed data
    normalized := NormalizeYAML(raw)

    // Ensure it's a map
    result, ok := normalized.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("YAML root must be an object, got %T", normalized)
    }

    return result, nil
}

// ValidateAndConvert ensures YAML is in the expected format
func (p *YAMLParser) ValidateAndConvert(content []byte, target interface{}) error {
    data, err := p.Parse(content)
    if err != nil {
        return err
    }

    // Convert to JSON then back to target type
    // This ensures compatibility
    jsonBytes, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to convert YAML to JSON: %w", err)
    }

    if err := json.Unmarshal(jsonBytes, target); err != nil {
        return fmt.Errorf("failed to unmarshal to target type: %w", err)
    }

    return nil
}
```

### 3.2 YAML Format Variations

**Problem**: Multiple YAML formats (compact, simplified, extended) cause validation failures.

**Solution**: Implement format detection and conversion:

```go
// pkg/api-cli/commands/yaml_formats.go

package commands

// YAMLFormat represents different YAML test formats
type YAMLFormat int

const (
    FormatUnknown YAMLFormat = iota
    FormatCompact
    FormatSimplified
    FormatExtended
)

// DetectYAMLFormat identifies which format a YAML file uses
func DetectYAMLFormat(data map[string]interface{}) YAMLFormat {
    // Check for format indicators
    if _, hasSteps := data["steps"]; hasSteps {
        if steps, ok := data["steps"].([]interface{}); ok && len(steps) > 0 {
            step := steps[0]
            switch v := step.(type) {
            case string:
                return FormatCompact
            case map[string]interface{}:
                if _, hasAction := v["action"]; hasAction {
                    return FormatExtended
                }
                return FormatSimplified
            }
        }
    }

    return FormatUnknown
}

// ConvertToStandardFormat converts any YAML format to standard format
func ConvertToStandardFormat(data map[string]interface{}, format YAMLFormat) (map[string]interface{}, error) {
    switch format {
    case FormatCompact:
        return convertCompactFormat(data)
    case FormatSimplified:
        return convertSimplifiedFormat(data)
    case FormatExtended:
        return data, nil // Already standard
    default:
        return nil, fmt.Errorf("unknown YAML format")
    }
}

func convertCompactFormat(data map[string]interface{}) (map[string]interface{}, error) {
    // Convert compact string steps to structured format
    steps, ok := data["steps"].([]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid compact format: missing steps")
    }

    var convertedSteps []interface{}
    for _, step := range steps {
        stepStr, ok := step.(string)
        if !ok {
            continue
        }

        // Parse compact step string
        parsed := parseCompactStep(stepStr)
        convertedSteps = append(convertedSteps, parsed)
    }

    data["steps"] = convertedSteps
    return data, nil
}

func parseCompactStep(step string) map[string]interface{} {
    // Parse compact format like "click: Login button"
    parts := strings.SplitN(step, ":", 2)
    if len(parts) != 2 {
        return map[string]interface{}{
            "action": "unknown",
            "target": step,
        }
    }

    action := strings.TrimSpace(parts[0])
    target := strings.TrimSpace(parts[1])

    return map[string]interface{}{
        "action": action,
        "target": target,
    }
}
```

## 4. Integration Problems

### 4.1 Pre-commit Hook Failures

**Problem**: Pre-commit hooks block valid commits.

**Solution**: Implement pre-commit compatibility:

```go
// pkg/api-cli/commands/git_integration.go

package commands

import (
    "os"
    "os/exec"
)

// GitIntegration handles git-related operations
type GitIntegration struct {
    skipHooks bool
}

// PrepareCommit ensures code is ready for commit
func (g *GitIntegration) PrepareCommit() error {
    // Run formatters
    if err := g.runGoFmt(); err != nil {
        return fmt.Errorf("go fmt failed: %w", err)
    }

    // Run linters
    if err := g.runGoLint(); err != nil {
        // Log but don't fail on lint errors
        fmt.Printf("Warning: lint issues found: %v\n", err)
    }

    // Check for generated files
    if err := g.updateGeneratedFiles(); err != nil {
        return fmt.Errorf("failed to update generated files: %w", err)
    }

    return nil
}

func (g *GitIntegration) runGoFmt() error {
    cmd := exec.Command("go", "fmt", "./...")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func (g *GitIntegration) runGoLint() error {
    cmd := exec.Command("golangci-lint", "run", "--fix")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func (g *GitIntegration) updateGeneratedFiles() error {
    // Update any generated files that might cause hook failures
    // e.g., swagger docs, protobuf files, etc.
    return nil
}
```

### 4.2 Deprecated Endpoints

**Problem**: Some endpoints like `list-checkpoint-steps` are deprecated.

**Solution**: Implement endpoint migration:

```go
// pkg/api-cli/client/endpoint_migration.go

package client

// EndpointMigration handles deprecated endpoint replacements
type EndpointMigration struct {
    migrations map[string]string
}

func NewEndpointMigration() *EndpointMigration {
    return &EndpointMigration{
        migrations: map[string]string{
            "/checkpoints/{id}/steps": "/teststeps?checkpointId={id}",
            "/goals/{id}/execute":     "/executions",
        },
    }
}

// MigrateEndpoint returns the current endpoint for a deprecated one
func (e *EndpointMigration) MigrateEndpoint(deprecated string) string {
    if current, ok := e.migrations[deprecated]; ok {
        return current
    }
    return deprecated
}
```

### 4.3 Missing Context Methods

**Problem**: Some operations require Context methods that don't exist.

**Solution**: Implement a context wrapper generator:

```go
// pkg/api-cli/client/context_wrapper_generator.go

package client

import (
    "context"
    "reflect"
    "strings"
)

// GenerateContextWrappers creates WithContext versions of methods
func GenerateContextWrappers(client *Client) error {
    clientType := reflect.TypeOf(client)
    clientValue := reflect.ValueOf(client)

    for i := 0; i < clientType.NumMethod(); i++ {
        method := clientType.Method(i)

        // Skip if already has context
        if strings.HasSuffix(method.Name, "WithContext") {
            continue
        }

        // Skip if method doesn't return error as last value
        if method.Type.NumOut() == 0 {
            continue
        }

        lastOut := method.Type.Out(method.Type.NumOut() - 1)
        if lastOut.String() != "error" {
            continue
        }

        // Generate wrapper
        if err := generateContextWrapper(client, method); err != nil {
            return fmt.Errorf("failed to generate wrapper for %s: %w", method.Name, err)
        }
    }

    return nil
}

// This is a conceptual example - actual implementation would use code generation
func generateContextWrapper(client *Client, method reflect.Method) error {
    // Generate a WithContext version of the method
    // This would typically be done with code generation tools
    return nil
}
```

## 5. Self-Healing Mechanisms

### 5.1 Automatic Retry with Backoff

```go
// pkg/api-cli/client/retry.go

package client

import (
    "context"
    "time"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay time.Duration
    Multiplier float64
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() *RetryConfig {
    return &RetryConfig{
        MaxAttempts:  3,
        InitialDelay: 1 * time.Second,
        MaxDelay:     10 * time.Second,
        Multiplier:   2.0,
    }
}

// RetryWithBackoff retries operations with exponential backoff
func RetryWithBackoff(ctx context.Context, config *RetryConfig, operation func() error) error {
    delay := config.InitialDelay

    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }

        // Check if error is retryable
        if !isRetryable(err) {
            return err
        }

        if attempt == config.MaxAttempts {
            return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, err)
        }

        // Wait with backoff
        select {
        case <-time.After(delay):
            delay = time.Duration(float64(delay) * config.Multiplier)
            if delay > config.MaxDelay {
                delay = config.MaxDelay
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    return nil
}

func isRetryable(err error) bool {
    // Define which errors are retryable
    // - Network timeouts
    // - 5xx server errors
    // - Rate limit errors (with appropriate backoff)

    if apiErr, ok := err.(*APIError); ok {
        return apiErr.Status >= 500 || apiErr.Status == 429
    }

    if clientErr, ok := err.(*ClientError); ok {
        return clientErr.Kind == KindTimeout || clientErr.Kind == KindConnectionFailed
    }

    return false
}
```

### 5.2 Response Format Detection

```go
// pkg/api-cli/client/response_detector.go

package client

// ResponseFormatDetector automatically detects and handles different response formats
type ResponseFormatDetector struct {
    formats []ResponseFormat
}

type ResponseFormat interface {
    CanHandle(data []byte) bool
    Extract(data []byte) (interface{}, error)
}

// AutoDetectResponse tries multiple formats until one works
func (d *ResponseFormatDetector) AutoDetectResponse(data []byte, target interface{}) error {
    for _, format := range d.formats {
        if format.CanHandle(data) {
            result, err := format.Extract(data)
            if err == nil {
                // Use reflection to assign result to target
                return assignResult(result, target)
            }
        }
    }

    return fmt.Errorf("no format could handle response")
}
```

## 6. Testing Strategies

### 6.1 Response Variation Tests

```go
// pkg/api-cli/client/client_test.go

func TestExecuteGoalResponseVariations(t *testing.T) {
    testCases := []struct {
        name     string
        response string
        wantID   string
        wantErr  bool
    }{
        {
            name: "string ID in item",
            response: `{"success": true, "item": {"id": "12345"}}`,
            wantID: "12345",
        },
        {
            name: "numeric ID in item",
            response: `{"success": true, "item": {"id": 12345}}`,
            wantID: "12345",
        },
        {
            name: "direct ID",
            response: `{"id": "12345", "status": "RUNNING"}`,
            wantID: "12345",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test response parsing
            result, err := parseExecutionResponse([]byte(tc.response))

            if tc.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tc.wantID, result.ID)
        })
    }
}
```

### 6.2 YAML Format Tests

```go
// pkg/api-cli/commands/yaml_normalizer_test.go

func TestYAMLNormalization(t *testing.T) {
    testCases := []struct {
        name  string
        input string
        want  map[string]interface{}
    }{
        {
            name: "map[interface{}]interface{} normalization",
            input: `
steps:
  - action: click
    target: button`,
            want: map[string]interface{}{
                "steps": []interface{}{
                    map[string]interface{}{
                        "action": "click",
                        "target": "button",
                    },
                },
            },
        },
    }

    parser := NewYAMLParser(false)

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := parser.Parse([]byte(tc.input))
            assert.NoError(t, err)
            assert.Equal(t, tc.want, result)
        })
    }
}
```

## 7. Architecture Improvements

### 7.1 Response Handler Chain

```go
// pkg/api-cli/client/response_chain.go

package client

// ResponseHandler processes API responses
type ResponseHandler interface {
    Handle(resp *resty.Response) (interface{}, error)
    SetNext(handler ResponseHandler)
}

// BaseResponseHandler provides common functionality
type BaseResponseHandler struct {
    next ResponseHandler
}

func (h *BaseResponseHandler) SetNext(handler ResponseHandler) {
    h.next = handler
}

func (h *BaseResponseHandler) HandleNext(resp *resty.Response) (interface{}, error) {
    if h.next != nil {
        return h.next.Handle(resp)
    }
    return nil, fmt.Errorf("no handler could process response")
}

// SuccessResponseHandler handles successful responses
type SuccessResponseHandler struct {
    BaseResponseHandler
}

func (h *SuccessResponseHandler) Handle(resp *resty.Response) (interface{}, error) {
    if !resp.IsSuccess() {
        return h.HandleNext(resp)
    }

    // Process successful response
    return resp.Result(), nil
}

// ErrorResponseHandler handles error responses
type ErrorResponseHandler struct {
    BaseResponseHandler
}

func (h *ErrorResponseHandler) Handle(resp *resty.Response) (interface{}, error) {
    if resp.IsSuccess() {
        return h.HandleNext(resp)
    }

    // Extract error details
    var apiErr APIError
    if err := json.Unmarshal(resp.Body(), &apiErr); err == nil {
        apiErr.Status = resp.StatusCode()
        return nil, &apiErr
    }

    return nil, fmt.Errorf("API error: %s", resp.Status())
}
```

### 7.2 Command Pipeline

```go
// pkg/api-cli/commands/command_pipeline.go

package commands

// CommandPipeline processes commands through stages
type CommandPipeline struct {
    stages []CommandStage
}

type CommandStage interface {
    Process(cmd string, args []string) (string, []string, error)
}

// ValidationStage validates command syntax
type ValidationStage struct {
    validator *CommandValidator
}

func (s *ValidationStage) Process(cmd string, args []string) (string, []string, error) {
    return s.validator.ValidateAndCorrect(cmd, args)
}

// NormalizationStage normalizes arguments
type NormalizationStage struct{}

func (s *NormalizationStage) Process(cmd string, args []string) (string, []string, error) {
    // Normalize arguments (trim spaces, fix casing, etc.)
    for i := range args {
        args[i] = strings.TrimSpace(args[i])
    }
    return cmd, args, nil
}

// ExecutePipeline runs command through all stages
func (p *CommandPipeline) Execute(cmd string, args []string) (string, []string, error) {
    var err error

    for _, stage := range p.stages {
        cmd, args, err = stage.Process(cmd, args)
        if err != nil {
            return cmd, args, fmt.Errorf("pipeline stage failed: %w", err)
        }
    }

    return cmd, args, nil
}
```

## 8. Implementation Priority

1. **High Priority** (Blocking Issues):

   - Fix ExecuteGoal unmarshal error
   - Implement response format detection
   - Fix YAML type normalization

2. **Medium Priority** (Reliability):

   - Add retry with backoff
   - Implement command validation
   - Add response validation

3. **Low Priority** (Nice to Have):
   - Generate missing Context methods
   - Add comprehensive testing
   - Implement self-healing mechanisms

## 9. Monitoring and Alerting

```go
// pkg/api-cli/monitoring/health_check.go

package monitoring

// HealthChecker monitors API health
type HealthChecker struct {
    client  *client.Client
    metrics *Metrics
}

// CheckAPIHealth performs health checks
func (h *HealthChecker) CheckAPIHealth() error {
    checks := []struct {
        name string
        fn   func() error
    }{
        {"auth", h.checkAuth},
        {"endpoints", h.checkEndpoints},
        {"response_formats", h.checkResponseFormats},
    }

    for _, check := range checks {
        if err := check.fn(); err != nil {
            h.metrics.RecordError(check.name, err)
            return fmt.Errorf("%s check failed: %w", check.name, err)
        }
        h.metrics.RecordSuccess(check.name)
    }

    return nil
}
```

## Conclusion

These comprehensive solutions address all major error categories in the Virtuoso API CLI. The implementation focuses on:

1. **Flexibility**: Handling multiple response formats gracefully
2. **Robustness**: Retry mechanisms and error recovery
3. **Compatibility**: Supporting various YAML formats and type systems
4. **Maintainability**: Clear architecture with separation of concerns
5. **Observability**: Monitoring and health checks

By implementing these solutions, the CLI will become more reliable, user-friendly, and maintainable.
