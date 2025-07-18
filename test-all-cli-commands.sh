#!/bin/bash

# Comprehensive Test Script for Virtuoso API CLI
# Tests ALL command variations, edge cases, and flag combinations
# Version: 1.0
# Date: $(date +%Y-%m-%d)

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
PROJECT_NAME="test_project_${TIMESTAMP}"
GOAL_NAME="test_goal_${TIMESTAMP}"
JOURNEY_NAME="test_journey_${TIMESTAMP}"
CHECKPOINT_NAME="test_checkpoint_${TIMESTAMP}"
TEST_URL="https://example.com"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILED_COMMANDS=()

# CLI path
CLI_PATH="./bin/api-cli"

# Log file
LOG_FILE="test_results_${TIMESTAMP}.log"

# Cleanup flag
CLEANUP=true

# Function to print section headers
print_section() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Function to print subsection headers
print_subsection() {
    echo -e "\n${YELLOW}--- $1 ---${NC}"
}

# Function to test a command
test_command() {
    local description="$1"
    local command="$2"
    local expected_success="${3:-true}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo -n "Testing: $description... "

    # Log the command
    echo "Command: $CLI_PATH $command" >> "$LOG_FILE"

    # Execute command and capture output
    if output=$(eval "$CLI_PATH $command" 2>&1); then
        if [ "$expected_success" = "true" ]; then
            echo -e "${GREEN}PASSED${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            echo "Output: $output" >> "$LOG_FILE"
        else
            echo -e "${RED}FAILED${NC} (expected to fail but succeeded)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_COMMANDS+=("$description: $command")
        fi
    else
        if [ "$expected_success" = "false" ]; then
            echo -e "${GREEN}PASSED${NC} (expected failure)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}FAILED${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_COMMANDS+=("$description: $command")
            echo "Error: $output" >> "$LOG_FILE"
        fi
    fi

    echo "---" >> "$LOG_FILE"
}

# Function to test all output formats for a command
test_all_formats() {
    local base_description="$1"
    local base_command="$2"

    test_command "$base_description (human)" "$base_command --output human"
    test_command "$base_description (json)" "$base_command --output json"
    test_command "$base_description (yaml)" "$base_command --output yaml"
    test_command "$base_description (ai)" "$base_command --output ai"
}

# Start testing
echo -e "${BLUE}Starting Comprehensive Virtuoso API CLI Test${NC}"
echo "Timestamp: ${TIMESTAMP}"
echo "Log file: ${LOG_FILE}"

print_section "SETUP: Creating Test Structure"

# Create project
test_command "Create project" "create-project \"$PROJECT_NAME\" --dry-run"

# Create goal
test_command "Create goal" "create-goal \"$GOAL_NAME\" --project \"$PROJECT_NAME\" --dry-run"

# Create journey
test_command "Create journey" "create-journey \"$JOURNEY_NAME\" --goal \"$GOAL_NAME\" --dry-run"

# Create checkpoint
test_command "Create checkpoint" "create-checkpoint \"$CHECKPOINT_NAME\" --journey \"$JOURNEY_NAME\" --dry-run"

# Set session context
export VIRTUOSO_SESSION_ID="$CHECKPOINT_NAME"

print_section "TESTING ASSERT COMMANDS"

print_subsection "Basic Assertions"
test_all_formats "assert exists" "assert exists \"Login button\" --dry-run"
test_all_formats "assert not-exists" "assert not-exists \"Error message\" --dry-run"
test_all_formats "assert contains" "assert contains \"#content\" \"Welcome\" --dry-run"
test_all_formats "assert not-contains" "assert not-contains \"#content\" \"Error\" --dry-run"
test_all_formats "assert equals" "assert equals \"#count\" \"10\" --dry-run"
test_all_formats "assert not-equals" "assert not-equals \"#status\" \"error\" --dry-run"

print_subsection "Numeric Assertions"
test_all_formats "assert gt" "assert gt \"#price\" \"100\" --dry-run"
test_all_formats "assert gte" "assert gte \"#quantity\" \"5\" --dry-run"
test_all_formats "assert lt" "assert lt \"#remaining\" \"50\" --dry-run"
test_all_formats "assert lte" "assert lte \"#total\" \"1000\" --dry-run"

print_subsection "Boolean Assertions"
test_all_formats "assert true" "assert true \"#checkbox\" --dry-run"
test_all_formats "assert false" "assert false \"#disabled\" --dry-run"

