#!/bin/bash

# ULTRATHINK Final Validation Test Suite
# Tests all 47 commands with modern session context pattern

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
export VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"
CHECKPOINT_ID="1680450"
LOG_FILE="ultrathink-final-validation.log"

echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}" | tee "$LOG_FILE"
echo -e "${PURPLE}║        ULTRATHINK FINAL VALIDATION TEST SUITE               ║${NC}" | tee -a "$LOG_FILE"
echo -e "${PURPLE}║   Testing all 47 step commands with modern pattern          ║${NC}" | tee -a "$LOG_FILE"
echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test function
test_command() {
    local category="$1"
    local command="$2"
    local description="$3"
    shift 3
    local args=("$@")
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "\n${BLUE}[$category] Testing: $command${NC}" | tee -a "$LOG_FILE"
    echo "Description: $description" | tee -a "$LOG_FILE"
    echo "Command: ./bin/api-cli $command ${args[*]}" | tee -a "$LOG_FILE"
    
    if ./bin/api-cli "$command" "${args[@]}" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}" | tee -a "$LOG_FILE"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAILED${NC}" | tee -a "$LOG_FILE"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

echo -e "${CYAN}Building CLI...${NC}" | tee -a "$LOG_FILE"
make build >/dev/null 2>&1
echo -e "${GREEN}✓ Build complete${NC}" | tee -a "$LOG_FILE"

echo -e "\n${CYAN}Setting checkpoint context...${NC}" | tee -a "$LOG_FILE"
./bin/api-cli set-checkpoint "$CHECKPOINT_ID" >/dev/null 2>&1
echo -e "${GREEN}✓ Checkpoint context set to $CHECKPOINT_ID${NC}" | tee -a "$LOG_FILE"

# Navigation Commands (4) - ALL MODERN
echo -e "\n${PURPLE}=== NAVIGATION COMMANDS ===  ${NC}" | tee -a "$LOG_FILE"
test_command "Navigation" "create-step-navigate" "Navigate to URL" "https://test.com"
test_command "Navigation" "create-step-wait-time" "Wait 2 seconds" "2"
test_command "Navigation" "create-step-wait-element" "Wait for element" "#button"
test_command "Navigation" "create-step-window" "Set window size" "1920" "1080"

# Mouse Commands (8) - ALL MODERN
echo -e "\n${PURPLE}=== MOUSE COMMANDS ===${NC}" | tee -a "$LOG_FILE"
test_command "Mouse" "create-step-click" "Click element" "#submit"
test_command "Mouse" "create-step-hover" "Hover menu" ".dropdown"
test_command "Mouse" "create-step-double-click" "Double click item" ".item"
test_command "Mouse" "create-step-right-click" "Right click" "#menu"
test_command "Mouse" "create-step-mouse-down" "Mouse down" ".drag"
test_command "Mouse" "create-step-mouse-up" "Mouse up" ".drop"
test_command "Mouse" "create-step-mouse-move" "Mouse move" "100" "200"
test_command "Mouse" "create-step-mouse-enter" "Mouse enter" ".zone"

# Input Commands (6) - ALL MODERN
echo -e "\n${PURPLE}=== INPUT COMMANDS ===${NC}" | tee -a "$LOG_FILE"
test_command "Input" "create-step-write" "Write text" "test@example.com" "#email"
test_command "Input" "create-step-key" "Press key" "Enter"
test_command "Input" "create-step-pick" "Pick by index" "#select" "2"
test_command "Input" "create-step-pick-value" "Pick by value" "#select" "value1"
test_command "Input" "create-step-pick-text" "Pick by text" "#select" "Option 1"
test_command "Input" "create-step-upload" "Upload file" "#file" "/tmp/test.pdf"

# Scroll Commands (4) - ALL MODERN
echo -e "\n${PURPLE}=== SCROLL COMMANDS ===${NC}" | tee -a "$LOG_FILE"
test_command "Scroll" "create-step-scroll-top" "Scroll to top"
test_command "Scroll" "create-step-scroll-bottom" "Scroll to bottom"
test_command "Scroll" "create-step-scroll-element" "Scroll to element" "#footer"
test_command "Scroll" "create-step-scroll-position" "Scroll to position" "100" "200"

