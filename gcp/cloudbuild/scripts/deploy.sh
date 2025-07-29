#!/bin/bash
# Deployment helper script for Virtuoso API CLI
# Handles blue-green deployments with traffic management

set -e

# Configuration
SERVICE_NAME="${SERVICE_NAME:-virtuoso-api-cli}"
REGION="${REGION:-us-central1}"
PROJECT_ID="${PROJECT_ID}"
ENVIRONMENT="${ENVIRONMENT:-dev}"
IMAGE="${IMAGE}"
VERSION="${VERSION:-${SHORT_SHA}}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check required variables
if [ -z "${PROJECT_ID}" ]; then
    log_error "PROJECT_ID is not set"
    exit 1
fi

if [ -z "${IMAGE}" ]; then
    log_error "IMAGE is not set"
    exit 1
fi

log_info "Starting deployment for ${SERVICE_NAME} in ${ENVIRONMENT}"
log_info "Image: ${IMAGE}"
log_info "Version: ${VERSION}"

# Step 1: Deploy new revision with no traffic
log_info "Deploying new revision (green)..."

DEPLOY_ARGS=(
    "--image=${IMAGE}"
    "--region=${REGION}"
    "--platform=managed"
    "--no-traffic"
    "--tag=green-${VERSION}"
    "--service-account=${SERVICE_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
)

# Add environment-specific configurations
case ${ENVIRONMENT} in
    prod)
        DEPLOY_ARGS+=(
            "--max-instances=100"
            "--min-instances=3"
            "--memory=1Gi"
            "--cpu=2"
            "--concurrency=1000"
        )
        ;;
    staging)
        DEPLOY_ARGS+=(
            "--max-instances=50"
            "--min-instances=1"
            "--memory=512Mi"
            "--cpu=1"
            "--concurrency=500"
        )
        ;;
    dev)
        DEPLOY_ARGS+=(
            "--max-instances=10"
            "--min-instances=0"
            "--memory=256Mi"
            "--cpu=1"
            "--concurrency=100"
        )
        ;;
esac

# Common configurations
DEPLOY_ARGS+=(
    "--timeout=300"
    "--set-env-vars=VERSION=${VERSION}"
    "--set-env-vars=ENVIRONMENT=${ENVIRONMENT}"
    "--set-env-vars=BUILD_ID=${BUILD_ID:-local}"
    "--set-env-vars=COMMIT_SHA=${COMMIT_SHA:-unknown}"
)

# Deploy the service
if gcloud run deploy ${SERVICE_NAME} "${DEPLOY_ARGS[@]}"; then
    log_success "New revision deployed successfully"
else
    log_error "Failed to deploy new revision"
    exit 1
fi

# Step 2: Get the green URL for testing
GREEN_URL=$(gcloud run services describe ${SERVICE_NAME} \
    --region=${REGION} \
    --format="value(status.address.url)" \
    | sed "s|https://|https://green-${VERSION}---|")

log_info "Green deployment URL: ${GREEN_URL}"

# Step 3: Health check
log_info "Running health check on green deployment..."

MAX_RETRIES=30
RETRY_COUNT=0

while [ ${RETRY_COUNT} -lt ${MAX_RETRIES} ]; do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${GREEN_URL}/health" || echo "000")

    if [ "${HTTP_CODE}" == "200" ]; then
        log_success "Health check passed"
        break
    else
        log_warning "Health check failed (${HTTP_CODE}), retrying... (${RETRY_COUNT}/${MAX_RETRIES})"
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 2
    fi
done

if [ ${RETRY_COUNT} -eq ${MAX_RETRIES} ]; then
    log_error "Health check failed after ${MAX_RETRIES} attempts"
    exit 1
fi

# Step 4: Run smoke tests
if [ -f "./gcp/cloudbuild/scripts/smoke-test.sh" ]; then
    log_info "Running smoke tests..."
    export TARGET_URL="${GREEN_URL}"
    if ./gcp/cloudbuild/scripts/smoke-test.sh; then
        log_success "Smoke tests passed"
    else
        log_error "Smoke tests failed"
        exit 1
    fi
fi