print_subsection "Advanced Assertions with Options"
test_command "assert with element type" "assert exists \"Submit\" --element-type BUTTON --dry-run"
test_command "assert with custom selector" "assert exists \"Submit\" --use-custom-selector --dry-run"
test_command "assert with explicit checkpoint" "assert exists \"Login\" --checkpoint \"$CHECKPOINT_NAME\" --dry-run"
test_command "assert with position" "assert exists \"Login\" --position 5 --dry-run"

print_section "TESTING INTERACT COMMANDS"

print_subsection "Click Variations"
test_all_formats "basic click" "interact click \"Submit\" --dry-run"
test_command "click with position enum" "interact click \"Submit\" --position CENTER --dry-run"
test_command "click top-left" "interact click \"Submit\" --position TOP_LEFT --dry-run"
test_command "click top-center" "interact click \"Submit\" --position TOP_CENTER --dry-run"
test_command "click top-right" "interact click \"Submit\" --position TOP_RIGHT --dry-run"
test_command "click middle-left" "interact click \"Submit\" --position MIDDLE_LEFT --dry-run"
test_command "click middle-right" "interact click \"Submit\" --position MIDDLE_RIGHT --dry-run"
test_command "click bottom-left" "interact click \"Submit\" --position BOTTOM_LEFT --dry-run"
test_command "click bottom-center" "interact click \"Submit\" --position BOTTOM_CENTER --dry-run"
test_command "click bottom-right" "interact click \"Submit\" --position BOTTOM_RIGHT --dry-run"

print_subsection "Click Modifiers"
test_command "double-click" "interact double-click \"Item\" --dry-run"
test_command "right-click" "interact right-click \"Menu\" --dry-run"
test_command "ctrl-click" "interact click \"Link\" --ctrl --dry-run"
test_command "shift-click" "interact click \"Item\" --shift --dry-run"
test_command "alt-click" "interact click \"Option\" --alt --dry-run"
test_command "meta-click" "interact click \"Link\" --meta --dry-run"
test_command "multiple modifiers" "interact click \"Text\" --ctrl --shift --dry-run"

print_subsection "Write/Type Commands"
test_all_formats "write text" "interact write \"username\" \"testuser\" --dry-run"
test_command "write with clear" "interact write \"#input\" \"new text\" --clear --dry-run"
test_command "write with element type" "interact write \"Email\" \"test@example.com\" --element-type INPUT --dry-run"
test_command "type command" "interact type \"password\" \"secret123\" --dry-run"

print_subsection "Keyboard Commands"
test_command "press single key" "interact press \"Enter\" --dry-run"
test_command "press with selector" "interact press \"Tab\" --selector \"#form\" --dry-run"
test_command "press escape" "interact press \"Escape\" --dry-run"
test_command "press space" "interact press \"Space\" --dry-run"
test_command "press arrow keys" "interact press \"ArrowDown\" --dry-run"

print_subsection "Other Interactions"
test_all_formats "hover" "interact hover \"Menu Item\" --dry-run"
test_all_formats "focus" "interact focus \"#input\" --dry-run"
test_all_formats "blur" "interact blur \"#input\" --dry-run"
test_all_formats "clear" "interact clear \"#search\" --dry-run"
test_all_formats "check" "interact check \"#terms\" --dry-run"
test_all_formats "uncheck" "interact uncheck \"#newsletter\" --dry-run"

print_section "TESTING NAVIGATE COMMANDS"

print_subsection "Basic Navigation"
test_all_formats "navigate to URL" "navigate to \"$TEST_URL\" --dry-run"
test_command "navigate with new tab" "navigate to \"$TEST_URL\" --new-tab --dry-run"
test_command "navigate with new window" "navigate to \"$TEST_URL\" --new-window --dry-run"
test_command "navigate with incognito" "navigate to \"$TEST_URL\" --incognito --dry-run"

print_subsection "Browser Navigation"
test_all_formats "navigate back" "navigate back --dry-run"
test_all_formats "navigate forward" "navigate forward --dry-run"
test_all_formats "navigate refresh" "navigate refresh --dry-run"
test_command "navigate hard refresh" "navigate refresh --hard --dry-run"

print_subsection "Multi-Step Navigation"
test_command "navigate back 3 steps" "navigate back --steps 3 --dry-run"
test_command "navigate forward 2 steps" "navigate forward --steps 2 --dry-run"

