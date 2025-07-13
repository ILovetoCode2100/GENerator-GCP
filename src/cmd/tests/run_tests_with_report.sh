#!/bin/bash
# run_tests_with_report.sh - Run all BATS tests and generate report for CI

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "========================================="
echo "Running API CLI Test Suite with Reporting"
echo "========================================="
echo

# First, run the report generation script which runs all tests
echo "Executing all BATS tests and generating report..."
if "$SCRIPT_DIR/generate_report.sh"; then
    echo
    echo "✅ All tests passed!"
else
    echo
    echo "❌ Some tests failed!"
fi

# Now run the 99_report.bats to display the aggregated summary
echo
echo "Running summary report test..."
if command -v bats &> /dev/null; then
    bats "$SCRIPT_DIR/99_report.bats"
else
    echo "Warning: BATS not found, cannot run summary test"
fi

# Display the markdown report location
echo
echo "========================================="
echo "Full test report available at:"
echo "$SCRIPT_DIR/report.md"
echo "========================================="

# Exit with the same code as generate_report.sh
if [ -f "$SCRIPT_DIR/report.md" ]; then
    if grep -q "## Failed Tests" "$SCRIPT_DIR/report.md"; then
        exit 1
    fi
fi

exit 0
