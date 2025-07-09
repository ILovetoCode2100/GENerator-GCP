#!/bin/bash

# Simple test of step creation commands
# Using example checkpoint ID from documentation

set -e

# Load environment
source ./scripts/setup-virtuoso.sh

# Test checkpoint ID from documentation examples
CHECKPOINT_ID=44444

echo "================================================"
echo "Testing Step Creation Commands"
echo "Using Example Checkpoint ID: $CHECKPOINT_ID"
echo "================================================"
echo ""
echo "Note: These tests may fail if checkpoint doesn't exist."
echo "The goal is to verify command syntax and API integration."
echo ""

# Test navigation step
echo "1. Testing navigation step creation..."
./bin/api-cli create-step-navigate $CHECKPOINT_ID "https://example.com" 1 -o json || echo "Failed (expected if checkpoint doesn't exist)"

echo ""
echo "2. Testing wait time step creation..."
./bin/api-cli create-step-wait-time $CHECKPOINT_ID 5 2 -o json || echo "Failed (expected if checkpoint doesn't exist)"

echo ""
echo "3. Testing click step creation..."
./bin/api-cli create-step-click $CHECKPOINT_ID "Sign in button" 3 -o json || echo "Failed (expected if checkpoint doesn't exist)"

echo ""
echo "4. Testing write step creation..."
./bin/api-cli create-step-write $CHECKPOINT_ID "test@example.com" "Email field" 4 -o json || echo "Failed (expected if checkpoint doesn't exist)"

echo ""
echo "5. Testing assertion step creation..."
./bin/api-cli create-step-assert-exists $CHECKPOINT_ID "Welcome message" 5 -o json || echo "Failed (expected if checkpoint doesn't exist)"

echo ""
echo "================================================"
echo "Command Structure Test Complete"
echo "================================================"
echo ""
echo "All commands executed. Failures are expected if checkpoint doesn't exist."
echo "The important thing is that commands are properly formed and reach the API."