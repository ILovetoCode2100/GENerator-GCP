#!/bin/bash

# Test script to create and analyze scroll steps

set -euo pipefail

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Creating test infrastructure...${NC}"

# Create project
PROJECT_ID=$(./bin/api-cli create-project "Scroll Test $(date +%s)" -o json | jq -r '.project_id')
echo "Created project: $PROJECT_ID"

# Create goal
GOAL_JSON=$(./bin/api-cli create-goal $PROJECT_ID "Scroll Test Goal" -o json)
GOAL_ID=$(echo "$GOAL_JSON" | jq -r '.goal_id')
SNAPSHOT_ID=$(echo "$GOAL_JSON" | jq -r '.snapshot_id')
echo "Created goal: $GOAL_ID"

# Create journey
JOURNEY_ID=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Scroll Test Journey" -o json | jq -r '.journey_id')
echo "Created journey: $JOURNEY_ID"

# Create checkpoint
CHECKPOINT_ID=$(./bin/api-cli create-checkpoint $JOURNEY_ID $GOAL_ID $SNAPSHOT_ID "Scroll Test Steps" -o json | jq -r '.checkpoint_id')
echo -e "${GREEN}Created checkpoint: $CHECKPOINT_ID${NC}"

# Set session for easier commands
export VIRTUOSO_SESSION_ID=$CHECKPOINT_ID

echo -e "\n${YELLOW}Creating scroll steps...${NC}"

# Create the same scroll steps that the user reported issues with
echo "1. Creating scroll-element step..."
./bin/api-cli step-navigate scroll-element "div.container" -o json | jq '.'

echo -e "\n2. Creating scroll-position step..."
./bin/api-cli step-navigate scroll-position 100,200 -o json | jq '.'

echo -e "\n3. Creating scroll-by step..."
./bin/api-cli step-navigate scroll-by 50,150 -o json | jq '.'

echo -e "\n4. Creating scroll-top step..."
./bin/api-cli step-navigate scroll-top -o json | jq '.'

echo -e "\n5. Creating scroll-bottom step..."
./bin/api-cli step-navigate scroll-bottom -o json | jq '.'

echo -e "\n${YELLOW}Retrieving steps from API...${NC}"

# Now analyze the steps using direct API call
CONFIG_FILE="$HOME/.api-cli/virtuoso-config.yaml"
AUTH_TOKEN=$(grep "auth_token:" "$CONFIG_FILE" | awk '{print $2}')
BASE_URL=$(grep "base_url:" "$CONFIG_FILE" | awk '{print $2}')

# Get steps
echo "Fetching steps for checkpoint $CHECKPOINT_ID..."
RESPONSE=$(curl -s -X GET \
  "${BASE_URL}/checkpoints/${CHECKPOINT_ID}/teststeps" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-Virtuoso-Client-ID: api-cli-generator" \
  -H "X-Virtuoso-Client-Name: api-cli-generator")

# Check response and display
if echo "$RESPONSE" | jq -e . >/dev/null 2>&1; then
    echo -e "\n${GREEN}Successfully retrieved steps:${NC}"
    echo "$RESPONSE" | jq '.items[] | {
        id: .id,
        action: .action,
        value: .value,
        meta: .meta,
        target: .target,
        description: .description
    }'

    # Analyze scroll steps specifically
    echo -e "\n${YELLOW}Scroll Steps Analysis:${NC}"
    echo "$RESPONSE" | jq '.items[] | select(.action == "SCROLL") | {
        id: .id,
        value: .value,
        meta: .meta,
        description: .description
    }'
else
    echo "Failed to retrieve steps. Raw response:"
    echo "$RESPONSE"
fi

echo -e "\n${GREEN}Test checkpoint ID: $CHECKPOINT_ID${NC}"
echo "You can manually check this in Virtuoso UI to compare"
