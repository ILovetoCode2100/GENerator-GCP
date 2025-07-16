#!/bin/bash
# Comprehensive test runner for Virtuoso MCP Server
# Executes all test suites and generates a comprehensive report

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Configuration
REPORT_DIR="test-reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORT_DIR/test-report-$TIMESTAMP.md"
SUMMARY_FILE="$REPORT_DIR/test-summary-$TIMESTAMP.json"
EXIT_CODE=0

# Create report directory
mkdir -p "$REPORT_DIR"

# Helper functions
print_header() {
    echo -e "\n${CYAN}${BOLD}=================================================================================${NC}"
    echo -e "${CYAN}${BOLD}$1${NC}"
    echo -e "${CYAN}${BOLD}=================================================================================${NC}\n"
}

print_section() {
    echo -e "\n${BLUE}${BOLD}► $1${NC}"
    echo -e "${BLUE}---------------------------------------------------------------------------------${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Initialize report
init_report() {
    cat > "$REPORT_FILE" << EOF
# Virtuoso MCP Server Test Report

**Generated**: $(date)
**Environment**: $(node -v), npm $(npm -v)
**Branch**: $(git branch --show-current 2>/dev/null || echo "unknown")
**Commit**: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

## Test Summary

| Test Suite | Status | Duration | Tests | Passed | Failed | Coverage |
|------------|--------|----------|-------|--------|--------|----------|
EOF
}

# Check prerequisites
check_prerequisites() {
    print_section "Checking Prerequisites"

    # Check Node.js
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed"
        exit 1
    fi
    print_success "Node.js: $(node -v)"

    # Check npm
    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed"
        exit 1
    fi
    print_success "npm: $(npm -v)"

    # Check TypeScript
    if ! npx tsc --version &> /dev/null; then
        print_warning "TypeScript not found globally, will use local version"
    else
        print_success "TypeScript: $(npx tsc --version)"
    fi

    # Check for config file
    CONFIG_FILE="$HOME/.api-cli/virtuoso-config.yaml"
    if [ ! -f "$CONFIG_FILE" ]; then
        print_warning "Config file not found at $CONFIG_FILE"
        print_info "Integration tests may fail without proper configuration"
    else
        print_success "Config file found"
    fi
}

# Install dependencies
install_dependencies() {
    print_section "Installing Dependencies"

    if [ ! -d "node_modules" ]; then
        print_info "Installing dependencies..."
        npm install --silent
        print_success "Dependencies installed"
    else
        print_success "Dependencies already installed"
    fi
}

# Run linting
run_lint() {
    print_section "Running TypeScript Linting"

    START_TIME=$(date +%s)

    if npm run lint > /dev/null 2>&1; then
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_success "TypeScript compilation check passed (${DURATION}s)"
        echo "| TypeScript Lint | ✅ Pass | ${DURATION}s | - | - | - | - |" >> "$REPORT_FILE"

        echo -e "\n### TypeScript Linting\n\n✅ No TypeScript errors found\n" >> "$REPORT_FILE"
    else
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_error "TypeScript compilation errors found"
        echo "| TypeScript Lint | ❌ Fail | ${DURATION}s | - | - | - | - |" >> "$REPORT_FILE"

        echo -e "\n### TypeScript Linting\n\n❌ TypeScript errors found:\n\n\`\`\`" >> "$REPORT_FILE"
        npm run lint 2>&1 | tail -20 >> "$REPORT_FILE"
        echo -e "\`\`\`\n" >> "$REPORT_FILE"

        EXIT_CODE=1
    fi
}

# Build the project
run_build() {
    print_section "Building Project"

    START_TIME=$(date +%s)

    if npm run build > /dev/null 2>&1; then
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_success "Build completed successfully (${DURATION}s)"
        echo "| Build | ✅ Pass | ${DURATION}s | - | - | - | - |" >> "$REPORT_FILE"

        echo -e "\n### Build\n\n✅ Project built successfully\n" >> "$REPORT_FILE"
    else
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_error "Build failed"
        echo "| Build | ❌ Fail | ${DURATION}s | - | - | - | - |" >> "$REPORT_FILE"

        echo -e "\n### Build\n\n❌ Build errors:\n\n\`\`\`" >> "$REPORT_FILE"
        npm run build 2>&1 | tail -20 >> "$REPORT_FILE"
        echo -e "\`\`\`\n" >> "$REPORT_FILE"

        EXIT_CODE=1
        return 1
    fi
}

# Run unit tests
run_unit_tests() {
    print_section "Running Unit Tests"

    START_TIME=$(date +%s)

    # Create temporary file for test output
    TEST_OUTPUT=$(mktemp)

    # Run tests with coverage
    if NODE_OPTIONS='--experimental-vm-modules' npx jest --ci --coverage --json --outputFile="$TEST_OUTPUT" 2>&1 | tee test-output.log; then
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))

        # Parse test results
        if [ -f "$TEST_OUTPUT" ]; then
            TOTAL_TESTS=$(jq -r '.numTotalTests' "$TEST_OUTPUT" 2>/dev/null || echo "0")
            PASSED_TESTS=$(jq -r '.numPassedTests' "$TEST_OUTPUT" 2>/dev/null || echo "0")
            FAILED_TESTS=$(jq -r '.numFailedTests' "$TEST_OUTPUT" 2>/dev/null || echo "0")

            # Get coverage data
            COVERAGE_LINE=$(jq -r '.coverageMap | to_entries | map(.value.data.s | to_entries | map(.value) | add) | add / length | floor' "$TEST_OUTPUT" 2>/dev/null || echo "0")

            print_success "Unit tests passed: $PASSED_TESTS/$TOTAL_TESTS (${DURATION}s)"
            echo "| Unit Tests | ✅ Pass | ${DURATION}s | $TOTAL_TESTS | $PASSED_TESTS | $FAILED_TESTS | ${COVERAGE_LINE}% |" >> "$REPORT_FILE"

            echo -e "\n### Unit Tests\n\n✅ All unit tests passed\n" >> "$REPORT_FILE"
            echo -e "- Total: $TOTAL_TESTS" >> "$REPORT_FILE"
            echo -e "- Passed: $PASSED_TESTS" >> "$REPORT_FILE"
            echo -e "- Failed: $FAILED_TESTS" >> "$REPORT_FILE"
            echo -e "- Coverage: ${COVERAGE_LINE}%\n" >> "$REPORT_FILE"
        else
            print_success "Unit tests passed (${DURATION}s)"
            echo "| Unit Tests | ✅ Pass | ${DURATION}s | - | - | - | - |" >> "$REPORT_FILE"
        fi
    else
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_error "Unit tests failed"
        echo "| Unit Tests | ❌ Fail | ${DURATION}s | - | - | - | - |" >> "$REPORT_FILE"

        echo -e "\n### Unit Tests\n\n❌ Unit test failures:\n\n\`\`\`" >> "$REPORT_FILE"
        tail -50 test-output.log >> "$REPORT_FILE"
        echo -e "\`\`\`\n" >> "$REPORT_FILE"

        EXIT_CODE=1
    fi

    # Cleanup
    rm -f "$TEST_OUTPUT" test-output.log
}

