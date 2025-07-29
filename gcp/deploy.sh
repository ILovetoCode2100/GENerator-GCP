#!/bin/bash
# Master Deployment Script for Virtuoso API CLI on GCP
# This script handles end-to-end deployment of the entire infrastructure

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Default values
PROJECT_ID="${GCP_PROJECT_ID:-}"
REGION="${GCP_REGION:-us-central1}"
ENVIRONMENT="${ENVIRONMENT:-production}"
DRY_RUN=false
SKIP_TERRAFORM=false
SKIP_BUILD=false
SKIP_FUNCTIONS=false
SKIP_MONITORING=false

# Progress tracking
STEPS_COMPLETED=0
TOTAL_STEPS=10

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

# Function to show progress
show_progress() {
    STEPS_COMPLETED=$((STEPS_COMPLETED + 1))
    local PERCENT=$((STEPS_COMPLETED * 100 / TOTAL_STEPS))
    echo -e "\n${GREEN}Progress: [$STEPS_COMPLETED/$TOTAL_STEPS] ${PERCENT}%${NC}"
    echo -e "${BLUE}===================================================${NC}\n"
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."

    local missing_tools=()

    # Check required tools
    command -v gcloud >/dev/null 2>&1 || missing_tools+=("gcloud")
    command -v terraform >/dev/null 2>&1 || missing_tools+=("terraform")
    command -v docker >/dev/null 2>&1 || missing_tools+=("docker")
    command -v jq >/dev/null 2>&1 || missing_tools+=("jq")
    command -v go >/dev/null 2>&1 || missing_tools+=("go")

    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_info "Please install missing tools and try again."
        exit 1
    fi

    # Check GCP authentication
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        print_error "No active GCP authentication found"
        print_info "Run 'gcloud auth login' to authenticate"
        exit 1
    fi

    # Check project ID
    if [ -z "$PROJECT_ID" ]; then
        PROJECT_ID=$(gcloud config get-value project 2>/dev/null || echo "")
        if [ -z "$PROJECT_ID" ]; then
            print_error "No GCP project ID specified"
            print_info "Set GCP_PROJECT_ID environment variable or run 'gcloud config set project PROJECT_ID'"
            exit 1
        fi
    fi

    print_success "All prerequisites met"
    show_progress
}

# Function to enable required APIs
enable_apis() {
    print_info "Enabling required GCP APIs..."

    local apis=(
        "cloudrun.googleapis.com"
        "cloudbuild.googleapis.com"
        "cloudfunctions.googleapis.com"
        "cloudscheduler.googleapis.com"
        "cloudtasks.googleapis.com"
        "pubsub.googleapis.com"
        "firestore.googleapis.com"
        "secretmanager.googleapis.com"
        "monitoring.googleapis.com"
        "logging.googleapis.com"
        "compute.googleapis.com"
        "storage.googleapis.com"
    )

    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would enable APIs: ${apis[*]}"
    else
        for api in "${apis[@]}"; do
            print_info "Enabling $api..."
            gcloud services enable "$api" --project="$PROJECT_ID" || true
        done
    fi

    print_success "APIs enabled"
    show_progress
}

# Function to setup Terraform
setup_terraform() {
    if [ "$SKIP_TERRAFORM" = true ]; then
        print_warning "Skipping Terraform setup"
        show_progress
        return
    fi

    print_info "Setting up Terraform infrastructure..."

    cd "$SCRIPT_DIR/terraform"

    # Initialize Terraform
    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would initialize Terraform"
    else
        terraform init -backend-config="bucket=${PROJECT_ID}-terraform-state"
    fi

    # Create terraform.tfvars
    cat > terraform.tfvars <<EOF
project_id = "$PROJECT_ID"
region = "$REGION"
environment = "$ENVIRONMENT"
virtuoso_api_key = "$VIRTUOSO_API_KEY"
virtuoso_org_id = "$VIRTUOSO_ORG_ID"
EOF

    # Plan Terraform changes
    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would plan Terraform changes"
    else
        terraform plan -out=tfplan
    fi

    # Apply Terraform changes
    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would apply Terraform changes"
    else
        read -p "Apply Terraform changes? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            terraform apply tfplan
        else
            print_warning "Terraform apply skipped"
        fi
    fi

    cd "$SCRIPT_DIR"
    print_success "Terraform infrastructure deployed"
    show_progress
}

