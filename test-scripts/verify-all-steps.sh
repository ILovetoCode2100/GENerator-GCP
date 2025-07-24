#!/bin/bash

# Script to verify all step types are created correctly
# Creates one of each step type and verifies the output

set -euo pipefail

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Use the existing checkpoint from earlier test
CHECKPOINT_ID="${1:-1682332}"

echo -e "${BLUE}========================================"
echo "Verifying All Step Types"
echo "========================================"
echo -e "Checkpoint ID: $CHECKPOINT_ID${NC}"
echo ""

# Counter for position
POS=100

# Function to create and verify a step
verify_step() {
    local desc="$1"
    local cmd="$2"

    echo -e "\n${YELLOW}Creating:${NC} $desc"
    echo "Command: $cmd"

    # Run command and capture output
    if output=$(eval "$cmd" 2>&1); then
        # Extract key information
        step_id=$(echo "$output" | grep "ID:" | awk '{print $2}')
        step_type=$(echo "$output" | grep "Type:" | awk '{print $2}')
        description=$(echo "$output" | grep "Description:" | sed 's/Description: //')
        selector=$(echo "$output" | grep "Selector:" | sed 's/Selector: //')

        echo -e "${GREEN}✓ Created${NC} - Step ID: $step_id"
        echo "  Type: $step_type"
        echo "  Description: $description"
        [ -n "$selector" ] && [ "$selector" != "Selector:" ] && echo "  Selector: $selector"

        # Show meta if it contains coordinates
        if echo "$output" | grep -q "Meta:"; then
            if echo "$output" | grep -q -E "(x:|y:)"; then
                echo "  Coordinates:"
                echo "$output" | grep -E "(x:|y:)" | sed 's/^/    /'
            fi
        fi
    else
        echo -e "${RED}✗ Failed${NC}"
        echo "$output" | head -5
    fi

    POS=$((POS + 1))
}

# Test Navigation Commands
echo -e "\n${BLUE}=== Navigation Commands ===${NC}"
verify_step "Navigate to URL" "./bin/api-cli step-navigate to $CHECKPOINT_ID 'https://example.com/verify' $POS"
verify_step "Navigate with new tab" "./bin/api-cli step-navigate to $CHECKPOINT_ID 'https://example.com/newtab' $((POS+1)) --new-tab"
verify_step "Scroll to top" "./bin/api-cli step-navigate scroll-top $CHECKPOINT_ID $((POS+2))"
verify_step "Scroll to bottom" "./bin/api-cli step-navigate scroll-bottom $CHECKPOINT_ID $((POS+3))"
verify_step "Scroll to h1 element" "./bin/api-cli step-navigate scroll-element $CHECKPOINT_ID 'h1' $((POS+4))"
verify_step "Scroll to position 150,250" "./bin/api-cli step-navigate scroll-position $CHECKPOINT_ID '150,250' $((POS+5))"
verify_step "Scroll by offset 50,400" "./bin/api-cli step-navigate scroll-by $CHECKPOINT_ID '50,400' $((POS+6))"
verify_step "Scroll up" "./bin/api-cli step-navigate scroll-up $CHECKPOINT_ID $((POS+7))"
verify_step "Scroll down" "./bin/api-cli step-navigate scroll-down $CHECKPOINT_ID $((POS+8))"
POS=$((POS+9))

# Test Interaction Commands
echo -e "\n${BLUE}=== Interaction Commands ===${NC}"
verify_step "Click button" "./bin/api-cli step-interact click $CHECKPOINT_ID 'button.verify' $POS"
verify_step "Double click card" "./bin/api-cli step-interact double-click $CHECKPOINT_ID 'div.card-verify' $((POS+1))"
verify_step "Right click menu" "./bin/api-cli step-interact right-click $CHECKPOINT_ID 'div.context-menu' $((POS+2))"
verify_step "Hover tooltip" "./bin/api-cli step-interact hover $CHECKPOINT_ID 'span.tooltip-verify' $((POS+3))"
verify_step "Write text" "./bin/api-cli step-interact write $CHECKPOINT_ID 'input#verify-field' 'Verification Text' $((POS+4))"
verify_step "Press Enter key" "./bin/api-cli step-interact key $CHECKPOINT_ID 'Enter' $((POS+5))"
POS=$((POS+6))

