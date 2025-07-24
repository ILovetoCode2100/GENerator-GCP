# Validation Rules and Error Messages

## Validation Layers

### 1. Schema Validation

```yaml
# Rule: Test name is required
Error: "Missing required field 'test'"
Fix: "Add 'test: Your Test Name' at the beginning of your YAML file"
Example:
  test: Login Test
  url: https://example.com

# Rule: Steps or 'do' must be a list
Error: "Field 'steps' must be a list of actions"
Fix: "Use dash (-) before each step"
Example:
  steps:
    - click: Login
    - type: {user: email}

# Rule: URL must be valid
Error: "Invalid URL format: 'example.com'"
Fix: "URLs must start with http:// or https://"
Example:
  url: https://example.com
```

### 2. Semantic Validation

```yaml
# Rule: Selector must be valid
Error: "Invalid CSS selector: 'button['"
Fix: "Close brackets in selector: 'button[type=submit]'"

# Rule: Target element required for actions
Error: "Action 'click' requires a target element"
Fix: "Specify what to click: 'click: button.submit'"

# Rule: Variable must be defined before use
Error: "Undefined variable '$username'"
Fix: "Define variable in 'vars' section before using it"
Example:
  vars:
    username: test@example.com
  steps:
    - type: {email: $username}
```

### 3. Cross-Reference Validation

```yaml
# Rule: Block references must exist
Error: "Unknown block 'login-flow'"
Fix: "Define the block before using it"
Example:
  blocks:
    login-flow:
      - type: {user: $user}
      - click: Submit
  steps:
    - use: login-flow

# Rule: Imported files must exist
Error: "Cannot find import: 'common/auth.yaml'"
Fix: "Check file path or create the missing file"
```

## Validation Rules by Category

### Action Validation

```go
type ValidationRule struct {
    Name        string
    Check       func(action Action) error
    ErrorMsg    string
    FixMsg      string
    Example     string
}

var actionRules = []ValidationRule{
    {
        Name: "RequireTarget",
        Check: func(a Action) error {
            if a.Target == "" && needsTarget(a.Type) {
                return fmt.Errorf("missing target")
            }
            return nil
        },
        ErrorMsg: "Action '%s' requires a target element",
        FixMsg: "Add a target selector or text",
        Example: "- click: button.submit",
    },
    {
        Name: "ValidSelector",
        Check: func(a Action) error {
            return validateSelector(a.Target)
        },
        ErrorMsg: "Invalid selector: '%s'",
        FixMsg: "Check CSS selector syntax",
        Example: "- click: [data-test='submit']",
    },
    {
        Name: "ValidKeyCombo",
        Check: func(a Action) error {
            if a.Type == "keys" {
                return validateKeyCombo(a.Value)
            }
            return nil
        },
        ErrorMsg: "Invalid key combination: '%s'",
        FixMsg: "Use format: Ctrl+A, Shift+Tab, Alt+F4",
        Example: "- keys: Ctrl+S",
    },
}
```

### Data Validation

```yaml
# Rule: Store actions need variable name
Error: "Store action missing variable name"
Fix: "Specify variable name to store value in"
Example:
  - store: {h1.title: pageTitle}
  - check: {title: $pageTitle}

# Rule: Variable names must be valid
Error: "Invalid variable name: '123var'"
Fix: "Variable names must start with letter"
Example:
  vars:
    var123: valid
    _private: valid
    123invalid: invalid
```

### Flow Control Validation

```yaml
# Rule: If conditions must be valid
Error: "Invalid condition in 'if' statement"
Fix: "Use valid condition format"
Examples:
  # Check existence
  - if: {exists: .element}

  # Compare values
  - if: {equals: [$var, "expected"]}

  # Complex conditions
  - if:
      and:
        - exists: .element
        - equals: [$count, 5]

# Rule: Loop counts must be positive
Error: "Invalid repeat count: -1"
Fix: "Use positive number for repeat count"
Example:
  - repeat: 3
    do:
      - click: Add
```

## Error Message Templates

### Helpful Error Format

```
ERROR: <What went wrong>
Line <N>: <Show the problematic line>
         ^
Problem: <Specific issue>
Fix: <How to fix it>
Example:
<Show correct usage>

Related documentation: <link>
```

