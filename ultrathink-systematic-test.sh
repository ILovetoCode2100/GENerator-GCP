#!/bin/bash

# ULTRATHINK Systematic Test of All 47 Step Commands
# Tests modern vs legacy syntax for each command

set -e

# Configuration
export VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"
CHECKPOINT_ID="1680450"
CLI="./bin/api-cli"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Counters
MODERN_COUNT=0
LEGACY_COUNT=0
POSITION=1

# Results arrays
MODERN_COMMANDS=()
LEGACY_COMMANDS=()

echo -e "${PURPLE}=== ULTRATHINK SYSTEMATIC TEST OF ALL 47 STEP COMMANDS ===${NC}"
echo "Checkpoint: $CHECKPOINT_ID"
echo "Testing modern (--checkpoint flag) vs legacy (checkpoint as arg) syntax"
echo ""

# Test function for modern syntax
test_modern() {
    local cmd="$1"
    local desc="$2"
    shift 2
    local args=("$@")
    
    echo -n "Testing $cmd (modern)... "
    
    if $CLI "$cmd" "${args[@]}" --checkpoint "$CHECKPOINT_ID" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ MODERN${NC}"
        MODERN_COMMANDS+=("$cmd")
        MODERN_COUNT=$((MODERN_COUNT + 1))
        return 0
    else
        return 1
    fi
}