print_subsection "Scroll Operations"
test_all_formats "scroll to element" "navigate scroll \"#footer\" --dry-run"
test_command "scroll with offset" "navigate scroll \"#section\" --offset 100 --dry-run"
test_command "scroll to top" "navigate scroll-to-top --dry-run"
test_command "scroll to bottom" "navigate scroll-to-bottom --dry-run"
test_command "scroll by pixels" "navigate scroll-by --x 0 --y 500 --dry-run"

print_section "TESTING DATA COMMANDS"

print_subsection "Store Operations"
test_all_formats "store text" "data store \"username\" --from \"#user\" --dry-run"
test_command "store with custom selector" "data store \"value\" --from \"div.content\" --use-custom-selector --dry-run"
test_command "store attribute" "data store \"link\" --from \"a\" --attribute \"href\" --dry-run"
test_command "store as variable" "data store-as \"myVar\" --from \"#data\" --dry-run"

print_subsection "Enhanced Cookie Operations"
test_all_formats "get all cookies" "data cookies get-all --dry-run"
test_command "get specific cookie" "data cookies get \"session_id\" --dry-run"
test_command "set cookie" "data cookies set \"user_pref\" \"dark_mode\" --dry-run"
test_command "set cookie with options" "data cookies set \"auth\" \"token123\" --domain \".example.com\" --path \"/\" --secure --http-only --same-site \"Strict\" --dry-run"
test_command "delete cookie" "data cookies delete \"temp_data\" --dry-run"
test_command "delete all cookies" "data cookies delete-all --dry-run"

print_subsection "Export Operations"
test_command "export data" "data export --format json --dry-run"
test_command "export CSV" "data export --format csv --dry-run"
test_command "export with filter" "data export --format json --filter \"user_*\" --dry-run"

print_section "TESTING DIALOG COMMANDS"

print_subsection "Alert Dialogs"
test_all_formats "accept alert" "dialog alert accept --dry-run"
test_all_formats "dismiss alert" "dialog alert dismiss --dry-run"
test_command "get alert text" "dialog alert get-text --dry-run"

print_subsection "Confirm Dialogs"
test_all_formats "accept confirm" "dialog confirm accept --dry-run"
test_all_formats "dismiss confirm" "dialog confirm dismiss --dry-run"

print_subsection "Prompt Dialogs"
test_all_formats "answer prompt" "dialog prompt answer \"My response\" --dry-run"
test_command "dismiss prompt" "dialog prompt dismiss --dry-run"

print_section "TESTING WAIT COMMANDS"

print_subsection "Element Wait"
test_all_formats "wait for element" "wait element \"#loading\" --dry-run"
test_command "wait with timeout" "wait element \"#result\" --timeout 10000 --dry-run"
test_command "wait for visibility" "wait element \"#modal\" --state visible --dry-run"
test_command "wait for hidden" "wait element \"#spinner\" --state hidden --dry-run"
test_command "wait for enabled" "wait element \"#submit\" --state enabled --dry-run"
test_command "wait for disabled" "wait element \"#submit\" --state disabled --dry-run"

print_subsection "Not Visible Wait"
test_all_formats "wait not visible" "wait not-visible \"#loader\" --dry-run"
test_command "wait not visible with timeout" "wait not-visible \"#popup\" --timeout 5000 --dry-run"

print_subsection "Time Wait"
test_all_formats "wait time" "wait time 2000 --dry-run"
test_command "wait 5 seconds" "wait time 5000 --dry-run"

print_subsection "Advanced Wait"
test_command "wait for text" "wait for-text \"Success\" --dry-run"
test_command "wait for URL" "wait for-url \"*/success\" --dry-run"
test_command "wait for title" "wait for-title \"Dashboard\" --dry-run"

print_section "TESTING WINDOW COMMANDS"

print_subsection "Window Management"
test_all_formats "close window" "window close --dry-run"
test_command "close specific tab" "window close --tab 2 --dry-run"
test_command "close all except current" "window close-others --dry-run"

print_subsection "Window Switching"
test_all_formats "switch to tab" "window switch --tab 1 --dry-run"
test_command "switch by title" "window switch --title \"Dashboard\" --dry-run"
test_command "switch by URL" "window switch --url \"*/admin\" --dry-run"
test_command "switch to new window" "window switch-to-new --dry-run"

print_subsection "Window Resizing"
test_all_formats "resize window" "window resize --width 1280 --height 720 --dry-run"
test_command "maximize window" "window maximize --dry-run"
test_command "minimize window" "window minimize --dry-run"
test_command "fullscreen window" "window fullscreen --dry-run"

