# Virtuoso API CLI - Stage 1 Implementation Task

## Context

You are working on the Virtuoso API CLI, a Go-based command-line tool that provides an AI-friendly interface for Virtuoso's test automation platform. The project is at version 2.0 with 98% test success rate and uses a consolidated command structure with 12 command groups.

## Project Structure

```
virtuoso-GENerator/
├── cmd/api-cli/           # Main entry point
├── pkg/api-cli/
│   ├── client/           # API client (40+ methods)
│   ├── commands/         # 12 command groups
│   └── config/           # Configuration management
├── bin/                  # Compiled binary
└── test-all-cli-commands.sh  # Test framework
```

## Stage 1 Implementation Requirements

Implement the following missing commands:

### 1. Navigate Commands (3 missing)

- `navigate back` - Browser back button
- `navigate forward` - Browser forward button
- `navigate refresh` - Refresh/reload page

### 2. Window Commands (3 missing)

- `window maximize` - Maximize browser window
- `window close` - Close current window/tab
- `window switch <index>` - Switch to tab by index (0-based)

### 3. Data Commands (1 missing)

- `data store attribute <element> <attribute> <variable>` - Store element attribute value

## Implementation Guide

### Step 1: Add Client Methods

In `pkg/api-cli/client/client.go`, add these methods:

```go
// Navigate methods
func (c *Client) CreateStepNavigateBack(checkpointID int, position int) (int, error)
func (c *Client) CreateStepNavigateForward(checkpointID int, position int) (int, error)
func (c *Client) CreateStepNavigateRefresh(checkpointID int, position int) (int, error)

// Window methods
func (c *Client) CreateStepWindowMaximize(checkpointID int, position int) (int, error)
func (c *Client) CreateStepWindowClose(checkpointID int, position int) (int, error)
func (c *Client) CreateStepSwitchTabByIndex(checkpointID int, index int, position int) (int, error)

// Data method
func (c *Client) CreateStepStoreAttribute(checkpointID int, selector string, attribute string, variable string, position int) (int, error)
```

### Step 2: Update Navigate Commands

In `pkg/api-cli/commands/navigate.go`:

1. Add subcommands to NavigateCmd():

```go
cmd.AddCommand(navigateBackSubCmd())
cmd.AddCommand(navigateForwardSubCmd())
cmd.AddCommand(navigateRefreshSubCmd())
```

2. Create the subcommand functions following the existing pattern (see navigateToSubCmd for reference)

3. Add cases to runNavigation() switch statement

4. Create execute functions following the pattern of executeNavigateToAction()

### Step 3: Update Window Commands

In `pkg/api-cli/commands/window.go`:

1. Add new window operations to the constants:

```go
windowMaximize  windowOperation = "maximize"
windowClose     windowOperation = "close"
windowSwitchTab windowOperation = "switch-tab"
```

2. Add entries to windowCommands map with proper metadata

3. Add subcommands in newWindowCmd()

4. Update callWindowAPI() to handle the new operations

### Step 4: Update Data Commands

In `pkg/api-cli/commands/data.go`:

1. Add new data type constant:

```go
dataStoreAttribute dataType = "store-attribute"
```

2. Add entry to dataCommands map with metadata

3. Add subcommand to the store command group

4. Update callDataAPI() to handle the new operation

### Step 5: API Request Format

All API requests should follow this pattern:

```go
step := map[string]interface{}{
    "type": "STEP_TYPE_HERE",
    "position": position,
    "meta": map[string]interface{}{
        // Step-specific metadata
    },
}
```

Expected step types:

- Navigate: "NAVIGATE_BACK", "NAVIGATE_FORWARD", "NAVIGATE_REFRESH"
- Window: "WINDOW_MAXIMIZE", "WINDOW_CLOSE", "SWITCH_TAB"
- Data: "STORE"

### Step 6: Testing

Update `test-all-cli-commands.sh` to test the new commands:

```bash
# Navigate commands (lines 91-95)
test_cmd "Navigate back" "$CLI navigate back $CHECKPOINT_ID $((POS++))"
test_cmd "Navigate forward" "$CLI navigate forward $CHECKPOINT_ID $((POS++))"
test_cmd "Navigate refresh" "$CLI navigate refresh $CHECKPOINT_ID $((POS++))"

# Window commands (add new section)
test_cmd "Window maximize" "$CLI window maximize $CHECKPOINT_ID $((POS++))"
test_cmd "Window close" "$CLI window close $CHECKPOINT_ID $((POS++))"
test_cmd "Window switch tab" "$CLI window switch 1 $CHECKPOINT_ID $((POS++))"

# Data command (add to data section)
test_cmd "Store attribute" "$CLI data store attribute $CHECKPOINT_ID '#element' 'href' 'link_url' $((POS++))"
```

## Important Patterns to Follow

1. **BaseCommand Usage**: All commands use the BaseCommand infrastructure from `pkg/api-cli/commands/base.go`

2. **Position Management**: Support both explicit position and auto-increment via session context

3. **Output Formats**: Support all formats (human, json, yaml, ai)

4. **Error Handling**: Use structured errors with helpful messages

5. **Argument Parsing**: Follow the pattern of ResolveCheckpointAndPosition()

## Build and Test

```bash
# Build
make build

# Test individual commands
./bin/api-cli navigate back --dry-run
./bin/api-cli window maximize --dry-run
./bin/api-cli data store attribute "button" "class" "btn_class" --dry-run

# Run full test suite
./test-all-cli-commands.sh
```

## Success Criteria

- All 7 commands implemented and working
- Following existing code patterns consistently
- Tests passing in test-all-cli-commands.sh
- Proper error handling and validation
- Support for all output formats
- Session context support with auto-increment

## Notes

- The client methods need to make POST requests to `/api/steps` endpoint
- Use existing client methods as reference (e.g., CreateStepNavigate, CreateStepWindowResize)
- Maintain backward compatibility with explicit checkpoint/position args
- Follow Go conventions and existing code style
