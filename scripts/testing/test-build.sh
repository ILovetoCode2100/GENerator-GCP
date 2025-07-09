#!/bin/bash
# test-build.sh - Test that the CLI builds successfully

echo "Testing Virtuoso CLI build..."
echo "============================"

# Navigate to project directory
cd "$(dirname "$0")/.." || exit 1

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Build the CLI
echo "Building CLI..."
if go build -o bin/api-cli ./src/cmd; then
    echo "✅ Build successful!"
    echo ""
    echo "Testing CLI execution..."
    ./bin/api-cli --version
    echo ""
    ./bin/api-cli --help
else
    echo "❌ Build failed!"
    exit 1
fi

echo ""
echo "Next steps:"
echo "1. Provide API endpoint details"
echo "2. Run: ./bin/api-cli create-structure --file examples/test-structure.json"