# Run tool validation
run_tool_validation() {
    print_section "Running Tool Validation"

    START_TIME=$(date +%s)

    if npm run validate > /dev/null 2>&1; then
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_success "Tool validation passed (${DURATION}s)"
        echo "| Tool Validation | ✅ Pass | ${DURATION}s | 12 | 12 | 0 | - |" >> "$REPORT_FILE"

        echo -e "\n### Tool Validation\n\n✅ All 12 tool groups validated successfully\n" >> "$REPORT_FILE"
    else
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_error "Tool validation failed"
        echo "| Tool Validation | ❌ Fail | ${DURATION}s | 12 | - | - | - |" >> "$REPORT_FILE"

        echo -e "\n### Tool Validation\n\n❌ Tool validation errors:\n\n\`\`\`" >> "$REPORT_FILE"
        npm run validate 2>&1 | tail -20 >> "$REPORT_FILE"
        echo -e "\`\`\`\n" >> "$REPORT_FILE"

        EXIT_CODE=1
    fi
}

# Run server integration test
run_server_test() {
    print_section "Running Server Integration Test"

    START_TIME=$(date +%s)

    if npm run test:server > /dev/null 2>&1; then
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_success "Server integration test passed (${DURATION}s)"
        echo "| Server Test | ✅ Pass | ${DURATION}s | 1 | 1 | 0 | - |" >> "$REPORT_FILE"

        echo -e "\n### Server Integration Test\n\n✅ MCP server responds correctly to protocol messages\n" >> "$REPORT_FILE"
    else
        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))
        print_error "Server integration test failed"
        echo "| Server Test | ❌ Fail | ${DURATION}s | 1 | 0 | 1 | - |" >> "$REPORT_FILE"

        echo -e "\n### Server Integration Test\n\n❌ Server test errors:\n\n\`\`\`" >> "$REPORT_FILE"
        npm run test:server 2>&1 | tail -20 >> "$REPORT_FILE"
        echo -e "\`\`\`\n" >> "$REPORT_FILE"

        EXIT_CODE=1
    fi
}

