# Virtuoso MCP Server

[![CI](https://github.com/yourusername/virtuoso-mcp-server/actions/workflows/ci.yml/badge.svg)](https://github.com/yourusername/virtuoso-mcp-server/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/yourusername/virtuoso-mcp-server/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/virtuoso-mcp-server)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Model Context Protocol (MCP) server that wraps the Virtuoso API CLI, enabling Claude Desktop to create and manage Virtuoso test automation steps.

## Features

- 🎯 **12 Command Groups**: Full coverage of all Virtuoso CLI commands
- 🔄 **Session Context**: Maintains checkpoint and position state across commands
- 🛡️ **Type Safety**: Full TypeScript implementation with Zod validation
- 📝 **Rich Responses**: Formatted output with success/error indicators
- 🔧 **Configurable**: Environment-based configuration
- 🐛 **Debug Mode**: Optional debug logging for troubleshooting

## Installation

### Prerequisites

1. **Virtuoso API CLI**: The compiled `api-cli` binary
2. **Node.js**: Version 18 or higher
3. **Virtuoso API Key**: Valid API credentials

### Setup Steps

1. **Clone and Install**

   ```bash
   cd virtuoso-mcp-server
   npm install
   ```

2. **Configure Environment**

   ```bash
   cp .env.example .env
   # Edit .env with your paths
   ```

3. **Build the Server**

   ```bash
   npm run build
   ```

4. **Test the Server**
   ```bash
   npm test
   ```

## Claude Desktop Configuration

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "virtuoso": {
      "command": "node",
      "args": ["/absolute/path/to/virtuoso-mcp-server/dist/index.js"],
      "env": {
        "VIRTUOSO_CLI_PATH": "/path/to/virtuoso-GENerator/bin/api-cli",
        "VIRTUOSO_CONFIG_PATH": "/path/to/.api-cli/virtuoso-config.yaml"
      }
    }
  }
}
```

## Available Tools

### Core Testing Tools

1. **`virtuoso_assert`** - Create assertion steps

   - Supports: exists, not-exists, equals, comparisons, regex matching
   - Example: "Assert that 'Login button' exists"

2. **`virtuoso_interact`** - User interaction steps

   - Supports: click, double-click, right-click, hover, write, key press
   - Example: "Click on 'Submit' button"

3. **`virtuoso_navigate`** - Navigation steps

   - Supports: URL navigation, scrolling
   - Example: "Navigate to 'https://example.com'"

4. **`virtuoso_data`** - Data management steps

   - Supports: store text/value, cookie operations
   - Example: "Store text from 'Username' field in variable 'user'"

5. **`virtuoso_wait`** - Wait steps
   - Supports: wait for element, wait for time
   - Example: "Wait for '#loader' element"

### Context Management

- **`virtuoso_set_context`** - Set session context (checkpoint ID, position)
- **`virtuoso_get_context`** - Get current session context

## Usage Examples

### In Claude Desktop

Once configured, you can use natural language:

1. "Set context to checkpoint 1681532"
2. "Assert that 'Login button' exists"
3. "Click on 'Email' field"
4. "Write 'test@example.com' in the current field"
5. "Click 'Submit' button"

### Session Context

The server maintains context across commands:

```
User: "Set context to checkpoint 1681532"
Assistant: ✅ Context updated: { checkpointId: "1681532", position: 1 }

User: "Assert 'Login' button exists"
Assistant: ✅ Created assertion: Element "Login" exists (Step 1)
// Position auto-increments to 2

User: "Click the Login button"
Assistant: ✅ Created click action on "Login" (Step 2)
// Position auto-increments to 3
```

## Development

### Project Structure

```
virtuoso-mcp-server/
├── src/
│   ├── index.ts              # Entry point
│   ├── server.ts             # MCP server
│   ├── cli-wrapper.ts        # CLI wrapper
│   ├── tools/               # Tool implementations
│   │   ├── assert.ts
│   │   ├── interact.ts
│   │   └── ...
│   ├── types/               # TypeScript types
│   └── utils/              # Utilities
├── dist/                    # Compiled output
├── package.json
├── tsconfig.json
└── README.md
```

### Adding New Tools

1. Create tool file in `src/tools/`
2. Define input schema with Zod
3. Implement command building logic
4. Register in `server.ts`

### Testing

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run comprehensive test suite
./scripts/run-all-tests.sh

# Run specific test suites
npm run lint                 # TypeScript linting
npm run test:ci              # CI-optimized test run
npm run validate             # Validate all tools
npm run test:server          # Test MCP server integration

# Test with debug logging
DEBUG=true npm run dev

# Build for production
npm run build:prod
```

#### Test Organization

The project includes several types of tests:

1. **Unit Tests** (`src/__tests__/`)

   - Tool-specific tests for each command group
   - Utility function tests
   - CLI wrapper tests

2. **Integration Tests** (`src/__tests__/integration/`)

   - End-to-end MCP protocol tests
   - Server integration tests

3. **Validation Tests** (`scripts/validate-tools.ts`)

   - Validates all 12 tool groups
   - Ensures proper schema definitions

4. **Manual Tests** (`test/manual-test.md`)
   - Step-by-step manual testing procedures
   - Used for final validation

#### Running the Comprehensive Test Suite

The comprehensive test runner provides detailed reporting:

```bash
# Run all tests with detailed reporting
./scripts/run-all-tests.sh

# Output includes:
# - TypeScript compilation check
# - Build validation
# - Unit tests with coverage
# - Tool validation
# - Server integration tests
# - Comprehensive test report
```

Test reports are saved in `test-reports/` with timestamps.

#### CI/CD Pipeline

The project uses GitHub Actions for continuous integration:

- **On Pull Request**: Linting, unit tests, build validation
- **On Push to Main**: Full test suite including integration tests
- **Release Process**: Automated releases with `[release]` commit tag

#### Coverage Requirements

The project maintains strict coverage thresholds:

- Statements: 80%
- Branches: 80%
- Functions: 80%
- Lines: 80%

View coverage reports:

```bash
npm run test:coverage
# Open coverage/lcov-report/index.html in browser
```

## Configuration

### Environment Variables

- `VIRTUOSO_CLI_PATH` - Path to api-cli binary (required)
- `VIRTUOSO_CONFIG_PATH` - Path to virtuoso-config.yaml
- `DEBUG` - Enable debug logging (true/false)
- `NODE_ENV` - Environment (development/production)

### Virtuoso Configuration

Create `~/.api-cli/virtuoso-config.yaml`:

```yaml
api:
  auth_token: your-api-key-here
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"
headers:
  X-Virtuoso-Client-ID: "api-cli-generator"
  X-Virtuoso-Client-Name: "api-cli-generator"
```

## Troubleshooting

### Common Issues

1. **"CLI not found"**

   - Check `VIRTUOSO_CLI_PATH` in .env
   - Ensure api-cli is built and executable

2. **"Authentication failed"**

   - Verify API token in virtuoso-config.yaml
   - Check organization ID is correct

3. **"Command not recognized"**
   - Ensure using latest tool names
   - Check Claude Desktop config is correct

### Debug Mode

Enable debug logging:

```bash
DEBUG=true node dist/index.js
```

## License

MIT
