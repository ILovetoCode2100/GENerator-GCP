#!/bin/bash
# Simple test to check goal creation and auto-journey

echo "🔍 Testing Goal Creation with Auto-Journey"
echo "========================================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Create project
echo "1️⃣ Creating project..."
PROJECT_ID=$(./bin/api-cli create-project "Auto Journey Test $(date +%s)" -o json | jq -r .project_id)
echo "Project ID: $PROJECT_ID"

# Create goal and capture everything
echo -e "\n2️⃣ Creating goal (should auto-create journey)..."
GOAL_JSON=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal With Auto Journey" --url "https://test.com" -o json)
echo "Full goal response:"
echo "$GOAL_JSON" | jq .

GOAL_ID=$(echo "$GOAL_JSON" | jq -r .goal_id)
SNAPSHOT_ID=$(echo "$GOAL_JSON" | jq -r .snapshot_id)

echo -e "\nExtracted:"
echo "- Goal ID: $GOAL_ID"
echo "- Snapshot ID: $SNAPSHOT_ID"

# Wait a moment for async operations
echo -e "\n3️⃣ Waiting 3 seconds for any async journey creation..."
sleep 3

# Check journeys
echo -e "\n4️⃣ Checking for journeys..."
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID

echo -e "\n5️⃣ Creating a manual journey to compare..."
MANUAL_JOURNEY=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Manual Journey" -o json)
echo "Manual journey created:"
echo "$MANUAL_JOURNEY" | jq .

echo -e "\n📊 Summary:"
echo "- Project: $PROJECT_ID"
echo "- Goal: $GOAL_ID"
echo "- Check Virtuoso UI for any additional journeys"
echo "- The 'createFirstJourney: true' flag is set in our API call"
echo "- But we're not seeing the auto-created journey in our API responses"