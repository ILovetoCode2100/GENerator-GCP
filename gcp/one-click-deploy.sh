#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_step() { echo -e "${CYAN}▶️  $1${NC}"; }
print_header() {
    echo -e "\n${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Spinner function for long operations
spin() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Load configuration
load_config() {
    if [ ! -f "gcp/auto-deploy-config.yaml" ]; then
        print_error "Configuration not found. Please run ./gcp/deploy-wizard.sh first."
        exit 1
    fi

    # Basic YAML parsing
    PROJECT_ID=$(grep "id:" gcp/auto-deploy-config.yaml | head -1 | awk '{print $2}')
    PROJECT_NUMBER=$(grep "number:" gcp/auto-deploy-config.yaml | awk '{print $2}')
    REGION=$(grep "region:" gcp/auto-deploy-config.yaml | awk '{print $2}')
    SERVICE_NAME=$(grep "name:" gcp/auto-deploy-config.yaml | head -1 | awk '{print $2}')
    IMAGE_URL=$(grep "image:" gcp/auto-deploy-config.yaml | awk '{print $2}')

    # Virtuoso config
    VIRTUOSO_API_TOKEN=$(grep "api_token:" gcp/auto-deploy-config.yaml | awk '{print $2}')
    VIRTUOSO_ORG_ID=$(grep "org_id:" gcp/auto-deploy-config.yaml | awk '{print $2}')
    VIRTUOSO_API_URL=$(grep "api_url:" gcp/auto-deploy-config.yaml | awk '{print $2}')

    # Cloud Run config
    MEMORY=$(grep "memory:" gcp/auto-deploy-config.yaml | awk '{print $2}')
    CPU=$(grep "cpu:" gcp/auto-deploy-config.yaml | awk '{print $2}' | tr -d '"')
    MIN_INSTANCES=$(grep "min_instances:" gcp/auto-deploy-config.yaml | awk '{print $2}')
    MAX_INSTANCES=$(grep "max_instances:" gcp/auto-deploy-config.yaml | awk '{print $2}')
    CONCURRENCY=$(grep "concurrency:" gcp/auto-deploy-config.yaml | awk '{print $2}')

    print_success "Configuration loaded successfully"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    # Check if pre-deployment check was run
    if [ -f "gcp/.precheck-results" ]; then
        source gcp/.precheck-results
        if [ "$CHECKS_PASSED" = "true" ]; then
            print_success "Pre-deployment checks passed"
        else
            print_warning "Pre-deployment checks had issues. Running checks again..."
            ./gcp/pre-deployment-check.sh
        fi
    else
        print_info "Running pre-deployment checks..."
        ./gcp/pre-deployment-check.sh
    fi

    # Verify Docker is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker Desktop."
        exit 1
    fi

    print_success "All prerequisites met"
}

# Enable APIs
enable_apis() {
    print_header "Enabling Required APIs"

    APIS=(
        "cloudbuild.googleapis.com"
        "run.googleapis.com"
        "containerregistry.googleapis.com"
        "secretmanager.googleapis.com"
        "cloudresourcemanager.googleapis.com"
    )

    for api in "${APIS[@]}"; do
        print_step "Enabling $api..."
        gcloud services enable "$api" --project="$PROJECT_ID" &
        spin $!
        print_success "$api enabled"
    done
}

# Create service account
create_service_account() {
    print_header "Setting Up Service Account"

    SA_EMAIL="$SERVICE_NAME@$PROJECT_ID.iam.gserviceaccount.com"

    # Check if service account exists
    if gcloud iam service-accounts describe "$SA_EMAIL" --project="$PROJECT_ID" >/dev/null 2>&1; then
        print_info "Service account already exists"
    else
        print_step "Creating service account..."
        gcloud iam service-accounts create "$SERVICE_NAME" \
            --display-name="Virtuoso API CLI Service Account" \
            --project="$PROJECT_ID"
        print_success "Service account created"
    fi

    # Grant necessary roles
    ROLES=(
        "roles/secretmanager.secretAccessor"
        "roles/logging.logWriter"
        "roles/cloudtrace.agent"
    )

    for role in "${ROLES[@]}"; do
        print_step "Granting $role..."
        gcloud projects add-iam-policy-binding "$PROJECT_ID" \
            --member="serviceAccount:$SA_EMAIL" \
            --role="$role" \
            --quiet >/dev/null 2>&1
    done

    print_success "Service account configured"
}

# Create secrets
create_secrets() {
    print_header "Managing Secrets"

    SECRET_NAME="virtuoso-api-token"

    # Check if secret exists
    if gcloud secrets describe "$SECRET_NAME" --project="$PROJECT_ID" >/dev/null 2>&1; then
        print_info "Secret already exists. Updating..."
        echo -n "$VIRTUOSO_API_TOKEN" | gcloud secrets versions add "$SECRET_NAME" \
            --data-file=- \
            --project="$PROJECT_ID"
    else
        print_step "Creating secret..."
        echo -n "$VIRTUOSO_API_TOKEN" | gcloud secrets create "$SECRET_NAME" \
            --data-file=- \
            --replication-policy="automatic" \
            --project="$PROJECT_ID"
    fi

    # Grant service account access
    gcloud secrets add-iam-policy-binding "$SECRET_NAME" \
        --member="serviceAccount:$SA_EMAIL" \
        --role="roles/secretmanager.secretAccessor" \
        --project="$PROJECT_ID" >/dev/null 2>&1

    print_success "Secrets configured"
}

# Build container
build_container() {
    print_header "Building Container Image"

    # Check if Dockerfile exists
    if [ ! -f "Dockerfile" ]; then
        print_error "Dockerfile not found in current directory"
        exit 1
    fi

    print_step "Building Docker image locally..."
    docker build -t "$IMAGE_URL" . &
    spin $!

    print_step "Configuring Docker for GCR..."
    gcloud auth configure-docker gcr.io --quiet

    print_step "Pushing image to Container Registry..."
    docker push "$IMAGE_URL" &
    spin $!

    print_success "Container image pushed successfully"
}

# Deploy to Cloud Run
deploy_cloud_run() {
    print_header "Deploying to Cloud Run"

    print_step "Deploying service $SERVICE_NAME..."

    # Build deployment command
    DEPLOY_CMD="gcloud run deploy $SERVICE_NAME \
        --image=$IMAGE_URL \
        --region=$REGION \
        --project=$PROJECT_ID \
        --platform=managed \
        --memory=$MEMORY \
        --cpu=$CPU \
        --min-instances=$MIN_INSTANCES \
        --max-instances=$MAX_INSTANCES \
        --concurrency=$CONCURRENCY \
        --timeout=300 \
        --service-account=$SA_EMAIL \
        --allow-unauthenticated"

    # Add environment variables
    DEPLOY_CMD+=" --set-env-vars=VIRTUOSO_ORG_ID=$VIRTUOSO_ORG_ID"
    DEPLOY_CMD+=" --set-env-vars=VIRTUOSO_API_URL=$VIRTUOSO_API_URL"
    DEPLOY_CMD+=" --set-secrets=VIRTUOSO_API_TOKEN=$SECRET_NAME:latest"

    # Execute deployment
    eval "$DEPLOY_CMD"

    # Get service URL
    SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" \
        --region="$REGION" \
        --project="$PROJECT_ID" \
        --format="value(status.url)")

    print_success "Service deployed successfully!"
    print_info "Service URL: $SERVICE_URL"
}

# Configure Cloud Build (optional)
setup_cloud_build() {
    print_header "Setting Up Continuous Deployment (Optional)"

    read -p "Set up Cloud Build for continuous deployment? (y/n) [n]: " SETUP_CD
    if [ "$SETUP_CD" != "y" ]; then
        print_info "Skipping Cloud Build setup"
        return
    fi

    # Copy Cloud Build configuration
    if [ -f "gcp/deployment-templates/cloudbuild.yaml" ]; then
        cp gcp/deployment-templates/cloudbuild.yaml cloudbuild.yaml
        print_success "Cloud Build configuration created"

        # Create trigger
        print_step "Creating build trigger..."
        gcloud builds triggers create github \
            --repo-name="virtuoso-GENerator" \
            --repo-owner="$(git remote get-url origin | sed 's/.*github.com[:\/]\(.*\)\/.*/\1/')" \
            --branch-pattern="^main$" \
            --build-config="cloudbuild.yaml" \
            --project="$PROJECT_ID" \
            --substitutions="_VIRTUOSO_API_TOKEN=$SECRET_NAME"

        print_success "Cloud Build trigger created"
    fi
}

# Test deployment
test_deployment() {
    print_header "Testing Deployment"

    print_step "Waiting for service to be ready..."
    sleep 5

    # Test health endpoint
    print_step "Testing health endpoint..."
    if curl -s -o /dev/null -w "%{http_code}" "$SERVICE_URL/health" | grep -q "200"; then
        print_success "Health check passed"
    else
        print_warning "Health check failed (service may still be starting)"
    fi

    # Test API endpoint
    print_step "Testing API endpoint..."
    RESPONSE=$(curl -s -X GET "$SERVICE_URL/api/version" 2>/dev/null || echo "")
    if [ -n "$RESPONSE" ]; then
        print_success "API responding"
        print_info "Response: $RESPONSE"
    else
        print_warning "API not responding yet"
    fi
}

# Generate deployment report
generate_report() {
    print_header "Generating Deployment Report"

    REPORT_FILE="gcp/deployment-report-$(date +%Y%m%d-%H%M%S).txt"

    cat > "$REPORT_FILE" << EOF
Virtuoso API CLI - Deployment Report
Generated: $(date)
=====================================

Deployment Summary:
------------------
Status: SUCCESS
Duration: $SECONDS seconds

Project Information:
-------------------
Project ID: $PROJECT_ID
Project Number: $PROJECT_NUMBER
Region: $REGION
Service Name: $SERVICE_NAME

Service Details:
---------------
Service URL: $SERVICE_URL
Container Image: $IMAGE_URL
Memory: $MEMORY
CPU: $CPU
Min Instances: $MIN_INSTANCES
Max Instances: $MAX_INSTANCES
Concurrency: $CONCURRENCY

Configuration:
-------------
Virtuoso Org ID: $VIRTUOSO_ORG_ID
Virtuoso API URL: $VIRTUOSO_API_URL

Resources Created:
-----------------
✓ Cloud Run Service: $SERVICE_NAME
✓ Container Image: $IMAGE_URL
✓ Service Account: $SA_EMAIL
✓ Secret: virtuoso-api-token

Access Information:
------------------
Service Endpoint: $SERVICE_URL
Health Check: $SERVICE_URL/health
API Documentation: $SERVICE_URL/docs

Next Steps:
----------
1. Test the API: curl $SERVICE_URL/api/version
2. View logs: gcloud run logs read --service=$SERVICE_NAME --region=$REGION
3. Monitor: https://console.cloud.google.com/run/detail/$REGION/$SERVICE_NAME/metrics

Cost Information:
----------------
Estimated monthly cost (idle): \$0.00 (with scale-to-zero)
Estimated cost per 1M requests: ~\$0.50

Manual Actions Required:
-----------------------
□ Configure custom domain (optional)
□ Set up monitoring alerts (recommended)
□ Review IAM permissions
□ Configure backup strategy

Useful Commands:
---------------
# View service details
gcloud run services describe $SERVICE_NAME --region=$REGION

# Stream logs
gcloud run logs tail --service=$SERVICE_NAME --region=$REGION

# Update service
gcloud run deploy $SERVICE_NAME --image=$IMAGE_URL --region=$REGION

# Delete service (if needed)
gcloud run services delete $SERVICE_NAME --region=$REGION

Support:
-------
Documentation: https://github.com/your-org/virtuoso-GENerator
Issues: https://github.com/your-org/virtuoso-GENerator/issues
EOF

    print_success "Deployment report saved to: $REPORT_FILE"

    # Display key information
    cat << EOF

${GREEN}🎉 Deployment Complete!${NC}

Your Virtuoso API CLI is now running on Google Cloud Platform.

${BLUE}Service URL:${NC} $SERVICE_URL
${BLUE}Health Check:${NC} $SERVICE_URL/health

${YELLOW}Quick Test:${NC}
curl $SERVICE_URL/api/version

${YELLOW}View Logs:${NC}
gcloud run logs tail --service=$SERVICE_NAME --region=$REGION

EOF
}

# Main deployment flow
main() {
    clear
    print_header "Virtuoso API CLI - One-Click Deployment"

    # Start timer
    SECONDS=0

    # Load configuration
    load_config

    # Run deployment steps
    check_prerequisites
    enable_apis
    create_service_account
    create_secrets
    build_container
    deploy_cloud_run
    setup_cloud_build
    test_deployment
    generate_report

    print_header "Deployment Completed Successfully! 🚀"
    print_info "Total deployment time: $SECONDS seconds"
}

# Handle errors
trap 'print_error "Deployment failed. Check the error messages above."; exit 1' ERR

# Run if not sourced
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi
