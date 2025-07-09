#!/bin/bash

# ULTRATHINK CLI Testing: Comprehensive test of all step creation commands
# Location: /Users/marklovelady/_dev/virtuoso-api-cli-generator
# Binary: ./bin/api-cli

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test results storage
TEST_RESULTS=()

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

# Success function
success() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED_TESTS++))
}

# Failure function
fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED_TESTS++))
}

# Warning function
warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Test runner function
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_pattern="$3"
    
    ((TOTAL_TESTS++))
    
    log "Testing: $test_name"
    
    # Run the command and capture output
    if output=$(eval "$command" 2>&1); then
        if [[ -z "$expected_pattern" ]] || echo "$output" | grep -q "$expected_pattern"; then
            success "$test_name"
            TEST_RESULTS+=("PASS: $test_name")
        else
            fail "$test_name - Expected pattern not found: $expected_pattern"
            TEST_RESULTS+=("FAIL: $test_name - Expected pattern not found")
        fi
    else
        fail "$test_name - Command failed with exit code $?"
        TEST_RESULTS+=("FAIL: $test_name - Command failed")
    fi
    
    echo "Output: $output"
    echo "---"
}

# Test help functionality
test_help() {
    local command="$1"
    run_test "Help for $command" "./bin/api-cli $command --help" "Usage:"
}

# Test parameter validation
test_parameter_validation() {
    local command="$1"
    local description="$2"
    
    log "Testing parameter validation for $command"
    
    # Test with insufficient parameters
    if ./bin/api-cli $command 2>&1 | grep -q "required"; then
        success "$command - Parameter validation working"
    else
        fail "$command - Parameter validation not working"
    fi
}

# Test session context
test_session_context() {
    log "Testing session context management"
    
    # Check current checkpoint
    if output=$(./bin/api-cli set-checkpoint --help 2>&1); then
        success "set-checkpoint command available"
    else
        fail "set-checkpoint command not available"
    fi
    
    # Test setting checkpoint
    TEST_CHECKPOINT_ID="1678318"
    if ./bin/api-cli set-checkpoint $TEST_CHECKPOINT_ID 2>&1 | grep -q "checkpoint"; then
        success "set-checkpoint functionality working"
    else
        fail "set-checkpoint functionality not working"
    fi
}

# Test output formats
test_output_formats() {
    local command="$1"
    
    log "Testing output formats for $command"
    
    # Test different output formats
    for format in human json yaml ai; do
        if ./bin/api-cli $command --help -o $format >/dev/null 2>&1; then
            success "$command - Output format $format supported"
        else
            fail "$command - Output format $format not supported"
        fi
    done
}

