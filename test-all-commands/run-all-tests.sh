#!/bin/bash

# Script to run all test files and verify the run-test command works correctly

CLI_BIN="../bin/api-cli"
TEST_DIR="$(pwd)"

echo "=========================================="
echo "Running Comprehensive Command Tests"
echo "=========================================="

# Check if CLI exists
if [ ! -f "$CLI_BIN" ]; then
    echo "Error: CLI binary not found at $CLI_BIN"
    echo "Please run 'make build' from the project root first"
    exit 1
fi

# Function to run a test
run_test() {
    local test_file=$1
    local test_name=$2

    echo ""
    echo "Running: $test_name"
    echo "File: $test_file"
    echo "------------------------------------------"

    # First do a dry run
    echo "🔍 Dry run..."
    $CLI_BIN run-test "$test_file" --dry-run

    echo ""
    echo "💡 To actually create this test, run:"
    echo "   $CLI_BIN run-test $test_file"
    echo ""
}

# Run each test file
run_test "01-assert-commands.yaml" "Assert Commands Test"
run_test "02-interact-commands.yaml" "Interact Commands Test"
run_test "03-navigate-data-commands.yaml" "Navigate and Data Commands Test"
run_test "04-window-dialog-misc.yaml" "Window, Dialog, and Misc Commands Test"
run_test "05-comprehensive-test.yaml" "Comprehensive Test Suite"

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo ""
echo "The run-test command supports the following simplified syntax:"
echo ""
echo "✅ Supported Commands:"
echo "  - navigate: URL navigation"
echo "  - click: Click elements"
echo "  - hover: Hover over elements"
echo "  - write: Type text into inputs"
echo "  - key: Press keyboard keys"
echo "  - select: Select dropdown options"
echo "  - assert: Check element/text exists"
echo "  - wait: Wait for time or element"
echo "  - scroll: Scroll to element or position"
echo "  - store: Store element text in variables"
echo "  - comment: Add test comments"
echo "  - execute: Run JavaScript code"
echo ""
echo "❌ Not Supported in Simplified Syntax:"
echo "  - Advanced assertions (equals, gt, matches, etc.)"
echo "  - Mouse operations (move-to, move-by, down/up)"
echo "  - Window management (resize, maximize, tabs)"
echo "  - Dialog handling (alerts, confirms, prompts)"
echo "  - Cookie operations"
echo "  - File uploads"
echo "  - Select by index or last"
echo ""
echo "📝 For full command access, use direct CLI commands with checkpoint IDs"
echo ""
echo "To create the comprehensive test, run:"
echo "  $CLI_BIN run-test 05-comprehensive-test.yaml"
echo ""
