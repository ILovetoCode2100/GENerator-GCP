#!/bin/bash
set -e

# Deployment script for Virtuoso API CLI

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="${SCRIPT_DIR}/../"
NAMESPACE="virtuoso-api"

# Default values
ENVIRONMENT=""
DRY_RUN=false
FORCE=false
WAIT_TIMEOUT="600s"
IMAGE_TAG=""

# Usage function
usage() {
    echo "Usage: $0 -e <environment> [-t <image-tag>] [-d] [-f] [-w <timeout>]"
    echo ""
    echo "Options:"
    echo "  -e <environment>  Environment to deploy (dev|staging|production)"
    echo "  -t <image-tag>    Docker image tag to deploy (default: latest)"
    echo "  -d                Dry run - show what would be deployed"
    echo "  -f                Force deployment without confirmation"
    echo "  -w <timeout>      Wait timeout (default: 600s)"
    echo "  -h                Show this help message"
    exit 1
}

# Parse command line arguments
while getopts "e:t:dfw:h" opt; do
    case ${opt} in
        e )
            ENVIRONMENT=$OPTARG
            ;;
        t )
            IMAGE_TAG=$OPTARG
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

# Validate environment
if [ -z "${ENVIRONMENT}" ]; then
    echo -e "${RED}Error: Environment is required${NC}"
    usage
fi

if [[ ! "${ENVIRONMENT}" =~ ^(dev|staging|production)$ ]]; then
    echo -e "${RED}Error: Invalid environment: ${ENVIRONMENT}${NC}"
    usage
fi

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

# Pre-deployment validation
validate_prerequisites() {
    log "Validating prerequisites..."

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        error "kubectl is not installed"
        exit 1
    fi

    # Check kustomize
    if ! command -v kustomize &> /dev/null; then
        error "kustomize is not installed"
        exit 1
    fi

    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    # Check if namespace exists
    if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
        warning "Namespace ${NAMESPACE} does not exist, it will be created"
    fi

    log "Prerequisites validated"
}

# Run pre-deployment validation
run_validation() {
    log "Running pre-deployment validation..."

    if [ -f "${SCRIPT_DIR}/validate.sh" ]; then
        bash "${SCRIPT_DIR}/validate.sh" -e "${ENVIRONMENT}" || {
            error "Pre-deployment validation failed"
            exit 1
        }
    else
        warning "validate.sh not found, skipping validation"
    fi
}

# Build deployment manifests
build_manifests() {
    log "Building deployment manifests for ${ENVIRONMENT}..."

    MANIFEST_DIR="${BASE_DIR}/overlays/${ENVIRONMENT}"

    if [ ! -d "${MANIFEST_DIR}" ]; then
        error "Environment directory not found: ${MANIFEST_DIR}"
        exit 1
    fi

    # Generate manifests
    if [ "${DRY_RUN}" = true ]; then
        kustomize build "${MANIFEST_DIR}"
    else
        kustomize build "${MANIFEST_DIR}" > /tmp/virtuoso-deployment.yaml

        # Update image tag if specified
        if [ -n "${IMAGE_TAG}" ]; then
            log "Updating image tag to: ${IMAGE_TAG}"
            sed -i "s|image: virtuoso-api-cli:.*|image: virtuoso-api-cli:${IMAGE_TAG}|g" /tmp/virtuoso-deployment.yaml
        fi
    fi
}

# Deploy to Kubernetes
deploy() {
    if [ "${DRY_RUN}" = true ]; then
        log "Dry run mode - no changes will be made"
        return
    fi

    log "Deploying to ${ENVIRONMENT}..."

    # Apply manifests
    kubectl apply -f /tmp/virtuoso-deployment.yaml

    # Wait for rollout
    log "Waiting for deployment rollout..."
    kubectl rollout status deployment/virtuoso-api-cli -n ${NAMESPACE} --timeout=${WAIT_TIMEOUT} || {
        error "Deployment rollout failed"

        # Show recent events
        log "Recent events:"
        kubectl get events -n ${NAMESPACE} --sort-by='.lastTimestamp' | tail -10

        # Show pod status
        log "Pod status:"
        kubectl get pods -n ${NAMESPACE} -l app=virtuoso-api-cli

        exit 1
    }

    log "Deployment successful!"
}

# Post-deployment checks
post_deployment_checks() {
    if [ "${DRY_RUN}" = true ]; then
        return
    fi

    log "Running post-deployment checks..."

    # Check pod status
    READY_PODS=$(kubectl get pods -n ${NAMESPACE} -l app=virtuoso-api-cli -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}' | grep -o "True" | wc -l)
    TOTAL_PODS=$(kubectl get pods -n ${NAMESPACE} -l app=virtuoso-api-cli --no-headers | wc -l)

    if [ "${READY_PODS}" -ne "${TOTAL_PODS}" ]; then
        warning "Not all pods are ready: ${READY_PODS}/${TOTAL_PODS}"
    else
        log "All pods are ready: ${READY_PODS}/${TOTAL_PODS}"
    fi

    # Check service endpoints
    ENDPOINTS=$(kubectl get endpoints virtuoso-api-cli -n ${NAMESPACE} -o jsonpath='{.subsets[*].addresses[*].ip}' | wc -w)
    if [ "${ENDPOINTS}" -eq 0 ]; then
        error "No service endpoints available"
    else
        log "Service has ${ENDPOINTS} endpoints"
    fi

    # Run health check
    log "Waiting 30 seconds before health check..."
    sleep 30

    # Trigger health check job
    kubectl create job --from=cronjob/virtuoso-health-check health-check-manual-$(date +%s) -n ${NAMESPACE} || true
}

# Show deployment summary
show_summary() {
    if [ "${DRY_RUN}" = true ]; then
        return
    fi

    echo ""
    log "Deployment Summary:"
    echo "======================="
    echo "Environment: ${ENVIRONMENT}"
    echo "Namespace: ${NAMESPACE}"
    if [ -n "${IMAGE_TAG}" ]; then
        echo "Image Tag: ${IMAGE_TAG}"
    fi
    echo ""

    # Show resource status
    kubectl get deployments,services,ingress -n ${NAMESPACE} -l app=virtuoso-api-cli

    echo ""
    echo "To view logs:"
    echo "  kubectl logs -n ${NAMESPACE} -l app=virtuoso-api-cli -f"
    echo ""
    echo "To check metrics:"
    echo "  kubectl port-forward -n ${NAMESPACE} svc/virtuoso-api-cli 8000:8000"
    echo "  curl http://localhost:8000/metrics"
}

# Main execution
main() {
    log "Starting deployment to ${ENVIRONMENT}"

    # Confirm deployment
    if ! confirm "Deploy to ${ENVIRONMENT} environment?"; then
        log "Deployment cancelled"
        exit 0
    fi

    # Run deployment steps
    validate_prerequisites
    run_validation
    build_manifests
    deploy
    post_deployment_checks
    show_summary

    log "Deployment completed!"
}

# Execute main function
main
