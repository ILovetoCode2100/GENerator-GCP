#!/bin/bash

# Run D365 Test Automation Suite Tests
# This script executes tests for the D365 Test Automation Suite

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_CLI="${API_CLI:-../bin/api-cli}"
CONFIG_FILE="${CONFIG_FILE:-d365-test-config.yaml}"
PROJECT_ID_FILE="$SCRIPT_DIR/.project-id"
RESULTS_DIR="${RESULTS_DIR:-$SCRIPT_DIR/test-results}"

# Parse command line arguments
MODULE_FILTER=""
TAG_FILTER=""
JOURNEY_FILTER=""
PARALLEL_EXECUTION=false
OUTPUT_FORMAT="human"
DRY_RUN=false

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  --module MODULE      Run tests for specific module (e.g., sales, marketing)"
    echo "  --tag TAG           Run tests with specific tag"
    echo "  --journey NAME      Run specific journey by name"
    echo "  --parallel          Run tests in parallel (default: sequential)"
    echo "  --output FORMAT     Output format: human, json, yaml (default: human)"
    echo "  --dry-run          Show what would be executed without running tests"
    echo "  --project-id ID     Specify project ID (otherwise uses saved ID)"
    echo "  --help              Show this help message"
    exit 1
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --module)
            MODULE_FILTER="$2"
            shift 2
            ;;
        --tag)
            TAG_FILTER="$2"
            shift 2
            ;;
        --journey)
            JOURNEY_FILTER="$2"
            shift 2
            ;;
        --parallel)
            PARALLEL_EXECUTION=true
            shift
            ;;
        --output)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        --help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

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

# Get project ID
if [ -z "$PROJECT_ID" ]; then
    if [ -f "$PROJECT_ID_FILE" ]; then
        PROJECT_ID=$(cat "$PROJECT_ID_FILE")
    else
        echo -e "${RED}Error: No project ID found. Please run deploy-d365-tests.sh first${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}=== D365 Test Automation Suite Execution ===${NC}"
echo "Project ID: $PROJECT_ID"
echo "Test Results Directory: $RESULTS_DIR"

# Create results directory
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RUN_DIR="$RESULTS_DIR/run_$TIMESTAMP"
mkdir -p "$RUN_DIR"

# Function to run a single test journey
run_journey() {
    local goal_id=$1
    local journey_id=$2
    local journey_name=$3
    local module=$4

    echo -e "\n${YELLOW}Running: $journey_name${NC}"

    if [ "$DRY_RUN" = true ]; then
        echo "  [DRY RUN] Would execute journey $journey_id in goal $goal_id"
        return 0
    fi

    # Create checkpoint for this test run
    local checkpoint_id=$($API_CLI create checkpoint "$journey_id" --output json | jq -r '.id // empty')

    if [ -z "$checkpoint_id" ]; then
        echo -e "${RED}  Failed to create checkpoint for journey $journey_id${NC}"
        return 1
    fi

    # Execute the test
    local result_file="$RUN_DIR/${module}_${journey_name// /_}_$checkpoint_id.json"

    if $API_CLI execute checkpoint "$checkpoint_id" --output json > "$result_file" 2>&1; then
        echo -e "${GREEN}  ✓ Test completed successfully${NC}"

        # Extract key metrics from results
        if [ -f "$result_file" ] && command -v jq >/dev/null 2>&1; then
            local duration=$(jq -r '.duration // "N/A"' "$result_file")
            local status=$(jq -r '.status // "N/A"' "$result_file")
            echo "    Duration: $duration"
            echo "    Status: $status"
        fi

        return 0
    else
        echo -e "${RED}  ✗ Test failed${NC}"
        return 1
    fi
}

# Function to run tests for a specific goal
run_goal_tests() {
    local goal_id=$1
    local goal_name=$2
    local module=$3

    echo -e "\n${BLUE}Module: $module - $goal_name${NC}"

    # Get journeys for this goal
    local journeys=$($API_CLI list journeys "$goal_id" --output json)

    if [ -z "$journeys" ] || [ "$journeys" = "[]" ]; then
        echo -e "${YELLOW}  No journeys found for this goal${NC}"
        return
    fi

    # Process each journey
    echo "$journeys" | jq -c '.[]' | while read -r journey; do
        local journey_id=$(echo "$journey" | jq -r '.id')
        local journey_name=$(echo "$journey" | jq -r '.name')
        local journey_tags=$(echo "$journey" | jq -r '.tags[]?' 2>/dev/null)

        # Apply journey filter if specified
        if [ -n "$JOURNEY_FILTER" ] && [[ ! "$journey_name" =~ $JOURNEY_FILTER ]]; then
            continue
        fi

        # Apply tag filter if specified
        if [ -n "$TAG_FILTER" ]; then
            local has_tag=false
            for tag in $journey_tags; do
                if [ "$tag" = "$TAG_FILTER" ]; then
                    has_tag=true
                    break
                fi
            done
            if [ "$has_tag" = false ]; then
                continue
            fi
        fi

        # Run the journey test
        if [ "$PARALLEL_EXECUTION" = true ]; then
            run_journey "$goal_id" "$journey_id" "$journey_name" "$module" &
        else
            run_journey "$goal_id" "$journey_id" "$journey_name" "$module"
        fi
    done

    # Wait for parallel executions to complete
    if [ "$PARALLEL_EXECUTION" = true ]; then
        wait
    fi
}

