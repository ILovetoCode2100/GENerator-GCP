#!/bin/bash

# Create proper goal structure for D365 tests
# This script creates goals for each module in project 9369

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ID=9369
API_CLI="./bin/api-cli"

echo -e "${BLUE}=== Creating D365 Goal Structure ===${NC}"
echo -e "Project ID: $PROJECT_ID"
echo ""

# Module definitions
declare -A modules=(
    ["Sales"]="Sales Module Tests"
    ["Customer Service"]="Customer Service Tests"
    ["Field Service"]="Field Service Tests"
    ["Marketing"]="Marketing Tests"
    ["Finance Operations"]="Finance and Operations Tests"
    ["Project Operations"]="Project Operations Tests"
    ["Human Resources"]="Human Resources Tests"
    ["Supply Chain"]="Supply Chain Management Tests"
    ["Commerce"]="Commerce Tests"
)

# Create goals for each module
echo -e "${BLUE}Creating goals for each D365 module...${NC}"

for module in "${!modules[@]}"; do
    description="${modules[$module]}"
    echo -n -e "Creating goal: $module... "

    # Create goal
    if goal_output=$($API_CLI create-goal "$PROJECT_ID" "$module" --description "$description" --output json 2>&1); then
        goal_id=$(echo "$goal_output" | jq -r '.goal.id' 2>/dev/null || echo "")
        if [ -n "$goal_id" ]; then
            echo -e "${GREEN}✓${NC} (Goal ID: $goal_id)"
            echo "$module:$goal_id" >> goal-mappings.txt
        else
            echo -e "${YELLOW}⚠${NC} Created but couldn't get ID"
        fi
    else
        echo -e "${RED}✗${NC}"
        echo "Error: $goal_output"
    fi
done

echo ""
echo -e "${GREEN}Goal structure created!${NC}"
echo "Goal mappings saved to: goal-mappings.txt"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Move tests to their respective goals"
echo "2. Or create new tests under the proper goal structure"
