#!/bin/bash

# Correct D365 Deployment Script - Uses existing goals and deploys tests properly

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

# Goal ID mappings (discovered from API)
declare -A GOAL_IDS
GOAL_IDS["sales"]=14137
GOAL_IDS["customer-service"]=14138
GOAL_IDS["field-service"]=14139
GOAL_IDS["marketing"]=14140
GOAL_IDS["finance-operations"]=14141
GOAL_IDS["project-operations"]=14142
GOAL_IDS["human-resources"]=14143
GOAL_IDS["supply-chain"]=14144
GOAL_IDS["commerce"]=14145

echo -e "${BLUE}=== Correct D365 Deployment ===${NC}"
echo "Project ID: $PROJECT_ID"
echo ""

# Create final directory for updated YAML files
mkdir -p "$FINAL_DIR"

# Step 1: Update YAML files with project and goal IDs
echo -e "${BLUE}Step 1: Updating YAML files with correct IDs...${NC}"

for module in "${!GOAL_IDS[@]}"; do
    goal_id=${GOAL_IDS[$module]}
    module_dir="$PROCESSED_DIR/$module"
    final_module_dir="$FINAL_DIR/$module"

    if [ -d "$module_dir" ]; then
        mkdir -p "$final_module_dir"
        echo -e "\n${YELLOW}Processing $module (Goal ID: $goal_id)...${NC}"

        for yaml_file in "$module_dir"/*.yaml; do
            if [ -f "$yaml_file" ]; then
                filename=$(basename "$yaml_file")
                final_file="$final_module_dir/$filename"

                # Add project and goal fields to the beginning of the file
                {
                    echo "project: $PROJECT_ID"
                    echo "goal: $goal_id"
                    cat "$yaml_file"
                } > "$final_file"

                echo -n "."
            fi
        done
        echo -e " ${GREEN}✓${NC}"
    fi
done

echo ""
echo -e "${GREEN}All YAML files updated!${NC}"

# Step 2: Deploy tests
echo ""
echo -e "${BLUE}Step 2: Deploying tests...${NC}"

total=0
success=0
failed=0

for module in "${!GOAL_IDS[@]}"; do
    goal_id=${GOAL_IDS[$module]}
    final_module_dir="$FINAL_DIR/$module"

    if [ -d "$final_module_dir" ]; then
        echo -e "\n${YELLOW}Deploying $module tests (Goal $goal_id)...${NC}"

        for yaml_file in "$final_module_dir"/*.yaml; do
            if [ -f "$yaml_file" ]; then
                test_name=$(basename "$yaml_file" .yaml)
                echo -n -e "  $test_name... "

                ((total++))

                # Deploy WITHOUT --project-name flag
                if ./bin/api-cli run-test "$yaml_file" >> deployment-correct.log 2>&1; then
                    echo -e "${GREEN}✓${NC}"
                    ((success++))
                else
                    echo -e "${RED}✗${NC}"
                    ((failed++))
                    # Log the error
                    echo "Failed: $yaml_file" >> deployment-correct-errors.log
                fi

                # Small delay to avoid rate limiting
                sleep 0.5
            fi
        done
    fi
done

# Step 3: Summary
echo ""
echo -e "${BLUE}=== Deployment Summary ===${NC}"
echo "Total tests: $total"
echo -e "Successful: ${GREEN}$success${NC}"
echo -e "Failed: ${RED}$failed${NC}"

if [ $success -gt 0 ]; then
    success_rate=$(( (success * 100) / total ))
    echo "Success rate: $success_rate%"
fi

echo ""
echo "Deployment log: deployment-correct.log"
if [ $failed -gt 0 ]; then
    echo "Error log: deployment-correct-errors.log"
fi

echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Check Virtuoso platform for deployed tests"
echo "2. Configure D365 credentials"
echo "3. Run sample tests to validate"

# Show link to project
echo ""
echo -e "${BLUE}View your tests at:${NC}"
echo "https://app.virtuoso.qa/#/project/$PROJECT_ID"