# Main execution
echo -e "\n${YELLOW}Fetching project goals...${NC}"

# Get all goals in the project
goals=$($API_CLI list goals "$PROJECT_ID" --output json)

if [ -z "$goals" ] || [ "$goals" = "[]" ]; then
    echo -e "${RED}No goals found in project $PROJECT_ID${NC}"
    exit 1
fi

# Track test execution statistics
total_goals=0
total_journeys=0
executed_journeys=0
passed_journeys=0

# Process each goal (module)
echo "$goals" | jq -c '.[]' | while read -r goal; do
    goal_id=$(echo "$goal" | jq -r '.id')
    goal_name=$(echo "$goal" | jq -r '.name')

    # Extract module name from goal name (e.g., "Sales Module Tests" -> "sales")
    module_name=$(echo "$goal_name" | sed -E 's/([^ ]+) Module Tests/\1/' | tr '[:upper:]' '[:lower:]' | tr ' ' '-')

    # Apply module filter if specified
    if [ -n "$MODULE_FILTER" ] && [ "$module_name" != "$MODULE_FILTER" ]; then
        continue
    fi

    ((total_goals++))
    run_goal_tests "$goal_id" "$goal_name" "$module_name"
done

# Generate execution summary
echo -e "\n${BLUE}=== Test Execution Summary ===${NC}"

SUMMARY_FILE="$RUN_DIR/execution-summary.txt"
{
    echo "D365 Test Automation Suite - Execution Summary"
    echo "============================================="
    echo "Date: $(date)"
    echo "Project ID: $PROJECT_ID"
    echo "Run ID: $TIMESTAMP"
    echo ""
    echo "Filters Applied:"
    echo "  Module: ${MODULE_FILTER:-None}"
    echo "  Tag: ${TAG_FILTER:-None}"
    echo "  Journey: ${JOURNEY_FILTER:-None}"
    echo ""
    echo "Execution Mode: $([ "$PARALLEL_EXECUTION" = true ] && echo "Parallel" || echo "Sequential")"
    echo ""
    echo "Results saved to: $RUN_DIR"
} > "$SUMMARY_FILE"

# Display summary
cat "$SUMMARY_FILE"

# Generate HTML report if available
if command -v python3 >/dev/null 2>&1 && [ "$DRY_RUN" = false ]; then
    echo -e "\n${YELLOW}Generating HTML report...${NC}"

    # Create a simple HTML report generator script
    cat > "$RUN_DIR/generate_report.py" << 'EOF'
import json
import os
import sys
from datetime import datetime

def generate_html_report(results_dir):
    results = []
    for filename in os.listdir(results_dir):
        if filename.endswith('.json'):
            with open(os.path.join(results_dir, filename), 'r') as f:
                try:
                    data = json.load(f)
                    results.append({
                        'name': filename.replace('.json', '').replace('_', ' '),
                        'status': data.get('status', 'Unknown'),
                        'duration': data.get('duration', 'N/A')
                    })
                except:
                    pass

    html = f"""
    <!DOCTYPE html>
    <html>
    <head>
        <title>D365 Test Results - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</title>
        <style>
            body {{ font-family: Arial, sans-serif; margin: 20px; }}
            table {{ border-collapse: collapse; width: 100%; }}
            th, td {{ border: 1px solid #ddd; padding: 8px; text-align: left; }}
            th {{ background-color: #4CAF50; color: white; }}
            .passed {{ color: green; }}
            .failed {{ color: red; }}
        </style>
    </head>
    <body>
        <h1>D365 Test Automation Suite Results</h1>
        <p>Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
        <table>
            <tr>
                <th>Test Name</th>
                <th>Status</th>
                <th>Duration</th>
            </tr>
    """

    for result in results:
        status_class = 'passed' if result['status'] == 'passed' else 'failed'
        html += f"""
            <tr>
                <td>{result['name']}</td>
                <td class="{status_class}">{result['status']}</td>
                <td>{result['duration']}</td>
            </tr>
        """

    html += """
        </table>
    </body>
    </html>
    """

    with open(os.path.join(results_dir, 'report.html'), 'w') as f:
        f.write(html)

if __name__ == '__main__':
    generate_html_report(sys.argv[1])
EOF

    python3 "$RUN_DIR/generate_report.py" "$RUN_DIR"
    echo -e "${GREEN}HTML report generated: $RUN_DIR/report.html${NC}"
fi

echo -e "\n${GREEN}=== Test Execution Complete ===${NC}"
echo "Results saved to: $RUN_DIR"

# Open results directory if on macOS
if [ "$(uname)" = "Darwin" ] && [ "$DRY_RUN" = false ]; then
    open "$RUN_DIR"
fi

exit 0
