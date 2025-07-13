#!/bin/bash

# Test all CLI commands to identify failures
# This script tests command parsing and help output

echo "Testing all CLI commands..."
echo "=========================="

# Build first
make build

COMMANDS=(
    "create-step-cookie-create"
    "create-step-cookie-wipe-all"
    "create-step-upload-url"
    "create-step-dismiss-prompt-with-text"
    "create-step-execute-script"
    "create-step-mouse-move-to"
    "create-step-mouse-move-by"
    "create-step-pick-index"
    "create-step-pick-last"
    "create-step-wait-for-element-timeout"
    "create-step-wait-for-element-default"
    "create-step-switch-iframe"
    "create-step-store-element-text"
    "create-step-store-literal-value"
    "create-step-switch-next-tab"
    "create-step-switch-parent-frame"
    "create-step-switch-prev-tab"
    "create-step-assert-not-equals"
    "create-step-assert-greater-than"
    "create-step-assert-greater-than-or-equal"
    "create-step-assert-matches"
)

FAILED_COMMANDS=()
PASSED_COMMANDS=()

for cmd in "${COMMANDS[@]}"; do
    echo -n "Testing $cmd... "
    
    # Test help output
    if ./bin/api-cli $cmd --help >/dev/null 2>&1; then
        echo "PASSED"
        PASSED_COMMANDS+=("$cmd")
    else
        echo "FAILED"
        FAILED_COMMANDS+=("$cmd")
    fi
done

echo ""
echo "Summary:"
echo "========="
echo "Total commands: ${#COMMANDS[@]}"
echo "Passed: ${#PASSED_COMMANDS[@]}"
echo "Failed: ${#FAILED_COMMANDS[@]}"

if [ ${#FAILED_COMMANDS[@]} -gt 0 ]; then
    echo ""
    echo "Failed commands:"
    for cmd in "${FAILED_COMMANDS[@]}"; do
        echo "  - $cmd"
    done
fi

echo ""
echo "Detailed help output test:"
echo "=========================="

# Test detailed help for a few commands
for cmd in "${COMMANDS[@]:0:3}"; do
    echo "--- $cmd ---"
    ./bin/api-cli $cmd --help 2>&1 | head -10
    echo ""
done