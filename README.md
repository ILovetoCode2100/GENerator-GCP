# Virtuoso API CLI

**Version:** 2.0  
**Status:** Production Ready (98% test success rate)  
**Purpose:** CLI tool for Virtuoso API test automation with AI-friendly design

## 🚀 Quick Start

```bash
# Build
make build

# Configure (create ~/.api-cli/virtuoso-config.yaml)
api:
  auth_token: your-api-key-here
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"

# Run commands
./bin/api-cli assert exists "Login button"
./bin/api-cli interact click "Submit"
./bin/api-cli navigate to "https://example.com"
```

## 📋 Commands and Variations

The CLI provides 60 commands organized into 12 groups:

### Assert (12 commands)
```bash
assert exists|not-exists|equals|not-equals|checked|selected|variable|gt|gte|lt|lte|matches
```
- **exists/not-exists**: Check element presence
- **equals/not-equals**: Compare element text/value
- **checked**: Verify checkbox state
- **selected**: Check dropdown selection (needs position)
- **variable**: Compare stored variables
- **gt/gte/lt/lte**: Numeric comparisons
- **matches**: Regex pattern matching (use single quotes)

### Interact (6 commands)
```bash
interact click|double-click|right-click|hover|write|key
```
- **click**: Standard click with optional position/element-type
- **write**: Type text with optional variable storage
- **key**: Send keyboard shortcuts (e.g., "ctrl+a")

### Navigate (5 commands)
```bash
navigate to|scroll-to|scroll-top|scroll-bottom|scroll-element
```
- **to**: Navigate URL with optional --new-tab
- **scroll-\***: Smooth scrolling operations

### Data (5 commands)
```bash
data store element-text|store literal|cookie create|cookie delete|cookie clear-all
```
- **store**: Save element text or literal values to variables
- **cookie**: Manage browser cookies

### Dialog (4 commands)
```bash
dialog dismiss alert|dismiss confirm|dismiss prompt|dismiss prompt-with-text
```
- Flags: --accept, --reject for confirm/prompt dialogs

### Wait (2 commands)
```bash
wait element|time
```
- **element**: Wait for selector with --timeout
- **time**: Sleep for milliseconds

### Window (5 commands)
```bash
window resize|switch tab|switch iframe|switch parent-frame
```
- **resize**: Format WIDTHxHEIGHT (e.g., 1024x768)
- **switch tab**: next/prev navigation

### Mouse (6 commands)
```bash
mouse move-to|move-by|move|down|up|enter
```
- Coordinate-based and element-based operations

### Select (3 commands)
```bash
select index|last|option
```
- Dropdown selection by index or position

### File (1 command)
```bash
file upload URL SELECTOR
```

### Misc (3 commands)
```bash
misc comment|execute|key
```
- **comment**: Add test comments
- **execute**: Run JavaScript

### Library (6 commands)
```bash
library add|get|attach|move-step|remove-step|update
```
- Manage reusable test components

## 🤖 AI Integration Guide

### Overview
This CLI is designed as an AI-friendly interface for generating Virtuoso test automation. AI systems can parse command patterns, generate test structures, and chain operations programmatically.

### Output Formats
All commands support AI-optimized output:
```bash
--output human  # Default readable format
--output json   # Structured data for parsing
--output yaml   # Configuration format
--output ai     # Conversational AI format with context
```

### AI Output Structure
The `--output ai` format provides contextual information for test building:
```json
{
  "command": "assert exists",
  "result": "success",
  "message": "Element 'Login button' found",
  "context": {
    "checkpoint_id": "1680930",
    "position": 1,
    "journey_id": "608926"
  },
  "next_steps": [
    "interact click 'Login button'",
    "wait element '#login-form'",
    "assert exists '#username-field'"
  ],
  "test_structure": {
    "current_checkpoint": "Setup",
    "total_steps": 5
  }
}
```

### Building Test Journeys

