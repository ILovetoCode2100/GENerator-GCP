# Stage 2 Enhanced Functionality - Implementation Prompt

## Context

You are implementing Stage 2 enhancements to the Virtuoso API CLI, a Go-based command-line tool for test automation. The codebase is located at `/Users/marklovelady/_dev/_projects/virtuoso-GENerator/`.

Stage 1 has already been completed and added:

- navigate back/forward/refresh
- window maximize/close/switch <index>
- data store attribute

## Stage 2 Requirements

### 1. Click with Position Enum Support

**Current State**: Click accepts a position string but doesn't validate against enum values
**Enhancement**: Add validation for position enum values and update the client method

**Files to Modify**:

- `/pkg/api-cli/commands/interact.go` - Update `executeClickAction` to validate position enum
- `/pkg/api-cli/client/client.go` - Update `CreateStepClickWithDetails` to handle position enums properly

**Position Enum Values**: TOP_LEFT, TOP_CENTER, TOP_RIGHT, CENTER_LEFT, CENTER, CENTER_RIGHT, BOTTOM_LEFT, BOTTOM_CENTER, BOTTOM_RIGHT

**Implementation**:

```go
// In interact.go, add validation function:
func isValidClickPosition(position string) bool {
    validPositions := []string{
        "TOP_LEFT", "TOP_CENTER", "TOP_RIGHT",
        "CENTER_LEFT", "CENTER", "CENTER_RIGHT",
        "BOTTOM_LEFT", "BOTTOM_CENTER", "BOTTOM_RIGHT",
    }
    for _, valid := range validPositions {
        if position == valid {
            return true
        }
    }
    return false
}

// Update executeClickAction to validate:
if positionType != "" && !isValidClickPosition(positionType) {
    return 0, fmt.Errorf("invalid position: %s. Valid positions: TOP_LEFT, TOP_CENTER, TOP_RIGHT, CENTER_LEFT, CENTER, CENTER_RIGHT, BOTTOM_LEFT, BOTTOM_CENTER, BOTTOM_RIGHT", positionType)
}
```

### 2. Wait Variations

**Current State**: Basic wait for element and wait time
**Enhancement**: Add "wait for element not visible" and enhance timeout handling

**Files to Modify**:

- `/pkg/api-cli/commands/wait.go` - Add new wait type for "element-not-visible"
- `/pkg/api-cli/client/client.go` - Add `CreateStepWaitForElementNotVisible` method

**Implementation**:

```go
// In wait.go, add to waitCommands map:
waitElementNotVisible: {
    stepType:    "WAIT_ELEMENT_NOT_VISIBLE",
    description: "Wait for an element to disappear",
    usage:       "wait element-not-visible SELECTOR [POSITION]",
    examples: []string{
        `api-cli wait element-not-visible "Loading spinner" 1`,
        `api-cli wait element-not-visible "#loader" --timeout 5000`,
    },
    argsCount: []int{1},
    parseStep: func(args []string, timeout int) string {
        if timeout > 0 {
            return fmt.Sprintf("wait until %s disappears (timeout: %dms)", args[0], timeout)
        }
        return fmt.Sprintf("wait until %s disappears", args[0])
    },
    hasTimeout: true,
}

// In client.go, add:
func (c *Client) CreateStepWaitForElementNotVisible(checkpointID int, selector string, timeoutMs int, position int) (int, error) {
    clueJSON := fmt.Sprintf(`{"clue":"%s"}`, selector)

    parsedStep := map[string]interface{}{
        "action": "WAIT",
        "meta": map[string]interface{}{
            "isElement":       true,
            "type":           "ELEMENT",
            "waitForInvisible": true,
        },
        "clue": clueJSON,
    }

    if timeoutMs > 0 {
        parsedStep["meta"].(map[string]interface{})["timeout"] = timeoutMs
    }

    return c.createTestStep(checkpointID, parsedStep, position)
}
```

### 3. Scroll by Relative Offset

**Current State**: Scroll to absolute position only
**Enhancement**: Add scroll by relative X,Y offset

**Files to Modify**:

- `/pkg/api-cli/commands/navigate.go` - Add new subcommand "scroll-by"
- `/pkg/api-cli/client/client.go` - Enhance `CreateStepScrollByOffset` if needed

**Implementation**:

```go
// In navigate.go, add new subcommand:
func scrollBySubCmd() *cobra.Command {
    var (
        x      int
        y      int
        smooth bool
    )

    cmd := &cobra.Command{
        Use:   "scroll-by [checkpoint-id] <x,y> [position]",
        Short: "Scroll by relative offset",
        Long: `Scroll the page by a relative X,Y offset from current position.

