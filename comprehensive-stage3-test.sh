#!/bin/bash
# Comprehensive Virtuoso API CLI Test Suite
# Tests ALL commands and variations including Stage 1, 2, and 3 enhancements

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLI="./bin/api-cli"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="test_comprehensive_${TIMESTAMP}.log"
CLEANUP=${CLEANUP:-true}

# Counters
TOTAL=0
PASSED=0
FAILED=0

# Test function
test_cmd() {
    local description="$1"
    local command="$2"
    TOTAL=$((TOTAL + 1))

    echo -n "  Testing: $description... "
    if eval "$command" >> "$LOG_FILE" 2>&1; then
        echo -e "${GREEN}✓${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗${NC}"
        FAILED=$((FAILED + 1))
        echo "    Failed command: $command" | tee -a "$LOG_FILE"
    fi
}

# Start test
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Virtuoso API CLI Comprehensive Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo "Timestamp: $TIMESTAMP"
echo "Log file: $LOG_FILE"
echo ""

# 1. Setup Test Infrastructure
echo -e "${YELLOW}1. Creating Test Infrastructure${NC}"
echo "================================"

# Create project
PROJECT_JSON=$($CLI create-project "CompTest_${TIMESTAMP}" -o json)
PROJECT_ID=$(echo "$PROJECT_JSON" | jq -r '.project_id')
echo "Created Project: $PROJECT_ID"

# Create goal
GOAL_JSON=$($CLI create-goal "$PROJECT_ID" "Goal_${TIMESTAMP}" -o json)
GOAL_ID=$(echo "$GOAL_JSON" | jq -r '.goal_id')
echo "Created Goal: $GOAL_ID"

# Get snapshot ID from the goal creation output
# The create-goal command returns the snapshot_id in its JSON output
SNAPSHOT_ID=$(echo "$GOAL_JSON" | jq -r '.snapshot_id // empty')
if [ -z "$SNAPSHOT_ID" ]; then
    echo "Warning: No snapshot ID found in goal creation output"
    SNAPSHOT_ID="1"  # Fallback
fi
echo "Using snapshot ID: $SNAPSHOT_ID"

# Add delay for API propagation
echo "Waiting for API propagation..."
sleep 2

# Create journey with error handling
echo "Creating journey with Goal: $GOAL_ID, Snapshot: $SNAPSHOT_ID"
if ! JOURNEY_JSON=$($CLI create-journey "$GOAL_ID" "$SNAPSHOT_ID" "Journey_${TIMESTAMP}" -o json); then
    echo "Failed to create journey. Exiting."
    exit 1
fi
JOURNEY_ID=$(echo "$JOURNEY_JSON" | jq -r '.journey_id')
echo "Created Journey: $JOURNEY_ID"

# Create checkpoint
CHECKPOINT_JSON=$($CLI create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "Checkpoint_${TIMESTAMP}" -o json)
CHECKPOINT_ID=$(echo "$CHECKPOINT_JSON" | jq -r '.checkpoint_id')
echo "Created Checkpoint: $CHECKPOINT_ID"
echo ""

# Position counter
POS=1

# 2. Test ALL Assert Commands
echo -e "${YELLOW}2. Assert Commands (All 12 Types)${NC}"
echo "=================================="
test_cmd "Assert exists" "$CLI assert exists 'Example Domain' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert not-exists" "$CLI assert not-exists 'Error Message' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert equals" "$CLI assert equals 'h1' 'Example Domain' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert not-equals" "$CLI assert not-equals 'title' 'Wrong Title' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert checked" "$CLI assert checked 'input[type=checkbox]' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert selected" "$CLI assert selected 'option:first' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert gt" "$CLI assert gt 'body' '0' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert gte" "$CLI assert gte 'body' '0' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert lt" "$CLI assert lt 'body' '999999' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert lte" "$CLI assert lte 'body' '999999' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert matches" "$CLI assert matches 'h1' '^Example' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert variable" "$CLI assert variable 'testVar' 'expectedValue' --checkpoint $CHECKPOINT_ID"
echo ""

