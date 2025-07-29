# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build & Development

```bash
# Build the CLI binary
make build

# Run all quality checks and build
make check

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean
```

### Testing

```bash
# Unit tests
make test
go test -v ./pkg/api-cli/commands/...  # Test specific package

# Integration tests (requires API access)
./test-scripts/test-all-69-commands.sh [checkpoint-id]    # Full test suite
./test-all-commands-simple.sh [checkpoint-id]              # Quick validation
./test-commands/test-yaml-end-to-end.sh                   # YAML functionality

# Make targets for specific tests
make test-commands       # Test CLI commands
make test-library        # Test library commands
```

### Common Development Tasks

```bash
# Create and run a test
./bin/api-cli run-test test.yaml

# Use session context for multiple commands
export VIRTUOSO_SESSION_ID=12345
./bin/api-cli step-navigate to "https://example.com"
./bin/api-cli step-interact click "button"

# Get command help
./bin/api-cli --help
./bin/api-cli step-assert --help
```

## Architecture

### Codebase Structure

The project follows a consolidated architecture where related commands are grouped into single files:

- **Entry Point**: `cmd/api-cli/main.go` - CLI initialization
- **Core Package**: `pkg/api-cli/` - Main implementation
  - `client/` - API client with 120+ methods using context-aware patterns
  - `commands/` - ~20 files (reduced from 35+) containing all CLI commands
  - `config/` - Configuration management using Viper
  - `constants/` - Shared constants and types
  - `yaml-layer/` - YAML test definition parsing and execution

### Command Consolidation Pattern

Commands are organized into logical groups to reduce code duplication:

1. **`interaction_commands.go`** - All user interactions (click, write, mouse, select)
2. **`browser_commands.go`** - Browser operations (navigate, scroll, window)
3. **`list.go`** - Generic list operations for all entities
4. **`project_management.go`** - CRUD operations for projects/goals/journeys
5. **`execution_management.go`** - Test execution workflow

Individual step commands remain in separate files (`step_*.go`) for clarity.

### Key Patterns

#### Command Structure

All commands follow the unified positional syntax:

```
api-cli <command> <subcommand> [checkpoint-id] <args...> [position]
```

Commands implement the `StepCommand` interface and extend `BaseCommand` for shared functionality.

#### Session Context

The CLI supports session-based checkpoint management via `VIRTUOSO_SESSION_ID` environment variable, eliminating the need to specify checkpoint IDs repeatedly.

#### Error Handling

Structured error types (`APIError`, `ClientError`) with consistent exit codes:

- 0: Success
- 1: General error
- 3: Authentication error
- 5: Not found error

#### Output Formats

All commands support multiple output formats via the `--output` flag:

- `human` - Default readable format
- `json` - Structured data
- `yaml` - Configuration format
- `ai` - AI-optimized with context

### YAML Test Layer

The `run-test` command provides a simplified interface for test creation:

- Auto-creates all required infrastructure (project, goal, journey, checkpoint)
- Supports multiple input formats (simplified, extended, compact)
- Progressive disclosure from simple to complex test definitions
- Comprehensive validation with helpful error messages

## Important Implementation Details

### API Client

- All methods are context-aware for timeout/cancellation support
- Automatic retry logic for transient failures
- Structured response handling with type safety
- Session management handled transparently

### Command Validation

The CLI includes an intelligent validator that:

- Auto-corrects common syntax errors (missing hyphens, deprecated commands)
- Validates flag compatibility
- Provides migration guidance for deprecated features
- Handles format conversions automatically

### Variable Handling

- Variables in commands should NOT include the `$` prefix (added automatically by the API)
- Store operations create variables that can be referenced in subsequent steps
- Variable names should be descriptive and follow camelCase convention

### Known Limitations

1. File upload commands only support URLs, not local file paths
2. Some browser navigation commands (back, forward, refresh) are not supported by the API
3. Window close and frame switching by index/name are not available
4. Library commands use checkpoint IDs (not journey IDs) for the `add` operation

## Configuration

The CLI uses a hierarchical configuration system:

1. CLI flags (highest priority)
2. Environment variables
3. Config file (`~/.api-cli/virtuoso-config.yaml`)
4. Default values

Key configuration:

```yaml
api:
  auth_token: your-api-key-here
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"
```

## Testing Guidelines

When adding new features:

1. Add unit tests in the appropriate `*_test.go` file
2. Update integration tests in `test-scripts/test-all-69-commands.sh`
3. Test all output formats (human, json, yaml, ai)
4. Verify session context support
5. Ensure backward compatibility with legacy syntax
6. Add examples to `examples/` directory

## Recent Changes (January 2025)

### Major Improvements

- **Context Support**: 80+ context-aware methods for better reliability
- **Command Validator**: Auto-correction of common syntax errors
- **Unified Test Runner**: `run-test` command for simplified test creation
- **Code Consolidation**: 43% file reduction through logical grouping
- **100% Success Rate**: All 69 commands tested and working

### Migration Notes

- Dialog commands now use hyphenated syntax (e.g., `dismiss-alert` instead of `alert accept`)
- Mouse and select commands moved under `step-interact` parent command
- Wait time commands expect milliseconds (auto-conversion from seconds)
- Store commands simplified (e.g., `store element-text` → `store text`)

## Update: 2025-07-29 20:17:20

### Changes Summary

- Added: 121 files
- Modified: 1 files
- Deleted: 0 files

### Repository: virtuoso-GENerator

### Modified Components

```
CLAUDE.md                                          | 554 ++++++---------------
 pkg/api-cli/client/client.go                       |  41 +-
 pkg/api-cli/client/client_fixes.go                 | 115 -----
 pkg/api-cli/client/execute_goal_robust.go          | 205 --------
 pkg/api-cli/client/response_handler.go             |  10 +-
 pkg/api-cli/client/response_handler_integration.go | 209 --------
 pkg/api-cli/commands/command_validator.go          |   3 +-
 pkg/api-cli/commands/execute_goal_fixed.go         | 152 ------
 pkg/api-cli/commands/manage_lists.go               |   6 +-
 pkg/api-cli/commands/register.go                   |   4 +-
 10 files changed, 180 insertions(+), 1119 deletions(-)
```

### Notes for Claude Code

- Automated commit at 2025-07-29 20:17:20
- Security scan passed
- All changes reviewed
