#!/bin/bash

# Comprehensive YAML Test Runner
# Run tests: ./run-tests.sh [smoke|full]

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_BINARY="${CLI_BINARY:-$SCRIPT_DIR/../bin/api-cli}"
MODE="${1:-smoke}"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counters
TOTAL=0
PASSED=0
FAILED=0

echo "========================================="
echo "YAML Test Suite Runner"
echo "Mode: $MODE"
echo "========================================="

# Check CLI binary
if [[ ! -f "$CLI_BINARY" ]]; then
    echo -e "${RED}Error: CLI binary not found at $CLI_BINARY${NC}"
    exit 1
fi

# Create test checkpoint
echo "Creating test infrastructure..."
PROJECT_ID=$($CLI_BINARY create-project "YAML Tests $(date +%s)" -o json | jq -r '.project_id')
GOAL_JSON=$($CLI_BINARY create-goal "$PROJECT_ID" "Test Goal" -o json)
GOAL_ID=$(echo "$GOAL_JSON" | jq -r '.goal_id')
SNAPSHOT_ID=$(echo "$GOAL_JSON" | jq -r '.snapshot_id')
JOURNEY_ID=$($CLI_BINARY create-journey "$GOAL_ID" "$SNAPSHOT_ID" "Test Journey" -o json | jq -r '.journey_id')
CHECKPOINT_ID=$($CLI_BINARY create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "YAML Tests" -o json | jq -r '.checkpoint_id')

echo "Created checkpoint: $CHECKPOINT_ID"
export VIRTUOSO_SESSION_ID="$CHECKPOINT_ID"

# Function to run test
run_test() {
    local yaml_file=$1
    local test_name=$(basename "$yaml_file" .yaml)

    ((TOTAL++))
    echo -n "Testing $test_name... "

    if $CLI_BINARY run-test -f "$yaml_file" -c "$CHECKPOINT_ID" >/dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC}"
        ((FAILED++))
    fi
}

# Run tests based on mode
if [[ "$MODE" == "smoke" ]]; then
    echo "Running smoke tests..."
    run_test "$SCRIPT_DIR/commands/step-navigate/positive/all-navigate-commands.yaml"
    run_test "$SCRIPT_DIR/commands/step-interact/positive/all-interact-commands.yaml"
    run_test "$SCRIPT_DIR/commands/step-assert/positive/all-assert-commands.yaml"
else
    echo "Running full test suite..."
    for yaml_file in $(find "$SCRIPT_DIR/commands" -name "*.yaml" -type f | sort); do
        run_test "$yaml_file"
    done
fi

# Summary
echo
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Total:  $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Success Rate: $(awk "BEGIN {printf \"%.1f\", ($PASSED/$TOTAL)*100}")%"

# Exit code
[[ $FAILED -eq 0 ]] && exit 0 || exit 1
