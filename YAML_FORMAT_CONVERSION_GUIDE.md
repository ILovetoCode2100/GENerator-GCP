# YAML Format Conversion Guide

This guide shows how to convert between the different YAML formats supported by Virtuoso API CLI.

## Format Overview

1. **Compact Format** - Used by `yaml` commands (validate, compile, run, generate)
2. **Simplified Format** - Used by `run-test` command
3. **Extended Format** - Not supported by CLI (documentation only)

## Conversion Examples

### Example 1: Basic Navigation and Assertion

#### Compact Format (for `yaml` commands)

```yaml
test: Basic Navigation Test
nav: https://example.com
do:
  - wait: body
  - ch: "Example Domain"
  - note: "Page loaded successfully"
```

#### Simplified Format (for `run-test` command)

```yaml
name: "Basic Navigation Test"
steps:
  - navigate: "https://example.com"
  - wait: "body"
  - assert: "Example Domain"
  - comment: "Page loaded successfully"
```

#### CLI Commands (generated from either format)

```bash
step-navigate to cp_12345 "https://example.com" 1
step-wait element cp_12345 "body" 2
step-assert exists cp_12345 "Example Domain" 3
step-misc comment cp_12345 "Page loaded successfully" 4
```

### Example 2: Form Interaction

#### Compact Format

```yaml
test: Login Form Test
nav: https://example.com/login
data:
  email: test@example.com
  password: SecurePass123!
do:
  - c: "#email"
  - t: $email
  - c: "#password"
  - t: $password
  - c: "button[type='submit']"
  - wait: .dashboard
  - ch: "Welcome"
```

#### Simplified Format

```yaml
name: "Login Form Test"
steps:
  - navigate: "https://example.com/login"
  - click: "#email"
  - write:
      selector: "#email"
      text: "test@example.com"
  - click: "#password"
  - write:
      selector: "#password"
      text: "SecurePass123!"
  - click: "button[type='submit']"
  - wait: ".dashboard"
  - assert: "Welcome"
```

### Example 3: Data Storage and Variables

#### Compact Format

```yaml
test: Data Storage Test
nav: https://example.com
do:
  - wait: body
  - store:
      element-text: ".user-name"
      variable: userName
  - note: "Logged in as: {{userName}}"
  - ch: $userName
```

#### Simplified Format

```yaml
name: "Data Storage Test"
steps:
  - navigate: "https://example.com"
  - wait: "body"
  - store:
      selector: ".user-name"
      as: "userName"
      type: "text"
  - comment: "Logged in as: {{userName}}"
  - assert: "{{userName}}"
```

## Conversion Mapping Table

| Compact                      | Simplified                                 | Description               |
| ---------------------------- | ------------------------------------------ | ------------------------- |
| `nav: URL`                   | `navigate: URL`                            | Navigate to URL           |
| `c: selector`                | `click: selector`                          | Click element             |
| `t: text`                    | `write: {text: "..."}`                     | Type text (current field) |
| `t: {selector: text}`        | `write: {selector: "...", text: "..."}`    | Type in specific field    |
| `ch: text`                   | `assert: text`                             | Check text exists         |
| `wait: selector`             | `wait: selector`                           | Wait for element          |
| `wait: 1000`                 | `wait: 1000`                               | Wait milliseconds         |
| `note: text`                 | `comment: text`                            | Add comment               |
| `store: {...}`               | `store: {...}`                             | Store data                |
| `k: Enter`                   | `key: Enter`                               | Press key                 |
| `h: selector`                | `hover: selector`                          | Hover over element        |
| `select: {selector: option}` | `select: {selector: "...", option: "..."}` | Select dropdown option    |

## Variable Usage

### Compact Format

- Define in `data:` section
- Reference with `$varName` in actions
- Use `{{varName}}` in strings for interpolation

### Simplified Format

- No separate data section
- Store variables with `store:` action
- Use `{{varName}}` for interpolation

## When to Use Each Format

### Use Compact Format When:

- You need AI-friendly test generation (`yaml generate`)
- You want to validate syntax before execution (`yaml validate`)
- You need to see CLI commands (`yaml compile`)
- You want minimal token usage
- You're using advanced features (loops, conditionals)

### Use Simplified Format When:

- You want quick test execution (`run-test`)
- You need automatic infrastructure creation
- You prefer more readable syntax
- You're creating simple sequential tests
- You want to specify project/goal names directly

## Conversion Script Example

To convert between formats programmatically:

```bash
# Compact to CLI commands
./bin/api-cli yaml compile compact-test.yaml -o commands > commands.txt

# Simplified to execution (dry run shows structure)
./bin/api-cli run-test simplified-test.yaml --dry-run -o json

# Validate compact format
./bin/api-cli yaml validate compact-test.yaml
```

## Important Notes

1. **Extended Format** (with `type:` and `command:` fields) is not supported by any CLI command
2. **Variable syntax** differs between formats (`$var` vs `{{var}}`)
3. **Project specification** is built-in for Simplified format but requires external setup for Compact
4. **Validation** is only available for Compact format
5. **Compilation to CLI commands** is only available for Compact format

## Best Practices

1. Choose one format and stick with it for consistency
2. Use Compact format for complex tests with conditions/loops
3. Use Simplified format for straightforward test sequences
4. Always validate Compact format files before execution
5. Test with `--dry-run` flag first when using `run-test`
