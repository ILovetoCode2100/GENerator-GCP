#!/bin/bash

# Simple test script for checkpoint 1680449
# Compatible with older bash versions

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Configuration
export VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"
CHECKPOINT_ID="1680449"
CLI="./bin/api-cli"
LOG_FILE="test-checkpoint-${CHECKPOINT_ID}-simple.log"

echo -e "${BLUE}=== TESTING CHECKPOINT $CHECKPOINT_ID ===${NC}" | tee "$LOG_FILE"
echo "Started: $(date)" | tee -a "$LOG_FILE"

# Counters
TOTAL=0
PASSED=0
FAILED=0

# Test function
test_command() {
    local category="$1"
    local command="$2"
    local description="$3"
    shift 3
    local args=("$@")
    
    TOTAL=$((TOTAL + 1))
    
    echo -e "\n${BLUE}[$category] Testing: $command${NC}" | tee -a "$LOG_FILE"
    echo "Description: $description" | tee -a "$LOG_FILE"
    
    if $CLI "$command" "${args[@]}" 2>&1 | tee -a "$LOG_FILE"; then
        echo -e "${GREEN}✓ PASSED${NC}" | tee -a "$LOG_FILE"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAILED${NC}" | tee -a "$LOG_FILE"
        FAILED=$((FAILED + 1))
    fi
}

# First, validate config
echo -e "\n${BLUE}Validating configuration...${NC}" | tee -a "$LOG_FILE"
if $CLI validate-config; then
    echo -e "${GREEN}✓ Configuration valid${NC}" | tee -a "$LOG_FILE"
else
    echo -e "${RED}✗ Configuration invalid - check API token${NC}" | tee -a "$LOG_FILE"
    exit 1
fi

# Set checkpoint context
echo -e "\n${BLUE}Setting checkpoint context...${NC}" | tee -a "$LOG_FILE"
if $CLI set-checkpoint "$CHECKPOINT_ID" 2>&1 | tee -a "$LOG_FILE"; then
    echo -e "${GREEN}✓ Checkpoint context set${NC}" | tee -a "$LOG_FILE"
else
    echo -e "${YELLOW}⚠ Failed to set checkpoint, using --checkpoint flag${NC}" | tee -a "$LOG_FILE"
fi

# Test a sample from each category

echo -e "\n${PURPLE}=== NAVIGATION TESTS ===${NC}" | tee -a "$LOG_FILE"
test_command "Navigation" "create-step-navigate" "Basic navigation" "https://example.com" "1" --checkpoint "$CHECKPOINT_ID"
test_command "Navigation" "create-step-wait-time" "Wait 2 seconds" "2000" "2" --checkpoint "$CHECKPOINT_ID"

echo -e "\n${PURPLE}=== MOUSE ACTION TESTS ===${NC}" | tee -a "$LOG_FILE"
test_command "Mouse" "create-step-click" "Click button" "#submit-button" "3" --checkpoint "$CHECKPOINT_ID"
test_command "Mouse" "create-step-hover" "Hover menu" ".dropdown-menu" "4" --checkpoint "$CHECKPOINT_ID"

echo -e "\n${PURPLE}=== INPUT TESTS ===${NC}" | tee -a "$LOG_FILE"
test_command "Input" "create-step-write" "Enter email" "test@example.com" "#email" "5" --checkpoint "$CHECKPOINT_ID"
test_command "Input" "create-step-key" "Press Enter" "Enter" "6" --checkpoint "$CHECKPOINT_ID"

echo -e "\n${PURPLE}=== ASSERTION TESTS ===${NC}" | tee -a "$LOG_FILE"
test_command "Assert" "create-step-assert-exists" "Check success message" ".success-msg" "7" --checkpoint "$CHECKPOINT_ID"
test_command "Assert" "create-step-assert-equals" "Check text" "#result" "Success!" "8" --checkpoint "$CHECKPOINT_ID"

echo -e "\n${PURPLE}=== SCROLL TESTS ===${NC}" | tee -a "$LOG_FILE"
test_command "Scroll" "create-step-scroll-bottom" "Scroll to bottom" "9" --checkpoint "$CHECKPOINT_ID"
test_command "Scroll" "create-step-scroll-element" "Scroll to element" "#footer" "10" --checkpoint "$CHECKPOINT_ID"

echo -e "\n${PURPLE}=== DATA TESTS ===${NC}" | tee -a "$LOG_FILE"
test_command "Data" "create-step-store" "Store username" "#username" "currentUser" "11" --checkpoint "$CHECKPOINT_ID"
test_command "Data" "create-step-execute-js" "Get page title" "return document.title;" "pageTitle" "12" --checkpoint "$CHECKPOINT_ID"

echo -e "\n${PURPLE}=== OUTPUT FORMAT TESTS ===${NC}" | tee -a "$LOG_FILE"
echo "Testing JSON output..." | tee -a "$LOG_FILE"
test_command "Format" "create-step-click" "JSON format test" "#test" "13" --checkpoint "$CHECKPOINT_ID" -o json

echo "Testing YAML output..." | tee -a "$LOG_FILE"
test_command "Format" "create-step-click" "YAML format test" "#test" "14" --checkpoint "$CHECKPOINT_ID" -o yaml

echo "Testing AI output..." | tee -a "$LOG_FILE"
test_command "Format" "create-step-click" "AI format test" "#test" "15" --checkpoint "$CHECKPOINT_ID" -o ai

# Test auto-increment
echo -e "\n${PURPLE}=== AUTO-INCREMENT TEST ===${NC}" | tee -a "$LOG_FILE"
$CLI set-checkpoint "$CHECKPOINT_ID" >/dev/null 2>&1
echo "Testing without position argument..." | tee -a "$LOG_FILE"
test_command "Auto" "create-step-navigate" "Auto position 1" "https://test.com"
test_command "Auto" "create-step-click" "Auto position 2" "#button"

# Summary
echo -e "\n${BLUE}=== TEST SUMMARY ===${NC}" | tee -a "$LOG_FILE"
echo "Total Tests: $TOTAL" | tee -a "$LOG_FILE"
echo -e "${GREEN}Passed: $PASSED${NC}" | tee -a "$LOG_FILE"
echo -e "${RED}Failed: $FAILED${NC}" | tee -a "$LOG_FILE"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ ALL TESTS PASSED!${NC}" | tee -a "$LOG_FILE"
    echo "Checkpoint $CHECKPOINT_ID is working correctly" | tee -a "$LOG_FILE"
else
    echo -e "\n${RED}✗ SOME TESTS FAILED${NC}" | tee -a "$LOG_FILE"
    echo "Success Rate: $(( PASSED * 100 / TOTAL ))%" | tee -a "$LOG_FILE"
fi

echo -e "\nCompleted: $(date)" | tee -a "$LOG_FILE"
echo "Full log: $LOG_FILE"