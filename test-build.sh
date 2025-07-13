#!/bin/bash

echo "Testing build of merged Version A..."

cd /Users/marklovelady/_dev/virtuoso-api-cli-generator

# Update dependencies
echo "Updating Go dependencies..."
go mod tidy

# Try to build
echo "Building api-cli..."
if go build -o bin/api-cli .; then
    echo "✅ Build successful!"
    echo "Binary created at: bin/api-cli"
else
    echo "❌ Build failed!"
    echo "Common issues to fix:"
    echo "1. Version B commands use different client constructor"
    echo "2. Missing config parameter in command files"
    echo "3. Import path mismatches"
fi