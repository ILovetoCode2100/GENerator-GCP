#!/bin/bash

# Script to add all Version B test steps to checkpoint 1680566 with JSON output

CHECKPOINT_ID=1680566
BINARY="/Users/marklovelady/_dev/virtuoso-api-cli-generator/bin/api-cli"

# Check if environment variables are set
if [ -z "$VIRTUOSO_API_TOKEN" ]; then
    echo "Error: VIRTUOSO_API_TOKEN environment variable is not set"
    echo "Please run: export VIRTUOSO_API_TOKEN='your-token-here'"
    exit 1
fi

if [ -z "$VIRTUOSO_API_BASE_URL" ]; then
    export VIRTUOSO_API_BASE_URL="https://api-app2.virtuoso.qa/api"
fi

echo '{'
echo '  "checkpoint_id": '$CHECKPOINT_ID','
echo '  "api_url": "'$VIRTUOSO_API_BASE_URL'",'
echo '  "steps": ['

POSITION=1
FIRST=true

# Function to run a command with JSON output
run_step_json() {
    local description="$1"
    local command="$2"
    
    # Add comma before each step except the first
    if [ "$FIRST" = true ]; then
        FIRST=false
    else
        echo ","
    fi
    
    echo -n '    {'
    echo -n '"position": '$POSITION', '
    echo -n '"description": "'$description'", '
    
    # Run command with JSON output
    if output=$($command -o json 2>&1); then
        if echo "$output" | grep -q "stepId"; then
            # Extract stepId from JSON output
            step_id=$(echo "$output" | grep -o '"stepId":[^,}]*' | cut -d: -f2 | tr -d ' ')
            echo -n '"status": "success", '
            echo -n '"step_id": '$step_id
        else
            echo -n '"status": "error", '
            echo -n '"error": "No step ID in response"'
        fi
    else
        # Escape error message for JSON
        error_msg=$(echo "$output" | sed 's/"/\\"/g' | tr '\n' ' ')
        echo -n '"status": "failed", '
        echo -n '"error": "'$error_msg'"'
    fi
    
    echo -n '}'
    
    POSITION=$((POSITION + 1))
}

# Add all the steps
run_step_json "Navigate to example.com" \
    "$BINARY create-step-navigate $CHECKPOINT_ID 'https://example.com' $POSITION"

run_step_json "Navigate to Google in new tab" \
    "$BINARY create-step-navigate $CHECKPOINT_ID 'https://google.com' $POSITION --new-tab"

run_step_json "Wait for search box" \
    "$BINARY create-step-wait-for-element-default $CHECKPOINT_ID 'search box' $POSITION"

run_step_json "Wait for results with timeout" \
    "$BINARY create-step-wait-for-element-timeout $CHECKPOINT_ID 'results' 5000 $POSITION"

run_step_json "Create session cookie" \
    "$BINARY create-step-cookie-create $CHECKPOINT_ID 'sessionId' 'abc123xyz' $POSITION"

run_step_json "Create preference cookie" \
    "$BINARY create-step-cookie-create $CHECKPOINT_ID 'userPref' 'darkMode=true' $POSITION"

run_step_json "Move mouse to (500,300)" \
    "$BINARY create-step-mouse-move-to $CHECKPOINT_ID 500 300 $POSITION"

run_step_json "Move mouse by (100,50)" \
    "$BINARY create-step-mouse-move-by $CHECKPOINT_ID 100 50 $POSITION"

run_step_json "Click submit button" \
    "$BINARY create-step-click $CHECKPOINT_ID 'Submit button' $POSITION"

run_step_json "Write in search field" \
    "$BINARY create-step-write $CHECKPOINT_ID 'search field' 'test automation' $POSITION"

run_step_json "Press Enter key" \
    "$BINARY create-step-key $CHECKPOINT_ID 'Enter' $POSITION"

run_step_json "Pick first dropdown option" \
    "$BINARY create-step-pick-index $CHECKPOINT_ID 'country dropdown' 0 $POSITION"

run_step_json "Pick last dropdown option" \
    "$BINARY create-step-pick-last $CHECKPOINT_ID 'state dropdown' $POSITION"

run_step_json "Store element text" \
    "$BINARY create-step-store-element-text $CHECKPOINT_ID 'product price' 'currentPrice' $POSITION"

run_step_json "Store literal value" \
    "$BINARY create-step-store-literal-value $CHECKPOINT_ID 'TestEnvironment' 'environment' $POSITION"

run_step_json "Scroll to top" \
    "$BINARY create-step-scroll-to-top $CHECKPOINT_ID $POSITION"

run_step_json "Scroll to position" \
    "$BINARY create-step-scroll-to-position $CHECKPOINT_ID 0 500 $POSITION"

run_step_json "Switch to iframe" \
    "$BINARY create-step-switch-iframe $CHECKPOINT_ID 'payment iframe' $POSITION"

run_step_json "Switch to parent frame" \
    "$BINARY create-step-switch-parent-frame $CHECKPOINT_ID $POSITION"

run_step_json "Switch to next tab" \
    "$BINARY create-step-switch-next-tab $CHECKPOINT_ID $POSITION"

run_step_json "Assert not equals" \
    "$BINARY create-step-assert-not-equals $CHECKPOINT_ID 'price' '0' $POSITION"

run_step_json "Assert greater than" \
    "$BINARY create-step-assert-greater-than $CHECKPOINT_ID 'quantity' '5' $POSITION"

run_step_json "Assert matches regex" \
    "$BINARY create-step-assert-matches $CHECKPOINT_ID 'email' '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' $POSITION"

run_step_json "Add comment" \
    "$BINARY create-step-comment $CHECKPOINT_ID 'Test section complete' $POSITION"

run_step_json "Execute script" \
    "$BINARY create-step-execute-script $CHECKPOINT_ID 'validateCheckout' $POSITION"

run_step_json "Upload file" \
    "$BINARY create-step-upload-url $CHECKPOINT_ID 'https://example.com/test.pdf' 'file input' $POSITION"

run_step_json "Dismiss prompt" \
    "$BINARY create-step-dismiss-prompt-with-text $CHECKPOINT_ID 'OK' $POSITION"

run_step_json "Resize window" \
    "$BINARY create-step-window-resize $CHECKPOINT_ID 1920 1080 $POSITION"

run_step_json "Clear all cookies" \
    "$BINARY create-step-cookie-wipe-all $CHECKPOINT_ID $POSITION"

echo ""
echo '  ],'
echo '  "total_steps": '$((POSITION - 1))
echo '}'