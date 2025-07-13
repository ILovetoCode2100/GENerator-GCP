#!/bin/bash
# generate_report.sh - Run BATS tests and generate test report

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"
REPORT_FILE="$SCRIPT_DIR/report.md"

# Check if bats is available
if ! command -v bats &> /dev/null; then
    echo "Error: BATS not found. Please install BATS (Bash Automated Testing System)"
    echo "Install with: npm install -g bats"
    exit 1
fi

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0
FAILED_DETAILS=()

# Run each test file and collect results
echo "Running BATS tests..."
echo

# Array to store test results
declare -A TEST_RESULTS

# Get all test files sorted
TEST_FILES=$(find "$SCRIPT_DIR" -name "*.bats" -type f | sort)

# Run tests and capture TAP output
for test_file in $TEST_FILES; do
    test_name=$(basename "$test_file")
    echo "Running $test_name..."
    
    # Run bats with TAP output
    if bats --tap "$test_file" > "$SCRIPT_DIR/tap_output_$test_name.txt" 2>&1; then
        TEST_RESULTS["$test_name"]="PASSED"
    else
        TEST_RESULTS["$test_name"]="FAILED"
    fi
    
    # Parse TAP output
    while IFS= read -r line; do
        if [[ "$line" =~ ^1\.\.([0-9]+) ]]; then
            # Test count line
            test_count="${BASH_REMATCH[1]}"
            TOTAL_TESTS=$((TOTAL_TESTS + test_count))
        elif [[ "$line" =~ ^ok\ [0-9]+\ (.*)$ ]]; then
            # Passed test
            PASSED_TESTS=$((PASSED_TESTS + 1))
        elif [[ "$line" =~ ^not\ ok\ [0-9]+\ (.*)$ ]]; then
            # Failed test
            FAILED_TESTS=$((FAILED_TESTS + 1))
            test_desc="${BASH_REMATCH[1]}"
            FAILED_DETAILS+=("$test_name: $test_desc")
        elif [[ "$line" =~ ^ok\ [0-9]+\ (.*)\ #\ skip ]]; then
            # Skipped test
            SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
            PASSED_TESTS=$((PASSED_TESTS - 1)) # Adjust since it was counted as passed
        fi
    done < "$SCRIPT_DIR/tap_output_$test_name.txt"
done

# Generate Markdown report
cat > "$REPORT_FILE" <<EOF
# Test Execution Report

**Generated on:** $(date)

## Summary

| Metric | Count |
|--------|-------|
| **Total Tests** | $TOTAL_TESTS |
| **Passed** | $PASSED_TESTS |
| **Failed** | $FAILED_TESTS |
| **Skipped** | $SKIPPED_TESTS |
| **Pass Rate** | $(echo "scale=2; $PASSED_TESTS * 100 / ($TOTAL_TESTS - $SKIPPED_TESTS)" | bc 2>/dev/null || echo "N/A")% |

## Test Files Status

| Test File | Status |
|-----------|--------|
EOF

# Add status for each test file
for test_file in $TEST_FILES; do
    test_name=$(basename "$test_file")
    status="${TEST_RESULTS[$test_name]}"
    if [[ "$status" == "PASSED" ]]; then
        echo "| $test_name | ✅ PASSED |" >> "$REPORT_FILE"
    else
        echo "| $test_name | ❌ FAILED |" >> "$REPORT_FILE"
    fi
done

# Add failed test details if any
if [ ${#FAILED_DETAILS[@]} -gt 0 ]; then
    cat >> "$REPORT_FILE" <<EOF

## Failed Tests

The following tests failed:

EOF
    for failed_test in "${FAILED_DETAILS[@]}"; do
        echo "- $failed_test" >> "$REPORT_FILE"
    done
fi

# Add test categories
cat >> "$REPORT_FILE" <<EOF

## Test Categories

1. **Environment Setup** (00_env.bats) - Verifies binary existence and basic functionality
2. **Authentication & Health** (10_auth.bats) - Tests authentication and health check endpoints
3. **Project Operations** (20_project.bats) - Tests project CRUD operations
4. **Journey & Goal Operations** (30_journey_goal.bats) - Tests journey and goal management
5. **Checkpoint Operations** (40_checkpoint.bats) - Tests checkpoint functionality
6. **Step Operations** (50_steps.bats) - Tests step management features
7. **Output Formats** (60_formats.bats) - Tests various output format options
8. **Session Management** (70_session.bats) - Tests session handling
9. **Error Handling** (80_errors.bats) - Tests error scenarios and handling
10. **Summary Report** (99_report.bats) - Generates test summary and cleanup

## Test Environment

- **Binary Path:** $PROJECT_ROOT/bin/api-cli
- **Test Directory:** $SCRIPT_DIR
- **Project Root:** $PROJECT_ROOT
EOF

# Clean up temporary TAP output files
rm -f "$SCRIPT_DIR"/tap_output_*.txt

echo
echo "Test report generated: $REPORT_FILE"
echo
echo "Summary: Total=$TOTAL_TESTS, Passed=$PASSED_TESTS, Failed=$FAILED_TESTS, Skipped=$SKIPPED_TESTS"

# Exit with failure if any tests failed
if [ $FAILED_TESTS -gt 0 ]; then
    exit 1
fi

exit 0