#### Complete E2E Test Example
```bash
# 1. Create project and journey structure
api-cli create-project "E-Commerce Tests"
PROJECT_ID=13961  # From response

api-cli create-journey $PROJECT_ID "User Purchase Flow"
JOURNEY_ID=608926  # From response

# 2. Create checkpoints with meaningful names
api-cli create-checkpoint $JOURNEY_ID "Homepage Navigation" 1
api-cli create-checkpoint $JOURNEY_ID "Product Search" 2
api-cli create-checkpoint $JOURNEY_ID "Add to Cart" 3
api-cli create-checkpoint $JOURNEY_ID "Checkout Process" 4
api-cli create-checkpoint $JOURNEY_ID "Order Confirmation" 5
```

#### Detailed YAML Test Templates

##### Basic Login Test
```yaml
# examples/test-template.yaml
journey:
  name: "Login Flow - Basic"
  project_id: 13961
  description: "Standard login flow with error handling"
  checkpoints:
    - name: "Navigate to Login"
      position: 1
      steps:
        - command: navigate to
          args: ["https://example.com/login"]
          description: "Open login page"
        
        - command: wait element
          args: ["#login-form"]
          options: 
            timeout: 5000
            description: "Ensure page loaded"
        
        - command: assert exists
          args: ["#username"]
          description: "Verify username field present"

    - name: "Enter Credentials"
      position: 2
      steps:
        - command: interact click
          args: ["#username"]
          description: "Focus username field"
        
        - command: interact write
          args: ["#username", "testuser@example.com"]
          options:
            variable: "login_email"
            clear_before: true
        
        - command: interact key
          args: ["Tab"]
          description: "Move to password field"
        
        - command: interact write
          args: ["#password", "${TEST_PASSWORD}"]
          options:
            masked: true
            
    - name: "Submit and Verify"
      position: 3
      steps:
        - command: interact click
          args: ["#login-button"]
          options:
            position: "CENTER"
            element_type: "BUTTON"
        
        - command: wait element
          args: [".dashboard", ".error-message"]
          options:
            timeout: 10000
            either_or: true
        
        - command: assert not-exists
          args: [".error-message"]
          description: "No login errors"
```

##### Advanced E-Commerce Flow
```yaml
journey:
  name: "E-Commerce Purchase - Full Flow"
  project_id: 13961
  config:
    retry_failed_steps: true
    screenshot_on_error: true
  variables:
    - name: "product_name"
      value: "Wireless Headphones"
    - name: "expected_price"
      value: "$99.99"
  checkpoints:
    - name: "Product Search"
      steps:
        # Handle cookie banner if present
        - command: conditional
          condition:
            command: assert exists
            args: ["#cookie-banner"]
            timeout: 2000
          then:
            - command: interact click
              args: ["#accept-cookies"]
        
        # Search for product
        - command: interact click
          args: ["#search-input"]
        - command: interact write
          args: ["#search-input", "${product_name}"]
        - command: interact key
          args: ["Enter"]
        
        # Wait for results
        - command: wait element
          args: [".search-results"]
          options:
            timeout: 5000
        
        # Verify search worked
        - command: assert gte
          args: [".results-count", "1"]
          description: "At least one result found"

    - name: "Product Selection"
      steps:
        # Store product details
        - command: data store element-text
          args: [".product-card:first-child .price", "actual_price"]
        
        # Click first product
        - command: interact click
          args: [".product-card:first-child"]
        
        # Verify product page
        - command: wait element
          args: [".product-details"]
        - command: assert equals
          args: ["h1.product-title", "${product_name}"]

    - name: "Add to Cart"
      steps:
        # Select options if available
        - command: conditional
          condition:
            command: assert exists
            args: ["#size-selector"]
          then:
            - command: select index
              args: ["#size-selector", "1"]
        
        # Add to cart
        - command: interact click
          args: ["#add-to-cart"]
          options:
            wait_after: 1000
        
        # Handle modal or notification
        - command: wait element
          args: [".cart-notification, .cart-sidebar"]
          options:
            either_or: true
        
        # Verify item added
        - command: assert equals
          args: [".cart-count", "1"]

    - name: "Checkout Process"
      steps:
        # Go to cart
        - command: interact click
          args: [".cart-icon, .view-cart"]
        
        # Verify cart contents
        - command: wait element
          args: [".cart-page"]
        - command: assert exists
          args: [".cart-item"]
        
        # Apply coupon if available
        - command: conditional
          condition:
            variable: "COUPON_CODE"
            exists: true
          then:
            - command: interact write
              args: ["#coupon-input", "${COUPON_CODE}"]
            - command: interact click
              args: ["#apply-coupon"]
            - command: wait element
              args: [".discount-applied"]
        
        # Proceed to checkout
        - command: interact click
          args: ["#checkout-button"]

    - name: "Order Completion"
      steps:
        # Fill shipping (if not saved)
        - command: conditional
          condition:
            command: assert exists
            args: ["#shipping-form"]
          then:
            - command: interact write
              args: ["#first-name", "Test"]
            - command: interact write
              args: ["#last-name", "User"]
            - command: interact write
              args: ["#email", "${login_email}"]
            - command: interact write
              args: ["#address", "123 Test Street"]
            - command: select option
              args: ["#country", "United States"]
        
        # Payment (test mode)
        - command: interact click
          args: ["#payment-test-mode"]
        
        # Place order
        - command: interact click
          args: ["#place-order"]
        
        # Verify success
        - command: wait element
          args: [".order-success"]
          options:
            timeout: 15000
        
        # Store order number
        - command: data store element-text
          args: [".order-number", "order_id"]
```