print_subsection "Frame Switching"
test_all_formats "switch to frame" "window switch-to-frame --index 0 --dry-run"
test_command "switch by frame name" "window switch-to-frame --name \"content\" --dry-run"
test_command "switch by frame selector" "window switch-to-frame --selector \"#iframe\" --dry-run"
test_command "switch to parent frame" "window switch-to-parent --dry-run"
test_command "switch to main frame" "window switch-to-main --dry-run"

print_section "TESTING MOUSE COMMANDS"

print_subsection "Mouse Movement"
test_all_formats "move to element" "mouse move \"#target\" --dry-run"
test_command "move with offset" "mouse move \"#button\" --offset-x 10 --offset-y 5 --dry-run"
test_command "move to coordinates" "mouse move-to --x 500 --y 300 --dry-run"

print_subsection "Drag and Drop"
test_all_formats "drag and drop" "mouse drag \"#source\" --to \"#target\" --dry-run"
test_command "drag with offset" "mouse drag \"#item\" --offset-x 100 --offset-y 0 --dry-run"

print_subsection "Advanced Mouse"
test_command "mouse down" "mouse down \"#element\" --dry-run"
test_command "mouse up" "mouse up \"#element\" --dry-run"
test_command "mouse wheel" "mouse wheel --delta-x 0 --delta-y 100 --dry-run"

print_section "TESTING SELECT COMMANDS"

print_subsection "Select by Value"
test_all_formats "select by value" "select by-value \"#dropdown\" \"option1\" --dry-run"
test_command "select multiple values" "select by-value \"#multi\" \"opt1,opt2,opt3\" --multiple --dry-run"

print_subsection "Select by Text"
test_all_formats "select by text" "select by-text \"#dropdown\" \"First Option\" --dry-run"
test_command "select by partial text" "select by-text \"#dropdown\" \"First\" --partial --dry-run"

print_subsection "Select by Index"
test_all_formats "select by index" "select by-index \"#dropdown\" 0 --dry-run"
test_command "select multiple indices" "select by-index \"#multi\" \"0,2,4\" --multiple --dry-run"

print_subsection "Deselect Operations"
test_command "deselect all" "select deselect-all \"#multi\" --dry-run"
test_command "deselect by value" "select deselect-by-value \"#multi\" \"opt2\" --dry-run"
test_command "deselect by text" "select deselect-by-text \"#multi\" \"Option 2\" --dry-run"
test_command "deselect by index" "select deselect-by-index \"#multi\" 1 --dry-run"

print_section "TESTING FILE COMMANDS"

print_subsection "File Upload"
test_all_formats "upload file" "file upload \"#file-input\" \"/path/to/file.pdf\" --dry-run"
test_command "upload multiple files" "file upload \"#multi-file\" \"/path/file1.pdf,/path/file2.pdf\" --multiple --dry-run"
test_command "upload with drag-drop" "file upload \"#dropzone\" \"/path/to/image.png\" --drag-drop --dry-run"

print_subsection "File Download"
test_command "download file" "file download \"Download PDF\" --dry-run"
test_command "download with custom path" "file download \"Export\" --path \"/tmp/export.csv\" --dry-run"
test_command "download with wait" "file download \"Generate Report\" --wait-for-download --dry-run"

print_section "TESTING MISC COMMANDS"

print_subsection "Comments"
test_all_formats "add comment" "misc comment \"This is a test comment\" --dry-run"
test_command "multi-line comment" "misc comment \"Line 1\nLine 2\nLine 3\" --dry-run"

print_subsection "JavaScript Execution"
test_all_formats "execute JS" "misc execute-js \"return document.title\" --dry-run"
test_command "execute JS with args" "misc execute-js \"arguments[0].click()\" --args \"#button\" --dry-run"
test_command "execute async JS" "misc execute-js \"setTimeout(() => arguments[0](), 1000)\" --async --dry-run"

print_subsection "Screenshots"
test_command "take screenshot" "misc screenshot --dry-run"
test_command "screenshot with name" "misc screenshot --name \"login_page\" --dry-run"
test_command "element screenshot" "misc screenshot --element \"#form\" --dry-run"

print_subsection "Debugging"
test_command "pause execution" "misc pause --duration 2000 --dry-run"
test_command "add breakpoint" "misc breakpoint --dry-run"
test_command "log message" "misc log \"Debug message\" --level \"info\" --dry-run"

