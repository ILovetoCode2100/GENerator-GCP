#!/bin/bash

# Virtuoso API CLI Test Suite Runner
# This script runs all test YAML files and reports results

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test file
run_test() {
    local test_file=$1
    local expect_failure=$2
    local test_name=$3

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo -e "\n${YELLOW}Running: ${test_name}${NC}"
    echo "File: ${test_file}"
    echo "Expected: ${expect_failure}"
    echo "----------------------------------------"

    # Run the test and capture the exit code
    if ./bin/api-cli run-test "${test_file}" --dry-run > /tmp/test_output.log 2>&1; then
        exit_code=0
    else
        exit_code=$?
    fi

    # Check if the result matches expectations
    if [ "${expect_failure}" = "should_fail" ]; then
        if [ $exit_code -ne 0 ]; then
            echo -e "${GREEN}✓ Test correctly failed with exit code ${exit_code}${NC}"
            # Check for error messages in output
            if grep -q -i "error" /tmp/test_output.log; then
                echo -e "${GREEN}✓ Error messages found in output${NC}"
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                echo -e "${RED}✗ No error messages found in output${NC}"
                FAILED_TESTS=$((FAILED_TESTS + 1))
            fi
        else
            echo -e "${RED}✗ Test should have failed but passed${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        if [ $exit_code -eq 0 ]; then
            echo -e "${GREEN}✓ Test passed successfully${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}✗ Test failed with exit code ${exit_code}${NC}"
            echo "Error output:"
            tail -n 20 /tmp/test_output.log
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    fi
}

# Function to run YAML validation tests
run_yaml_validation() {
    local test_file="test-yaml-validation.yaml"

    echo -e "\n${YELLOW}Running: YAML Validation Tests${NC}"
    echo "File: ${test_file}"
    echo "Note: This file contains multiple invalid YAML sections"
    echo "----------------------------------------"

    # Extract and test each YAML document separately
    local doc_count=0
    local valid_count=0
    local invalid_count=0

    # Split the file by '---' markers and test each section
    awk '/^---/{n++}{print > "yaml_test_"n".yaml"}' "${test_file}"

    for yaml_file in yaml_test_*.yaml; do
        if [ -f "$yaml_file" ]; then
            doc_count=$((doc_count + 1))
            echo -e "\nTesting YAML document #${doc_count}..."

            if ./bin/api-cli run-test "${yaml_file}" --dry-run > /tmp/yaml_output.log 2>&1; then
                valid_count=$((valid_count + 1))
                echo -e "${GREEN}✓ Valid YAML structure${NC}"
            else
                invalid_count=$((invalid_count + 1))
                echo -e "${YELLOW}✓ Invalid YAML detected (expected)${NC}"
                grep -i "error" /tmp/yaml_output.log | head -n 3
            fi

            rm -f "$yaml_file"
        fi
    done

    echo -e "\nYAML Validation Summary:"
    echo "Total documents: ${doc_count}"
    echo "Valid documents: ${valid_count}"
    echo "Invalid documents: ${invalid_count} (most are intentionally invalid)"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if [ $invalid_count -gt 0 ]; then
        echo -e "${GREEN}✓ YAML validation correctly detected invalid structures${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ YAML validation did not detect expected errors${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Main test execution
echo "======================================"
echo "Virtuoso API CLI Test Suite"
echo "======================================"
echo "Date: $(date)"
echo "Working Directory: $(pwd)"

# Check if api-cli binary exists
if [ ! -f "./bin/api-cli" ]; then
    echo -e "${RED}Error: api-cli binary not found at ./bin/api-cli${NC}"
    echo "Please build the project first with: make build"
    exit 1
fi

# Run the tests
run_test "test-all-commands-positive.yaml" "should_pass" "Positive Test Cases (All 69 Commands)"
run_test "test-negative-cases.yaml" "should_fail" "Negative Test Cases (Error Handling)"
run_test "test-edge-cases.yaml" "should_pass" "Edge Cases and Boundary Conditions"
run_yaml_validation

# Summary
echo -e "\n======================================"
echo "Test Suite Summary"
echo "======================================"
echo "Total Tests Run: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests completed as expected!${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Some tests did not behave as expected${NC}"
    exit 1
fi
