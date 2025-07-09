#!/bin/bash
# step-by-step-test.sh - Manual step-through test

echo "🔍 Step-by-Step Journey Investigation"
echo "====================================="
echo "We'll pause after each step so you can verify what's happening"
echo ""

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Step 1: Create a fresh project
echo "STEP 1: Creating a new project"
echo "------------------------------"
read -p "Press Enter to create project..."
PROJECT_ID=$(./bin/api-cli create-project "Manual Test $(date +%H%M%S)" -o json | jq -r .project_id)
echo "✅ Created project: $PROJECT_ID"
echo ""

# Step 2: Create a goal
echo "STEP 2: Creating a goal"
echo "-----------------------"
echo "Watch for any auto-created journey in the response..."
read -p "Press Enter to create goal..."
./bin/api-cli create-goal $PROJECT_ID "Test Goal" --url "https://example.com"
echo ""
echo "Now let's check the JSON response:"
GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal 2" --url "https://example.com" -o json)
echo "Full response:"
echo "$GOAL_RESULT" | jq .
GOAL_ID=$(echo $GOAL_RESULT | jq -r .goal_id)
SNAPSHOT_ID=$(echo $GOAL_RESULT | jq -r .snapshot_id)
echo ""
echo "Extracted:"
echo "- Goal ID: $GOAL_ID"
echo "- Snapshot ID: $SNAPSHOT_ID"
echo ""

# Step 3: Check for journeys
echo "STEP 3: Looking for journeys"
echo "----------------------------"
read -p "Press Enter to list journeys..."
echo "Using list-journeys command:"
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID
echo ""
echo "JSON format:"
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq .
echo ""

# Step 4: Manually create a journey
echo "STEP 4: Creating a journey manually"
echo "-----------------------------------"
read -p "Press Enter to create journey..."
JOURNEY_RESULT=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Manual Journey" -o json)
echo "Journey creation result:"
echo "$JOURNEY_RESULT" | jq .
JOURNEY_ID=$(echo $JOURNEY_RESULT | jq -r .journey_id)
echo "Created Journey ID: $JOURNEY_ID"
echo ""

# Step 5: List journeys again
echo "STEP 5: List journeys after creation"
echo "------------------------------------"
read -p "Press Enter to list journeys again..."
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq .
echo ""

# Step 6: Check the Virtuoso UI
echo "STEP 6: Manual verification"
echo "---------------------------"
echo "Please check the Virtuoso UI:"
echo "1. Navigate to project $PROJECT_ID"
echo "2. Look at goal $GOAL_ID"
echo "3. Check if there are any journeys we're missing"
echo "4. Note any journey IDs you see"
echo ""
read -p "What journey IDs do you see in the UI? Enter them here: " UI_JOURNEY_IDS
echo ""

# Step 7: Try direct API calls
echo "STEP 7: Direct API investigation"
echo "--------------------------------"
echo "Let's try some direct API calls to understand what's happening..."
read -p "Press Enter to continue..."

# Try to get journey details if user provided an ID
if [ ! -z "$UI_JOURNEY_IDS" ]; then
    for JID in $UI_JOURNEY_IDS; do
        echo "Trying to get details for journey $JID:"
        curl -s -X GET "https://api-app2.virtuoso.qa/api/testsuites/$JID" \
            -H "Authorization: Bearer $(grep api_key $API_CLI_CONFIG | awk '{print $2}')" \
            -H "Content-Type: application/json" | jq .
        echo ""
    done
fi

echo "INVESTIGATION SUMMARY"
echo "===================="
echo "Project ID: $PROJECT_ID"
echo "Goal ID: $GOAL_ID"
echo "Created Journey ID: $JOURNEY_ID"
echo "UI Journey IDs: $UI_JOURNEY_IDS"
echo ""
echo "Key Questions:"
echo "1. Did the goal creation show any journey in its response?"
echo "2. Are there journeys in the UI that our API doesn't show?"
echo "3. Is there a journey with an ID before $JOURNEY_ID that we should have found?"
