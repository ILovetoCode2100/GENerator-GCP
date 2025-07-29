#!/bin/bash
# Initial GCP Project Setup Script for Virtuoso API CLI
# This script handles the initial setup of a GCP project

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default values
PROJECT_ID="${GCP_PROJECT_ID:-}"
PROJECT_NAME="${GCP_PROJECT_NAME:-Virtuoso API CLI}"
BILLING_ACCOUNT_ID="${GCP_BILLING_ACCOUNT_ID:-}"
ORGANIZATION_ID="${GCP_ORGANIZATION_ID:-}"
FOLDER_ID="${GCP_FOLDER_ID:-}"
REGION="${GCP_REGION:-us-central1}"
CREATE_PROJECT=false

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

# Function to generate random project ID
generate_project_id() {
    local prefix="virtuoso-api"
    local suffix=$(date +%Y%m%d%H%M%S)
    echo "${prefix}-${suffix}"
}

# Function to check if project exists
project_exists() {
    gcloud projects describe "$1" &>/dev/null
}

# Function to create project
create_project() {
    print_info "Creating new GCP project..."

    # Generate project ID if not provided
    if [ -z "$PROJECT_ID" ]; then
        PROJECT_ID=$(generate_project_id)
        print_info "Generated project ID: $PROJECT_ID"
    fi

    # Check if project already exists
    if project_exists "$PROJECT_ID"; then
        print_warning "Project $PROJECT_ID already exists"
        return
    fi

    # Create project command
    local create_cmd="gcloud projects create $PROJECT_ID --name=\"$PROJECT_NAME\""

    # Add organization or folder if specified
    if [ -n "$ORGANIZATION_ID" ]; then
        create_cmd="$create_cmd --organization=$ORGANIZATION_ID"
    elif [ -n "$FOLDER_ID" ]; then
        create_cmd="$create_cmd --folder=$FOLDER_ID"
    fi

    # Create the project
    eval "$create_cmd"

    print_success "Project $PROJECT_ID created"
}

# Function to link billing account
link_billing() {
    print_info "Linking billing account..."

    if [ -z "$BILLING_ACCOUNT_ID" ]; then
        print_info "Fetching available billing accounts..."
        gcloud billing accounts list

        read -p "Enter billing account ID: " BILLING_ACCOUNT_ID

        if [ -z "$BILLING_ACCOUNT_ID" ]; then
            print_error "Billing account ID is required"
            exit 1
        fi
    fi

    gcloud billing projects link "$PROJECT_ID" \
        --billing-account="$BILLING_ACCOUNT_ID"

    print_success "Billing account linked"
}

