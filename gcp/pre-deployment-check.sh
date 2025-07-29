#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_header() {
    echo -e "\n${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Load configuration if exists
load_config() {
    if [ -f "gcp/auto-deploy-config.yaml" ]; then
        # Parse YAML (basic parsing)
        PROJECT_ID=$(grep "id:" gcp/auto-deploy-config.yaml | head -1 | awk '{print $2}')
        REGION=$(grep "region:" gcp/auto-deploy-config.yaml | awk '{print $2}')
        SERVICE_NAME=$(grep "name:" gcp/auto-deploy-config.yaml | head -1 | awk '{print $2}')
        MEMORY=$(grep "memory:" gcp/auto-deploy-config.yaml | awk '{print $2}')
        CPU=$(grep "cpu:" gcp/auto-deploy-config.yaml | awk '{print $2}' | tr -d '"')
        MIN_INSTANCES=$(grep "min_instances:" gcp/auto-deploy-config.yaml | awk '{print $2}')
        MAX_INSTANCES=$(grep "max_instances:" gcp/auto-deploy-config.yaml | awk '{print $2}')
    fi
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    local all_good=true

    # Check required tools
    print_info "Checking required tools..."
    for tool in gcloud docker terraform; do
        if command_exists "$tool"; then
            VERSION=$($tool --version 2>&1 | head -1)
            print_success "$tool: $VERSION"
        else
            print_error "$tool: NOT INSTALLED"
            all_good=false
        fi
    done

    # Check optional tools
    print_info "\nChecking optional tools..."
    for tool in jq yq curl; do
        if command_exists "$tool"; then
            print_success "$tool: installed"
        else
            print_warning "$tool: not installed (optional)"
        fi
    done

    if [ "$all_good" = false ]; then
        print_error "\nSome required tools are missing. Please install them first."
        return 1
    fi

    return 0
}

# Check GCP authentication
check_gcp_auth() {
    print_header "Checking GCP Authentication"

    # Check if authenticated
    if gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        ACCOUNT=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")
        print_success "Authenticated as: $ACCOUNT"

        # Check application default credentials
        if [ -f "$HOME/.config/gcloud/application_default_credentials.json" ]; then
            print_success "Application default credentials configured"
        else
            print_warning "Application default credentials not set"
            print_info "Run: gcloud auth application-default login"
        fi
    else
        print_error "Not authenticated with GCP"
        print_info "Run: gcloud auth login"
        return 1
    fi

    # Check current project
    CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null || echo "")
    if [ -n "$CURRENT_PROJECT" ]; then
        print_success "Current project: $CURRENT_PROJECT"
    else
        print_warning "No default project set"
    fi

    return 0
}

# Check project permissions
check_permissions() {
    print_header "Checking Project Permissions"

    if [ -z "$PROJECT_ID" ]; then
        print_warning "No project configured. Run deploy-wizard.sh first."
        return 1
    fi

    print_info "Checking permissions for project: $PROJECT_ID"

    # Required roles
    REQUIRED_ROLES=(
        "roles/cloudbuild.builds.editor"
        "roles/run.admin"
        "roles/storage.admin"
        "roles/secretmanager.admin"
        "roles/iam.serviceAccountUser"
    )

    # Get current user's roles
    USER_EMAIL=$(gcloud config get-value account)
    CURRENT_ROLES=$(gcloud projects get-iam-policy "$PROJECT_ID" \
        --flatten="bindings[].members" \
        --filter="bindings.members:user:$USER_EMAIL" \
        --format="value(bindings.role)" 2>/dev/null || echo "")

    # Check each required role
    local missing_roles=""
    for role in "${REQUIRED_ROLES[@]}"; do
        if echo "$CURRENT_ROLES" | grep -q "$role"; then
            print_success "$role"
        else
            # Check if user has owner or editor role (which includes these permissions)
            if echo "$CURRENT_ROLES" | grep -q -E "(roles/owner|roles/editor)"; then
                print_success "$role (via owner/editor)"
            else
                print_warning "$role: MISSING"
                missing_roles+="$role "
            fi
        fi
    done

    if [ -n "$missing_roles" ]; then
        print_warning "\nSome permissions may be missing. Deployment might require additional permissions."
    fi

    return 0
}

