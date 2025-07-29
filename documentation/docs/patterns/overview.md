---
id: overview
title: Pattern Library Overview
sidebar_position: 1
---

# Pattern Library

The Virtuoso Pattern Library contains over 100+ conversion patterns that map test commands from various frameworks to Virtuoso's format. Each pattern includes confidence scores, alternative approaches, and performance considerations.

## What are Patterns?

Patterns are reusable templates that define how to convert specific testing actions from one framework to another. They include:

- **Source Pattern**: The original framework command
- **Target Pattern**: The Virtuoso equivalent
- **Confidence Score**: How accurately the pattern converts
- **Alternatives**: Other ways to achieve the same result
- **Performance Impact**: Expected execution time difference

## Pattern Categories

### 🧭 Navigation Patterns

Handle page navigation, URL changes, and browser actions:

- Basic navigation (go to URL)
- Browser back/forward
- Page refresh
- Scroll operations
- Window management

### 🖱️ Interaction Patterns

User interactions with page elements:

- Click actions (single, double, right-click)
- Form filling and text input
- Keyboard actions
- Mouse movements
- Drag and drop

### ✅ Assertion Patterns

Validate page state and content:

- Element existence
- Text content verification
- Attribute checking
- Value comparisons
- Custom validations

### 💾 Data Patterns

Handle data storage and retrieval:

- Variable storage
- Cookie management
- Local/session storage
- API responses
- File operations

## Pattern Structure

Each pattern follows this structure:

```yaml
pattern:
  id: "NAV-001"
  name: "Basic URL Navigation"
  category: "navigation"

  source:
    framework: "selenium"
    languages: ["java", "python", "javascript", "csharp"]
    commands:
      - java: "driver.get(url)"
      - python: "driver.get(url)"
      - javascript: "await driver.get(url)"
      - csharp: "driver.Navigate().GoToUrl(url)"

  target:
    framework: "virtuoso"
    command: "navigate"
    syntax: "navigate: {url}"

  confidence: 0.99

  notes:
    - "Direct 1:1 mapping"
    - "URL validation included"

  examples:
    - source: 'driver.get("https://example.com")'
      target: 'navigate: "https://example.com"'

  alternatives:
    - command: "navigate-and-wait"
      when: "Page load needs explicit wait"
      confidence: 0.95
```

## Confidence Scores

Patterns are rated on conversion accuracy:

| Score     | Meaning       | Description                                          |
| --------- | ------------- | ---------------------------------------------------- |
| 0.95-1.0  | **Perfect**   | Direct 1:1 mapping, no functionality loss            |
| 0.85-0.94 | **Excellent** | Minor syntax differences, same behavior              |
| 0.70-0.84 | **Good**      | Some adaptation needed, core functionality preserved |
| 0.50-0.69 | **Fair**      | Significant differences, manual review recommended   |
| < 0.50    | **Poor**      | Major changes required, consider alternatives        |

## Using Patterns

### 1. Browse by Category

Explore patterns organized by testing action:

- [Navigation Patterns](./navigation/basic-navigation)
- [Interaction Patterns](./interaction/click-patterns)
- [Assertion Patterns](./assertion/element-assertions)
- [Data Patterns](./data/variable-storage)

### 2. Search by Framework

Find patterns for your specific framework:

```bash
# API endpoint to search patterns
GET /v1/patterns?source_framework=selenium&language=python
```

### 3. Test Pattern Matching

Test if your code matches a pattern:

```bash
curl -X POST https://api.virtuoso.qa/v1/patterns/match \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "code": "driver.find_element(By.ID, \"submit\").click()",
    "framework": "selenium-python"
  }'
```

## Pattern Examples

### High Confidence Pattern

**Selenium Click → Virtuoso Click**

```python
# Selenium (Python)
driver.find_element(By.ID, "submit-button").click()

# Virtuoso
click: "#submit-button"
```

- Confidence: 0.98
- Direct mapping
- No functionality loss

### Medium Confidence Pattern

**Cypress Custom Command → Virtuoso Steps**

```javascript
// Cypress
cy.login('user@example.com', 'password123')

// Virtuoso (expanded)
steps:
  - navigate: "/login"
  - write:
      selector: "#email"
      text: "user@example.com"
  - write:
      selector: "#password"
      text: "password123"
  - click: "#submit"
```

- Confidence: 0.75
- Requires expansion of custom command
- Manual review recommended

### Low Confidence Pattern

**Selenium Action Chains → Virtuoso Mouse Actions**

```python
# Selenium (Python)
actions = ActionChains(driver)
actions.move_to_element(element)
actions.click_and_hold()
actions.move_by_offset(100, 50)
actions.release()
actions.perform()

# Virtuoso (alternative approach)
steps:
  - hover: "#drag-element"
  - mouse-down: "#drag-element"
  - mouse-move: {x: 100, y: 50}
  - mouse-up: {x: 100, y: 50}
```

- Confidence: 0.65
- Complex action chain
- May need manual adjustment

## Contributing Patterns

Help improve the pattern library:

### Submit New Pattern

```json
{
  "source_framework": "playwright",
  "source_code": "await page.waitForSelector('.loading', {state: 'hidden'})",
  "suggested_target": "wait-hidden: '.loading'",
  "use_case": "Wait for loading indicator to disappear",
  "frequency": "common"
}
```

### Report Pattern Issues

```json
{
  "pattern_id": "INT-042",
  "issue_type": "incorrect_conversion",
  "description": "Pattern doesn't handle dynamic selectors",
  "example": "driver.find_element(By.XPATH, f'//div[@id=\"{item_id}\"]')",
  "expected_result": "Parameterized selector support"
}
```

## Pattern Performance

Understanding performance implications:

| Pattern Type | Selenium | Cypress  | Playwright | Virtuoso  |
| ------------ | -------- | -------- | ---------- | --------- |
| Navigation   | 2-3s     | 1-2s     | 1-2s       | 1-2s      |
| Click        | 100ms    | 50ms     | 50ms       | 50ms      |
| Assert       | 50ms     | 30ms     | 30ms       | 40ms      |
| Wait         | Variable | Variable | Variable   | Optimized |

## Best Practices

1. **Review Low Confidence Conversions**: Always manually review patterns with confidence < 0.8
2. **Test Converted Patterns**: Run converted tests in a staging environment first
3. **Submit Feedback**: Help improve pattern accuracy by reporting issues
4. **Use Alternatives**: Consider alternative patterns for better performance
5. **Combine Patterns**: Some complex actions require multiple patterns

## Next Steps

- **[Browse All Patterns](./navigation/basic-navigation)** - Explore the complete library
- **[Submit Feedback](../api/endpoints/feedback/submit-feedback)** - Help improve patterns
- **[Custom Patterns](../advanced/custom-patterns)** - Create organization-specific patterns
- **[Pattern API](../api/endpoints/patterns/list-patterns)** - Programmatic pattern access
