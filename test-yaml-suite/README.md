# Virtuoso API CLI - Comprehensive YAML Test Suite

## Overview

This comprehensive test suite validates all 69 commands of the Virtuoso API CLI through structured YAML test files. The suite includes positive tests, edge cases, complex workflows, and error scenarios.

## Test Suite Structure

```
test-yaml-suite/
├── README.md                    # This file
├── run-tests.sh                 # Main test runner script
├── config/
│   └── test-config.yaml        # Global configuration and test data
├── commands/                    # Command-specific tests
│   ├── step-navigate/          # Navigation and scroll tests
│   ├── step-interact/          # Interaction tests (click, write, key, mouse, select)
│   ├── step-assert/            # Assertion tests
│   ├── step-data/              # Data storage and cookie tests
│   ├── step-window/            # Window management tests
│   ├── step-dialog/            # Dialog handling tests
│   ├── step-wait/              # Wait operation tests
│   ├── step-file/              # File upload tests
│   └── step-misc/              # Comment and JavaScript tests
├── workflows/
│   └── e2e-scenarios/          # End-to-end workflow tests
└── error-scenarios/            # Error and negative test cases
```

## Test Coverage

### Command Tests (All 69 Commands)

#### Navigation (10 commands)

- `navigate to` - URL navigation with new tab support
- `navigate scroll-top/bottom` - Page scroll extremes
- `navigate scroll-element` - Scroll to specific elements
- `navigate scroll-position` - Absolute positioning
- `navigate scroll-by` - Relative scrolling
- `navigate scroll-up/down` - Directional scrolling

#### Interaction (21 commands)

- **Click**: click, double-click, right-click, hover
- **Text**: write with clear/delay options
- **Keyboard**: key press with modifiers (ctrl, shift, alt, meta)
- **Mouse**: move-to, move-by, move, down, up, enter
- **Select**: by option text, by index, select last

#### Assertions (12 commands)

- Existence: exists, not-exists
- Equality: equals, not-equals
- State: checked, selected
- Variables: variable assertions
- Comparisons: gt, gte, lt, lte
- Pattern: matches (regex)

#### Data Management (10 commands)

- Store: element-text, literal, element-value, attribute
- Cookies: create, delete, clear-all

#### Window/Dialog (10 commands)

- Window: resize, maximize
- Switch: tab navigation, iframe switching
- Dialogs: alert, confirm, prompt handling

#### Other (6 commands)

- Wait: for element, for time
- File: upload via URL
- Misc: comment, execute JavaScript

### Test Categories

1. **Positive Tests** (`positive/`)

   - Valid inputs for all commands
   - Expected success scenarios
   - Common use cases

2. **Edge Cases** (`edge-cases/`)

   - Boundary values (0, negative, very large)
   - Special characters and unicode
   - Empty/whitespace inputs
   - Long strings
   - Extreme coordinates

3. **Complex Workflows** (`workflows/`)

   - Shopping cart flow (browse → add → checkout)
   - Login flow with validation and MFA
   - Multi-step form submissions
   - Data persistence across pages

4. **Error Scenarios** (`error-scenarios/`)
   - Invalid arguments
   - Malformed YAML
   - API limitations
   - Missing required fields

## Running Tests

### Prerequisites

1. Build the CLI binary:

   ```bash
   make build
   ```

2. Set up test environment:
   ```bash
   export VIRTUOSO_API_TOKEN="your-api-token"
   ```

### Quick Start

```bash
# Make test runner executable
chmod +x test-yaml-suite/run-tests.sh

# Run smoke tests (quick validation)
./test-yaml-suite/run-tests.sh smoke

# Run full test suite
./test-yaml-suite/run-tests.sh full
```

### Test Execution Modes

- **smoke**: Runs key tests from each category (~3 minutes)
- **full**: Runs entire test suite (~30 minutes)

### Using Individual Tests

You can run individual YAML tests directly:

```bash
# Run a specific test file
./bin/api-cli run-test -f test-yaml-suite/commands/step-navigate/positive/all-navigate-commands.yaml -c <checkpoint-id>

# Using session context
export VIRTUOSO_SESSION_ID=<checkpoint-id>
./bin/api-cli run-test -f test-yaml-suite/workflows/e2e-scenarios/shopping-cart-flow.yaml
```

## YAML Test Format

### Basic Structure

```yaml
name: "Test Name"
description: "Test description"

config:
  continue_on_error: false
  output_format: "json"

steps:
  - navigate:
      action: "to"
      url: "https://example.com"

  - interact:
      action: "click"
      target: "button.submit"

  - assert:
      type: "exists"
      target: "div.success"
```

### Using Variables

```yaml
steps:
  - data:
      action: "store"
      type: "element-text"
      target: "h1"
      variable: "pageTitle"

  - assert:
      type: "equals"
      target: "title"
      expected: "{{pageTitle}}"
```

### Conditional Execution

```yaml
steps:
  - interact:
      action: "click"
      target: "button.optional"
      continue_on_error: true
```

## Test Data and Fixtures

The `config/test-config.yaml` file contains:

- Common selectors
- Test user credentials
- Reusable test data
- Timing configurations
- Browser settings

## Key Features

1. **Comprehensive Coverage**: All 69 CLI commands tested
2. **Edge Case Testing**: Unicode, special chars, boundaries
3. **Real-world Workflows**: E2E scenarios like shopping/login
4. **Error Handling**: Negative tests for robustness
5. **Reusable Components**: Shared config and test data
6. **Detailed Reporting**: JSON and text reports generated

## Expected Test Results

Based on current API limitations, expect:

- ~87% pass rate for command tests
- 100% pass rate for supported commands
- Known failures:
  - Some mouse operations (API limitations)
  - Store element-value (API limitation)
  - Keyboard modifiers via API

## Contributing

To add new tests:

1. Choose appropriate category/directory
2. Follow naming convention: `<feature>-<type>.yaml`
3. Use existing patterns for consistency
4. Include both positive and negative cases
5. Document any special requirements

## Troubleshooting

### Common Issues

1. **Authentication errors**: Check VIRTUOSO_API_TOKEN
2. **Checkpoint not found**: Ensure valid checkpoint ID
3. **Timeout errors**: Increase wait times in tests
4. **Element not found**: Verify selectors match target app

### Debug Mode

```bash
# Enable verbose output
VERBOSE=true ./test-yaml-suite/run-tests.sh

# Check individual test logs
ls test-yaml-suite/logs/
```

## Summary

This comprehensive YAML test suite provides:

- Full command coverage validation
- Edge case and error testing
- Real-world workflow validation
- Automated test execution
- Detailed reporting

The suite ensures the Virtuoso API CLI works correctly across all supported operations and handles edge cases gracefully.
