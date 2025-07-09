#\!/bin/bash
# test-enhanced-cli.sh - Comprehensive test suite for enhanced Virtuoso CLI

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Base directory
BASE_DIR="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator"
CONFIG="./config/virtuoso-config.yaml"

# Change to base directory
cd "$BASE_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Virtuoso CLI Enhanced Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_success="${3:-true}"
    
    echo -e "\n${YELLOW}Test: ${test_name}${NC}"
    echo "Command: $command"
    
    if $expected_success; then
        if eval "$command" > /tmp/test_output.txt 2>&1; then
            echo -e "${GREEN}✓ PASSED${NC}"
            ((TESTS_PASSED++))
            # Show first few lines of output
            head -n 5 /tmp/test_output.txt | sed 's/^/  /'
            if [ $(wc -l < /tmp/test_output.txt) -gt 5 ]; then
                echo "  ..."
            fi
        else
            echo -e "${RED}✗ FAILED${NC}"
            ((TESTS_FAILED++))
            cat /tmp/test_output.txt | sed 's/^/  /'
        fi
    else
        # Expecting failure
        if eval "$command" > /tmp/test_output.txt 2>&1; then
            echo -e "${RED}✗ FAILED (expected to fail but passed)${NC}"
            ((TESTS_FAILED++))
        else
            echo -e "${GREEN}✓ PASSED (failed as expected)${NC}"
            ((TESTS_PASSED++))
        fi
    fi
}

# Test 1: Validate Configuration
run_test "Validate Configuration" \
    "./bin/api-cli validate-config --config $CONFIG"

# Test 2: Validate Config - JSON Output
run_test "Validate Config (JSON)" \
    "./bin/api-cli validate-config --config $CONFIG -o json"

# Test 3: List Projects
run_test "List Projects" \
    "./bin/api-cli list-projects --config $CONFIG"

# Test 4: Create Project with Timestamp
TIMESTAMP=$(date +%s)
PROJECT_NAME="Test Suite $TIMESTAMP"
run_test "Create Project" \
    "./bin/api-cli create-project \"$PROJECT_NAME\" --config $CONFIG"

# Extract project ID from output
PROJECT_ID=$(./bin/api-cli create-project "Test Project $TIMESTAMP-2" --config $CONFIG -o json | jq -r .project_id)
echo -e "  Created project ID: ${PROJECT_ID}"

# Test 5: List Goals for Project
run_test "List Goals" \
    "./bin/api-cli list-goals $PROJECT_ID --config $CONFIG"

# Test 6: Create Goal
run_test "Create Goal" \
    "./bin/api-cli create-goal $PROJECT_ID \"Test Goal $TIMESTAMP\" --url \"https://example.com\" --config $CONFIG"

# Get goal details
GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "Goal $TIMESTAMP-2" --url "https://test.com" --config $CONFIG -o json)
GOAL_ID=$(echo $GOAL_RESULT | jq -r .goal_id)
SNAPSHOT_ID=$(echo $GOAL_RESULT | jq -r .snapshot_id)
echo -e "  Created goal ID: ${GOAL_ID}, snapshot: ${SNAPSHOT_ID}"

# Test 7: List Journeys
run_test "List Journeys" \
    "./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID --config $CONFIG"

# Test 8: Create Journey
run_test "Create Journey" \
    "./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID \"Test Journey $TIMESTAMP\" --config $CONFIG"

# Test 9: Test Help Commands
run_test "Help Command" \
    "./bin/api-cli --help"

run_test "Create Structure Help" \
    "./bin/api-cli create-structure --help"

# Test 10: Dry Run Structure Creation
run_test "Structure Creation (Dry Run)" \
    "./bin/api-cli create-structure --file examples/test-small.yaml --config $CONFIG --dry-run"

# Test 11: Invalid Command
run_test "Invalid Command (Should Fail)" \
    "./bin/api-cli invalid-command" \
    false

# Test 12: Missing Required Args
run_test "Missing Args (Should Fail)" \
    "./bin/api-cli create-goal" \
    false

# Test 13: Create small structure with unique name
cat > /tmp/test-structure-$TIMESTAMP.yaml << EOFINNER
project:
  name: "Auto Test Suite $TIMESTAMP"
  description: "Automated test created at $(date)"
goals:
  - name: "Simple Goal"
    url: "https://example.com"
    journeys:
      - name: "Simple Journey"
        checkpoints:
          - name: "Simple Checkpoint"
            steps:
              - type: navigate
                url: "https://example.com"
              - type: wait
                selector: "body"
                timeout: 1000
EOFINNER

run_test "Create Structure from File" \
    "./bin/api-cli create-structure --file /tmp/test-structure-$TIMESTAMP.yaml --config $CONFIG"

# Test 14: Test different output formats
run_test "JSON Output Format" \
    "./bin/api-cli list-projects --config $CONFIG -o json"

run_test "YAML Output Format" \
    "./bin/api-cli validate-config --config $CONFIG -o yaml"

run_test "AI Output Format" \
    "./bin/api-cli validate-config --config $CONFIG -o ai"

# Test 15: Test complete workflow script
run_test "Complete Test Script" \
    "./examples/create-complete-test.sh"

# Summary
echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
else
    echo -e "${GREEN}Failed: $TESTS_FAILED${NC}"
fi
echo -e "${BLUE}Total:  $((TESTS_PASSED + TESTS_FAILED))${NC}"

# Cleanup
rm -f /tmp/test-structure-$TIMESTAMP.yaml /tmp/test_output.txt

# Exit with appropriate code
if [ $TESTS_FAILED -gt 0 ]; then
    exit 1
else
    echo -e "\n${GREEN}All tests passed\! 🎉${NC}"
    exit 0
fi
EOF < /dev/null