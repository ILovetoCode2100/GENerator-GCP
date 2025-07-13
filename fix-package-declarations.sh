#!/bin/bash

# Fix package declarations in src/cmd directory

echo "Fixing package declarations in src/cmd..."

# Change all "package cmd" to "package main"
find /Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd -name "*.go" -type f | while read file; do
    if grep -q "^package cmd" "$file"; then
        echo "Fixing: $(basename "$file")"
        sed -i '' 's/^package cmd$/package main/' "$file"
    fi
done

echo "Package declarations fixed!"