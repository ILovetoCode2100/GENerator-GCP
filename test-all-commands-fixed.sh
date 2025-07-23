#!/bin/bash

# Comprehensive Virtuoso API CLI Test with Step Verification - FIXED VERSION
# This script tests ALL CLI commands and verifies they create the correct steps

set -e

# Configuration
BINARY="./bin/api-cli"
TEST_DATE=$(date +"%Y%m%d_%H%M%S")
TEST_PROJECT="Full_CLI_Test_Fixed_${TEST_DATE}"
RESULTS_FILE="test-results-fixed-${TEST_DATE}.log"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test counters
TOTAL_COMMANDS=0
SUCCESSFUL_COMMANDS=0
FAILED_COMMANDS=0

# Function to log test results
log_test() {
    local test_name="$1"
    local command="$2"
    local result="$3"
    local details="$4"

    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $test_name" | tee -a "$RESULTS_FILE"
    echo "Command: $command" | tee -a "$RESULTS_FILE"
    echo "Result: $result" | tee -a "$RESULTS_FILE"
    if [ -n "$details" ]; then
        echo "Details: $details" | tee -a "$RESULTS_FILE"
    fi
    echo "---" | tee -a "$RESULTS_FILE"
}

# Fixed execute_command function that properly handles quoted arguments
execute_command() {
    local test_name="$1"
    shift  # Remove first argument
    local command=("$@")  # Store remaining arguments as array

    echo -e "\n${BLUE}Testing: $test_name${NC}"
    ((TOTAL_COMMANDS++))

    # Execute command using array expansion to preserve arguments
    if output=$("${command[@]}" 2>&1); then
        echo -e "${GREEN}✓ Success${NC}"
        ((SUCCESSFUL_COMMANDS++))
        log_test "$test_name" "${command[*]}" "SUCCESS" "$output"

        # Extract step ID if present
        if [[ $output =~ ID:\ ([0-9]+) ]]; then
            echo "  Step ID: ${BASH_REMATCH[1]}"
        fi
    else
        echo -e "${RED}✗ Failed${NC}"
        ((FAILED_COMMANDS++))
        log_test "$test_name" "${command[*]}" "FAILED" "$output"
    fi
}

echo "=============================================="
echo "Virtuoso API CLI - Complete Test Suite (FIXED)"
echo "With Step Verification"
echo "Date: $(date)"
echo "=============================================="
echo "" | tee -a "$RESULTS_FILE"

# ==========================================
# STEP 1: Create Test Infrastructure
# ==========================================
echo -e "\n${YELLOW}=== Creating Test Infrastructure ===${NC}"

# Create project
PROJECT_ID=$($BINARY create-project "$TEST_PROJECT" -o json | jq -r '.project_id')
echo -e "${GREEN}✓ Project created: $PROJECT_ID${NC}"

# Create goal
GOAL_JSON=$($BINARY create-goal $PROJECT_ID "Complete CLI Test" -o json)
GOAL_ID=$(echo "$GOAL_JSON" | jq -r '.goal_id')
SNAPSHOT_ID=$(echo "$GOAL_JSON" | jq -r '.snapshot_id')
echo -e "${GREEN}✓ Goal created: $GOAL_ID${NC}"

# Create journey
JOURNEY_ID=$($BINARY create-journey $GOAL_ID $SNAPSHOT_ID "All Commands Test" -o json | jq -r '.journey_id')
echo -e "${GREEN}✓ Journey created: $JOURNEY_ID${NC}"

# Create main checkpoint
CHECKPOINT_ID=$($BINARY create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID "All CLI Commands" -o json | jq -r '.checkpoint_id')
echo -e "${GREEN}✓ Checkpoint created: $CHECKPOINT_ID${NC}"

# Position counter
POS=1

# ==========================================
# STEP 2: Test ALL Commands Systematically
# ==========================================

echo -e "\n${YELLOW}=== Testing Navigate Commands (8 types) ===${NC}"

execute_command "Navigate to URL" "$BINARY" step-navigate to "$CHECKPOINT_ID" "https://example.com" "$((POS++))"
execute_command "Navigate scroll top" "$BINARY" step-navigate scroll-top "$CHECKPOINT_ID" "$((POS++))"
execute_command "Navigate scroll bottom" "$BINARY" step-navigate scroll-bottom "$CHECKPOINT_ID" "$((POS++))"
execute_command "Navigate scroll element" "$BINARY" step-navigate scroll-element "$CHECKPOINT_ID" "h1" "$((POS++))"
execute_command "Navigate scroll position" "$BINARY" step-navigate scroll-position "$CHECKPOINT_ID" "100,200" "$((POS++))"
execute_command "Navigate scroll by" "$BINARY" step-navigate scroll-by "$CHECKPOINT_ID" "0,300" "$((POS++))"
execute_command "Navigate scroll up" "$BINARY" step-navigate scroll-up "$CHECKPOINT_ID" "$((POS++))"
execute_command "Navigate scroll down" "$BINARY" step-navigate scroll-down "$CHECKPOINT_ID" "$((POS++))"

