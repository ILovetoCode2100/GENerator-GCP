#!/bin/bash
# test-fixed-list-journeys.sh - Test the list-journeys fix

echo "🔧 Testing Fixed List Journeys"
echo "=============================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Test with known IDs that have journeys
echo "1. Testing with your known goal/snapshot that has journeys"
echo "----------------------------------------------------------"
GOAL_ID=13807
SNAPSHOT_ID=43830

echo "Using list-journeys command:"
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID

echo ""
echo "Expected to see:"
echo "- Journey 608093 (Suite 1) - The auto-created journey"
echo "- Journey 608094 (Suite 2) - The manually created journey"
echo ""

# Create a new test to verify auto-journey detection
echo "2. Creating new goal to test auto-journey detection"
echo "---------------------------------------------------"
PROJECT_ID=$(./bin/api-cli create-project "Fix Test $(date +%s)" -o json | jq -r .project_id)
echo "Created project: $PROJECT_ID"

GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "Test Auto Journey" --url "https://example.com" -o json)
GOAL_ID=$(echo $GOAL_RESULT | jq -r .goal_id)
SNAPSHOT_ID=$(echo $GOAL_RESULT | jq -r .snapshot_id)

echo "Created goal: $GOAL_ID with snapshot: $SNAPSHOT_ID"
echo ""

echo "Listing journeys (should show 1 auto-created journey):"
./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID

echo ""
echo "3. Testing batch structure with auto-journey handling"
echo "----------------------------------------------------"
cat > /tmp/test-auto-journey.yaml << EOF
project:
  id: $PROJECT_ID
  
goals:
  - name: "Test Goal with Auto Journey"
    url: "https://example.com"
    journeys:
      - name: "This Should Rename Auto Journey"
        checkpoints:
          - name: "Updated Navigation"
            navigation_url: "https://example.com/updated"
            steps:
              - type: wait
                selector: "body"
                timeout: 1000
EOF

echo "Running create-structure (should rename auto-journey):"
./bin/api-cli create-structure --file /tmp/test-auto-journey.yaml --dry-run

echo ""
echo "✅ Once ListJourneys is fixed, this test should:"
echo "   - Show the auto-created journey in listings"
echo "   - Allow batch structure to rename it"
echo "   - Update the navigation URL properly"