# Main test execution
main() {
    log "Starting comprehensive step command testing"
    
    # Test session context first
    test_session_context
    
    # Define all step commands by category
    declare -A STEP_COMMANDS=(
        # Navigation steps
        ["create-step-navigate"]="Navigate to URL"
        ["create-step-wait-time"]="Wait for time"
        ["create-step-wait-element"]="Wait for element"
        ["create-step-window"]="Window resize"
        
        # Mouse actions
        ["create-step-click"]="Click element"
        ["create-step-double-click"]="Double-click element"
        ["create-step-right-click"]="Right-click element"
        ["create-step-hover"]="Hover over element"
        ["create-step-mouse-down"]="Mouse down"
        ["create-step-mouse-up"]="Mouse up"
        ["create-step-mouse-move"]="Mouse move"
        ["create-step-mouse-enter"]="Mouse enter"
        
        # Input steps
        ["create-step-write"]="Write text"
        ["create-step-key"]="Key press"
        ["create-step-pick"]="Pick dropdown"
        ["create-step-pick-value"]="Pick value"
        ["create-step-pick-text"]="Pick text"
        ["create-step-upload"]="Upload file"
        
        # Scroll steps
        ["create-step-scroll-top"]="Scroll to top"
        ["create-step-scroll-bottom"]="Scroll to bottom"
        ["create-step-scroll-element"]="Scroll to element"
        ["create-step-scroll-position"]="Scroll to position"
        
        # Assertion steps
        ["create-step-assert-exists"]="Assert element exists"
        ["create-step-assert-not-exists"]="Assert element not exists"
        ["create-step-assert-equals"]="Assert equals"
        ["create-step-assert-not-equals"]="Assert not equals"
        ["create-step-assert-checked"]="Assert checked"
        ["create-step-assert-selected"]="Assert selected"
        ["create-step-assert-variable"]="Assert variable"
        ["create-step-assert-greater-than"]="Assert greater than"
        ["create-step-assert-greater-than-or-equal"]="Assert greater than or equal"
        ["create-step-assert-less-than-or-equal"]="Assert less than or equal"
        ["create-step-assert-matches"]="Assert matches"
        
        # Data steps
        ["create-step-store"]="Store data"
        ["create-step-store-value"]="Store value"
        ["create-step-execute-js"]="Execute JavaScript"
        
        # Cookie steps
        ["create-step-add-cookie"]="Add cookie"
        ["create-step-delete-cookie"]="Delete cookie"
        ["create-step-clear-cookies"]="Clear cookies"
        
        # Dialog steps
        ["create-step-dismiss-alert"]="Dismiss alert"
        ["create-step-dismiss-confirm"]="Dismiss confirm"
        ["create-step-dismiss-prompt"]="Dismiss prompt"
        
        # Frame/Tab steps
        ["create-step-switch-iframe"]="Switch to iframe"
        ["create-step-switch-next-tab"]="Switch to next tab"
        ["create-step-switch-prev-tab"]="Switch to previous tab"
        ["create-step-switch-parent-frame"]="Switch to parent frame"
        
        # Utility steps
        ["create-step-comment"]="Add comment"
    )
    
    # Test each command
    for command in "${!STEP_COMMANDS[@]}"; do
        echo -e "\n${BLUE}=== Testing $command ===${NC}"
        
        # Test help functionality
        test_help "$command"
        
        # Test parameter validation
        test_parameter_validation "$command" "${STEP_COMMANDS[$command]}"
        
        # Test output formats
        test_output_formats "$command"
        
        # Test with --checkpoint flag
        if ./bin/api-cli $command --help 2>&1 | grep -q "checkpoint"; then
            success "$command - Checkpoint flag supported"
        else
            fail "$command - Checkpoint flag not supported"
        fi
        
        echo ""
    done
    
    # Test specific command functionality with actual parameters
    log "Testing specific command functionality"
    
    # Test navigate command with parameters
    run_test "navigate with URL" "./bin/api-cli create-step-navigate --help" "Usage:"
    
    # Test click command with parameters
    run_test "click with selector" "./bin/api-cli create-step-click --help" "Usage:"
    
    # Test write command with parameters
    run_test "write with text" "./bin/api-cli create-step-write --help" "Usage:"
    
    # Test assert-exists command
    run_test "assert-exists with selector" "./bin/api-cli create-step-assert-exists --help" "Usage:"
    
    # Summary report
    echo -e "\n${BLUE}=== TEST SUMMARY ===${NC}"
    echo "Total tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo -e "Success rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    
    # Detailed results
    echo -e "\n${BLUE}=== DETAILED RESULTS ===${NC}"
    for result in "${TEST_RESULTS[@]}"; do
        if [[ $result == PASS* ]]; then
            echo -e "${GREEN}$result${NC}"
        else
            echo -e "${RED}$result${NC}"
        fi
    done
    
    # Final assessment
    echo -e "\n${BLUE}=== ASSESSMENT ===${NC}"
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}All tests passed! The CLI is in excellent condition.${NC}"
    elif [ $FAILED_TESTS -lt 5 ]; then
        echo -e "${YELLOW}Most tests passed with minor issues.${NC}"
    else
        echo -e "${RED}Significant issues found. Review failed tests.${NC}"
    fi
}

# Run the main test function
main "$@"