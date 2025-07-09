#!/bin/bash

# Script to create comprehensive test suite in Virtuoso
# This creates: 1 project, 3 goals, multiple journeys, checkpoints and steps

echo "🚀 Creating Comprehensive E-Commerce Test Suite"
echo "================================================"

# Change to project directory
cd /Users/marklovelady/_dev/claude-desktop/projects/api-cli-generator

# Source the Virtuoso setup
echo "📋 Setting up Virtuoso environment..."
source ./scripts/setup-virtuoso.sh

# First, let's do a dry run to preview what will be created
echo ""
echo "👀 Preview Mode - Showing what will be created:"
echo "------------------------------------------------"
./bin/api-cli create-structure --file examples/comprehensive-test-suite.yaml --dry-run

# Ask for confirmation
echo ""
echo "📊 Summary:"
echo "- 1 Project: 'Comprehensive E-Commerce Test Suite'"
echo "- 3 Goals: Homepage, Product Catalog, Checkout Process"
echo "- 10 Journeys total (3 + 3 + 4)"
echo "- 27 Checkpoints total"
echo "- 100+ Steps including all available step types"
echo ""
read -p "Do you want to create this test structure? (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]
then
    echo ""
    echo "🔨 Creating test structure..."
    echo "------------------------------------------------"
    
    # Run the actual creation with verbose output
    ./bin/api-cli create-structure --file examples/comprehensive-test-suite.yaml --verbose
    
    echo ""
    echo "✅ Test structure creation complete!"
    echo ""
    echo "📋 What was created:"
    echo "- 1 Project with comprehensive test coverage"
    echo "- 3 Goals covering different aspects of the e-commerce site"
    echo "- Multiple journeys per goal including:"
    echo "  - Basic flows"
    echo "  - Mobile testing"
    echo "  - Performance testing"
    echo "  - Complete step type showcase (4th journey in Checkout goal)"
    echo ""
    echo "🎯 The 'Complete Step Types Showcase' journey demonstrates ALL available steps:"
    echo "  - Navigation & control steps"
    echo "  - Mouse interactions (click, hover, drag)"
    echo "  - Form inputs (text, dropdowns, checkboxes)"
    echo "  - File operations"
    echo "  - Keyboard actions"
    echo "  - Scrolling variations"
    echo "  - Wait conditions"
    echo "  - All assertion types"
    echo "  - Advanced features (JS execution, cookies, alerts)"
    echo ""
    echo "💡 Next steps:"
    echo "1. Log into Virtuoso to see your created test suite"
    echo "2. Run the tests to verify they work correctly"
    echo "3. Customize the selectors and URLs for your actual application"
else
    echo "❌ Creation cancelled"
fi