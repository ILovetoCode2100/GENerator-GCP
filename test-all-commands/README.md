# Comprehensive Test Suite for Virtuoso API CLI

This directory contains comprehensive test files that demonstrate all available commands and their variations in the Virtuoso API CLI.

## Test Files

### 1. `01-assert-commands.yaml`

Tests all 12 assert command variations:

- exists, not-exists, equals, not-equals
- checked, selected, variable
- gt, gte, lt, lte, matches

**Note**: The simplified run-test syntax only supports basic `assert` (for exists).

### 2. `02-interact-commands.yaml`

Tests all 15 interact command variations:

- click, double-click, right-click, hover
- write, key
- mouse operations (move-to, move-by, move, down, up, enter)
- select operations (option, index, last)

**Supported in simplified syntax**: click, hover, write, key, select option

### 3. `03-navigate-data-commands.yaml`

Tests navigation and data commands:

- navigate to URL
- scroll operations (top, bottom, element, position, by, up, down)
- data storage (element-text, value, attribute)
- cookie operations (create, delete, clear)

**Supported in simplified syntax**: navigate, basic scroll, store element-text

### 4. `04-window-dialog-misc.yaml`

Tests window, dialog, and miscellaneous commands:

- window operations (resize, maximize, switch tab/iframe)
- dialog handling (alert, confirm, prompt)
- wait commands (element, time)
- file upload
- misc commands (comment, execute)

**Supported in simplified syntax**: wait, comment, execute

### 5. `05-comprehensive-test.yaml`

A complete end-to-end test that uses all supported commands in the simplified syntax.
This is the best example of what you can do with the run-test command.

### 6. `06-all-step-commands-direct.yaml`

Documentation of all 69 available step commands that can be used directly with the CLI
(not through the simplified run-test syntax).

## Running the Tests

### Quick Test (Dry Run)

```bash
./run-all-tests.sh
```

This will perform a dry run of all tests, showing what would be created without actually making API calls.

### Create a Test

To actually create a test in Virtuoso:

```bash
../bin/api-cli run-test 05-comprehensive-test.yaml
```

### Create with JSON Output

```bash
../bin/api-cli run-test 05-comprehensive-test.yaml -o json
```

## Simplified Syntax Support

The `run-test` command supports a simplified YAML/JSON syntax for the most common operations:

### ✅ Supported Commands

- `navigate`: URL navigation
- `click`: Click elements
- `hover`: Hover over elements
- `write`: Type text into inputs
- `key`: Press keyboard keys
- `select`: Select dropdown options (by text only)
- `assert`: Check element/text exists
- `wait`: Wait for time (ms) or element
- `scroll`: Scroll to element or position (top/bottom)
- `store`: Store element text in variables
- `comment`: Add test comments
- `execute`: Run JavaScript code

### ❌ Not Supported in Simplified Syntax

- Advanced assertions (equals, gt, matches, etc.)
- Mouse operations (move-to, move-by, down/up)
- Window management (resize, maximize, tabs)
- Dialog handling (alerts, confirms, prompts)
- Cookie operations
- File uploads
- Select by index or last
- Advanced scroll operations (by, position, up/down)

## Direct CLI Commands

For full access to all 69 commands, use the direct CLI syntax with checkpoint IDs:

```bash
# Set session context
export VIRTUOSO_SESSION_ID=12345

# Use any command
api-cli step-assert equals "h1" "Welcome"
api-cli step-interact mouse move-to "button"
api-cli step-window resize 1024x768
api-cli step-dialog dismiss-alert
```

## Example: Creating a Complete Test

```bash
# 1. Create the test (this creates project, goal, journey, checkpoint, and steps)
../bin/api-cli run-test 05-comprehensive-test.yaml

# 2. View the output (includes links to Virtuoso UI)
# Output will show:
# - Project ID
# - Goal ID
# - Journey ID
# - Checkpoint ID
# - Step creation results
# - Direct link to view in Virtuoso

# 3. Optional: Use JSON output for scripting
../bin/api-cli run-test 05-comprehensive-test.yaml -o json | jq '.checkpoint_id'
```

## Notes

1. The run-test command automatically creates all required infrastructure
2. You can specify an existing project by name or ID in the YAML
3. Use `--dry-run` to preview what will be created
4. The `config.continue_on_error` option allows tests to continue even if steps fail
5. All commands maintain backward compatibility with the original CLI syntax
