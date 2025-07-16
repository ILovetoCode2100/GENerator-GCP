# Virtuoso API CLI - AI-Optimized Documentation

## 🎯 Project Overview

This is an AI-optimized CLI tool for the Virtuoso API, designed specifically for AI-driven test automation. The repository has been optimized to reduce lines of code by 60% while enhancing AI integration capabilities.

## 📊 Optimization Results (January 2025)

### Code Reduction
- **Files**: 75+ → 51 (32% reduction)
- **Lines of Code**: ~60% reduction through consolidation
- **Commands**: 54 individual → 12 logical groups
- **Documentation**: 7 files → 1 comprehensive README
- **Test Success Rate**: 98% (55/56 commands)

### Key Improvements
- Consolidated command structure for AI clarity
- Enhanced output formats with AI context
- Viper configuration for flexible test building
- Template-driven test generation system
- Intelligent next-step suggestions

## 🤖 AI Integration Features

### 1. Enhanced Output Format
The `--output ai` format provides contextual information for test building:
```json
{
  "command": "assert exists",
  "result": "success",
  "context": {
    "checkpoint_id": "1680930",
    "position": 1,
    "max_steps": 20
  },
  "next_steps": [
    "interact click 'Login button'",
    "wait element '#login-form'"
  ],
  "description": "Verify login button is present on page"
}
```

### 2. Viper Configuration System
Centralized configuration management for AI-driven testing:

#### Test Configuration
```yaml
test:
  batch_dir: ./test-batches          # AI places test definitions here
  output_format: ai                  # Default to AI-friendly output
  template_dir: ./examples           # Reference test patterns
  auto_validate: true                # Catch errors early
  max_steps_per_checkpoint: 20       # Maintainable test structure
```

#### AI Configuration
```yaml
ai:
  enable_suggestions: true           # Intelligent next steps
  context_depth: 3                   # Balance detail vs size
  auto_generate_descriptions: true   # Self-documenting tests
  template_inference: true           # Learn from patterns
```

### 3. Template-Driven Testing
YAML templates for complex test scenarios:
```yaml
journey:
  name: "E-Commerce Purchase Flow"
  checkpoints:
    - name: "Product Search"
      steps:
        - command: navigate to
          args: ["https://shop.example.com"]
        - command: interact write
          args: ["#search", "wireless headphones"]
        - command: assert gte
          args: [".results-count", "1"]
```

### 4. Command Consolidation
12 logical groups for AI understanding:
- **assert** (12 subcommands): Element state verification
- **interact** (6 subcommands): User actions
- **navigate** (5 subcommands): Page navigation
- **data** (5 subcommands): Variable/cookie management
- **dialog** (4 subcommands): Alert handling
- **wait** (2 subcommands): Timing control
- **window** (5 subcommands): Tab/frame management
- **mouse** (6 subcommands): Advanced interactions
- **select** (3 subcommands): Dropdown operations
- **file** (1 subcommand): Upload handling
- **misc** (3 subcommands): Comments/scripts
- **library** (6 subcommands): Reusable components

## 🏗️ Architecture for AI

### Simplified Structure
```
pkg/api-cli/
├── client/          # 40+ API methods
├── commands/        # 12 consolidated groups
├── config/          # Viper configuration
└── base.go         # Shared infrastructure
```

### Key Design Principles
1. **Clear Command Patterns**: Consistent structure for AI parsing
2. **Type Safety**: Compile-time validation
3. **Shared Infrastructure**: 60% code reduction
4. **Extensible Design**: Easy to add features

## 🚀 AI Usage Patterns

### Basic Test Generation
```bash
# Single command with AI output
api-cli assert exists "Login" --output ai

# Session-based sequential testing
export VIRTUOSO_SESSION_ID=$CHECKPOINT_ID
api-cli navigate to "https://example.com"  # Position 1
api-cli assert exists "header"              # Position 2
api-cli interact click "Login"              # Position 3
```

### Batch Processing
```bash
# Process YAML templates from configured directory
api-cli load-template login-test.yaml
api-cli generate-commands e-commerce.yaml --output script

# List available templates
api-cli get-templates ./examples --output json
```

### Dynamic Test Building
```bash
# Parse AI suggestions for next steps
RESULT=$(api-cli assert exists "#banner" --output json)
NEXT_STEPS=$(echo $RESULT | jq -r '.next_steps[]')

# Execute suggested steps
for step in $NEXT_STEPS; do
  api-cli $step
done
```

## 📋 Recent Updates (January 2025)

### Viper Enhancement (Latest)
- Refactored configuration into centralized `loadConfig()` function
- Added `TestConfig` and `AIConfig` structures
- Enhanced inline documentation for AI understanding
- Added helper methods for AI-specific settings
- Updated README with configuration guide

### Previous Optimizations
- Consolidated 54 commands into 12 groups
- Added AI output format with context
- Created template processing system
- Implemented session management
- Fixed config loading issues
- Added library checkpoint commands

## 🔧 Configuration Best Practices

### For AI Systems
1. **Set output format to "ai"** for context-aware responses
2. **Use batch_dir** for template organization
3. **Enable suggestions** for intelligent workflows
4. **Set appropriate context_depth** (3 recommended)
5. **Use auto_validate** to catch errors early

### Environment Variables
```bash
export VIRTUOSO_API_TOKEN=your-token
export VIRTUOSO_TEST_OUTPUT_FORMAT=ai
export VIRTUOSO_AI_ENABLE_SUGGESTIONS=true
```

## 📖 Command Reference

### Most Used by AI
1. **assert exists/equals** - Verify UI state
2. **interact click/write** - User actions
3. **navigate to** - Page navigation
4. **wait element** - Timing control
5. **data store** - Variable management

### Advanced Patterns
- **Conditional execution** based on element presence
- **Variable extraction** and reuse
- **Library checkpoints** for reusable components
- **Session context** for sequential operations

## 🧪 Testing and Validation

### Test Coverage
- 98% success rate (55/56 commands)
- All command groups tested
- Multiple output formats verified
- Session management validated

### Known Limitations
- execute-script command (security considerations)
- Some edge cases in regex matching
- Limited to Virtuoso API constraints

## 🎯 Future Enhancements

### Planned Features
- Natural language to test conversion
- AI-powered selector generation
- Automatic test maintenance
- Pattern recognition improvements
- Enhanced error recovery

### Community Contributions
- Fork and enhance AI capabilities
- Add new command variations
- Improve template examples
- Enhance documentation

## 📚 Resources

### Documentation
- **README.md** - Comprehensive user guide
- **CLAUDE.md** - This AI-focused documentation
- **config.yaml** - Configuration reference
- **examples/** - Test template library

### Support
- GitHub Issues: Report bugs or request features
- Pull Requests: Contribute improvements
- Discussions: Share AI integration patterns

## 🔑 Key Takeaways for AI

1. **Use structured commands**: `api-cli [group] [subcommand] [args]`
2. **Leverage AI output format**: Rich context and suggestions
3. **Utilize templates**: YAML for complex test scenarios
4. **Configure appropriately**: Viper settings for optimal AI usage
5. **Chain commands**: Session context for workflows

---

**Version**: 2.0 (AI-Optimized)  
**Last Updated**: January 2025  
**Optimization Level**: 60% LOC reduction  
**AI Integration**: Enhanced with Viper configuration