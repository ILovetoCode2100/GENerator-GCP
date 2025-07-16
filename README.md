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

#### Single Test Flow
```bash
# Create journey with checkpoints
api-cli create-journey PROJECT_ID "Login Test"
api-cli create-checkpoint JOURNEY_ID "Setup" 1
api-cli assert exists "Login form" --checkpoint CHECKPOINT_ID 1
api-cli interact write "#username" "test@example.com" --checkpoint CHECKPOINT_ID 2
api-cli interact click "Submit" --checkpoint CHECKPOINT_ID 3
```

#### Batch Test Generation (YAML Template)
Create `test-journey.yaml`:
```yaml
journey:
  name: "E2E User Registration"
  project_id: 13961
  checkpoints:
    - name: "Navigate to Registration"
      steps:
        - command: navigate to
          args: ["https://example.com/register"]
          options: {new_tab: true}
        - command: wait element
          args: ["#registration-form"]
          options: {timeout: 5000}
    
    - name: "Fill Registration Form"
      steps:
        - command: interact write
          args: ["#email", "test@example.com"]
          options: {variable: "user_email"}
        - command: interact write
          args: ["#password", "SecurePass123!"]
        - command: interact click
          args: ["#terms-checkbox"]
        - command: assert checked
          args: ["#terms-checkbox"]
    
    - name: "Submit and Verify"
      steps:
        - command: interact click
          args: ["#submit-button"]
        - command: wait element
          args: [".success-message"]
        - command: assert equals
          args: [".welcome-text", "Welcome, test@example.com"]
```

Process with AI:
```bash
# Parse YAML and generate commands
ai-cli process-journey test-journey.yaml --output ai

# Or use direct command generation
ai-cli generate-from-yaml test-journey.yaml | bash
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

## 🏗️ Architecture

### Consolidated Structure (40 files total)
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