#!/bin/bash

# Deploy Health Check Cloud Function
# Usage: ./deploy.sh [project-id]

set -e

PROJECT_ID=${1:-virtuoso-api}
FUNCTION_NAME="health-check"
REGION="us-central1"
RUNTIME="python311"
MEMORY="256MB"
TIMEOUT="30s"
REDIS_HOST="10.0.0.3"  # Update with actual Memorystore IP
CLOUD_RUN_URL="https://virtuoso-api-service-abcdef.run.app"  # Update with actual URL

echo "Deploying ${FUNCTION_NAME} to project ${PROJECT_ID}..."

gcloud functions deploy ${FUNCTION_NAME} \
  --runtime ${RUNTIME} \
  --trigger-http \
  --allow-unauthenticated \
  --entry-point health_check \
  --memory ${MEMORY} \
  --timeout ${TIMEOUT} \
  --region ${REGION} \
  --project ${PROJECT_ID} \
  --set-env-vars "GCP_PROJECT=${PROJECT_ID},REDIS_HOST=${REDIS_HOST},CLOUD_RUN_URL=${CLOUD_RUN_URL}" \
  --max-instances 10 \
  --min-instances 0

echo "Function deployed successfully!"
echo "Endpoint: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}"

# Set up Cloud Scheduler for periodic health checks (optional)
read -p "Set up Cloud Scheduler for periodic health checks? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  gcloud scheduler jobs create http ${FUNCTION_NAME}-schedule \
    --location ${REGION} \
    --schedule "*/5 * * * *" \
    --uri "https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}" \
    --http-method GET \
    --attempt-deadline 30s \
    --project ${PROJECT_ID} || echo "Scheduler job might already exist"
fi
