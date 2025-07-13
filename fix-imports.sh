#!/bin/bash

# Fix imports in all Version B command files

echo "Fixing imports in Version B command files..."

# List of Version B command files that need import fixes
VERSION_B_COMMANDS=(
    "create-step-cookie-create.go"
    "create-step-cookie-wipe-all.go"
    "create-step-dismiss-prompt-with-text.go"
    "create-step-execute-script.go"
    "create-step-mouse-move-by.go"
    "create-step-mouse-move-to.go"
    "create-step-pick-index.go"
    "create-step-pick-last.go"
    "create-step-scroll.go"
    "create-step-store-element-text.go"
    "create-step-store-literal-value.go"
    "create-step-upload-url.go"
    "create-step-wait-for-element-default.go"
    "create-step-wait-for-element-timeout.go"
    "create-step-window-resize.go"
    "create-step-navigate.go"
    "create-step-click.go"
    "create-step-write.go"
    "create-step-key.go"
    "create-step-comment.go"
    "create-step-switch-iframe.go"
    "create-step-switch-parent-frame.go"
    "create-step-switch-next-tab.go"
    "create-step-switch-prev-tab.go"
    "create-step-assert-not-equals.go"
    "create-step-assert-greater-than.go"
    "create-step-assert-greater-than-or-equal.go"
    "create-step-assert-matches.go"
)

for cmd in "${VERSION_B_COMMANDS[@]}"; do
    file="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/$cmd"
    if [ -f "$file" ]; then
        echo "Fixing imports in $cmd"
        # Replace Version B import with Version A import
        sed -i '' 's|github.com/virtuoso/api-cli-generator/pkg/virtuoso|github.com/marklovelady/api-cli-generator/pkg/virtuoso|g' "$file"
        
        # Also update the package declaration if needed
        sed -i '' 's|package main|package cmd|g' "$file"
    else
        echo "Warning: File not found: $file"
    fi
done

echo "Import fixes complete!"