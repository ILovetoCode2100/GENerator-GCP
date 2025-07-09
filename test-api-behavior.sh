#!/bin/bash
# Test actual Virtuoso API behavior

echo "🔍 Testing Virtuoso API Behavior"
echo "================================"

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Create a test project
echo -e "\n1️⃣ Creating test project..."
PROJECT_ID=$(./bin/api-cli create-project "API Test $(date +%Y%m%d-%H%M%S)" -o json | jq -r .project_id)
echo "Project ID: $PROJECT_ID"

# Create a goal
echo -e "\n2️⃣ Creating goal..."
GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal" --url "https://test.com" -o json)
echo "Goal Result: $GOAL_RESULT"
GOAL_ID=$(echo $GOAL_RESULT | jq -r .goal_id)
SNAPSHOT_ID=$(echo $GOAL_RESULT | jq -r .snapshot_id)

echo "Goal ID: $GOAL_ID"
echo "Snapshot ID: $SNAPSHOT_ID"

# Check if journeys exist
echo -e "\n3️⃣ Checking for auto-created journeys..."
JOURNEYS=$(./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json)
echo "Journeys: $JOURNEYS"
JOURNEY_COUNT=$(echo $JOURNEYS | jq -r .count)
echo "Journey count: $JOURNEY_COUNT"

if [ "$JOURNEY_COUNT" -eq "0" ]; then
    echo -e "\n❌ No auto-created journey found! Creating one manually..."
    JOURNEY_RESULT=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Manual Journey" -o json)
    echo "Journey Result: $JOURNEY_RESULT"
    JOURNEY_ID=$(echo $JOURNEY_RESULT | jq -r .journey_id)
else
    echo -e "\n✅ Found auto-created journey!"
    JOURNEY_ID=$(echo $JOURNEYS | jq -r '.journeys[0].id')
fi

echo "Journey ID: $JOURNEY_ID"

# Check checkpoints
echo -e "\n4️⃣ Checking for checkpoints..."
./bin/api-cli list-checkpoints $JOURNEY_ID

echo -e "\n✨ Test Complete!"
echo "Project: $PROJECT_ID"
echo "Goal: $GOAL_ID" 
echo "Journey: $JOURNEY_ID"