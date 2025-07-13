#!/bin/bash

# Simple fix for double closing braces

echo "Applying simple syntax fix to all command files..."

# Process each Go file in src/cmd
find /Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd -name "*.go" -type f | while read file; do
    # Use perl to fix the double closing brace pattern before return nil
    perl -i -pe 's/^\s*}\s*\n\s*}\s*\n\s*\n\s*return nil/\t}\n\n\treturn nil/g' "$file"
    
    # Also fix the pattern where there's no empty line between braces and return
    perl -i -pe 's/^\s*}\s*\n\s*}\s*\n\s*return nil/\t}\n\n\treturn nil/g' "$file"
    
    # Fix any remaining double braces on the same line
    sed -i '' 's/}[[:space:]]*}$/}/' "$file"
    
    echo "Processed: $(basename "$file")"
done

echo "Simple syntax fix complete!"