#!/bin/bash

# GCP API endpoint
GCP_API_URL="https://virtuoso-api-5e22h3hywa-uc.a.run.app"

# Use the GCP API key (from Secret Manager)
GCP_API_KEY="6a54e405ab1277b555f13ccfcd68f32343a21debcb2f7fe12ce845ca8dfd5e2d"

# Read Virtuoso API token from config
if [ -f "$HOME/.api-cli/virtuoso-config.yaml" ]; then
    VIRTUOSO_TOKEN=$(grep 'auth_token:' "$HOME/.api-cli/virtuoso-config.yaml" | awk '{print $2}' | tr -d '"')
else
    VIRTUOSO_TOKEN="${VIRTUOSO_API_TOKEN}"
fi

echo "Sending Rocketshop test to GCP API..."
echo "Endpoint: $GCP_API_URL"
echo "Using GCP API Key: ${GCP_API_KEY:0:10}..."
echo ""

# Send the YAML test file via the GCP API
curl -X POST "$GCP_API_URL/api/v1/tests/run" \
    -H "X-API-Key: $GCP_API_KEY" \
    -H "Content-Type: application/json" \
    -d @- << 'EOF'
{
  "definition": {
    "name": "Rocketshop Purchase Flow - GCP Deployment",
    "description": "E-commerce test sent via GCP Cloud Run API",
    "steps": [
      {
        "action": "navigate",
        "url": "https://rocketshop.virtuoso.qa"
      },
      {
        "action": "assert",
        "hint": "Border Not Found"
      },
      {
        "action": "wait",
        "time": 20000
      },
      {
        "action": "click",
        "hint": "Add to Bag"
      },
      {
        "action": "click",
        "hint": "Shopping Bag"
      },
      {
        "action": "assert",
        "hint": "Shopping Bag",
        "element_type": "h1"
      },
      {
        "action": "click",
        "hint": "Go to Checkout"
      },
      {
        "action": "write",
        "hint": "Full name",
        "text": "John Doe"
      },
      {
        "action": "write",
        "hint": "Email",
        "text": "johndoe@example.com"
      },
      {
        "action": "write",
        "hint": "Address",
        "text": "123 Elm Street"
      },
      {
        "action": "write",
        "hint": "Phone numbers",
        "text": "555-1234"
      },
      {
        "action": "write",
        "hint": "ZIP code",
        "text": "90210"
      },
      {
        "action": "write",
        "hint": "Card number",
        "text": "4111 1111 1111 1111"
      },
      {
        "action": "write",
        "hint": "xxx",
        "text": "234"
      },
      {
        "action": "click",
        "hint": "Confirm and Pay"
      },
      {
        "action": "wait",
        "time": 20000
      },
      {
        "action": "assert",
        "hint": "Purchase Confirmed!",
        "element_type": "h3"
      },
      {
        "action": "click",
        "hint": "Download Confirmation"
      }
    ],
    "config": {
      "project_id": 9349,
      "checkpoint_id": 1682637
    }
  },
  "execute": true,
  "environment": "production"
}
EOF

echo ""
echo "Test sent to GCP Cloud Run API!"
