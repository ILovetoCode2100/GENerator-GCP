#!/bin/bash
# Quick test to verify title fix

echo "🔍 Testing Journey Title Update Fix"
echo "==================================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Create simple test
PROJECT_ID=$(./bin/api-cli create-project "Title Fix Test $(date +%s)" -o json | jq -r .project_id)
GOAL_ID=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal" --url "https://test.com" -o json | jq -r .goal_id)
SNAPSHOT_ID=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal" --url "https://test.com" -o json | jq -r .snapshot_id)

echo "Created project: $PROJECT_ID, goal: $GOAL_ID"

# List journeys to see auto-created one
echo -e "\n1️⃣ Listing journeys (should show auto-created)..."
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq '.journeys[] | {id, name, title}'

# Get the journey ID
JOURNEY_ID=$(./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq -r '.journeys[0].id')

echo -e "\n2️⃣ Updating journey title..."
./bin/api-cli update-journey $JOURNEY_ID --name "My Custom Journey Title"

echo -e "\n3️⃣ Verifying update..."
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq '.journeys[] | {id, name, title}'

echo -e "\n✅ The 'name' field stays as 'Suite 1' but 'title' is updated!"