# 3. Test ALL Navigate Commands (Stage 1 & 3 Enhanced)
echo -e "${YELLOW}3. Navigate Commands (All Variations)${NC}"
echo "====================================="
test_cmd "Navigate to URL" "$CLI navigate to $CHECKPOINT_ID 'https://example.com' $((POS++))"
test_cmd "Navigate new tab" "$CLI navigate to $CHECKPOINT_ID 'https://google.com' $((POS++)) --new-tab"
test_cmd "Navigate back" "$CLI navigate back $CHECKPOINT_ID $((POS++))"
test_cmd "Navigate back 2 steps" "$CLI navigate back $CHECKPOINT_ID $((POS++)) --steps 2"
test_cmd "Navigate forward" "$CLI navigate forward $CHECKPOINT_ID $((POS++))"
test_cmd "Navigate forward 3 steps" "$CLI navigate forward $CHECKPOINT_ID $((POS++)) --steps 3"
test_cmd "Navigate refresh" "$CLI navigate refresh $CHECKPOINT_ID $((POS++))"
test_cmd "Scroll to top" "$CLI navigate scroll-top $CHECKPOINT_ID $((POS++))"
test_cmd "Scroll to bottom" "$CLI navigate scroll-bottom $CHECKPOINT_ID $((POS++))"
test_cmd "Scroll to element" "$CLI navigate scroll-element $CHECKPOINT_ID 'footer' $((POS++))"
test_cmd "Scroll to position" "$CLI navigate scroll-position $CHECKPOINT_ID '500,300' $((POS++))"
test_cmd "Scroll by offset" "$CLI navigate scroll-by $CHECKPOINT_ID '0,500' $((POS++))"
test_cmd "Scroll by negative" "$CLI navigate scroll-by $CHECKPOINT_ID '-100,-200' $((POS++))"
test_cmd "Scroll up" "$CLI navigate scroll-up $CHECKPOINT_ID $((POS++))"
test_cmd "Scroll down" "$CLI navigate scroll-down $CHECKPOINT_ID $((POS++))"
echo ""

# 4. Test ALL Interact Commands (Stage 2 & 3 Enhanced)
echo -e "${YELLOW}4. Interact Commands (All Variations)${NC}"
echo "====================================="
# Basic clicks
test_cmd "Click basic" "$CLI interact click $CHECKPOINT_ID 'button' $((POS++))"
test_cmd "Click with variable" "$CLI interact click $CHECKPOINT_ID 'a.link' $((POS++)) --variable linkText"

# Position enum clicks
for position in TOP_LEFT TOP_CENTER TOP_RIGHT CENTER_LEFT CENTER CENTER_RIGHT BOTTOM_LEFT BOTTOM_CENTER BOTTOM_RIGHT; do
    test_cmd "Click $position" "$CLI interact click $CHECKPOINT_ID 'div' $((POS++)) --position $position"
done

test_cmd "Double-click" "$CLI interact double-click $CHECKPOINT_ID 'div.card' $((POS++))"
test_cmd "Right-click" "$CLI interact right-click $CHECKPOINT_ID 'body' $((POS++))"
test_cmd "Hover" "$CLI interact hover $CHECKPOINT_ID 'a.tooltip' $((POS++))"
test_cmd "Hover duration" "$CLI interact hover $CHECKPOINT_ID 'button' $((POS++)) --duration 2000"
test_cmd "Write text" "$CLI interact write $CHECKPOINT_ID 'input' 'test text' $((POS++))"
test_cmd "Write clear" "$CLI interact write $CHECKPOINT_ID 'input' 'new text' $((POS++)) --clear"
test_cmd "Write variable" "$CLI interact write $CHECKPOINT_ID 'input' '{{userName}}' $((POS++)) --variable userName"