#### Command Variations and Nuances

##### Assert Command Variations
```bash
# Text matching variations
api-cli assert equals "#title" "Exact Text"
api-cli assert equals "#title" "Exact Text" --case-insensitive
api-cli assert matches "#title" "Partial.*Text"  # Regex
api-cli assert matches "#title" '^Start.*End$'   # Full regex

# Numeric comparisons with type coercion
api-cli assert gt ".price" "50"      # Extracts number from "$50.99"
api-cli assert gte ".count" "10"     # Works with "10 items"
api-cli assert lt ".stock" "5"       # Handles "Only 3 left!"

# Special selectors
api-cli assert exists "text=Login"    # Text content selector
api-cli assert exists "button:contains('Submit')"  # jQuery-style
api-cli assert selected "#country" 2  # Dropdown by position
```

##### Interact Command Nuances
```bash
# Click variations
api-cli interact click "Submit"  # Matches by text
api-cli interact click "#submit" --position TOP_LEFT
api-cli interact click ".btn" --element-type BUTTON --force

# Write with special handling
api-cli interact write "#email" "test@example.com" --clear-first
api-cli interact write "#password" "pass123" --masked
api-cli interact write "#bio" "Line 1\nLine 2" --multiline

# Keyboard shortcuts
api-cli interact key "ctrl+a"         # Select all
api-cli interact key "cmd+v"          # Paste (Mac)
api-cli interact key "shift+Tab"      # Reverse tab
api-cli interact key "Escape"         # Close modals
```

##### Navigation Edge Cases
```bash
# Handle popups and new tabs
api-cli navigate to "https://example.com" --new-tab
api-cli window switch tab next
api-cli navigate to "javascript:alert('test')"  # JS URLs

# Scroll variations
api-cli navigate scroll-to "#footer" --smooth
api-cli navigate scroll-element ".sidebar" 500  # Scroll within element
api-cli navigate scroll-bottom --wait-after 1000
```

##### Data Operations
```bash
# Cookie management
api-cli data cookie create "session" "abc123" --domain ".example.com"
api-cli data cookie create "preferences" '{"theme":"dark"}' --http-only
api-cli data cookie delete "tracking" --all-domains

# Variable storage and math
api-cli data store element-text ".price" "price"
api-cli data store literal "50" "discount"
api-cli data calculate "price - discount" "final_price"  # Future feature
```

##### Wait Strategies
```bash
# Element waits
api-cli wait element "#loader" --not-exists --timeout 10000
api-cli wait element ".content" --visible --stable
api-cli wait element "iframe#payment" --interactive

# Smart waits
api-cli wait network-idle --timeout 5000
api-cli wait animation-complete ".modal"
```

#### Library Checkpoint Patterns
```bash
# Create reusable components
api-cli library add $CHECKPOINT_ID  # Convert to library
LIBRARY_ID=7023

# Attach to multiple journeys
api-cli library attach $JOURNEY1 $LIBRARY_ID 1
api-cli library attach $JOURNEY2 $LIBRARY_ID 5

# Modify library checkpoints
api-cli library update $LIBRARY_ID "Login Flow v2"
api-cli library move-step $LIBRARY_ID $STEP_ID 1
api-cli library remove-step $LIBRARY_ID $OLD_STEP_ID
```

