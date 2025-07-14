#!/bin/bash

# Test script that creates a new project, goal, journey, checkpoint
# and then tests all CLI commands by creating steps

CLI="./bin/api-cli"
RESULTS_FILE="test-results-all-commands.txt"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
PROJECT_NAME="Test_Project_${TIMESTAMP}"
GOAL_NAME="Test_Goal_${TIMESTAMP}"
JOURNEY_NAME="Test_Journey_${TIMESTAMP}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize results file
echo "=== COMPREHENSIVE CLI COMMANDS TEST ===" > $RESULTS_FILE
echo "Test Date: $(date)" >> $RESULTS_FILE
echo "Project Name: $PROJECT_NAME" >> $RESULTS_FILE
echo "" >> $RESULTS_FILE

# Function to test a command and log results
test_command() {
    local group=$1
    local command=$2
    local args=$3

    echo -e "\n${YELLOW}Testing: $group - $command${NC}"
    echo -e "\n--- Testing: $group - $command ---" >> $RESULTS_FILE
    echo "Command: $CLI $command $args" >> $RESULTS_FILE

    # Run the command
    output=$($CLI $command $args 2>&1)
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✓ Success${NC}"
        echo "Status: SUCCESS" >> $RESULTS_FILE
        echo "Output: $output" >> $RESULTS_FILE
        return 0
    else
        echo -e "${RED}✗ Failed${NC}"
        echo "Status: FAILED" >> $RESULTS_FILE
        echo "Exit Code: $exit_code" >> $RESULTS_FILE
        echo "Error: $output" >> $RESULTS_FILE
        return 1
    fi
}

echo -e "${BLUE}=== STEP 1: Creating Test Infrastructure ===${NC}"

# Create Project
echo -e "\n${YELLOW}Creating new project: $PROJECT_NAME${NC}"
project_output=$($CLI create-project "$PROJECT_NAME" --output json 2>&1)
if [ $? -eq 0 ]; then
    PROJECT_ID=$(echo "$project_output" | grep -o '"project_id":[^,]*' | sed 's/"project_id"://' | tr -d ' ')
    echo -e "${GREEN}✓ Project created with ID: $PROJECT_ID${NC}"
    echo "Project ID: $PROJECT_ID" >> $RESULTS_FILE
else
    echo -e "${RED}✗ Failed to create project${NC}"
    echo "Failed to create project: $project_output" >> $RESULTS_FILE
    exit 1
fi

# Create Goal
echo -e "\n${YELLOW}Creating goal: $GOAL_NAME${NC}"
goal_output=$($CLI create-goal "$PROJECT_ID" "$GOAL_NAME" --output json 2>&1)
if [ $? -eq 0 ]; then
    GOAL_ID=$(echo "$goal_output" | grep -o '"goal_id":[^,]*' | sed 's/"goal_id"://' | tr -d ' ')
    SNAPSHOT_ID=$(echo "$goal_output" | sed -n 's/.*"snapshot_id": *"\([^"]*\)".*/\1/p')
    echo -e "${GREEN}✓ Goal created with ID: $GOAL_ID (Snapshot: $SNAPSHOT_ID)${NC}"
    echo "Goal ID: $GOAL_ID" >> $RESULTS_FILE
    echo "Snapshot ID: $SNAPSHOT_ID" >> $RESULTS_FILE
else
    echo -e "${RED}✗ Failed to create goal${NC}"
    echo "Failed to create goal: $goal_output" >> $RESULTS_FILE
    exit 1
fi

# Create Journey
echo -e "\n${YELLOW}Creating journey: $JOURNEY_NAME${NC}"
journey_output=$($CLI create-journey "$GOAL_ID" "$SNAPSHOT_ID" "$JOURNEY_NAME" --output json 2>&1)
if [ $? -eq 0 ]; then
    JOURNEY_ID=$(echo "$journey_output" | grep -o '"journey_id":[^,]*' | sed 's/"journey_id"://' | tr -d ' ')
    echo -e "${GREEN}✓ Journey created with ID: $JOURNEY_ID${NC}"
    echo "Journey ID: $JOURNEY_ID" >> $RESULTS_FILE
