#!/bin/bash

# Script to add all Version B test steps to checkpoint 1680566 - Fixed version

CHECKPOINT_ID=1680566
BINARY="/Users/marklovelady/_dev/virtuoso-api-cli-generator/bin/api-cli"

# Check if environment variables are set
if [ -z "$VIRTUOSO_API_TOKEN" ]; then
    echo "Error: VIRTUOSO_API_TOKEN environment variable is not set"
    echo "Please run: export VIRTUOSO_API_TOKEN='your-token-here'"
    exit 1
fi

if [ -z "$VIRTUOSO_API_BASE_URL" ]; then
    echo "Setting default VIRTUOSO_API_BASE_URL..."
    export VIRTUOSO_API_BASE_URL="https://api-app2.virtuoso.qa/api"
fi

echo "========================================="
echo "Adding All Test Steps to Checkpoint $CHECKPOINT_ID"
echo "========================================="
echo "API URL: $VIRTUOSO_API_BASE_URL"
echo ""

# Counter for position
POSITION=1

# Function to run a command and report status
run_step() {
    local description="$1"
    shift
    
    echo "[$POSITION] $description"
    echo "Command: $*"
    
    if output=$("$@" 2>&1); then
        if echo "$output" | grep -q "successfully"; then
            echo "✅ Success"
            if echo "$output" | grep -q "Step ID:"; then
                step_id=$(echo "$output" | grep "Step ID:" | awk '{print $3}')
                echo "   Step ID: $step_id"
            fi
        else
            echo "⚠️  Command executed but may have issues"
            echo "   Output: $output" | head -5
        fi
    else
        echo "❌ Failed"
        echo "   Error: $output" | head -5
    fi
    echo ""
    
    POSITION=$((POSITION + 1))
}

echo "📋 Adding Navigation and Basic Steps"
echo "===================================="

# Navigation
run_step "Navigate to example.com" \
    "$BINARY" create-step-navigate "$CHECKPOINT_ID" "https://example.com" "$POSITION"

run_step "Navigate to Google in new tab" \
    "$BINARY" create-step-navigate "$CHECKPOINT_ID" "https://google.com" "$POSITION" --new-tab

# Wait commands
run_step "Wait for search box (default timeout)" \
    "$BINARY" create-step-wait-for-element-default "$CHECKPOINT_ID" "search box" "$POSITION"

run_step "Wait for results with 5s timeout" \
    "$BINARY" create-step-wait-for-element-timeout "$CHECKPOINT_ID" "results" 5000 "$POSITION"

echo "📋 Adding Cookie Management Steps"
echo "================================="

# Cookie management
run_step "Create session cookie" \
    "$BINARY" create-step-cookie-create "$CHECKPOINT_ID" "sessionId" "abc123xyz" "$POSITION"

run_step "Create user preference cookie" \
    "$BINARY" create-step-cookie-create "$CHECKPOINT_ID" "userPref" "darkMode=true" "$POSITION"

echo "📋 Adding Mouse Movement Steps"
echo "=============================="

# Mouse movements
run_step "Move mouse to coordinates (500, 300)" \
    "$BINARY" create-step-mouse-move-to "$CHECKPOINT_ID" 500 300 "$POSITION"

run_step "Move mouse by offset (100, 50)" \
    "$BINARY" create-step-mouse-move-by "$CHECKPOINT_ID" 100 50 "$POSITION"

echo "📋 Adding Click and Input Steps"
echo "==============================="

# Click actions
run_step "Click on submit button" \
    "$BINARY" create-step-click "$CHECKPOINT_ID" "Submit button" "$POSITION"

run_step "Click using variable target" \
    "$BINARY" create-step-click "$CHECKPOINT_ID" "" "$POSITION" --variable "dynamicButton"

# Write actions
run_step "Write text in search field" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "search field" "test automation" "$POSITION"

run_step "Write and store in variable" \
    "$BINARY" create-step-write "$CHECKPOINT_ID" "username field" "testuser" "$POSITION" --variable "username"

echo "📋 Adding Keyboard Steps"
echo "========================"

# Key presses
run_step "Press Enter key globally" \
    "$BINARY" create-step-key "$CHECKPOINT_ID" "Enter" "$POSITION"