# Function to enable APIs
enable_apis() {
    print_info "Enabling required APIs..."

    local apis=(
        # Core services
        "compute.googleapis.com"
        "storage.googleapis.com"
        "cloudresourcemanager.googleapis.com"

        # Cloud Run and Build
        "cloudrun.googleapis.com"
        "cloudbuild.googleapis.com"
        "containerregistry.googleapis.com"
        "artifactregistry.googleapis.com"

        # Cloud Functions
        "cloudfunctions.googleapis.com"
        "cloudbuildv2.googleapis.com"

        # Async processing
        "cloudscheduler.googleapis.com"
        "cloudtasks.googleapis.com"
        "pubsub.googleapis.com"

        # Data storage
        "firestore.googleapis.com"
        "redis.googleapis.com"

        # Security and secrets
        "secretmanager.googleapis.com"
        "iamcredentials.googleapis.com"

        # Monitoring and logging
        "monitoring.googleapis.com"
        "logging.googleapis.com"
        "cloudtrace.googleapis.com"
        "clouderrorreporting.googleapis.com"

        # Networking
        "servicenetworking.googleapis.com"
        "vpcaccess.googleapis.com"
    )

    # Enable APIs in batches to avoid rate limiting
    local batch_size=5
    local api_count=${#apis[@]}

    for ((i=0; i<api_count; i+=batch_size)); do
        local batch=("${apis[@]:i:batch_size}")
        print_info "Enabling APIs batch $((i/batch_size + 1))..."

        for api in "${batch[@]}"; do
            gcloud services enable "$api" --project="$PROJECT_ID" &
        done

        # Wait for batch to complete
        wait

        # Small delay between batches
        sleep 2
    done

    print_success "All APIs enabled"
}

# Function to create service accounts
create_service_accounts() {
    print_info "Creating service accounts..."

    # Cloud Run service account
    gcloud iam service-accounts create virtuoso-api-cli \
        --display-name="Virtuoso API CLI Service" \
        --description="Service account for Cloud Run API service" \
        --project="$PROJECT_ID" || true

    # Cloud Functions service account
    gcloud iam service-accounts create virtuoso-functions \
        --display-name="Virtuoso Cloud Functions" \
        --description="Service account for Cloud Functions" \
        --project="$PROJECT_ID" || true

    # Cloud Build service account (already exists, just configure)
    local build_sa="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

    # Cloud Scheduler service account
    gcloud iam service-accounts create virtuoso-scheduler \
        --display-name="Virtuoso Cloud Scheduler" \
        --description="Service account for Cloud Scheduler jobs" \
        --project="$PROJECT_ID" || true

    print_success "Service accounts created"
}

# Function to set up IAM roles
setup_iam_roles() {
    print_info "Setting up IAM roles..."

    local project_number=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")

    # Cloud Run service account roles
    local cloudrun_sa="virtuoso-api-cli@${PROJECT_ID}.iam.gserviceaccount.com"
    local roles_cloudrun=(
        "roles/firestore.dataEditor"
        "roles/secretmanager.secretAccessor"
        "roles/cloudtasks.enqueuer"
        "roles/pubsub.publisher"
        "roles/storage.objectViewer"
        "roles/logging.logWriter"
        "roles/cloudtrace.agent"
    )

    for role in "${roles_cloudrun[@]}"; do
        gcloud projects add-iam-policy-binding "$PROJECT_ID" \
            --member="serviceAccount:$cloudrun_sa" \
            --role="$role" \
            --quiet || true
    done

    # Cloud Functions service account roles
    local functions_sa="virtuoso-functions@${PROJECT_ID}.iam.gserviceaccount.com"
    local roles_functions=(
        "roles/firestore.dataEditor"
        "roles/secretmanager.secretAccessor"
        "roles/pubsub.subscriber"
        "roles/storage.objectAdmin"
        "roles/logging.logWriter"
        "roles/monitoring.metricWriter"
    )

    for role in "${roles_functions[@]}"; do
        gcloud projects add-iam-policy-binding "$PROJECT_ID" \
            --member="serviceAccount:$functions_sa" \
            --role="$role" \
            --quiet || true
    done

    # Cloud Build service account roles
    local build_sa="${project_number}@cloudbuild.gserviceaccount.com"
    local roles_build=(
        "roles/cloudfunctions.developer"
        "roles/run.admin"
        "roles/iam.serviceAccountUser"
        "roles/secretmanager.secretAccessor"
    )

    for role in "${roles_build[@]}"; do
        gcloud projects add-iam-policy-binding "$PROJECT_ID" \
            --member="serviceAccount:$build_sa" \
            --role="$role" \
            --quiet || true
    done

    # Cloud Scheduler service account roles
    local scheduler_sa="virtuoso-scheduler@${PROJECT_ID}.iam.gserviceaccount.com"
    local roles_scheduler=(
        "roles/cloudtasks.enqueuer"
        "roles/cloudfunctions.invoker"
        "roles/run.invoker"
    )

    for role in "${roles_scheduler[@]}"; do
        gcloud projects add-iam-policy-binding "$PROJECT_ID" \
            --member="serviceAccount:$scheduler_sa" \
            --role="$role" \
            --quiet || true
    done

    print_success "IAM roles configured"
}

# Function to configure project defaults
configure_defaults() {
    print_info "Configuring project defaults..."

    # Set default project
    gcloud config set project "$PROJECT_ID"

    # Set default region
    gcloud config set compute/region "$REGION"
    gcloud config set run/region "$REGION"
    gcloud config set functions/region "$REGION"

    # Enable required features
    gcloud config set builds/use_kaniko True
    gcloud config set builds/timeout 1200

    print_success "Project defaults configured"
}

# Function to create storage buckets
create_storage_buckets() {
    print_info "Creating storage buckets..."

    # Terraform state bucket
    gsutil mb -p "$PROJECT_ID" -c STANDARD -l "$REGION" \
        "gs://${PROJECT_ID}-terraform-state" || true

    # Enable versioning for Terraform state
    gsutil versioning set on "gs://${PROJECT_ID}-terraform-state"

    # Cloud Build artifacts bucket
    gsutil mb -p "$PROJECT_ID" -c STANDARD -l "$REGION" \
        "gs://${PROJECT_ID}-build-artifacts" || true

    # Test results bucket
    gsutil mb -p "$PROJECT_ID" -c STANDARD -l "$REGION" \
        "gs://${PROJECT_ID}-test-results" || true

    # Set lifecycle rules for artifacts
    cat > lifecycle.json <<EOF
{
  "lifecycle": {
    "rule": [
      {
        "action": {"type": "Delete"},
        "condition": {
          "age": 30,
          "matchesPrefix": ["builds/", "artifacts/"]
        }
      }
    ]
  }
}
EOF

    gsutil lifecycle set lifecycle.json "gs://${PROJECT_ID}-build-artifacts"
    rm lifecycle.json

    print_success "Storage buckets created"
}

# Function to create Firestore database
create_firestore() {
    print_info "Creating Firestore database..."

    # Create Firestore database in native mode
    gcloud firestore databases create \
        --location="$REGION" \
        --project="$PROJECT_ID" || true

    print_success "Firestore database created"
}

# Function to setup initial secrets
setup_secrets() {
    print_info "Setting up initial secrets..."

    # Create secrets (values will be added later)
    echo -n "placeholder" | gcloud secrets create virtuoso-api-key \
        --data-file=- \
        --replication-policy="automatic" \
        --project="$PROJECT_ID" || true

    echo -n "placeholder" | gcloud secrets create virtuoso-org-id \
        --data-file=- \
        --replication-policy="automatic" \
        --project="$PROJECT_ID" || true

    print_warning "Remember to update secrets with actual values using:"
    print_info "  ./secrets-setup.sh"

    print_success "Initial secrets created"
}

# Function to create VPC network
create_vpc_network() {
    print_info "Creating VPC network..."

    # Create custom VPC
    gcloud compute networks create virtuoso-vpc \
        --subnet-mode=custom \
        --bgp-routing-mode=regional \
        --project="$PROJECT_ID" || true

    # Create subnet
    gcloud compute networks subnets create virtuoso-subnet \
        --network=virtuoso-vpc \
        --region="$REGION" \
        --range=10.0.0.0/24 \
        --project="$PROJECT_ID" || true

    # Create Cloud NAT for outbound connectivity
    gcloud compute routers create virtuoso-router \
        --network=virtuoso-vpc \
        --region="$REGION" \
        --project="$PROJECT_ID" || true

    gcloud compute routers nats create virtuoso-nat \
        --router=virtuoso-router \
        --region="$REGION" \
        --nat-all-subnet-ip-ranges \
        --auto-allocate-nat-external-ips \
        --project="$PROJECT_ID" || true

    print_success "VPC network created"
}

# Function to print setup summary
print_summary() {
    print_info "Project Setup Summary"
    echo -e "${BLUE}===================================================${NC}"
    echo -e "Project ID: ${GREEN}$PROJECT_ID${NC}"
    echo -e "Project Name: ${GREEN}$PROJECT_NAME${NC}"
    echo -e "Region: ${GREEN}$REGION${NC}"
    echo -e "\nResources created:"
    echo -e "- Service accounts"
    echo -e "- Storage buckets"
    echo -e "- Firestore database"
    echo -e "- VPC network"
    echo -e "- Secret placeholders"
    echo -e "\nNext steps:"
    echo -e "1. Run ${GREEN}./secrets-setup.sh${NC} to configure secrets"
    echo -e "2. Run ${GREEN}./deploy.sh${NC} to deploy the application"
    echo -e "${BLUE}===================================================${NC}"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        --project-name)
            PROJECT_NAME="$2"
            shift 2
            ;;
        --billing-account)
            BILLING_ACCOUNT_ID="$2"
            shift 2
            ;;
        --organization)
            ORGANIZATION_ID="$2"
            shift 2
            ;;
        --folder)
            FOLDER_ID="$2"
            shift 2
            ;;
        --region)
            REGION="$2"
            shift 2
            ;;
        --create-project)
            CREATE_PROJECT=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --project-id ID          GCP project ID"
            echo "  --project-name NAME      Project display name"
            echo "  --billing-account ID     Billing account ID"
            echo "  --organization ID        Organization ID (optional)"
            echo "  --folder ID             Folder ID (optional)"
            echo "  --region REGION         GCP region (default: us-central1)"
            echo "  --create-project        Create new project"
            echo "  --help                  Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Main setup flow
print_info "Starting GCP project setup for Virtuoso API CLI"
echo -e "${BLUE}===================================================${NC}\n"

# Create project if requested
if [ "$CREATE_PROJECT" = true ]; then
    create_project
elif [ -z "$PROJECT_ID" ]; then
    # Try to get current project
    PROJECT_ID=$(gcloud config get-value project 2>/dev/null || echo "")
    if [ -z "$PROJECT_ID" ]; then
        print_error "No project ID specified. Use --project-id or --create-project"
        exit 1
    fi
fi

# Verify project exists
if ! project_exists "$PROJECT_ID"; then
    print_error "Project $PROJECT_ID does not exist"
    print_info "Use --create-project to create a new project"
    exit 1
fi

print_info "Using project: $PROJECT_ID"

# Set project as active
gcloud config set project "$PROJECT_ID"

# Link billing (required for most services)
link_billing

# Execute setup steps
enable_apis
create_service_accounts
setup_iam_roles
configure_defaults
create_storage_buckets
create_firestore
setup_secrets
create_vpc_network

# Print summary
print_summary

print_success "Project setup completed successfully!"
