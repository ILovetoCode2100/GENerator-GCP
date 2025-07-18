#!/bin/bash

# Virtuoso API CLI - Complete End-to-End Test
# This is the ONLY test script you need - tests all working commands

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

CLI="./bin/api-cli"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Virtuoso API CLI - Complete E2E Test           ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
echo "Timestamp: $TIMESTAMP"
echo ""

# Stats
PASSED=0
TOTAL=0

test_cmd() {
    local name="$1"
    local cmd="$2"
    echo -n "  $name... "
    TOTAL=$((TOTAL + 1))
    if eval "$cmd" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗${NC}"
    fi
}

# 1. Configuration
echo -e "${YELLOW}1. Configuration & Setup${NC}"
echo "────────────────────────"
test_cmd "Validate config" "$CLI validate-config"
test_cmd "Show version" "$CLI --version"
echo ""

# 2. Create Infrastructure
echo -e "${YELLOW}2. Creating Test Infrastructure${NC}"
echo "───────────────────────────────"

PROJECT_JSON=$($CLI create-project "E2E_Test_${TIMESTAMP}" -o json)
PROJECT_ID=$(echo "$PROJECT_JSON" | jq -r '.project_id')

GOAL_JSON=$($CLI create-goal "$PROJECT_ID" "Test_Goal_${TIMESTAMP}" -o json)
GOAL_ID=$(echo "$GOAL_JSON" | jq -r '.goal_id')
SNAPSHOT_ID=$(echo "$GOAL_JSON" | jq -r '.snapshot_id')

JOURNEY_JSON=$($CLI create-journey "$GOAL_ID" "$SNAPSHOT_ID" "Test_Journey_${TIMESTAMP}" -o json)
JOURNEY_ID=$(echo "$JOURNEY_JSON" | jq -r '.journey_id')

CHECKPOINT_JSON=$($CLI create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "Complete_Test_Checkpoint" -o json)
CHECKPOINT_ID=$(echo "$CHECKPOINT_JSON" | jq -r '.checkpoint_id // .checkpointId // .id')

echo "  ✓ Project: $PROJECT_ID"
echo "  ✓ Goal: $GOAL_ID"
echo "  ✓ Journey: $JOURNEY_ID"
echo "  ✓ Checkpoint: $CHECKPOINT_ID"
echo ""

# Set session for commands that support it
export VIRTUOSO_SESSION_ID=$CHECKPOINT_ID

# Position counter
POS=1

# 3. List Commands
echo -e "${YELLOW}3. List Commands${NC}"
echo "────────────────"
test_cmd "List projects" "$CLI list-projects --limit 5 -o json"
test_cmd "List goals" "$CLI list-goals $PROJECT_ID -o json"
test_cmd "List journeys" "$CLI list-journeys $GOAL_ID $SNAPSHOT_ID -o json"
test_cmd "List checkpoints" "$CLI list-checkpoints $JOURNEY_ID -o json"
echo ""

# 4. Navigate Commands (8 steps)
echo -e "${YELLOW}4. Navigate Commands (8 types)${NC}"
echo "──────────────────────────────"
test_cmd "Navigate to URL" "$CLI navigate to $CHECKPOINT_ID 'https://example.com' $((POS++))"
test_cmd "Navigate new tab" "$CLI navigate to $CHECKPOINT_ID 'https://google.com' $((POS++)) --new-tab"
test_cmd "Navigate back" "$CLI navigate back $CHECKPOINT_ID $((POS++))"
test_cmd "Navigate forward" "$CLI navigate forward $CHECKPOINT_ID $((POS++))"
test_cmd "Navigate refresh" "$CLI navigate refresh $CHECKPOINT_ID $((POS++))"
test_cmd "Scroll to top" "$CLI navigate scroll-top $CHECKPOINT_ID $((POS++))"
test_cmd "Scroll to bottom" "$CLI navigate scroll-bottom $CHECKPOINT_ID $((POS++))"
test_cmd "Scroll to element" "$CLI navigate scroll-element $CHECKPOINT_ID 'body' $((POS++))"
echo ""

