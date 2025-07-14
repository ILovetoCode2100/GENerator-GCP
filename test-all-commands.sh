#!/bin/bash

# Test script for all Virtuoso API CLI commands
# This script tests all 63 commands with example parameters
# Updated to test both old (legacy) and new (consolidated) command formats

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0
TOTAL=0
LEGACY_PASSED=0
LEGACY_FAILED=0
CONSOLIDATED_PASSED=0
CONSOLIDATED_FAILED=0

# Variables to store created IDs
PROJECT_ID=""
GOAL_ID=""
JOURNEY_ID=""
CHECKPOINT_ID=""
POSITION=1

# Generate unique names with timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
PROJECT_NAME="Test Project $TIMESTAMP"
GOAL_NAME="Test Goal $TIMESTAMP"
JOURNEY_NAME="Test Journey $TIMESTAMP"

# Function to test a command
test_command() {
    local cmd="$1"
    local description="$2"
    TOTAL=$((TOTAL + 1))

    echo -n "Testing: $description... "

    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
        LEGACY_PASSED=$((LEGACY_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Command: $cmd"
        FAILED=$((FAILED + 1))
        LEGACY_FAILED=$((LEGACY_FAILED + 1))
    fi
}

# Function to test both legacy and consolidated commands
test_dual_command() {
    local legacy_cmd="$1"
    local consolidated_cmd="$2"
    local description="$3"

    # Test legacy command
    TOTAL=$((TOTAL + 1))
    echo -n "Testing (Legacy): $description... "

    if eval "$legacy_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
        LEGACY_PASSED=$((LEGACY_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Command: $legacy_cmd"
        FAILED=$((FAILED + 1))
        LEGACY_FAILED=$((LEGACY_FAILED + 1))
    fi

    # Test consolidated command if provided
    if [ -n "$consolidated_cmd" ]; then
        TOTAL=$((TOTAL + 1))
        echo -n "Testing (New): $description... "

        if eval "$consolidated_cmd" > /dev/null 2>&1; then
            echo -e "${GREEN}PASSED${NC}"
            PASSED=$((PASSED + 1))
            CONSOLIDATED_PASSED=$((CONSOLIDATED_PASSED + 1))
        else
            echo -e "${RED}FAILED${NC}"
            echo "  Command: $consolidated_cmd"
            FAILED=$((FAILED + 1))
            CONSOLIDATED_FAILED=$((CONSOLIDATED_FAILED + 1))
        fi
    fi
}

# Function to test command with expected failure
test_command_expect_fail() {
    local cmd="$1"
    local description="$2"
    TOTAL=$((TOTAL + 1))

    echo -n "Testing: $description (expect fail)... "

    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${RED}UNEXPECTED SUCCESS${NC}"
        FAILED=$((FAILED + 1))
    else
        echo -e "${GREEN}FAILED AS EXPECTED${NC}"
        PASSED=$((PASSED + 1))
    fi
}

echo "=== Testing All Virtuoso API CLI Commands ==="
echo ""

# Function to extract ID from JSON output
extract_id() {
    local json="$1"
    local field="$2"
    # Handle both numeric IDs and string IDs (with quotes)
    local result=$(echo "$json" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*[0-9]*" | grep -o "[0-9]*$")
    if [ -z "$result" ]; then
        # Try extracting string value
        result=$(echo "$json" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | sed 's/.*"\([^"]*\)"$/\1/')
    fi
    echo "$result"
}

# Step 1: Create a new project
echo -e "${BLUE}Step 1: Creating new project...${NC}"
if PROJECT_OUTPUT=$(./bin/api-cli create-project "$PROJECT_NAME" --output json 2>&1); then
    PROJECT_ID=$(extract_id "$PROJECT_OUTPUT" "project_id")
    if [ -n "$PROJECT_ID" ]; then
        echo -e "${GREEN}✓ Created project: $PROJECT_NAME (ID: $PROJECT_ID)${NC}"
    else
        echo -e "${RED}✗ Failed to extract project ID${NC}"
        echo "Output: $PROJECT_OUTPUT"
        exit 1
    fi
else
    echo -e "${RED}✗ Failed to create project${NC}"
    echo "Output: $PROJECT_OUTPUT"
    exit 1
fi

# Step 2: Create a goal in the project
echo -e "${BLUE}Step 2: Creating goal in project...${NC}"
if GOAL_OUTPUT=$(./bin/api-cli create-goal "$PROJECT_ID" "$GOAL_NAME" --output json 2>&1); then
    GOAL_ID=$(extract_id "$GOAL_OUTPUT" "goal_id")
    SNAPSHOT_ID=$(extract_id "$GOAL_OUTPUT" "snapshot_id")
    if [ -n "$GOAL_ID" ]; then
        echo -e "${GREEN}✓ Created goal: $GOAL_NAME (ID: $GOAL_ID, Snapshot: $SNAPSHOT_ID)${NC}"
    else
        echo -e "${RED}✗ Failed to extract goal ID${NC}"
        echo "Output: $GOAL_OUTPUT"
        exit 1
    fi
else
    echo -e "${RED}✗ Failed to create goal${NC}"
    echo "Output: $GOAL_OUTPUT"
    exit 1
fi

# Step 3: Create a journey in the goal
echo -e "${BLUE}Step 3: Creating journey in goal...${NC}"
if JOURNEY_OUTPUT=$(./bin/api-cli create-journey "$GOAL_ID" "$SNAPSHOT_ID" "$JOURNEY_NAME" --output json 2>&1); then
    JOURNEY_ID=$(extract_id "$JOURNEY_OUTPUT" "journey_id")
    if [ -n "$JOURNEY_ID" ]; then
        echo -e "${GREEN}✓ Created journey: $JOURNEY_NAME (ID: $JOURNEY_ID)${NC}"
    else
        echo -e "${RED}✗ Failed to extract journey ID${NC}"
        echo "Output: $JOURNEY_OUTPUT"
        exit 1
    fi
else
    echo -e "${RED}✗ Failed to create journey${NC}"
    echo "Output: $JOURNEY_OUTPUT"
    exit 1
fi

# Step 4: Create a checkpoint in the journey
echo -e "${BLUE}Step 4: Creating checkpoint in journey...${NC}"
if CHECKPOINT_OUTPUT=$(./bin/api-cli create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "Test Checkpoint" --output json 2>&1); then
    CHECKPOINT_ID=$(extract_id "$CHECKPOINT_OUTPUT" "checkpoint_id")
    if [ -n "$CHECKPOINT_ID" ]; then
        echo -e "${GREEN}✓ Created checkpoint (ID: $CHECKPOINT_ID)${NC}"
    else
        echo -e "${RED}✗ Failed to extract checkpoint ID${NC}"
        echo "Output: $CHECKPOINT_OUTPUT"
        exit 1
    fi
else
    echo -e "${RED}✗ Failed to create checkpoint${NC}"
    echo "Output: $CHECKPOINT_OUTPUT"
    exit 1
fi

# Set the checkpoint in session context for commands that support it
echo -e "${BLUE}Step 5: Setting checkpoint in session context...${NC}"
if ./bin/api-cli set-checkpoint "$CHECKPOINT_ID" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Set checkpoint $CHECKPOINT_ID in session context${NC}"
else
    echo -e "${YELLOW}⚠ Could not set checkpoint in session context (will use explicit IDs)${NC}"
fi

echo ""
echo -e "${GREEN}=== Setup Complete ===${NC}"
echo "Project ID: $PROJECT_ID"
echo "Goal ID: $GOAL_ID"
echo "Journey ID: $JOURNEY_ID"
echo "Checkpoint ID: $CHECKPOINT_ID"
echo ""
echo -e "${BLUE}=== Testing Step Creation Commands ===${NC}"
echo ""

# Test navigation command
test_command "./bin/api-cli create-step-navigate $CHECKPOINT_ID 'https://example.com' $POSITION --output json" "create-step-navigate (basic)"
test_command "./bin/api-cli create-step-navigate $CHECKPOINT_ID 'https://example.com' $POSITION --new-tab --output json" "create-step-navigate (new tab)"

# Test click command
test_command "./bin/api-cli create-step-click $CHECKPOINT_ID 'button.submit' $POSITION --output json" "create-step-click (basic)"
test_command "./bin/api-cli create-step-click $CHECKPOINT_ID 'button.submit' $POSITION --variable 'clickResult' --output json" "create-step-click (with variable)"

# Test write command
test_command "./bin/api-cli create-step-write $CHECKPOINT_ID 'input#username' 'testuser' $POSITION --output json" "create-step-write (basic)"
test_command "./bin/api-cli create-step-write $CHECKPOINT_ID 'input#username' 'testuser' $POSITION --variable 'username' --output json" "create-step-write (with variable)"

# Test cookie commands
test_command "./bin/api-cli create-step-cookie-create $CHECKPOINT_ID 'session' 'abc123' $POSITION --output json" "create-step-cookie-create"
test_command "./bin/api-cli create-step-cookie-wipe-all $CHECKPOINT_ID $POSITION --output json" "create-step-cookie-wipe-all"
test_command "./bin/api-cli create-step-add-cookie 'session' 'abc123' 'example.com' $POSITION --output json" "create-step-add-cookie"
test_command "./bin/api-cli create-step-delete-cookie 'session' $POSITION --output json" "create-step-delete-cookie"
test_command "./bin/api-cli create-step-clear-cookies $POSITION --output json" "create-step-clear-cookies"

# Test upload commands
test_command "./bin/api-cli create-step-upload-url $CHECKPOINT_ID 'https://example.com/file.pdf' 'input[type=file]' $POSITION --output json" "create-step-upload-url"
test_command "./bin/api-cli create-step-upload 'https://example.com/dummy-file.pdf' 'input[type=file]' $POSITION --output json" "create-step-upload"

# Test script execution
test_command "./bin/api-cli create-step-execute-script $CHECKPOINT_ID 'console.log(\"test\")' $POSITION --output json" "create-step-execute-script"
test_command "./bin/api-cli create-step-execute-js 'return document.title' $POSITION --output json" "create-step-execute-js (no variable)"
test_command "./bin/api-cli create-step-execute-js 'return document.title' 'pageTitle' $POSITION --output json" "create-step-execute-js (with variable)"

# Test dismiss commands
test_command "./bin/api-cli create-step-dismiss-prompt-with-text $CHECKPOINT_ID 'OK' $POSITION --output json" "create-step-dismiss-prompt-with-text"
test_command "./bin/api-cli create-step-dismiss-alert $POSITION --output json" "create-step-dismiss-alert"
test_command "./bin/api-cli create-step-dismiss-confirm $POSITION --output json" "create-step-dismiss-confirm (cancel)"
test_command "./bin/api-cli create-step-dismiss-confirm $POSITION --accept --output json" "create-step-dismiss-confirm (accept)"
test_command "./bin/api-cli create-step-dismiss-prompt 'test input' $POSITION --output json" "create-step-dismiss-prompt"

# Test pick commands
test_command "./bin/api-cli create-step-pick-index $CHECKPOINT_ID 'select#dropdown' 2 $POSITION --output json" "create-step-pick-index"
test_command "./bin/api-cli create-step-pick-last $CHECKPOINT_ID 'select#dropdown' $POSITION --output json" "create-step-pick-last"
test_command "./bin/api-cli create-step-pick 'select#dropdown' 'option-value' $POSITION --output json" "create-step-pick"
test_command "./bin/api-cli create-step-pick-text 'select#dropdown' 'Option Text' $POSITION --output json" "create-step-pick-text"
test_command "./bin/api-cli create-step-pick-value 'select#dropdown' 'option-value' $POSITION --output json" "create-step-pick-value"

# Test wait commands
test_command "./bin/api-cli create-step-wait-for-element-timeout $CHECKPOINT_ID 'div.loaded' 5000 $POSITION --output json" "create-step-wait-for-element-timeout"
test_command "./bin/api-cli create-step-wait-for-element-default $CHECKPOINT_ID 'div.loaded' $POSITION --output json" "create-step-wait-for-element-default"
test_command "./bin/api-cli create-step-wait-element 'div.loaded' 5000 $POSITION --output json" "create-step-wait-element"
test_command "./bin/api-cli create-step-wait-time 3000 $POSITION --output json" "create-step-wait-time"

# Test store commands
test_command "./bin/api-cli create-step-store-element-text $CHECKPOINT_ID 'h1.title' 'pageTitle' $POSITION --output json" "create-step-store-element-text"
test_command "./bin/api-cli create-step-store-literal-value $CHECKPOINT_ID 'test-value' 'myVariable' $POSITION --output json" "create-step-store-literal-value"
test_command "./bin/api-cli create-step-store 'h1.title' 'pageTitle' $POSITION --output json" "create-step-store"
test_command "./bin/api-cli create-step-store-value 'input#username' 'usernameValue' $POSITION --output json" "create-step-store-value"

# Test mouse commands
test_command "./bin/api-cli create-step-mouse-move-to $CHECKPOINT_ID 100 200 $POSITION --output json" "create-step-mouse-move-to"
test_command "./bin/api-cli create-step-mouse-move-by $CHECKPOINT_ID 50 50 $POSITION --output json" "create-step-mouse-move-by"
test_command "./bin/api-cli create-step-double-click 'button.action' $POSITION --output json" "create-step-double-click"
test_command "./bin/api-cli create-step-hover 'a.tooltip' $POSITION --output json" "create-step-hover"
test_command "./bin/api-cli create-step-mouse-down 'div.draggable' $POSITION --output json" "create-step-mouse-down"
test_command "./bin/api-cli create-step-mouse-enter 'div.hover-area' $POSITION --output json" "create-step-mouse-enter"
test_command "./bin/api-cli create-step-mouse-move 'div.target' $POSITION --output json" "create-step-mouse-move"
test_command "./bin/api-cli create-step-mouse-up 'div.drop-zone' $POSITION --output json" "create-step-mouse-up"
test_command "./bin/api-cli create-step-right-click 'div.context-menu' $POSITION --output json" "create-step-right-click"

# Test switch commands
test_command "./bin/api-cli create-step-switch-iframe $CHECKPOINT_ID 'iframe#content' $POSITION --output json" "create-step-switch-iframe"
test_command "./bin/api-cli create-step-switch-next-tab $CHECKPOINT_ID $POSITION --output json" "create-step-switch-next-tab"
test_command "./bin/api-cli create-step-switch-parent-frame $CHECKPOINT_ID $POSITION --output json" "create-step-switch-parent-frame"
test_command "./bin/api-cli create-step-switch-prev-tab $CHECKPOINT_ID $POSITION --output json" "create-step-switch-prev-tab"

# Test assertion commands
test_command "./bin/api-cli create-step-assert-not-equals $CHECKPOINT_ID 'span.status' 'error' $POSITION --output json" "create-step-assert-not-equals"
test_command "./bin/api-cli create-step-assert-greater-than $CHECKPOINT_ID 'span.count' '10' $POSITION --output json" "create-step-assert-greater-than"
test_command "./bin/api-cli create-step-assert-greater-than-or-equal $CHECKPOINT_ID 'span.count' '10' $POSITION --output json" "create-step-assert-greater-than-or-equal"
test_command "./bin/api-cli create-step-assert-matches $CHECKPOINT_ID 'span.email' '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$' $POSITION --output json" "create-step-assert-matches"
test_command "./bin/api-cli create-step-assert-checked 'input#terms' $POSITION --output json" "create-step-assert-checked"
test_command "./bin/api-cli create-step-assert-equals 'span.status' 'success' $POSITION --output json" "create-step-assert-equals"
test_command "./bin/api-cli create-step-assert-exists 'div.content' $POSITION --output json" "create-step-assert-exists"
test_command "./bin/api-cli create-step-assert-less-than 'span.count' '100' $POSITION --output json" "create-step-assert-less-than"
test_command "./bin/api-cli create-step-assert-less-than-or-equal 'span.count' '100' $POSITION --output json" "create-step-assert-less-than-or-equal"
test_command "./bin/api-cli create-step-assert-not-exists 'div.error' $POSITION --output json" "create-step-assert-not-exists"
test_command "./bin/api-cli create-step-assert-selected 'select#country' 'USA' $POSITION --output json" "create-step-assert-selected"
test_command "./bin/api-cli create-step-assert-variable 'username' 'testuser' $POSITION --output json" "create-step-assert-variable"

# Test scroll commands
test_command "./bin/api-cli create-step-scroll-to-position $CHECKPOINT_ID 0 500 $POSITION --output json" "create-step-scroll-to-position"
test_command "./bin/api-cli create-step-scroll-by-offset $CHECKPOINT_ID 0 200 $POSITION --output json" "create-step-scroll-by-offset"
test_command "./bin/api-cli create-step-scroll-to-top $CHECKPOINT_ID $POSITION --output json" "create-step-scroll-to-top"
test_command "./bin/api-cli create-step-scroll-bottom $POSITION --output json" "create-step-scroll-bottom"
test_command "./bin/api-cli create-step-scroll-element 'div#footer' $POSITION --output json" "create-step-scroll-element"
test_command "./bin/api-cli create-step-scroll-position 0 1000 $POSITION --output json" "create-step-scroll-position"
test_command "./bin/api-cli create-step-scroll-top $POSITION --output json" "create-step-scroll-top"

# Test window commands
test_command "./bin/api-cli create-step-window-resize $CHECKPOINT_ID 1280 720 $POSITION --output json" "create-step-window-resize"
test_command "./bin/api-cli create-step-window 'resize' '400x400' $POSITION --output json" "create-step-window"

# Test key command
test_command "./bin/api-cli create-step-key $CHECKPOINT_ID 'Enter' $POSITION --output json" "create-step-key (global)"
test_command "./bin/api-cli create-step-key $CHECKPOINT_ID 'Tab' $POSITION --target 'input#username' --output json" "create-step-key (targeted)"

# Test comment command
test_command "./bin/api-cli create-step-comment $CHECKPOINT_ID 'This is a test comment' $POSITION --output json" "create-step-comment"

echo ""
echo -e "${YELLOW}=== Testing Consolidated Commands (New Format) ===${NC}"
echo ""

# Test some key commands in both formats
test_dual_command \
    "./bin/api-cli create-step-assert-equals $CHECKPOINT_ID 'span.username' 'john.doe' $POSITION" \
    "./bin/api-cli assert equals 'span.username' 'john.doe' $POSITION --checkpoint $CHECKPOINT_ID" \
    "assert equals (dual format)"

test_dual_command \
    "./bin/api-cli create-step-click $CHECKPOINT_ID 'button.submit' $POSITION" \
    "./bin/api-cli interact click 'button.submit' $POSITION --checkpoint $CHECKPOINT_ID" \
    "interact click (dual format)"

test_dual_command \
    "./bin/api-cli create-step-navigate $CHECKPOINT_ID 'https://example.com' $POSITION" \
    "./bin/api-cli navigate url 'https://example.com' $POSITION --checkpoint $CHECKPOINT_ID" \
    "navigate url (dual format)"

test_dual_command \
    "./bin/api-cli create-step-write $CHECKPOINT_ID 'input#email' 'test@example.com' $POSITION" \
    "./bin/api-cli interact write 'input#email' 'test@example.com' $POSITION --checkpoint $CHECKPOINT_ID" \
    "interact write (dual format)"

test_dual_command \
    "./bin/api-cli create-step-wait-element $CHECKPOINT_ID 'div.spinner' $POSITION" \
    "./bin/api-cli wait element 'div.spinner' $POSITION --checkpoint $CHECKPOINT_ID" \
    "wait element (dual format)"

echo ""
echo "=== Test Summary ==="
echo -e "Total tests: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""
echo "Legacy Commands:"
echo -e "  Passed: ${GREEN}$LEGACY_PASSED${NC}"
echo -e "  Failed: ${RED}$LEGACY_FAILED${NC}"
echo ""
echo "Consolidated Commands:"
echo -e "  Passed: ${GREEN}$CONSOLIDATED_PASSED${NC}"
echo -e "  Failed: ${RED}$CONSOLIDATED_FAILED${NC}"

# Optional: Clean up the created test data
echo ""
echo -e "${BLUE}Cleanup: Do you want to delete the test project? (y/N)${NC}"
read -r CLEANUP_RESPONSE
if [[ "$CLEANUP_RESPONSE" =~ ^[Yy]$ ]]; then
    echo "Deleting test project..."
    if ./bin/api-cli delete-project "$PROJECT_ID" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Test project deleted${NC}"
    else
        echo -e "${YELLOW}⚠ Could not delete test project${NC}"
    fi
fi

if [ "$FAILED" -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed.${NC}"
    exit 1
fi
