#!/bin/bash

# Add missing closing braces to files

echo "Adding missing closing braces..."

# List of files that need a closing brace
FILES_NEED_BRACE=(
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

for file in "${FILES_NEED_BRACE[@]}"; do
    filepath="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/$file"
    if [ -f "$filepath" ]; then
        # Check if file ends with "return nil" without a closing brace after it
        if tail -2 "$filepath" | grep -q "return nil" && ! tail -1 "$filepath" | grep -q "}"; then
            echo "Adding closing brace to: $file"
            echo "}" >> "$filepath"
        fi
    fi
done

echo "Missing closing braces added!"