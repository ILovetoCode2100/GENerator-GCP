#!/bin/bash
# deploy-tests.sh - Main deployment script for D365 Virtuoso tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname $(dirname "$SCRIPT_DIR"))"
API_CLI="$PROJECT_ROOT/bin/api-cli"
PROCESSED_DIR="$PROJECT_ROOT/deployment/processed-tests"
STATE_DIR="$PROJECT_ROOT/deployment/state"
LOG_DIR="$PROJECT_ROOT/deployment/logs"

# Deployment state file
STATE_FILE="$STATE_DIR/deployment-state.json"
LOG_FILE="$LOG_DIR/deployment-$(date +%Y%m%d_%H%M%S).log"

# Global variables
PROJECT_ID=""
DRY_RUN=false
CONTINUE_FROM_CHECKPOINT=false

# Log functions
log() {
    local message="[$(date +'%Y-%m-%d %H:%M:%S')] $1"
    echo -e "${BLUE}${message}${NC}"
    echo "$message" >> "$LOG_FILE"
}

error() {
    local message="[ERROR] $1"
    echo -e "${RED}${message}${NC}" >&2
    echo "$message" >> "$LOG_FILE"
}

warning() {
    local message="[WARNING] $1"
    echo -e "${YELLOW}${message}${NC}"
    echo "$message" >> "$LOG_FILE"
}

success() {
    local message="[SUCCESS] $1"
    echo -e "${GREEN}${message}${NC}"
    echo "$message" >> "$LOG_FILE"
}

info() {
    local message="[INFO] $1"
    echo -e "${CYAN}${message}${NC}"
    echo "$message" >> "$LOG_FILE"
}

# Initialize deployment
init_deployment() {
    log "Initializing deployment..."

    # Create necessary directories
    mkdir -p "$STATE_DIR"
    mkdir -p "$LOG_DIR"

    # Check for existing state
    if [ -f "$STATE_FILE" ] && [ "$CONTINUE_FROM_CHECKPOINT" = true ]; then
        log "Loading existing deployment state..."
        PROJECT_ID=$(jq -r '.project_id // empty' "$STATE_FILE" 2>/dev/null || echo "")

        if [ -n "$PROJECT_ID" ]; then
            info "Continuing deployment for project: $PROJECT_ID"
        fi
    else
        # Initialize new state
        cat > "$STATE_FILE" <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "project_id": null,
  "goals": {},
  "tests_deployed": [],
  "tests_failed": [],
  "status": "initialized"
}
EOF
    fi

    success "Deployment initialized"
}

# Create or get project
setup_project() {
    log "Setting up Virtuoso project..."

    if [ -n "$PROJECT_ID" ]; then
        # Verify project exists
        if "$API_CLI" get project "$PROJECT_ID" --output json &>/dev/null; then
            success "Using existing project: $PROJECT_ID"
            return 0
        else
            warning "Project $PROJECT_ID not found, creating new project..."
            PROJECT_ID=""
        fi
    fi

    # Create new project
    local project_name="D365 Test Suite - $(date +%Y%m%d_%H%M%S)"
    local project_desc="Comprehensive D365 test automation suite covering all modules"

    if [ "$DRY_RUN" = true ]; then
        info "[DRY RUN] Would create project: $project_name"
        PROJECT_ID="dry-run-project-id"
    else
        log "Creating new project: $project_name"

        local create_output=$("$API_CLI" create project "$project_name" "$project_desc" --output json 2>&1) || {
            error "Failed to create project: $create_output"
            return 1
        }

        PROJECT_ID=$(echo "$create_output" | jq -r '.id // empty')

        if [ -z "$PROJECT_ID" ]; then
            error "Failed to get project ID from creation response"
            return 1
        fi

        success "Project created with ID: $PROJECT_ID"

        # Update state
        jq ".project_id = \"$PROJECT_ID\"" "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"
    fi
}

# Create goal for a module
create_goal() {
    local module_name="$1"
    local module_desc="$2"

    log "Creating goal for module: $module_name"

    local goal_name="${module_name} Tests"

    if [ "$DRY_RUN" = true ]; then
        info "[DRY RUN] Would create goal: $goal_name"
        echo "dry-run-goal-id-$module_name"
        return 0
    fi

    local create_output=$("$API_CLI" create goal "$PROJECT_ID" "$goal_name" "$module_desc" --output json 2>&1) || {
        error "Failed to create goal: $create_output"
        return 1
    }

    local goal_id=$(echo "$create_output" | jq -r '.id // empty')

    if [ -z "$goal_id" ]; then
        error "Failed to get goal ID from creation response"
        return 1
    fi

    success "Goal created with ID: $goal_id"

    # Update state
    jq ".goals.\"$module_name\" = \"$goal_id\"" "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"

    echo "$goal_id"
}