run_step "Press Tab in specific field" \
    "$BINARY" create-step-key "$CHECKPOINT_ID" "Tab" "$POSITION" --target "input field"

echo "📋 Adding Dropdown Selection Steps"
echo "=================================="

# Dropdown selections
run_step "Pick first option by index" \
    "$BINARY" create-step-pick-index "$CHECKPOINT_ID" "country dropdown" 0 "$POSITION"

run_step "Pick last option" \
    "$BINARY" create-step-pick-last "$CHECKPOINT_ID" "state dropdown" "$POSITION"

echo "📋 Adding Storage Steps"
echo "======================="

# Storage operations
run_step "Store element text" \
    "$BINARY" create-step-store-element-text "$CHECKPOINT_ID" "product price" "currentPrice" "$POSITION"

run_step "Store literal value" \
    "$BINARY" create-step-store-literal-value "$CHECKPOINT_ID" "TestEnvironment" "environment" "$POSITION"

echo "📋 Adding Scroll Steps"
echo "======================"

# Scrolling
run_step "Scroll to top" \
    "$BINARY" create-step-scroll-to-top "$CHECKPOINT_ID" "$POSITION"

run_step "Scroll to position (0, 500)" \
    "$BINARY" create-step-scroll-to-position "$CHECKPOINT_ID" 0 500 "$POSITION"

run_step "Scroll by offset (0, 200)" \
    "$BINARY" create-step-scroll-by-offset "$CHECKPOINT_ID" 0 200 "$POSITION"

echo "📋 Adding Frame/Tab Navigation Steps"
echo "===================================="

# Frame and tab switching
run_step "Switch to iframe" \
    "$BINARY" create-step-switch-iframe "$CHECKPOINT_ID" "payment iframe" "$POSITION"

run_step "Switch to parent frame" \
    "$BINARY" create-step-switch-parent-frame "$CHECKPOINT_ID" "$POSITION"

run_step "Switch to next tab" \
    "$BINARY" create-step-switch-next-tab "$CHECKPOINT_ID" "$POSITION"

run_step "Switch to previous tab" \
    "$BINARY" create-step-switch-prev-tab "$CHECKPOINT_ID" "$POSITION"

echo "📋 Adding Assertion Steps"
echo "========================"

# Assertions
run_step "Assert price not equals zero" \
    "$BINARY" create-step-assert-not-equals "$CHECKPOINT_ID" "price" "0" "$POSITION"

run_step "Assert quantity greater than 5" \
    "$BINARY" create-step-assert-greater-than "$CHECKPOINT_ID" "quantity" "5" "$POSITION"

run_step "Assert score >= 75" \
    "$BINARY" create-step-assert-greater-than-or-equal "$CHECKPOINT_ID" "score" "75" "$POSITION"

run_step "Assert email matches pattern" \
    "$BINARY" create-step-assert-matches "$CHECKPOINT_ID" "email" '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' "$POSITION"

echo "📋 Adding Utility Steps"
echo "======================="

# Utility steps
run_step "Add comment about test section" \
    "$BINARY" create-step-comment "$CHECKPOINT_ID" "Starting checkout process validation" "$POSITION"

run_step "Execute custom script" \
    "$BINARY" create-step-execute-script "$CHECKPOINT_ID" "validateCheckout" "$POSITION"

run_step "Upload file from URL" \
    "$BINARY" create-step-upload-url "$CHECKPOINT_ID" "https://example.com/test-file.pdf" "file input" "$POSITION"

run_step "Dismiss prompt with text" \
    "$BINARY" create-step-dismiss-prompt-with-text "$CHECKPOINT_ID" "OK" "$POSITION"

run_step "Resize window to 1920x1080" \
    "$BINARY" create-step-window-resize "$CHECKPOINT_ID" 1920 1080 "$POSITION"

# Final cookie cleanup
run_step "Clear all cookies at end" \
    "$BINARY" create-step-cookie-wipe-all "$CHECKPOINT_ID" "$POSITION"

echo "========================================="
echo "Summary"
echo "========================================="
echo "Total steps attempted: $((POSITION - 1))"
echo ""
echo "All Version B commands have been added to checkpoint $CHECKPOINT_ID"
echo "Check the Virtuoso UI to see all the steps in order."
echo "========================================="