# Git History Analysis - Failed Commands

## Summary of Findings

After reviewing the git history and codebase, I've identified the root causes and previous fixes for the failing commands:

### 1. Mouse Commands (move-by, move, down, up)

**Previous Fix (commit a03c62c):**

- Fixed mouse move-by command with negative numbers
- Issue: Cobra interprets negative numbers as flags
- Solution: Updated documentation to use `--` separator
- Example: `api-cli mouse move-by -- 50 -30`

**Current Implementation:**

- Commands are properly implemented in `step_interact.go`
- `parseCoordinates()` function handles x,y coordinate parsing
- Client methods exist: `CreateStepMouseMoveBy`, `CreateStepMouseMoveTo`, `CreateStepMouseDown`, `CreateStepMouseUp`

**Likely Issue:**

- The test script may not be using the `--` separator for negative coordinates
- Coordinate format validation may be too strict

### 2. Keyboard with Modifiers (--ctrl)

**Current Implementation:**

- Command properly accepts `--modifiers` flag in `step_interact.go`
- Client has methods: `CreateStepKeyGlobalWithModifiers`, `CreateStepKeyTargetedWithModifiers`
- TODO comments indicate context support is not yet implemented for modifier methods

**Likely Issue:**

- The methods with modifiers don't have context support, which may cause issues with the new error handling
- API may require specific format for modifiers array

### 3. Select by Index

**Current Implementation:**

- Command is properly implemented in `step_interact.go`
- Validates index is a non-negative integer
- Client method: `CreateStepPickIndex` exists and formats index as string

**Likely Issue:**

- The API expects the index value in a specific format (as seen in client: `"value": fmt.Sprintf("%d", index)`)
- May be an API response parsing issue

### 4. Store element-value

**Root Cause - Missing Implementation:**

- The `element-value` operation is NOT implemented in `dataConfigs` map in `step_data.go`
- Only `element-text`, `literal`, and `attribute` are implemented
- Documentation and tests reference it, but the actual command config is missing

**Fix Required:**

- Add `"store.element-value"` configuration to `dataConfigs` map
- The client method `CreateStepStoreValue` already exists

### 5. Cookie clear

**Root Cause - Command Name Mismatch:**

- The command is implemented as `cookie clear-all` (with hyphen)
- Documentation and tests reference it as `cookie clear` (without hyphen)
- The actual implementation uses `clear-all`

**Fix Required:**

- Either update tests/docs to use `clear-all` or add alias for `clear`

## Key Insights from Git History

### From commit 315d796 (Remove unsupported API commands):

- Removed several unsupported commands that the API doesn't handle
- This commit shows pattern of API limitations requiring command removal

### From commit 6cbc1f4 (Context support implementation):

- Added comprehensive context support with 80+ context-aware methods
- Some methods (like keyboard with modifiers) still have TODO comments for context support

### From CLAUDE.md Recent Changes:

- Mouse commands: "Fixed coordinate parsing for move-by and move operations"
- Wait time: "Fixed argument parsing and milliseconds handling"
- Dialog commands: "Updated to hyphenated syntax"

## Recommendations

1. **Mouse Commands**: Ensure test scripts use `--` for negative numbers
2. **Keyboard Modifiers**: Implement context support for modifier methods
3. **Select Index**: Check API response handling
4. **Store element-value**: Add missing configuration to `dataConfigs`
5. **Cookie clear**: Use `clear-all` or add alias support

## Previous Patterns

The codebase shows a pattern of:

- API limitations requiring workarounds or command removal
- Command syntax mismatches between documentation and implementation
- Missing context support for some specialized methods
- Hyphenated vs non-hyphenated command naming inconsistencies
