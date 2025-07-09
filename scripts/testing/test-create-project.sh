#!/bin/bash
# test-create-project.sh - Test the create-project command

echo "Testing Virtuoso CLI - Create Project Command"
echo "==========================================="

# Set up environment
source ./scripts/setup-virtuoso.sh

# Build the CLI
echo "Building CLI..."
go build -o bin/api-cli ./src/cmd

# Test 1: Create project with human output
echo -e "\n1. Testing create-project with default output:"
./bin/api-cli create-project "Test Project $(date +%s)"

# Test 2: Create project with JSON output
echo -e "\n2. Testing create-project with JSON output:"
./bin/api-cli create-project "Test Project JSON $(date +%s)" -o json

# Test 3: Create project with AI output
echo -e "\n3. Testing create-project with AI output:"
./bin/api-cli create-project "Test Project AI $(date +%s)" -o ai

# Test 4: Create project with description
echo -e "\n4. Testing create-project with description:"
./bin/api-cli create-project "Test Project Desc $(date +%s)" --description "This is a test project"

# Test 5: Test error handling (missing name)
echo -e "\n5. Testing error handling (missing name):"
./bin/api-cli create-project || echo "✅ Error handling works"

echo -e "\n✅ Test complete!"