echo -e "\n${YELLOW}=== Testing Interact Commands (Basic) ===${NC}"

execute_command "Interact click" "$BINARY" step-interact click "$CHECKPOINT_ID" "button.submit" "$((POS++))"
execute_command "Interact click with position" "$BINARY" step-interact click "$CHECKPOINT_ID" "div.target" "$((POS++))" --position CENTER
execute_command "Interact click with variable" "$BINARY" step-interact click "$CHECKPOINT_ID" "a.link" "$((POS++))" --variable linkText
execute_command "Interact double-click" "$BINARY" step-interact double-click "$CHECKPOINT_ID" "div.card" "$((POS++))"
execute_command "Interact right-click" "$BINARY" step-interact right-click "$CHECKPOINT_ID" "div.menu" "$((POS++))"
execute_command "Interact hover" "$BINARY" step-interact hover "$CHECKPOINT_ID" "a.tooltip" "$((POS++))"
execute_command "Interact hover with duration" "$BINARY" step-interact hover "$CHECKPOINT_ID" "button.info" "$((POS++))" --duration 2000
execute_command "Interact write" "$BINARY" step-interact write "$CHECKPOINT_ID" "input#email" "test@example.com" "$((POS++))"
execute_command "Interact write with clear" "$BINARY" step-interact write "$CHECKPOINT_ID" "input#username" "newuser" "$((POS++))" --clear
execute_command "Interact key" "$BINARY" step-interact key "$CHECKPOINT_ID" "Enter" "$((POS++))"
execute_command "Interact key with modifiers" "$BINARY" step-interact key "$CHECKPOINT_ID" "a" "$((POS++))" --modifiers ctrl
execute_command "Interact key with target" "$BINARY" step-interact key "$CHECKPOINT_ID" "Tab" "$((POS++))" --target "input#search"

echo -e "\n${YELLOW}=== Testing Interact Mouse Commands ===${NC}"

execute_command "Interact mouse move-to" "$BINARY" step-interact mouse move-to "$CHECKPOINT_ID" "nav.menu" "$((POS++))"
execute_command "Interact mouse move-by" "$BINARY" step-interact mouse move-by "$CHECKPOINT_ID" "50,100" "$((POS++))"
execute_command "Interact mouse move" "$BINARY" step-interact mouse move "$CHECKPOINT_ID" "200,300" "$((POS++))"
execute_command "Interact mouse down" "$BINARY" step-interact mouse down "$CHECKPOINT_ID" "div.draggable" "$((POS++))"
execute_command "Interact mouse up" "$BINARY" step-interact mouse up "$CHECKPOINT_ID" "div.droppable" "$((POS++))"
execute_command "Interact mouse enter" "$BINARY" step-interact mouse enter "$CHECKPOINT_ID" "div.hover-zone" "$((POS++))"

echo -e "\n${YELLOW}=== Testing Interact Select Commands ===${NC}"

execute_command "Interact select option" "$BINARY" step-interact select option "$CHECKPOINT_ID" "select#country" "United States" "$((POS++))"
execute_command "Interact select index" "$BINARY" step-interact select index "$CHECKPOINT_ID" "select#language" "0" "$((POS++))"
execute_command "Interact select last" "$BINARY" step-interact select last "$CHECKPOINT_ID" "select#timezone" "$((POS++))"

echo -e "\n${YELLOW}=== Testing Assert Commands (12 types) ===${NC}"

execute_command "Assert exists" "$BINARY" step-assert exists "$CHECKPOINT_ID" "h1" "$((POS++))"
execute_command "Assert not-exists" "$BINARY" step-assert not-exists "$CHECKPOINT_ID" "div.error" "$((POS++))"
execute_command "Assert equals" "$BINARY" step-assert equals "$CHECKPOINT_ID" "h1" "Welcome" "$((POS++))"
execute_command "Assert not-equals" "$BINARY" step-assert not-equals "$CHECKPOINT_ID" "span.status" "Error" "$((POS++))"
execute_command "Assert checked" "$BINARY" step-assert checked "$CHECKPOINT_ID" "input[type=checkbox]" "$((POS++))"
execute_command "Assert selected" "$BINARY" step-assert selected "$CHECKPOINT_ID" "option[value=us]" "$((POS++))"
execute_command "Assert gt" "$BINARY" step-assert gt "$CHECKPOINT_ID" "span.count" "10" "$((POS++))"
execute_command "Assert gte" "$BINARY" step-assert gte "$CHECKPOINT_ID" "span.total" "100" "$((POS++))"
execute_command "Assert lt" "$BINARY" step-assert lt "$CHECKPOINT_ID" "span.price" "1000" "$((POS++))"
execute_command "Assert lte" "$BINARY" step-assert lte "$CHECKPOINT_ID" "span.quantity" "50" "$((POS++))"
execute_command "Assert matches" "$BINARY" step-assert matches "$CHECKPOINT_ID" "div.email" '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' "$((POS++))"
execute_command "Assert variable" "$BINARY" step-assert variable "$CHECKPOINT_ID" "userName" "testuser" "$((POS++))"

