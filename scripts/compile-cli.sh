#!/bin/bash
# compile-cli.sh - Compile the Virtuoso CLI

echo "Compiling Virtuoso CLI..."
echo "========================"

# Navigate to project directory
cd "$(dirname "$0")/.." || exit 1

# Clean previous build
echo "Cleaning previous build..."
rm -f bin/api-cli

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Build the CLI
echo "Building CLI..."
if go build -o bin/api-cli ./src/cmd; then
    echo "✅ Build successful!"
    echo ""
    # Make it executable
    chmod +x bin/api-cli
    
    # Show version
    echo "Testing CLI..."
    ./bin/api-cli --version
    
    echo ""
    echo "CLI is ready at: ./bin/api-cli"
    echo ""
    echo "Available commands:"
    ./bin/api-cli --help
else
    echo "❌ Build failed!"
    exit 1
fi
