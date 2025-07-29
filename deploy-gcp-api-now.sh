#!/bin/bash

echo "Deploying API to GCP Cloud Run with authentication disabled..."
echo ""

# First, let's create a simple test to verify our changes
cd api

# Deploy with explicit revision suffix to force new deployment
REVISION_SUFFIX="noauth-$(date +%s)"
echo "Deploying revision: $REVISION_SUFFIX"

gcloud run deploy virtuoso-api \
  --source . \
  --region=us-central1 \
  --platform=managed \
  --allow-unauthenticated \
  --revision-suffix=$REVISION_SUFFIX \
  --set-env-vars="AUTH_ENABLED=false,SKIP_AUTH=true,VIRTUOSO_API_KEY=f7a55516-5cc4-4529-b2ae-8e106a7d164e" \
  --timeout=600 \
  --memory=1Gi \
  --min-instances=1 \
  --max-instances=10 \
  --cpu=2 \
  --concurrency=1000

echo ""
echo "Deployment complete! Testing the API..."
echo ""

# Wait for deployment to stabilize
sleep 15

# Test the API
echo "Testing health endpoint:"
curl -s https://virtuoso-api-5e22h3hywa-uc.a.run.app/health | jq .

echo ""
echo "Testing test run endpoint without auth:"
curl -X POST "https://virtuoso-api-5e22h3hywa-uc.a.run.app/api/v1/tests/run" \
  -H "Content-Type: application/json" \
  -d '{
    "definition": {
      "name": "Test Auth Disabled",
      "steps": [{"action": "navigate", "url": "https://example.com"}]
    },
    "dry_run": true
  }' | jq .

echo ""
echo "If you still see auth errors, the deployment might need more time to propagate."
