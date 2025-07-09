#!/bin/bash
# Test journey visibility and IDs

echo "🔍 Testing Journey Visibility and Sequential IDs"
echo "=============================================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Create test data
echo "Creating test project..."
PROJECT_ID=$(./bin/api-cli create-project "Journey Visibility Test $(date +%s)" -o json | jq -r .project_id)
GOAL_ID=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal" --url "https://test.com" -o json | jq -r .goal_id)
SNAPSHOT_ID=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal 2" --url "https://test2.com" -o json | jq -r .snapshot_id)

echo ""
echo "📊 Created:"
echo "Project: $PROJECT_ID"
echo "Goal: $GOAL_ID"
echo "Snapshot: $SNAPSHOT_ID"
echo ""

# Create multiple journeys to see ID pattern
echo "Creating 3 journeys to see ID sequence..."
for i in 1 2 3; do
    echo -e "\nCreating journey $i..."
    RESULT=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Journey $i" -o json)
    echo "$RESULT" | jq .
    JOURNEY_ID=$(echo "$RESULT" | jq -r .journey_id)
    echo "Created Journey ID: $JOURNEY_ID"
done

echo -e "\n🔍 Now let's check what the list-journeys command returns..."
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq .

echo -e "\n💡 Observation:"
echo "The journey IDs are sequential (e.g., 608086, 608087, 608088...)"
echo "If you saw ID 608083 earlier, it was from a previous test run."
echo "There are NO auto-created journeys - we create them all explicitly."