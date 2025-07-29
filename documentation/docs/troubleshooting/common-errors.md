---
id: common-errors
title: Common Errors and Solutions
sidebar_position: 1
---

# Common Errors and Solutions

This guide covers the most frequently encountered errors when using the Virtuoso Test Converter API and provides detailed solutions for each.

## Conversion Errors

### Error: Unsupported Framework Format

```json
{
  "error": {
    "code": "UNSUPPORTED_FORMAT",
    "message": "Framework format 'selenium-ruby' is not supported",
    "supported_formats": [
      "selenium-java",
      "selenium-python",
      "selenium-javascript",
      "selenium-csharp",
      "cypress",
      "playwright"
    ]
  }
}
```

**Causes:**

- Using an unsupported language variant
- Typo in format specification
- Outdated API version

**Solutions:**

1. **Check supported formats:**

```bash
curl -X GET https://api.virtuoso.qa/v1/convert/formats \
  -H "X-API-Key: YOUR_API_KEY"
```

2. **Use correct format identifier:**

```json
{
  "source_format": "selenium-python", // ✓ Correct
  "source_format": "selenium-py" // ✗ Incorrect
}
```

3. **Consider preprocessing:**

```ruby
# Convert Ruby Selenium to Python format first
def convert_ruby_to_python(ruby_code)
  # Custom conversion logic
end
```

### Error: Pattern Not Recognized

```json
{
  "error": {
    "code": "PATTERN_NOT_FOUND",
    "message": "No matching pattern found for: driver.execute_cdp_cmd()",
    "confidence": 0.0,
    "suggestions": [
      "Consider using execute_script instead",
      "Submit this pattern for review"
    ]
  }
}
```

**Causes:**

- Using framework-specific advanced features
- Custom helper methods
- Newer API methods not yet supported

**Solutions:**

1. **Use alternative patterns:**

```python
# Instead of CDP commands
driver.execute_cdp_cmd('Network.enable', {})

# Use JavaScript execution
driver.execute_script('// equivalent JS code')
```

2. **Expand custom methods:**

```javascript
// Instead of custom command
cy.loginUser("admin");

// Expand to basic commands
cy.visit("/login");
cy.get("#username").type("admin");
cy.get("#password").type("password");
cy.get("#submit").click();
```

3. **Submit pattern for support:**

```bash
curl -X POST https://api.virtuoso.qa/v1/feedback/pattern \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "pattern": "driver.execute_cdp_cmd()",
    "use_case": "Network throttling",
    "frequency": "common"
  }'
```

### Error: Syntax Parse Error

```json
{
  "error": {
    "code": "PARSE_ERROR",
    "message": "Failed to parse test content",
    "details": {
      "line": 15,
      "column": 23,
      "error": "Unexpected token '{'"
    }
  }
}
```

**Causes:**

- Invalid syntax in source code
- Encoding issues
- Incomplete code snippets

**Solutions:**

1. **Validate syntax before submission:**

```python
import ast

def validate_python_syntax(code):
    try:
        ast.parse(code)
        return True, None
    except SyntaxError as e:
        return False, str(e)
```

2. **Ensure proper encoding:**

```bash
# UTF-8 encoding
curl -X POST https://api.virtuoso.qa/v1/convert \
  -H "Content-Type: application/json; charset=utf-8" \
  -d @test.json
```

3. **Submit complete test files:**

```javascript
// ✗ Incomplete
cy.get(".button").click();

// ✓ Complete context
describe("Test Suite", () => {
  it("should click button", () => {
    cy.visit("/page");
    cy.get(".button").click();
  });
});
```

## Authentication Errors

### Error: Invalid API Key

```json
{
  "error": {
    "code": "INVALID_API_KEY",
    "message": "The provided API key is invalid or has been revoked"
  }
}
```

**Solutions:**

1. **Verify API key format:**

```bash
# Correct format: vrt_[env]_[identifier]
export API_KEY="vrt_live_abc123xyz789"  # ✓
export API_KEY="abc123xyz789"           # ✗
```

2. **Check for typos:**

```python
# Common issues
api_key = "vrt_live_abc123 "  # Extra space
api_key = "vrt_live_abc123\n" # Newline character
api_key = api_key.strip()      # Clean the key
```

3. **Regenerate if needed:**
   - Log into dashboard
   - Revoke compromised key
   - Generate new key
   - Update all environments

### Error: Rate Limit Exceeded

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "API rate limit exceeded",
    "retry_after": 3600,
    "limit": 1000,
    "remaining": 0,
    "reset": "2024-01-15T12:00:00Z"
  }
}
```

**Solutions:**

1. **Implement exponential backoff:**

```python
import time
import random

def retry_with_backoff(func, max_retries=5):
    for i in range(max_retries):
        try:
            return func()
        except RateLimitError as e:
            wait_time = min(2 ** i + random.random(), e.retry_after)
            time.sleep(wait_time)
    raise Exception("Max retries exceeded")
```

2. **Use batch endpoints:**

```json
// Instead of multiple single conversions
POST /v1/convert/batch
{
  "tests": [
    {"name": "test1", "content": "..."},
    {"name": "test2", "content": "..."},
    {"name": "test3", "content": "..."}
  ]
}
```

3. **Monitor usage:**

```javascript
// Check remaining quota
const response = await fetch("/v1/patterns", {
  headers: { "X-API-Key": apiKey },
});