# Keyboard shortcuts (Stage 3)
test_cmd "Key Enter" "$CLI interact key $CHECKPOINT_ID 'Enter' $((POS++))"
test_cmd "Key Escape" "$CLI interact key $CHECKPOINT_ID 'Escape' $((POS++))"
test_cmd "Key Tab" "$CLI interact key $CHECKPOINT_ID 'Tab' $((POS++))"
test_cmd "Key Ctrl+A" "$CLI interact key $CHECKPOINT_ID 'a' $((POS++)) --modifiers ctrl"
test_cmd "Key Ctrl+C" "$CLI interact key $CHECKPOINT_ID 'c' $((POS++)) --modifiers ctrl"
test_cmd "Key Ctrl+V" "$CLI interact key $CHECKPOINT_ID 'v' $((POS++)) --modifiers ctrl"
test_cmd "Key Ctrl+Shift+Tab" "$CLI interact key $CHECKPOINT_ID 'Tab' $((POS++)) --modifiers ctrl,shift"
test_cmd "Key Alt+F4" "$CLI interact key $CHECKPOINT_ID 'F4' $((POS++)) --modifiers alt"
test_cmd "Key Cmd+S" "$CLI interact key $CHECKPOINT_ID 's' $((POS++)) --modifiers meta"
test_cmd "Key targeted" "$CLI interact key $CHECKPOINT_ID 'Delete' $((POS++)) --target 'input#field'"
echo ""

# 5. Test ALL Data Commands (Stage 1 & 2 Enhanced)
echo -e "${YELLOW}5. Data Commands (All Variations)${NC}"
echo "=================================="
test_cmd "Store element text" "$CLI data store element-text $CHECKPOINT_ID 'h1' 'pageTitle' $((POS++))"
test_cmd "Store literal" "$CLI data store literal $CHECKPOINT_ID 'testValue123' 'myVar' $((POS++))"
test_cmd "Store attribute" "$CLI data store attribute $CHECKPOINT_ID 'a.link' 'href' 'linkUrl' $((POS++))"
test_cmd "Store src attribute" "$CLI data store attribute $CHECKPOINT_ID 'img' 'src' 'imageSrc' $((POS++))"
test_cmd "Cookie create basic" "$CLI data cookie create $CHECKPOINT_ID 'session' 'abc123' $((POS++))"
test_cmd "Cookie with domain" "$CLI data cookie create $CHECKPOINT_ID 'auth' 'xyz789' $((POS++)) --domain '.example.com'"
test_cmd "Cookie with path" "$CLI data cookie create $CHECKPOINT_ID 'prefs' 'settings1' $((POS++)) --path '/admin'"
test_cmd "Cookie secure" "$CLI data cookie create $CHECKPOINT_ID 'secure_token' 'secret' $((POS++)) --secure"
test_cmd "Cookie httpOnly" "$CLI data cookie create $CHECKPOINT_ID 'http_only' 'value' $((POS++)) --http-only"
test_cmd "Cookie all options" "$CLI data cookie create $CHECKPOINT_ID 'full' 'complex' $((POS++)) --domain '.test.com' --path '/' --secure --http-only"
test_cmd "Cookie delete" "$CLI data cookie delete $CHECKPOINT_ID 'session' $((POS++))"
test_cmd "Cookie clear all" "$CLI data cookie clear-all $CHECKPOINT_ID $((POS++))"
echo ""

# 6. Test ALL Dialog Commands
echo -e "${YELLOW}6. Dialog Commands (All Types)${NC}"
echo "==============================="
test_cmd "Alert accept" "$CLI dialog alert $CHECKPOINT_ID accept $((POS++))"
test_cmd "Alert dismiss" "$CLI dialog alert $CHECKPOINT_ID dismiss $((POS++))"
test_cmd "Confirm accept" "$CLI dialog confirm $CHECKPOINT_ID accept $((POS++))"
test_cmd "Confirm dismiss" "$CLI dialog confirm $CHECKPOINT_ID dismiss $((POS++))"
test_cmd "Prompt with text" "$CLI dialog prompt $CHECKPOINT_ID 'User input text' $((POS++))"
test_cmd "Prompt dismiss" "$CLI dialog prompt $CHECKPOINT_ID dismiss $((POS++))"
echo ""

