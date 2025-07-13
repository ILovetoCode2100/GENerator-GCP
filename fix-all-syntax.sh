#!/bin/bash

# Fix all syntax errors in Version B command files

echo "Fixing all syntax errors in command files..."

# Find all Go files with the pattern of extra closing braces
find /Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd -name "*.go" -type f | while read file; do
    # Check if file has the problematic pattern
    if grep -q "}\s*}\s*$" "$file"; then
        echo "Fixing: $(basename "$file")"
        
        # Use sed to fix the double closing brace pattern
        sed -i '' -E 's/^(\s*})\s*}$/\1/' "$file"
        
        # Remove "No newline at end of file" markers
        sed -i '' '/No newline at end of file/d' "$file"
        
        # Ensure file ends with newline
        if [ -n "$(tail -c 1 "$file")" ]; then
            echo >> "$file"
        fi
    fi
done

echo "All syntax errors fixed!"