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

## 🤖 AI Usage

### Output Formats
All commands support AI-optimized output:
```bash
--output human  # Default readable format
--output json   # Structured data
--output yaml   # Configuration format
--output ai     # Conversational AI format
```

### Building Tests
Create Virtuoso test structures:
```bash
# Create journey with checkpoints
api-cli create-journey PROJECT_ID "Login Test"
api-cli create-checkpoint JOURNEY_ID "Setup" 1
api-cli assert exists "Login form" --checkpoint CHECKPOINT_ID 1
api-cli interact write "#username" "test@example.com" --checkpoint CHECKPOINT_ID 2
api-cli interact click "Submit" --checkpoint CHECKPOINT_ID 3
```

### Batch Operations
Use session context for sequential commands:
```bash
export VIRTUOSO_SESSION_ID=CHECKPOINT_ID
api-cli assert exists "Header"  # Position 1
api-cli interact click "Login"   # Position 2
api-cli wait element "#form"    # Position 3
```

### Command Chaining
Parse AI output for test generation:
```bash
# Get command in AI format
api-cli assert exists "Button" --output ai | jq -r '.next_steps[]'

# Chain commands programmatically
RESULT=$(api-cli data store element-text "#user" "username" --output json)
USERNAME=$(echo $RESULT | jq -r '.variable_value')
```

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