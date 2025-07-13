#!/usr/bin/env bats
# 99_report.bats - Summary and reporting tests

load 00_env

setup() {
    load 00_env
    setup
    
    # This file runs last and can generate summary reports
    export TEST_SUMMARY_FILE="$TEST_DIR/test-summary.txt"
}

@test "generate test summary report" {
    # Run the report generation script if it exists
    if [ -f "$TEST_DIR/generate_report.sh" ]; then
        # Note: We don't run all tests again here, just check if report was generated
        # The generate_report.sh should be run separately before this test suite
        if [ -f "$TEST_DIR/report.md" ]; then
            echo "Test report found at: $TEST_DIR/report.md"
            # Copy key metrics to summary file
            cat > "$TEST_SUMMARY_FILE" <<EOF
=====================================
Test Execution Summary
=====================================
Date: $(date)
Binary: $BINARY
Test Directory: $TEST_DIR

Detailed report available at: $TEST_DIR/report.md

EOF
            # Extract summary from markdown report
            if grep -A 10 "## Summary" "$TEST_DIR/report.md" >> "$TEST_SUMMARY_FILE"; then
                echo "Summary extracted from report.md"
            fi
        else
            echo "Note: Run generate_report.sh to create full test report"
            # Create basic summary
            cat > "$TEST_SUMMARY_FILE" <<EOF
=====================================
Test Execution Summary
=====================================
Date: $(date)
Binary: $BINARY
Test Directory: $TEST_DIR

Note: Run $TEST_DIR/generate_report.sh for detailed report

Test Categories:
- Environment Setup (00_env.bats)
- Authentication & Health (10_auth.bats)
- Project Operations (20_project.bats)
- Journey & Goal Operations (30_journey_goal.bats)
- Checkpoint Operations (40_checkpoint.bats)
- Step Operations (50_steps.bats)
- Output Formats (60_formats.bats)
- Session Management (70_session.bats)
- Error Handling (80_errors.bats)
- Summary Report (99_report.bats)

EOF
        fi
    fi
    
    [ -f "$TEST_SUMMARY_FILE" ]
}

@test "check test coverage metrics" {
    skip "Implement based on your coverage tools"
    # If you have coverage tools, generate coverage report
    # run coverage_tool --report
    # [ "$status" -eq 0 ]
}

@test "performance benchmarks" {
    skip "Implement performance tests"
    # Record execution times for key operations
    # start_time=$(date +%s.%N)
    # run "$BINARY" project list
    # end_time=$(date +%s.%N)
    # execution_time=$(echo "$end_time - $start_time" | bc)
    # echo "Project list execution time: $execution_time seconds" >> "$TEST_SUMMARY_FILE"
}

@test "memory usage check" {
    skip "Implement memory usage monitoring"
    # Check for memory leaks or excessive usage
    # Useful for long-running operations
}

@test "verify all test files are executable" {
    for test_file in "$TEST_DIR"/*.bats; do
        [ -x "$test_file" ] || chmod +x "$test_file"
    done
}

@test "check for skipped tests" {
    # Count skipped tests across all files
    SKIPPED_COUNT=0
    for test_file in "$TEST_DIR"/*.bats; do
        count=$(grep -c "skip " "$test_file" 2>/dev/null || echo 0)
        SKIPPED_COUNT=$((SKIPPED_COUNT + count))
    done
    
    echo "Total skipped tests: $SKIPPED_COUNT" >> "$TEST_SUMMARY_FILE"
    echo "Skipped test count: $SKIPPED_COUNT"
    
    # This test always passes but reports the count
    [ "$SKIPPED_COUNT" -ge 0 ]
}

@test "generate API compatibility report" {
    skip "Implement API version checking"
    # run "$BINARY" version --api
    # echo "API Version: $output" >> "$TEST_SUMMARY_FILE"
}

@test "cleanup test artifacts" {
    # Clean up any remaining test data
    # Be careful not to delete important data
    
    # List test projects that might need cleanup
    run "$BINARY" project list --format json
    if [ "$status" -eq 0 ]; then
        echo "$output" | jq -r '.[] | select(.name | startswith("test-") or startswith("format-test-") or startswith("session-test-") or startswith("steps-test-") or startswith("checkpoint-test-") or startswith("journey-test-") or startswith("error-test-") or startswith("dup-test-")) | .name' > "$TEST_DIR/test-projects-to-cleanup.txt" || true
    fi
    
    if [ -f "$TEST_DIR/test-projects-to-cleanup.txt" ]; then
        echo "Test projects that may need cleanup:" >> "$TEST_SUMMARY_FILE"
        cat "$TEST_DIR/test-projects-to-cleanup.txt" >> "$TEST_SUMMARY_FILE"
    fi
    
    [ 0 -eq 0 ]  # Always pass
}

@test "display test summary" {
    echo "===== TEST EXECUTION SUMMARY ====="
    
    # Display from test summary file
    if [ -f "$TEST_SUMMARY_FILE" ]; then
        cat "$TEST_SUMMARY_FILE"
    fi
    
    # Also display key metrics from markdown report if available
    if [ -f "$TEST_DIR/report.md" ]; then
        echo ""
        echo "===== AGGREGATED TEST RESULTS ====="
        # Extract the summary table from the markdown report
        sed -n '/## Summary/,/## Test Files Status/p' "$TEST_DIR/report.md" | head -n -2
        
        # Check for failed tests
        if grep -q "## Failed Tests" "$TEST_DIR/report.md"; then
            echo ""
            echo "⚠️  FAILED TESTS DETECTED ⚠️"
            sed -n '/## Failed Tests/,/## Test Categories/p' "$TEST_DIR/report.md" | head -n -2
        else
            echo ""
            echo "✅ ALL TESTS PASSED ✅"
        fi
    else
        echo ""
        echo "ℹ️  To generate detailed test report, run:"
        echo "   $TEST_DIR/generate_report.sh"
    fi
    
    echo "===================================="
    
    [ 0 -eq 0 ]  # Always pass
}