# Deploy a single test
deploy_test() {
    local test_file="$1"
    local goal_id="$2"
    local test_number="$3"
    local total_tests="$4"

    local test_name=$(basename "$test_file" .yaml)
    log "[$test_number/$total_tests] Deploying test: $test_name"

    if [ "$DRY_RUN" = true ]; then
        info "[DRY RUN] Would deploy: $test_file to goal $goal_id"
        return 0
    fi

    # Check if test was already deployed
    if jq -r '.tests_deployed[]' "$STATE_FILE" 2>/dev/null | grep -q "^$test_name$"; then
        info "Test already deployed, skipping: $test_name"
        return 0
    fi

    # Deploy the test using run-test command
    local deploy_output=$("$API_CLI" run-test "$test_file" --project-id "$PROJECT_ID" --goal-id "$goal_id" --output json 2>&1) || {
        error "Failed to deploy test: $deploy_output"

        # Record failure
        jq ".tests_failed += [\"$test_name\"]" "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"

        return 1
    }

    success "Test deployed successfully: $test_name"

    # Record success
    jq ".tests_deployed += [\"$test_name\"]" "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"

    return 0
}

# Deploy tests for a module
deploy_module() {
    local module_name="$1"
    local module_path="$PROCESSED_DIR/$module_name"

    log "Deploying module: $module_name"

    # Check if module directory exists
    if [ ! -d "$module_path" ]; then
        warning "Module directory not found: $module_path"
        return 0
    fi

    # Get or create goal
    local goal_id=$(jq -r ".goals.\"$module_name\" // empty" "$STATE_FILE" 2>/dev/null)

    if [ -z "$goal_id" ]; then
        # Define module descriptions
        local module_desc=""
        case "$module_name" in
            "commerce") module_desc="All commerce module tests including B2B, CX, POS, etc." ;;
            "customer-service") module_desc="Customer service module tests including cases, KB, SLA, etc." ;;
            "field-service") module_desc="Field service module tests including work orders, IoT, scheduling, etc." ;;
            "finance-operations") module_desc="Finance and operations tests including AP, AR, GL, etc." ;;
            "human-resources") module_desc="HR module tests including benefits, leave, performance, etc." ;;
            "marketing") module_desc="Marketing module tests including email, events, journeys, etc." ;;
            "project-operations") module_desc="Project operations tests including PM, resources, time tracking, etc." ;;
            "sales") module_desc="Sales module tests including leads, opportunities, orders, etc." ;;
            "supply-chain") module_desc="Supply chain tests including inventory, production, quality, etc." ;;
            *) module_desc="Tests for $module_name module" ;;
        esac

        goal_id=$(create_goal "$module_name" "$module_desc") || return 1
    else
        info "Using existing goal: $goal_id"
    fi

    # Deploy tests in the module
    local test_files=($(find "$module_path" -name "*.yaml" -type f | sort))
    local total_tests=${#test_files[@]}
    local deployed=0
    local failed=0

    info "Found $total_tests tests in module $module_name"

    for i in "${!test_files[@]}"; do
        local test_file="${test_files[$i]}"
        local test_number=$((i + 1))

        if deploy_test "$test_file" "$goal_id" "$test_number" "$total_tests"; then
            ((deployed++))
        else
            ((failed++))

            # Check max consecutive failures
            if [ $failed -ge 3 ]; then
                error "Too many consecutive failures in module $module_name"
                return 1
            fi
        fi

        # Small delay to avoid overwhelming the API
        sleep 1
    done

    log "Module $module_name deployment complete: $deployed deployed, $failed failed"
}

