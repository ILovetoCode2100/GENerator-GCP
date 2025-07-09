#!/bin/bash
# test-virtuoso-api.sh - Test Virtuoso API connection

echo "Testing Virtuoso API Connection..."
echo "================================="

# Configuration
BASE_URL="https://api-app2.virtuoso.qa/api"
AUTH_TOKEN="f7a55516-5cc4-4529-b2ae-8e106a7d164e"
CLIENT_ID="api-cli-generator"
CLIENT_NAME="api-cli-generator"

# Test API connection
echo "Testing API endpoint..."
response=$(curl -s -w "\n%{http_code}" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "X-Virtuoso-Client-ID: $CLIENT_ID" \
  -H "X-Virtuoso-Client-Name: $CLIENT_NAME" \
  -H "Content-Type: application/json" \
  "$BASE_URL/health" 2>/dev/null)

http_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | head -n -1)

echo "HTTP Status: $http_code"
echo "Response: $body"

if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 204 ]; then
  echo "✅ API connection successful!"
else
  echo "❌ API connection failed. Please check credentials."
fi

echo ""
echo "To test project creation, update the script with the correct endpoint."
