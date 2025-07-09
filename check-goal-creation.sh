#!/bin/bash
# check-goal-creation.sh - Detailed goal creation check

echo "🔬 Analyzing Goal Creation Response"
echo "==================================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Get API key for direct curl
API_KEY=$(grep api_key $API_CLI_CONFIG | awk '{print $2}')

# Create a project first
echo "Creating test project..."
PROJECT_ID=$(./bin/api-cli create-project "Goal Test $(date +%s)" -o json | jq -r .project_id)
echo "Project ID: $PROJECT_ID"
echo ""

# Now let's make a direct API call to create a goal and see the full response
echo "Making direct API call to create goal..."
echo "----------------------------------------"

RESPONSE=$(curl -s -X POST "https://api-app2.virtuoso.qa/api/goals" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{
        \"projectId\": $PROJECT_ID,
        \"goalName\": \"Direct API Goal\",
        \"applicationUrl\": \"https://example.com\",
        \"createFirstJourney\": true
    }")

echo "Full API Response:"
echo "$RESPONSE" | jq .
echo ""

# Extract any journey-related fields
echo "Looking for journey-related fields in response:"
echo "$RESPONSE" | jq 'keys[] | select(contains("journey") or contains("suite"))'
echo ""

# Get the goal ID
GOAL_ID=$(echo "$RESPONSE" | jq -r .id)
echo "Created Goal ID: $GOAL_ID"

# Try different endpoints to find journeys
echo ""
echo "Trying different endpoints to find journeys:"
echo "-------------------------------------------"

echo "1. /api/goals/$GOAL_ID/testsuites"
curl -s -X GET "https://api-app2.virtuoso.qa/api/goals/$GOAL_ID/testsuites" \
    -H "Authorization: Bearer $API_KEY" | jq .

echo ""
echo "2. /api/projects/$PROJECT_ID/goals/$GOAL_ID/testsuites"
curl -s -X GET "https://api-app2.virtuoso.qa/api/projects/$PROJECT_ID/goals/$GOAL_ID/testsuites" \
    -H "Authorization: Bearer $API_KEY" | jq .

echo ""
echo "3. Just /api/testsuites with goal filter"
curl -s -X GET "https://api-app2.virtuoso.qa/api/testsuites?goalId=$GOAL_ID" \
    -H "Authorization: Bearer $API_KEY" | jq .

echo ""
echo "Please check: Does the goal creation response contain any journey/testsuite ID?"