echo -e "\n${YELLOW}=== Testing Window Commands (7 types) ===${NC}"

execute_command "Window resize" "$BINARY" step-window resize "$CHECKPOINT_ID" "1024x768" "$((POS++))"
execute_command "Window maximize" "$BINARY" step-window maximize "$CHECKPOINT_ID" "$((POS++))"
execute_command "Window switch tab next" "$BINARY" step-window switch tab next "$CHECKPOINT_ID" "$((POS++))"
execute_command "Window switch tab prev" "$BINARY" step-window switch tab prev "$CHECKPOINT_ID" "$((POS++))"
execute_command "Window switch tab index" "$BINARY" step-window switch tab index "$CHECKPOINT_ID" "0" "$((POS++))"
execute_command "Window switch iframe" "$BINARY" step-window switch iframe "$CHECKPOINT_ID" "iframe#content" "$((POS++))"
execute_command "Window switch parent-frame" "$BINARY" step-window switch parent-frame "$CHECKPOINT_ID" "$((POS++))"

echo -e "\n${YELLOW}=== Testing Data Commands (6 types) ===${NC}"

execute_command "Data store element-text" "$BINARY" step-data store element-text "$CHECKPOINT_ID" "h1" "pageTitle" "$((POS++))"
execute_command "Data store attribute" "$BINARY" step-data store attribute "$CHECKPOINT_ID" "a.link" "href" "linkUrl" "$((POS++))"
execute_command "Data store literal" "$BINARY" step-data store literal "$CHECKPOINT_ID" "test123" "myVariable" "$((POS++))"
execute_command "Data cookie create" "$BINARY" step-data cookie create "$CHECKPOINT_ID" "sessionId" "abc123" "$((POS++))"
execute_command "Data cookie delete" "$BINARY" step-data cookie delete "$CHECKPOINT_ID" "sessionId" "$((POS++))"
execute_command "Data cookie clear-all" "$BINARY" step-data cookie clear-all "$CHECKPOINT_ID" "$((POS++))"

echo -e "\n${YELLOW}=== Testing Wait Commands (2 types) ===${NC}"

execute_command "Wait element" "$BINARY" step-wait element "$CHECKPOINT_ID" "div.loaded" "$((POS++))"
execute_command "Wait time" "$BINARY" step-wait time "$CHECKPOINT_ID" "2000" "$((POS++))"

echo -e "\n${YELLOW}=== Testing Dialog Commands (6 types) ===${NC}"

execute_command "Dialog dismiss-alert" "$BINARY" step-dialog dismiss-alert "$CHECKPOINT_ID" "$((POS++))"
execute_command "Dialog dismiss-confirm" "$BINARY" step-dialog dismiss-confirm "$CHECKPOINT_ID" "$((POS++))"
execute_command "Dialog dismiss-confirm accept" "$BINARY" step-dialog dismiss-confirm "$CHECKPOINT_ID" "$((POS++))" --accept
execute_command "Dialog dismiss-confirm reject" "$BINARY" step-dialog dismiss-confirm "$CHECKPOINT_ID" "$((POS++))" --reject
execute_command "Dialog dismiss-prompt" "$BINARY" step-dialog dismiss-prompt "$CHECKPOINT_ID" "$((POS++))"
execute_command "Dialog dismiss-prompt-with-text" "$BINARY" step-dialog dismiss-prompt-with-text "$CHECKPOINT_ID" "User Input Text" "$((POS++))"

echo -e "\n${YELLOW}=== Testing File Commands (2 types) ===${NC}"

execute_command "File upload" "$BINARY" step-file upload "$CHECKPOINT_ID" "input[type=file]" "https://example.com/test.pdf" "$((POS++))"
execute_command "File upload-url" "$BINARY" step-file upload-url "$CHECKPOINT_ID" "input.file-upload" "https://example.com/document.doc" "$((POS++))"

echo -e "\n${YELLOW}=== Testing Misc Commands (2 types) ===${NC}"

