#!/bin/bash

# Deploy Analytics Cloud Function
# Usage: ./deploy.sh [project-id]

set -e

PROJECT_ID=${1:-virtuoso-api}
FUNCTION_NAME="analytics"
REGION="us-central1"
RUNTIME="python311"
MEMORY="512MB"
TIMEOUT="300s"  # 5 minutes
BIGQUERY_DATASET="virtuoso_analytics"
ANALYTICS_BUCKET="${PROJECT_ID}-analytics"

echo "Deploying ${FUNCTION_NAME} to project ${PROJECT_ID}..."

# Create BigQuery dataset if it doesn't exist
echo "Creating BigQuery dataset..."
bq mk --dataset --location=${REGION} ${PROJECT_ID}:${BIGQUERY_DATASET} 2>/dev/null || echo "Dataset ${BIGQUERY_DATASET} already exists"

# Create tables
echo "Creating BigQuery tables..."
# Daily reports table
bq mk --table \
  ${PROJECT_ID}:${BIGQUERY_DATASET}.daily_reports \
  report_type:STRING,generated_at:TIMESTAMP,period:JSON,command_usage:JSON,api_metrics:JSON,user_analytics:JSON,processed_at:TIMESTAMP \
  2>/dev/null || echo "Table daily_reports already exists"

# Weekly reports table
bq mk --table \
  ${PROJECT_ID}:${BIGQUERY_DATASET}.weekly_reports \
  report_type:STRING,generated_at:TIMESTAMP,period:JSON,command_usage:JSON,api_metrics:JSON,user_analytics:JSON,processed_at:TIMESTAMP \
  2>/dev/null || echo "Table weekly_reports already exists"

# Monthly reports table
bq mk --table \
  ${PROJECT_ID}:${BIGQUERY_DATASET}.monthly_reports \
  report_type:STRING,generated_at:TIMESTAMP,period:JSON,command_usage:JSON,api_metrics:JSON,user_analytics:JSON,processed_at:TIMESTAMP \
  2>/dev/null || echo "Table monthly_reports already exists"

# Create analytics bucket if it doesn't exist
echo "Creating analytics bucket..."
gsutil mb -p ${PROJECT_ID} -c STANDARD -l ${REGION} gs://${ANALYTICS_BUCKET} 2>/dev/null || echo "Bucket ${ANALYTICS_BUCKET} already exists"

# Deploy function
gcloud functions deploy ${FUNCTION_NAME} \
  --runtime ${RUNTIME} \
  --trigger-http \
  --no-allow-unauthenticated \
  --entry-point analytics \
  --memory ${MEMORY} \
  --timeout ${TIMEOUT} \
  --region ${REGION} \
  --project ${PROJECT_ID} \
  --set-env-vars "GCP_PROJECT=${PROJECT_ID},BIGQUERY_DATASET=${BIGQUERY_DATASET},ANALYTICS_BUCKET=${ANALYTICS_BUCKET}" \
  --max-instances 5 \
  --min-instances 0

echo "Function deployed successfully!"
echo "Endpoint: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}"

# Set up Cloud Scheduler for periodic analytics
read -p "Set up Cloud Scheduler for periodic analytics? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  # Create service account for scheduler
  gcloud iam service-accounts create analytics-scheduler \
    --display-name "Analytics Scheduler" \
    --project ${PROJECT_ID} 2>/dev/null || echo "Service account already exists"

  # Grant invoker permission
  gcloud functions add-iam-policy-binding ${FUNCTION_NAME} \
    --member="serviceAccount:analytics-scheduler@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/cloudfunctions.invoker" \
    --region=${REGION} \
    --project=${PROJECT_ID}

  # Create daily analytics job
  gcloud scheduler jobs create http ${FUNCTION_NAME}-daily \
    --location ${REGION} \
    --schedule "0 1 * * *" \
    --uri "https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}?type=comprehensive&period=daily" \
    --http-method GET \
    --oidc-service-account-email "analytics-scheduler@${PROJECT_ID}.iam.gserviceaccount.com" \
    --attempt-deadline 300s \
    --project ${PROJECT_ID} || echo "Daily job might already exist"

  # Create weekly analytics job
  gcloud scheduler jobs create http ${FUNCTION_NAME}-weekly \
    --location ${REGION} \
    --schedule "0 2 * * 1" \
    --uri "https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}?type=comprehensive&period=weekly" \
    --http-method GET \
    --oidc-service-account-email "analytics-scheduler@${PROJECT_ID}.iam.gserviceaccount.com" \
    --attempt-deadline 300s \
    --project ${PROJECT_ID} || echo "Weekly job might already exist"

  # Create monthly analytics job
  gcloud scheduler jobs create http ${FUNCTION_NAME}-monthly \
    --location ${REGION} \
    --schedule "0 3 1 * *" \
    --uri "https://${REGION}-${PROJECT_ID}.cloudfunctions.net/${FUNCTION_NAME}?type=comprehensive&period=monthly" \
    --http-method GET \
    --oidc-service-account-email "analytics-scheduler@${PROJECT_ID}.iam.gserviceaccount.com" \
    --attempt-deadline 300s \
    --project ${PROJECT_ID} || echo "Monthly job might already exist"

  echo "Analytics scheduled:"
  echo "- Daily: 1:00 AM"
  echo "- Weekly: Mondays at 2:00 AM"
  echo "- Monthly: First day of month at 3:00 AM"
fi
