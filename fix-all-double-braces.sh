#!/bin/bash

# Fix all double closing braces in Version B commands

echo "Fixing all double closing braces..."

# List of all affected files
AFFECTED_FILES=(
    "create-step-assert-greater-than.go"
    "create-step-assert-matches.go"
    "create-step-assert-not-equals.go"
    "create-step-click.go"
    "create-step-comment.go"
    "create-step-cookie-wipe-all.go"
    "create-step-dismiss-prompt-with-text.go"
    "create-step-key.go"
    "create-step-pick-index.go"
    "create-step-pick-last.go"
    "create-step-mouse-move-by.go"
    "create-step-mouse-move-to.go"
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

for file in "${AFFECTED_FILES[@]}"; do
    filepath="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/$file"
    if [ -f "$filepath" ]; then
        echo "Fixing: $file"
        
        # Use sed to find and fix the pattern of double closing braces before return nil
        # This looks for a line with just }, followed by another line with just }, then empty line, then return nil
        sed -i '' '/^[[:space:]]*}[[:space:]]*$/{
            N
            /^[[:space:]]*}[[:space:]]*\n[[:space:]]*}[[:space:]]*$/{
                N
                N
                /^[[:space:]]*}[[:space:]]*\n[[:space:]]*}[[:space:]]*\n[[:space:]]*\n[[:space:]]*return nil/s/^[[:space:]]*}[[:space:]]*\n//
            }
        }' "$filepath"
    fi
done

echo "All double closing braces fixed!"