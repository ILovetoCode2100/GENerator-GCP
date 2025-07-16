# Virtuoso MCP Server Test Suite

This directory contains the comprehensive test suite for the Virtuoso MCP Server.

## Test Structure

```
__tests__/
├── cli-wrapper.test.ts      # Tests for VirtuosoCliWrapper class
├── server.test.ts           # Tests for VirtuosoMcpServer class
├── tools/                   # Tests for tool implementations
│   ├── assert.test.ts       # Assert tool tests
│   └── ...                  # Other tool tests
└── utils/                   # Tests for utility functions
    ├── formatting.test.ts   # Formatting utility tests
    └── ...                  # Other utility tests
```

## Running Tests

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run tests in CI mode
npm run test:ci
```

## Writing Tests

### Test File Naming

- Test files should be named `[module-name].test.ts`
- Place tests in the same directory structure as the source files

### Test Structure

```typescript
import { jest } from "@jest/globals";

describe("Module Name", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe("Function/Method Name", () => {
    test("should do something", () => {
      // Arrange
      const input = "test";

      // Act
      const result = functionUnderTest(input);

      // Assert
      expect(result).toBe("expected");
    });
  });
});
```

### Mocking

#### Mocking Modules

```typescript
jest.mock("../module-path.js", () => ({
  exportedFunction: jest.fn(),
}));
```

#### Mocking child_process

```typescript
jest.mock("child_process", () => ({
  spawn: jest.fn(),
}));
```

#### Creating Mock Processes

```typescript
const mockProcess = new EventEmitter();
mockProcess.stdout = new EventEmitter();
mockProcess.stderr = new EventEmitter();
mockSpawn.mockReturnValue(mockProcess);
```

### Common Test Patterns

#### Testing CLI Commands

```typescript
test("should execute CLI command", async () => {
  const executePromise = wrapper.execute(["command", "arg"]);

  // Simulate process output
  mockProcess.stdout.emit("data", JSON.stringify({ success: true }));
  mockProcess.emit("close", 0);

  const result = await executePromise;
  expect(result.success).toBe(true);
});
```

#### Testing Tool Registration

```typescript
test("should register tool handler", () => {
  registerTool(mockServer, mockCli);

  expect(mockServer.setRequestHandler).toHaveBeenCalledWith(
    CallToolRequestSchema,
    expect.any(Function),
  );
});
```

#### Testing Error Handling

```typescript
test("should handle errors gracefully", async () => {
  mockFunction.mockRejectedValue(new Error("Test error"));

  const result = await functionUnderTest();

  expect(result.success).toBe(false);
  expect(result.error).toContain("Test error");
});
```

## Coverage Requirements

The project maintains the following coverage thresholds:

- **Branches**: 80%
- **Functions**: 80%
- **Lines**: 80%
- **Statements**: 80%

Coverage reports are generated in the `coverage/` directory:

- `coverage/lcov-report/index.html` - HTML coverage report
- `coverage/lcov.info` - LCOV format for CI tools

## Debugging Tests

### Running a Single Test File

```bash
NODE_OPTIONS='--experimental-vm-modules' jest src/__tests__/cli-wrapper.test.ts
```

### Running Tests with Verbose Output

```bash
NODE_OPTIONS='--experimental-vm-modules' jest --verbose
```

### Debugging with VS Code

Add this configuration to `.vscode/launch.json`:

```json
{
  "type": "node",
  "request": "launch",
  "name": "Jest Debug",
  "program": "${workspaceFolder}/node_modules/.bin/jest",
  "args": ["--runInBand", "${relativeFile}"],
  "env": {
    "NODE_OPTIONS": "--experimental-vm-modules"
  },
  "console": "integratedTerminal",
  "internalConsoleOptions": "neverOpen"
}
```

## Continuous Integration

Tests are automatically run in CI with:

```bash
npm run test:ci
```

This command:

- Runs in CI mode (no watch, no interactive)
- Generates coverage reports
- Limits workers to 2 for consistent CI performance

## Test Helpers

Global test helpers are available via `jest.setup.js`:

```typescript
// Create mock responses
const response = global.testHelpers.createMockResponse({ id: 1 });

// Create mock errors
const error = global.testHelpers.createMockError("Test failed");
```

## Best Practices

1. **Clear mocks between tests** - Use `jest.clearAllMocks()` in `beforeEach`
2. **Test both success and failure paths** - Include error cases
3. **Use descriptive test names** - Should read like documentation
4. **Follow AAA pattern** - Arrange, Act, Assert
5. **Mock external dependencies** - Don't make real API calls
6. **Test edge cases** - Empty arrays, null values, etc.
7. **Keep tests focused** - One assertion per test when possible
8. **Use test.each for similar tests** - Reduce duplication