# Step 5: Traffic migration (blue-green deployment)
log_info "Starting traffic migration..."

# Get current traffic allocation
CURRENT_TRAFFIC=$(gcloud run services describe ${SERVICE_NAME} \
    --region=${REGION} \
    --format=json | jq -r '.spec.traffic')

# Define traffic split percentages
if [ "${ENVIRONMENT}" == "prod" ]; then
    TRAFFIC_STEPS=(5 10 25 50 75 100)
    STEP_DELAY=300  # 5 minutes
    ERROR_THRESHOLD=5
else
    TRAFFIC_STEPS=(10 50 100)
    STEP_DELAY=60   # 1 minute
    ERROR_THRESHOLD=10
fi

# Gradually shift traffic
for PERCENTAGE in "${TRAFFIC_STEPS[@]}"; do
    log_info "Shifting ${PERCENTAGE}% traffic to green deployment..."

    if gcloud run services update-traffic ${SERVICE_NAME} \
        --region=${REGION} \
        --to-tags="green-${VERSION}=${PERCENTAGE}"; then
        log_success "Traffic updated to ${PERCENTAGE}%"
    else
        log_error "Failed to update traffic"
        exit 1
    fi

    # Skip waiting on 100%
    if [ ${PERCENTAGE} -eq 100 ]; then
        break
    fi

    log_info "Monitoring for ${STEP_DELAY} seconds..."
    sleep ${STEP_DELAY}

    # Check error rate
    ERROR_COUNT=$(gcloud logging read \
        "resource.type=\"cloud_run_revision\" \
         AND resource.labels.revision_name=\"${SERVICE_NAME}-green-${VERSION}\" \
         AND severity>=ERROR \
         AND timestamp>=\"$(date -u -d '5 minutes ago' +%Y-%m-%dT%H:%M:%SZ)\"" \
        --limit=100 \
        --format=json | jq length)

    if [ ${ERROR_COUNT} -gt ${ERROR_THRESHOLD} ]; then
        log_error "High error rate detected (${ERROR_COUNT} errors)"
        log_warning "Rolling back deployment..."

        # Rollback to previous revision
        ./gcp/cloudbuild/scripts/rollback.sh
        exit 1
    fi

    log_success "No issues detected, continuing..."
done

# Step 6: Update latest tag
log_info "Updating latest tag..."
gcloud run services update-traffic ${SERVICE_NAME} \
    --region=${REGION} \
    --to-latest

# Step 7: Clean up old revisions (keep last 5)
log_info "Cleaning up old revisions..."
OLD_REVISIONS=$(gcloud run revisions list \
    --service=${SERVICE_NAME} \
    --region=${REGION} \
    --format="value(name)" \
    --sort-by="~creationTimestamp" | tail -n +6)

if [ -n "${OLD_REVISIONS}" ]; then
    for REVISION in ${OLD_REVISIONS}; do
        log_info "Deleting old revision: ${REVISION}"
        gcloud run revisions delete ${REVISION} \
            --region=${REGION} \
            --quiet || log_warning "Failed to delete ${REVISION}"
    done
fi

# Step 8: Update deployment metadata
log_info "Updating deployment metadata..."

# Store deployment info
cat > deployment-info.json << EOF
{
  "service": "${SERVICE_NAME}",
  "environment": "${ENVIRONMENT}",
  "version": "${VERSION}",
  "image": "${IMAGE}",
  "deployed_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "deployed_by": "${USER:-system}",
  "build_id": "${BUILD_ID:-local}",
  "commit_sha": "${COMMIT_SHA:-unknown}",
  "service_url": "$(gcloud run services describe ${SERVICE_NAME} --region=${REGION} --format='value(status.url)')"
}
EOF

# Upload to GCS
gsutil cp deployment-info.json \
    "gs://${PROJECT_ID}-deployments/${ENVIRONMENT}/${VERSION}/deployment-info.json"

# Final status
log_success "Deployment completed successfully!"
log_info "Service URL: $(gcloud run services describe ${SERVICE_NAME} --region=${REGION} --format='value(status.url)')"
log_info "Version: ${VERSION}"

exit 0