# Generate coverage report
generate_coverage_report() {
    print_section "Coverage Report"

    if [ -d "coverage" ]; then
        print_info "Coverage report generated at: coverage/lcov-report/index.html"

        echo -e "\n## Coverage Details\n" >> "$REPORT_FILE"

        # Extract coverage summary from coverage-summary.json if available
        if [ -f "coverage/coverage-summary.json" ]; then
            echo -e "### Coverage Summary\n" >> "$REPORT_FILE"
            echo -e "\`\`\`json" >> "$REPORT_FILE"
            jq '.total' coverage/coverage-summary.json >> "$REPORT_FILE" 2>/dev/null || echo "Coverage data not available" >> "$REPORT_FILE"
            echo -e "\`\`\`\n" >> "$REPORT_FILE"
        fi
    else
        print_warning "No coverage report generated"
    fi
}

# Generate summary JSON
generate_summary() {
    cat > "$SUMMARY_FILE" << EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "environment": {
    "node": "$(node -v)",
    "npm": "$(npm -v)",
    "platform": "$(uname -s)",
    "arch": "$(uname -m)"
  },
  "git": {
    "branch": "$(git branch --show-current 2>/dev/null || echo "unknown")",
    "commit": "$(git rev-parse HEAD 2>/dev/null || echo "unknown")",
    "shortCommit": "$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
  },
  "results": {
    "overall": $([ $EXIT_CODE -eq 0 ] && echo "true" || echo "false"),
    "exitCode": $EXIT_CODE
  }
}
EOF
}

# Main execution
main() {
    print_header "Virtuoso MCP Server - Comprehensive Test Suite"

    # Initialize report
    init_report

    # Run all test phases
    check_prerequisites
    install_dependencies
    run_lint

    # Only continue if build succeeds
    if run_build; then
        run_unit_tests
        run_tool_validation
        run_server_test
        generate_coverage_report
    fi

    # Generate final summary
    generate_summary

    # Print final results
    print_header "Test Results Summary"

    if [ $EXIT_CODE -eq 0 ]; then
        print_success "All tests passed! 🎉"
        echo -e "\n## Overall Result\n\n✅ **All tests passed successfully!**\n" >> "$REPORT_FILE"
    else
        print_error "Some tests failed"
        echo -e "\n## Overall Result\n\n❌ **Some tests failed. Please check the details above.**\n" >> "$REPORT_FILE"
    fi

    # Print report location
    print_info "Full test report: $REPORT_FILE"
    print_info "Summary JSON: $SUMMARY_FILE"

    # Show report preview
    echo -e "\n${MAGENTA}${BOLD}Report Preview:${NC}"
    echo -e "${MAGENTA}---------------------------------------------------------------------------------${NC}"
    head -20 "$REPORT_FILE"
    echo -e "${MAGENTA}...${NC}"

    # Exit with appropriate code
    exit $EXIT_CODE
}

# Handle interrupts
trap 'echo -e "\n${RED}Test run interrupted${NC}"; exit 130' INT TERM

# Run main function
main
