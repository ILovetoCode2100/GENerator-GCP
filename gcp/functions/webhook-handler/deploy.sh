#!/bin/bash

# Deploy Webhook Handler Cloud Function
# Usage: ./deploy.sh [project-id]

set -e

PROJECT_ID=${1:-virtuoso-api}
FUNCTION_NAME="webhook-handler"
REGION="us-central1"
RUNTIME="python311"
MEMORY="256MB"
TIMEOUT="60s"

echo "Deploying ${FUNCTION_NAME} to project ${PROJECT_ID}..."

# Create Pub/Sub topics if they don't exist
echo "Creating Pub/Sub topics..."
gcloud pubsub topics create github-webhooks --project ${PROJECT_ID} 2>/dev/null || echo "Topic github-webhooks already exists"
gcloud pubsub topics create virtuoso-webhooks --project ${PROJECT_ID} 2>/dev/null || echo "Topic virtuoso-webhooks already exists"

# Deploy function
gcloud functions deploy ${FUNCTION_NAME} \
  --runtime ${RUNTIME} \
  --trigger-http \
  --allow-unauthenticated \
  --entry-point webhook_handler \
  --memory ${MEMORY} \
  --timeout ${TIMEOUT} \
  --region ${REGION} \
  --project ${PROJECT_ID} \
  --set-env-vars "GCP_PROJECT=${PROJECT_ID}" \
  --set-secrets "GITHUB_WEBHOOK_SECRET=github-webhook-secret:latest,VIRTUOSO_WEBHOOK_SECRET=virtuoso-webhook-secret:latest" \
  --max-instances 50 \
  --min-instances 0

echo "Function deployed successfully!"
echo "Webhook endpoint: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}"

# Create sample webhook secrets if they don't exist
read -p "Create sample webhook secrets in Secret Manager? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  # GitHub webhook secret
  echo -n "Enter GitHub webhook secret (or press Enter to generate): "
  read GITHUB_SECRET
  if [ -z "$GITHUB_SECRET" ]; then
    GITHUB_SECRET=$(openssl rand -hex 32)
    echo "Generated GitHub secret: $GITHUB_SECRET"
  fi
  echo -n "$GITHUB_SECRET" | gcloud secrets create github-webhook-secret \
    --data-file=- \
    --project ${PROJECT_ID} 2>/dev/null || echo "Secret github-webhook-secret already exists"

  # Virtuoso webhook secret
  echo -n "Enter Virtuoso webhook secret (or press Enter to generate): "
  read VIRTUOSO_SECRET
  if [ -z "$VIRTUOSO_SECRET" ]; then
    VIRTUOSO_SECRET=$(openssl rand -hex 32)
    echo "Generated Virtuoso secret: $VIRTUOSO_SECRET"
  fi
  echo -n "$VIRTUOSO_SECRET" | gcloud secrets create virtuoso-webhook-secret \
    --data-file=- \
    --project ${PROJECT_ID} 2>/dev/null || echo "Secret virtuoso-webhook-secret already exists"
fi
