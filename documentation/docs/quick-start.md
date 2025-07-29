---
id: quick-start
title: Quick Start Guide
sidebar_position: 2
---

# Quick Start Guide

Get your first test converted and running in under 5 minutes! This guide will walk you through the essential steps to start using the Virtuoso Test Converter API.

## Prerequisites

Before you begin, make sure you have:

- An API key for the Virtuoso Test Converter API
- A test script to convert (Selenium, Cypress, or Playwright)
- Basic knowledge of REST APIs
- `curl` or a REST client like Postman

## Step 1: Get Your API Key

First, you'll need an API key to authenticate your requests.

```bash
# Request an API key from your Virtuoso account dashboard
# Or contact support@virtuoso.qa
```

## Step 2: Your First Conversion

Let's convert a simple Selenium test to Virtuoso format:

### Example Selenium Test (Python)

```python
from selenium import webdriver
from selenium.webdriver.common.by import By

driver = webdriver.Chrome()
driver.get("https://example.com")
driver.find_element(By.ID, "login-button").click()
driver.find_element(By.NAME, "username").send_keys("testuser@example.com")
driver.find_element(By.NAME, "password").send_keys("password123")
driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
assert "Dashboard" in driver.title
driver.quit()
```

### Convert Using the API

```bash
curl -X POST https://api.virtuoso.qa/v1/convert \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "source_format": "selenium-python",
    "test_content": "from selenium import webdriver\nfrom selenium.webdriver.common.by import By\n\ndriver = webdriver.Chrome()\ndriver.get(\"https://example.com\")\ndriver.find_element(By.ID, \"login-button\").click()\ndriver.find_element(By.NAME, \"username\").send_keys(\"testuser@example.com\")\ndriver.find_element(By.NAME, \"password\").send_keys(\"password123\")\ndriver.find_element(By.CSS_SELECTOR, \"button[type='\''submit'\'']\").click()\nassert \"Dashboard\" in driver.title\ndriver.quit()",
    "target_format": "virtuoso-yaml"
  }'
```

### Response

```json
{
  "job_id": "job_123456",
  "status": "completed",
  "result": {
    "converted_test": {
      "name": "Converted Selenium Test",
      "steps": [
        { "navigate": "https://example.com" },
        { "click": "#login-button" },
        {
          "write": {
            "selector": "input[name='username']",
            "text": "testuser@example.com"
          }
        },
        {
          "write": {
            "selector": "input[name='password']",
            "text": "password123"
          }
        },
        { "click": "button[type='submit']" },
        { "assert": { "type": "title-contains", "value": "Dashboard" } }
      ]
    },
    "confidence_score": 0.95,
    "conversion_notes": [
      "Successfully converted all commands",
      "Title assertion converted to Virtuoso format"
    ]
  }
}
```

## Step 3: Run the Converted Test

Now let's execute the converted test:

```bash
curl -X POST https://api.virtuoso.qa/v1/tests/run \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "test_definition": {
      "name": "Login Test",
      "steps": [
        {"navigate": "https://example.com"},
        {"click": "#login-button"},
        {"write": {"selector": "input[name='\''username'\'']", "text": "testuser@example.com"}},
        {"write": {"selector": "input[name='\''password'\'']", "text": "password123"}},
        {"click": "button[type='\''submit'\'']"},
        {"assert": {"type": "title-contains", "value": "Dashboard"}}
      ]
    },
    "execute": true
  }'
```

### Response

```json
{
  "test_id": "test_789",
  "execution_id": "exec_456",
  "status": "running",
  "webhook_url": "https://api.virtuoso.qa/v1/webhooks/exec_456",
  "estimated_duration": "45s"
}
```

## Step 4: Check Test Status

Monitor your test execution:

```bash
curl -X GET https://api.virtuoso.qa/v1/status/exec_456 \
  -H "X-API-Key: YOUR_API_KEY"
```

### Response

```json
{
  "execution_id": "exec_456",
  "status": "completed",
  "result": "passed",
  "duration": "42s",
  "steps_executed": 6,
  "steps_passed": 6,
  "screenshot_url": "https://virtuoso.qa/screenshots/exec_456",
  "video_url": "https://virtuoso.qa/videos/exec_456"
}
```

## Next Steps

Congratulations! You've successfully converted and run your first test. Here's what to explore next:

### 1. Try Different Formats

Convert tests from other frameworks:

- **[Cypress Conversion](./formats/cypress)** - Convert Cypress tests
- **[Playwright Conversion](./formats/playwright)** - Convert Playwright tests
- **[YAML Format](./formats/yaml-format)** - Write tests directly in YAML

### 2. Explore Advanced Features

- **[Batch Conversion](../api/endpoints/convert/batch-conversion)** - Convert multiple tests at once
- **[Pattern Library](./patterns)** - Browse supported test patterns
- **[Webhooks](../api/webhooks/overview)** - Set up async notifications

### 3. Integrate with Your Workflow

- **[CI/CD Integration](./developer-guide/integration-patterns)** - Automate conversions
- **[SDK Usage](../api/sdks/overview)** - Use our SDKs for easier integration
- **[Custom Patterns](./advanced/custom-patterns)** - Define custom conversion rules

## Common Issues

### Authentication Error

```json
{
  "error": "Invalid API key"
}
```

**Solution**: Verify your API key is correct and active.

### Unsupported Pattern

```json
{
  "error": "Pattern not recognized",
  "details": "Custom wait function not supported"
}
```

**Solution**: Check the [Pattern Library](./patterns) for supported patterns or submit feedback for new pattern support.

### Rate Limiting

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

**Solution**: Implement exponential backoff or upgrade your plan for higher limits.

## Get Help

- 📚 **[Full API Reference](../api/overview)**
- 🔍 **[Troubleshooting Guide](./troubleshooting/common-errors)**
- 💬 **[Discord Community](https://discord.gg/virtuoso)**
- 📧 **Email**: support@virtuoso.qa