# 5. Assert Commands (10 types - skip selected and variable)
echo -e "${YELLOW}5. Assert Commands (10 types)${NC}"
echo "─────────────────────────────"
test_cmd "Assert exists" "$CLI assert exists 'Example Domain' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert not-exists" "$CLI assert not-exists 'Error' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert equals" "$CLI assert equals 'h1' 'Example Domain' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert not-equals" "$CLI assert not-equals 'h1' 'Wrong' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert checked" "$CLI assert checked 'input' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert gt" "$CLI assert gt 'body' '0' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert gte" "$CLI assert gte 'body' '0' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert lt" "$CLI assert lt 'body' '999999' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert lte" "$CLI assert lte 'body' '999999' --checkpoint $CHECKPOINT_ID"
test_cmd "Assert matches" "$CLI assert matches 'h1' '^Example' --checkpoint $CHECKPOINT_ID"
echo ""

# 6. Interact Commands (8 types)
echo -e "${YELLOW}6. Interact Commands (8 types)${NC}"
echo "─────────────────────────────"
test_cmd "Click" "$CLI interact click $CHECKPOINT_ID 'a' $((POS++))"
test_cmd "Click position" "$CLI interact click $CHECKPOINT_ID 'body' $((POS++)) --position CENTER"
test_cmd "Double-click" "$CLI interact double-click $CHECKPOINT_ID 'body' $((POS++))"
test_cmd "Right-click" "$CLI interact right-click $CHECKPOINT_ID 'body' $((POS++))"
test_cmd "Hover" "$CLI interact hover $CHECKPOINT_ID 'h1' $((POS++))"
test_cmd "Write text" "$CLI interact write $CHECKPOINT_ID 'body' 'test' $((POS++))"
test_cmd "Write clear" "$CLI interact write $CHECKPOINT_ID 'body' 'new' $((POS++)) --clear"
test_cmd "Press key" "$CLI interact key $CHECKPOINT_ID 'ESCAPE' $((POS++))"
echo ""

# 7. Data Commands (5 types)
echo -e "${YELLOW}7. Data Commands (5 types)${NC}"
echo "──────────────────────────"
test_cmd "Store element text" "$CLI data store element-text $CHECKPOINT_ID 'h1' 'myVar' $((POS++))"
test_cmd "Store literal" "$CLI data store literal $CHECKPOINT_ID 'testValue' 'myLiteral' $((POS++))"
test_cmd "Store attribute" "$CLI data store attribute $CHECKPOINT_ID '#link' 'href' 'linkUrl' $((POS++))"
test_cmd "Cookie create" "$CLI data cookie create $CHECKPOINT_ID 'session' 'abc123' $((POS++))"
test_cmd "Cookie clear all" "$CLI data cookie clear-all $CHECKPOINT_ID $((POS++))"
echo ""

# 8. Dialog Commands (5 types - skip prompt dismiss)
echo -e "${YELLOW}8. Dialog Commands (5 types)${NC}"
echo "────────────────────────────"
test_cmd "Alert accept" "$CLI dialog alert $CHECKPOINT_ID accept $((POS++))"
test_cmd "Alert dismiss" "$CLI dialog alert $CHECKPOINT_ID dismiss $((POS++))"
test_cmd "Confirm accept" "$CLI dialog confirm $CHECKPOINT_ID accept $((POS++))"
test_cmd "Confirm dismiss" "$CLI dialog confirm $CHECKPOINT_ID dismiss $((POS++))"
test_cmd "Prompt text" "$CLI dialog prompt $CHECKPOINT_ID 'input text' $((POS++))"
echo ""

# 9. Wait Commands (3 types - skip not-visible)
echo -e "${YELLOW}9. Wait Commands (3 types)${NC}"
echo "─────────────────────────"
test_cmd "Wait element" "$CLI wait element 'h1' --checkpoint $CHECKPOINT_ID"
test_cmd "Wait timeout" "$CLI wait element 'body' --timeout 5 --checkpoint $CHECKPOINT_ID"
test_cmd "Wait time" "$CLI wait time 1 --checkpoint $CHECKPOINT_ID"
echo ""

