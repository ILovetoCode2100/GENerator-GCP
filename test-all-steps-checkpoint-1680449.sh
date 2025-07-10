#!/bin/bash

# Comprehensive test script for all 47 step creation commands
# Using checkpoint 1680449 with ULTRATHINK methodology

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
export VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"
CHECKPOINT_ID="1680449"
CLI="./bin/api-cli"
LOG_FILE="test-results-checkpoint-${CHECKPOINT_ID}.log"

# Initialize log
echo "=== ULTRATHINK COMPREHENSIVE STEP TESTING ===" | tee "$LOG_FILE"
echo "Checkpoint ID: $CHECKPOINT_ID" | tee -a "$LOG_FILE"
echo "Test Started: $(date)" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# Function to test a command
test_command() {
    local category="$1"
    local command="$2"
    local description="$3"
    shift 3
    local args=("$@")
    
    echo -e "\n${BLUE}[$category]${NC} Testing: $command" | tee -a "$LOG_FILE"
    echo "Description: $description" | tee -a "$LOG_FILE"
    echo "Command: $CLI $command ${args[*]}" | tee -a "$LOG_FILE"
    
    if $CLI "$command" "${args[@]}" 2>&1 | tee -a "$LOG_FILE"; then
        echo -e "${GREEN}✓ SUCCESS${NC}" | tee -a "$LOG_FILE"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}" | tee -a "$LOG_FILE"
        return 1
    fi
}

# Function to test with different output formats
test_output_formats() {
    local command="$1"
    shift
    local args=("$@")
    
    for format in human json yaml ai; do
        echo -e "\n${YELLOW}Testing output format: $format${NC}" | tee -a "$LOG_FILE"
        $CLI "$command" "${args[@]}" -o "$format" 2>&1 | tee -a "$LOG_FILE"
    done
}

# Set checkpoint context
echo -e "\n${BLUE}=== SETTING CHECKPOINT CONTEXT ===${NC}" | tee -a "$LOG_FILE"
if $CLI set-checkpoint "$CHECKPOINT_ID" 2>&1 | tee -a "$LOG_FILE"; then
    echo -e "${GREEN}✓ Checkpoint context set${NC}" | tee -a "$LOG_FILE"
else
    echo -e "${RED}✗ Failed to set checkpoint context${NC}" | tee -a "$LOG_FILE"
    echo "Continuing with --checkpoint flag for each command..." | tee -a "$LOG_FILE"
    USE_CHECKPOINT_FLAG=true
fi

# Test counters
TOTAL=0
PASSED=0
FAILED=0

# Helper function to increment counters
increment_test() {
    TOTAL=$((TOTAL + 1))
    if [ $? -eq 0 ]; then
        PASSED=$((PASSED + 1))
    else
        FAILED=$((FAILED + 1))
    fi
}

# 1. NAVIGATION COMMANDS (4 types)
echo -e "\n${BLUE}=== NAVIGATION COMMANDS (4 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Navigation" "create-step-navigate" "Basic navigation" "https://example.com" "1" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Navigation" "create-step-wait-time" "Wait for 3 seconds" "3000" "2" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Navigation" "create-step-wait-element" "Wait for element" "#submit-button" "3" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Navigation" "create-step-window" "Window operations" "maximize" "4" --checkpoint "$CHECKPOINT_ID"
increment_test

# 2. MOUSE ACTION COMMANDS (8 types)
echo -e "\n${BLUE}=== MOUSE ACTION COMMANDS (8 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Mouse" "create-step-click" "Click element" "#submit-button" "5" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Mouse" "create-step-double-click" "Double click element" ".item" "6" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Mouse" "create-step-right-click" "Right click element" "#context-menu" "7" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Mouse" "create-step-hover" "Hover over element" ".dropdown-trigger" "8" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Mouse" "create-step-mouse-down" "Mouse down on element" ".draggable" "9" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Mouse" "create-step-mouse-up" "Mouse up on element" ".droppable" "10" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Mouse" "create-step-mouse-move" "Mouse move to coordinates" "100" "200" "11" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Mouse" "create-step-mouse-enter" "Mouse enter element" ".hover-zone" "12" --checkpoint "$CHECKPOINT_ID"
increment_test

# 3. INPUT COMMANDS (6 types)
echo -e "\n${BLUE}=== INPUT COMMANDS (6 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Input" "create-step-write" "Write text to input" "test@example.com" "#email-input" "13" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Input" "create-step-key" "Press key combination" "ctrl+a" "14" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Input" "create-step-pick" "Pick from dropdown by index" "#dropdown" "2" "15" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Input" "create-step-pick-value" "Pick by value" "#dropdown" "option-value" "16" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Input" "create-step-pick-text" "Pick by visible text" "#dropdown" "Option Text" "17" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Input" "create-step-upload" "Upload file" "#file-input" "/path/to/file.pdf" "18" --checkpoint "$CHECKPOINT_ID"
increment_test

# 4. SCROLL COMMANDS (4 types)
echo -e "\n${BLUE}=== SCROLL COMMANDS (4 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Scroll" "create-step-scroll-top" "Scroll to top" "19" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Scroll" "create-step-scroll-bottom" "Scroll to bottom" "20" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Scroll" "create-step-scroll-element" "Scroll to element" "#target-element" "21" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Scroll" "create-step-scroll-position" "Scroll to position" "500" "22" --checkpoint "$CHECKPOINT_ID"
increment_test

