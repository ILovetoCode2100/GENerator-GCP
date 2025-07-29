#!/bin/bash
# setup-environment.sh - Setup and validate environment for D365 Virtuoso test deployment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname $(dirname "$SCRIPT_DIR"))"

# Log function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Create necessary directories
setup_directories() {
    log "Creating deployment directories..."

    mkdir -p "$PROJECT_ROOT/deployment/config"
    mkdir -p "$PROJECT_ROOT/deployment/scripts"
    mkdir -p "$PROJECT_ROOT/deployment/state"
    mkdir -p "$PROJECT_ROOT/deployment/logs"
    mkdir -p "$PROJECT_ROOT/deployment/backups"
    mkdir -p "$PROJECT_ROOT/deployment/reports"
    mkdir -p "$PROJECT_ROOT/deployment/processed-tests"

    success "Deployment directories created"
}

# Validate environment variables
validate_environment() {
    log "Validating environment variables..."

    local errors=0

    # Check D365_INSTANCE
    if [ -z "${D365_INSTANCE:-}" ]; then
        error "D365_INSTANCE environment variable is not set"
        echo "  Please set: export D365_INSTANCE=your-instance-name"
        ((errors++))
    else
        success "D365_INSTANCE is set: $D365_INSTANCE"
    fi

    # Check VIRTUOSO_API_TOKEN
    if [ -z "${VIRTUOSO_API_TOKEN:-}" ]; then
        error "VIRTUOSO_API_TOKEN environment variable is not set"
        echo "  Please set: export VIRTUOSO_API_TOKEN=your-api-token"
        ((errors++))
    else
        success "VIRTUOSO_API_TOKEN is set (hidden for security)"
    fi

    # Validate D365 instance URL format
    if [ -n "${D365_INSTANCE:-}" ]; then
        # Test if the instance URL is reachable
        local test_url="https://${D365_INSTANCE}.crm.dynamics.com"
        log "Testing D365 instance URL: $test_url"

        if curl -s --head --request GET "$test_url" | grep "200\|302\|401" > /dev/null; then
            success "D365 instance URL is reachable"
        else
            warning "Could not verify D365 instance URL. This might be due to authentication requirements."
        fi
    fi

    if [ $errors -gt 0 ]; then
        error "Environment validation failed with $errors errors"
        return 1
    fi

    success "Environment validation passed"
}

# Check API CLI availability
check_api_cli() {
    log "Checking API CLI availability..."

    local api_cli="$PROJECT_ROOT/bin/api-cli"

    if [ ! -f "$api_cli" ]; then
        error "API CLI not found at: $api_cli"
        echo "  Please build the CLI first: make build"
        return 1
    fi

    if [ ! -x "$api_cli" ]; then
        error "API CLI is not executable"
        echo "  Please make it executable: chmod +x $api_cli"
        return 1
    fi

    # Test API CLI
    if "$api_cli" --version &>/dev/null; then
        success "API CLI is available and working"
    else
        error "API CLI test failed"
        return 1
    fi
}

# Validate Virtuoso configuration
validate_virtuoso_config() {
    log "Validating Virtuoso configuration..."

    local config_file="$HOME/.api-cli/virtuoso-config.yaml"

    # Create config directory if it doesn't exist
    mkdir -p "$HOME/.api-cli"

    # Create or update config file
    cat > "$config_file" <<EOF
api:
  auth_token: ${VIRTUOSO_API_TOKEN}
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"
EOF

    success "Virtuoso configuration updated"

    # Test API connection
    log "Testing Virtuoso API connection..."
    if "$PROJECT_ROOT/bin/api-cli" list projects --output json &>/dev/null; then
        success "Virtuoso API connection successful"
    else
        error "Failed to connect to Virtuoso API"
        echo "  Please check your API token and network connection"
        return 1
    fi
}

# Count and validate test files
validate_test_files() {
    log "Validating test files..."

    local test_dir="$PROJECT_ROOT/d365-virtuoso-tests-final"

    if [ ! -d "$test_dir" ]; then
        error "Test directory not found: $test_dir"
        return 1
    fi

    local total_files=$(find "$test_dir" -name "*.yaml" -type f | wc -l | tr -d ' ')
    log "Found $total_files test files"

    if [ "$total_files" -ne 169 ]; then
        warning "Expected 169 test files, found $total_files"
    fi

    # Check for files with [instance] placeholder
    local files_with_placeholder=$(grep -l "\[instance\]" "$test_dir"/*.yaml "$test_dir"/*/*.yaml 2>/dev/null | wc -l | tr -d ' ')

    if [ "$files_with_placeholder" -gt 0 ]; then
        log "Found $files_with_placeholder files with [instance] placeholder that need updating"
    fi

    success "Test files validated"
}

# Create environment file template
create_env_template() {
    log "Creating environment file template..."

    local env_file="$PROJECT_ROOT/deployment/.env.template"

    cat > "$env_file" <<'EOF'
# D365 Virtuoso Test Deployment Environment Variables
# Copy this file to .env and fill in your values

# D365 Instance Configuration
# Your D365 instance name (without .crm.dynamics.com)
# Example: contoso-dev, mycompany-test
D365_INSTANCE=your-instance-name

# Virtuoso API Configuration
# Get your API token from Virtuoso platform
VIRTUOSO_API_TOKEN=your-api-token-here

# Optional: Override default organization ID
# VIRTUOSO_ORG_ID=2242

# Optional: Deployment settings
# DEPLOYMENT_PARALLEL_UPLOADS=5
# DEPLOYMENT_BATCH_SIZE=10
# DEPLOYMENT_RETRY_ATTEMPTS=3

# Optional: Enable debug logging
# DEBUG=true
EOF

    success "Environment template created at: $env_file"

    # Create .gitignore for deployment directory
    cat > "$PROJECT_ROOT/deployment/.gitignore" <<EOF
# Ignore sensitive files
.env
*.log
state/
backups/
reports/
processed-tests/
EOF
}

# Main setup function
main() {
    echo -e "${BLUE}=== D365 Virtuoso Test Deployment Environment Setup ===${NC}"
    echo

    # Run all setup steps
    setup_directories || exit 1

    if ! validate_environment; then
        echo
        error "Environment setup incomplete. Please fix the errors above and run again."
        exit 1
    fi

    check_api_cli || exit 1
    validate_virtuoso_config || exit 1
    validate_test_files || exit 1
    create_env_template || exit 1

    echo
    success "Environment setup completed successfully!"
    echo
    echo "Next steps:"
    echo "1. Review the environment template at: deployment/.env.template"
    echo "2. Ensure your environment variables are set correctly"
    echo "3. Run the test preprocessor: ./deployment/scripts/preprocess-tests.sh"
    echo "4. Run the deployment: ./deployment/scripts/deploy-tests.sh"
}

# Run main function
main "$@"
