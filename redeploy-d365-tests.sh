#!/bin/bash

# Redeploy D365 Tests - Create new journeys with all steps in existing goals

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ID=9369

echo -e "${BLUE}=== D365 Test Redeployment Script ===${NC}"
echo "This script will create new journeys with ALL steps in existing goals"
echo ""

# Function to get goal ID and snapshot for module
get_goal_info() {
    local module=$1
    case "$module" in
        "sales") echo "14137" ;;
        "customer-service") echo "14138" ;;
        "field-service") echo "14139" ;;
        "marketing") echo "14140" ;;
        "finance-operations") echo "14141" ;;
        "project-operations") echo "14142" ;;
        "human-resources") echo "14143" ;;
        "supply-chain") echo "14144" ;;
        "commerce") echo "14145" ;;
        *) echo "" ;;
    esac
}

# Function to deploy test to existing goal
deploy_to_existing_goal() {
    local goal_id=$1
    local yaml_file=$2
    local test_name=$(basename "$yaml_file" .yaml)

    echo -n -e "  $test_name... "

    # Get snapshot ID for the goal
    local snapshot_id=$(./bin/api-cli goal-get $goal_id --output json 2>/dev/null | jq -r '.snapshotId' || echo "")

    if [ -z "$snapshot_id" ] || [ "$snapshot_id" = "null" ]; then
        echo -e "${RED}✗ (no snapshot)${NC}"
        return 1
    fi

    # Extract test name from YAML
    local journey_name=$(yq eval '.name' "$yaml_file")

    # Create journey
    local journey_output=$(./bin/api-cli journey-create $goal_id "$journey_name" $snapshot_id --output json 2>&1)
    local journey_id=$(echo "$journey_output" | jq -r '.id' 2>/dev/null || echo "")

    if [ -z "$journey_id" ] || [ "$journey_id" = "null" ]; then
        echo -e "${RED}✗ (journey creation failed)${NC}"
        echo "$journey_output" >> redeploy-errors.log
        return 1
    fi

    # Create checkpoint
    local checkpoint_output=$(./bin/api-cli checkpoint-create $goal_id "Test Steps" $snapshot_id --output json 2>&1)
    local checkpoint_id=$(echo "$checkpoint_output" | jq -r '.id' 2>/dev/null || echo "")

    if [ -z "$checkpoint_id" ] || [ "$checkpoint_id" = "null" ]; then
        echo -e "${RED}✗ (checkpoint creation failed)${NC}"
        echo "$checkpoint_output" >> redeploy-errors.log
        return 1
    fi

    # Add checkpoint to journey
    ./bin/api-cli journey-add-checkpoint $journey_id $checkpoint_id >/dev/null 2>&1

    # Add all steps
    local step_count=$(yq eval '.steps | length' "$yaml_file")
    local steps_added=0

    for ((i=0; i<$step_count; i++)); do
        local step_type=$(yq eval ".steps[$i] | keys | .[0]" "$yaml_file")
        local position=$((i + 1))

        case "$step_type" in
            "navigate")
                local url=$(yq eval ".steps[$i].navigate" "$yaml_file")
                ./bin/api-cli step-navigate to $checkpoint_id "$url" $position >/dev/null 2>&1 && ((steps_added++))
                ;;
            "wait")
                local duration=$(yq eval ".steps[$i].wait" "$yaml_file")
                ./bin/api-cli step-wait time $checkpoint_id $duration $position >/dev/null 2>&1 && ((steps_added++))
                ;;
            "click")
                local selector=$(yq eval ".steps[$i].click" "$yaml_file")
                ./bin/api-cli step-interact click $checkpoint_id "$selector" $position >/dev/null 2>&1 && ((steps_added++))
                ;;
            "write")
                local selector=$(yq eval ".steps[$i].write.selector" "$yaml_file")
                local text=$(yq eval ".steps[$i].write.text" "$yaml_file")
                ./bin/api-cli step-interact write $checkpoint_id "$selector" "$text" $position >/dev/null 2>&1 && ((steps_added++))
                ;;
            "select")
                local selector=$(yq eval ".steps[$i].select.selector" "$yaml_file")
                local option=$(yq eval ".steps[$i].select.option" "$yaml_file")
                ./bin/api-cli step-interact select-option $checkpoint_id "$selector" "$option" $position >/dev/null 2>&1 && ((steps_added++))
                ;;
            "assert")
                local text=$(yq eval ".steps[$i].assert" "$yaml_file")
                ./bin/api-cli step-assert text $checkpoint_id "$text" $position >/dev/null 2>&1 && ((steps_added++))
                ;;
            "store")
                local type=$(yq eval ".steps[$i].store.type" "$yaml_file")
                local selector=$(yq eval ".steps[$i].store.selector" "$yaml_file")
                local variable=$(yq eval ".steps[$i].store.variable" "$yaml_file")
                # Remove $ from variable name if present
                variable=${variable#\$}
                ./bin/api-cli step-data store-$type $checkpoint_id "$selector" "$variable" $position >/dev/null 2>&1 && ((steps_added++))
                ;;
        esac

        # Rate limiting
        sleep 0.1
    done

    if [ $steps_added -eq $step_count ]; then
        echo -e "${GREEN}✓${NC} (Journey: $journey_id, Steps: $steps_added)"
        echo "$module,$test_name,$journey_id,$checkpoint_id,success" >> redeploy-success.log
    else
        echo -e "${YELLOW}⚠${NC} (Journey: $journey_id, Steps: $steps_added/$step_count)"
        echo "$module,$test_name,$journey_id,$checkpoint_id,partial,$steps_added/$step_count" >> redeploy-partial.log
    fi

    return 0
}

# Clear logs
> redeploy-success.log
> redeploy-partial.log
> redeploy-errors.log

# Main processing
modules=("sales" "customer-service" "field-service" "marketing" "finance-operations" "project-operations")

# Test with just one module first
echo -e "${BLUE}Testing with Sales module first...${NC}"

module="sales"
goal_id=$(get_goal_info "$module")
yaml_dir="deployment/final-tests/$module"

if [ -d "$yaml_dir" ]; then
    echo -e "\n${YELLOW}Deploying $module tests to goal $goal_id...${NC}"

    # Process first 3 tests only for testing
    count=0
    for yaml_file in "$yaml_dir"/*.yaml; do
        if [ -f "$yaml_file" ] && [ $count -lt 3 ]; then
            deploy_to_existing_goal "$goal_id" "$yaml_file"
            ((count++))
        fi
    done
fi

# Summary
echo -e "\n${BLUE}=== Redeployment Summary ===${NC}"
success_count=$(wc -l < redeploy-success.log 2>/dev/null || echo "0")
partial_count=$(wc -l < redeploy-partial.log 2>/dev/null || echo "0")
error_count=$(wc -l < redeploy-errors.log 2>/dev/null || echo "0")

echo "Successfully deployed: $success_count"
echo "Partially deployed: $partial_count"
echo "Failed: $error_count"

echo -e "\n${GREEN}Test deployment complete!${NC}"
echo "Check the Virtuoso UI to verify the tests were created correctly."
echo ""
echo "To deploy ALL tests, modify the script to process all modules and remove the count limit."