else
    echo -e "${RED}✗ Failed to create journey${NC}"
    echo "Failed to create journey: $journey_output" >> $RESULTS_FILE
    exit 1
fi

# Create Checkpoint
echo -e "\n${YELLOW}Creating checkpoint${NC}"
checkpoint_output=$($CLI create-checkpoint "$JOURNEY_ID" "$GOAL_ID" "$SNAPSHOT_ID" "Test_Checkpoint_${TIMESTAMP}" --output json 2>&1)
if [ $? -eq 0 ]; then
    CHECKPOINT_ID=$(echo "$checkpoint_output" | grep -o '"checkpoint_id":[^,]*' | sed 's/"checkpoint_id"://' | tr -d ' ')
    echo -e "${GREEN}✓ Checkpoint created with ID: $CHECKPOINT_ID${NC}"
    echo "Checkpoint ID: $CHECKPOINT_ID" >> $RESULTS_FILE
else
    echo -e "${RED}✗ Failed to create checkpoint${NC}"
    echo "Failed to create checkpoint: $checkpoint_output" >> $RESULTS_FILE
    exit 1
fi

echo -e "\n${BLUE}=== STEP 2: Setting Checkpoint Context ===${NC}"
$CLI set-checkpoint $CHECKPOINT_ID

echo -e "\n${BLUE}=== STEP 3: Testing All CLI Commands ===${NC}"

# Position counter for steps
POSITION=1

# 1. Assert commands (12 subcommands)
echo -e "\n${BLUE}=== ASSERT COMMANDS (12 variations) ===${NC}"
test_command "assert" "assert" "exists '#login-button' $POSITION" && ((POSITION++))
test_command "assert" "assert" "not-exists '#old-element' $POSITION" && ((POSITION++))
test_command "assert" "assert" "equals '#username' 'test@example.com' $POSITION" && ((POSITION++))
test_command "assert" "assert" "not-equals '#status' 'offline' $POSITION" && ((POSITION++))
test_command "assert" "assert" "checked '#remember-me' $POSITION" && ((POSITION++))
test_command "assert" "assert" "selected '#country-usa' $POSITION" && ((POSITION++))
test_command "assert" "assert" "variable 'userRole' 'admin' $POSITION" && ((POSITION++))
test_command "assert" "assert" "gt '#price' '100' $POSITION" && ((POSITION++))
test_command "assert" "assert" "gte '#quantity' '10' $POSITION" && ((POSITION++))
test_command "assert" "assert" "lt '#discount' '50' $POSITION" && ((POSITION++))
test_command "assert" "assert" "lte '#tax' '15' $POSITION" && ((POSITION++))
test_command "assert" "assert" "matches '#email' '^[a-zA-Z0-9+_.-]+@[a-zA-Z0-9.-]+$' $POSITION" && ((POSITION++))

# 2. Interact commands (6 subcommands)
echo -e "\n${BLUE}=== INTERACT COMMANDS (6 variations) ===${NC}"
test_command "interact" "interact" "click '#submit-button' $POSITION" && ((POSITION++))
test_command "interact" "interact" "double-click '#header-logo' $POSITION" && ((POSITION++))
test_command "interact" "interact" "right-click '#context-menu' $POSITION" && ((POSITION++))
test_command "interact" "interact" "hover '#dropdown-menu' $POSITION" && ((POSITION++))
test_command "interact" "interact" "write '#search-input' 'test search query' $POSITION" && ((POSITION++))
test_command "interact" "interact" "key 'Enter' $POSITION --target '#search-input'" && ((POSITION++))

