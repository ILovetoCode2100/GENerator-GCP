#!/bin/bash

# Deploy all Cloud Functions
# Usage: ./deploy-all.sh [project-id]

set -e

PROJECT_ID=${1:-virtuoso-api}
REGION="us-central1"

echo "==================================="
echo "Deploying all Cloud Functions"
echo "Project: ${PROJECT_ID}"
echo "Region: ${REGION}"
echo "==================================="
echo ""

# Make all deploy scripts executable
chmod +x */deploy.sh

# Get Memorystore Redis IP
echo "Please enter your Memorystore Redis IP address:"
read REDIS_HOST
export REDIS_HOST

# Get Cloud Run service URL
echo "Please enter your Cloud Run service URL:"
read CLOUD_RUN_URL
export CLOUD_RUN_URL

# Deploy functions in order
FUNCTIONS=(
  "health-check"
  "auth-validator"
  "webhook-handler"
  "analytics"
  "cleanup"
)

FAILED_DEPLOYMENTS=()

for func in "${FUNCTIONS[@]}"; do
  echo ""
  echo "==================================="
  echo "Deploying ${func}..."
  echo "==================================="

  cd ${func}

  # Update environment variables in deploy script
  if [[ "$func" == "health-check" ]]; then
    sed -i.bak "s|REDIS_HOST=\".*\"|REDIS_HOST=\"${REDIS_HOST}\"|g" deploy.sh
    sed -i.bak "s|CLOUD_RUN_URL=\".*\"|CLOUD_RUN_URL=\"${CLOUD_RUN_URL}\"|g" deploy.sh
  elif [[ "$func" == "auth-validator" ]]; then
    sed -i.bak "s|REDIS_HOST=\".*\"|REDIS_HOST=\"${REDIS_HOST}\"|g" deploy.sh
  fi

  # Deploy function
  if ./deploy.sh ${PROJECT_ID}; then
    echo "✓ ${func} deployed successfully"
  else
    echo "✗ ${func} deployment failed"
    FAILED_DEPLOYMENTS+=($func)
  fi

  # Clean up backup files
  rm -f deploy.sh.bak

  cd ..
done

echo ""
echo "==================================="
echo "Deployment Summary"
echo "==================================="
echo ""

if [ ${#FAILED_DEPLOYMENTS[@]} -eq 0 ]; then
  echo "✓ All functions deployed successfully!"
else
  echo "✗ Failed deployments: ${FAILED_DEPLOYMENTS[@]}"
  echo ""
  echo "To retry failed deployments:"
  for func in "${FAILED_DEPLOYMENTS[@]}"; do
    echo "  cd ${func} && ./deploy.sh ${PROJECT_ID}"
  done
fi

echo ""
echo "Function Endpoints:"
echo "- Health Check: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/health-check"
echo "- Auth Validator: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/auth-validator"
echo "- Webhook Handler: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/webhook-handler"
echo "- Analytics: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/analytics"
echo "- Cleanup: https://${REGION}-${PROJECT_ID}.cloudfunctions.net/cleanup"

echo ""
echo "Next Steps:"
echo "1. Configure webhook secrets in Secret Manager"
echo "2. Set up monitoring alerts in Cloud Monitoring"
echo "3. Test each function endpoint"
echo "4. Configure Cloud Scheduler jobs (if not done during deployment)"

# Optional: Set up monitoring
echo ""
read -p "Set up basic monitoring alerts? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo "Creating monitoring alerts..."

  # Function error rate alert
  gcloud alpha monitoring policies create \
    --notification-channels=[] \
    --display-name="Cloud Functions Error Rate" \
    --condition="rate(metric.type=\"cloudfunctions.googleapis.com/function/execution_count\" AND resource.type=\"cloud_function\" AND metric.label.status!=\"ok\") > 0.1" \
    --project=${PROJECT_ID} || echo "Error rate alert might already exist"

  # Function latency alert
  gcloud alpha monitoring policies create \
    --notification-channels=[] \
    --display-name="Cloud Functions High Latency" \
    --condition="fetch(metric.type=\"cloudfunctions.googleapis.com/function/execution_times\" AND resource.type=\"cloud_function\").percentile(99) > 5000" \
    --project=${PROJECT_ID} || echo "Latency alert might already exist"

  echo "Basic monitoring alerts created (no notification channels configured)"
fi

echo ""
echo "Deployment complete!"
