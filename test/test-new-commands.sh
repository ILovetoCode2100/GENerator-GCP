#!/bin/bash

# Test script for new Virtuoso CLI commands
# This script tests the newly implemented commands

set -e  # Exit on error

echo "====================================="
echo "Testing New Virtuoso CLI Commands"
echo "====================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test configuration
TEST_JOURNEY_ID="${JOURNEY_ID:-608048}"
TEST_STEP_ID="${STEP_ID:-19636330}"

echo -e "\n${YELLOW}Test Configuration:${NC}"
echo "Journey ID: $TEST_JOURNEY_ID"
echo "Step ID: $TEST_STEP_ID"

# Function to check command result
check_result() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $1 succeeded${NC}"
    else
        echo -e "${RED}✗ $1 failed${NC}"
        exit 1
    fi
}

echo -e "\n${YELLOW}1. Testing update-journey command${NC}"
echo "Renaming journey to 'Renamed Journey Test'..."
./bin/api-cli update-journey $TEST_JOURNEY_ID --name "Renamed Journey Test"
check_result "Journey rename"

echo -e "\n${YELLOW}2. Testing get-step command${NC}"
echo "Getting step details for step $TEST_STEP_ID..."
STEP_DETAILS=$(./bin/api-cli get-step $TEST_STEP_ID -o json)
check_result "Get step details"

# Extract canonical ID
CANONICAL_ID=$(echo $STEP_DETAILS | jq -r .canonical_id)
echo "Extracted Canonical ID: $CANONICAL_ID"

if [ "$CANONICAL_ID" == "null" ] || [ -z "$CANONICAL_ID" ]; then
    echo -e "${RED}Failed to extract canonical ID${NC}"
    exit 1
fi

echo -e "\n${YELLOW}3. Testing update-navigation command${NC}"
echo "Updating navigation URL..."
./bin/api-cli update-navigation $TEST_STEP_ID "$CANONICAL_ID" --url "https://updated.example.com"
check_result "Navigation update"

echo -e "\n${YELLOW}4. Testing list-checkpoints command${NC}"
echo "Listing checkpoints for journey $TEST_JOURNEY_ID..."
./bin/api-cli list-checkpoints $TEST_JOURNEY_ID
check_result "List checkpoints"

echo -e "\n${YELLOW}5. Testing create-structure command (dry-run)${NC}"
# Create a test structure file
cat > test-structure-temp.yaml << 'EOF'
project:
  name: "Test Structure Project"
  description: "Created by test script"

goals:
  - name: "Test Goal"
    url: "https://example.com"
    journeys:
      - name: "Test Journey"
        checkpoints:
          - name: "Navigate to Home"
            navigation_url: "https://example.com/home"
            steps:
              - type: wait
                selector: ".homepage"
                timeout: 5000
          - name: "Click Login"
            steps:
              - type: click
                selector: "#login-button"
              - type: wait
                selector: ".login-form"
EOF

echo "Testing dry-run mode..."
./bin/api-cli create-structure --file test-structure-temp.yaml --dry-run
check_result "Structure dry-run"

echo -e "\n${YELLOW}6. Testing create-structure command (verbose dry-run)${NC}"
./bin/api-cli create-structure --file test-structure-temp.yaml --dry-run --verbose | head -20
check_result "Structure verbose dry-run"

# Cleanup
rm -f test-structure-temp.yaml

echo -e "\n${GREEN}====================================="
echo "All tests completed successfully!"
echo "=====================================${NC}"

echo -e "\n${YELLOW}Additional Manual Tests:${NC}"
echo "1. To test actual structure creation (creates real resources):"
echo "   ./bin/api-cli create-structure --file examples/enhanced-structure.yaml --verbose"
echo ""
echo "2. To test with existing project ID:"
echo "   ./bin/api-cli create-structure --file examples/enhanced-structure.yaml --project-id 9056"
echo ""
echo "3. To test navigation update with new tab:"
echo "   ./bin/api-cli update-navigation $TEST_STEP_ID \"$CANONICAL_ID\" --url \"https://example.com\" --new-tab"
echo ""
echo "4. To test different output formats:"
echo "   ./bin/api-cli list-checkpoints $TEST_JOURNEY_ID -o json"
echo "   ./bin/api-cli list-checkpoints $TEST_JOURNEY_ID -o yaml"
echo "   ./bin/api-cli list-checkpoints $TEST_JOURNEY_ID -o ai"