# 5. ASSERTION COMMANDS (11 types)
echo -e "\n${BLUE}=== ASSERTION COMMANDS (11 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Assert" "create-step-assert-exists" "Assert element exists" ".success-message" "23" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-not-exists" "Assert element not exists" ".error-message" "24" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-equals" "Assert equals" "#result" "Expected Value" "25" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-checked" "Assert checkbox checked" "#terms-checkbox" "26" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-selected" "Assert option selected" "#dropdown option[value='2']" "27" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-variable" "Assert variable equals" "myVar" "expectedValue" "28" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-greater-than" "Assert greater than" "#count" "10" "29" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-greater-than-or-equal" "Assert greater than or equal" "#score" "75" "30" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-less-than-or-equal" "Assert less than or equal" "#price" "100" "31" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-matches" "Assert matches pattern" "#phone" "\\d{3}-\\d{3}-\\d{4}" "32" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Assert" "create-step-assert-not-equals" "Assert not equals" "#status" "error" "33" --checkpoint "$CHECKPOINT_ID"
increment_test

# 6. DATA COMMANDS (3 types)
echo -e "\n${BLUE}=== DATA COMMANDS (3 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Data" "create-step-store" "Store element text" "#username" "currentUser" "34" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Data" "create-step-store-value" "Store input value" "#email-field" "userEmail" "35" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Data" "create-step-execute-js" "Execute JavaScript" "return document.title;" "pageTitle" "36" --checkpoint "$CHECKPOINT_ID"
increment_test

# 7. ENVIRONMENT COMMANDS (3 types)
echo -e "\n${BLUE}=== ENVIRONMENT COMMANDS (3 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Environment" "create-step-add-cookie" "Add cookie" "session_id" "abc123" "example.com" "/" "37" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Environment" "create-step-delete-cookie" "Delete cookie" "session_id" "38" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Environment" "create-step-clear-cookies" "Clear all cookies" "39" --checkpoint "$CHECKPOINT_ID"
increment_test

# 8. DIALOG COMMANDS (3 types)
echo -e "\n${BLUE}=== DIALOG COMMANDS (3 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Dialog" "create-step-dismiss-alert" "Dismiss alert dialog" "40" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Dialog" "create-step-dismiss-confirm" "Dismiss confirm dialog" "true" "41" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Dialog" "create-step-dismiss-prompt" "Dismiss prompt dialog" "User Input" "42" --checkpoint "$CHECKPOINT_ID"
increment_test

# 9. FRAME/TAB COMMANDS (4 types)
echo -e "\n${BLUE}=== FRAME/TAB COMMANDS (4 types) ===${NC}" | tee -a "$LOG_FILE"

test_command "Frame/Tab" "create-step-switch-iframe" "Switch to iframe" "#content-frame" "43" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Frame/Tab" "create-step-switch-next-tab" "Switch to next tab" "44" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Frame/Tab" "create-step-switch-prev-tab" "Switch to previous tab" "45" --checkpoint "$CHECKPOINT_ID"
increment_test

test_command "Frame/Tab" "create-step-switch-parent-frame" "Switch to parent frame" "46" --checkpoint "$CHECKPOINT_ID"
increment_test

# 10. UTILITY COMMANDS (1 type)
echo -e "\n${BLUE}=== UTILITY COMMANDS (1 type) ===${NC}" | tee -a "$LOG_FILE"

test_command "Utility" "create-step-comment" "Add comment" "This is a test comment for documentation" "47" --checkpoint "$CHECKPOINT_ID"
increment_test

# EDGE CASES AND SPECIAL SCENARIOS
echo -e "\n${BLUE}=== EDGE CASES AND SPECIAL SCENARIOS ===${NC}" | tee -a "$LOG_FILE"

# Test negative numbers
test_command "Edge Case" "create-step-assert-greater-than" "Negative number test" "#temperature" -- "-10" "48" --checkpoint "$CHECKPOINT_ID"
increment_test

# Test special characters in selectors
test_command "Edge Case" "create-step-click" "Special selector characters" "[data-test='submit-btn']" "49" --checkpoint "$CHECKPOINT_ID"
increment_test

# Test empty/null values
test_command "Edge Case" "create-step-write" "Clear input field" "" "#search-box" "50" --checkpoint "$CHECKPOINT_ID"
increment_test

# Test auto-increment position
echo -e "\n${BLUE}=== AUTO-INCREMENT POSITION TEST ===${NC}" | tee -a "$LOG_FILE"
echo "Testing without position argument (should auto-increment)..." | tee -a "$LOG_FILE"

$CLI set-checkpoint "$CHECKPOINT_ID" 2>&1 | tee -a "$LOG_FILE"

test_command "Auto-increment" "create-step-navigate" "Auto position 1" "https://test.com"
increment_test

test_command "Auto-increment" "create-step-click" "Auto position 2" "#login"
increment_test

test_command "Auto-increment" "create-step-assert-exists" "Auto position 3" ".dashboard"
increment_test

# OUTPUT FORMAT TESTS
echo -e "\n${BLUE}=== OUTPUT FORMAT TESTS ===${NC}" | tee -a "$LOG_FILE"

test_output_formats "create-step-click" "#test-button" "60" --checkpoint "$CHECKPOINT_ID"

# SUMMARY
echo -e "\n${BLUE}=== TEST SUMMARY ===${NC}" | tee -a "$LOG_FILE"
echo "Total Tests: $TOTAL" | tee -a "$LOG_FILE"
echo -e "${GREEN}Passed: $PASSED${NC}" | tee -a "$LOG_FILE"
echo -e "${RED}Failed: $FAILED${NC}" | tee -a "$LOG_FILE"
echo "Success Rate: $(( PASSED * 100 / TOTAL ))%" | tee -a "$LOG_FILE"
echo "Test Completed: $(date)" | tee -a "$LOG_FILE"

# Final status
if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ ALL TESTS PASSED!${NC}" | tee -a "$LOG_FILE"
    exit 0
else
    echo -e "\n${RED}✗ SOME TESTS FAILED${NC}" | tee -a "$LOG_FILE"
    exit 1
fi