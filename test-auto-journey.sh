#!/bin/bash
# Test to check if goals auto-create journeys

echo "🔍 Testing for Auto-Created Journeys"
echo "===================================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Step 1: Create a project
echo "1️⃣ Creating project..."
PROJECT_ID=$(./bin/api-cli create-project "Auto Journey Test $(date +%s)" -o json | jq -r .project_id)
echo "Project ID: $PROJECT_ID"

# Step 2: Create a goal
echo -e "\n2️⃣ Creating goal..."
GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal" --url "https://test.com" -o json)
echo "Goal Result: $GOAL_RESULT"
GOAL_ID=$(echo $GOAL_RESULT | jq -r .goal_id)
SNAPSHOT_ID=$(echo $GOAL_RESULT | jq -r .snapshot_id)

# Step 3: Immediately check for journeys
echo -e "\n3️⃣ Checking for journeys immediately after goal creation..."
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq .

# Step 4: Wait a moment and check again
echo -e "\n4️⃣ Waiting 2 seconds and checking again..."
sleep 2
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq .

# Step 5: Create a journey manually
echo -e "\n5️⃣ Creating a journey manually..."
JOURNEY_RESULT=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Manual Journey" -o json)
echo "Journey Result: $JOURNEY_RESULT"
JOURNEY_ID=$(echo $JOURNEY_RESULT | jq -r .journey_id)

# Step 6: Check all journeys now
echo -e "\n6️⃣ Listing all journeys after manual creation..."
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq .

echo -e "\n📊 Summary:"
echo "Project: $PROJECT_ID"
echo "Goal: $GOAL_ID"
echo "Manually created Journey: $JOURNEY_ID"