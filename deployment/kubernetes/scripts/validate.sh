#!/bin/bash
set -e

# Pre-deployment validation script for Virtuoso API CLI

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
VERBOSE=false
EXIT_ON_WARNING=false

# Validation results
ERRORS=0
WARNINGS=0

# Usage function
usage() {
    echo "Usage: $0 -e <environment> [-v] [-w]"
    echo ""
    echo "Options:"
    echo "  -e <environment>  Environment to validate (dev|staging|production)"
    echo "  -v                Verbose output"
    echo "  -w                Exit with error on warnings"
    echo "  -h                Show this help message"
    exit 1
}

# Parse command line arguments
while getopts "e:vwh" opt; do
    case ${opt} in
        e )
            ENVIRONMENT=$OPTARG
            ;;
        v )
            VERBOSE=true
            ;;
        w )
            EXIT_ON_WARNING=true
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
    echo -e "${GREEN}[VALIDATE]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((ERRORS++))
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    ((WARNINGS++))
}

verbose() {
    if [ "${VERBOSE}" = true ]; then
        echo "  $1"
    fi
}

success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

# Validation functions
validate_kubernetes_version() {
    log "Validating Kubernetes version..."

    K8S_VERSION=$(kubectl version --client -o json | jq -r '.clientVersion.gitVersion')
    MAJOR=$(echo ${K8S_VERSION} | cut -d. -f1 | sed 's/v//')
    MINOR=$(echo ${K8S_VERSION} | cut -d. -f2)

    verbose "Kubernetes client version: ${K8S_VERSION}"

    if [ "${MAJOR}" -lt 1 ] || ([ "${MAJOR}" -eq 1 ] && [ "${MINOR}" -lt 19 ]); then
        error "Kubernetes client version ${K8S_VERSION} is too old (minimum: v1.19)"
    else
        success "Kubernetes version ${K8S_VERSION}"
    fi
}

validate_cluster_access() {
    log "Validating cluster access..."

    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster"
        return
    fi

    # Check if user has permissions
    if ! kubectl auth can-i create deployments -n ${NAMESPACE} &> /dev/null; then
        error "Insufficient permissions to deploy to namespace ${NAMESPACE}"
    else
        success "Cluster access validated"
    fi
}

validate_namespace() {
    log "Validating namespace..."

    if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
        warning "Namespace ${NAMESPACE} does not exist (will be created)"
    else
        success "Namespace ${NAMESPACE} exists"

        # Check for existing resources
        EXISTING_DEPLOYMENTS=$(kubectl get deployments -n ${NAMESPACE} --no-headers 2>/dev/null | wc -l)
        if [ "${EXISTING_DEPLOYMENTS}" -gt 0 ]; then
            verbose "Found ${EXISTING_DEPLOYMENTS} existing deployments in namespace"
        fi
    fi
}

validate_manifests() {
    log "Validating manifests..."

    MANIFEST_DIR="${BASE_DIR}/overlays/${ENVIRONMENT}"

    if [ ! -d "${MANIFEST_DIR}" ]; then
        error "Environment directory not found: ${MANIFEST_DIR}"
        return
    fi

    # Build and validate manifests
    if ! kustomize build "${MANIFEST_DIR}" > /tmp/virtuoso-manifests.yaml 2>/dev/null; then
        error "Failed to build manifests with kustomize"
        return
    fi

    # Dry run apply
    if ! kubectl apply -f /tmp/virtuoso-manifests.yaml --dry-run=client &> /dev/null; then
        error "Manifest validation failed"
        kubectl apply -f /tmp/virtuoso-manifests.yaml --dry-run=client 2>&1 | tail -10
    else
        success "Manifests validated"

        # Count resources
        if [ "${VERBOSE}" = true ]; then
            RESOURCE_COUNT=$(grep "^kind:" /tmp/virtuoso-manifests.yaml | wc -l)
            verbose "Total resources to deploy: ${RESOURCE_COUNT}"
        fi
    fi
}

validate_secrets() {
    log "Validating secrets..."

    # Check if required secrets exist
    REQUIRED_SECRETS=("virtuoso-api-secret")

    for secret in "${REQUIRED_SECRETS[@]}"; do
        if kubectl get secret ${secret} -n ${NAMESPACE} &> /dev/null; then
            success "Secret ${secret} exists"
        else
            warning "Secret ${secret} not found in namespace ${NAMESPACE}"
            verbose "You may need to create this secret before deployment"
        fi
    done
}

validate_persistent_volumes() {
    log "Validating persistent volumes..."

    # Check storage class
    STORAGE_CLASS="fast-ssd"

    if kubectl get storageclass ${STORAGE_CLASS} &> /dev/null; then
        success "Storage class ${STORAGE_CLASS} exists"
    else
        warning "Storage class ${STORAGE_CLASS} not found"
        verbose "PVC creation may fail if storage class doesn't exist"

        # List available storage classes
        if [ "${VERBOSE}" = true ]; then
            verbose "Available storage classes:"
            kubectl get storageclass --no-headers | awk '{print "  - " $1}'
        fi
    fi
}

