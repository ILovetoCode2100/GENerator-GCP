#!/bin/bash

# Fix undefined types and variables in command files

echo "Fixing undefined types and variables..."

# Fix create-step-execute-script.go
file="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/create-step-execute-script.go"
if [ -f "$file" ]; then
    echo "Fixing: create-step-execute-script.go"
    # Remove any references to response variable that weren't properly converted
    sed -i '' 's/response\./stepID/g' "$file"
    # Remove StepResponse type references
    sed -i '' 's/virtuoso\.StepResponse/int/g' "$file"
fi

# Fix create-step-mouse-move-by.go
file="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/create-step-mouse-move-by.go"
if [ -f "$file" ]; then
    echo "Fixing: create-step-mouse-move-by.go"
    sed -i '' 's/virtuoso\.StepResponse/int/g' "$file"
fi

# Fix create-step-mouse-move-to.go
file="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/create-step-mouse-move-to.go"
if [ -f "$file" ]; then
    echo "Fixing: create-step-mouse-move-to.go"
    sed -i '' 's/virtuoso\.StepResponse/int/g' "$file"
fi

# Fix create-step-click.go
file="/Users/marklovelady/_dev/virtuoso-api-cli-generator/src/cmd/create-step-click.go"
if [ -f "$file" ]; then
    echo "Fixing: create-step-click.go"
    # Remove the response variable declaration line
    sed -i '' '/response.*:=.*virtuoso\.StepResponse/d' "$file"
    sed -i '' 's/virtuoso\.StepResponse/int/g' "$file"
    
    # Fix the variable references
    sed -i '' 's/response, err/stepID, err/g' "$file"
fi

echo "Type fixes complete!"