# Virtuoso YAML Schema

## Minimal Test Syntax

```yaml
# Simplest possible test
test: Login Test
url: https://example.com
steps:
  - click: Login
  - type: {user: test@example.com}
  - type: {pass: password123}
  - click: Submit
  - check: Welcome back

# With shortcuts expanded
test: Login Test
nav: https://example.com  # 'nav' shorter than 'navigate'
do:                       # 'do' shorter than 'steps'
  - c: Login              # 'c' for click
  - t: {user: test@example.com}  # 't' for type
  - t: {pass: password123}
  - c: Submit
  - ch: Welcome back      # 'ch' for check
```

## Complete Schema

```yaml
# Test metadata (all optional except 'test')
test: string # Test name (required)
desc: string # Description
tags: [string] # Tags for organization
env: string # Environment (dev/staging/prod)
retry: number # Retry count on failure
timeout: number # Test timeout in seconds

# Navigation
url: string # Starting URL (alias: nav)
base: string # Base URL for relative paths

# Variables and data
vars: # Variables (alias: v)
  key: value
data: # Data sets for iteration
  - { user: user1, pass: pass1 }
  - { user: user2, pass: pass2 }

# Main test steps
steps: # Test steps (alias: do)
  - action: target # Simple syntax
  - action: # Extended syntax
      target: selector
      value: input
      wait: number
      retry: number
      optional: bool
      context: string

# Reusable blocks
blocks: # Reusable step groups
  login:
    - type: { user: $user }
    - type: { pass: $pass }
    - click: Submit

# Includes
include: # Include other YAML files
  - common/login.yaml
  - common/logout.yaml

# Assertions
assert: # Global assertions
  - exists: .success
  - text: { h1: Welcome }

# Cleanup
cleanup: # Always run after test
  - click: Logout
  - clear: cookies
```

## Action Types

### Interaction Actions

```yaml
# Click variants
- click: button # Simple click
- c: button # Short form
- dblclick: item # Double click
- rightclick: menu # Right click
- hover: tooltip # Hover

# Text input
- type: { input: value } # Type into element
- t: { input: value } # Short form
- write: { input: value } # Alias for type
- clear: input # Clear field

# Selection
- select: { dropdown: Option }
- choose: { radio: value }
- check: checkbox
- uncheck: checkbox

# Keyboard
- key: Enter # Single key
- keys: Ctrl+A # Key combination
```

### Navigation Actions

```yaml
# Page navigation
- nav: https://url # Navigate to URL
- goto: /relative/path # Relative navigation
- back: # Browser back
- forward: # Browser forward
- refresh: # Reload page

# Scrolling
- scroll: bottom # Scroll to position
- scroll: { to: element } # Scroll to element
- scroll: { by: 200 } # Scroll by pixels
```

### Wait Actions

```yaml
# Element waits
- wait: element # Wait for visible
- wait: { for: element } # Explicit syntax
- wait: { gone: element } # Wait for disappear
- pause: 1000 # Wait milliseconds
```

### Assertions

```yaml
# Existence
- check: element # Element exists (alias: ch)
- exists: element # Explicit exists
- missing: element # Element not exists

# Text assertions
- text: { el: "Expected" } # Text equals
- contains: { el: "text" } # Text contains
- matches: { el: "regex" } # Text matches regex

# Attribute assertions
- attr: { el.class: active }
- value: { input: expected }
- checked: checkbox
- selected: { option: value }

# Comparison assertions
- equals: { $var: value }
- gt: { $count: 5 }
- lt: { $time: 1000 }
```

### Data Actions

```yaml
# Store values
- store: { element: varName } # Store text
- store: { attr.href: linkVar } # Store attribute
- store: { value.input: inputVar } # Store value

# Cookies
- cookie: { set: { name: value } }
- cookie: { delete: name }
- cookie: clear
```

### Advanced Actions

```yaml
# JavaScript execution
- js: "return document.title"
- script: |
    const elements = document.querySelectorAll('.item');
    return elements.length;

# Dialogs
- alert: accept
- confirm: { accept: true }
- prompt: { text: "Input", accept: true }

# Windows/tabs
- window: { resize: 1024x768 }
- tab: new
- tab: { switch: next }
- iframe: { switch: frame1 }
- iframe: parent

# File operations
- upload: { input: https://url/file.pdf }
```

## Selectors

```yaml
# CSS selectors (default)
- click: button.primary
- click: "#submit"
- click: "[data-test=login]"

# Text selectors
- click: "Login"         # Exact text
- click: ~Login          # Contains text
- click: /Log.*in/       # Regex

# XPath (prefix with //)
- click: //button[text()='Submit']

# Special selectors
- click: $variable       # Variable reference
- click: @login-button   # Data-test-id shorthand
```

## Variables and Templating

```yaml
vars:
  user: test@example.com
  pass: ${ENV:PASSWORD} # Environment variable

steps:
  - type: { email: $user }
  - type: { password: $pass }
  - store: { h1: pageTitle }
  - check: { title: $pageTitle }
```

## Control Flow

```yaml
# Conditionals
- if: { exists: .premium }
  then:
    - click: Premium features
  else:
    - click: Upgrade

# Loops
- repeat: 3
  do:
    - click: Add item

- foreach: $data
  do:
    - type: { input: $item.name }
    - click: Save

# Error handling
- try:
    - click: Optional element
  catch:
    - log: Element not found
  finally:
    - screenshot: attempt
```

## Imports and Composition

```yaml
# Import common steps
include:
  - auth/login.yaml
  - utils/helpers.yaml

# Use imported blocks
steps:
  - use: login
    with:
      user: admin@example.com
      pass: admin123
  - use: navigate-to-dashboard
```

## Metadata and Configuration

```yaml
# Test configuration
config:
  viewport: 1920x1080
  device: desktop
  browser: chrome
  headless: false

# Test organization
suite: Regression
priority: high
tags: [critical, auth, smoke]

# Execution control
parallel: false
retry:
  count: 2
  delay: 5000
timeout: 30000

# Reporting
report:
  screenshots: failure # always/failure/never
  video: true
  traces: true
```

## Shortcuts Reference

| Long Form | Short | Purpose          |
| --------- | ----- | ---------------- |
| navigate  | nav   | Navigation       |
| steps     | do    | Step list        |
| click     | c     | Click action     |
| type      | t     | Type action      |
| check     | ch    | Assertion        |
| variable  | var   | Variable         |
| element   | el    | Element selector |
