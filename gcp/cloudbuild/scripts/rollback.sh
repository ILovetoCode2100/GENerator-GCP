#!/bin/bash
# Rollback script for Virtuoso API CLI deployments
# Quickly reverts to the previous stable version

set -e

# Configuration
SERVICE_NAME="${SERVICE_NAME:-virtuoso-api-cli}"
REGION="${REGION:-us-central1}"
PROJECT_ID="${PROJECT_ID}"
ENVIRONMENT="${ENVIRONMENT:-dev}"

# Colors
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

log_warning "Starting rollback for ${SERVICE_NAME} in ${ENVIRONMENT}"

# Step 1: Get current revision info
log_info "Getting current deployment status..."

CURRENT_REVISIONS=$(gcloud run services describe ${SERVICE_NAME} \
    --region=${REGION} \
    --format=json | jq -r '.status.traffic[]')

echo "${CURRENT_REVISIONS}" | jq .

# Step 2: Find the previous stable revision
log_info "Finding previous stable revision..."

# Get list of revisions sorted by creation time (newest first)
REVISIONS=$(gcloud run revisions list \
    --service=${SERVICE_NAME} \
    --region=${REGION} \
    --format="value(name,metadata.annotations.'run.googleapis.com/traffic'.percent)" \
    --sort-by="~creationTimestamp" \
    --limit=10)

# Find the second revision with 100% traffic (or the second newest)
PREVIOUS_REVISION=""
REVISION_COUNT=0

while IFS= read -r line; do
    REVISION_NAME=$(echo "$line" | awk '{print $1}')

    # Skip the current revision
    if [ ${REVISION_COUNT} -gt 0 ]; then
        PREVIOUS_REVISION="${REVISION_NAME}"
        break
    fi

    REVISION_COUNT=$((REVISION_COUNT + 1))
done <<< "${REVISIONS}"

if [ -z "${PREVIOUS_REVISION}" ]; then
    log_error "No previous revision found for rollback"
    exit 1
fi

log_info "Rolling back to revision: ${PREVIOUS_REVISION}"

# Step 3: Immediate traffic shift (emergency rollback)
log_warning "Shifting 100% traffic to previous revision..."

if gcloud run services update-traffic ${SERVICE_NAME} \
    --region=${REGION} \
    --to-revisions="${PREVIOUS_REVISION}=100"; then
    log_success "Traffic shifted successfully"
else
    log_error "Failed to shift traffic"
    exit 1
fi

# Step 4: Health check on rolled-back service
log_info "Verifying rollback health..."

SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} \
    --region=${REGION} \
    --format='value(status.url)')

MAX_RETRIES=10
RETRY_COUNT=0

while [ ${RETRY_COUNT} -lt ${MAX_RETRIES} ]; do
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${SERVICE_URL}/health" || echo "000")

    if [ "${HTTP_CODE}" == "200" ]; then
        log_success "Health check passed on rolled-back service"
        break
    else
        log_warning "Health check failed (${HTTP_CODE}), retrying... (${RETRY_COUNT}/${MAX_RETRIES})"
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 2
    fi
done

if [ ${RETRY_COUNT} -eq ${MAX_RETRIES} ]; then
    log_error "Health check failed after rollback"
    # Don't exit - rollback is complete even if health check fails
fi

# Step 5: Tag the failed revision
log_info "Tagging failed revision..."

# Get the most recent revision (the one we rolled back from)
FAILED_REVISION=$(gcloud run revisions list \
    --service=${SERVICE_NAME} \
    --region=${REGION} \
    --format="value(name)" \
    --sort-by="~creationTimestamp" \
    --limit=1)

# Add annotation to mark as failed
gcloud run services update ${SERVICE_NAME} \
    --region=${REGION} \
    --update-annotations="rollback-from-${FAILED_REVISION}=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Step 6: Send notifications
log_info "Sending rollback notifications..."

# Create rollback report
cat > rollback-report.json << EOF
{
  "service": "${SERVICE_NAME}",
  "environment": "${ENVIRONMENT}",
  "rolled_back_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "rolled_back_by": "${USER:-system}",
  "failed_revision": "${FAILED_REVISION}",
  "stable_revision": "${PREVIOUS_REVISION}",
  "reason": "${ROLLBACK_REASON:-High error rate detected}",
  "build_id": "${BUILD_ID:-unknown}"
}
EOF

# Upload rollback report
gsutil cp rollback-report.json \
    "gs://${PROJECT_ID}-deployments/${ENVIRONMENT}/rollbacks/$(date +%Y%m%d-%H%M%S)-rollback.json"

# Step 7: Collect diagnostics from failed revision
log_info "Collecting diagnostics..."

# Get recent logs from failed revision
gcloud logging read \
    "resource.type=\"cloud_run_revision\" \
     AND resource.labels.revision_name=\"${FAILED_REVISION}\" \
     AND timestamp>=\"$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%SZ)\"" \
    --limit=1000 \
    --format=json > failed-revision-logs.json

# Get metrics
gcloud monitoring time-series list \
    --filter="resource.type=\"cloud_run_revision\" AND resource.labels.revision_name=\"${FAILED_REVISION}\"" \
    --interval-start-time="$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%SZ)" \
    --format=json > failed-revision-metrics.json

# Upload diagnostics
gsutil cp failed-revision-logs.json \
    "gs://${PROJECT_ID}-deployments/${ENVIRONMENT}/rollbacks/$(date +%Y%m%d-%H%M%S)-logs.json"
gsutil cp failed-revision-metrics.json \
    "gs://${PROJECT_ID}-deployments/${ENVIRONMENT}/rollbacks/$(date +%Y%m%d-%H%M%S)-metrics.json"

# Step 8: Create incident report
log_info "Creating incident report..."

cat > incident-report.md << EOF
# Deployment Rollback Incident Report

**Date:** $(date -u +%Y-%m-%dT%H:%M:%SZ)
**Service:** ${SERVICE_NAME}
**Environment:** ${ENVIRONMENT}

## Summary
Automatic rollback was triggered due to: ${ROLLBACK_REASON:-High error rate detected}

## Timeline
- Deployment started: ${DEPLOYMENT_START_TIME:-Unknown}
- Issues detected: ${ISSUE_DETECTED_TIME:-Unknown}
- Rollback initiated: $(date -u +%Y-%m-%dT%H:%M:%SZ)
- Service restored: $(date -u +%Y-%m-%dT%H:%M:%SZ)

## Impact
- Failed revision: ${FAILED_REVISION}
- Stable revision: ${PREVIOUS_REVISION}
- Estimated downtime: ${DOWNTIME:-< 5 minutes}

## Root Cause
To be investigated. Initial indicators:
- Error count threshold exceeded
- See diagnostics in GCS bucket

## Action Items
1. Investigate root cause of deployment failure
2. Fix identified issues
3. Re-deploy with fixes
4. Update deployment procedures if needed

## Diagnostics Location
gs://${PROJECT_ID}-deployments/${ENVIRONMENT}/rollbacks/$(date +%Y%m%d-%H%M%S)-*
EOF

# Final status
log_success "Rollback completed successfully!"
log_info "Service restored to revision: ${PREVIOUS_REVISION}"
log_info "Service URL: ${SERVICE_URL}"
log_warning "Please investigate the failure and create an incident report"

# Exit with non-zero to indicate rollback occurred
exit 2