# Function to build CLI binary
build_cli() {
    if [ "$SKIP_BUILD" = true ]; then
        print_warning "Skipping CLI build"
        show_progress
        return
    fi

    print_info "Building CLI binary..."

    cd "$PROJECT_ROOT"

    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would build CLI binary"
    else
        make build

        # Verify binary exists
        if [ ! -f "bin/api-cli" ]; then
            print_error "CLI binary not found after build"
            exit 1
        fi
    fi

    print_success "CLI binary built"
    show_progress
}

# Function to build and deploy Cloud Run service
deploy_cloud_run() {
    print_info "Deploying Cloud Run service..."

    cd "$PROJECT_ROOT"

    # Build container image
    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would build container image"
    else
        gcloud builds submit --tag "gcr.io/$PROJECT_ID/virtuoso-api-cli:$ENVIRONMENT" .
    fi

    # Deploy to Cloud Run
    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would deploy to Cloud Run"
    else
        gcloud run deploy virtuoso-api-cli \
            --image "gcr.io/$PROJECT_ID/virtuoso-api-cli:$ENVIRONMENT" \
            --platform managed \
            --region "$REGION" \
            --allow-unauthenticated \
            --set-env-vars "ENVIRONMENT=$ENVIRONMENT" \
            --set-env-vars "GCP_PROJECT_ID=$PROJECT_ID" \
            --set-secrets "VIRTUOSO_API_KEY=virtuoso-api-key:latest" \
            --set-secrets "VIRTUOSO_ORG_ID=virtuoso-org-id:latest" \
            --memory 2Gi \
            --cpu 2 \
            --timeout 300 \
            --concurrency 100 \
            --max-instances 10
    fi

    print_success "Cloud Run service deployed"
    show_progress
}

# Function to deploy Cloud Functions
deploy_functions() {
    if [ "$SKIP_FUNCTIONS" = true ]; then
        print_warning "Skipping Cloud Functions deployment"
        show_progress
        return
    fi

    print_info "Deploying Cloud Functions..."

    cd "$SCRIPT_DIR/functions"

    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would deploy all Cloud Functions"
    else
        ./deploy-all.sh
    fi

    cd "$SCRIPT_DIR"
    print_success "Cloud Functions deployed"
    show_progress
}

# Function to setup Cloud Build triggers
setup_cloud_build() {
    print_info "Setting up Cloud Build triggers..."

    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would create Cloud Build triggers"
    else
        # Create trigger for main branch
        gcloud builds triggers create github \
            --repo-name="virtuoso-GENerator" \
            --repo-owner="$GITHUB_OWNER" \
            --branch-pattern="^main$" \
            --build-config="cloudbuild.yaml" \
            --name="deploy-main" \
            --description="Deploy on push to main branch"

        # Create trigger for tags
        gcloud builds triggers create github \
            --repo-name="virtuoso-GENerator" \
            --repo-owner="$GITHUB_OWNER" \
            --tag-pattern="^v[0-9]+\.[0-9]+\.[0-9]+$" \
            --build-config="cloudbuild.yaml" \
            --name="deploy-release" \
            --description="Deploy on version tags"
    fi

    print_success "Cloud Build triggers created"
    show_progress
}

# Function to setup monitoring
setup_monitoring() {
    if [ "$SKIP_MONITORING" = true ]; then
        print_warning "Skipping monitoring setup"
        show_progress
        return
    fi

    print_info "Setting up monitoring..."

    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would run monitoring setup"
    else
        ./monitoring-setup.sh
    fi

    print_success "Monitoring configured"
    show_progress
}

