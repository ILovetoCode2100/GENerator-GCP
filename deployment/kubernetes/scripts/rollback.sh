#!/bin/bash
set -e

# Rollback script for Virtuoso API CLI

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="virtuoso-api"
DEPLOYMENT="virtuoso-api-cli"

# Default values
REVISION=""
DRY_RUN=false
FORCE=false
WAIT_TIMEOUT="300s"

# Usage function
usage() {
    echo "Usage: $0 [-r <revision>] [-d] [-f] [-w <timeout>]"
    echo ""
    echo "Options:"
    echo "  -r <revision>     Revision to rollback to (default: previous revision)"
    echo "  -d                Dry run - show what would be rolled back"
    echo "  -f                Force rollback without confirmation"
    echo "  -w <timeout>      Wait timeout (default: 300s)"
    echo "  -h                Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Rollback to previous revision"
    echo "  $0 -r 5               # Rollback to revision 5"
    echo "  $0 -d                 # Show rollback plan without executing"
    exit 1
}

# Parse command line arguments
while getopts "r:dfw:h" opt; do
    case ${opt} in
        r )
            REVISION=$OPTARG
            ;;
        d )
            DRY_RUN=true
            ;;
        f )
            FORCE=true
            ;;
        w )
            WAIT_TIMEOUT=$OPTARG
            ;;
        h )
            usage
            ;;
        \? )
            usage
            ;;
    esac
done

# Functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"
}

warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"
}

confirm() {
    if [ "${FORCE}" = true ]; then
        return 0
    fi

    read -p "$1 (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        return 1
    fi
    return 0
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        error "kubectl is not installed"
        exit 1
    fi

    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    # Check if deployment exists
    if ! kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE} &> /dev/null; then
        error "Deployment ${DEPLOYMENT} not found in namespace ${NAMESPACE}"
        exit 1
    fi
}

# Get rollout history
show_rollout_history() {
    log "Rollout history for ${DEPLOYMENT}:"
    echo ""
    kubectl rollout history deployment/${DEPLOYMENT} -n ${NAMESPACE}
    echo ""
}

# Get current deployment info
get_current_info() {
    log "Current deployment status:"

    # Get current revision
    CURRENT_REVISION=$(kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE} -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
    echo "Current revision: ${CURRENT_REVISION}"

    # Get current image
    CURRENT_IMAGE=$(kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE} -o jsonpath='{.spec.template.spec.containers[0].image}')
    echo "Current image: ${CURRENT_IMAGE}"

    # Get replica status
    READY_REPLICAS=$(kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE} -o jsonpath='{.status.readyReplicas}')
    DESIRED_REPLICAS=$(kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE} -o jsonpath='{.spec.replicas}')
    echo "Replicas: ${READY_REPLICAS}/${DESIRED_REPLICAS} ready"
    echo ""
}

# Get target revision info
get_target_info() {
    if [ -z "${REVISION}" ]; then
        # Get previous revision
        TARGET_REVISION=$((CURRENT_REVISION - 1))
        if [ "${TARGET_REVISION}" -lt 1 ]; then
            error "Cannot rollback from revision 1"
            exit 1
        fi
    else
        TARGET_REVISION="${REVISION}"
    fi

    log "Target revision: ${TARGET_REVISION}"

    # Validate target revision exists
    if ! kubectl rollout history deployment/${DEPLOYMENT} -n ${NAMESPACE} --revision=${TARGET_REVISION} &> /dev/null; then
        error "Revision ${TARGET_REVISION} not found in rollout history"
        exit 1
    fi

    # Show target revision details
    echo ""
    kubectl rollout history deployment/${DEPLOYMENT} -n ${NAMESPACE} --revision=${TARGET_REVISION}
}

# Create backup of current state
create_backup() {
    if [ "${DRY_RUN}" = true ]; then
        log "Would create backup of current deployment state"
        return
    fi

    log "Creating backup of current deployment state..."

    BACKUP_FILE="/tmp/virtuoso-deployment-backup-$(date +%Y%m%d-%H%M%S).yaml"
    kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE} -o yaml > "${BACKUP_FILE}"

    log "Backup saved to: ${BACKUP_FILE}"
}

