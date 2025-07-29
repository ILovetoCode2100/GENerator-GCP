#!/bin/bash

# Final D365 Deployment Script - Deploys all tests to correct goals

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ID=9369
PROCESSED_DIR="deployment/processed-tests"
FINAL_DIR="deployment/final-tests"

echo -e "${BLUE}=== Final D365 Deployment ===${NC}"
echo "Project ID: $PROJECT_ID"
echo ""

# Create final directory
mkdir -p "$FINAL_DIR"

# Function to get goal ID for module
get_goal_id() {
    case "$1" in
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

# Step 1: Update YAML files
echo -e "${BLUE}Step 1: Preparing YAML files...${NC}"

modules=("sales" "customer-service" "field-service" "marketing" "finance-operations" "project-operations" "human-resources" "supply-chain" "commerce")

for module in "${modules[@]}"; do
    goal_id=$(get_goal_id "$module")
    module_dir="$PROCESSED_DIR/$module"
    final_module_dir="$FINAL_DIR/$module"

    if [ -d "$module_dir" ]; then
        mkdir -p "$final_module_dir"
        echo -e "Processing $module (Goal: $goal_id)..."

        count=0
        for yaml_file in "$module_dir"/*.yaml; do
            if [ -f "$yaml_file" ]; then
                filename=$(basename "$yaml_file")
                final_file="$final_module_dir/$filename"

                # Add project and goal to YAML
                {
                    echo "project: $PROJECT_ID"
                    echo "goal: $goal_id"
                    cat "$yaml_file"
                } > "$final_file"

                ((count++))
            fi
        done
        echo "  Prepared $count test files"
    fi
done

# Step 2: Deploy tests
echo ""
echo -e "${BLUE}Step 2: Deploying tests...${NC}"

total=0
success=0
failed=0

# Clear logs
> deployment-final.log
> deployment-final-errors.log

for module in "${modules[@]}"; do
    goal_id=$(get_goal_id "$module")
    final_module_dir="$FINAL_DIR/$module"

    if [ -d "$final_module_dir" ]; then
        echo -e "\n${YELLOW}Deploying $module tests...${NC}"

        module_count=0
        module_failed=0

        for yaml_file in "$final_module_dir"/*.yaml; do
            if [ -f "$yaml_file" ]; then
                test_name=$(basename "$yaml_file" .yaml)
                echo -n -e "  $test_name... "

                ((total++))

                # Deploy test
                if output=$(./bin/api-cli run-test "$yaml_file" --output json 2>&1); then
                    # Check if journey was created
                    journey_id=$(echo "$output" | jq -r '.journey_id' 2>/dev/null || echo "")
                    if [ -n "$journey_id" ] && [ "$journey_id" != "null" ]; then
                        echo -e "${GREEN}✓${NC} (Journey: $journey_id)"
                        ((success++))
                        ((module_count++))
                        echo "$module,$test_name,$journey_id,success" >> deployment-final.log
                    else
                        echo -e "${YELLOW}⚠${NC} (Partial)"
                        ((failed++))
                        ((module_failed++))
                        echo "$module,$test_name,error,$output" >> deployment-final-errors.log
                    fi
                else
                    echo -e "${RED}✗${NC}"
                    ((failed++))
                    ((module_failed++))
                    echo "$module,$test_name,error,$output" >> deployment-final-errors.log
                fi

                # Rate limiting
                sleep 0.3
            fi
        done

        echo "  Module summary: $module_count deployed, $module_failed failed"
    fi
done

# Step 3: Summary
echo ""
echo -e "${BLUE}=== Final Deployment Summary ===${NC}"
echo "Total tests attempted: $total"
echo -e "Successfully deployed: ${GREEN}$success${NC}"
echo -e "Failed deployments: ${RED}$failed${NC}"

if [ $total -gt 0 ]; then
    success_rate=$(( (success * 100) / total ))
    echo "Success rate: $success_rate%"
fi

# Save summary
cat > deployment-summary.json << EOF
{
  "project_id": $PROJECT_ID,
  "deployment_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "total_tests": $total,
  "successful": $success,
  "failed": $failed,
  "success_rate": "${success_rate}%",
  "goals": {
    "sales": 14137,
    "customer_service": 14138,
    "field_service": 14139,
    "marketing": 14140,
    "finance_operations": 14141,
    "project_operations": 14142,
    "human_resources": 14143,
    "supply_chain": 14144,
    "commerce": 14145
  }
}
EOF

echo ""
echo "Deployment complete!"
echo "Summary saved to: deployment-summary.json"
echo "Success log: deployment-final.log"
if [ $failed -gt 0 ]; then
    echo "Error log: deployment-final-errors.log"
fi

echo ""
echo -e "${BLUE}View your tests:${NC}"
echo "https://app.virtuoso.qa/#/project/$PROJECT_ID"