# Function to run smoke tests
run_smoke_tests() {
    print_info "Running smoke tests..."

    if [ "$DRY_RUN" = true ]; then
        print_info "[DRY RUN] Would run smoke tests"
        show_progress
        return
    fi

    # Get Cloud Run service URL
    SERVICE_URL=$(gcloud run services describe virtuoso-api-cli \
        --platform managed \
        --region "$REGION" \
        --format 'value(status.url)')

    if [ -z "$SERVICE_URL" ]; then
        print_error "Could not get Cloud Run service URL"
        exit 1
    fi

    print_info "Testing service at: $SERVICE_URL"

    # Test health endpoint
    if curl -s "$SERVICE_URL/health" | jq -e '.status == "healthy"' > /dev/null; then
        print_success "Health check passed"
    else
        print_error "Health check failed"
        exit 1
    fi

    # Test API endpoint
    if curl -s -H "X-API-Key: $VIRTUOSO_API_KEY" "$SERVICE_URL/api/v1/commands/list" | jq -e '.commands | length > 0' > /dev/null; then
        print_success "API endpoint test passed"
    else
        print_error "API endpoint test failed"
        exit 1
    fi

    print_success "All smoke tests passed"
    show_progress
}

# Function to print deployment summary
print_summary() {
    print_info "Deployment Summary"
    echo -e "${BLUE}===================================================${NC}"
    echo -e "Project ID: ${GREEN}$PROJECT_ID${NC}"
    echo -e "Region: ${GREEN}$REGION${NC}"
    echo -e "Environment: ${GREEN}$ENVIRONMENT${NC}"

    if [ "$DRY_RUN" = false ]; then
        # Get service URL
        SERVICE_URL=$(gcloud run services describe virtuoso-api-cli \
            --platform managed \
            --region "$REGION" \
            --format 'value(status.url)' 2>/dev/null || echo "Not deployed")

        echo -e "Service URL: ${GREEN}$SERVICE_URL${NC}"
        echo -e "\nNext steps:"
        echo -e "1. Visit the service URL to verify deployment"
        echo -e "2. Check Cloud Console for monitoring dashboards"
        echo -e "3. Review Cloud Build history for CI/CD pipeline"
        echo -e "4. Configure alerting policies as needed"
    fi

    echo -e "${BLUE}===================================================${NC}"
}

# Function to handle errors and rollback
handle_error() {
    print_error "Deployment failed at step $STEPS_COMPLETED"

    if [ "$DRY_RUN" = false ]; then
        read -p "Attempt rollback? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            ./rollback.sh
        fi
    fi

    exit 1
}

# Set error handler
trap handle_error ERR

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
        --environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-terraform)
            SKIP_TERRAFORM=true
            TOTAL_STEPS=$((TOTAL_STEPS - 1))
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            TOTAL_STEPS=$((TOTAL_STEPS - 1))
            shift
            ;;
        --skip-functions)
            SKIP_FUNCTIONS=true
            TOTAL_STEPS=$((TOTAL_STEPS - 1))
            shift
            ;;
        --skip-monitoring)
            SKIP_MONITORING=true
            TOTAL_STEPS=$((TOTAL_STEPS - 1))
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --project-id ID      GCP project ID"
            echo "  --region REGION      GCP region (default: us-central1)"
            echo "  --environment ENV    Environment (default: production)"
            echo "  --dry-run           Show what would be done without doing it"
            echo "  --skip-terraform    Skip Terraform infrastructure setup"
            echo "  --skip-build        Skip CLI binary build"
            echo "  --skip-functions    Skip Cloud Functions deployment"
            echo "  --skip-monitoring   Skip monitoring setup"
            echo "  --help              Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Main deployment flow
print_info "Starting Virtuoso API CLI deployment to GCP"
echo -e "${BLUE}===================================================${NC}\n"

# Check required environment variables
if [ -z "${VIRTUOSO_API_KEY:-}" ]; then
    print_error "VIRTUOSO_API_KEY environment variable not set"
    exit 1
fi

if [ -z "${VIRTUOSO_ORG_ID:-}" ]; then
    print_error "VIRTUOSO_ORG_ID environment variable not set"
    exit 1
fi

# Execute deployment steps
check_prerequisites
enable_apis
setup_terraform
build_cli
deploy_cloud_run
deploy_functions
setup_cloud_build
setup_monitoring
run_smoke_tests
print_summary

print_success "Deployment completed successfully!"
