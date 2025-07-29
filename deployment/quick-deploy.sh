#!/bin/bash
# quick-deploy.sh - Quick deployment script for D365 Virtuoso tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Log functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Main function
main() {
    echo -e "${BLUE}=== D365 Virtuoso Test Quick Deployment ===${NC}"
    echo

    # Check environment variables
    if [ -z "${D365_INSTANCE:-}" ]; then
        error "D365_INSTANCE environment variable is not set"
        echo
        echo "Please set your D365 instance name:"
        echo "  export D365_INSTANCE=your-instance-name"
        echo
        echo "Example:"
        echo "  export D365_INSTANCE=contoso-dev"
        exit 1
    fi

    if [ -z "${VIRTUOSO_API_TOKEN:-}" ]; then
        error "VIRTUOSO_API_TOKEN environment variable is not set"
        echo
        echo "Please set your Virtuoso API token:"
        echo "  export VIRTUOSO_API_TOKEN=your-api-token"
        echo
        echo "Get your API token from the Virtuoso platform"
        exit 1
    fi

    log "Using D365 instance: $D365_INSTANCE"
    log "API token is configured"
    echo

    # Make scripts executable
    chmod +x "$SCRIPT_DIR"/scripts/*.sh

    # Step 1: Setup environment
    log "Step 1/4: Setting up environment..."
    if "$SCRIPT_DIR/scripts/setup-environment.sh"; then
        success "Environment setup complete"
    else
        error "Environment setup failed"
        exit 1
    fi

    echo
    sleep 2

    # Step 2: Preprocess tests
    log "Step 2/4: Preprocessing test files..."
    if "$SCRIPT_DIR/scripts/preprocess-tests.sh"; then
        success "Test preprocessing complete"
    else
        error "Test preprocessing failed"
        exit 1
    fi

    echo
    sleep 2

    # Step 3: Deploy tests
    log "Step 3/4: Deploying tests to Virtuoso..."
    if "$SCRIPT_DIR/scripts/deploy-tests.sh"; then
        success "Test deployment complete"
    else
        error "Test deployment failed"
        echo
        echo "You can resume deployment with:"
        echo "  $SCRIPT_DIR/scripts/deploy-tests.sh --continue"
        exit 1
    fi

    echo
    sleep 2

    # Step 4: Validate deployment
    log "Step 4/4: Validating deployment..."
    if "$SCRIPT_DIR/scripts/validate-deployment.sh"; then
        success "Deployment validation complete"
    else
        error "Deployment validation failed"
        echo
        echo "Check the health report for details"
    fi

    echo
    success "Quick deployment completed successfully!"
    echo
    echo "Next steps:"
    echo "1. Review deployment reports in: deployment/reports/"
    echo "2. Access your tests in the Virtuoso platform"
    echo "3. Configure test execution schedules as needed"
    echo
    echo "Useful commands:"
    echo "  View deployment status: cat deployment/state/deployment-state.json | jq"
    echo "  Rollback if needed: ./deployment/scripts/rollback-deployment.sh"
    echo "  Validate health: ./deployment/scripts/validate-deployment.sh"
}

# Run main function
main "$@"
