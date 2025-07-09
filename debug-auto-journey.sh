#!/bin/bash
# debug-auto-journey.sh - Find the auto-created journey

echo "🔍 Debugging Auto-Created Journey Detection"
echo "=========================================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Get API key for direct calls
API_KEY=$(grep api_key $API_CLI_CONFIG | awk '{print $2}')

# Test with the known auto-created journey
echo "1. Testing with known auto-created journey ID 608093"
echo "----------------------------------------------------"
echo "Direct API call to get journey details:"
curl -s -X GET "https://api-app2.virtuoso.qa/api/testsuites/608093" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" | jq .

echo ""
echo "2. Let's create a new goal and watch carefully"
echo "----------------------------------------------"

# Create project
PROJECT_ID=$(./bin/api-cli create-project "Debug Test $(date +%s)" -o json | jq -r .project_id)
echo "Created project: $PROJECT_ID"

# Create goal with verbose curl to see response
echo ""
echo "Creating goal with full response capture:"
GOAL_RESPONSE=$(curl -s -X POST "https://api-app2.virtuoso.qa/api/goals" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{
        \"projectId\": $PROJECT_ID,
        \"goalName\": \"Test Goal with Journey\",
        \"applicationUrl\": \"https://example.com\",
        \"createFirstJourney\": true
    }")

echo "Goal creation response:"
echo "$GOAL_RESPONSE" | jq .

GOAL_ID=$(echo "$GOAL_RESPONSE" | jq -r .id)
SNAPSHOT_ID=$(echo "$GOAL_RESPONSE" | jq -r .latestSnapshotId)

echo ""
echo "Extracted:"
echo "- Goal ID: $GOAL_ID"
echo "- Snapshot ID: $SNAPSHOT_ID"

# Now let's try different ways to find the journey
echo ""
echo "3. Trying different endpoints to find the auto-created journey"
echo "--------------------------------------------------------------"

echo ""
echo "a) Using our CLI list-journeys:"
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq .

echo ""
echo "b) Direct API - testsuites with goalId:"
curl -s -X GET "https://api-app2.virtuoso.qa/api/testsuites?goalId=$GOAL_ID" \
    -H "Authorization: Bearer $API_KEY" | jq .

echo ""
echo "c) Direct API - testsuites/latest_status:"
curl -s -X GET "https://api-app2.virtuoso.qa/api/testsuites/latest_status?goalIds=$GOAL_ID&includeDataDriven=true&includePlans=true&snapshotId=$SNAPSHOT_ID" \
    -H "Authorization: Bearer $API_KEY" | jq .

echo ""
echo "d) Check if goal has a firstJourneyId field:"
curl -s -X GET "https://api-app2.virtuoso.qa/api/goals/$GOAL_ID" \
    -H "Authorization: Bearer $API_KEY" | jq .

echo ""
echo "4. Let's check what happens when we DON'T use createFirstJourney"
echo "----------------------------------------------------------------"
GOAL_RESPONSE2=$(curl -s -X POST "https://api-app2.virtuoso.qa/api/goals" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{
        \"projectId\": $PROJECT_ID,
        \"goalName\": \"Test Goal WITHOUT Journey\",
        \"applicationUrl\": \"https://example.com\",
        \"createFirstJourney\": false
    }")

echo "Goal creation without createFirstJourney:"
echo "$GOAL_RESPONSE2" | jq .

echo ""
echo "🤔 ANALYSIS:"
echo "- You see journey 608093 in the UI"
echo "- Our API calls don't show it"
echo "- This suggests the journey might be created asynchronously"
echo "- Or there's a different endpoint we need to use"
