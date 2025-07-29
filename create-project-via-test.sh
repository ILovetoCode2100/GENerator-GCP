#!/bin/bash

# Create a project by running a simple test (which auto-creates project/goal/journey/checkpoint)

PROJECT_NAME="${1:-My New Project via GCP API}"

echo "Creating project: $PROJECT_NAME"
echo ""

curl -X POST "https://virtuoso-api-5e22h3hywa-uc.a.run.app/api/v1/tests/run" \
  -H "X-API-Key: 6a54e405ab1277b555f13ccfcd68f32343a21debcb2f7fe12ce845ca8dfd5e2d" \
  -H "Content-Type: application/json" \
  -d "{
    \"definition\": {
      \"name\": \"$PROJECT_NAME - Initial Test\",
      \"description\": \"Created via GCP API\",
      \"steps\": [
        {
          \"action\": \"navigate\",
          \"url\": \"https://example.com\"
        },
        {
          \"action\": \"assert\",
          \"hint\": \"Example Domain\"
        }
      ],
      \"config\": {
        \"project_name\": \"$PROJECT_NAME\"
      }
    },
    \"dry_run\": false,
    \"execute\": false
  }" | jq .

echo ""
echo "Note: This creates a project with an initial test. The project_id will be in the response."
