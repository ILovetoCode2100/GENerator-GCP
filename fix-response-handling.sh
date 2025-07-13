#!/bin/bash

# Fix response handling in Version B command files
# Version B commands expect StepResponse but client methods now return int

echo "Fixing response handling in Version B command files..."

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
        echo "Fixing response handling in $cmd"
        
        # Replace response, err := with stepID, err :=
        sed -i '' 's/response, err :=/stepID, err :=/g' "$file"
        
        # Replace response.StepID with stepID
        sed -i '' 's/response\.StepID/stepID/g' "$file"
        
        # Replace response.CheckpointID with checkpointID (from args)
        sed -i '' 's/response\.CheckpointID/checkpointID/g' "$file"
        
        # Remove response.Action, response.Value, response.Message references
        sed -i '' '/response\.Action/d' "$file"
        sed -i '' '/response\.Value/d' "$file"
        sed -i '' '/response\.Message/d' "$file"
        sed -i '' '/if response\.Message/,+2d' "$file"
        
        # Fix the JSON/YAML output to use a simple structure
        sed -i '' 's/json\.MarshalIndent(response/json.MarshalIndent(map[string]interface{}{"stepId": stepID, "checkpointId": checkpointID}/g' "$file"
        sed -i '' 's/yaml\.Marshal(response)/yaml.Marshal(map[string]interface{}{"stepId": stepID, "checkpointId": checkpointID})/g' "$file"
        
    else
        echo "Warning: File not found: $file"
    fi
done

echo "Response handling fixes complete!"