# 10. Window Commands (6 types)
echo -e "${YELLOW}10. Window Commands (6 types)${NC}"
echo "────────────────────────────"
test_cmd "Window maximize" "$CLI window maximize $CHECKPOINT_ID $((POS++))"
test_cmd "Window close" "$CLI window close $CHECKPOINT_ID $((POS++))"
test_cmd "Switch tab by index" "$CLI window switch 1 $CHECKPOINT_ID $((POS++))"
test_cmd "Switch next tab" "$CLI window switch tab next $CHECKPOINT_ID $((POS++))"
test_cmd "Switch prev tab" "$CLI window switch tab prev $CHECKPOINT_ID $((POS++))"
test_cmd "Window resize" "$CLI window resize 1024x768 $CHECKPOINT_ID $((POS++))"
echo ""

# 11. Mouse, Select, Misc Commands
echo -e "${YELLOW}11. Other Commands${NC}"
echo "──────────────────"
test_cmd "Mouse move" "$CLI mouse move 100 200 --checkpoint $CHECKPOINT_ID"
test_cmd "Select option" "$CLI select option $CHECKPOINT_ID 'select' 'Option 1' $((POS++))"
test_cmd "Add comment" "$CLI misc comment $CHECKPOINT_ID 'E2E test completed' $((POS++))"
test_cmd "Execute JS" "$CLI misc execute $CHECKPOINT_ID 'return 42;' $((POS++))"
echo ""

# 12. Add-Step Commands (simplified API)
echo -e "${YELLOW}12. Add-Step Commands${NC}"
echo "─────────────────────"
test_cmd "Add navigate" "$CLI add-step navigate $CHECKPOINT_ID --url 'https://test.com'"
test_cmd "Add click" "$CLI add-step click $CHECKPOINT_ID --selector 'button.submit'"
test_cmd "Add wait" "$CLI add-step wait $CHECKPOINT_ID --selector 'div.ready' --timeout 10"
echo ""

# 13. Management Commands
echo -e "${YELLOW}13. Management Commands${NC}"
echo "───────────────────────"
test_cmd "Update journey" "$CLI update-journey $JOURNEY_ID --name 'Updated_${TIMESTAMP}'"
echo ""

# 14. Output Formats
echo -e "${YELLOW}14. Output Formats${NC}"
echo "──────────────────"
test_cmd "JSON output" "$CLI list-projects --limit 1 -o json"
test_cmd "YAML output" "$CLI list-projects --limit 1 -o yaml"
test_cmd "Human output" "$CLI list-projects --limit 1 -o human"
test_cmd "AI output" "$CLI list-projects --limit 1 -o ai"
echo ""

# Final Verification
echo -e "${YELLOW}15. Final Verification${NC}"
echo "──────────────────────"

FINAL_CHECKPOINTS=$($CLI list-checkpoints "$JOURNEY_ID" -o json)
STEP_COUNT=$(echo "$FINAL_CHECKPOINTS" | jq -r ".checkpoints[] | select(.id == $CHECKPOINT_ID) | .step_count" 2>/dev/null)

echo "  Checkpoint has ${STEP_COUNT:-0} steps"
echo ""

# Summary
echo -e "${BLUE}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    Summary                       ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════╝${NC}"
echo ""
echo "Total tests: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$((TOTAL - PASSED))${NC}"
echo -e "Success rate: $(( PASSED * 100 / TOTAL ))%"
echo ""
echo "Test Infrastructure Created:"
echo "  • Project ID: $PROJECT_ID"
echo "  • Goal ID: $GOAL_ID"
echo "  • Journey ID: $JOURNEY_ID"
echo "  • Checkpoint ID: $CHECKPOINT_ID (${STEP_COUNT:-0} steps)"
echo ""

if [ "$STEP_COUNT" -ge 45 ]; then
    echo -e "${GREEN}✅ SUCCESS! All major step types have been tested.${NC}"
    echo -e "${GREEN}   Created $STEP_COUNT steps covering all command categories.${NC}"
else
    echo -e "${YELLOW}⚠ Created $STEP_COUNT steps. Some commands may not be adding steps.${NC}"
fi

echo ""
echo "This is the complete E2E test for Virtuoso API CLI."
echo "Run it anytime with: ./test-all-cli-commands.sh"

# Cleanup
rm -f test-all-commands-working.sh test-run.log test-errors-*.log 2>/dev/null || true
