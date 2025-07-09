#!/bin/bash
# create-complete-test.sh - Example of creating a complete test structure

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Creating Complete Virtuoso Test Structure${NC}"
echo "========================================"

# Check if CLI exists
if [ ! -f "./bin/api-cli" ]; then
    echo "Error: CLI not found. Run 'make build' first."
    exit 1
fi

# Create project with timestamp to ensure uniqueness
echo -e "\n${GREEN}1. Creating Project...${NC}"
PROJECT_NAME="Demo E2E Test Suite $(date +%s)"
PROJECT_RESULT=$(./bin/api-cli create-project "$PROJECT_NAME" -o json)
PROJECT_ID=$(echo $PROJECT_RESULT | jq -r .project_id)
echo "Created project with ID: $PROJECT_ID"

# Create goal (auto-creates initial journey)
echo -e "\n${GREEN}2. Creating Goal...${NC}"
GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "User Authentication Tests" --url "https://demo.example.com" -o json)
GOAL_ID=$(echo $GOAL_RESULT | jq -r .goal_id)
SNAPSHOT_ID=$(echo $GOAL_RESULT | jq -r .snapshot_id)
echo "Created goal with ID: $GOAL_ID, Snapshot ID: $SNAPSHOT_ID"

# Create additional journey
echo -e "\n${GREEN}3. Creating Journey...${NC}"
JOURNEY_RESULT=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Login Happy Path" -o json)
JOURNEY_ID=$(echo $JOURNEY_RESULT | jq -r .journey_id)
echo "Created journey with ID: $JOURNEY_ID"

# Create checkpoints
echo -e "\n${GREEN}4. Creating Checkpoints...${NC}"
CP1_RESULT=$(./bin/api-cli create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID "Navigate to Login Page" -o json)
CP1=$(echo $CP1_RESULT | jq -r .checkpoint_id)
echo "Created checkpoint 1 with ID: $CP1"

CP2_RESULT=$(./bin/api-cli create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID "Enter Credentials" --position 3 -o json)
CP2=$(echo $CP2_RESULT | jq -r .checkpoint_id)
echo "Created checkpoint 2 with ID: $CP2"

CP3_RESULT=$(./bin/api-cli create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID "Verify Dashboard" --position 4 -o json)
CP3=$(echo $CP3_RESULT | jq -r .checkpoint_id)
echo "Created checkpoint 3 with ID: $CP3"

# Add steps to checkpoints
echo -e "\n${GREEN}5. Adding Steps...${NC}"

# Checkpoint 1: Navigate to login
echo "Adding steps to checkpoint 1..."
./bin/api-cli add-step navigate $CP1 --url "https://demo.example.com/login" -o human
./bin/api-cli add-step wait $CP1 --selector "Login Form" --timeout 5000 -o human

# Checkpoint 2: Login action
echo -e "\nAdding steps to checkpoint 2..."
./bin/api-cli add-step click $CP2 --selector "Email Field" -o human
./bin/api-cli add-step click $CP2 --selector "Password Field" -o human
./bin/api-cli add-step click $CP2 --selector "Login Button" -o human

# Checkpoint 3: Verify dashboard
echo -e "\nAdding steps to checkpoint 3..."
./bin/api-cli add-step wait $CP3 --selector "Dashboard" --timeout 10000 -o human
./bin/api-cli add-step wait $CP3 --selector "Welcome Message" --timeout 5000 -o human

echo -e "\n${BLUE}Test Structure Created Successfully!${NC}"
echo "===================================="
echo -e "${GREEN}Summary:${NC}"
echo "- Project ID: $PROJECT_ID"
echo "- Goal ID: $GOAL_ID"
echo "- Journey ID: $JOURNEY_ID"
echo "- Checkpoints: $CP1, $CP2, $CP3"
echo -e "\n${GREEN}Next Steps:${NC}"
echo "1. View in Virtuoso UI: https://app2.virtuoso.qa"
echo "2. Run the test journey"
echo "3. Add more checkpoints or steps as needed"
