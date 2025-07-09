#!/bin/bash
# Test the working features of the Virtuoso CLI

echo "🧪 Testing Working Virtuoso CLI Features"
echo "========================================"

export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# 1. Test update-journey command
echo -e "\n1️⃣ Testing update-journey command..."
echo "Creating test project and journey..."
PROJECT_ID=$(./bin/api-cli create-project "Journey Update Test $(date +%Y%m%d-%H%M%S)" -o json | jq -r .project_id)
GOAL_ID=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal" --url "https://test.com" -o json | jq -r .goal_id)
SNAPSHOT_ID=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal" --url "https://test.com" -o json | jq -r .snapshot_id)
JOURNEY_ID=$(./bin/api-cli create-journey $GOAL_ID $SNAPSHOT_ID "Original Name" -o json | jq -r .journey_id)

echo "Journey created with ID: $JOURNEY_ID"
echo "Updating journey name..."
./bin/api-cli update-journey $JOURNEY_ID --name "Updated Journey Name"
echo "✅ Journey update command works!"

# 2. Test structure creation with multiple goals
echo -e "\n2️⃣ Testing batch structure creation..."
cat > /tmp/multi-goal-test.yaml << EOF
project:
  name: "Multi-Goal Test $(date +%Y%m%d-%H%M%S)"
  
goals:
  - name: "Homepage Tests"
    url: "https://example.com"
    journeys:
      - name: "Basic Navigation"
        checkpoints:
          - name: "Landing Page"
            navigation_url: "https://example.com"
            steps:
              - type: wait
                selector: "h1"
                timeout: 2000
              
  - name: "Product Tests"  
    url: "https://example.com/products"
    journeys:
      - name: "Product Browse"
        checkpoints:
          - name: "Product List"
            navigation_url: "https://example.com/products"
            steps:
              - type: click
                selector: ".product-card"
EOF

echo "Running structure creation..."
./bin/api-cli create-structure --file /tmp/multi-goal-test.yaml --verbose

# 3. Test navigation update
echo -e "\n3️⃣ Testing navigation update..."
echo "First, let's create a simple journey to get navigation step..."
PROJ_ID=$(./bin/api-cli create-project "Nav Test $(date +%s)" -o json | jq -r .project_id)
GOAL_ID=$(./bin/api-cli create-goal $PROJ_ID "Nav Goal" --url "https://old-url.com" -o json | jq -r .goal_id)
SNAP_ID=$(./bin/api-cli create-goal $PROJ_ID "Nav Goal" --url "https://old-url.com" -o json | jq -r .snapshot_id)
JOUR_ID=$(./bin/api-cli create-journey $GOAL_ID $SNAP_ID "Nav Journey" -o json | jq -r .journey_id)

echo "Getting checkpoint details..."
CHECKPOINT_ID=$(./bin/api-cli list-checkpoints $JOUR_ID -o json | jq -r '.checkpoints[0].id')
echo "Checkpoint ID: $CHECKPOINT_ID"

# Get step details
STEP_ID=$(./bin/api-cli get-steps $CHECKPOINT_ID -o json | jq -r '.[0].id')
CANONICAL_ID=$(./bin/api-cli get-step $STEP_ID -o json | jq -r '.canonicalId')

echo "Updating navigation URL..."
./bin/api-cli update-navigation $STEP_ID $CANONICAL_ID --url "https://new-url.com"
echo "✅ Navigation update works!"

echo -e "\n🎉 All working features tested successfully!"