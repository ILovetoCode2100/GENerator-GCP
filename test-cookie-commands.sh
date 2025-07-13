#!/bin/bash

# Test script for the new cookie commands
# This script demonstrates how to use the new cookie functionality

echo "Testing Virtuoso API CLI Cookie Commands"
echo "========================================"

# Check if the binary exists
if [[ ! -f "./bin/api-cli" ]]; then
    echo "Error: API CLI binary not found. Please run 'go build -o bin/api-cli' first."
    exit 1
fi

# Check if API token is set
if [[ -z "$VIRTUOSO_API_TOKEN" ]]; then
    echo "Warning: VIRTUOSO_API_TOKEN environment variable is not set."
    echo "You need to set this to actually call the API."
    echo "Example: export VIRTUOSO_API_TOKEN='your-token-here'"
    echo ""
fi

echo "1. Testing create-step-cookie-create command help:"
./bin/api-cli create-step-cookie-create --help
echo ""

echo "2. Testing create-step-cookie-wipe-all command help:"
./bin/api-cli create-step-cookie-wipe-all --help
echo ""

echo "3. Example usage (requires valid CHECKPOINT_ID and VIRTUOSO_API_TOKEN):"
echo "   ./bin/api-cli create-step-cookie-create 1678318 \"session\" \"abc123\" 1"
echo "   ./bin/api-cli create-step-cookie-wipe-all 1678318 2"
echo ""

echo "4. Example with different output formats:"
echo "   ./bin/api-cli create-step-cookie-create 1678318 \"session\" \"abc123\" 1 -o json"
echo "   ./bin/api-cli create-step-cookie-create 1678318 \"session\" \"abc123\" 1 -o yaml"
echo "   ./bin/api-cli create-step-cookie-create 1678318 \"session\" \"abc123\" 1 -o ai"
echo ""

echo "JSON Request bodies that will be sent:"
echo "--------------------------------------"
echo "For create-step-cookie-create:"
echo '{"action": "ENVIRONMENT", "value": "abc123", "meta": {"type": "ADD", "name": "session"}, "position": 1}'
echo ""
echo "For create-step-cookie-wipe-all:"
echo '{"action": "ENVIRONMENT", "meta": {"type": "CLEAR"}, "position": 2}'
echo ""

echo "Test completed successfully!"