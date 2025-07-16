# Virtuoso MCP Server Scripts

This directory contains build, test, and validation scripts for the Virtuoso MCP Server.

## Scripts Overview

### 🧪 test-server.ts

Test script that simulates Claude Desktop interaction with the MCP server.

**Usage:**

```bash
npm run test:server
# or
npx tsx scripts/test-server.ts
```

**Features:**

- Connects to MCP server as a client
- Tests all 12 tool groups with sample calls
- Validates responses and error handling
- Provides colored output for easy debugging
- Reports success/failure statistics

### ✅ validate-tools.ts

Validation script that checks all tool schemas for consistency and generates documentation.

**Usage:**

```bash
npm run validate
# or
npx tsx scripts/validate-tools.ts
```

**Features:**

- Validates tool naming conventions
- Checks for duplicate tool names
- Ensures description quality
- Generates TOOLS.md documentation
- Reports validation errors and warnings

### 🔨 build.sh

Production build script that compiles TypeScript and creates a distribution package.

**Usage:**

```bash
npm run build:prod
# or
./scripts/build.sh
```

**Features:**

- Cleans previous builds
- Runs linting and tests
- Compiles TypeScript
- Creates distribution package
- Generates tarball for deployment
- Includes build metadata

### 🚀 quick-test.sh

Quick validation script that runs both validation and server tests.

**Usage:**

```bash
./scripts/quick-test.sh
```

**Features:**

- Checks for required configuration
- Builds server if needed
- Runs validation
- Runs integration tests
- Provides quick health check

## Configuration Requirements

All scripts require a valid Virtuoso configuration file at:

```
~/.api-cli/virtuoso-config.yaml
```

Example configuration:

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

## Development Workflow

1. **Make changes** to tool implementations in `src/tools/`
2. **Run validation** to ensure consistency: `npm run validate`
3. **Test locally** with the test server: `npm run test:server`
4. **Build for production**: `npm run build:prod`
5. **Deploy** the generated tarball or dist directory

## Script Output

- **test-server.ts**: Console output with test results
- **validate-tools.ts**: Console output + generates `TOOLS.md`
- **build.sh**: Creates `dist/` directory and `.tgz` package
- **quick-test.sh**: Combined output from validation and testing

## Troubleshooting

### Config not found

Ensure `~/.api-cli/virtuoso-config.yaml` exists with valid API credentials.

### Build failures

1. Check Node.js version (requires 18+)
2. Run `npm install` to ensure dependencies
3. Check for TypeScript errors with `npm run lint`

### Test failures

1. Verify API credentials are valid
2. Check network connectivity to Virtuoso API
3. Review error messages for specific issues

## Adding New Tools

When adding new tools:

1. Add tool implementation in `src/tools/`
2. Run `npm run validate` to check naming
3. Add test case in `test-server.ts`
4. Update documentation as needed
