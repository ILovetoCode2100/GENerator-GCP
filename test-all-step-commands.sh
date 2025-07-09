#!/bin/bash

# Test script for all step creation commands
# This script tests each of the 24 step creation commands

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load environment
source ./scripts/setup-virtuoso.sh

# Counter for tests
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test checkpoint ID - you'll need to replace this with a valid checkpoint ID
CHECKPOINT_ID="${TEST_CHECKPOINT_ID:-1678318}"

echo "================================================"
echo "Testing All Step Creation Commands"
echo "Using Checkpoint ID: $CHECKPOINT_ID"
echo "================================================"

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -n "Testing $test_name... "
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAILED${NC}"
        echo "  Command: $command"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

echo ""
echo "=== Navigation and Control Steps ==="
run_test "create-step-navigate" "./bin/api-cli create-step-navigate $CHECKPOINT_ID 'https://example.com' 1 -o json"
run_test "create-step-wait-time" "./bin/api-cli create-step-wait-time $CHECKPOINT_ID 5 2 -o json"
run_test "create-step-wait-element" "./bin/api-cli create-step-wait-element $CHECKPOINT_ID 'Loading Complete' 3 -o json"
run_test "create-step-window" "./bin/api-cli create-step-window $CHECKPOINT_ID 1920 1080 4 -o json"

echo ""
echo "=== Mouse Action Steps ==="
run_test "create-step-click" "./bin/api-cli create-step-click $CHECKPOINT_ID 'Sign in button' 5 -o json"
run_test "create-step-double-click" "./bin/api-cli create-step-double-click $CHECKPOINT_ID 'Header element' 6 -o json"
run_test "create-step-hover" "./bin/api-cli create-step-hover $CHECKPOINT_ID 'Menu item' 7 -o json"
run_test "create-step-right-click" "./bin/api-cli create-step-right-click $CHECKPOINT_ID 'Context menu trigger' 8 -o json"

echo ""
echo "=== Input and Form Steps ==="
run_test "create-step-write" "./bin/api-cli create-step-write $CHECKPOINT_ID 'test@example.com' 'Email field' 9 -o json"
run_test "create-step-key" "./bin/api-cli create-step-key $CHECKPOINT_ID 'Enter' 10 -o json"
run_test "create-step-pick" "./bin/api-cli create-step-pick $CHECKPOINT_ID 'United States' 'Country dropdown' 11 -o json"
run_test "create-step-upload" "./bin/api-cli create-step-upload $CHECKPOINT_ID 'document.pdf' 'File input' 12 -o json"

echo ""
echo "=== Scroll Steps ==="
run_test "create-step-scroll-top" "./bin/api-cli create-step-scroll-top $CHECKPOINT_ID 13 -o json"
run_test "create-step-scroll-bottom" "./bin/api-cli create-step-scroll-bottom $CHECKPOINT_ID 14 -o json"
run_test "create-step-scroll-element" "./bin/api-cli create-step-scroll-element $CHECKPOINT_ID 'Submit button' 15 -o json"

echo ""
echo "=== Assertion Steps ==="
run_test "create-step-assert-exists" "./bin/api-cli create-step-assert-exists $CHECKPOINT_ID 'Welcome message' 16 -o json"
run_test "create-step-assert-not-exists" "./bin/api-cli create-step-assert-not-exists $CHECKPOINT_ID 'Error message' 17 -o json"
run_test "create-step-assert-equals" "./bin/api-cli create-step-assert-equals $CHECKPOINT_ID 'Total price' '\$99.99' 18 -o json"
run_test "create-step-assert-checked" "./bin/api-cli create-step-assert-checked $CHECKPOINT_ID 'Terms checkbox' 19 -o json"

echo ""
echo "=== Data and Browser Management Steps ==="
run_test "create-step-store" "./bin/api-cli create-step-store $CHECKPOINT_ID 'Order ID' 'orderId' 20 -o json"
run_test "create-step-execute-js" "./bin/api-cli create-step-execute-js $CHECKPOINT_ID 'console.log(\"test\")' 21 -o json"
run_test "create-step-add-cookie" "./bin/api-cli create-step-add-cookie $CHECKPOINT_ID 'session' 'abc123' 22 -o json"
run_test "create-step-dismiss-alert" "./bin/api-cli create-step-dismiss-alert $CHECKPOINT_ID 23 -o json"
run_test "create-step-comment" "./bin/api-cli create-step-comment $CHECKPOINT_ID 'This is a test comment' 24 -o json"

echo ""
echo "================================================"
echo "Test Results Summary"
echo "================================================"
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi