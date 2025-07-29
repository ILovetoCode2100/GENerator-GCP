#!/bin/bash

echo "Deploying GCP API without authentication..."

# Build and deploy from source
gcloud run deploy virtuoso-api \
  --source=api \
  --region=us-central1 \
  --platform=managed \
  --allow-unauthenticated \
  --set-env-vars="AUTH_ENABLED=false,SKIP_AUTH=true,VIRTUOSO_API_KEY=f7a55516-5cc4-4529-b2ae-8e106a7d164e" \
  --quiet

echo ""
echo "Deployment complete!"
echo "Testing the API..."

# Wait for deployment
sleep 5

# Test the API
curl -X POST "https://virtuoso-api-5e22h3hywa-uc.a.run.app/api/v1/tests/run" \
  -H "Content-Type: application/json" \
  -d '{
    "definition": {
      "name": "Test API",
      "steps": [{"action": "navigate", "url": "https://example.com"}]
    }
  }' | jq .

echo ""
echo "API is ready for use without authentication!"