print_section "TESTING LIBRARY COMMANDS"

print_subsection "Component Usage"
test_all_formats "use component" "library use \"login-flow\" --dry-run"
test_command "use with parameters" "library use \"checkout\" --params '{\"product\":\"ABC123\"}' --dry-run"
test_command "use specific version" "library use \"navigation\" --version \"2.0\" --dry-run"

print_subsection "Component Management"
test_command "list components" "library list --dry-run"
test_command "list with filter" "library list --filter \"auth\" --dry-run"
test_command "get component info" "library info \"login-flow\" --dry-run"
test_command "validate component" "library validate \"checkout\" --dry-run"

print_section "TESTING EDGE CASES"

print_subsection "Special Characters and Escaping"
test_command "selector with quotes" 'assert exists "Button with \"quotes\"" --dry-run'
test_command "selector with special chars" "assert exists \"Price: \$99.99\" --dry-run"
test_command "Unicode characters" "interact write \"#input\" \"Hello 世界 🌍\" --dry-run"

print_subsection "Empty and Null Values"
test_command "empty string write" "interact write \"#input\" \"\" --dry-run"
test_command "assert empty" "assert equals \"#field\" \"\" --dry-run"

print_subsection "Large Values"
test_command "long timeout" "wait element \"#slow\" --timeout 600000 --dry-run"
test_command "large coordinate" "mouse move-to --x 9999 --y 9999 --dry-run"

print_subsection "Invalid Commands (Expected Failures)"
test_command "invalid command group" "invalid-group test --dry-run" false
test_command "missing required arg" "assert exists --dry-run" false
test_command "invalid output format" "assert exists \"test\" --output invalid --dry-run" false

print_section "TESTING SESSION CONTEXT"

print_subsection "Session Management"
test_command "get session info" "get-session-info --dry-run"
test_command "with explicit checkpoint" "assert exists \"test\" --checkpoint \"$CHECKPOINT_NAME\" --dry-run"
test_command "with explicit position" "assert exists \"test\" --position 10 --dry-run"

# Clear session for next tests
unset VIRTUOSO_SESSION_ID

print_subsection "No Session Context"
test_command "command without session" "assert exists \"test\" --checkpoint \"$CHECKPOINT_NAME\" --position 1 --dry-run"

print_section "TESTING COMPLEX SCENARIOS"

print_subsection "Chained Commands with Session"
export VIRTUOSO_SESSION_ID="$CHECKPOINT_NAME"
test_command "chained 1: navigate" "navigate to \"$TEST_URL\" --dry-run"
test_command "chained 2: wait" "wait element \"#content\" --dry-run"
test_command "chained 3: assert" "assert exists \"Welcome\" --dry-run"
test_command "chained 4: interact" "interact click \"Login\" --dry-run"
test_command "chained 5: write" "interact write \"#username\" \"testuser\" --dry-run"

print_subsection "All Flags Combined"
test_command "maximum flags" "interact click \"Submit\" --position CENTER --element-type BUTTON --use-custom-selector --checkpoint \"$CHECKPOINT_NAME\" --output json --dry-run"

print_section "CLEANUP (if enabled)"

if [ "$CLEANUP" = true ]; then
    print_subsection "Cleaning up test resources"
    test_command "Delete checkpoint" "delete-checkpoint \"$CHECKPOINT_NAME\" --dry-run"
    test_command "Delete journey" "delete-journey \"$JOURNEY_NAME\" --dry-run"
    test_command "Delete goal" "delete-goal \"$GOAL_NAME\" --dry-run"
    test_command "Delete project" "delete-project \"$PROJECT_NAME\" --dry-run"
fi

print_section "TEST RESULTS SUMMARY"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}FINAL TEST RESULTS${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total Tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"
echo -e "Success Rate: $(awk "BEGIN {printf \"%.2f\", $PASSED_TESTS/$TOTAL_TESTS*100}")%"

if [ ${#FAILED_COMMANDS[@]} -gt 0 ]; then
    echo -e "\n${RED}Failed Commands:${NC}"
    for cmd in "${FAILED_COMMANDS[@]}"; do
        echo -e "  - $cmd"
    done
fi

echo -e "\nDetailed results saved to: ${LOG_FILE}"
echo -e "\n${BLUE}Test completed at: $(date)${NC}"

# Exit with appropriate code
if [ $FAILED_TESTS -gt 0 ]; then
    exit 1
else
    exit 0
fi
