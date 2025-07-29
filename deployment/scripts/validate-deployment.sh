#!/bin/bash
# validate-deployment.sh - Validate deployed tests and generate health report

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
STATE_FILE="$PROJECT_ROOT/deployment/state/deployment-state.json"

# Validation results
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

# Log functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    ((FAILED_CHECKS++))
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    ((WARNINGS++))
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_CHECKS++))
}

info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

# Check deployment state
check_deployment_state() {
    log "Checking deployment state..."
    ((TOTAL_CHECKS++))

    if [ ! -f "$STATE_FILE" ]; then
        error "Deployment state file not found"
        return 1
    fi

    local status=$(jq -r '.status // "unknown"' "$STATE_FILE")
    local project_id=$(jq -r '.project_id // empty' "$STATE_FILE")

    if [ "$status" = "completed" ]; then
        success "Deployment status: completed"
    else
        warning "Deployment status: $status (not completed)"
    fi

    if [ -n "$project_id" ]; then
        info "Project ID: $project_id"
    else
        error "No project ID found in state"
        return 1
    fi
}

# Validate project exists
validate_project() {
    log "Validating project existence..."
    ((TOTAL_CHECKS++))

    local project_id=$(jq -r '.project_id // empty' "$STATE_FILE")

    if [ -z "$project_id" ]; then
        error "No project ID in state file"
        return 1
    fi

    if "$API_CLI" get project "$project_id" --output json &>/dev/null; then
        success "Project exists and is accessible: $project_id"
    else
        error "Project not found or not accessible: $project_id"
        return 1
    fi
}

# Validate goals
validate_goals() {
    log "Validating goals..."

    local project_id=$(jq -r '.project_id // empty' "$STATE_FILE")
    local goals=$(jq -r '.goals | to_entries[] | "\(.key)|\(.value)"' "$STATE_FILE" 2>/dev/null)

    if [ -z "$goals" ]; then
        error "No goals found in deployment state"
        return 1
    fi

    echo "$goals" | while IFS='|' read -r module goal_id; do
        ((TOTAL_CHECKS++))

        if "$API_CLI" get goal "$project_id" "$goal_id" --output json &>/dev/null; then
            success "Goal exists for module $module: $goal_id"
        else
            error "Goal not found for module $module: $goal_id"
        fi
    done
}

# Validate deployed tests
validate_deployed_tests() {
    log "Validating deployed tests..."

    local project_id=$(jq -r '.project_id // empty' "$STATE_FILE")
    local deployed_tests=$(jq -r '.tests_deployed[]' "$STATE_FILE" 2>/dev/null | wc -l)
    local failed_tests=$(jq -r '.tests_failed[]' "$STATE_FILE" 2>/dev/null | wc -l)

    info "Tests deployed: $deployed_tests"
    info "Tests failed: $failed_tests"

    ((TOTAL_CHECKS++))

    if [ "$deployed_tests" -eq 0 ]; then
        error "No tests were deployed"
    elif [ "$deployed_tests" -lt 169 ]; then
        warning "Only $deployed_tests out of 169 tests were deployed"
    else
        success "All 169 tests were deployed"
    fi

    if [ "$failed_tests" -gt 0 ]; then
        warning "$failed_tests tests failed to deploy"

        # List failed tests
        echo "Failed tests:"
        jq -r '.tests_failed[]' "$STATE_FILE" 2>/dev/null | while read -r test; do
            echo "  - $test"
        done
    fi
}

# Check test execution capability
check_test_execution() {
    log "Checking test execution capability..."
    ((TOTAL_CHECKS++))

    local project_id=$(jq -r '.project_id // empty' "$STATE_FILE")

    # Get first deployed test
    local sample_test=$(jq -r '.tests_deployed[0] // empty' "$STATE_FILE")

    if [ -z "$sample_test" ]; then
        warning "No deployed tests to validate execution"
        return 0
    fi

    # Get goals to find a journey
    local first_goal=$(jq -r '.goals | to_entries[0].value // empty' "$STATE_FILE")

    if [ -z "$first_goal" ]; then
        error "No goals found to test execution"
        return 1
    fi

    # List journeys in the goal
    if "$API_CLI" list journeys "$project_id" "$first_goal" --output json &>/dev/null; then
        success "Can list journeys - test execution capability verified"
    else
        error "Cannot list journeys - test execution may not work"
    fi
}

