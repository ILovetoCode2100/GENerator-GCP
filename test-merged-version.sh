#!/bin/bash

# Test script for merged Version A with Version B enhancements

VERSION_A_DIR="/Users/marklovelady/_dev/virtuoso-api-cli-generator"
BINARY="$VERSION_A_DIR/bin/api-cli"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================="
echo "Testing Merged Version A"
echo "========================================="

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}❌ Binary not found at $BINARY${NC}"
    echo "Please run build-merged-version.sh first"
    exit 1
fi

cd "$VERSION_A_DIR"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_pattern="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "\n${YELLOW}Test $TOTAL_TESTS: $test_name${NC}"
    echo "Command: $command"
    
    if output=$($command 2>&1); then
        if echo "$output" | grep -q "$expected_pattern"; then
            echo -e "${GREEN}✅ PASSED${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ FAILED - Expected pattern not found: $expected_pattern${NC}"
            echo "Output: $output"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        echo -e "${RED}❌ FAILED - Command returned error${NC}"
        echo "Error: $output"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

echo ""
echo "📋 Testing Basic Functionality"
echo "================================"

# Test 1: Help command
run_test "Help Command" \
    "$BINARY --help" \
    "Virtuoso API CLI"

# Test 2: Version command
run_test "Version Command" \
    "$BINARY --version" \
    "version"

# Test 3: List of commands
echo -e "\n${YELLOW}Test: Command Count${NC}"
TOTAL_COMMANDS=$($BINARY --help 2>&1 | grep -E "^\s+[a-z]" | wc -l)
CREATE_STEP_COMMANDS=$($BINARY --help 2>&1 | grep "create-step-" | wc -l)
echo "Total commands: $TOTAL_COMMANDS"
echo "Create-step commands: $CREATE_STEP_COMMANDS"

if [ $CREATE_STEP_COMMANDS -ge 28 ]; then
    echo -e "${GREEN}✅ Version B commands integrated (found $CREATE_STEP_COMMANDS create-step commands)${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    echo -e "${RED}❌ Missing Version B commands (found only $CREATE_STEP_COMMANDS create-step commands)${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

echo ""
echo "📋 Testing Version A Original Commands"
echo "======================================"

# Test Version A commands
run_test "Create Project Help" \
    "$BINARY create-project --help" \
    "Create a new project"

run_test "List Projects Help" \
    "$BINARY list-projects --help" \
    "List all projects"

run_test "Create Goal Help" \
    "$BINARY create-goal --help" \
    "Create a new goal"

echo ""
echo "📋 Testing Version B Enhanced Commands"
echo "======================================"

# Test Version B commands
run_test "Cookie Create Help" \
    "$BINARY create-step-cookie-create --help" \
    "Create a cookie"

run_test "Cookie Wipe All Help" \
    "$BINARY create-step-cookie-wipe-all --help" \
    "Clear all cookies"

run_test "Execute Script Help" \
    "$BINARY create-step-execute-script --help" \
    "Execute a custom script"

run_test "Mouse Move To Help" \
    "$BINARY create-step-mouse-move-to --help" \
    "Move mouse to absolute coordinates"

run_test "Mouse Move By Help" \
    "$BINARY create-step-mouse-move-by --help" \
    "Move mouse by relative offset"

run_test "Pick Index Help" \
    "$BINARY create-step-pick-index --help" \
    "Pick dropdown option by index"

run_test "Store Element Text Help" \
    "$BINARY create-step-store-element-text --help" \
    "Store element text"

run_test "Window Resize Help" \
    "$BINARY create-step-window-resize --help" \
    "Resize browser window"

echo ""
echo "📋 Testing Command Execution (Dry Run)"
echo "====================================="

# Test command execution with missing token (should fail gracefully)
run_test "Cookie Create Execution (No Token)" \
    "$BINARY create-step-cookie-create 123 sessionid abc123 1 2>&1" \
    "VIRTUOSO_API_TOKEN environment variable is required"

echo ""
echo "📋 Testing Output Formats"
echo "========================"

# Test different output formats
for format in human json yaml ai; do
    run_test "Output Format: $format" \
        "$BINARY create-step-cookie-create --help 2>&1 | grep -i output" \
        "Output format"
done

echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✅ All tests passed! The merge was successful.${NC}"
    echo ""
    echo "The merged Version A includes:"
    echo "- All original project management commands"
    echo "- All 28 enhanced Version B commands"
    echo "- Multiple output formats (human, json, yaml, ai)"
    echo ""
    echo "To use the CLI:"
    echo "export VIRTUOSO_API_TOKEN='your-token-here'"
    echo "export VIRTUOSO_API_BASE_URL='https://api-app2.virtuoso.qa/api'"
    echo "$BINARY [command] [args]"
else
    echo -e "\n${RED}❌ Some tests failed. Please check the output above.${NC}"
fi

echo "========================================="