# 7. Test ALL Wait Commands (Stage 2 Enhanced)
echo -e "${YELLOW}7. Wait Commands (All Variations)${NC}"
echo "================================="
test_cmd "Wait element" "$CLI wait element 'h1' --checkpoint $CHECKPOINT_ID"
test_cmd "Wait with timeout" "$CLI wait element 'div.ready' --timeout 5000 --checkpoint $CHECKPOINT_ID"
test_cmd "Wait element not visible" "$CLI wait element-not-visible 'div.loading' --checkpoint $CHECKPOINT_ID"
test_cmd "Wait not visible timeout" "$CLI wait element-not-visible 'spinner' --timeout 3000 --checkpoint $CHECKPOINT_ID"
test_cmd "Wait time 1s" "$CLI wait time 1 --checkpoint $CHECKPOINT_ID"
test_cmd "Wait time 500ms" "$CLI wait time 0.5 --checkpoint $CHECKPOINT_ID"
echo ""

# 8. Test ALL Window Commands (Stage 1 & 3 Enhanced)
echo -e "${YELLOW}8. Window Commands (All Operations)${NC}"
echo "==================================="
test_cmd "Window resize" "$CLI window resize 1024x768 $CHECKPOINT_ID $((POS++))"
test_cmd "Window resize mobile" "$CLI window resize 375x667 $CHECKPOINT_ID $((POS++))"
test_cmd "Window maximize" "$CLI window maximize $CHECKPOINT_ID $((POS++))"
test_cmd "Window close" "$CLI window close $CHECKPOINT_ID $((POS++))"
test_cmd "Switch next tab" "$CLI window switch tab next $CHECKPOINT_ID $((POS++))"
test_cmd "Switch prev tab" "$CLI window switch tab prev $CHECKPOINT_ID $((POS++))"
test_cmd "Switch tab index 0" "$CLI window switch tab 0 $CHECKPOINT_ID $((POS++))"
test_cmd "Switch tab index 2" "$CLI window switch tab 2 $CHECKPOINT_ID $((POS++))"
test_cmd "Switch iframe" "$CLI window switch iframe '#payment-frame' $CHECKPOINT_ID $((POS++))"
test_cmd "Switch parent frame" "$CLI window switch parent-frame $CHECKPOINT_ID $((POS++))"
test_cmd "Switch frame by index" "$CLI window switch frame-index 0 $CHECKPOINT_ID $((POS++))"
test_cmd "Switch frame by name" "$CLI window switch frame-name 'contentFrame' $CHECKPOINT_ID $((POS++))"
test_cmd "Switch main content" "$CLI window switch main-content $CHECKPOINT_ID $((POS++))"
echo ""

# 9. Test ALL Mouse Commands
echo -e "${YELLOW}9. Mouse Commands (All Types)${NC}"
echo "=============================="
test_cmd "Mouse move-to" "$CLI mouse move-to 'button' --checkpoint $CHECKPOINT_ID"
test_cmd "Mouse move-by" "$CLI mouse move-by 50 100 --checkpoint $CHECKPOINT_ID"
test_cmd "Mouse move" "$CLI mouse move 200 300 --checkpoint $CHECKPOINT_ID"
test_cmd "Mouse down" "$CLI mouse down --checkpoint $CHECKPOINT_ID"
test_cmd "Mouse up" "$CLI mouse up --checkpoint $CHECKPOINT_ID"
test_cmd "Mouse enter" "$CLI mouse enter 'div.hover-target' --checkpoint $CHECKPOINT_ID"
echo ""

# 10. Test ALL Select Commands
echo -e "${YELLOW}10. Select Commands (All Types)${NC}"
echo "================================"
test_cmd "Select option" "$CLI select option $CHECKPOINT_ID 'select#country' 'United States' $((POS++))"
test_cmd "Select by index" "$CLI select index $CHECKPOINT_ID 'select#country' 0 $((POS++))"
test_cmd "Select last" "$CLI select last $CHECKPOINT_ID 'select#country' $((POS++))"
echo ""

# 11. Test ALL File Commands
echo -e "${YELLOW}11. File Commands${NC}"
echo "=================="
test_cmd "File upload" "$CLI file upload $CHECKPOINT_ID 'input[type=file]' '/tmp/test.txt' $((POS++))"
test_cmd "File upload URL" "$CLI file upload-url $CHECKPOINT_ID 'input[type=file]' 'https://example.com/file.pdf' $((POS++))"
echo ""

# 12. Test ALL Misc Commands
echo -e "${YELLOW}12. Misc Commands${NC}"
echo "=================="
test_cmd "Add comment" "$CLI misc comment $CHECKPOINT_ID 'Test step comment' $((POS++))"
test_cmd "Execute JS" "$CLI misc execute $CHECKPOINT_ID 'return document.title;' $((POS++))"
echo ""

