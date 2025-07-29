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

# Module definitions - using arrays to preserve order
modules=(
    "Sales:Sales Module Tests"
    "Customer_Service:Customer Service Tests"
    "Field_Service:Field Service Tests"
    "Marketing:Marketing Tests"
    "Finance_Operations:Finance and Operations Tests"
    "Project_Operations:Project Operations Tests"
    "Human_Resources:Human Resources Tests"
    "Supply_Chain:Supply Chain Management Tests"
    "Commerce:Commerce Tests"
)

# Clear previous mappings
> goal-mappings.txt

# Create goals for each module
echo -e "${BLUE}Creating goals for each D365 module...${NC}"

for module_info in "${modules[@]}"; do
    IFS=':' read -r module_key module_name <<< "$module_info"

    # Replace underscores with spaces for display
    display_name="${module_key//_/ }"

    echo -n -e "Creating goal: $display_name... "

    # Create goal (API doesn't support description flag, will use name only)
    if goal_output=$($API_CLI create-goal "$PROJECT_ID" "$display_name" --output json 2>&1); then
        goal_id=$(echo "$goal_output" | jq -r '.goal_id' 2>/dev/null || echo "")
        if [ -n "$goal_id" ]; then
            echo -e "${GREEN}✓${NC} (Goal ID: $goal_id)"
            echo "$display_name:$goal_id:$module_name" >> goal-mappings.txt
        else
            echo -e "${YELLOW}⚠${NC} Created but couldn't get ID"
            echo "$goal_output" >> create-goals-debug.log
        fi
    else
        echo -e "${RED}✗${NC}"
        echo "Error: $goal_output" | tee -a create-goals-debug.log
    fi
done

echo ""

# Display created goals
if [ -f goal-mappings.txt ] && [ -s goal-mappings.txt ]; then
    echo -e "${GREEN}Successfully created goals:${NC}"
    cat goal-mappings.txt
else
    echo -e "${YELLOW}No goals were created successfully${NC}"
fi

echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Tests created with run-test are in the project but not organized by goals"
echo "2. Future tests should be created under the appropriate goal"
echo "3. Use: ./bin/api-cli create-journey <goal-id> <journey-name>"