### Real Examples

```
ERROR: Invalid selector syntax
Line 15: - click: button[type=submit
                                    ^
Problem: Unclosed bracket in CSS selector
Fix: Add closing bracket ']'
Example:
  - click: button[type=submit]
  - click: "[data-test='login']"  # Quote if needed

Related: https://docs.virtuoso.qa/selectors
```

```
ERROR: Undefined variable reference
Line 23: - type: {email: $userEmail}
                         ^
Problem: Variable 'userEmail' is not defined
Fix: Define the variable in the 'vars' section
Example:
  vars:
    userEmail: test@example.com
  steps:
    - type: {email: $userEmail}

Related: https://docs.virtuoso.qa/variables
```

## Validation Implementation

```go
package validation

import (
    "fmt"
    "regexp"
    "strings"
)

type Validator struct {
    errors   []ValidationError
    warnings []ValidationWarning
    context  ValidationContext
}

type ValidationError struct {
    Line     int
    Column   int
    Field    string
    Message  string
    Fix      string
    Example  string
}

func (v *Validator) ValidateTest(test *Test) error {
    // Phase 1: Schema validation
    if err := v.validateSchema(test); err != nil {
        return err
    }

    // Phase 2: Semantic validation
    if err := v.validateSemantics(test); err != nil {
        return err
    }

    // Phase 3: Cross-reference validation
    if err := v.validateReferences(test); err != nil {
        return err
    }

    // Phase 4: Best practices (warnings)
    v.checkBestPractices(test)

    return v.formatErrors()
}

func (v *Validator) validateSelector(selector string) error {
    // Check for common selector mistakes
    checks := []struct {
        pattern string
        message string
        fix     string
    }{
        {
            pattern: `\[(?:[^"\]]*(?:"[^"]*")?)*[^"\]]*$`,
            message: "Unclosed bracket in selector",
            fix:     "Add closing bracket ']'",
        },
        {
            pattern: `^[#.]\s`,
            message: "Space after # or . in selector",
            fix:     "Remove space: '#id' not '# id'",
        },
        {
            pattern: `:contains\(`,
            message: "jQuery :contains() not supported",
            fix:     "Use text selector: ~text or 'exact text'",
        },
    }

    for _, check := range checks {
        if matched, _ := regexp.MatchString(check.pattern, selector); matched {
            return &ValidationError{
                Message: check.message,
                Fix:     check.fix,
            }
        }
    }

    return nil
}
```

## Best Practice Warnings

```yaml
# Warning: Hardcoded credentials
Warning: "Hardcoded password detected"
Suggestion: "Use environment variables for sensitive data"
Example:
  vars:
    password: ${ENV:TEST_PASSWORD}

# Warning: No assertions
Warning: "Test has no assertions"
Suggestion: "Add checks to verify expected behavior"
Example:
  steps:
    - click: Submit
    - check: Success message
    - exists: .confirmation

# Warning: Brittle selectors
Warning: "Using tag-only selector 'button'"
Suggestion: "Use more specific selectors"
Better:
  - click: button.primary
  - click: [data-test=submit]
  - click: Submit  # Text selector

# Warning: Missing wait
Warning: "No wait after navigation"
Suggestion: "Wait for page to load"
Example:
  - nav: https://example.com
  - wait: .main-content
  - click: Login
```

## Progressive Validation

```yaml
# Level 1: Critical Errors (block execution)
- Missing required fields
- Invalid syntax
- Undefined references

# Level 2: Semantic Errors (likely to fail)
- Invalid selectors
- Type mismatches
- Circular dependencies

# Level 3: Warnings (may cause issues)
- Brittle selectors
- Missing waits
- No assertions

# Level 4: Style Suggestions
- Use shortcuts for brevity
- Group related actions
- Add descriptions
```

## Validation Config

```yaml
validation:
  strict: false # Fail on warnings
  max_errors: 10 # Stop after N errors
  spell_check: true # Check selector typos
  security_check: true # Warn on hardcoded secrets

  custom_rules:
    - name: "require-data-test-ids"
      pattern: "^\\[data-test="
      message: "Use data-test attributes for selectors"

    - name: "max-step-count"
      max: 50
      message: "Consider breaking into smaller tests"
```
