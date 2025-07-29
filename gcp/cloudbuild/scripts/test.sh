#!/bin/bash
# Comprehensive test runner for Virtuoso API CLI
# This script runs all test suites and generates reports

set -e

echo "========================================="
echo "Virtuoso API CLI - Test Suite"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test suite
run_test_suite() {
    local suite_name=$1
    local test_command=$2

    echo -e "\n${YELLOW}Running ${suite_name}...${NC}"

    if eval "${test_command}"; then
        echo -e "${GREEN}✓ ${suite_name} passed${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ ${suite_name} failed${NC}"
        ((TESTS_FAILED++))

        # Don't exit immediately - run all tests
        if [ "${FAIL_FAST}" == "true" ]; then
            exit 1
        fi
    fi
}

# Create test results directory
mkdir -p test-results

# 1. Unit Tests
run_test_suite "Unit Tests" "go test -v -race -coverprofile=test-results/coverage.out -covermode=atomic ./..."

# 2. Generate coverage report
echo -e "\n${YELLOW}Generating coverage report...${NC}"
go tool cover -html=test-results/coverage.out -o test-results/coverage.html
COVERAGE=$(go tool cover -func=test-results/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Total coverage: ${COVERAGE}%"

# 3. Integration Tests (if enabled)
if [ "${INTEGRATION_TEST}" == "true" ]; then
    # Set up test environment
    export VIRTUOSO_TEST_MODE=true
    export VIRTUOSO_API_URL="${TEST_API_URL:-https://api-app2.virtuoso.qa/api}"

    # Run integration tests
    run_test_suite "Integration Tests" "go test -v -tags=integration ./test/integration/..."

    # Test unified commands
    if [ -f "./test-commands/test-unified-commands.sh" ]; then
        run_test_suite "Unified Commands Test" "./test-commands/test-unified-commands.sh"
    fi
fi

# 4. Benchmark Tests
if [ "${RUN_BENCHMARKS}" == "true" ]; then
    run_test_suite "Benchmark Tests" "go test -bench=. -benchmem ./... > test-results/benchmark.txt"
fi

# 5. Static Analysis
echo -e "\n${YELLOW}Running static analysis...${NC}"

# gofmt check
echo "Checking code formatting..."
GOFMT_OUTPUT=$(gofmt -l -s .)
if [ -n "${GOFMT_OUTPUT}" ]; then
    echo -e "${RED}The following files need formatting:${NC}"
    echo "${GOFMT_OUTPUT}"
    ((TESTS_FAILED++))
else
    echo -e "${GREEN}✓ Code formatting check passed${NC}"
    ((TESTS_PASSED++))
fi

# go vet
run_test_suite "Go Vet" "go vet ./..."

# 6. Security Tests
if [ "${RUN_SECURITY_SCAN}" == "true" ]; then
    # Check for security vulnerabilities
    run_test_suite "Vulnerability Check" "go list -json -deps ./... | nancy sleuth"

    # Run gosec if available
    if command -v gosec &> /dev/null; then
        run_test_suite "Security Scan (gosec)" "gosec -fmt json -out test-results/gosec-report.json ./..."
    fi
fi

# 7. License Check
if [ "${CHECK_LICENSES}" == "true" ]; then
    run_test_suite "License Check" "go-licenses check ./..."
fi

# 8. API Contract Tests
if [ "${RUN_CONTRACT_TESTS}" == "true" ] && [ -d "./test/contracts" ]; then
    run_test_suite "API Contract Tests" "go test -v ./test/contracts/..."
fi

# 9. E2E Tests (minimal)
if [ "${RUN_E2E_TESTS}" == "true" ]; then
    echo -e "\n${YELLOW}Running E2E tests...${NC}"

    # Build the CLI
    go build -o test-results/api-cli-test ./cmd/api-cli

    # Run basic E2E test
    if ./test-results/api-cli-test version > /dev/null 2>&1; then
        echo -e "${GREEN}✓ CLI binary test passed${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ CLI binary test failed${NC}"
        ((TESTS_FAILED++))
    fi
fi

# 10. Generate test report
echo -e "\n${YELLOW}Generating test report...${NC}"

cat > test-results/report.json << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "total_tests": $((TESTS_PASSED + TESTS_FAILED)),
  "passed": ${TESTS_PASSED},
  "failed": ${TESTS_FAILED},
  "coverage": ${COVERAGE:-0},
  "duration_seconds": ${SECONDS},
  "environment": "${ENVIRONMENT:-unknown}",
  "build_id": "${BUILD_ID:-local}",
  "commit": "${SHORT_SHA:-$(git rev-parse --short HEAD)}"
}
EOF

# Generate JUnit XML report (for CI systems)
if command -v go-junit-report &> /dev/null; then
    go test -v ./... | go-junit-report > test-results/junit.xml
fi

# Summary
echo -e "\n========================================="
echo -e "Test Summary"
echo -e "========================================="
echo -e "Total tests run: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"
echo -e "Coverage: ${COVERAGE}%"
echo -e "Duration: ${SECONDS}s"
echo -e "========================================="

# Exit with appropriate code
if [ ${TESTS_FAILED} -gt 0 ]; then
    echo -e "${RED}TESTS FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}ALL TESTS PASSED${NC}"
    exit 0
fi
