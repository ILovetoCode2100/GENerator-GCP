#!/bin/bash

# This script ACTUALLY creates a project by calling the CLI through the GCP API

PROJECT_NAME="${1:-Real Project via GCP $(date +%H:%M:%S)}"

echo "Creating REAL project: $PROJECT_NAME"
echo ""

# The GCP API needs to execute actual CLI commands
# Let's use the commands/execute endpoint if it exists

echo "Method 1: Using command execution endpoint..."
curl -X POST "https://virtuoso-api-5e22h3hywa-uc.a.run.app/api/v1/commands/batch" \
  -H "X-API-Key: 6a54e405ab1277b555f13ccfcd68f32343a21debcb2f7fe12ce845ca8dfd5e2d" \
  -H "Content-Type: application/json" \
  -d "{
    \"commands\": [
      {
        \"command\": \"create-project\",
        \"args\": [\"$PROJECT_NAME\"]
      }
    ]
  }" | jq .

echo ""
echo "The issue: The GCP API is returning mock data, not actually creating projects!"
echo "The project IDs like 'proj_96fcd0d7' are fake UUIDs, not real Virtuoso IDs."
echo ""
echo "Solution: The GCP API needs to be updated to actually execute CLI commands"
echo "instead of returning mock responses."