# Check APIs
check_apis() {
    print_header "Checking Required APIs"

    if [ -z "$PROJECT_ID" ]; then
        print_warning "No project configured. Skipping API check."
        return 0
    fi

    REQUIRED_APIS=(
        "cloudbuild.googleapis.com"
        "run.googleapis.com"
        "containerregistry.googleapis.com"
        "secretmanager.googleapis.com"
        "cloudresourcemanager.googleapis.com"
    )

    # Get enabled APIs
    ENABLED_APIS=$(gcloud services list --enabled --project="$PROJECT_ID" --format="value(config.name)" 2>/dev/null || echo "")

    local all_enabled=true
    for api in "${REQUIRED_APIS[@]}"; do
        if echo "$ENABLED_APIS" | grep -q "$api"; then
            print_success "$api: ENABLED"
        else
            print_warning "$api: NOT ENABLED"
            all_enabled=false
        fi
    done

    if [ "$all_enabled" = false ]; then
        print_info "\nSome APIs are not enabled. They will be enabled during deployment."
    fi

    return 0
}

# Check quota and limits
check_quota() {
    print_header "Checking Quotas and Limits"

    if [ -z "$PROJECT_ID" ]; then
        print_warning "No project configured. Skipping quota check."
        return 0
    fi

    print_info "Checking Cloud Run quotas for region: ${REGION:-us-central1}"

    # Check if we can access quota information
    if gcloud compute regions describe "${REGION:-us-central1}" --project="$PROJECT_ID" >/dev/null 2>&1; then
        print_success "Region ${REGION:-us-central1} is available"

        # Note: Detailed quota checking requires specific API calls
        print_info "Cloud Run service limits:"
        print_info "  - Services per region: 1000"
        print_info "  - Concurrent requests: 250 per container"
        print_info "  - Max containers: 1000"
        print_info "  - Request timeout: 60 minutes"
    else
        print_warning "Unable to check detailed quotas"
    fi

    return 0
}

# Estimate costs
estimate_costs() {
    print_header "Cost Estimation"

    # Load configuration
    load_config

    # Default values if not configured
    MEMORY=${MEMORY:-"512Mi"}
    CPU=${CPU:-"1"}
    MIN_INSTANCES=${MIN_INSTANCES:-0}
    MAX_INSTANCES=${MAX_INSTANCES:-100}

    # Parse memory to GB
    MEMORY_GB=$(echo "$MEMORY" | sed 's/Mi//' | awk '{print $1/1024}')
    if [[ "$MEMORY" == *"Gi"* ]]; then
        MEMORY_GB=$(echo "$MEMORY" | sed 's/Gi//')
    fi

    # Calculate costs
    print_info "Configuration:"
    print_info "  - Memory: $MEMORY"
    print_info "  - CPU: $CPU"
    print_info "  - Min instances: $MIN_INSTANCES"
    print_info "  - Max instances: $MAX_INSTANCES"

    # Base costs (USD)
    CPU_COST_PER_SECOND=0.00002400
    MEMORY_COST_PER_GB_SECOND=0.00000250
    REQUEST_COST_PER_MILLION=0.40

    # Calculate idle cost (min instances always running)
    if [ "$MIN_INSTANCES" -gt 0 ]; then
        SECONDS_PER_MONTH=2628000  # 30.4 days
        IDLE_CPU_COST=$(echo "$MIN_INSTANCES * $CPU * $CPU_COST_PER_SECOND * $SECONDS_PER_MONTH" | bc -l)
        IDLE_MEMORY_COST=$(echo "$MIN_INSTANCES * $MEMORY_GB * $MEMORY_COST_PER_GB_SECOND * $SECONDS_PER_MONTH" | bc -l)
        IDLE_TOTAL=$(echo "$IDLE_CPU_COST + $IDLE_MEMORY_COST" | bc -l)

        print_info "\nIdle cost (always-on instances):"
        printf "  - CPU: \$%.2f/month\n" "$IDLE_CPU_COST"
        printf "  - Memory: \$%.2f/month\n" "$IDLE_MEMORY_COST"
        printf "  - Total idle: \$%.2f/month\n" "$IDLE_TOTAL"
    else
        print_success "\nIdle cost: \$0.00/month (scale-to-zero enabled)"
    fi

    # Calculate active costs
    print_info "\nActive usage costs:"
    print_info "  - Per 1M requests: ~\$0.50"
    print_info "  - Per 1K requests: ~\$0.0005"

    # Additional costs
    print_info "\nAdditional costs:"
    print_info "  - Container Registry storage: ~\$0.10/month"
    print_info "  - Cloud Build: ~\$0.10/deployment"
    print_info "  - Network egress: \$0.12/GB (after 1GB free)"

    # Free tier
    print_success "\nFree tier includes:"
    print_success "  - 2 million requests/month"
    print_success "  - 360,000 vCPU-seconds/month"
    print_success "  - 180,000 GiB-seconds/month"
    print_success "  - 1 GB network egress/month"

    return 0
}