# Assertion Commands (11) - ALREADY MODERN
echo -e "\n${PURPLE}=== ASSERTION COMMANDS ===${NC}" | tee -a "$LOG_FILE"
test_command "Assert" "create-step-assert-exists" "Assert exists" ".success"
test_command "Assert" "create-step-assert-not-exists" "Assert not exists" ".error"
test_command "Assert" "create-step-assert-equals" "Assert equals" "#result" "Success"
test_command "Assert" "create-step-assert-checked" "Assert checked" "#checkbox"
test_command "Assert" "create-step-assert-selected" "Assert selected" "#option"
test_command "Assert" "create-step-assert-variable" "Assert variable" "myVar" "value"
test_command "Assert" "create-step-assert-greater-than" "Assert GT" "#count" "5"
test_command "Assert" "create-step-assert-greater-than-or-equal" "Assert GTE" "#score" "75"
test_command "Assert" "create-step-assert-less-than-or-equal" "Assert LTE" "#price" "100"
test_command "Assert" "create-step-assert-matches" "Assert matches" "#phone" "\\d{3}-\\d{3}-\\d{4}"
test_command "Assert" "create-step-assert-not-equals" "Assert not equals" "#status" "error"

# Data Commands (3) - ALL MODERN
echo -e "\n${PURPLE}=== DATA COMMANDS ===${NC}" | tee -a "$LOG_FILE"
test_command "Data" "create-step-store" "Store element" "#user" "userName"
test_command "Data" "create-step-store-value" "Store value" "test@example.com" "email"
test_command "Data" "create-step-execute-js" "Execute JS" "return document.title;" "pageTitle"

# Environment Commands (3) - ALL MODERN
echo -e "\n${PURPLE}=== ENVIRONMENT COMMANDS ===${NC}" | tee -a "$LOG_FILE"
test_command "Environment" "create-step-add-cookie" "Add cookie" "session" "abc123" "example.com" "/"
test_command "Environment" "create-step-delete-cookie" "Delete cookie" "session"
test_command "Environment" "create-step-clear-cookies" "Clear cookies"

# Dialog Commands (3) - ALL MODERN
echo -e "\n${PURPLE}=== DIALOG COMMANDS ===${NC}" | tee -a "$LOG_FILE"
test_command "Dialog" "create-step-dismiss-alert" "Dismiss alert"
test_command "Dialog" "create-step-dismiss-confirm" "Dismiss confirm" "true"
test_command "Dialog" "create-step-dismiss-prompt" "Dismiss prompt" "User input"

# Frame/Tab Commands (4) - ALL MODERN
echo -e "\n${PURPLE}=== FRAME/TAB COMMANDS ===${NC}" | tee -a "$LOG_FILE"
test_command "Frame" "create-step-switch-iframe" "Switch iframe" "#content"
test_command "Frame" "create-step-switch-next-tab" "Next tab"
test_command "Frame" "create-step-switch-prev-tab" "Previous tab"
test_command "Frame" "create-step-switch-parent-frame" "Parent frame"

# Utility Command (1) - ALL MODERN
echo -e "\n${PURPLE}=== UTILITY COMMAND ===${NC}" | tee -a "$LOG_FILE"
test_command "Utility" "create-step-comment" "Add comment" "Test comment"

# Summary
echo -e "\n${CYAN}═══════════════════════════════════════════════════════════════${NC}" | tee -a "$LOG_FILE"
echo -e "${CYAN}                    FINAL VALIDATION SUMMARY                   ${NC}" | tee -a "$LOG_FILE"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"
echo "Total Commands Tested: $TOTAL_TESTS / 47" | tee -a "$LOG_FILE"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}" | tee -a "$LOG_FILE"
echo -e "${RED}Failed: $FAILED_TESTS${NC}" | tee -a "$LOG_FILE"

if [ $TOTAL_TESTS -eq 47 ] && [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✅ ALL 47 COMMANDS FULLY MODERNIZED AND WORKING!${NC}" | tee -a "$LOG_FILE"
    echo -e "${GREEN}🎉 ULTRATHINK MISSION ACCOMPLISHED!${NC}" | tee -a "$LOG_FILE"
    exit 0
else
    echo -e "\n${YELLOW}⚠ Some commands need attention${NC}" | tee -a "$LOG_FILE"
    echo "Success Rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%" | tee -a "$LOG_FILE"
    exit 1
fi