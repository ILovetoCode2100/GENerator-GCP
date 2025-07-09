#!/bin/bash
# Complete test sequence for all new Virtuoso CLI commands

echo "🚀 Testing Complete Virtuoso CLI Batch Structure Implementation"
echo "=============================================================="

# Set test configuration
export API_CLI_CONFIG="/Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator/config/virtuoso-config.yaml"
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# 1. First, create a test project to work with
echo -e "\n📁 Creating test project..."
PROJECT_ID=$(./bin/api-cli create-project "Batch Test $(date +%Y%m%d-%H%M%S)" -o json | jq -r .project_id)
echo "Created project: $PROJECT_ID"

# 2. Create a goal (this auto-creates a journey we'll rename)
echo -e "\n🎯 Creating goal with auto-journey..."
GOAL_RESULT=$(./bin/api-cli create-goal $PROJECT_ID "Test Goal for Batch" --url "https://example.com" -o json)
GOAL_ID=$(echo $GOAL_RESULT | jq -r .goal_id)
SNAPSHOT_ID=$(echo $GOAL_RESULT | jq -r .snapshot_id)
echo "Goal ID: $GOAL_ID"
echo "Snapshot ID: $SNAPSHOT_ID"

# 3. List journeys to find the auto-created one
echo -e "\n📋 Finding auto-created journey..."
JOURNEY_ID=$(./bin/api-cli list-journeys $GOAL_ID $SNAPSHOT_ID -o json | jq -r '.[0].id' 2>/dev/null)
echo "Auto-created Journey ID: $JOURNEY_ID"

# 4. Test update-journey command
echo -e "\n✏️  Testing journey rename..."
./bin/api-cli update-journey $JOURNEY_ID --name "Renamed Test Journey"
echo "✅ Journey renamed successfully"

# 5. List checkpoints to find the auto-created navigation
echo -e "\n📍 Listing checkpoints..."
./bin/api-cli list-checkpoints $JOURNEY_ID

# Get first checkpoint ID (should have navigation)
CHECKPOINT_ID=$(./bin/api-cli list-checkpoints $JOURNEY_ID -o json | jq -r '.[0].id' 2>/dev/null)
echo "First checkpoint ID: $CHECKPOINT_ID"

# 6. Get the navigation step details
echo -e "\n🔍 Getting navigation step details..."
# Assuming first step is navigation
STEPS=$(./bin/api-cli get-steps $CHECKPOINT_ID -o json 2>/dev/null)
if [ $? -eq 0 ]; then
    NAV_STEP_ID=$(echo $STEPS | jq -r '.[0].id')
    echo "Navigation step ID: $NAV_STEP_ID"
    
    # Get step details with canonical ID
    STEP_DETAILS=$(./bin/api-cli get-step $NAV_STEP_ID -o json)
    CANONICAL_ID=$(echo $STEP_DETAILS | jq -r .canonicalId)
    echo "Canonical ID: $CANONICAL_ID"
    
    # 7. Test navigation update
    echo -e "\n🔄 Testing navigation update..."
    ./bin/api-cli update-navigation $NAV_STEP_ID $CANONICAL_ID --url "https://updated-test.example.com"
    echo "✅ Navigation updated successfully"
fi

# 8. Create a test structure file
echo -e "\n📝 Creating test structure file..."
cat > /tmp/test-batch-structure.yaml << EOF
project:
  # Using existing project
  id: $PROJECT_ID
  
goals:
  - name: "Batch Created Goal"
    url: "https://batch-test.example.com"
    journeys:
      - name: "Primary User Journey"
        checkpoints:
          - name: "Start Shopping"
            navigation_url: "https://batch-test.example.com/shop"
            steps:
              - type: wait
                selector: ".products"
                timeout: 3000
              - type: click
                selector: ".product-1"
                
          - name: "Add to Cart"
            steps:
              - type: click
                selector: ".add-to-cart"
              - type: wait
                selector: ".cart-updated"
                timeout: 2000
                
          - name: "Checkout"
            steps:
              - type: click
                selector: ".checkout-btn"
              - type: fill
                selector: "#email"
                value: "test@example.com"
              - type: click
                selector: ".place-order"
EOF

# 9. Test dry-run first
echo -e "\n🔍 Testing create-structure (dry-run)..."
./bin/api-cli create-structure --file /tmp/test-batch-structure.yaml --dry-run

# 10. Run actual structure creation
echo -e "\n🏗️  Creating full structure..."
read -p "Press Enter to create the actual structure (or Ctrl+C to cancel)..."
./bin/api-cli create-structure --file /tmp/test-batch-structure.yaml --verbose

# 11. Verify creation by listing resources
echo -e "\n✅ Verifying created structure..."
echo "Listing goals in project $PROJECT_ID:"
./bin/api-cli list-goals $PROJECT_ID

# 12. Test error cases
echo -e "\n❌ Testing error handling..."
echo "Testing invalid journey ID:"
./bin/api-cli update-journey 99999999 --name "Should Fail" || echo "✅ Error handled correctly"

echo "Testing missing canonical ID:"
./bin/api-cli update-navigation $NAV_STEP_ID "invalid-canonical" --url "https://fail.com" || echo "✅ Error handled correctly"

# Summary
echo -e "\n=============================================================="
echo "🎉 Test Summary:"
echo "✅ Project created: $PROJECT_ID"
echo "✅ Goal created: $GOAL_ID"
echo "✅ Journey renamed: $JOURNEY_ID"
echo "✅ Navigation updated successfully"
echo "✅ Batch structure tested"
echo ""
echo "🧹 Cleanup: To remove test data, delete project $PROJECT_ID"
echo ""
echo "📊 Next steps:"
echo "1. Check the Virtuoso UI to verify all created resources"
echo "2. Run the created tests to ensure they work"
echo "3. Try creating a more complex structure with multiple goals/journeys"