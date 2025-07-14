# Assert Command Consolidation

## Overview

The assert command consolidation is a proof of concept that demonstrates how we can simplify the CLI by grouping related commands under a single parent command with subcommands.

## Before: Individual Commands

Previously, we had 12 separate assertion commands:

- `api-cli create-step-assert-exists`
- `api-cli create-step-assert-not-exists`
- `api-cli create-step-assert-equals`
- `api-cli create-step-assert-not-equals`
- `api-cli create-step-assert-checked`
- `api-cli create-step-assert-selected`
- `api-cli create-step-assert-greater-than`
- `api-cli create-step-assert-greater-than-or-equal`
- `api-cli create-step-assert-less-than`
- `api-cli create-step-assert-less-than-or-equal`
- `api-cli create-step-assert-matches`
- `api-cli create-step-assert-variable`

## After: Consolidated Command

Now we have a single `assert` command with subcommands:

- `api-cli assert exists`
- `api-cli assert not-exists`
- `api-cli assert equals`
- `api-cli assert not-equals`
- `api-cli assert checked`
- `api-cli assert selected`
- `api-cli assert gt`
- `api-cli assert gte`
- `api-cli assert lt`
- `api-cli assert lte`
- `api-cli assert matches`
- `api-cli assert variable`

## Benefits

1. **Cleaner Command Structure**: Related commands are grouped together
2. **Better Discoverability**: `api-cli assert --help` shows all assertion options
3. **Shared Logic**: Common functionality is extracted and reused
4. **Maintainability**: Less code duplication, easier to maintain
5. **Consistency**: All assertion commands follow the same patterns

## Implementation Details

### Code Structure

The consolidation is implemented in `pkg/api-cli/commands/assert.go` with:

1. **Type Definitions**:

   - `assertType` enum for different assertion types
   - `assertCommandInfo` struct containing metadata for each type

2. **Shared Logic**:

   - `resolveStepContext()` - Handles checkpoint and position resolution
   - `outputStepResult()` - Formats output in various formats
   - `validateAssertArgs()` - Validates arguments for each assertion type
   - `callAssertAPI()` - Routes to appropriate client API method

3. **Dynamic Command Generation**:
   - Commands are generated from the `assertCommands` map
   - Each subcommand shares the same command structure and validation

### Usage Examples

```bash
# With explicit checkpoint and position
api-cli assert exists "Login button" 1 --checkpoint 1680449

# Using session context and auto-increment
api-cli set-checkpoint 1680449
api-cli assert exists "Login button"
api-cli assert equals "Username" "john@example.com"
api-cli assert gt "Price" "10"

# Different output formats
api-cli assert exists "Dashboard" --output json
api-cli assert matches "Email" "^[\\w.-]+@[\\w.-]+\\.\\w+$" --output yaml
```

## Extending the Pattern

This consolidation pattern can be applied to other command groups:

1. **Mouse Commands**:

   - `api-cli mouse move-to X Y`
   - `api-cli mouse move-by DX DY`
   - `api-cli mouse click`
   - `api-cli mouse double-click`
   - `api-cli mouse right-click`

2. **Wait Commands**:

   - `api-cli wait element SELECTOR`
   - `api-cli wait time SECONDS`
   - `api-cli wait for-element SELECTOR TIMEOUT`

3. **Cookie Commands**:

   - `api-cli cookie create NAME VALUE`
   - `api-cli cookie delete NAME`
   - `api-cli cookie wipe-all`

4. **Scroll Commands**:
   - `api-cli scroll to-top`
   - `api-cli scroll to-bottom`
   - `api-cli scroll to-position X Y`
   - `api-cli scroll by-offset DX DY`
   - `api-cli scroll to-element SELECTOR`

## Testing

The consolidated assert command maintains 100% compatibility with the original commands:

- All API calls remain the same
- Output formats are preserved
- Session context and auto-increment work identically
- All flags and options are supported

Run the test script to verify functionality:

```bash
./test-consolidated-assert.sh
```

## Migration Path

For backward compatibility during migration:

1. The old commands are commented out but can be re-enabled if needed
2. Both command styles could coexist temporarily
3. Scripts using old commands can be updated gradually
4. A migration script could automate the command translation
