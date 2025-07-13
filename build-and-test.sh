#!/bin/bash

# Build and test the merged Version A

echo "Building merged Version A..."

cd /Users/marklovelady/_dev/virtuoso-api-cli-generator

# Update dependencies
echo "Updating Go dependencies..."
go mod tidy

# Build the binary
echo "Building api-cli..."
go build -o bin/api-cli .

if [ $? -eq 0 ]; then
    echo "✓ Build successful!"
    echo ""
    echo "Binary created at: bin/api-cli"
    echo ""
    echo "To test the new commands, run:"
    echo "  export VIRTUOSO_API_BASE_URL='https://api-app2.virtuoso.qa/api'"
    echo "  export VIRTUOSO_API_TOKEN='your-token-here'"
    echo "  ./test-all-commands-variations.sh"
    echo "  ./test-new-commands.sh"
else
    echo "✗ Build failed! Please check the errors above."
    echo ""
    echo "Common issues:"
    echo "1. Missing client methods - run ./integrate-client-methods.sh"
    echo "2. Missing command registrations - run ./integrate-commands.sh"
    echo "3. Check import statements in the new command files"
fi