#### Session Management for Complex Flows
```bash
# Set up session for test sequence
export VIRTUOSO_SESSION_ID=$CHECKPOINT_ID
export VIRTUOSO_PROJECT_ID=$PROJECT_ID

# Commands auto-increment position
api-cli navigate to "https://example.com"     # Position 1
api-cli wait element "body"                    # Position 2
api-cli assert exists ".header"                # Position 3
api-cli interact click ".login"                # Position 4

# Check current position
api-cli get-session-info --output json
```

#### Error Handling Patterns
```yaml
journey:
  name: "Robust Test with Error Handling"
  error_handling:
    screenshot_on_failure: true
    continue_on_error: false
    retry_count: 2
  checkpoints:
    - name: "Safe Navigation"
      steps:
        - command: try-catch
          try:
            - command: navigate to
              args: ["${BASE_URL}/page"]
          catch:
            - command: navigate to
              args: ["${FALLBACK_URL}"]
          finally:
            - command: wait element
              args: ["body"]
```

### Command Chaining Patterns

#### Sequential Test Steps
```bash
# Use session context for auto-incrementing positions
export VIRTUOSO_SESSION_ID=CHECKPOINT_ID

# Commands automatically chain with position tracking
api-cli assert exists "Header"  # Position 1
api-cli interact click "Login"   # Position 2
api-cli wait element "#form"    # Position 3
```

#### Conditional Flows
```bash
# Parse AI output for dynamic test generation
RESULT=$(api-cli assert exists "#promo-banner" --output json)
if [ $(echo $RESULT | jq -r '.found') = "true" ]; then
  api-cli interact click "#close-promo"
fi

# Continue main flow
api-cli interact click "#main-action"
```

#### Variable Extraction and Reuse
```bash
# Store dynamic values
api-cli data store element-text "#order-id" "orderId" --output json
ORDER_ID=$(api-cli get-variable "orderId" --output json | jq -r '.value')

# Use in subsequent steps
api-cli navigate to "https://example.com/orders/$ORDER_ID"
api-cli assert equals "#order-status" "Processing"
```

### AI Schema Parsing

#### Command Structure Schema
```json
{
  "command_groups": {
    "assert": {
      "subcommands": ["exists", "not-exists", "equals", ...],
      "parameters": {
        "selector": "string",
        "value": "string (optional)",
        "position": "number (for selected)"
      }
    },
    "interact": {
      "subcommands": ["click", "write", "hover", ...],
      "parameters": {
        "selector": "string",
        "value": "string (for write)",
        "options": {
          "variable": "string",
          "position": "enum[TOP_LEFT, TOP_RIGHT, ...]"
        }
      }
    }
  }
}
```

#### Test Structure Schema
```json
{
  "journey": {
    "id": "string",
    "name": "string",
    "checkpoints": [{
      "id": "string",
      "name": "string",
      "position": "number",
      "steps": [{
        "id": "string",
        "type": "string",
        "meta": "object",
        "position": "number"
      }]
    }]
  }
}
```

### Advanced AI Integration

#### Dynamic Test Generation
```bash
# Generate test from natural language
echo "Test the login flow with invalid credentials" | \
  ai-cli generate-test --output yaml > login-negative-test.yaml

# Execute generated test
ai-cli run-journey login-negative-test.yaml
```

#### Pattern Recognition
```bash
# Analyze page and suggest test steps
api-cli analyze-page "https://example.com/form" --output ai | \
  jq -r '.suggested_tests[]' | \
  while read cmd; do
    eval "api-cli $cmd"
  done
```

#### Test Maintenance
```bash
# Update selectors based on page changes
api-cli update-selectors JOURNEY_ID --auto-detect --output ai

# Get maintenance suggestions
api-cli analyze-journey JOURNEY_ID --suggest-improvements --output ai
```

### Best Practices for AI Integration

