#!/bin/bash

# Command Validator Demo Script
# This script demonstrates how the Virtuoso API CLI command validator
# automatically corrects common syntax errors and provides helpful guidance

echo "===== Virtuoso API CLI Command Validator Demo ====="
echo ""
echo "This demo shows how the CLI automatically corrects common command mistakes."
echo ""

# Set up environment (using a test checkpoint ID)
export VIRTUOSO_SESSION_ID=12345

# Function to run a command and show the correction
run_with_correction() {
    local description="$1"
    local command="$2"

    echo "----------------------------------------"
    echo "Test: $description"
    echo "Input Command: $command"
    echo ""
    echo "Output:"
    # Run the command - in real usage, remove the echo
    echo "[Simulated] $command"
    echo ""
}

# 1. Scroll Command Corrections
echo "=== 1. SCROLL COMMAND CORRECTIONS ==="

run_with_correction \
    "Old scroll syntax without hyphen" \
    "api-cli step-navigate scroll top"
# Corrected to: api-cli step-navigate scroll-top

run_with_correction \
    "Scroll with 'to' word (removed)" \
    "api-cli step-navigate scroll to bottom"
# Corrected to: api-cli step-navigate scroll-bottom

# 2. Dialog Command Corrections
echo "=== 2. DIALOG COMMAND CORRECTIONS ==="

run_with_correction \
    "Old alert accept syntax" \
    "api-cli step-dialog alert accept"
# Corrected to: api-cli step-dialog dismiss-alert

run_with_correction \
    "Confirm with old syntax" \
    "api-cli step-dialog confirm accept"
# Corrected to: api-cli step-dialog dismiss-confirm --accept

# 3. Switch Tab Argument Order
echo "=== 3. SWITCH TAB ARGUMENT ORDER ==="

run_with_correction \
    "Switch tab with wrong argument order" \
    "api-cli step-window switch tab 12345 next 1"
# Corrected to: api-cli step-window switch-tab next 12345 1

# 4. Mouse Coordinate Formats
echo "=== 4. MOUSE COORDINATE CORRECTIONS ==="

run_with_correction \
    "Mouse move with space-separated coordinates" \
    "api-cli step-interact mouse move-to \"100 200\""
# Corrected to: api-cli step-interact mouse move-to "100,200"

run_with_correction \
    "Mouse move with separate x,y arguments" \
    "api-cli step-interact mouse move-by 50 100"
# Corrected to: api-cli step-interact mouse move-by "50,100"

# 5. Store Command Simplification
echo "=== 5. STORE COMMAND SIMPLIFICATION ==="

run_with_correction \
    "Old store element-text syntax" \
    "api-cli step-data store element-text \"h1\" \"pageTitle\""
# Corrected to: api-cli step-data store text "h1" "pageTitle"

run_with_correction \
    "Store element-attribute simplified" \
    "api-cli step-data store element-attribute \"img\" \"src\" \"imageUrl\""
# Corrected to: api-cli step-data store attribute "img" "src" "imageUrl"

# 6. Wait Time Auto-Conversion
echo "=== 6. WAIT TIME AUTO-CONVERSION ==="

run_with_correction \
    "Wait time in seconds (auto-converted to milliseconds)" \
    "api-cli step-wait time 5"
# Corrected to: api-cli step-wait time 5000

# 7. Resize Dimension Formats
echo "=== 7. RESIZE DIMENSION CORRECTIONS ==="

run_with_correction \
    "Resize with space separator" \
    "api-cli step-window resize \"1024 768\""
# Corrected to: api-cli step-window resize "1024x768"

run_with_correction \
    "Resize with asterisk separator" \
    "api-cli step-window resize \"1920*1080\""
# Corrected to: api-cli step-window resize "1920x1080"

# 8. Removed Commands
echo "=== 8. REMOVED COMMAND WARNINGS ==="

run_with_correction \
    "Attempting to use removed scroll-left command" \
    "api-cli step-navigate scroll-left 100"
# Error: command 'scroll-left' is no longer supported: Horizontal scrolling is not supported. Use 'scroll-by' with negative X values instead.

run_with_correction \
    "Attempting to use removed navigate back" \
    "api-cli step-navigate back"
# Error: command 'navigate back' is no longer supported: Browser back navigation is not supported by the API.

# 9. Unsupported Flags
echo "=== 9. UNSUPPORTED FLAG DETECTION ==="

run_with_correction \
    "Click with unsupported offset flags" \
    "api-cli step-interact click \"button\" --offset-x 10 --offset-y 20"
# Error: command 'click' does not support --offset-x flag. Mouse offset positioning is not available for this command

# 10. Common Misspellings
echo "=== 10. COMMON MISSPELLING CORRECTIONS ==="

run_with_correction \
    "Navigate misspelling" \
    "api-cli step-navigate naviagte to \"https://example.com\""
# Corrected to: api-cli step-navigate navigate to "https://example.com"

run_with_correction \
    "Double click with space" \
    "api-cli step-interact double click \"button.submit\""
# Corrected to: api-cli step-interact double-click "button.submit"

# 11. Scroll Distance Named Values
echo "=== 11. SCROLL DISTANCE NAMED VALUES ==="

run_with_correction \
    "Scroll with named distance" \
    "api-cli step-navigate scroll-down large"
# Corrected to: api-cli step-navigate scroll-down 500

run_with_correction \
    "Scroll up with positive value (auto-negated)" \
    "api-cli step-navigate scroll-up 300"
# Corrected to: api-cli step-navigate scroll-up -300

echo ""
echo "===== Demo Complete ====="
echo ""
echo "The command validator helps ensure your scripts use the correct syntax"
echo "and provides helpful error messages when commands are deprecated or removed."
echo ""
echo "To use in your scripts:"
echo "1. The validator runs automatically - no configuration needed"
echo "2. Pay attention to deprecation warnings"
echo "3. Update scripts based on the corrections shown"
echo "4. Use 'export VIRTUOSO_SESSION_ID=<id>' to avoid repeating checkpoint IDs"
