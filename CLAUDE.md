# Virtuoso API CLI - Claude Development Guide

## Project Overview

The Virtuoso API CLI is a Go-based command-line tool that provides an AI-friendly interface for Virtuoso's test automation platform. This CLI enables programmatic creation and management of automated tests through a consistent, well-structured command interface.

**Version:** 2.0
**Status:** Production Ready (98% test success rate)
**Language:** Go 1.21+

## Architecture

### Project Structure

```
virtuoso-GENerator/
├── cmd/api-cli/           # Main entry point
├── pkg/api-cli/           # Core implementation
│   ├── client/           # API client (40+ methods)
│   ├── commands/         # 12 command groups
│   └── config/           # Configuration management
├── bin/                  # Compiled binary output
└── examples/             # YAML test templates
```

### Key Components

1. **Command Groups** (12 total):

   - `assert` - Validation commands (exists, equals, gt, etc.)
   - `interact` - User interactions (click, write, hover, etc.)
   - `navigate` - Navigation commands (to, scroll, back, etc.)
   - `data` - Data management (store, cookies)
   - `dialog` - Dialog handling (alerts, confirms, prompts)
   - `wait` - Wait operations (element, time)
   - `window` - Window management (resize, switch)
   - `mouse` - Mouse operations (move, drag)
   - `select` - Dropdown operations
   - `file` - File upload
   - `misc` - Miscellaneous (comments, execute JS)
   - `library` - Reusable test components

2. **Shared Infrastructure**:

   - `BaseCommand` struct provides common functionality
   - Reduces code duplication by ~60%
   - Consistent error handling and output formatting

3. **Configuration**:
   - Uses Viper for flexible config management
   - Config file: `~/.api-cli/virtuoso-config.yaml`
   - Environment variable overrides supported

## Development Guidelines

### Adding New Commands

1. **Update the Client** (`pkg/api-cli/client/client.go`):

   ```go
   func (c *Client) YourNewMethod(params) (*Response, error) {
       // Implementation
   }
   ```

2. **Add to Command Group** (e.g., `pkg/api-cli/commands/interact.go`):

   ```go
   case "your-command":
       return i.handleYourCommand(args)
   ```

3. **Follow Patterns**:
   - Use `BaseCommand` for shared functionality
   - Support all output formats (human, json, yaml, ai)
   - Include proper error messages
   - Add position management for test steps

### Testing

Run the comprehensive test suite:

```bash
make test
./test-consolidated-commands-final.sh
```

Test individual commands:

```bash
./bin/api-cli assert exists "test" --dry-run
```

### Code Style

- Follow standard Go conventions
- Use meaningful variable names
- Add comments for complex logic
- Maintain consistent error messages
- Keep functions focused and small

## Important Patterns

### 1. Command Structure

All commands follow: `api-cli [GROUP] [SUBCOMMAND] [ARGS...] [OPTIONS]`

### 2. Output Formats

- `--output human` - Human-readable (default)
- `--output json` - Structured JSON
- `--output yaml` - YAML format
- `--output ai` - AI-optimized with context

### 3. Session Management

Commands can use session context for automatic position tracking:

```bash
export VIRTUOSO_SESSION_ID=checkpoint_id
# Commands auto-increment position
```

### 4. Error Handling

- Always return structured errors
- Include helpful error messages
- Support --dry-run for testing

## Related Projects

### MCP Server

The Model Context Protocol (MCP) server has been moved to a separate repository at `/Users/marklovelady/_dev/_projects/virtuoso-mcp-server`. It provides:

- Bridge between Claude Desktop and this CLI
- TypeScript/Node.js implementation
- Exposes all CLI commands as MCP tools

The MCP server depends on the compiled binary from this project (`bin/api-cli`) but is otherwise independent.

## Common Tasks

### Building

```bash
make build
# Output: bin/api-cli
```

### Running Commands

```bash
# Basic assertion
./bin/api-cli assert exists "Login button"

# Complex interaction
./bin/api-cli interact click "#submit" --position CENTER --element-type BUTTON

# Navigation with options
./bin/api-cli navigate to "https://example.com" --new-tab
```

### Configuration

Create `~/.api-cli/virtuoso-config.yaml`:

```yaml
api:
  auth_token: your-api-key-here
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"
```

## Debugging Tips

1. **Enable Debug Output**:

   ```bash
   export DEBUG=true
   ./bin/api-cli [command]
   ```

2. **Check API Responses**:
   Use `--output json` to see raw API responses

3. **Dry Run Mode**:
   Use `--dry-run` to test commands without making API calls

4. **Session Info**:
   ```bash
   ./bin/api-cli get-session-info --output json
   ```

## Key Files to Understand

1. **`pkg/api-cli/commands/base.go`** - Shared command infrastructure
2. **`pkg/api-cli/client/client.go`** - API client implementation
3. **`pkg/api-cli/commands/register.go`** - Command registration
4. **`pkg/api-cli/config/config.go`** - Configuration management
5. **`cmd/api-cli/main.go`** - Entry point and CLI setup

## Notes for AI Development

- This CLI is designed to be AI-friendly with structured outputs
- The `--output ai` format includes context and suggestions
- Commands map directly to Virtuoso API step types
- All commands support consistent meta field patterns
- Test templates in `examples/` show common patterns

## Maintenance

- Keep commands consolidated in their groups
- Maintain backward compatibility
- Update tests when adding features
- Document new command variations
- Follow the existing error message patterns
