#!/bin/bash

# Virtuoso API CLI Helm Chart Installation Script
# This script helps with installing and managing the Virtuoso API CLI Helm chart

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
NAMESPACE="virtuoso-api"
RELEASE_NAME="virtuoso-api-cli"
CHART_PATH="./virtuoso-api-cli"
VALUES_FILE=""
ENVIRONMENT=""

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS] COMMAND

Commands:
    install     Install the Virtuoso API CLI Helm chart
    upgrade     Upgrade an existing installation
    uninstall   Uninstall the Helm release
    template    Render templates locally for debugging
    lint        Lint the Helm chart
    package     Package the Helm chart

Options:
    -n, --namespace NAME        Kubernetes namespace (default: virtuoso-api)
    -r, --release NAME          Helm release name (default: virtuoso-api-cli)
    -f, --values FILE           Values file to use
    -e, --environment ENV       Environment (dev, staging, production)
    --api-token TOKEN          API token (required for install)
    --org-id ID                Organization ID (required for install)
    --dry-run                  Perform a dry run
    -h, --help                 Show this help message

Examples:
    # Install for development
    $0 -e dev --api-token "your-token" --org-id "2242" install

    # Install for production with custom values
    $0 -e production -f custom-values.yaml install

    # Upgrade existing installation
    $0 -n virtuoso-prod -r api-cli-prod upgrade

    # Render templates for debugging
    $0 -e staging template

EOF
}

# Parse command line arguments
ARGS=()
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -r|--release)
            RELEASE_NAME="$2"
            shift 2
            ;;
        -f|--values)
            VALUES_FILE="$2"
            shift 2
            ;;
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --api-token)
            API_TOKEN="$2"
            shift 2
            ;;
        --org-id)
            ORG_ID="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN="--dry-run"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            ARGS+=("$1")
            shift
            ;;
    esac
done

# Restore positional parameters
set -- "${ARGS[@]}"

# Get command
COMMAND=${1:-""}

# Check if command is provided
if [ -z "$COMMAND" ]; then
    print_error "No command provided"
    usage
    exit 1
fi

# Check if Helm is installed
if ! command -v helm &> /dev/null; then
    print_error "Helm is not installed. Please install Helm first."
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Function to build Helm command with values files
build_helm_values_args() {
    local args=""

    # Add base values file
    args="$args -f ${CHART_PATH}/values.yaml"

    # Add environment-specific values if specified
    if [ -n "$ENVIRONMENT" ]; then
        local env_values="${CHART_PATH}/values.${ENVIRONMENT}.yaml"
        if [ -f "$env_values" ]; then
            args="$args -f $env_values"
            print_info "Using environment values: $env_values"
        else
            print_warning "Environment values file not found: $env_values"
        fi
    fi

    # Add custom values file if specified
    if [ -n "$VALUES_FILE" ]; then
        if [ -f "$VALUES_FILE" ]; then
            args="$args -f $VALUES_FILE"
            print_info "Using custom values: $VALUES_FILE"
        else
            print_error "Values file not found: $VALUES_FILE"
            exit 1
        fi
    fi

    # Add API token and org ID if provided
    if [ -n "$API_TOKEN" ]; then
        args="$args --set secret.apiToken=\"$API_TOKEN\""
    fi

    if [ -n "$ORG_ID" ]; then
        args="$args --set config.organization.id=\"$ORG_ID\""
    fi

    echo "$args"
}

# Execute command
case "$COMMAND" in
    install)
        print_info "Installing Virtuoso API CLI Helm chart..."

        # Check required parameters for install
        if [ -z "$API_TOKEN" ] && [ -z "$VALUES_FILE" ]; then
            print_error "API token is required. Use --api-token or provide a values file."
            exit 1
        fi

        if [ -z "$ORG_ID" ] && [ -z "$VALUES_FILE" ]; then
            print_error "Organization ID is required. Use --org-id or provide a values file."
            exit 1
        fi

        # Create namespace if it doesn't exist
        if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
            print_info "Creating namespace: $NAMESPACE"
            kubectl create namespace "$NAMESPACE"
        fi

        # Build Helm command
        VALUES_ARGS=$(build_helm_values_args)

        # Install the chart
        helm install "$RELEASE_NAME" "$CHART_PATH" \
            --namespace "$NAMESPACE" \
            $VALUES_ARGS \
            $DRY_RUN

        if [ -z "$DRY_RUN" ]; then
            print_info "Installation complete!"
            print_info "Run 'kubectl get all -n $NAMESPACE' to check the deployment status."
        fi
        ;;

    upgrade)
        print_info "Upgrading Virtuoso API CLI Helm chart..."

        # Check if release exists
        if ! helm list -n "$NAMESPACE" | grep -q "$RELEASE_NAME"; then
            print_error "Release '$RELEASE_NAME' not found in namespace '$NAMESPACE'"
            exit 1
        fi

        # Build Helm command
        VALUES_ARGS=$(build_helm_values_args)

        # Upgrade the chart
        helm upgrade "$RELEASE_NAME" "$CHART_PATH" \
            --namespace "$NAMESPACE" \
            $VALUES_ARGS \
            $DRY_RUN

        if [ -z "$DRY_RUN" ]; then
            print_info "Upgrade complete!"
        fi
        ;;

    uninstall)
        print_info "Uninstalling Virtuoso API CLI Helm chart..."

        # Confirm uninstall
        if [ -z "$DRY_RUN" ]; then
            read -p "Are you sure you want to uninstall $RELEASE_NAME from namespace $NAMESPACE? (y/N) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_info "Uninstall cancelled."
                exit 0
            fi
        fi

        # Uninstall the release
        helm uninstall "$RELEASE_NAME" \
            --namespace "$NAMESPACE" \
            $DRY_RUN

        if [ -z "$DRY_RUN" ]; then
            print_info "Uninstall complete!"
        fi
        ;;

    template)
        print_info "Rendering Helm templates..."

        # Build Helm command
        VALUES_ARGS=$(build_helm_values_args)

        # Template the chart
        helm template "$RELEASE_NAME" "$CHART_PATH" \
            --namespace "$NAMESPACE" \
            $VALUES_ARGS
        ;;

    lint)
        print_info "Linting Helm chart..."

        # Build Helm command
        VALUES_ARGS=$(build_helm_values_args)

        # Lint the chart
        helm lint "$CHART_PATH" $VALUES_ARGS

        print_info "Lint complete!"
        ;;

    package)
        print_info "Packaging Helm chart..."

        # Package the chart
        helm package "$CHART_PATH"

        print_info "Package complete!"
        ;;

    *)
        print_error "Unknown command: $COMMAND"
        usage
        exit 1
        ;;
esac