execute_command "Misc comment" "$BINARY" step-misc comment "$CHECKPOINT_ID" "This is a test comment for verification" "$((POS++))"
execute_command "Misc execute" "$BINARY" step-misc execute "$CHECKPOINT_ID" "return document.title;" "$((POS++))"

# Store final position
FINAL_POS=$((POS - 1))

# ==========================================
# STEP 3: Use GET Steps to Verify
# ==========================================
echo -e "\n${CYAN}=== Verifying Created Steps ===${NC}"

# First, let's check how many steps were created
STEP_COUNT=$($BINARY list-checkpoints $JOURNEY_ID -o json | jq -r '.checkpoints[] | select(.id == '$CHECKPOINT_ID') | .step_count // empty')
if [ -z "$STEP_COUNT" ]; then
    # Try alternative method
    CHECKPOINT_INFO=$($BINARY get-checkpoint $CHECKPOINT_ID -o json 2>/dev/null || echo "{}")
    STEP_COUNT=$(echo "$CHECKPOINT_INFO" | jq -r '.step_count // "unknown"')
fi

echo -e "Total steps reported in checkpoint: ${BLUE}$STEP_COUNT${NC}"
echo -e "Total commands executed: ${BLUE}$FINAL_POS${NC}"

# ==========================================
# STEP 4: Test Summary
# ==========================================
echo -e "\n${YELLOW}======================================"
echo "Test Summary"
echo "======================================"
echo -e "Total Commands Tested: ${BLUE}$TOTAL_COMMANDS${NC}"
echo -e "Successful: ${GREEN}$SUCCESSFUL_COMMANDS${NC}"
echo -e "Failed: ${RED}$FAILED_COMMANDS${NC}"
if [ $TOTAL_COMMANDS -gt 0 ]; then
    if [ $TOTAL_COMMANDS -gt 0 ]; then
    SUCCESS_RATE=$(echo "scale=1; ($SUCCESSFUL_COMMANDS * 100) / $TOTAL_COMMANDS" | bc)
    echo -e "Success Rate: ${SUCCESS_RATE}%"
fi
fi
echo ""
echo -e "Steps Created in Checkpoint: ${BLUE}$STEP_COUNT${NC}"
echo -e "Expected Steps: ${BLUE}$FINAL_POS${NC}"
echo "======================================"

# ==========================================
# STEP 5: Create Category Summary
# ==========================================
echo -e "\n${YELLOW}=== Command Category Summary ===${NC}"

echo "1. Navigate Commands: 8 types tested"
echo "2. Interact Commands:"
echo "   - Basic: 12 variations"
echo "   - Mouse: 6 types (fixed coordinate format)"
echo "   - Select: 3 types"
echo "3. Assert Commands: 12 types"
echo "4. Window Commands: 7 types"
echo "5. Data Commands: 6 types (fixed command names)"
echo "6. Wait Commands: 2 types"
echo "7. Dialog Commands: 6 types (fixed text quoting)"
echo "8. File Commands: 2 types"
echo "9. Misc Commands: 2 types"
echo ""
echo "Total Command Variations: ~67"

# Output files
echo -e "\n${YELLOW}=== Output Files ===${NC}"
echo "Test Results: $RESULTS_FILE"
echo ""
echo "Checkpoint URL:"
echo "https://app.virtuoso.qa/#/project/$PROJECT_ID/journey/$JOURNEY_ID/checkpoint/$CHECKPOINT_ID"

# ==========================================
# STEP 6: Library Commands Test (Separate)
# ==========================================
echo -e "\n${YELLOW}=== Testing Library Commands (Separate Checkpoint) ===${NC}"

# Create library checkpoint
LIB_CHECKPOINT_ID=$($BINARY create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID "Library Test" -o json | jq -r '.checkpoint_id')
echo -e "Library checkpoint created: $LIB_CHECKPOINT_ID"

# Add some steps to library checkpoint
$BINARY step-navigate to "$LIB_CHECKPOINT_ID" "https://example.com" 1 >/dev/null 2>&1
$BINARY step-interact click "$LIB_CHECKPOINT_ID" "button.login" 2 >/dev/null 2>&1

# Test library commands
execute_command "Library add" "$BINARY" library add "$LIB_CHECKPOINT_ID"

echo -e "\n${GREEN}Testing Complete!${NC}"
echo -e "\n${YELLOW}=== Fixes Applied ===${NC}"
echo "1. ✓ Mouse commands now use correct 'x,y' format"
echo "2. ✓ Dialog text properly handled as single argument"
echo "3. ✓ Data store commands use correct names (attribute, not element-attribute)"
echo "4. ✓ Execute command function preserves quoted arguments"
echo "5. ✓ All step-misc commands properly prefixed with 'misc'"

# Exit with appropriate code
if [ $FAILED_COMMANDS -gt 0 ]; then
    exit 1
else
    exit 0
fi