# 3. Navigate commands (5 subcommands)
echo -e "\n${BLUE}=== NAVIGATE COMMANDS (5 variations) ===${NC}"
test_command "navigate" "navigate" "to https://example.com $POSITION" && ((POSITION++))
test_command "navigate" "navigate" "to https://example.com/page2 $POSITION --new-tab" && ((POSITION++))
test_command "navigate" "navigate" "scroll-to '#footer' $POSITION" && ((POSITION++))
test_command "navigate" "navigate" "scroll-top $POSITION" && ((POSITION++))
test_command "navigate" "navigate" "scroll-bottom $POSITION" && ((POSITION++))

# 4. Data commands (5 subcommands)
echo -e "\n${BLUE}=== DATA COMMANDS (5 variations) ===${NC}"
test_command "data" "data" "store-text '#product-name' 'productVar' $POSITION" && ((POSITION++))
test_command "data" "data" "store-value 'TestValue123' 'testVar' $POSITION" && ((POSITION++))
test_command "data" "data" "cookie-create 'sessionId' 'abc123xyz' $POSITION" && ((POSITION++))
test_command "data" "data" "cookie-delete 'sessionId' $POSITION" && ((POSITION++))
test_command "data" "data" "cookie-clear $POSITION" && ((POSITION++))

# 5. Dialog commands (4 subcommands - note: dismiss-popup removed as it's not in the code)
echo -e "\n${BLUE}=== DIALOG COMMANDS (3 variations) ===${NC}"
test_command "dialog" "dialog" "dismiss-alert $POSITION" && ((POSITION++))
test_command "dialog" "dialog" "dismiss-confirm $POSITION" && ((POSITION++))
test_command "dialog" "dialog" "dismiss-prompt 'User Input' $POSITION" && ((POSITION++))

# 6. Wait commands (2 subcommands)
echo -e "\n${BLUE}=== WAIT COMMANDS (2 variations) ===${NC}"
test_command "wait" "wait" "element '#loading-spinner' $POSITION" && ((POSITION++))
test_command "wait" "wait" "element '#dynamic-content' $POSITION --timeout 10000" && ((POSITION++))
test_command "wait" "wait" "time 2000 $POSITION" && ((POSITION++))

# 7. Window commands (5 subcommands - resize uses different formats)
echo -e "\n${BLUE}=== WINDOW COMMANDS (5 variations) ===${NC}"
test_command "window" "window" "resize 1920x1080 $POSITION" && ((POSITION++))
test_command "window" "window" "resize 1024x768 $POSITION" && ((POSITION++))
test_command "window" "window" "switch-tab next $POSITION" && ((POSITION++))
test_command "window" "window" "switch-tab prev $POSITION" && ((POSITION++))
test_command "window" "window" "switch-frame '#iframe-content' $POSITION" && ((POSITION++))

# 8. Mouse commands (6 subcommands)
echo -e "\n${BLUE}=== MOUSE COMMANDS (6 variations) ===${NC}"
test_command "mouse" "mouse" "move-to 500 300 $POSITION" && ((POSITION++))
test_command "mouse" "mouse" "move-by 100 -50 $POSITION" && ((POSITION++))
test_command "mouse" "mouse" "move 'smooth' $POSITION" && ((POSITION++))
test_command "mouse" "mouse" "down $POSITION" && ((POSITION++))
test_command "mouse" "mouse" "up $POSITION" && ((POSITION++))
test_command "mouse" "mouse" "enter $POSITION" && ((POSITION++))

# 9. Select commands (3 subcommands)
echo -e "\n${BLUE}=== SELECT COMMANDS (3 variations) ===${NC}"
test_command "select" "select" "option '#country-dropdown' 'United States' $POSITION" && ((POSITION++))
test_command "select" "select" "index '#state-dropdown' 5 $POSITION" && ((POSITION++))
test_command "select" "select" "last '#year-dropdown' $POSITION" && ((POSITION++))

# 10. File commands (1 subcommand)
echo -e "\n${BLUE}=== FILE COMMANDS (1 variation) ===${NC}"
test_command "file" "file" "upload '#file-upload' 'https://example.com/sample.pdf' $POSITION" && ((POSITION++))