# Test function for legacy syntax
test_legacy() {
    local cmd="$1"
    local desc="$2"
    shift 2
    local args=("$@")
    
    echo -n "Testing $cmd (legacy)... "
    
    if $CLI "$cmd" "$CHECKPOINT_ID" "${args[@]}" >/dev/null 2>&1; then
        echo -e "${YELLOW}✓ LEGACY${NC}"
        LEGACY_COMMANDS+=("$cmd")
        LEGACY_COUNT=$((LEGACY_COUNT + 1))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

# Test each command
test_command() {
    local cmd="$1"
    local desc="$2"
    local type="$3"
    shift 3
    local args=("$@")
    
    echo -e "\n${BLUE}[$type] $cmd - $desc${NC}"
    
    # Try modern first
    if test_modern "$cmd" "$desc" "${args[@]}" "$POSITION"; then
        :
    else
        # Try legacy
        test_legacy "$cmd" "$desc" "${args[@]}" "$POSITION"
    fi
    
    POSITION=$((POSITION + 1))
}

# Set checkpoint context
echo "Setting checkpoint context..."
$CLI set-checkpoint "$CHECKPOINT_ID" >/dev/null 2>&1

# NAVIGATION COMMANDS (4)
echo -e "\n${PURPLE}=== NAVIGATION COMMANDS (4) ===${NC}"
test_command "create-step-navigate" "Navigate to URL" "Navigation" "https://example.com"
test_command "create-step-wait-time" "Wait for time" "Navigation" "2000"
test_command "create-step-wait-element" "Wait for element" "Navigation" "#submit"
test_command "create-step-window" "Window operations" "Navigation" "1024" "768"

# MOUSE COMMANDS (8)
echo -e "\n${PURPLE}=== MOUSE COMMANDS (8) ===${NC}"
test_command "create-step-click" "Click element" "Mouse" "#button"
test_command "create-step-double-click" "Double click" "Mouse" ".item"
test_command "create-step-right-click" "Right click" "Mouse" "#menu"
test_command "create-step-hover" "Hover element" "Mouse" ".dropdown"
test_command "create-step-mouse-down" "Mouse down" "Mouse" ".drag"
test_command "create-step-mouse-up" "Mouse up" "Mouse" ".drop"
test_command "create-step-mouse-move" "Mouse move" "Mouse" "100" "200"
test_command "create-step-mouse-enter" "Mouse enter" "Mouse" ".zone"

# INPUT COMMANDS (6)
echo -e "\n${PURPLE}=== INPUT COMMANDS (6) ===${NC}"
test_command "create-step-write" "Write text" "Input" "test@example.com" "#email"
test_command "create-step-key" "Press key" "Input" "Enter"
test_command "create-step-pick" "Pick by index" "Input" "#select" "2"
test_command "create-step-pick-value" "Pick by value" "Input" "#select" "opt1"
test_command "create-step-pick-text" "Pick by text" "Input" "#select" "Option 1"
test_command "create-step-upload" "Upload file" "Input" "#file" "/tmp/test.pdf"

# SCROLL COMMANDS (4)
echo -e "\n${PURPLE}=== SCROLL COMMANDS (4) ===${NC}"
test_command "create-step-scroll-top" "Scroll to top" "Scroll"
test_command "create-step-scroll-bottom" "Scroll to bottom" "Scroll"
test_command "create-step-scroll-element" "Scroll to element" "Scroll" "#target"
test_command "create-step-scroll-position" "Scroll to position" "Scroll" "500"

# ASSERTION COMMANDS (11)
echo -e "\n${PURPLE}=== ASSERTION COMMANDS (11) ===${NC}"
test_command "create-step-assert-exists" "Assert exists" "Assert" ".success"
test_command "create-step-assert-not-exists" "Assert not exists" "Assert" ".error"
test_command "create-step-assert-equals" "Assert equals" "Assert" "#result" "Success"
test_command "create-step-assert-checked" "Assert checked" "Assert" "#checkbox"
test_command "create-step-assert-selected" "Assert selected" "Assert" "#option"
test_command "create-step-assert-variable" "Assert variable" "Assert" "myVar" "value"
test_command "create-step-assert-greater-than" "Assert GT" "Assert" "#count" "5"
test_command "create-step-assert-greater-than-or-equal" "Assert GTE" "Assert" "#score" "75"
test_command "create-step-assert-less-than-or-equal" "Assert LTE" "Assert" "#price" "100"
test_command "create-step-assert-matches" "Assert matches" "Assert" "#phone" "\\d{3}-\\d{3}-\\d{4}"
test_command "create-step-assert-not-equals" "Assert not equals" "Assert" "#status" "error"

# DATA COMMANDS (3)
echo -e "\n${PURPLE}=== DATA COMMANDS (3) ===${NC}"
test_command "create-step-store" "Store element" "Data" "#user" "userName"
test_command "create-step-store-value" "Store value" "Data" "#input" "inputVal"
test_command "create-step-execute-js" "Execute JS" "Data" "return document.title;" "pageTitle"

# ENVIRONMENT COMMANDS (3)
echo -e "\n${PURPLE}=== ENVIRONMENT COMMANDS (3) ===${NC}"
test_command "create-step-add-cookie" "Add cookie" "Environment" "session" "abc123" "example.com" "/"
test_command "create-step-delete-cookie" "Delete cookie" "Environment" "session"
test_command "create-step-clear-cookies" "Clear cookies" "Environment"

# DIALOG COMMANDS (3)
echo -e "\n${PURPLE}=== DIALOG COMMANDS (3) ===${NC}"
test_command "create-step-dismiss-alert" "Dismiss alert" "Dialog"
test_command "create-step-dismiss-confirm" "Dismiss confirm" "Dialog" "true"
test_command "create-step-dismiss-prompt" "Dismiss prompt" "Dialog" "Input text"

# FRAME/TAB COMMANDS (4)
echo -e "\n${PURPLE}=== FRAME/TAB COMMANDS (4) ===${NC}"
test_command "create-step-switch-iframe" "Switch iframe" "Frame" "#iframe"
test_command "create-step-switch-next-tab" "Next tab" "Frame"
test_command "create-step-switch-prev-tab" "Previous tab" "Frame"
test_command "create-step-switch-parent-frame" "Parent frame" "Frame"

# UTILITY COMMAND (1)
echo -e "\n${PURPLE}=== UTILITY COMMAND (1) ===${NC}"
test_command "create-step-comment" "Add comment" "Utility" "Test comment"

# SUMMARY
echo -e "\n${PURPLE}=== SUMMARY ===${NC}"
echo "Total Commands: 47"
echo -e "${GREEN}Modern Syntax Commands: $MODERN_COUNT${NC}"
echo -e "${YELLOW}Legacy Syntax Commands: $LEGACY_COUNT${NC}"
echo ""

echo -e "${GREEN}Modern Commands (support --checkpoint flag):${NC}"
for cmd in "${MODERN_COMMANDS[@]}"; do
    echo "  ✓ $cmd"
done

echo -e "\n${YELLOW}Legacy Commands (checkpoint as first argument):${NC}"
for cmd in "${LEGACY_COMMANDS[@]}"; do
    echo "  ⚠ $cmd"
done

# Save results
{
    echo "ULTRATHINK TEST RESULTS - $(date)"
    echo "================================"
    echo "Checkpoint: $CHECKPOINT_ID"
    echo "Total Commands: 47"
    echo "Modern Syntax: $MODERN_COUNT"
    echo "Legacy Syntax: $LEGACY_COUNT"
    echo ""
    echo "MODERN COMMANDS:"
    printf '%s\n' "${MODERN_COMMANDS[@]}"
    echo ""
    echo "LEGACY COMMANDS:"
    printf '%s\n' "${LEGACY_COMMANDS[@]}"
} > ultrathink-test-results.txt

echo -e "\n✅ Results saved to: ultrathink-test-results.txt"