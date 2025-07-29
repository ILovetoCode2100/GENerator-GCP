#!/bin/bash
# Emergency Rollback Script for Virtuoso API CLI on GCP
# This script handles rollback of deployments in case of failures

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default values
PROJECT_ID="${GCP_PROJECT_ID:-$(gcloud config get-value project)}"
REGION="${GCP_REGION:-us-central1}"
SERVICE_NAME="virtuoso-api-cli"
ROLLBACK_TYPE="all"
REVISION=""
TERRAFORM_STATE_BUCKET=""

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to rollback Cloud Run service
rollback_cloud_run() {
    print_info "Rolling back Cloud Run service..."

    if [ -z "$REVISION" ]; then
        # Get previous revision
        print_info "Finding previous stable revision..."

        REVISIONS=$(gcloud run revisions list \
            --service="$SERVICE_NAME" \
            --platform=managed \
            --region="$REGION" \
            --format="value(name)" \
            --limit=5)

        if [ -z "$REVISIONS" ]; then
            print_error "No revisions found for service $SERVICE_NAME"
            return 1
        fi

        # Show available revisions
        print_info "Available revisions:"
        echo "$REVISIONS" | nl -nrz

        # Get current revision
        CURRENT=$(gcloud run services describe "$SERVICE_NAME" \
            --platform=managed \
            --region="$REGION" \
            --format="value(status.latestReadyRevisionName)")

        print_info "Current revision: $CURRENT"

        # Find previous revision
        PREVIOUS=""
        for rev in $REVISIONS; do
            if [ "$rev" != "$CURRENT" ]; then
                PREVIOUS="$rev"
                break
            fi
        done

        if [ -z "$PREVIOUS" ]; then
            print_error "No previous revision found to rollback to"
            return 1
        fi

        REVISION="$PREVIOUS"
        print_info "Rolling back to revision: $REVISION"
    fi

    # Update traffic to previous revision
    gcloud run services update-traffic "$SERVICE_NAME" \
        --to-revisions="$REVISION=100" \
        --platform=managed \
        --region="$REGION"

    print_success "Cloud Run service rolled back to $REVISION"
}

# Function to rollback Terraform state
rollback_terraform() {
    print_info "Rolling back Terraform state..."

    if [ -z "$TERRAFORM_STATE_BUCKET" ]; then
        TERRAFORM_STATE_BUCKET="${PROJECT_ID}-terraform-state"
    fi

    cd "$SCRIPT_DIR/terraform"

    # List available state backups
    print_info "Available Terraform state backups:"
    gsutil ls -l "gs://$TERRAFORM_STATE_BUCKET/terraform.tfstate.backup*" | tail -10

    # Get user confirmation
    read -p "Enter backup file name to restore (or 'skip' to skip): " backup_file

    if [ "$backup_file" != "skip" ] && [ -n "$backup_file" ]; then
        # Backup current state
        print_info "Backing up current state..."
        terraform state pull > "terraform.tfstate.current.$(date +%Y%m%d%H%M%S)"

        # Restore from backup
        print_info "Restoring state from $backup_file..."
        gsutil cp "$backup_file" terraform.tfstate
        terraform state push terraform.tfstate

        # Apply the restored state
        print_info "Applying restored state..."
        terraform apply -auto-approve

        print_success "Terraform state restored and applied"
    else
        print_warning "Terraform rollback skipped"
    fi

    cd "$SCRIPT_DIR"
}

# Function to rollback Cloud Functions
rollback_cloud_functions() {
    print_info "Rolling back Cloud Functions..."

    local functions=(
        "analytics"
        "auth-validator"
        "cleanup"
        "health-check"
        "webhook-handler"
    )

    for func in "${functions[@]}"; do
        print_info "Checking function: $func"

        # Get current generation
        CURRENT_GEN=$(gcloud functions describe "$func" \
            --region="$REGION" \
            --format="value(serviceConfig.revision)" 2>/dev/null || echo "")

        if [ -n "$CURRENT_GEN" ]; then
            # List available versions
            print_info "Current version of $func: $CURRENT_GEN"

            # For Gen2 functions, we can't directly rollback
            # Instead, we redeploy from the previous source
            print_warning "Cloud Functions Gen2 doesn't support direct rollback"
            print_info "To rollback $func, redeploy from a previous git commit"
        fi
    done

    print_info "Cloud Functions rollback requires manual intervention"
    print_info "Use 'git checkout <previous-commit> && ./deploy-all.sh' in functions directory"
}

# Function to clear caches
clear_caches() {
    print_info "Clearing caches..."

    # Clear CDN cache if using Cloud CDN
    if gcloud compute url-maps list --format="value(name)" | grep -q "virtuoso-lb"; then
        print_info "Invalidating CDN cache..."
        gcloud compute url-maps invalidate-cdn-cache virtuoso-lb \
            --path="/*" \
            --async
    fi

    # Clear Redis cache if accessible
    print_info "Clearing Redis cache..."
    # This would require Redis connection details
    print_warning "Redis cache clearing requires manual intervention"

    print_success "Cache clearing initiated"
}

