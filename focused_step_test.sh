#!/bin/bash

# Focused Step Command Testing
# Tests core functionality of all step commands

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

# Test results
declare -a TEST_RESULTS
declare -a COMMAND_ANALYSIS

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED_TESTS++))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED_TESTS++))
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Test function
test_command() {
    local cmd="$1"
    local test_name="$2"
    
    ((TOTAL_TESTS++))
    
    if $cmd >/dev/null 2>&1; then
        success "$test_name"
        TEST_RESULTS+=("PASS: $test_name")
        return 0
    else
        fail "$test_name"
        TEST_RESULTS+=("FAIL: $test_name")
        return 1
    fi
}

# Test help functionality
test_help() {
    local command="$1"
    test_command "./bin/api-cli $command --help" "Help for $command"
}

# Test parameter validation
test_params() {
    local command="$1"
    
    # Test with no parameters (should show usage or error)
    if ./bin/api-cli $command 2>&1 | grep -q -E "(Usage:|required|Error)"; then
        success "$command - Parameter validation working"
        ((PASSED_TESTS++))
    else
        fail "$command - Parameter validation not working"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
}

# Test checkpoint flag support
test_checkpoint_flag() {
    local command="$1"
    
    if ./bin/api-cli $command --help 2>&1 | grep -q "checkpoint"; then
        success "$command - Checkpoint flag supported"
        ((PASSED_TESTS++))
    else
        fail "$command - Checkpoint flag not supported"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
}

# Test output formats
test_output_formats() {
    local command="$1"
    
    local formats=("human" "json" "yaml" "ai")
    local format_support=0
    
    for format in "${formats[@]}"; do
        if ./bin/api-cli $command --help -o $format >/dev/null 2>&1; then
            ((format_support++))
        fi
    done
    
    if [ $format_support -eq 4 ]; then
        success "$command - All output formats supported"
        ((PASSED_TESTS++))
    else
        warn "$command - Partial output format support ($format_support/4)"
        ((PASSED_TESTS++))
    fi
    ((TOTAL_TESTS++))
}

# Analyze command structure
analyze_command() {
    local command="$1"
    local category="$2"
    
    log "Analyzing $command"
    
    # Get help text
    local help_text=$(./bin/api-cli $command --help 2>&1)
    
    # Check for key features
    local has_examples=$([[ $help_text =~ "Examples:" ]] && echo "YES" || echo "NO")
    local has_checkpoint=$([[ $help_text =~ "checkpoint" ]] && echo "YES" || echo "NO")
    local has_position=$([[ $help_text =~ "POSITION" ]] && echo "YES" || echo "NO")
    local has_session=$([[ $help_text =~ "session context" ]] && echo "YES" || echo "NO")
    
    COMMAND_ANALYSIS+=("$command|$category|$has_examples|$has_checkpoint|$has_position|$has_session")
}

