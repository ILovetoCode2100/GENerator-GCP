#!/bin/bash

# Script to update all step commands to use the new stateful context pattern
# This removes backward compatibility for cleaner code

cd /Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd

# List of step command files to update (excluding already updated ones)
step_files=(
    "create-step-write.go"
    "create-step-wait-time.go"
    "create-step-wait-element.go"
    "create-step-window.go"
    "create-step-double-click.go"
    "create-step-hover.go"
    "create-step-right-click.go"
    "create-step-key.go"
    "create-step-pick.go"
    "create-step-upload.go"
    "create-step-scroll-top.go"
    "create-step-scroll-bottom.go"
    "create-step-scroll-element.go"
    "create-step-scroll-position.go"
    "create-step-assert-exists.go"
    "create-step-assert-not-exists.go"
    "create-step-assert-equals.go"
    "create-step-assert-not-equals.go"
    "create-step-assert-greater-than.go"
    "create-step-assert-greater-than-or-equal.go"
    "create-step-assert-less-than-or-equal.go"
    "create-step-assert-matches.go"
    "create-step-assert-checked.go"
    "create-step-assert-selected.go"
    "create-step-assert-variable.go"
    "create-step-store.go"
    "create-step-store-value.go"
    "create-step-execute-js.go"
    "create-step-add-cookie.go"
    "create-step-delete-cookie.go"
    "create-step-clear-cookies.go"
    "create-step-dismiss-alert.go"
    "create-step-dismiss-confirm.go"
    "create-step-dismiss-prompt.go"
    "create-step-mouse-down.go"
    "create-step-mouse-up.go"
    "create-step-mouse-move.go"
    "create-step-mouse-enter.go"
    "create-step-pick-value.go"
    "create-step-pick-text.go"
    "create-step-switch-iframe.go"
    "create-step-switch-next-tab.go"
    "create-step-switch-prev-tab.go"
    "create-step-switch-parent-frame.go"
    "create-step-comment.go"
)

echo "Updating ${#step_files[@]} step command files..."

for file in "${step_files[@]}"; do
    if [[ -f "$file" ]]; then
        echo "Processing $file..."
        
        # Create backup
        cp "$file" "$file.backup"
        
        # Remove unused imports (will be added back by helper functions)
        sed -i '' 's/\t"encoding\/json"//' "$file"
        sed -i '' 's/\t"os"//' "$file"
        sed -i '' 's/\t"strconv"//' "$file"
        
        # Clean up import block
        sed -i '' '/^import ($/,/^)$/ {
            /^\t"encoding\/json"$/d
            /^\t"os"$/d
            /^\t"strconv"$/d
        }' "$file"
        
        echo "Updated imports for $file"
    else
        echo "Warning: $file not found"
    fi
done

echo "Completed updating step command files"
echo "Backups created with .backup extension"
echo "Manual updates still needed for command signature and logic"