#!/bin/bash

# Fix syntax errors in Version B command files

echo "Fixing syntax errors in Version B command files..."

# List of files with syntax errors
FILES_WITH_ERRORS=(
    "create-step-assert-greater-than-or-equal.go"
    "create-step-assert-greater-than.go"
    "create-step-assert-matches.go"
    "create-step-assert-not-equals.go"
    "create-step-click.go"
    "create-step-comment.go"
    "create-step-cookie-create.go"
    "create-step-cookie-wipe-all.go"
    "create-step-dismiss-prompt-with-text.go"
    "create-step-key.go"
    "create-step-mouse-move-by.go"
    "create-step-mouse-move-to.go"
    "create-step-navigate.go"
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

for file in "${FILES_WITH_ERRORS[@]}"; do
    filepath="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/$file"
    if [ -f "$filepath" ]; then
        echo "Fixing: $file"
        
        # Remove the extra closing brace before the final return statement
        # This pattern looks for lines with just "}" followed by "}" on the next line
        awk '
        {
            lines[NR] = $0
        }
        END {
            for (i = 1; i <= NR; i++) {
                if (i < NR-2 && lines[i] ~ /^\s*}\s*$/ && lines[i+1] ~ /^\s*}\s*$/ && lines[i+2] ~ /^\s*$/ && lines[i+3] ~ /^\s*return nil/) {
                    # Skip the first closing brace
                    continue
                } else {
                    print lines[i]
                }
            }
        }
        ' "$filepath" > "$filepath.tmp" && mv "$filepath.tmp" "$filepath"
        
        # Also remove any "No newline at end of file" markers
        sed -i '' '/No newline at end of file/d' "$filepath"
        
        # Ensure file ends with a newline
        if [ -n "$(tail -c 1 "$filepath")" ]; then
            echo >> "$filepath"
        fi
    fi
done

echo "Syntax errors fixed!"