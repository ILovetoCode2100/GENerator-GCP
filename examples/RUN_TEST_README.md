# Run Test Command

The `run-test` command provides a simplified way to create and execute Virtuoso tests using YAML or JSON definitions.

## Features

- **Automatic Infrastructure Creation**: Creates project, goal, journey, and checkpoint automatically
- **Simplified Syntax**: Focus on test steps, not boilerplate
- **Multiple Input Formats**: YAML, JSON, files, or stdin
- **Dry Run Mode**: Preview what will be created
- **Execution Support**: Optionally execute tests after creation

## Basic Usage

```bash
# Run from file
api-cli run-test test.yaml

# Run from stdin
cat test.yaml | api-cli run-test -

# Dry run to preview
api-cli run-test test.yaml --dry-run

# Create and execute
api-cli run-test test.yaml --execute

# Auto-generate unique names
api-cli run-test test.yaml --auto-name
```

## Simplified Test Format

### Minimal Example

```yaml
name: "My Test"
steps:
  - navigate: "https://example.com"
  - click: "#login"
  - assert: "Welcome"
```

### Common Steps

```yaml
steps:
  # Navigation
  - navigate: "https://example.com"

  # Interactions
  - click: "button.submit"
  - hover: ".menu-item"
  - key: "Enter"

  # Text input
  - write:
      selector: "#email"
      text: "test@example.com"

  # Assertions
  - assert: "Success message" # Defaults to 'exists'

  # Waiting
  - wait: 2000 # Time in milliseconds
  - wait: ".loading-done" # Wait for element

  # Scrolling
  - scroll: "#footer" # Scroll to element
  - scroll:
      to: "bottom" # Scroll to top/bottom

  # Data storage
  - store:
      selector: ".price"
      as: "productPrice"

  # Comments
  - comment: "Manual verification needed here"

  # JavaScript
  - execute: "console.log('Test step')"
```

### Advanced Example

```yaml
name: "Complete User Flow"
project: 123 # Use existing project ID
config:
  base_url: "https://app.example.com"
  continue_on_error: true

steps:
  - navigate: "https://app.example.com/login"

  - write:
      selector: "#username"
      text: "testuser"

  - write:
      selector: "#password"
      text: "password123"

  - click: "button[type='submit']"

  - wait: ".dashboard"

  - assert: "Dashboard"

  - store:
      selector: ".user-name"
      as: "userName"

  - execute: |
      console.log('Logged in as:', variables.userName);

  - scroll:
      to: "bottom"

  - click: ".logout"

  - assert: "You have been logged out"
```

## Verbose Format

The command also supports the verbose format for more control:

```yaml
name: "Detailed Test"
steps:
  - type: navigate
    target: "https://example.com"
    description: "Go to homepage"

  - type: interact
    command: click
    target: "#login-button"

  - type: assert
    command: exists
    target: ".welcome-message"

  - type: wait
    command: time
    value: "3000"
```

## Project Handling

The `project` field accepts multiple formats:

```yaml
# Create new project (default)
# project field omitted

# Use existing project ID
project: 123

# Create project with specific name
project: "My Test Project"

# Deprecated but still supported
project_id: 123
project_name: "My Test Project"
```

## Output Formats

```bash
# Human-readable (default)
api-cli run-test test.yaml

# JSON output
api-cli run-test test.yaml -o json

# YAML output
api-cli run-test test.yaml -o yaml
```

## Test Results

The command provides detailed results including:

- Created infrastructure IDs (project, goal, journey, checkpoint)
- Step creation status
- Execution ID (if --execute used)
- Direct links to Virtuoso UI
- Error messages for failed steps

## Tips

1. **Start Simple**: Begin with basic navigation and assertions
2. **Use Comments**: Add comments for manual verification steps
3. **Store Variables**: Use store commands to capture dynamic values
4. **Error Handling**: Use `continue_on_error` in config for resilient tests
5. **Dry Run First**: Always do a dry run before creating infrastructure

## Examples Directory

See the `examples/` directory for more test examples:

- `simple_login_test.yaml` - Basic login flow
- `advanced_test.yaml` - E-commerce purchase with variables
- More examples in `test-commands/yaml/`
