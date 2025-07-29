#!/bin/bash

# Verify D365 Virtuoso Deployment
# This script checks that the deployment was successful

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== D365 Deployment Verification ===${NC}"
echo ""

# Check deployment state
if [ -f "deployment-state.json" ]; then
    echo -e "${GREEN}✓${NC} Deployment state file found"

    # Extract key information
    PROJECT_ID=$(jq -r '.project_id' deployment-state.json)
    PROJECT_NAME=$(jq -r '.project_name' deployment-state.json)
    D365_INSTANCE=$(jq -r '.d365_instance' deployment-state.json)
    TOTAL_TESTS=$(jq -r '.total_tests' deployment-state.json)
    PROCESSED_TESTS=$(jq -r '.processed_tests' deployment-state.json)
    FAILED_TESTS=$(jq -r '.failed_tests' deployment-state.json)
    STATUS=$(jq -r '.status' deployment-state.json)

    echo ""
    echo "Deployment Details:"
    echo "  Project: $PROJECT_NAME (ID: $PROJECT_ID)"
    echo "  D365 Instance: $D365_INSTANCE"
    echo "  Total Tests: $TOTAL_TESTS"
    echo "  Successful: $PROCESSED_TESTS"
    echo "  Failed: $FAILED_TESTS"
    echo "  Status: $STATUS"
    echo ""

    # Verify project exists in Virtuoso
    echo -n "Verifying project exists in Virtuoso... "
    if ./bin/api-cli list-projects --output json | jq -e ".projects[] | select(.id == $PROJECT_ID)" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${YELLOW}⚠${NC} Could not verify project"
    fi

    # Check processed tests
    echo -n "Checking processed test files... "
    PROCESSED_COUNT=$(find deployment/processed-tests -name "*.yaml" | wc -l)
    echo -e "${GREEN}✓${NC} Found $PROCESSED_COUNT processed files"

    # Module breakdown
    echo ""
    echo "Module Breakdown:"
    jq -r '.modules | to_entries[] | "  \(.key): \(.value.success)/\(.value.total) tests deployed"' deployment-state.json

else
    echo -e "${YELLOW}⚠${NC} No deployment state file found"
    echo "Run ./deploy-d365-comprehensive.sh first"
fi

echo ""
echo -e "${BLUE}Quick Links:${NC}"
echo "  Virtuoso Platform: https://app.virtuoso.qa/"
echo "  Project ID: $PROJECT_ID"
echo "  D365 Test Instance: https://$D365_INSTANCE.crm.dynamics.com"

echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Log into Virtuoso and navigate to project $PROJECT_ID"
echo "2. Configure D365 test credentials"
echo "3. Run a test to verify setup"
echo "4. Set up execution schedules"
