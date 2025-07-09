#!/bin/bash
# test-batch-structure.sh - Test the batch structure creation functionality

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

echo "🧪 Testing Virtuoso CLI Batch Structure Commands"
echo "================================================"

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_success="$3"
    
    echo -ne "Testing $test_name... "
    
    if eval "$command" > /tmp/test_output.txt 2>&1; then
        if [ "$expected_success" = "true" ]; then
            echo -e "${GREEN}✓ PASSED${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}✗ FAILED (expected failure)${NC}"
            cat /tmp/test_output.txt
            ((TESTS_FAILED++))
        fi
    else
        if [ "$expected_success" = "false" ]; then
            echo -e "${GREEN}✓ PASSED (correctly failed)${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}✗ FAILED${NC}"
            cat /tmp/test_output.txt
            ((TESTS_FAILED++))
        fi
    fi
}

# Test 1: Update Journey Name
echo -e "\n${YELLOW}1. Testing Journey Management${NC}"
# You'll need to provide a valid journey ID here
JOURNEY_ID=${TEST_JOURNEY_ID:-608048}
run_test "update-journey command" "./bin/api-cli update-journey $JOURNEY_ID --name 'Test Journey $(date +%s)'" "true"

# Test 2: Get Step Details
echo -e "\n${YELLOW}2. Testing Step Details Retrieval${NC}"
# You'll need to provide a valid step ID here
STEP_ID=${TEST_STEP_ID:-19636330}
run_test "get-step command" "./bin/api-cli get-step $STEP_ID -o json" "true"

# Extract canonical ID if successful
if [ -f /tmp/test_output.txt ]; then
    CANONICAL_ID=$(cat /tmp/test_output.txt | jq -r .canonicalId 2>/dev/null || echo "")
    echo "Extracted Canonical ID: $CANONICAL_ID"
fi

# Test 3: Update Navigation Step
echo -e "\n${YELLOW}3. Testing Navigation Update${NC}"
if [ ! -z "$CANONICAL_ID" ]; then
    run_test "update-navigation command" "./bin/api-cli update-navigation $STEP_ID $CANONICAL_ID --url 'https://test-$(date +%s).example.com'" "true"
else
    echo -e "${YELLOW}⚠️  Skipping navigation update test (no canonical ID)${NC}"
fi

# Test 4: List Checkpoints
echo -e "\n${YELLOW}4. Testing Checkpoint Listing${NC}"
run_test "list-checkpoints command" "./bin/api-cli list-checkpoints $JOURNEY_ID" "true"

# Test 5: Dry Run Structure Creation
echo -e "\n${YELLOW}5. Testing Structure Creation (Dry Run)${NC}"
run_test "create-structure dry-run" "./bin/api-cli create-structure --file examples/simple-test-structure.yaml --dry-run" "true"

# Test 6: Validate Structure File
echo -e "\n${YELLOW}6. Testing Structure Validation${NC}"
run_test "invalid structure file" "./bin/api-cli create-structure --file /nonexistent/file.yaml" "false"

# Test 7: Test with existing project ID
echo -e "\n${YELLOW}7. Testing with Existing Project${NC}"
PROJECT_ID=${TEST_PROJECT_ID:-9056}
run_test "create-structure with project-id" "./bin/api-cli create-structure --file examples/simple-test-structure.yaml --project-id $PROJECT_ID --dry-run" "true"

# Summary
echo -e "\n================================================"
echo -e "Test Summary:"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed!${NC}"
    exit 1
fi