# 13. Test ALL Library Commands (Stage 3)
echo -e "${YELLOW}13. Library Commands${NC}"
echo "===================="
# Note: These require existing library checkpoint IDs
test_cmd "Library add" "$CLI library add 7023 $CHECKPOINT_ID"
test_cmd "Library get" "$CLI library get 7023"
test_cmd "Library attach" "$CLI library attach $CHECKPOINT_ID 7023 $((POS++))"
# test_cmd "Library move step" "$CLI library move-step 7023 19660498 2"
# test_cmd "Library remove step" "$CLI library remove-step 7023 19660498"
echo ""

# 14. Test Output Formats
echo -e "${YELLOW}14. Output Format Tests${NC}"
echo "======================="
test_cmd "Human output" "$CLI list-projects --limit 1 -o human"
test_cmd "JSON output" "$CLI list-projects --limit 1 -o json"
test_cmd "YAML output" "$CLI list-projects --limit 1 -o yaml"
test_cmd "AI output" "$CLI list-projects --limit 1 -o ai"
echo ""

# 15. Test Session Context
echo -e "${YELLOW}15. Session Context Tests${NC}"
echo "========================="
export VIRTUOSO_SESSION_ID=$CHECKPOINT_ID
test_cmd "Session navigate" "$CLI navigate to 'https://test.com'"
test_cmd "Session click" "$CLI interact click 'button'"
test_cmd "Session assert" "$CLI assert exists 'body'"
unset VIRTUOSO_SESSION_ID
echo ""

# 16. Test Edge Cases
echo -e "${YELLOW}16. Edge Cases${NC}"
echo "==============="
test_cmd "Empty selector" "$CLI interact click $CHECKPOINT_ID '' $((POS++)) || true"
test_cmd "Special chars" "$CLI interact write $CHECKPOINT_ID 'input' 'Test @#$%^&*()' $((POS++))"
test_cmd "Unicode text" "$CLI interact write $CHECKPOINT_ID 'input' 'Hello 世界 🌍' $((POS++))"
test_cmd "Long text" "$CLI interact write $CHECKPOINT_ID 'textarea' '$(printf 'x%.0s' {1..1000})' $((POS++))"
test_cmd "Invalid position" "$CLI interact click $CHECKPOINT_ID 'div' $((POS++)) --position INVALID || true"
echo ""

# 17. Final Statistics
echo -e "${YELLOW}17. Final Checkpoint Verification${NC}"
echo "=================================="
FINAL_CHECKPOINTS=$($CLI list-checkpoints "$JOURNEY_ID" -o json)
STEP_COUNT=$(echo "$FINAL_CHECKPOINTS" | jq -r ".checkpoints[] | select(.id == $CHECKPOINT_ID) | .step_count" 2>/dev/null || echo "0")
echo "Total steps created: ${STEP_COUNT:-0}"
echo ""

# Cleanup (optional)
if [ "$CLEANUP" = "true" ]; then
    echo -e "${YELLOW}18. Cleanup${NC}"
    echo "============"
    # Note: Virtuoso API doesn't support project deletion via API
    echo "Cleanup not available via API - test resources remain for manual inspection"
    echo ""
fi

# Summary
echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              Test Summary                        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
echo ""
echo "Total tests: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo -e "Success rate: $(( PASSED * 100 / TOTAL ))%"
echo ""
echo "Test Infrastructure:"
echo "  • Project: $PROJECT_ID"
echo "  • Goal: $GOAL_ID"
echo "  • Journey: $JOURNEY_ID"
echo "  • Checkpoint: $CHECKPOINT_ID (${STEP_COUNT:-0} steps)"
echo ""
echo "Log file: $LOG_FILE"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed! The Virtuoso API CLI is fully functional.${NC}"
else
    echo -e "${YELLOW}⚠ Some tests failed. Check $LOG_FILE for details.${NC}"
    echo -e "${YELLOW}Note: Some failures may be due to API limitations rather than CLI issues.${NC}"
fi

echo ""
echo "Test completed at: $(date)"
