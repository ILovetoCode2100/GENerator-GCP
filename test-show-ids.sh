#!/bin/bash
# Test that shows all IDs clearly

echo "🔍 Test Showing Project, Goal, and Journey IDs"
echo "=============================================="

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Create a simple test structure
cat > /tmp/show-ids-test.yaml << EOF
project:
  name: "ID Demo Project $(date +%Y%m%d-%H%M%S)"
  
goals:
  - name: "ID Demo Goal"
    url: "https://demo.example.com"
    journeys:
      - name: "This Journey Will Be Renamed"
        checkpoints:
          - name: "Demo Checkpoint"
            navigation_url: "https://demo.example.com"
            steps:
              - type: wait
                selector: "body"
                timeout: 1000
EOF

echo "📄 Structure file created"
echo ""
echo "🚀 Creating structure and capturing IDs..."
echo ""

# Run creation and capture output
OUTPUT=$(./bin/api-cli create-structure --file /tmp/show-ids-test.yaml --verbose 2>&1)

# Extract IDs from output
PROJECT_ID=$(echo "$OUTPUT" | grep -oE "Created project ID: [0-9]+" | grep -oE "[0-9]+")
GOAL_ID=$(echo "$OUTPUT" | grep -oE "Created goal ID: [0-9]+" | grep -oE "[0-9]+")
JOURNEY_ID=$(echo "$OUTPUT" | grep -oE "Created journey ID: [0-9]+" | grep -oE "[0-9]+")

# Show the output
echo "$OUTPUT"

echo ""
echo "=========================================="
echo "📊 CAPTURED IDs:"
echo "=========================================="
echo "PROJECT ID: $PROJECT_ID"
echo "GOAL ID: $GOAL_ID"
echo "JOURNEY ID: $JOURNEY_ID"
echo "=========================================="
echo ""
echo "The journey with ID $JOURNEY_ID was created with a default name"
echo "and then renamed to 'This Journey Will Be Renamed'"
echo ""
echo "You can verify in Virtuoso UI:"
echo "- Project: $PROJECT_ID"
echo "- Goal: $GOAL_ID"
echo "- Journey: $JOURNEY_ID (should be named 'This Journey Will Be Renamed')"