Examples:
  # Using session context
  api-cli navigate scroll-by "0,500"    # Scroll down 500px
  api-cli navigate scroll-by "-100,0"  # Scroll left 100px
  api-cli navigate scroll-by --x 0 --y -500  # Scroll up 500px

  # Using explicit checkpoint
  api-cli navigate scroll-by cp_12345 "0,500" 1`,
        Aliases: []string{"by", "offset"},
        Args:    cobra.RangeArgs(0, 3),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Parse coordinates logic similar to scroll-position
            return runNavigation(cmd, args, "scroll-by", map[string]interface{}{
                "x":      x,
                "y":      y,
                "smooth": smooth,
            })
        },
    }

    cmd.Flags().IntVar(&x, "x", 0, "X offset (negative for left)")
    cmd.Flags().IntVar(&y, "y", 0, "Y offset (negative for up)")
    cmd.Flags().BoolVar(&smooth, "smooth", false, "Use smooth scrolling")

    return cmd
}

// Add to NavigateCmd():
cmd.AddCommand(scrollBySubCmd())

// Add case in runNavigation:
case "scroll-by":
    stepID, err = executeScrollByAction(base.Client, checkpointID, base.Position, options)

// Add execute function:
func executeScrollByAction(c *client.Client, checkpointID int, position int, options map[string]interface{}) (int, error) {
    x, _ := options["x"].(int)
    y, _ := options["y"].(int)

    return c.CreateStepScrollByOffset(checkpointID, x, y, position)
}
```

### 4. Cookie Operations with Domain

**Current State**: Basic cookie create without domain/path options
**Enhancement**: Update client to pass domain/path/secure/httpOnly to API

**Files to Modify**:

- `/pkg/api-cli/client/client.go` - Enhance `CreateStepCookieCreate` to accept options

**Implementation**:

```go
// In client.go, update CreateStepCookieCreate:
func (c *Client) CreateStepCookieCreateWithOptions(checkpointID int, name, value string, options map[string]interface{}, position int) (int, error) {
    meta := map[string]interface{}{
        "envOption": "CREATE",
        "cookieName": name,
        "cookieValue": value,
    }

    // Add optional fields if provided
    if domain, ok := options["domain"].(string); ok && domain != "" {
        meta["cookieDomain"] = domain
    }
    if path, ok := options["path"].(string); ok && path != "" {
        meta["cookiePath"] = path
    }
    if secure, ok := options["secure"].(bool); ok && secure {
        meta["cookieSecure"] = true
    }
    if httpOnly, ok := options["httpOnly"].(bool); ok && httpOnly {
        meta["cookieHttpOnly"] = true
    }

    parsedStep := map[string]interface{}{
        "action": "ENVIRONMENT",
        "meta":   meta,
    }

    return c.createTestStep(checkpointID, parsedStep, position)
}

// In data.go, update callDataAPI for dataCookieCreate:
case dataCookieCreate:
    if flags != nil && (flags["domain"] != nil || flags["path"] != nil || flags["secure"] != nil || flags["http-only"] != nil) {
        options := map[string]interface{}{
            "domain":   flags["domain"],
            "path":     flags["path"],
            "secure":   flags["secure"],
            "httpOnly": flags["http-only"],
        }
        return apiClient.CreateStepCookieCreateWithOptions(ctx.CheckpointID, args[0], args[1], options, ctx.Position)
    }
    return apiClient.CreateStepCookieCreate(ctx.CheckpointID, args[0], args[1], ctx.Position)
```

### 5. Store Operations for Element Attributes

**Current State**: Already implemented in Stage 1 as `data store attribute`
**Action**: Verify implementation is complete and working

## Testing Requirements

1. **Unit Tests**: Add tests for new validation functions
2. **Integration Tests**: Update `test-consolidated-commands-final.sh` to include:

   ```bash
   # Test click with position enum
   ./bin/api-cli interact click "#button" --position TOP_LEFT --dry-run

   # Test wait element not visible
   ./bin/api-cli wait element-not-visible "#loader" --timeout 3000 --dry-run

   # Test scroll by offset
   ./bin/api-cli navigate scroll-by "0,500" --dry-run
   ./bin/api-cli navigate scroll-by --x -100 --y 200 --dry-run

   # Test cookie with domain
   ./bin/api-cli data cookie create "session" "abc123" --domain ".example.com" --secure --dry-run
   ```

## Implementation Order

1. **Click Position Enum** (simplest, validation only)
2. **Scroll By Offset** (new subcommand, uses existing client method)
3. **Wait Element Not Visible** (new wait type, new client method)
4. **Cookie Domain Options** (enhance existing, new client method)

## Backward Compatibility

- Keep existing methods unchanged
- Add new methods alongside existing ones
- Use feature detection in commands (check if enhanced method exists)
- Default to basic behavior if enhanced options not provided

## Key Patterns to Follow

1. Use `BaseCommand` for shared functionality
2. Support all output formats (human, json, yaml, ai)
3. Include proper error messages
4. Add position management for test steps
5. Follow existing code style and patterns

## Build and Test

```bash
# Build the project
make build

# Run tests
make test

# Test individual enhancements
./bin/api-cli interact click "#test" --position TOP_LEFT --dry-run
./bin/api-cli wait element-not-visible "#loader" --dry-run
./bin/api-cli navigate scroll-by "0,500" --dry-run
./bin/api-cli data cookie create "test" "value" --domain ".test.com" --dry-run
```

## Success Criteria

- All new commands work with --dry-run flag
- Position enum validation prevents invalid values
- Cookie domain/path/secure/httpOnly passed to API
- Scroll by offset works with negative values
- Wait element not visible has timeout support
- All existing commands continue to work unchanged
- Test success rate remains at 98%+
