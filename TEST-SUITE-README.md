# Virtuoso API CLI Test Suite

This directory contains comprehensive test YAML definitions for the Virtuoso API CLI. The test suite is designed to validate all 69 commands, error handling, edge cases, and YAML parsing.

## Test Files

### 1. test-all-commands-positive.yaml

**Purpose**: Test all 69 commands with valid inputs

- Tests every command variant with proper syntax
- Includes all parameter combinations
- Uses variables to demonstrate interpolation
- Validates both simple and complex command forms
- Expected outcome: All commands should execute successfully

### 2. test-negative-cases.yaml

**Purpose**: Test error conditions and invalid inputs

- Missing required fields
- Invalid parameter values
- Malformed selectors
- Type mismatches
- Conflicting options
- Expected outcome: Proper error messages with guidance

### 3. test-edge-cases.yaml

**Purpose**: Test boundary conditions and unusual inputs

- Maximum/minimum values
- Very long strings
- Special characters and Unicode
- Empty values
- Complex selectors
- Numeric boundaries
- Expected outcome: Graceful handling of edge cases

### 4. test-yaml-validation.yaml

**Purpose**: Test YAML parsing and validation

- Invalid YAML syntax
- Missing required fields
- Type validation
- Structure validation
- Special YAML features (anchors, references)
- Malformed structures
- Expected outcome: Clear parsing errors with line numbers

## Running the Tests

### Individual Test Execution

```bash
# Run positive tests
./bin/api-cli run-test test-all-commands-positive.yaml

# Run with dry run to validate without execution
./bin/api-cli run-test test-all-commands-positive.yaml --dry-run

# Run negative tests (expect errors)
./bin/api-cli run-test test-negative-cases.yaml

# Run edge cases
./bin/api-cli run-test test-edge-cases.yaml

# Test YAML validation (expect parsing errors)
./bin/api-cli run-test test-yaml-validation.yaml
```

### Automated Test Suite

```bash
# Create a test runner script
cat > run-all-tests.sh << 'EOF'
#!/bin/bash

echo "Running Virtuoso API CLI Test Suite"
echo "===================================="

# Positive tests (should succeed)
echo -e "\n1. Running positive tests..."
./bin/api-cli run-test test-all-commands-positive.yaml --dry-run
if [ $? -eq 0 ]; then
    echo "✓ Positive tests passed"
else
    echo "✗ Positive tests failed"
fi

# Negative tests (should fail with proper errors)
echo -e "\n2. Running negative tests..."
./bin/api-cli run-test test-negative-cases.yaml --dry-run 2>&1 | grep -q "error"
if [ $? -eq 0 ]; then
    echo "✓ Negative tests properly reported errors"
else
    echo "✗ Negative tests did not report expected errors"
fi

# Edge cases (should handle gracefully)
echo -e "\n3. Running edge case tests..."
./bin/api-cli run-test test-edge-cases.yaml --dry-run
if [ $? -eq 0 ]; then
    echo "✓ Edge cases handled successfully"
else
    echo "✗ Edge cases encountered issues"
fi

# YAML validation (should report parsing errors)
echo -e "\n4. Running YAML validation tests..."
./bin/api-cli run-test test-yaml-validation.yaml --dry-run 2>&1 | grep -q "error"
if [ $? -eq 0 ]; then
    echo "✓ YAML validation properly reported errors"
else
    echo "✗ YAML validation did not report expected errors"
fi

echo -e "\nTest suite complete!"
EOF

chmod +x run-all-tests.sh
```

## Expected Behaviors

### Successful Commands

- Clear output showing step creation
- Proper variable interpolation
- Correct parameter handling
- Appropriate default values

### Error Conditions

- User-friendly error messages
- Specific guidance on fixing issues
- Proper exit codes
- No stack traces for user errors

### Edge Cases

- Graceful degradation
- Sensible limits
- Proper escaping and encoding
- Consistent behavior

### YAML Validation

- Line number reporting
- Clear syntax error messages
- Type validation errors
- Structure validation

## Test Coverage

### Command Categories (69 total)

1. **Navigation** (10): URL navigation, scrolling variants
2. **Interaction** (15): Click, write, mouse, select operations
3. **Assertions** (12): All comparison and validation types
4. **Window** (7): Resize, maximize, tab/frame switching
5. **Data** (6): Storage and cookie operations
6. **Wait** (2): Element and time-based waits
7. **Dialog** (5): Alert, confirm, prompt handling
8. **File** (2): URL-based file uploads
9. **Misc** (2): Comments and JavaScript execution
10. **Syntax variants** (8): Alternative command syntaxes

### Error Categories

- Missing required parameters
- Invalid parameter types
- Malformed values
- Syntax errors
- Logical conflicts
- Resource errors

### Edge Conditions

- Boundary values
- Empty inputs
- Very large inputs
- Special characters
- Unicode handling
- Complex selectors
- Numeric limits

## Maintenance

When adding new commands or features:

1. Add positive test cases to `test-all-commands-positive.yaml`
2. Add error cases to `test-negative-cases.yaml`
3. Add boundary tests to `test-edge-cases.yaml`
4. Update this README with new test coverage

## Integration with CI/CD

These test files can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v2
    - name: Run positive tests
      run: ./bin/api-cli run-test test-all-commands-positive.yaml --dry-run
    - name: Validate error handling
      run: |
        ./bin/api-cli run-test test-negative-cases.yaml --dry-run || true
        ./bin/api-cli run-test test-yaml-validation.yaml --dry-run || true
```

## Notes

- The `test-yaml-validation.yaml` file intentionally contains invalid YAML to test parser error handling
- Some negative tests are expected to fail - success means they failed with appropriate error messages
- Edge case tests help ensure the CLI handles unusual but valid inputs gracefully
- Use `--dry-run` flag to validate tests without creating actual test infrastructure
