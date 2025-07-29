#!/bin/bash

# Comprehensive D365 Virtuoso Test Deployment Script
# This script ensures complete deployment with proper error handling

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
STATE_FILE="deployment-state.json"
PROCESSED_DIR="deployment/processed-tests"

# D365 Instance Configuration
# IMPORTANT: This uses the D365_INSTANCE environment variable
# If not set, defaults to 'demo' - MUST be updated for production
D365_INSTANCE="${D365_INSTANCE:-demo}"

echo -e "${BLUE}=== D365 Virtuoso Comprehensive Test Deployment ===${NC}"
echo -e "D365 Instance: ${YELLOW}$D365_INSTANCE${NC}"
echo -e "Log file: $LOG_FILE"
echo ""

# Verify D365 instance is set
if [ "$D365_INSTANCE" = "demo" ]; then
    echo -e "${YELLOW}WARNING: Using default 'demo' instance.${NC}"
    echo -e "${YELLOW}For production deployment, set D365_INSTANCE environment variable:${NC}"
    echo -e "${YELLOW}  export D365_INSTANCE=your-actual-instance${NC}"
    echo ""
    read -p "Continue with 'demo' instance? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Deployment cancelled."
        exit 1
    fi
fi

# Create processed directory
mkdir -p "$PROCESSED_DIR"

# Initialize deployment state
init_state() {
    cat > "$STATE_FILE" << EOF
{
    "project_id": null,
    "project_name": "$PROJECT_NAME",
    "d365_instance": "$D365_INSTANCE",
    "total_tests": 0,
    "processed_tests": 0,
    "failed_tests": 0,
    "modules": {},
    "start_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "status": "in_progress"
}
EOF
}

# Update state
update_state() {
    local key=$1
    local value=$2
    local temp_file=$(mktemp)
    jq "$key = $value" "$STATE_FILE" > "$temp_file" && mv "$temp_file" "$STATE_FILE"
}