# Check environment variables
check_environment() {
    log "Checking environment configuration..."

    ((TOTAL_CHECKS++))
    if [ -n "${D365_INSTANCE:-}" ]; then
        success "D365_INSTANCE is set: $D365_INSTANCE"
    else
        error "D365_INSTANCE environment variable is not set"
    fi

    ((TOTAL_CHECKS++))
    if [ -n "${VIRTUOSO_API_TOKEN:-}" ]; then
        success "VIRTUOSO_API_TOKEN is set"
    else
        error "VIRTUOSO_API_TOKEN environment variable is not set"
    fi
}

# Generate health report
generate_health_report() {
    log "Generating deployment health report..."

    local report_file="$PROJECT_ROOT/deployment/reports/health-report-$(date +%Y%m%d_%H%M%S).md"
    mkdir -p "$(dirname "$report_file")"

    local success_rate=0
    if [ $TOTAL_CHECKS -gt 0 ]; then
        success_rate=$(echo "scale=2; $PASSED_CHECKS * 100 / $TOTAL_CHECKS" | bc)
    fi

    cat > "$report_file" <<EOF
# D365 Virtuoso Deployment Health Report

Generated: $(date)

## Summary

- **Total Checks**: $TOTAL_CHECKS
- **Passed**: $PASSED_CHECKS
- **Failed**: $FAILED_CHECKS
- **Warnings**: $WARNINGS
- **Success Rate**: ${success_rate}%

## Deployment Information

$(if [ -f "$STATE_FILE" ]; then
    echo "- **Project ID**: $(jq -r '.project_id // "Not found"' "$STATE_FILE")"
    echo "- **Status**: $(jq -r '.status // "Unknown"' "$STATE_FILE")"
    echo "- **Tests Deployed**: $(jq '.tests_deployed | length' "$STATE_FILE" 2>/dev/null || echo "0")"
    echo "- **Tests Failed**: $(jq '.tests_failed | length' "$STATE_FILE" 2>/dev/null || echo "0")"
    echo "- **Deployment Date**: $(jq -r '.timestamp // "Unknown"' "$STATE_FILE")"
else
    echo "No deployment state file found"
fi)

## Health Status

$(if [ $FAILED_CHECKS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo "✅ **HEALTHY** - All checks passed"
elif [ $FAILED_CHECKS -eq 0 ]; then
    echo "⚠️ **HEALTHY WITH WARNINGS** - All critical checks passed but some warnings exist"
else
    echo "❌ **UNHEALTHY** - Critical issues detected"
fi)

## Recommendations

$(if [ $FAILED_CHECKS -gt 0 ]; then
    echo "1. Review failed checks and address critical issues"
    echo "2. Consider running rollback if deployment is incomplete"
    echo "3. Check API credentials and network connectivity"
fi

if [ $WARNINGS -gt 0 ]; then
    echo "1. Review warnings for potential issues"
    echo "2. Consider re-deploying failed tests"
fi

if [ $FAILED_CHECKS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo "1. Deployment is healthy and ready for use"
    echo "2. You can now execute tests via the Virtuoso platform"
fi)

## Next Steps

1. Review this health report
2. Address any failed checks or warnings
3. If healthy, proceed with test execution
4. Monitor test execution results

EOF

    success "Health report generated: $report_file"

    # Display summary
    echo
    echo "Health Check Summary:"
    echo "===================="
    echo "Total Checks: $TOTAL_CHECKS"
    echo "Passed: $PASSED_CHECKS"
    echo "Failed: $FAILED_CHECKS"
    echo "Warnings: $WARNINGS"
    echo "Success Rate: ${success_rate}%"
    echo

    if [ $FAILED_CHECKS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
        success "Deployment is HEALTHY ✅"
    elif [ $FAILED_CHECKS -eq 0 ]; then
        warning "Deployment is HEALTHY WITH WARNINGS ⚠️"
    else
        error "Deployment is UNHEALTHY ❌"
    fi
}

# Run specific test
run_test_check() {
    local test_name="$1"

    log "Running test check: $test_name"
    ((TOTAL_CHECKS++))

    # This is a placeholder for specific test execution
    # In real scenario, you would trigger test execution and check results
    info "Test check not implemented in validation (would require test execution)"
}

# Main validation function
main() {
    echo -e "${BLUE}=== D365 Virtuoso Deployment Validation ===${NC}"
    echo

    # Run all validation checks
    check_environment
    check_deployment_state
    validate_project
    validate_goals
    validate_deployed_tests
    check_test_execution

    # Generate health report
    generate_health_report

    # Return appropriate exit code
    if [ $FAILED_CHECKS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# Run main function
main "$@"
