#!/bin/bash

# Adapt Version B commands to use Version A's config-based client pattern

echo "Adapting Version B commands to Version A pattern..."

# List of Version B command files
VERSION_B_COMMANDS=(
    "create-step-cookie-create.go"
    "create-step-cookie-wipe-all.go"
    "create-step-dismiss-prompt-with-text.go"
    "create-step-execute-script.go"
    "create-step-mouse-move-by.go"
    "create-step-mouse-move-to.go"
    "create-step-pick-index.go"
    "create-step-pick-last.go"
    "create-step-store-element-text.go"
    "create-step-store-literal-value.go"
    "create-step-upload-url.go"
    "create-step-wait-for-element-default.go"
    "create-step-wait-for-element-timeout.go"
    "create-step-window-resize.go"
)

for cmd in "${VERSION_B_COMMANDS[@]}"; do
    file="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/$cmd"
    if [ -f "$file" ]; then
        echo "Adapting $cmd"
        
        # Replace the client creation pattern
        # FROM: token := os.Getenv("VIRTUOSO_API_TOKEN")
        #       if token == "" { return fmt.Errorf("VIRTUOSO_API_TOKEN environment variable is required") }
        #       baseURL := os.Getenv("VIRTUOSO_API_BASE_URL")
        #       if baseURL == "" { baseURL = "https://api-app2.virtuoso.qa/api" }
        #       client := virtuoso.NewClient(baseURL, token)
        # TO:   client := virtuoso.NewClient(cfg)
        
        # Create a temporary file
        tmp_file="${file}.tmp"
        
        # Use awk to replace the pattern
        awk '
        BEGIN { in_replacement = 0 }
        /token := os\.Getenv\("VIRTUOSO_API_TOKEN"\)/ {
            in_replacement = 1
            print "\t// Create Virtuoso client"
            print "\tclient := virtuoso.NewClient(cfg)"
            print ""
            next
        }
        in_replacement && /client := virtuoso\.NewClient/ {
            in_replacement = 0
            next
        }
        in_replacement && /^[[:space:]]*$/ {
            next
        }
        in_replacement && /^[[:space:]]*\/\// {
            next
        }
        in_replacement {
            next
        }
        { print }
        ' "$file" > "$tmp_file"
        
        # Replace the original file
        mv "$tmp_file" "$file"
        
        # Also remove the os import if it's no longer needed
        # Check if os is still used in the file
        if ! grep -q 'os\.' "$file" | grep -v 'import'; then
            sed -i '' '/"os"/d' "$file"
        fi
    else
        echo "Warning: File not found: $file"
    fi
done

# Also need to add the config import to the files
echo "Adding config import to adapted files..."

for cmd in "${VERSION_B_COMMANDS[@]}"; do
    file="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/$cmd"
    if [ -f "$file" ]; then
        # Check if config import already exists
        if ! grep -q '"github.com/marklovelady/api-cli-generator/pkg/config"' "$file"; then
            # Add config import after the standard library imports
            sed -i '' '/^import (/,/^)/ {
                /"strconv"/a\
\
	"github.com/marklovelady/api-cli-generator/pkg/config"
            }' "$file"
        fi
    fi
done

echo "Adaptation complete!"