# Process YAML files to replace [instance] with environment variable
process_yaml_files() {
    echo -e "${BLUE}Processing YAML files to use D365 instance: $D365_INSTANCE${NC}"

    local count=0
    for module_dir in "$TEST_DIR"/*; do
        if [ -d "$module_dir" ]; then
            local module_name=$(basename "$module_dir")
            mkdir -p "$PROCESSED_DIR/$module_name"

            for yaml_file in "$module_dir"/*.yaml; do
                if [ -f "$yaml_file" ]; then
                    local filename=$(basename "$yaml_file")
                    local processed_file="$PROCESSED_DIR/$module_name/$filename"

                    # Replace [instance] with actual instance name
                    sed "s/\[instance\]/$D365_INSTANCE/g" "$yaml_file" > "$processed_file"
                    ((count++))

                    echo -n "."
                fi
            done
        fi
    done

    echo ""
    echo -e "${GREEN}Processed $count YAML files${NC}"
    update_state '.total_tests' "$count"
}

# Deploy a single test
deploy_test() {
    local test_file=$1
    local module=$2
    local test_name=$(basename "$test_file" .yaml)

    echo -n -e "Deploying ${module}/${test_name}... "

    # Create project if it doesn't exist (first test only)
    local project_id=$(jq -r '.project_id' "$STATE_FILE")
    if [ "$project_id" = "null" ]; then
        # Try to find existing project first
        project_id=$($API_CLI list-projects --output json | jq -r ".projects[] | select(.name == \"$PROJECT_NAME\") | .id" | head -1)

        if [ -z "$project_id" ]; then
            # Create new project
            echo -e "\n${BLUE}Creating project: $PROJECT_NAME${NC}"
            local create_output=$($API_CLI create-project "$PROJECT_NAME" --output json 2>&1 | tee -a "$LOG_FILE")
            project_id=$(echo "$create_output" | jq -r '.project.id' 2>/dev/null || echo "")

            if [ -z "$project_id" ]; then
                echo -e "${RED}Failed to create project${NC}"
                echo "$create_output" >> "$LOG_FILE"
                return 1
            fi
        fi

        update_state '.project_id' "\"$project_id\""
        echo -e "\n${GREEN}Using project ID: $project_id${NC}"
    fi

    # Deploy the test
    if $API_CLI run-test "$test_file" --project-name "$PROJECT_NAME" >> "$LOG_FILE" 2>&1; then
        echo -e "${GREEN}✓${NC}"

        # Update state
        local processed=$(jq -r '.processed_tests' "$STATE_FILE")
        update_state '.processed_tests' "$((processed + 1))"

        # Update module stats
        local module_key=".modules[\"$module\"]"
        local module_stats=$(jq -r "$module_key // {}" "$STATE_FILE")
        if [ "$module_stats" = "{}" ]; then
            update_state "$module_key" '{"total": 0, "success": 0, "failed": 0}'
        fi

        local module_success=$(jq -r "$module_key.success // 0" "$STATE_FILE")
        update_state "$module_key.success" "$((module_success + 1))"

        return 0
    else
        echo -e "${RED}✗${NC}"
        echo -e "${RED}Failed to deploy: $test_file${NC}" | tee -a "$LOG_FILE"

        # Update failed count
        local failed=$(jq -r '.failed_tests' "$STATE_FILE")
        update_state '.failed_tests' "$((failed + 1))"

        # Update module stats
        local module_key=".modules[\"$module\"]"
        local module_failed=$(jq -r "$module_key.failed // 0" "$STATE_FILE")
        update_state "$module_key.failed" "$((module_failed + 1))"

        return 1
    fi
}

# Deploy tests for a module
deploy_module() {
    local module_dir=$1
    local module_name=$2
    local test_count=0
    local success_count=0
    local failed_count=0

    echo -e "\n${YELLOW}Deploying $module_name tests...${NC}"

    # Count tests first
    for test_file in "$module_dir"/*.yaml; do
        if [ -f "$test_file" ]; then
            ((test_count++))
        fi
    done

    # Update module total
    update_state ".modules[\"$module_name\"].total" "$test_count"

    # Deploy tests
    for test_file in "$module_dir"/*.yaml; do
        if [ -f "$test_file" ]; then
            if deploy_test "$test_file" "$module_name"; then
                ((success_count++))
            else
                ((failed_count++))

                # Check if we should continue
                if [ $failed_count -ge 3 ]; then
                    echo -e "${RED}Too many failures in module $module_name, stopping${NC}"
                    break
                fi
            fi
        fi
    done

    echo -e "Module $module_name: ${GREEN}$success_count successful${NC}, ${RED}$failed_count failed${NC} out of $test_count tests"

    return $failed_count
}

# Main deployment process
main() {
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

    # Initialize state
    init_state

    # Process YAML files
    process_yaml_files

    # Deploy each module
    local modules=(
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

    local total_failed=0

    for module_info in "${modules[@]}"; do
        IFS=':' read -r module_dir module_name <<< "$module_info"
        module_path="$PROCESSED_DIR/$module_dir"

        if [ -d "$module_path" ]; then
            deploy_module "$module_path" "$module_name"
            failed=$?
            total_failed=$((total_failed + failed))

            # Stop if too many total failures
            if [ $total_failed -ge 10 ]; then
                echo -e "${RED}Too many total failures, stopping deployment${NC}"
                break
            fi
        else
            echo -e "${YELLOW}Warning: Module directory not found: $module_path${NC}"
        fi
    done

    # Update final state
    update_state '.end_time' "\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\""
    if [ $total_failed -eq 0 ]; then
        update_state '.status' '"completed"'
    else
        update_state '.status' '"completed_with_errors"'
    fi

    # Print summary
    echo -e "\n${BLUE}=== Deployment Summary ===${NC}"
    echo -e "State file: $STATE_FILE"

    local total=$(jq -r '.total_tests' "$STATE_FILE")
    local processed=$(jq -r '.processed_tests' "$STATE_FILE")
    local failed=$(jq -r '.failed_tests' "$STATE_FILE")
    local success=$((processed - failed))

    echo -e "Total tests: $total"
    echo -e "Successful: ${GREEN}$success${NC}"
    echo -e "Failed: ${RED}$failed${NC}"

    if [ $total -gt 0 ]; then
        local success_rate=$(( (success * 100) / total ))
        echo -e "Success rate: $success_rate%"
    fi

    echo -e "\nProject ID: $(jq -r '.project_id' "$STATE_FILE")"
    echo -e "D365 Instance: $D365_INSTANCE"

    # Save summary to log
    {
        echo ""
        echo "=== Deployment Summary ==="
        echo "Total tests: $total"
        echo "Successful: $success"
        echo "Failed: $failed"
        echo "Project ID: $(jq -r '.project_id' "$STATE_FILE")"
        echo "Completed at: $(date)"
    } >> "$LOG_FILE"

    # Exit status based on failures
    if [ $failed -gt 0 ]; then
        echo -e "\n${RED}Deployment completed with errors. Check $LOG_FILE for details.${NC}"
        echo -e "\n${YELLOW}IMPORTANT Manual Steps Required:${NC}"
        echo -e "1. Update D365 test user credentials in Virtuoso platform"
        echo -e "2. Configure test execution schedules"
        echo -e "3. Set up email notifications for test results"
        echo -e "4. Verify tests can access your D365 instance at: https://$D365_INSTANCE.crm.dynamics.com"
        exit 1
    else
        echo -e "\n${GREEN}All tests deployed successfully!${NC}"
        echo -e "\n${YELLOW}IMPORTANT Manual Steps Required:${NC}"
        echo -e "1. ${YELLOW}Update D365 test user credentials${NC} in Virtuoso platform"
        echo -e "   - Go to Virtuoso platform > Settings > Test Credentials"
        echo -e "   - Add credentials for D365 test users with appropriate permissions"
        echo -e ""
        echo -e "2. ${YELLOW}Configure test execution schedules${NC}"
        echo -e "   - Navigate to your project in Virtuoso"
        echo -e "   - Set up execution schedules for different test suites"
        echo -e ""
        echo -e "3. ${YELLOW}Set up notifications${NC}"
        echo -e "   - Configure email/webhook notifications for test results"
        echo -e ""
        echo -e "4. ${YELLOW}Verify D365 access${NC}"
        echo -e "   - Ensure tests can access: https://$D365_INSTANCE.crm.dynamics.com"
        echo -e "   - Check firewall rules if running from specific IPs"
        echo -e ""
        echo -e "5. ${YELLOW}Run a smoke test${NC}"
        echo -e "   - Execute a simple test like 'sales-lead-001' to verify setup"
        echo -e ""
        echo -e "Project deployed as: '${GREEN}$PROJECT_NAME${NC}'"
        echo -e "View in Virtuoso at: https://app.virtuoso.qa/"
    fi
}

# Run main function
main "$@"
