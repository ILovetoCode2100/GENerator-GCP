#!/bin/bash

# Read config file to get API credentials
CONFIG_FILE="$HOME/.api-cli/virtuoso-config.yaml"

# Extract values from YAML config
AUTH_TOKEN=$(grep "auth_token:" "$CONFIG_FILE" | awk '{print $2}')
BASE_URL=$(grep "base_url:" "$CONFIG_FILE" | awk '{print $2}')

# Checkpoint ID - using the one from the user's report
CHECKPOINT_ID=1682332

echo "Analyzing steps for checkpoint $CHECKPOINT_ID..."
echo ""

# First, try to get checkpoint details
echo "Getting checkpoint details..."
CHECKPOINT_RESPONSE=$(curl -s -X GET \
  "${BASE_URL}/checkpoints/${CHECKPOINT_ID}" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-Virtuoso-Client-ID: api-cli-generator" \
  -H "X-Virtuoso-Client-Name: api-cli-generator")

echo "Checkpoint response:"
echo "$CHECKPOINT_RESPONSE" | jq '.' 2>/dev/null || echo "$CHECKPOINT_RESPONSE"
echo ""

# Try different endpoints to get steps
echo "Trying to get steps with different endpoints..."
echo ""

# Try endpoint 1: /checkpoints/{id}/teststeps
echo "1. Trying /checkpoints/${CHECKPOINT_ID}/teststeps"
RESPONSE1=$(curl -s -X GET \
  "${BASE_URL}/checkpoints/${CHECKPOINT_ID}/teststeps" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-Virtuoso-Client-ID: api-cli-generator" \
  -H "X-Virtuoso-Client-Name: api-cli-generator")

if echo "$RESPONSE1" | jq -e . >/dev/null 2>&1; then
    echo "Success! Found steps:"
    echo "$RESPONSE1" | jq '.'
else
    echo "Failed. Response: $RESPONSE1"
fi
echo ""

# Try endpoint 2: /checkpoints/{id}/steps
echo "2. Trying /checkpoints/${CHECKPOINT_ID}/steps"
RESPONSE2=$(curl -s -X GET \
  "${BASE_URL}/checkpoints/${CHECKPOINT_ID}/steps" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-Virtuoso-Client-ID: api-cli-generator" \
  -H "X-Virtuoso-Client-Name: api-cli-generator")

if echo "$RESPONSE2" | jq -e . >/dev/null 2>&1; then
    echo "Success! Found steps:"
    echo "$RESPONSE2" | jq '.'
else
    echo "Failed. Response: $RESPONSE2"
fi
echo ""

# Try endpoint 3: /teststeps with query parameter
echo "3. Trying /teststeps?checkpointId=${CHECKPOINT_ID}"
RESPONSE3=$(curl -s -X GET \
  "${BASE_URL}/teststeps?checkpointId=${CHECKPOINT_ID}" \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -H "Content-Type: application/json" \
  -H "X-Virtuoso-Client-ID: api-cli-generator" \
  -H "X-Virtuoso-Client-Name: api-cli-generator")

if echo "$RESPONSE3" | jq -e . >/dev/null 2>&1; then
    echo "Success! Found steps:"
    echo "$RESPONSE3" | jq '.'
else
    echo "Failed. Response: $RESPONSE3"
fi