# Check existing resources
check_existing_resources() {
    print_header "Checking Existing Resources"

    if [ -z "$PROJECT_ID" ]; then
        print_warning "No project configured. Skipping resource check."
        return 0
    fi

    load_config

    # Check if service already exists
    if [ -n "$SERVICE_NAME" ] && [ -n "$REGION" ]; then
        if gcloud run services describe "$SERVICE_NAME" --region="$REGION" --project="$PROJECT_ID" >/dev/null 2>&1; then
            print_warning "Cloud Run service '$SERVICE_NAME' already exists in region $REGION"
            SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" --region="$REGION" --project="$PROJECT_ID" --format="value(status.url)")
            print_info "Service URL: $SERVICE_URL"
        else
            print_success "Cloud Run service '$SERVICE_NAME' does not exist (will be created)"
        fi
    fi

    # Check container images
    if [ -n "$SERVICE_NAME" ]; then
        IMAGES=$(gcloud container images list --filter="name~$SERVICE_NAME" --project="$PROJECT_ID" 2>/dev/null || echo "")
        if [ -n "$IMAGES" ]; then
            print_info "Existing container images found:"
            echo "$IMAGES"
        else
            print_info "No existing container images found"
        fi
    fi

    # Check secrets
    SECRETS=$(gcloud secrets list --project="$PROJECT_ID" --filter="name~virtuoso" --format="value(name)" 2>/dev/null || echo "")
    if [ -n "$SECRETS" ]; then
        print_info "Existing secrets found:"
        echo "$SECRETS"
    else
        print_info "No existing secrets found"
    fi

    return 0
}

# Generate deployment preview
generate_preview() {
    print_header "Deployment Preview"

    load_config

    print_info "The following resources will be created/updated:"
    echo ""
    echo "1. Container Image:"
    echo "   - Registry: gcr.io/$PROJECT_ID/$SERVICE_NAME"
    echo "   - Build from: ./Dockerfile"
    echo ""
    echo "2. Cloud Run Service:"
    echo "   - Name: $SERVICE_NAME"
    echo "   - Region: $REGION"
    echo "   - Memory: $MEMORY"
    echo "   - CPU: $CPU"
    echo "   - Scaling: $MIN_INSTANCES to $MAX_INSTANCES instances"
    echo ""
    echo "3. Secrets:"
    echo "   - virtuoso-api-token"
    echo ""
    echo "4. IAM:"
    echo "   - Service account: $SERVICE_NAME@$PROJECT_ID.iam.gserviceaccount.com"
    echo "   - Roles: Cloud Run Invoker, Secret Accessor"
    echo ""
    echo "5. Networking:"
    echo "   - Public endpoint (HTTPS)"
    echo "   - Automatic TLS certificate"
    echo ""

    return 0
}

# Main check function
run_checks() {
    clear
    print_header "Virtuoso API CLI - Pre-deployment Check"

    local all_checks_passed=true

    # Run all checks
    check_prerequisites || all_checks_passed=false
    check_gcp_auth || all_checks_passed=false
    check_permissions
    check_apis
    check_quota
    check_existing_resources
    estimate_costs
    generate_preview

    # Summary
    print_header "Pre-deployment Check Summary"

    if [ "$all_checks_passed" = true ]; then
        print_success "All critical checks passed!"
        print_info "You are ready to deploy."
        print_info "\nNext steps:"
        print_info "1. Review the deployment preview above"
        print_info "2. Run: ./gcp/deploy-wizard.sh (if not already done)"
        print_info "3. Run: ./gcp/one-click-deploy.sh"
    else
        print_error "Some critical checks failed."
        print_info "Please resolve the issues above before proceeding."
    fi

    # Save check results
    cat > gcp/.precheck-results << EOF
CHECK_TIME=$(date +%s)
CHECKS_PASSED=$all_checks_passed
PROJECT_ID=$PROJECT_ID
REGION=$REGION
SERVICE_NAME=$SERVICE_NAME
EOF
}

# Run checks
run_checks