const remaining = response.headers.get("X-RateLimit-Remaining");
if (remaining < 100) {
  console.warn("Low API quota:", remaining);
}
```

## Execution Errors

### Error: Test Execution Timeout

```json
{
  "error": {
    "code": "EXECUTION_TIMEOUT",
    "message": "Test execution exceeded maximum time of 300 seconds",
    "execution_id": "exec_123",
    "duration": 300000
  }
}
```

**Causes:**

- Infinite loops in test
- Waiting for elements that never appear
- Network issues

**Solutions:**

1. **Add explicit timeouts:**

```yaml
steps:
  - wait:
      element: ".slow-loading"
      timeout: 10000 # 10 seconds max
  - assert:
      exists: ".content"
      timeout: 5000 # 5 seconds max
```

2. **Break long tests:**

```javascript
// Instead of one long test
describe("Complete user flow", () => {
  it("should complete all steps", () => {
    // 50+ steps
  });
});

// Split into smaller tests
describe("User flow", () => {
  it("should complete registration", () => {
    // 10 steps
  });

  it("should complete profile setup", () => {
    // 10 steps
  });
});
```

3. **Use checkpoints:**

```yaml
checkpoints:
  - name: "Registration"
    steps: [...]
  - name: "Profile Setup"
    steps: [...]
  - name: "First Purchase"
    steps: [...]
```

### Error: Element Not Found

```json
{
  "error": {
    "code": "ELEMENT_NOT_FOUND",
    "message": "Element '#submit-button' not found after 30s",
    "selector": "#submit-button",
    "timeout": 30000,
    "suggestions": [
      "Verify selector is correct",
      "Check if element is in iframe",
      "Ensure page is fully loaded"
    ]
  }
}
```

**Solutions:**

1. **Use robust selectors:**

```yaml
# ✗ Fragile
click: "#btn_1234"

# ✓ Robust
click: "[data-test-id='submit-button']"
click: "button[type='submit']"
click: "Submit Order"  # Text selector
```

2. **Handle dynamic content:**

```yaml
steps:
  # Wait for page to stabilize
  - wait:
      element: ".loading"
      state: "hidden"
  - wait:
      element: "#submit-button"
      state: "visible"
  - click: "#submit-button"
```

3. **Check for iframes:**

```yaml
steps:
  - switch-to-iframe: "#payment-frame"
  - write:
      selector: "#card-number"
      text: "4111111111111111"
  - switch-to-parent-frame
```

## Data Errors

### Error: Invalid Test Data

```json
{
  "error": {
    "code": "INVALID_TEST_DATA",
    "message": "Test data validation failed",
    "details": [
      {
        "field": "steps[2].text",
        "issue": "Special characters not properly escaped"
      }
    ]
  }
}
```

**Solutions:**

1. **Escape special characters:**

```yaml
# ✗ Incorrect
write:
  selector: "#input"
  text: "Test with "quotes""

# ✓ Correct
write:
  selector: "#input"
  text: "Test with \"quotes\""
```

2. **Use proper data types:**

```json
{
  "timeout": "5000", // ✗ String
  "timeout": 5000, // ✓ Number

  "enabled": "true", // ✗ String
  "enabled": true // ✓ Boolean
}
```

3. **Validate before submission:**

```javascript
const Ajv = require("ajv");
const ajv = new Ajv();

const schema = {
  type: "object",
  properties: {
    name: { type: "string" },
    steps: {
      type: "array",
      items: {
        type: "object",
        required: ["action"],
      },
    },
  },
};

const valid = ajv.validate(schema, testData);
if (!valid) {
  console.error(ajv.errors);
}
```

## Network Errors

### Error: Connection Timeout

```json
{
  "error": {
    "code": "CONNECTION_TIMEOUT",
    "message": "Request timeout after 30000ms",
    "endpoint": "/v1/convert/batch"
  }
}
```

**Solutions:**

1. **Increase client timeout:**

```javascript
const response = await fetch(url, {
  method: "POST",
  headers: headers,
  body: JSON.stringify(data),
  timeout: 60000, // 60 seconds
});
```

2. **Use async processing:**

```bash
# Submit for async processing
curl -X POST https://api.virtuoso.qa/v1/convert/batch \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "X-Async: true" \
  -d @large-batch.json

# Returns immediately with job ID
{
  "job_id": "job_789",
  "status": "queued",
  "webhook_url": "https://api.virtuoso.qa/v1/webhooks/job_789"
}
```

3. **Implement retry logic:**

```python
from tenacity import retry, stop_after_attempt, wait_exponential

@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=4, max=10)
)
def call_api(url, data):
    response = requests.post(url, json=data, timeout=30)
    response.raise_for_status()
    return response.json()
```

## Getting Help

If you encounter errors not covered here:

1. **Check API Status**: https://status.virtuoso.qa
2. **Search Documentation**: Use the search bar above
3. **Community Support**: [Discord](https://discord.gg/virtuoso)
4. **Submit Support Ticket**: support@virtuoso.qa

Include in your support request:

- Request ID from error response
- Complete error message
- Sample code that reproduces the issue
- API version and SDK version
