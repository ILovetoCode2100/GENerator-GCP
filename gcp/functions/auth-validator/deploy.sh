#!/bin/bash

# Deploy Auth Validator Cloud Function
# Usage: ./deploy.sh [project-id]

set -e

PROJECT_ID=${1:-virtuoso-api}
FUNCTION_NAME="auth-validator"
REGION="us-central1"
RUNTIME="python311"
MEMORY="256MB"
TIMEOUT="10s"  # Fast response required
REDIS_HOST="10.0.0.3"  # Update with actual Memorystore IP

echo "Deploying ${FUNCTION_NAME} to project ${PROJECT_ID}..."

# Deploy function with high performance settings
gcloud functions deploy ${FUNCTION_NAME} \
  --runtime ${RUNTIME} \
  --trigger-http \
  --allow-unauthenticated \
  --entry-point auth_validator \
  --memory ${MEMORY} \
  --timeout ${TIMEOUT} \
  --region ${REGION} \
  --project ${PROJECT_ID} \
  --set-env-vars "GCP_PROJECT=${PROJECT_ID},REDIS_HOST=${REDIS_HOST}" \
  --max-instances 100 \
  --min-instances 1 \
  --concurrency 100

echo "Function deployed successfully!"
echo "Endpoint: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}"

# Create sample API keys collection structure
read -p "Create sample Firestore collections for API keys? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo "Creating Firestore collections..."

  # Create a Python script to set up collections
  cat > setup_collections.py << 'EOF'
import asyncio
from google.cloud import firestore
from datetime import datetime
import hashlib

async def setup_collections():
    db = firestore.AsyncClient()

    # Create plans collection
    plans = {
        'free': {
            'name': 'Free Plan',
            'rate_limit': 1000,
            'features': ['basic_api_access'],
            'resources': {
                'max_projects': 5,
                'max_tests_per_project': 10
            }
        },
        'pro': {
            'name': 'Pro Plan',
            'rate_limit': 10000,
            'features': ['basic_api_access', 'advanced_analytics', 'priority_support'],
            'resources': {
                'max_projects': 50,
                'max_tests_per_project': 100
            }
        },
        'enterprise': {
            'name': 'Enterprise Plan',
            'rate_limit': 100000,
            'features': ['basic_api_access', 'advanced_analytics', 'priority_support', 'custom_integrations'],
            'resources': {
                'max_projects': -1,  # unlimited
                'max_tests_per_project': -1
            }
        }
    }

    for plan_id, plan_data in plans.items():
        await db.collection('plans').document(plan_id).set(plan_data)

    print("Sample collections created successfully!")

# Run the setup
asyncio.run(setup_collections())
EOF

  python setup_collections.py
  rm setup_collections.py
fi

echo ""
echo "Auth Validator deployment complete!"
echo ""
echo "Usage example:"
echo "curl -X POST https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME} \\"
echo "  -H 'Authorization: Bearer YOUR_API_KEY' \\"
echo "  -H 'Content-Type: application/json'"
