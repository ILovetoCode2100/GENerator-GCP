#!/bin/bash

# Virtuoso API CLI - Render Deployment Script
# This script handles deployment to Render with validation and configuration

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="${RENDER_SERVICE_NAME:-virtuoso-api-cli}"
RENDER_API_KEY="${RENDER_API_KEY:-}"
GIT_BRANCH="${GIT_BRANCH:-main}"
ENVIRONMENT="${ENVIRONMENT:-production}"

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." &> /dev/null && pwd )"

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

check_requirements() {
    log_info "Checking requirements..."

    # Check if render CLI is installed
    if ! command -v render &> /dev/null; then
        log_error "Render CLI not found. Please install it first:"
        echo "  brew install render"
        echo "  or visit: https://render.com/docs/cli"
        exit 1
    fi

    # Check if git is installed
    if ! command -v git &> /dev/null; then
        log_error "Git not found. Please install Git first."
        exit 1
    fi

    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository. Please run from the project root."
        exit 1
    fi

    log_success "All requirements met"
}

validate_environment() {
    log_info "Validating environment variables..."

    local missing_vars=()

    # Check required environment variables
    if [ -z "$VIRTUOSO_API_TOKEN" ]; then
        missing_vars+=("VIRTUOSO_API_TOKEN")
    fi

    if [ -z "$VIRTUOSO_ORG_ID" ]; then
        missing_vars+=("VIRTUOSO_ORG_ID")
    fi

    if [ ${#missing_vars[@]} -ne 0 ]; then
        log_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        echo ""
        echo "Please set these variables or create a .env file"
        exit 1
    fi

    log_success "Environment variables validated"
}

load_env_file() {
    local env_file="$1"

    if [ -f "$env_file" ]; then
        log_info "Loading environment from $env_file"
        set -a
        source "$env_file"
        set +a
    fi
}

check_git_status() {
    log_info "Checking git status..."

    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        log_warning "You have uncommitted changes:"
        git status --short
        echo ""
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Deployment cancelled"
            exit 0
        fi
    fi

    # Get current branch
    local current_branch=$(git branch --show-current)
    if [ "$current_branch" != "$GIT_BRANCH" ]; then
        log_warning "You're on branch '$current_branch' but deploying from '$GIT_BRANCH'"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Deployment cancelled"
            exit 0
        fi
    fi

    log_success "Git status checked"
}

setup_render_config() {
    log_info "Setting up Render configuration..."

    # Create render.yaml if it doesn't exist
    if [ ! -f "$PROJECT_ROOT/render.yaml" ]; then
        cat > "$PROJECT_ROOT/render.yaml" << EOF
services:
  - type: web
    name: ${SERVICE_NAME}
    env: docker
    dockerfilePath: ./Dockerfile
    envVars:
      - key: VIRTUOSO_API_TOKEN
        sync: false
      - key: VIRTUOSO_ORG_ID
        sync: false
      - key: VIRTUOSO_BASE_URL
        value: https://api-app2.virtuoso.qa/api
      - key: PORT
        value: 8080
      - key: LOG_LEVEL
        value: info
    healthCheckPath: /health
    autoDeploy: true
EOF
        log_success "Created render.yaml"
    else
        log_info "render.yaml already exists"
    fi
}

build_docker_image() {
    log_info "Building Docker image locally for validation..."

    cd "$PROJECT_ROOT"

    if docker build -t virtuoso-api-cli:latest .; then
        log_success "Docker build successful"

        # Test the image
        log_info "Testing Docker image..."
        if docker run --rm virtuoso-api-cli:latest ./api-cli --version; then
            log_success "Docker image test passed"
        else
            log_error "Docker image test failed"
            exit 1
        fi
    else
        log_error "Docker build failed"
        exit 1
    fi
}

deploy_to_render() {
    log_info "Deploying to Render..."

    cd "$PROJECT_ROOT"

    # Set environment variables
    if [ -n "$RENDER_API_KEY" ]; then
        export RENDER_API_KEY
    fi

    # Deploy using render CLI
    if render deploy --service "$SERVICE_NAME"; then
        log_success "Deployment initiated successfully"
    else
        log_error "Deployment failed"
        exit 1
    fi
}

wait_for_deployment() {
    log_info "Waiting for deployment to complete..."

    local max_attempts=60
    local attempt=0
    local service_url="https://${SERVICE_NAME}.onrender.com"

    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "${service_url}/health" > /dev/null 2>&1; then
            log_success "Deployment completed successfully!"
            echo "Service URL: $service_url"
            return 0
        fi

        attempt=$((attempt + 1))
        echo -n "."
        sleep 5
    done

    echo ""
    log_error "Deployment health check timed out"
    return 1
}

run_post_deployment_checks() {
    log_info "Running post-deployment checks..."

    # Run the health check script
    if [ -f "$SCRIPT_DIR/health-check.sh" ]; then
        bash "$SCRIPT_DIR/health-check.sh"
    else
        log_warning "health-check.sh not found, skipping post-deployment checks"
    fi
}

show_deployment_summary() {
    echo ""
    echo "======================================"
    echo "   Deployment Summary"
    echo "======================================"
    echo "Service Name: $SERVICE_NAME"
    echo "Environment: $ENVIRONMENT"
    echo "Git Branch: $GIT_BRANCH"
    echo "Service URL: https://${SERVICE_NAME}.onrender.com"
    echo "Dashboard: https://dashboard.render.com/web/${SERVICE_NAME}"
    echo "======================================"
}

# Main execution
main() {
    echo "======================================"
    echo "   Virtuoso API CLI - Render Deploy"
    echo "======================================"
    echo ""

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --env-file)
                ENV_FILE="$2"
                shift 2
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-checks)
                SKIP_CHECKS=true
                shift
                ;;
            --service)
                SERVICE_NAME="$2"
                shift 2
                ;;
            --branch)
                GIT_BRANCH="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --env-file FILE    Load environment from file"
                echo "  --skip-build       Skip local Docker build test"
                echo "  --skip-checks      Skip pre-deployment checks"
                echo "  --service NAME     Render service name (default: virtuoso-api-cli)"
                echo "  --branch BRANCH    Git branch to deploy (default: main)"
                echo "  --help             Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # Load environment file if specified
    if [ -n "$ENV_FILE" ]; then
        load_env_file "$ENV_FILE"
    else
        # Try to load default .env files
        load_env_file "$SCRIPT_DIR/.env"
        load_env_file "$PROJECT_ROOT/.env"
    fi

    # Run deployment steps
    check_requirements

    if [ "$SKIP_CHECKS" != "true" ]; then
        validate_environment
        check_git_status
    fi

    setup_render_config

    if [ "$SKIP_BUILD" != "true" ] && command -v docker &> /dev/null; then
        build_docker_image
    fi

    deploy_to_render
    wait_for_deployment
    run_post_deployment_checks
    show_deployment_summary

    log_success "Deployment completed!"
}

# Run main function
main "$@"