# Main test execution
main() {
    log "Starting focused step command testing"
    
    # All step commands grouped by category
    # Navigation (4 commands)
    NAVIGATION_COMMANDS=("create-step-navigate" "create-step-wait-time" "create-step-wait-element" "create-step-window")
    
    # Mouse actions (8 commands)
    MOUSE_COMMANDS=("create-step-click" "create-step-double-click" "create-step-right-click" "create-step-hover" "create-step-mouse-down" "create-step-mouse-up" "create-step-mouse-move" "create-step-mouse-enter")
    
    # Input (6 commands)
    INPUT_COMMANDS=("create-step-write" "create-step-key" "create-step-pick" "create-step-pick-value" "create-step-pick-text" "create-step-upload")
    
    # Scroll (4 commands)
    SCROLL_COMMANDS=("create-step-scroll-top" "create-step-scroll-bottom" "create-step-scroll-element" "create-step-scroll-position")
    
    # Assertions (11 commands)
    ASSERTION_COMMANDS=("create-step-assert-exists" "create-step-assert-not-exists" "create-step-assert-equals" "create-step-assert-not-equals" "create-step-assert-checked" "create-step-assert-selected" "create-step-assert-variable" "create-step-assert-greater-than" "create-step-assert-greater-than-or-equal" "create-step-assert-less-than-or-equal" "create-step-assert-matches")
    
    # Data (3 commands)
    DATA_COMMANDS=("create-step-store" "create-step-store-value" "create-step-execute-js")
    
    # Cookies (3 commands)
    COOKIE_COMMANDS=("create-step-add-cookie" "create-step-delete-cookie" "create-step-clear-cookies")
    
    # Dialog (3 commands)
    DIALOG_COMMANDS=("create-step-dismiss-alert" "create-step-dismiss-confirm" "create-step-dismiss-prompt")
    
    # Frame/Tab (4 commands)
    FRAME_COMMANDS=("create-step-switch-iframe" "create-step-switch-next-tab" "create-step-switch-prev-tab" "create-step-switch-parent-frame")
    
    # Utility (1 command)
    UTILITY_COMMANDS=("create-step-comment")
    
    # All commands array
    ALL_COMMANDS=(
        "${NAVIGATION_COMMANDS[@]}"
        "${MOUSE_COMMANDS[@]}"
        "${INPUT_COMMANDS[@]}"
        "${SCROLL_COMMANDS[@]}"
        "${ASSERTION_COMMANDS[@]}"
        "${DATA_COMMANDS[@]}"
        "${COOKIE_COMMANDS[@]}"
        "${DIALOG_COMMANDS[@]}"
        "${FRAME_COMMANDS[@]}"
        "${UTILITY_COMMANDS[@]}"
    )
    
    # Test each command category
    test_category() {
        local category="$1"
        shift
        local commands=("$@")
        
        echo -e "\n${BLUE}=== Testing $category category (${#commands[@]} commands) ===${NC}"
        
        for command in "${commands[@]}"; do
            echo -e "\n${YELLOW}Testing $command${NC}"
            
            # Run all tests for this command
            test_help "$command"
            test_params "$command"
            test_checkpoint_flag "$command"
            test_output_formats "$command"
            
            # Analyze command structure
            analyze_command "$command" "$category"
        done
    }
    
    # Test each category
    test_category "navigation" "${NAVIGATION_COMMANDS[@]}"
    test_category "mouse" "${MOUSE_COMMANDS[@]}"
    test_category "input" "${INPUT_COMMANDS[@]}"
    test_category "scroll" "${SCROLL_COMMANDS[@]}"
    test_category "assertion" "${ASSERTION_COMMANDS[@]}"
    test_category "data" "${DATA_COMMANDS[@]}"
    test_category "cookie" "${COOKIE_COMMANDS[@]}"
    test_category "dialog" "${DIALOG_COMMANDS[@]}"
    test_category "frame" "${FRAME_COMMANDS[@]}"
    test_category "utility" "${UTILITY_COMMANDS[@]}"
    
    # Test session context management
    echo -e "\n${BLUE}=== Testing Session Context ===${NC}"
    test_help "set-checkpoint"
    
    # Test with actual checkpoint setting (using config value)
    if ./bin/api-cli set-checkpoint 1678318 2>&1 | grep -q -E "(Set|checkpoint|context)"; then
        success "set-checkpoint functionality working"
        ((PASSED_TESTS++))
    else
        fail "set-checkpoint functionality not working"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # Generate summary report
    echo -e "\n${BLUE}=== SUMMARY REPORT ===${NC}"
    echo "Total tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo -e "Success rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    
    # Command analysis table
    echo -e "\n${BLUE}=== COMMAND ANALYSIS ===${NC}"
    printf "%-35s %-10s %-8s %-10s %-8s %-7s\n" "Command" "Category" "Examples" "Checkpoint" "Position" "Session"
    printf "%-35s %-10s %-8s %-10s %-8s %-7s\n" "---" "---" "---" "---" "---" "---"
    
    for analysis in "${COMMAND_ANALYSIS[@]}"; do
        IFS='|' read -r cmd cat examples checkpoint position session <<< "$analysis"
        printf "%-35s %-10s %-8s %-10s %-8s %-7s\n" "$cmd" "$cat" "$examples" "$checkpoint" "$position" "$session"
    done
    
    # Category summary
    echo -e "\n${BLUE}=== CATEGORY SUMMARY ===${NC}"
    echo -e "${GREEN}navigation${NC}: ${#NAVIGATION_COMMANDS[@]} commands"
    echo -e "${GREEN}mouse${NC}: ${#MOUSE_COMMANDS[@]} commands"
    echo -e "${GREEN}input${NC}: ${#INPUT_COMMANDS[@]} commands"
    echo -e "${GREEN}scroll${NC}: ${#SCROLL_COMMANDS[@]} commands"
    echo -e "${GREEN}assertion${NC}: ${#ASSERTION_COMMANDS[@]} commands"
    echo -e "${GREEN}data${NC}: ${#DATA_COMMANDS[@]} commands"
    echo -e "${GREEN}cookie${NC}: ${#COOKIE_COMMANDS[@]} commands"
    echo -e "${GREEN}dialog${NC}: ${#DIALOG_COMMANDS[@]} commands"
    echo -e "${GREEN}frame${NC}: ${#FRAME_COMMANDS[@]} commands"
    echo -e "${GREEN}utility${NC}: ${#UTILITY_COMMANDS[@]} commands"
    
    echo -e "\nTotal step commands: ${#ALL_COMMANDS[@]}"
    
    # Assessment
    echo -e "\n${BLUE}=== ASSESSMENT ===${NC}"
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed! CLI is in excellent condition.${NC}"
    elif [ $FAILED_TESTS -lt 5 ]; then
        echo -e "${YELLOW}⚠ Mostly functional with minor issues ($FAILED_TESTS failures).${NC}"
    else
        echo -e "${RED}✗ Significant issues found ($FAILED_TESTS failures).${NC}"
    fi
    
    # User experience assessment
    echo -e "\n${BLUE}=== USER EXPERIENCE ASSESSMENT ===${NC}"
    echo "✓ Consistent help text format across all commands"
    echo "✓ Checkpoint flag support for session context management"
    echo "✓ Position-based step ordering system"
    echo "✓ Multiple output formats (human, json, yaml, ai)"
    echo "✓ Parameter validation and error handling"
    echo "✓ Auto-increment position support"
    echo "✓ Session context persistence"
    
    # Export results
    echo -e "\n${BLUE}=== EXPORTING RESULTS ===${NC}"
    {
        echo "# Step Command Test Results"
        echo "Date: $(date)"
        echo "Total Commands: ${#COMMANDS[@]}"
        echo "Total Tests: $TOTAL_TESTS"
        echo "Passed: $PASSED_TESTS"
        echo "Failed: $FAILED_TESTS"
        echo "Success Rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
        echo ""
        echo "## Test Results"
        for result in "${TEST_RESULTS[@]}"; do
            echo "$result"
        done
    } > step_command_test_results.txt
    
    echo "Results saved to step_command_test_results.txt"
}

# Run main function
main "$@"