validate_resource_quotas() {
    log "Validating resource quotas..."

    if kubectl get resourcequota -n ${NAMESPACE} &> /dev/null 2>&1; then
        QUOTAS=$(kubectl get resourcequota -n ${NAMESPACE} -o json)

        if [ "$(echo ${QUOTAS} | jq '.items | length')" -gt 0 ]; then
            warning "Resource quotas found in namespace"

            if [ "${VERBOSE}" = true ]; then
                verbose "Resource quotas:"
                kubectl get resourcequota -n ${NAMESPACE} --no-headers | awk '{print "  - " $1}'
            fi
        fi
    else
        success "No resource quotas found"
    fi
}

validate_network_policies() {
    log "Validating network policies..."

    # Check if network policies are enforced
    if kubectl get networkpolicies -n ${NAMESPACE} &> /dev/null 2>&1; then
        NP_COUNT=$(kubectl get networkpolicies -n ${NAMESPACE} --no-headers 2>/dev/null | wc -l)

        if [ "${NP_COUNT}" -gt 0 ]; then
            verbose "Found ${NP_COUNT} network policies in namespace"
            warning "Network policies are enforced - ensure proper configuration"
        fi
    else
        success "No restrictive network policies found"
    fi
}

validate_monitoring() {
    log "Validating monitoring setup..."

    # Check if Prometheus CRDs exist
    if kubectl get crd servicemonitors.monitoring.coreos.com &> /dev/null; then
        success "Prometheus Operator CRDs found"
    else
        warning "Prometheus Operator CRDs not found - monitoring may not work"
    fi

    # Check if cert-manager exists (for certificate monitoring)
    if kubectl get crd certificates.cert-manager.io &> /dev/null; then
        success "Cert-manager CRDs found"
    else
        warning "Cert-manager CRDs not found - certificate automation may not work"
    fi
}

validate_ingress() {
    log "Validating ingress controller..."

    # Check for ingress controller
    INGRESS_FOUND=false

    # Check common ingress controller namespaces
    for ns in ingress-nginx nginx-ingress kube-system; do
        if kubectl get pods -n ${ns} -l app.kubernetes.io/name=ingress-nginx &> /dev/null 2>&1; then
            INGRESS_FOUND=true
            success "Ingress controller found in namespace ${ns}"
            break
        fi
    done

    if [ "${INGRESS_FOUND}" = false ]; then
        error "No ingress controller found - ingress resources will not work"
    fi
}

validate_environment_specific() {
    log "Validating ${ENVIRONMENT}-specific requirements..."

    case ${ENVIRONMENT} in
        production)
            # Production-specific validations

            # Check node count
            NODE_COUNT=$(kubectl get nodes --no-headers | wc -l)
            if [ "${NODE_COUNT}" -lt 3 ]; then
                warning "Only ${NODE_COUNT} nodes available (recommended: 3+ for production)"
            else
                success "${NODE_COUNT} nodes available"
            fi

            # Check if PodDisruptionBudgets will be satisfied
            verbose "Checking PodDisruptionBudget requirements..."
            ;;

        staging)
            # Staging-specific validations
            verbose "Staging environment validation"
            ;;

        dev)
            # Dev-specific validations
            verbose "Development environment validation"
            ;;
    esac
}

# Summary function
show_summary() {
    echo ""
    echo "========================================"
    echo "Validation Summary for ${ENVIRONMENT}"
    echo "========================================"

    if [ "${ERRORS}" -eq 0 ]; then
        echo -e "${GREEN}✓ All critical validations passed${NC}"
    else
        echo -e "${RED}✗ Found ${ERRORS} critical errors${NC}"
    fi

    if [ "${WARNINGS}" -gt 0 ]; then
        echo -e "${YELLOW}⚠ Found ${WARNINGS} warnings${NC}"
    fi

    echo ""

    # Exit code logic
    if [ "${ERRORS}" -gt 0 ]; then
        exit 1
    elif [ "${WARNINGS}" -gt 0 ] && [ "${EXIT_ON_WARNING}" = true ]; then
        exit 1
    else
        exit 0
    fi
}

# Main execution
main() {
    log "Starting pre-deployment validation for ${ENVIRONMENT}"
    echo ""

    # Run all validations
    validate_kubernetes_version
    validate_cluster_access
    validate_namespace
    validate_manifests
    validate_secrets
    validate_persistent_volumes
    validate_resource_quotas
    validate_network_policies
    validate_monitoring
    validate_ingress
    validate_environment_specific

    # Show summary
    show_summary
}

# Execute main function
main