# Deploy all modules
deploy_all_modules() {
    log "Starting deployment of all modules..."

    local modules=(
        "commerce"
        "customer-service"
        "field-service"
        "finance-operations"
        "human-resources"
        "marketing"
        "project-operations"
        "sales"
        "supply-chain"
    )

    local total_modules=${#modules[@]}
    local completed_modules=0

    for module in "${modules[@]}"; do
        ((completed_modules++))
        echo
        log "Processing module $completed_modules/$total_modules: $module"

        if deploy_module "$module"; then
            success "Module $module deployed successfully"
        else
            error "Module $module deployment failed"

            if [ "$CONTINUE_FROM_CHECKPOINT" = false ]; then
                error "Stopping deployment due to module failure"
                return 1
            fi
        fi

        # Update state
        jq ".last_module = \"$module\" | .status = \"in_progress\"" "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"
    done

    # Update final state
    jq ".status = \"completed\" | .completed_at = \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"" "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"

    success "All modules deployed successfully!"
}

# Generate deployment report
generate_report() {
    log "Generating deployment report..."

    local report_file="$PROJECT_ROOT/deployment/reports/deployment-report-$(date +%Y%m%d_%H%M%S).md"
    mkdir -p "$(dirname "$report_file")"

    local total_deployed=$(jq '.tests_deployed | length' "$STATE_FILE")
    local total_failed=$(jq '.tests_failed | length' "$STATE_FILE")
    local total_tests=$((total_deployed + total_failed))

    cat > "$report_file" <<EOF
# D365 Virtuoso Test Deployment Report

Generated: $(date)

## Summary

- **Project ID**: $PROJECT_ID
- **Total Tests**: $total_tests
- **Successfully Deployed**: $total_deployed
- **Failed**: $total_failed
- **Success Rate**: $(echo "scale=2; $total_deployed * 100 / $total_tests" | bc)%

## Module Breakdown

| Module | Goal ID | Tests Deployed |
|--------|---------|----------------|
EOF

    # Add module details
    jq -r '.goals | to_entries[] | "\(.key)|\(.value)"' "$STATE_FILE" | while IFS='|' read -r module goal_id; do
        local module_tests=$(jq -r '.tests_deployed[]' "$STATE_FILE" | grep "^$module-" | wc -l)
        echo "| $module | $goal_id | $module_tests |" >> "$report_file"
    done

    # Add failed tests if any
    if [ $total_failed -gt 0 ]; then
        echo -e "\n## Failed Tests\n" >> "$report_file"
        jq -r '.tests_failed[]' "$STATE_FILE" | while read -r test_name; do
            echo "- $test_name" >> "$report_file"
        done
    fi

    # Add deployment log summary
    echo -e "\n## Deployment Log\n" >> "$report_file"
    echo "Full deployment log available at: $LOG_FILE" >> "$report_file"

    success "Deployment report generated: $report_file"
}

# Show usage
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Deploy D365 Virtuoso tests to the platform

Options:
    -h, --help              Show this help message
    -d, --dry-run           Perform a dry run without actual deployment
    -c, --continue          Continue from last checkpoint
    -p, --project-id ID     Use existing project ID

Environment variables required:
    D365_INSTANCE          Your D365 instance name
    VIRTUOSO_API_TOKEN     Your Virtuoso API token

Example:
    $0                     # Normal deployment
    $0 --dry-run          # Test deployment without making changes
    $0 --continue         # Resume interrupted deployment
EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -c|--continue)
                CONTINUE_FROM_CHECKPOINT=true
                shift
                ;;
            -p|--project-id)
                PROJECT_ID="$2"
                shift 2
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Main deployment function
main() {
    echo -e "${BLUE}=== D365 Virtuoso Test Deployment ===${NC}"
    echo

    # Parse arguments
    parse_args "$@"

    # Check environment
    if [ -z "${D365_INSTANCE:-}" ]; then
        error "D365_INSTANCE environment variable is not set"
        exit 1
    fi

    if [ -z "${VIRTUOSO_API_TOKEN:-}" ]; then
        error "VIRTUOSO_API_TOKEN environment variable is not set"
        exit 1
    fi

    # Check processed tests exist
    if [ ! -d "$PROCESSED_DIR" ]; then
        error "Processed tests not found. Please run preprocess-tests.sh first"
        exit 1
    fi

    if [ "$DRY_RUN" = true ]; then
        warning "Running in DRY RUN mode - no actual changes will be made"
    fi

    # Initialize deployment
    init_deployment || exit 1

    # Setup project
    setup_project || exit 1

    # Deploy all modules
    deploy_all_modules || exit 1

    # Generate report
    generate_report || exit 1

    echo
    success "Deployment completed successfully!"
    echo
    echo "Project ID: $PROJECT_ID"
    echo "View deployment report: deployment/reports/"
    echo "View deployment logs: $LOG_FILE"

    if [ "$DRY_RUN" = true ]; then
        echo
        warning "This was a DRY RUN - no actual deployment was performed"
        echo "Run without --dry-run flag to perform actual deployment"
    fi
}

# Run main function
main "$@"