1. **Use Structured Output**: Always use `--output json` or `--output ai` for parsing
2. **Session Context**: Leverage `VIRTUOSO_SESSION_ID` for sequential operations
3. **Error Handling**: Check command results before proceeding
4. **Variable Management**: Store and reuse dynamic values
5. **Batch Templates**: Use YAML for complex test structures
6. **Command Validation**: Verify commands before execution

### Changelog Integration
Recent updates affecting AI usage:
- v2.0: Added `--output ai` format with contextual information
- v2.0: Library commands for reusable test components
- v2.0: Session context for automatic position management
- v2.0: Enhanced error messages for better AI parsing

## 📖 API Spec Quick Reference

### Command Structure
The CLI follows a consistent pattern for all commands:
```
api-cli [GROUP] [SUBCOMMAND] [ARGS...] [OPTIONS]
```

### Step Types (Virtuoso API)
All commands map to specific Virtuoso step types:
- **NAVIGATE** - URL navigation
- **CLICK**, **DOUBLE_CLICK**, **RIGHT_CLICK** - Mouse interactions
- **WRITE**, **KEY** - Keyboard input
- **ASSERT_EXISTS**, **ASSERT_EQUALS**, etc. - Assertions
- **WAIT_FOR_ELEMENT**, **WAIT_TIME** - Wait operations
- **STORE_ELEMENT_TEXT**, **STORE_VALUE** - Data storage
- **DISMISS_ALERT**, **DISMISS_CONFIRM** - Dialog handling
- **SWITCH_TAB**, **SWITCH_FRAME** - Window management
- **SCROLL_POSITION**, **SCROLL_TO_ELEMENT** - Scrolling
- **SELECT_OPTION**, **SELECT_INDEX** - Dropdown selection

### Meta Field Structures
Commands use specific meta field patterns:
```json
// Click with position
{ "position": "TOP_LEFT", "elementType": "BUTTON" }

// Write with variable
{ "variable": "username", "clearBefore": true }

// Navigate with new tab
{ "newTab": true }

// Assert with comparison
{ "comparisonType": "GREATER_THAN", "value": "50" }
```

### Test Template Commands
New AI-focused commands for template management:
```bash
# Load and validate a test template
api-cli load-template examples/login-test.yaml

# Generate executable commands from template
api-cli generate-commands examples/e-commerce-test.yaml --output script > test.sh

# List available templates
api-cli get-templates ./examples --output json
```

### Response Formats
All commands support consistent response formats:
```json
{
  "id": "step_12345",
  "checkpointId": "cp_67890",
  "type": "CLICK",
  "position": 1,
  "meta": {},
  "createdAt": "2025-01-16T10:00:00Z"
}
```

## 🏗️ Architecture

### Consolidated Structure (40+ files total)
```
pkg/api-cli/
├── client/client.go        # 40+ API methods
├── commands/               # 12 command groups
│   ├── assert.go          
│   ├── interact.go        
│   ├── navigate.go        
│   └── ...
├── base.go                # Shared infrastructure
└── config/config.go       # Configuration management
```

### Key Design Principles
- **Type Safety**: All commands validated at compile time
- **Shared Infrastructure**: 60% code reduction via BaseCommand
- **AI-Friendly**: Structured output, clear command patterns
- **Extensible**: Easy to add new commands/subcommands

## 🛠️ Development

### Adding Commands
1. Add method to `client/client.go`
2. Update command group file
3. Follow existing patterns
4. Test with all output formats

### Testing
```bash
# Run full test suite
./test-consolidated-commands-final.sh

# Test specific command
./bin/api-cli assert exists "test" --dry-run
```

## 📊 Changelog

### v2.0 (2025-01-16)
- Consolidated 54 commands into 12 groups
- Added library checkpoint commands
- Fixed config loading and recursion bugs
- Achieved 98% test success rate
- Reduced codebase by 60%

### v1.0 (2025-01-14)
- Initial release with 54 individual commands
- Basic API integration
- Multiple output formats

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Follow Go conventions
4. Add tests for new features
5. Submit pull request

## 📄 License

MIT License - see LICENSE file

## 🔗 Resources

- [Virtuoso API Documentation](https://api-app2.virtuoso.qa/api)
- [Command Reference](#commands-and-variations)
- [GitHub Issues](https://github.com/ILovetoCode2100/virtuoso-api-cli/issues)