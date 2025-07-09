#!/bin/bash
# setup-virtuoso.sh - Quick setup script for Virtuoso API CLI

# Virtuoso API Configuration
export VIRTUOSO_API_BASE_URL="https://api-app2.virtuoso.qa/api"
export VIRTUOSO_API_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"
export VIRTUOSO_ORGANIZATION_ID="2242"
export VIRTUOSO_HEADERS_X_VIRTUOSO_CLIENT_ID="api-cli-generator"
export VIRTUOSO_HEADERS_X_VIRTUOSO_CLIENT_NAME="api-cli-generator"

# Optional: Set output format
export VIRTUOSO_OUTPUT_DEFAULT_FORMAT="json"  # or "human", "yaml", "ai"

echo "✅ Virtuoso API environment configured"
echo ""
echo "Base URL: $VIRTUOSO_API_BASE_URL"
echo "Org ID: $VIRTUOSO_ORGANIZATION_ID"
echo "Client ID: $VIRTUOSO_HEADERS_X_VIRTUOSO_CLIENT_ID"
echo ""
echo "You can now run the CLI commands:"
echo "  ./bin/api-cli create-structure --file structure.json"
echo "  ./bin/api-cli add-step <checkpoint-id> <step-type> [options]"
