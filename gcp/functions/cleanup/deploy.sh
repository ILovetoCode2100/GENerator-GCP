#!/bin/bash

# Deploy Cleanup Cloud Function
# Usage: ./deploy.sh [project-id]

set -e

PROJECT_ID=${1:-virtuoso-api}
FUNCTION_NAME="cleanup"
REGION="us-central1"
RUNTIME="python311"
MEMORY="512MB"
TIMEOUT="540s"  # 9 minutes
CLEANUP_BUCKET="${PROJECT_ID}-archives"

echo "Deploying ${FUNCTION_NAME} to project ${PROJECT_ID}..."

# Create archive bucket if it doesn't exist
echo "Creating archive bucket..."
gsutil mb -p ${PROJECT_ID} -c STANDARD -l ${REGION} gs://${CLEANUP_BUCKET} 2>/dev/null || echo "Bucket ${CLEANUP_BUCKET} already exists"

# Create temp bucket if it doesn't exist
echo "Creating temp bucket..."
gsutil mb -p ${PROJECT_ID} -c STANDARD -l ${REGION} gs://${PROJECT_ID}-temp 2>/dev/null || echo "Bucket ${PROJECT_ID}-temp already exists"

# Deploy function
gcloud functions deploy ${FUNCTION_NAME} \
  --runtime ${RUNTIME} \
  --trigger-http \
  --no-allow-unauthenticated \
  --entry-point cleanup \
  --memory ${MEMORY} \
  --timeout ${TIMEOUT} \
  --region ${REGION} \
  --project ${PROJECT_ID} \
  --set-env-vars "GCP_PROJECT=${PROJECT_ID},CLEANUP_BUCKET=${CLEANUP_BUCKET}" \
  --max-instances 1 \
  --min-instances 0

echo "Function deployed successfully!"
echo "Endpoint: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}"

# Set up Cloud Scheduler for daily cleanup
read -p "Set up Cloud Scheduler for daily cleanup? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  # Create service account for scheduler
  gcloud iam service-accounts create cleanup-scheduler \
    --display-name "Cleanup Scheduler" \
    --project ${PROJECT_ID} 2>/dev/null || echo "Service account already exists"

  # Grant invoker permission
  gcloud functions add-iam-policy-binding ${FUNCTION_NAME} \
    --member="serviceAccount:cleanup-scheduler@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/cloudfunctions.invoker" \
    --region=${REGION} \
    --project=${PROJECT_ID}

  # Create scheduler job
  gcloud scheduler jobs create http ${FUNCTION_NAME}-daily \
    --location ${REGION} \
    --schedule "0 2 * * *" \
    --uri "https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}" \
    --http-method POST \
    --oidc-service-account-email "cleanup-scheduler@${PROJECT_ID}.iam.gserviceaccount.com" \
    --attempt-deadline 540s \
    --project ${PROJECT_ID} || echo "Scheduler job might already exist"

  echo "Daily cleanup scheduled for 2:00 AM"
fi
