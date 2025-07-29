#!/bin/bash

# Create a new Virtuoso project via GCP API

PROJECT_NAME="${1:-My New Project via GCP API}"

echo "Creating project: $PROJECT_NAME"
echo ""

curl -X POST "https://virtuoso-api-5e22h3hywa-uc.a.run.app/api/v1/commands/step/create/project" \
  -H "X-API-Key: 6a54e405ab1277b555f13ccfcd68f32343a21debcb2f7fe12ce845ca8dfd5e2d" \
  -H "Content-Type: application/json" \
  -d "{
    \"args\": [\"$PROJECT_NAME\"]
  }" | jq .

echo ""
echo "Usage: ./create-project-curl.sh \"Your Project Name\""
