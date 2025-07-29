#!/bin/bash

# D365 Virtuoso Tests Deployment Script
# This script deploys all D365 test suites to the Virtuoso platform

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_CLI="./bin/api-cli"
TEST_DIR="d365-virtuoso-tests-final"
PROJECT_NAME="D365 Test Automation Suite"
LOG_FILE="deployment-$(date +%Y%m%d-%H%M%S).log"

# Check if API CLI exists
if [ ! -f "$API_CLI" ]; then
    echo -e "${RED}Error: API CLI not found at $API_CLI${NC}"
    echo "Please build the API CLI first with: make build"
    exit 1
fi

# Check if test directory exists
if [ ! -d "$TEST_DIR" ]; then
    echo -e "${RED}Error: Test directory not found at $TEST_DIR${NC}"
    echo "Please run the conversion script first"
    exit 1
fi

echo -e "${BLUE}=== D365 Virtuoso Tests Deployment ===${NC}"
echo -e "Starting at: $(date)"
echo -e "Project: $PROJECT_NAME"
echo -e "Log file: $LOG_FILE"
echo ""

# Function to deploy a single test
deploy_test() {
    local test_file=$1
    local module=$2
    local test_name=$(basename "$test_file" .yaml)

    echo -n -e "Deploying ${module}/${test_name}... "

    if $API_CLI run-test "$test_file" --project-name "$PROJECT_NAME" >> "$LOG_FILE" 2>&1; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        echo -e "${RED}Failed to deploy: $test_file${NC}" | tee -a "$LOG_FILE"
        return 1
    fi
}

# Function to deploy tests for a module
deploy_module() {
    local module_dir=$1
    local module_name=$2
    local test_count=0
    local success_count=0
    local failed_count=0

    echo -e "\n${YELLOW}Deploying $module_name tests...${NC}"

    for test_file in "$module_dir"/*.yaml; do
        if [ -f "$test_file" ]; then
            ((test_count++))
            if deploy_test "$test_file" "$module_name"; then
                ((success_count++))
            else
                ((failed_count++))
            fi
        fi
    done

    echo -e "Module $module_name: ${GREEN}$success_count successful${NC}, ${RED}$failed_count failed${NC} out of $test_count tests"

    return $failed_count
}

# Track overall statistics
total_tests=0
total_success=0
total_failed=0

# Deploy each module
modules=(
    "sales:Sales Module"
    "customer-service:Customer Service"
    "field-service:Field Service"
    "marketing:Marketing"
    "finance-operations:Finance Operations"
    "project-operations:Project Operations"
    "human-resources:Human Resources"
    "supply-chain:Supply Chain"
    "commerce:Commerce"
)

for module_info in "${modules[@]}"; do
    IFS=':' read -r module_dir module_name <<< "$module_info"
    module_path="$TEST_DIR/$module_dir"

    if [ -d "$module_path" ]; then
        deploy_module "$module_path" "$module_name"
        failed=$?

        # Count tests in module
        module_test_count=$(find "$module_path" -name "*.yaml" | wc -l)
        module_success=$((module_test_count - failed))

        total_tests=$((total_tests + module_test_count))
        total_success=$((total_success + module_success))
        total_failed=$((total_failed + failed))
    else
        echo -e "${YELLOW}Warning: Module directory not found: $module_path${NC}"
    fi
done

# Print summary
echo -e "\n${BLUE}=== Deployment Summary ===${NC}"
echo -e "Total tests: $total_tests"
echo -e "Successful: ${GREEN}$total_success${NC}"
echo -e "Failed: ${RED}$total_failed${NC}"
echo -e "Success rate: $(( (total_success * 100) / total_tests ))%"
echo -e "Completed at: $(date)"

# Save summary to log
{
    echo ""
    echo "=== Deployment Summary ==="
    echo "Total tests: $total_tests"
    echo "Successful: $total_success"
    echo "Failed: $total_failed"
    echo "Success rate: $(( (total_success * 100) / total_tests ))%"
    echo "Completed at: $(date)"
} >> "$LOG_FILE"

# Exit with error if any tests failed
if [ $total_failed -gt 0 ]; then
    echo -e "\n${RED}Deployment completed with errors. Check $LOG_FILE for details.${NC}"
    exit 1
else
    echo -e "\n${GREEN}All tests deployed successfully!${NC}"
    echo -e "\nNext steps:"
    echo -e "1. Update the D365 instance URLs in the YAML files"
    echo -e "2. Configure test credentials in Virtuoso"
    echo -e "3. Execute tests from the Virtuoso platform"
    echo -e "4. Review test results and reports"
fi
