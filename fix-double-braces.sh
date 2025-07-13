#!/bin/bash

# Fix double closing braces in all Version B command files

echo "Fixing double closing braces in command files..."

# List of files that likely have this issue
VERSION_B_COMMANDS=(
    "create-step-assert-greater-than-or-equal.go"
    "create-step-assert-greater-than.go"
    "create-step-assert-matches.go"
    "create-step-assert-not-equals.go"
    "create-step-click.go"
    "create-step-comment.go"
    "create-step-cookie-wipe-all.go"
    "create-step-dismiss-prompt-with-text.go"
    "create-step-key.go"
    "create-step-mouse-move-by.go"
    "create-step-mouse-move-to.go"
    "create-step-pick-index.go"
    "create-step-pick-last.go"
    "create-step-scroll.go"
    "create-step-store-element-text.go"
    "create-step-store-literal-value.go"
    "create-step-switch-iframe.go"
    "create-step-switch-next-tab.go"
    "create-step-switch-parent-frame.go"
    "create-step-switch-prev-tab.go"
    "create-step-upload-url.go"
    "create-step-wait-for-element-default.go"
    "create-step-wait-for-element-timeout.go"
    "create-step-window-resize.go"
    "create-step-write.go"
)

for file in "${VERSION_B_COMMANDS[@]}"; do
    filepath="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/$file"
    if [ -f "$filepath" ]; then
        # Check if file has the problematic pattern (closing brace, then another closing brace, then return)
        if grep -B2 "return nil" "$filepath" | grep -q "}\s*}"; then
            echo "Fixing: $file"
            
            # Create a temporary file with the fix
            awk '
            BEGIN { prev = ""; prevprev = "" }
            {
                if (prevprev ~ /^\s*}\s*$/ && prev ~ /^\s*}\s*$/ && $0 ~ /^\s*$/) {
                    # Skip the second closing brace
                    prev = $0
                    next
                } else {
                    if (prevprev != "") print prevprev
                    prevprev = prev
                    prev = $0
                }
            }
            END {
                if (prevprev != "") print prevprev
                if (prev != "") print prev
            }
            ' "$filepath" > "$filepath.tmp" && mv "$filepath.tmp" "$filepath"
        fi
    fi
done

echo "Double closing braces fixed!"