# 11. Misc commands (3 subcommands)
echo -e "\n${BLUE}=== MISC COMMANDS (3 variations) ===${NC}"
test_command "misc" "misc" "comment 'Starting checkout process' $POSITION" && ((POSITION++))
test_command "misc" "misc" "execute-script 'return document.title' $POSITION" && ((POSITION++))
test_command "misc" "misc" "key 'Escape' $POSITION" && ((POSITION++))

# Additional command variations with options
echo -e "\n${BLUE}=== COMMANDS WITH SPECIAL OPTIONS ===${NC}"
test_command "interact" "interact" "click '#save-button' $POSITION --position TOP_RIGHT" && ((POSITION++))
test_command "interact" "interact" "click '#menu-item' $POSITION --element-type LINK" && ((POSITION++))
test_command "interact" "interact" "write '#password' 'secret123' $POSITION --variable passwordVar" && ((POSITION++))
test_command "navigate" "navigate" "scroll-element '#sidebar' 200 $POSITION" && ((POSITION++))
test_command "window" "window" "switch-frame parent $POSITION" && ((POSITION++))

# Summary
echo -e "\n${BLUE}=== TEST SUMMARY ===${NC}"
echo -e "\n=== TEST SUMMARY ===" >> $RESULTS_FILE

# Count successes and failures
success_count=$(grep -c "Status: SUCCESS" $RESULTS_FILE)
fail_count=$(grep -c "Status: FAILED" $RESULTS_FILE)
total_count=$((success_count + fail_count))

echo "Total Commands Tested: $total_count" | tee -a $RESULTS_FILE
echo "Successful: $success_count" | tee -a $RESULTS_FILE
echo "Failed: $fail_count" | tee -a $RESULTS_FILE
if [ $total_count -gt 0 ]; then
    success_rate=$(( success_count * 100 / total_count ))
    echo "Success Rate: ${success_rate}%" | tee -a $RESULTS_FILE
fi

echo -e "\n${YELLOW}Test Infrastructure Created:${NC}"
echo "Project: $PROJECT_NAME (ID: $PROJECT_ID)" | tee -a $RESULTS_FILE
echo "Goal: $GOAL_NAME (ID: $GOAL_ID)" | tee -a $RESULTS_FILE
echo "Journey: $JOURNEY_NAME (ID: $JOURNEY_ID)" | tee -a $RESULTS_FILE
echo "Checkpoint ID: $CHECKPOINT_ID" | tee -a $RESULTS_FILE
echo "Total Steps Created: $((POSITION - 1))" | tee -a $RESULTS_FILE

echo -e "\n${YELLOW}Full test results saved to: $RESULTS_FILE${NC}"

# Show failed commands if any
if [ $fail_count -gt 0 ]; then
    echo -e "\n${RED}=== FAILED COMMANDS ===${NC}"
    echo -e "\n=== FAILED COMMANDS ===" >> $RESULTS_FILE
    grep -B1 -A3 "Status: FAILED" $RESULTS_FILE | tail -20
fi

# Show command group summary
echo -e "\n${BLUE}=== COMMAND GROUP SUMMARY ===${NC}"
echo -e "\n=== COMMAND GROUP SUMMARY ===" >> $RESULTS_FILE

for group in "assert" "interact" "navigate" "data" "dialog" "wait" "window" "mouse" "select" "file" "misc"; do
    group_total=$(grep -c "Testing: $group - " $RESULTS_FILE || echo "0")
    group_success=$(grep -A2 "Testing: $group - " $RESULTS_FILE | grep -c "Status: SUCCESS" || echo "0")
    echo "$group commands: $group_success/$group_total passed" | tee -a $RESULTS_FILE
done

# Exit with appropriate code
if [ $fail_count -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed successfully!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed. Check the results above.${NC}"
    exit 1
fi