# Function to restore from backup
restore_from_backup() {
    print_info "Restoring from backup..."

    # List available Firestore backups
    print_info "Available Firestore backups:"
    gcloud firestore operations list \
        --filter="name:backup" \
        --limit=10 \
        --format="table(name,metadata.startTime,done)"

    read -p "Enter backup operation name to restore (or 'skip'): " backup_name

    if [ "$backup_name" != "skip" ] && [ -n "$backup_name" ]; then
        # Restore Firestore
        gcloud firestore import "gs://${PROJECT_ID}-backups/firestore/$backup_name" \
            --async

        print_success "Firestore restore initiated"
    fi

    print_success "Backup restoration completed"
}

# Function to verify rollback
verify_rollback() {
    print_info "Verifying rollback..."

    # Check Cloud Run service
    SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" \
        --platform=managed \
        --region="$REGION" \
        --format="value(status.url)")

    if [ -n "$SERVICE_URL" ]; then
        print_info "Testing service health..."
        if curl -s "$SERVICE_URL/health" | grep -q "healthy"; then
            print_success "Service is healthy"
        else
            print_error "Service health check failed"
        fi
    fi

    # Check metrics
    print_info "Recent error rate:"
    gcloud monitoring time-series list \
        --filter='metric.type="run.googleapis.com/request_count" AND
                 resource.type="cloud_run_revision" AND
                 metric.label.response_code_class="5xx"' \
        --format="table(resource.labels.service_name,points[0].value.int64_value,points[0].interval.endTime)" \
        --limit=5

    print_success "Rollback verification completed"
}

# Function to send notification
send_notification() {
    local status="$1"
    local message="$2"

    print_info "Sending rollback notification..."

    # Log to Cloud Logging
    gcloud logging write rollback-log "$message" \
        --severity="$status" \
        --resource="global" \
        --project="$PROJECT_ID"

    # If Slack webhook is configured
    if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
        curl -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"text\":\"Rollback $status: $message\"}"
    fi

    print_success "Notification sent"
}

# Function to show rollback plan
show_rollback_plan() {
    print_info "Rollback Plan"
    echo -e "${BLUE}===================================================${NC}"
    echo -e "Project: ${GREEN}$PROJECT_ID${NC}"
    echo -e "Region: ${GREEN}$REGION${NC}"
    echo -e "Service: ${GREEN}$SERVICE_NAME${NC}"
    echo -e "Rollback Type: ${GREEN}$ROLLBACK_TYPE${NC}"

    echo -e "\nComponents to rollback:"
    case "$ROLLBACK_TYPE" in
        "all")
            echo -e "  ✓ Cloud Run service"
            echo -e "  ✓ Terraform infrastructure"
            echo -e "  ✓ Cloud Functions"
            echo -e "  ✓ Clear caches"
            ;;
        "service")
            echo -e "  ✓ Cloud Run service only"
            ;;
        "terraform")
            echo -e "  ✓ Terraform infrastructure only"
            ;;
        "functions")
            echo -e "  ✓ Cloud Functions only"
            ;;
    esac
    echo -e "${BLUE}===================================================${NC}"

    read -p "Proceed with rollback? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Rollback cancelled"
        exit 0
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        --region)
            REGION="$2"
            shift 2
            ;;
        --service)
            SERVICE_NAME="$2"
            shift 2
            ;;
        --type)
            ROLLBACK_TYPE="$2"
            shift 2
            ;;
        --revision)
            REVISION="$2"
            shift 2
            ;;
        --skip-verification)
            SKIP_VERIFICATION=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --project-id ID         GCP project ID"
            echo "  --region REGION         GCP region"
            echo "  --service NAME          Service name (default: virtuoso-api-cli)"
            echo "  --type TYPE            Rollback type: all|service|terraform|functions"
            echo "  --revision REV          Specific revision to rollback to"
            echo "  --skip-verification     Skip rollback verification"
            echo "  --help                  Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Main rollback flow
print_warning "EMERGENCY ROLLBACK PROCEDURE"
echo -e "${RED}===================================================${NC}\n"

# Verify project
if [ -z "$PROJECT_ID" ]; then
    print_error "Project ID not specified"
    exit 1
fi

# Show rollback plan
show_rollback_plan

# Start rollback
START_TIME=$(date +%s)
send_notification "INFO" "Starting rollback for $SERVICE_NAME in $PROJECT_ID"

# Execute rollback based on type
case "$ROLLBACK_TYPE" in
    "all")
        rollback_cloud_run
        rollback_terraform
        rollback_cloud_functions
        clear_caches
        ;;
    "service")
        rollback_cloud_run
        ;;
    "terraform")
        rollback_terraform
        ;;
    "functions")
        rollback_cloud_functions
        ;;
    *)
        print_error "Invalid rollback type: $ROLLBACK_TYPE"
        exit 1
        ;;
esac

# Verify rollback if not skipped
if [ "${SKIP_VERIFICATION:-false}" != true ]; then
    verify_rollback
fi

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Send completion notification
send_notification "INFO" "Rollback completed in ${DURATION}s for $SERVICE_NAME"

print_success "Rollback completed successfully!"
print_info "Duration: ${DURATION} seconds"
print_warning "Remember to investigate and fix the root cause before redeploying"
