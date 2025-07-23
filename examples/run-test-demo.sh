#!/bin/bash

# Virtuoso API CLI - run-test Command Demo
# This script demonstrates various ways to use the new unified test runner

echo "=== Virtuoso API CLI - run-test Command Demo ==="
echo

# Set the path to the CLI binary
CLI_BIN="../bin/api-cli"

# Example 1: Dry run with minimal test
echo "1. Dry run with minimal test file:"
echo "   $CLI_BIN run-test minimal-test.yaml --dry-run"
$CLI_BIN run-test minimal-test.yaml --dry-run
echo

# Example 2: JSON output format
echo "2. JSON output format:"
echo "   $CLI_BIN run-test minimal-test.yaml --dry-run --output json"
$CLI_BIN run-test minimal-test.yaml --dry-run --output json
echo

# Example 3: Complex test with all step types
echo "3. Complex login test:"
echo "   $CLI_BIN run-test simple-login-test.yaml --dry-run"
$CLI_BIN run-test simple-login-test.yaml --dry-run
echo

# Example 4: Test from stdin (YAML)
echo "4. Test from stdin (YAML):"
echo "   cat minimal-test.yaml | $CLI_BIN run-test - --dry-run"
cat minimal-test.yaml | $CLI_BIN run-test - --dry-run
echo

# Example 5: Test from stdin (JSON)
echo "5. Test from stdin (JSON):"
echo '   echo '"'"'{"name":"Quick Test","steps":[...]}'"'"' | $CLI_BIN run-test - --dry-run'
echo '{"name":"Quick Test","steps":[{"type":"navigate","target":"https://example.com"},{"type":"assert","command":"exists","target":"body"}]}' | $CLI_BIN run-test - --dry-run
echo

# Example 6: Auto-naming feature
echo "6. Auto-naming feature (generates unique names):"
echo "   $CLI_BIN run-test minimal-test.yaml --dry-run --auto-name"
$CLI_BIN run-test minimal-test.yaml --dry-run --auto-name
echo

# Example 7: Help text
echo "7. Command help:"
echo "   $CLI_BIN run-test --help"
$CLI_BIN run-test --help

echo
echo "=== Demo Complete ==="
echo
echo "To run tests for real (creating infrastructure and steps):"
echo "  $CLI_BIN run-test test.yaml"
echo
echo "To run tests and execute them immediately:"
echo "  $CLI_BIN run-test test.yaml --execute"
echo
echo "Note: Real execution requires valid API credentials in ~/.api-cli/virtuoso-config.yaml"
