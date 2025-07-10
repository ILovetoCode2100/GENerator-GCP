#!/bin/bash

# Test script for pick-text and upload commands with modern session context

echo "Testing pick-text and upload commands with modern session context..."
echo "================================================"

# Build the CLI
echo "Building CLI..."
make build

# Set a test checkpoint for session context
echo -e "\nSetting test checkpoint context..."
./bin/api-cli set-checkpoint 1678318

echo -e "\n1. Testing create-step-pick-text - Modern syntax with position"
./bin/api-cli create-step-pick-text "Country dropdown" "United States" 1 -o json

echo -e "\n2. Testing create-step-pick-text - Modern syntax without position (auto-increment)"
./bin/api-cli create-step-pick-text "#subscription-select" "Premium Plan" -o json

echo -e "\n3. Testing create-step-pick-text - Override checkpoint"
./bin/api-cli create-step-pick-text "State select" "California" 3 --checkpoint 1678319 -o json

echo -e "\n4. Testing create-step-pick-text - Legacy syntax"
./bin/api-cli create-step-pick-text 1678320 "Basic Plan" "Plan dropdown" 1 -o json

echo -e "\n5. Testing create-step-upload - Modern syntax with position"
./bin/api-cli create-step-upload "file upload" "document.pdf" 1 -o yaml

echo -e "\n6. Testing create-step-upload - Modern syntax without position (auto-increment)"
./bin/api-cli create-step-upload "#file-input" "image.jpg" -o yaml

echo -e "\n7. Testing create-step-upload - Override checkpoint"
./bin/api-cli create-step-upload "input[type='file']" "data.csv" 3 --checkpoint 1678319 -o yaml

echo -e "\n8. Testing create-step-upload - Legacy syntax"
./bin/api-cli create-step-upload 1678320 "report.xlsx" "Upload button" 2 -o yaml

echo -e "\n9. Testing output formats"
echo -e "\n   Human format:"
./bin/api-cli create-step-pick-text "Format test" "Human" 10

echo -e "\n   AI format:"
./bin/api-cli create-step-upload "AI test" "test.txt" 11 -o ai

echo -e "\nAll tests completed!"