# Perform rollback
perform_rollback() {
    if [ "${DRY_RUN}" = true ]; then
        log "Dry run mode - no changes will be made"
        log "Would rollback ${DEPLOYMENT} to revision ${TARGET_REVISION}"
        return
    fi

    log "Performing rollback..."

    # Execute rollback
    if [ -z "${REVISION}" ]; then
        kubectl rollout undo deployment/${DEPLOYMENT} -n ${NAMESPACE}
    else
        kubectl rollout undo deployment/${DEPLOYMENT} -n ${NAMESPACE} --to-revision=${TARGET_REVISION}
    fi

    # Wait for rollout to complete
    log "Waiting for rollback to complete..."
    kubectl rollout status deployment/${DEPLOYMENT} -n ${NAMESPACE} --timeout=${WAIT_TIMEOUT} || {
        error "Rollback failed"

        # Show recent events
        log "Recent events:"
        kubectl get events -n ${NAMESPACE} --sort-by='.lastTimestamp' | tail -10

        # Show pod status
        log "Pod status:"
        kubectl get pods -n ${NAMESPACE} -l app=${DEPLOYMENT}

        exit 1
    }

    log "Rollback completed successfully!"
}

# Post-rollback checks
post_rollback_checks() {
    if [ "${DRY_RUN}" = true ]; then
        return
    fi

    log "Running post-rollback checks..."

    # Get new revision
    NEW_REVISION=$(kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE} -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
    log "New revision: ${NEW_REVISION}"

    # Check pod status
    READY_PODS=$(kubectl get pods -n ${NAMESPACE} -l app=${DEPLOYMENT} -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}' | grep -o "True" | wc -l)
    TOTAL_PODS=$(kubectl get pods -n ${NAMESPACE} -l app=${DEPLOYMENT} --no-headers | wc -l)

    if [ "${READY_PODS}" -ne "${TOTAL_PODS}" ]; then
        warning "Not all pods are ready: ${READY_PODS}/${TOTAL_PODS}"
    else
        log "All pods are ready: ${READY_PODS}/${TOTAL_PODS}"
    fi

    # Run health check
    log "Waiting 30 seconds before health check..."
    sleep 30

    # Check service health
    SERVICE_IP=$(kubectl get svc ${DEPLOYMENT} -n ${NAMESPACE} -o jsonpath='{.spec.clusterIP}')
    if kubectl run health-check-temp --rm -i --restart=Never --image=curlimages/curl:latest -- \
        curl -s -o /dev/null -w "%{http_code}" "http://${SERVICE_IP}:8000/health" | grep -q "200"; then
        log "Health check passed"
    else
        warning "Health check failed"
    fi
}

# Show rollback summary
show_summary() {
    if [ "${DRY_RUN}" = true ]; then
        return
    fi

    echo ""
    log "Rollback Summary:"
    echo "======================="
    echo "Deployment: ${DEPLOYMENT}"
    echo "Namespace: ${NAMESPACE}"
    echo "Previous revision: ${CURRENT_REVISION}"
    echo "Current revision: $(kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE} -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')"
    echo ""

    # Show current deployment info
    kubectl get deployment ${DEPLOYMENT} -n ${NAMESPACE}
    echo ""

    # Show pod status
    kubectl get pods -n ${NAMESPACE} -l app=${DEPLOYMENT}
}

# Main execution
main() {
    log "Starting rollback process"

    # Check prerequisites
    check_prerequisites

    # Get current state
    get_current_info

    # Show rollout history
    show_rollout_history

    # Get target revision info
    get_target_info

    # Confirm rollback
    if ! confirm "Rollback ${DEPLOYMENT} from revision ${CURRENT_REVISION} to ${TARGET_REVISION}?"; then
        log "Rollback cancelled"
        exit 0
    fi

    # Execute rollback
    create_backup
    perform_rollback
    post_rollback_checks
    show_summary

    log "Rollback process completed!"
}

# Execute main function
main