# Test Mouse Commands
echo -e "\n${BLUE}=== Mouse Commands ===${NC}"
verify_step "Mouse move to nav" "./bin/api-cli step-interact mouse move-to $CHECKPOINT_ID 'nav.verify-menu' $POS"
verify_step "Mouse move by 75,125" "./bin/api-cli step-interact mouse move-by $CHECKPOINT_ID '75,125' $((POS+1))"
verify_step "Mouse move to 300,400" "./bin/api-cli step-interact mouse move $CHECKPOINT_ID '300,400' $((POS+2))"
verify_step "Mouse down" "./bin/api-cli step-interact mouse down $CHECKPOINT_ID $((POS+3))"
verify_step "Mouse up" "./bin/api-cli step-interact mouse up $CHECKPOINT_ID $((POS+4))"
POS=$((POS+5))

# Test Select Commands
echo -e "\n${BLUE}=== Select Commands ===${NC}"
verify_step "Select by option" "./bin/api-cli step-interact select option $CHECKPOINT_ID 'select#verify-country' 'Canada' $POS"
verify_step "Select by index" "./bin/api-cli step-interact select index $CHECKPOINT_ID 'select#verify-lang' 2 $((POS+1))"
verify_step "Select last" "./bin/api-cli step-interact select last $CHECKPOINT_ID 'select#verify-zone' $((POS+2))"
POS=$((POS+3))

# Test Assertion Commands
echo -e "\n${BLUE}=== Assertion Commands ===${NC}"
verify_step "Assert exists" "./bin/api-cli step-assert exists $CHECKPOINT_ID 'h1.verify-title' $POS"
verify_step "Assert not exists" "./bin/api-cli step-assert not-exists $CHECKPOINT_ID 'div.verify-error' $((POS+1))"
verify_step "Assert equals" "./bin/api-cli step-assert equals $CHECKPOINT_ID 'h2.verify-heading' 'Verified' $((POS+2))"
verify_step "Assert checked" "./bin/api-cli step-assert checked $CHECKPOINT_ID 'input#verify-checkbox' $((POS+3))"
verify_step "Assert greater than" "./bin/api-cli step-assert gt $CHECKPOINT_ID 'span.verify-count' '25' $((POS+4))"
POS=$((POS+5))

# Test Wait Commands
echo -e "\n${BLUE}=== Wait Commands ===${NC}"
verify_step "Wait for element" "./bin/api-cli step-wait element $CHECKPOINT_ID 'div.verify-loaded' $POS"
verify_step "Wait 2 seconds" "./bin/api-cli step-wait time $CHECKPOINT_ID 2000 $((POS+1))"
POS=$((POS+2))

# Test Window Commands
echo -e "\n${BLUE}=== Window Commands ===${NC}"
verify_step "Resize window" "./bin/api-cli step-window resize $CHECKPOINT_ID '1280x720' $POS"
verify_step "Maximize window" "./bin/api-cli step-window maximize $CHECKPOINT_ID $((POS+1))"
verify_step "Switch tab next" "./bin/api-cli step-window switch tab next $CHECKPOINT_ID $((POS+2))"
verify_step "Switch to iframe" "./bin/api-cli step-window switch iframe $CHECKPOINT_ID 'iframe#verify-frame' $((POS+3))"
POS=$((POS+4))

# Test Data Commands
echo -e "\n${BLUE}=== Data Commands ===${NC}"
verify_step "Store text" "./bin/api-cli step-data store element-text $CHECKPOINT_ID 'h1.verify' 'verifyTitle' $POS"
verify_step "Store value" "./bin/api-cli step-data store element-value $CHECKPOINT_ID 'input#verify' 'verifyValue' $((POS+1))"
verify_step "Create cookie" "./bin/api-cli step-data cookie create $CHECKPOINT_ID 'verify_session' 'test123' $((POS+2))"
POS=$((POS+3))

# Test Dialog Commands
echo -e "\n${BLUE}=== Dialog Commands ===${NC}"
verify_step "Dismiss alert" "./bin/api-cli step-dialog dismiss-alert $CHECKPOINT_ID $POS"
verify_step "Accept confirm" "./bin/api-cli step-dialog dismiss-confirm $CHECKPOINT_ID $((POS+1)) --accept"
POS=$((POS+2))

# Test File Commands
echo -e "\n${BLUE}=== File Commands ===${NC}"
verify_step "Upload file" "./bin/api-cli step-file upload $CHECKPOINT_ID 'input[type=file]' 'https://example.com/verify.pdf' $POS"
POS=$((POS+1))

# Test Misc Commands
echo -e "\n${BLUE}=== Misc Commands ===${NC}"
verify_step "Add comment" "./bin/api-cli step-misc comment $CHECKPOINT_ID 'Verification test complete' $POS"
verify_step "Execute JS" "./bin/api-cli step-misc execute $CHECKPOINT_ID 'console.log(\"Verified\")' $((POS+1))"

echo -e "\n${BLUE}========================================"
echo -e "Verification Complete${NC}"
echo ""
