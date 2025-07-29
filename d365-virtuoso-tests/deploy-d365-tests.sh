#!/bin/bash

# Deploy D365 Test Automation Suite to Virtuoso Platform
# This script creates the project structure and uploads all test files

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_CLI="${API_CLI:-../bin/api-cli}"
CONFIG_FILE="${CONFIG_FILE:-d365-test-config.yaml}"
MASTER_PROJECT="${MASTER_PROJECT:-d365-master-project.yaml}"

# Ensure API CLI exists
if [ ! -f "$API_CLI" ]; then
    echo -e "${RED}Error: API CLI not found at $API_CLI${NC}"
    echo "Please build the CLI first with: make build"
    exit 1
fi

# Load configuration if exists
if [ -f "$CONFIG_FILE" ]; then
    echo -e "${GREEN}Loading configuration from $CONFIG_FILE${NC}"
    export VIRTUOSO_CONFIG_FILE="$CONFIG_FILE"
fi

echo -e "${GREEN}=== D365 Test Automation Suite Deployment ===${NC}"
echo "Starting deployment process..."

# Function to check command success
check_success() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $1${NC}"
    else
        echo -e "${RED}✗ $1${NC}"
        exit 1
    fi
}

# Step 1: Create the main project using the master YAML
echo -e "\n${YELLOW}Step 1: Creating D365 Test Automation Suite project...${NC}"
PROJECT_ID=$($API_CLI run-test "$MASTER_PROJECT" --output json | jq -r '.project_id // empty')

if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}Failed to create project. Checking if it already exists...${NC}"

    # Try to find existing project
    PROJECT_ID=$($API_CLI list projects --output json | jq -r '.[] | select(.name == "D365 Test Automation Suite") | .id // empty' | head -1)

    if [ -z "$PROJECT_ID" ]; then
        echo -e "${RED}Project creation failed and no existing project found${NC}"
        exit 1
    else
        echo -e "${GREEN}Found existing project with ID: $PROJECT_ID${NC}"
    fi
else
    echo -e "${GREEN}Created project with ID: $PROJECT_ID${NC}"
fi

# Step 2: Deploy each module's tests
echo -e "\n${YELLOW}Step 2: Deploying module tests...${NC}"

# Array of modules and their test counts
declare -A modules=(
    ["sales"]=5
    ["customer-service"]=6
    ["field-service"]=6
    ["marketing"]=6
    ["finance-operations"]=6
    ["project-operations"]=5
    ["human-resources"]=5
    ["supply-chain"]=6
    ["commerce"]=6
)

# Deploy test suites for each module
total_tests=0
deployed_tests=0

for module in "${!modules[@]}"; do
    echo -e "\n${YELLOW}Deploying $module module tests...${NC}"

    # Find and deploy all YAML files in the module directory
    module_dir="$SCRIPT_DIR/$module"
    if [ -d "$module_dir" ]; then
        for test_file in "$module_dir"/*.yaml; do
            if [ -f "$test_file" ]; then
                test_name=$(basename "$test_file" .yaml)
                echo -n "  Deploying $test_name... "

                # Deploy the test file
                if $API_CLI run-test "$test_file" --project-id "$PROJECT_ID" --output json > /dev/null 2>&1; then
                    echo -e "${GREEN}✓${NC}"
                    ((deployed_tests++))
                else
                    echo -e "${RED}✗${NC}"
                    echo -e "${YELLOW}    Warning: Failed to deploy $test_file${NC}"
                fi
                ((total_tests++))
            fi
        done
    else
        echo -e "${RED}  Warning: Module directory $module_dir not found${NC}"
    fi
done

# Step 3: Create test execution schedules (optional)
echo -e "\n${YELLOW}Step 3: Setting up test execution schedules...${NC}"

# This is a placeholder for scheduling configuration
# You can add scheduling logic here if the API supports it

echo -e "${GREEN}Skipping schedule setup (configure manually if needed)${NC}"

# Step 4: Generate deployment summary
echo -e "\n${YELLOW}Step 4: Generating deployment summary...${NC}"

SUMMARY_FILE="$SCRIPT_DIR/deployment-summary.txt"
{
    echo "D365 Test Automation Suite Deployment Summary"
    echo "============================================="
    echo "Date: $(date)"
    echo "Project ID: $PROJECT_ID"
    echo "Total Tests: $total_tests"
    echo "Successfully Deployed: $deployed_tests"
    echo "Failed Deployments: $((total_tests - deployed_tests))"
    echo ""
    echo "Module Breakdown:"
    for module in "${!modules[@]}"; do
        echo "  - $module: ${modules[$module]} test suites"
    done
    echo ""
    echo "Next Steps:"
    echo "1. Run all tests: ./run-all-d365-tests.sh"
    echo "2. Run specific module: ./run-all-d365-tests.sh --module sales"
    echo "3. View results in Virtuoso platform"
} > "$SUMMARY_FILE"

cat "$SUMMARY_FILE"

# Step 5: Validate deployment
echo -e "\n${YELLOW}Step 5: Validating deployment...${NC}"

# List all goals in the project to verify
echo "Verifying project structure..."
GOAL_COUNT=$($API_CLI list goals "$PROJECT_ID" --output json | jq '. | length')

if [ "$GOAL_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Found $GOAL_COUNT goals in the project${NC}"
else
    echo -e "${RED}✗ No goals found in the project${NC}"
fi

echo -e "\n${GREEN}=== Deployment Complete ===${NC}"
echo -e "Successfully deployed $deployed_tests out of $total_tests tests"
echo -e "Project ID: ${YELLOW}$PROJECT_ID${NC}"

# Save project ID for use by other scripts
echo "$PROJECT_ID" > "$SCRIPT_DIR/.project-id"

exit 0
