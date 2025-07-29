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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to get user input with default
get_input() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"

    if [ -n "$default" ]; then
        read -p "$prompt [$default]: " value
        value="${value:-$default}"
    else
        read -p "$prompt: " value
    fi

    eval "$var_name='$value'"
}

# Function to generate random suffix
generate_suffix() {
    echo "-$(date +%s | tail -c 5)"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    local missing_tools=""

    # Check for required tools
    for tool in gcloud docker terraform; do
        if command_exists "$tool"; then
            print_success "$tool is installed"
        else
            print_error "$tool is not installed"
            missing_tools+="$tool "
        fi
    done

    if [ -n "$missing_tools" ]; then
        print_error "Missing required tools: $missing_tools"
        print_info "Please install missing tools and run again."
        exit 1
    fi

    # Check gcloud authentication
    if gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        ACCOUNT=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")
        print_success "Authenticated as: $ACCOUNT"
    else
        print_error "Not authenticated with gcloud"
        print_info "Please run: gcloud auth login"
        exit 1
    fi
}

# Auto-detect GCP configuration
detect_gcp_config() {
    print_header "Detecting GCP Configuration"

    # Try to get current project
    CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null || echo "")
    if [ -n "$CURRENT_PROJECT" ]; then
        print_info "Current GCP project: $CURRENT_PROJECT"
        get_input "Use this project? (y/n)" "y" USE_CURRENT
        if [ "$USE_CURRENT" = "y" ]; then
            PROJECT_ID="$CURRENT_PROJECT"
        fi
    fi

    # If no project selected, list available projects
    if [ -z "$PROJECT_ID" ]; then
        print_info "Available projects:"
        gcloud projects list --format="table(projectId,name)" 2>/dev/null || true
        get_input "Enter project ID (or 'new' to create)" "" PROJECT_ID

        if [ "$PROJECT_ID" = "new" ]; then
            get_input "Enter new project ID" "virtuoso-cli$(generate_suffix)" PROJECT_ID
            get_input "Enter billing account ID" "" BILLING_ACCOUNT

            print_info "Creating new project: $PROJECT_ID"
            gcloud projects create "$PROJECT_ID" --name="Virtuoso API CLI"
            gcloud config set project "$PROJECT_ID"

            if [ -n "$BILLING_ACCOUNT" ]; then
                gcloud beta billing projects link "$PROJECT_ID" --billing-account="$BILLING_ACCOUNT"
            fi
        else
            gcloud config set project "$PROJECT_ID"
        fi
    fi

    # Get project number
    PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")

    # Detect region
    REGIONS=("us-central1" "us-east1" "europe-west1" "asia-southeast1")
    print_info "Available regions:"
    for i in "${!REGIONS[@]}"; do
        echo "  $((i+1)). ${REGIONS[$i]}"
    done
    get_input "Select region (1-4)" "1" REGION_CHOICE
    REGION="${REGIONS[$((REGION_CHOICE-1))]}"
}

# Configure deployment settings
configure_deployment() {
    print_header "Deployment Configuration"

    # Service name
    get_input "Service name" "virtuoso-api-cli" SERVICE_NAME

    # Virtuoso API configuration
    print_info "\nVirtuoso API Configuration:"
    get_input "Virtuoso API token" "" VIRTUOSO_API_TOKEN
    get_input "Virtuoso organization ID" "2242" VIRTUOSO_ORG_ID
    get_input "Virtuoso API URL" "https://api-app2.virtuoso.qa/api" VIRTUOSO_API_URL

    # Cloud Run configuration
    print_info "\nCloud Run Configuration:"
    get_input "Memory allocation" "512Mi" MEMORY
    get_input "CPU allocation" "1" CPU
    get_input "Min instances" "0" MIN_INSTANCES
    get_input "Max instances" "100" MAX_INSTANCES
    get_input "Concurrency" "80" CONCURRENCY

    # Cost optimization
    print_info "\nCost Optimization:"
    get_input "Enable CPU boost?" "n" CPU_BOOST
    get_input "Enable always-on (min 1 instance)?" "n" ALWAYS_ON
    if [ "$ALWAYS_ON" = "y" ]; then
        MIN_INSTANCES="1"
    fi
}

# Generate configuration files
generate_configs() {
    print_header "Generating Configuration Files"

    # Create auto-deploy-config.yaml
    cat > gcp/auto-deploy-config.yaml << EOF
# Auto-generated deployment configuration
# Generated on: $(date)

project:
  id: $PROJECT_ID
  number: $PROJECT_NUMBER
  region: $REGION

service:
  name: $SERVICE_NAME
  image: gcr.io/$PROJECT_ID/$SERVICE_NAME

virtuoso:
  api_token: $VIRTUOSO_API_TOKEN
  org_id: $VIRTUOSO_ORG_ID
  api_url: $VIRTUOSO_API_URL

cloud_run:
  memory: $MEMORY
  cpu: "$CPU"
  min_instances: $MIN_INSTANCES
  max_instances: $MAX_INSTANCES
  concurrency: $CONCURRENCY
  cpu_boost: $CPU_BOOST

deployment:
  timestamp: $(date +%s)
  deployer: $ACCOUNT
EOF
    print_success "Generated auto-deploy-config.yaml"

    # Create Terraform tfvars
    cat > gcp/deployment-templates/terraform.tfvars << EOF
# Auto-generated Terraform variables
# Generated on: $(date)

project_id = "$PROJECT_ID"
project_number = "$PROJECT_NUMBER"
region = "$REGION"
service_name = "$SERVICE_NAME"

# Cloud Run configuration
cloud_run_config = {
  memory = "$MEMORY"
  cpu = "$CPU"
  min_instances = $MIN_INSTANCES
  max_instances = $MAX_INSTANCES
  concurrency = $CONCURRENCY
  cpu_boost = $([ "$CPU_BOOST" = "y" ] && echo "true" || echo "false")
}

# Virtuoso configuration
virtuoso_config = {
  org_id = "$VIRTUOSO_ORG_ID"
  api_url = "$VIRTUOSO_API_URL"
}
EOF
    print_success "Generated terraform.tfvars"

    # Create Cloud Build configuration
    cat > gcp/deployment-templates/cloudbuild.yaml << EOF
# Auto-generated Cloud Build configuration
# Generated on: $(date)

steps:
  # Build the container image
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/$SERVICE_NAME', '-f', 'Dockerfile', '.']
    dir: '.'

  # Push the container image to Container Registry
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/$SERVICE_NAME']

  # Deploy to Cloud Run
  - name: 'gcr.io/cloud-builders/gcloud'
    args:
      - 'run'
      - 'deploy'
      - '$SERVICE_NAME'
      - '--image=gcr.io/$PROJECT_ID/$SERVICE_NAME'
      - '--region=$REGION'
      - '--platform=managed'
      - '--memory=$MEMORY'
      - '--cpu=$CPU'
      - '--min-instances=$MIN_INSTANCES'
      - '--max-instances=$MAX_INSTANCES'
      - '--concurrency=$CONCURRENCY'
      - '--allow-unauthenticated'
      - '--set-env-vars=VIRTUOSO_API_TOKEN=\${_VIRTUOSO_API_TOKEN}'
      - '--set-env-vars=VIRTUOSO_ORG_ID=$VIRTUOSO_ORG_ID'
      - '--set-env-vars=VIRTUOSO_API_URL=$VIRTUOSO_API_URL'

images:
  - 'gcr.io/$PROJECT_ID/$SERVICE_NAME'

options:
  logging: CLOUD_LOGGING_ONLY
  machineType: 'E2_HIGHCPU_8'

timeout: '1200s'
EOF
    print_success "Generated cloudbuild.yaml"

    # Create secret template
    cat > gcp/deployment-templates/secrets.json << EOF
{
  "virtuoso_api_token": "$VIRTUOSO_API_TOKEN",
  "project_id": "$PROJECT_ID",
  "service_name": "$SERVICE_NAME"
}
EOF
    print_success "Generated secrets.json"

    # Create example API configuration
    cat > gcp/deployment-templates/api-config-example.yaml << EOF
# Example API configuration for local development
# Copy to ~/.api-cli/virtuoso-config.yaml

api:
  auth_token: $VIRTUOSO_API_TOKEN
  base_url: $VIRTUOSO_API_URL
organization:
  id: "$VIRTUOSO_ORG_ID"
EOF
    print_success "Generated api-config-example.yaml"
}

# Enable required APIs
enable_apis() {
    print_header "Enabling Required APIs"

    REQUIRED_APIS=(
        "cloudbuild.googleapis.com"
        "run.googleapis.com"
        "containerregistry.googleapis.com"
        "secretmanager.googleapis.com"
        "cloudresourcemanager.googleapis.com"
    )

    for api in "${REQUIRED_APIS[@]}"; do
        print_info "Enabling $api..."
        gcloud services enable "$api" --project="$PROJECT_ID" || true
    done

    print_success "All required APIs enabled"
}

# Create deployment summary
create_summary() {
    print_header "Deployment Summary"

    cat > gcp/deployment-summary.txt << EOF
Virtuoso API CLI - GCP Deployment Summary
Generated: $(date)
=========================================

Project Configuration:
- Project ID: $PROJECT_ID
- Project Number: $PROJECT_NUMBER
- Region: $REGION
- Service Name: $SERVICE_NAME

Cloud Run Configuration:
- Memory: $MEMORY
- CPU: $CPU
- Min Instances: $MIN_INSTANCES
- Max Instances: $MAX_INSTANCES
- Concurrency: $CONCURRENCY
- CPU Boost: $CPU_BOOST

Estimated Costs (per month):
- Idle Cost: \$$([ "$MIN_INSTANCES" = "0" ] && echo "0.00" || echo "5.40")
- Active Cost: ~\$0.00002 per request
- Storage: ~\$0.10

Next Steps:
1. Review generated configuration files
2. Run: ./gcp/one-click-deploy.sh
3. Monitor deployment progress

Manual Steps Required:
- Verify billing is enabled
- Review IAM permissions
- Configure custom domain (optional)

Files Generated:
- gcp/auto-deploy-config.yaml
- gcp/deployment-templates/terraform.tfvars
- gcp/deployment-templates/cloudbuild.yaml
- gcp/deployment-templates/secrets.json
- gcp/deployment-templates/api-config-example.yaml
EOF

    print_success "Deployment summary created"
    cat gcp/deployment-summary.txt
}

# Main execution
main() {
    clear
    print_header "Virtuoso API CLI - GCP Deployment Wizard"

    check_prerequisites
    detect_gcp_config
    configure_deployment
    generate_configs
    enable_apis
    create_summary

    print_header "Configuration Complete!"
    print_success "All configuration files have been generated."
    print_info "Next step: Run ./gcp/one-click-deploy.sh to deploy"

    # Save configuration state
    cat > gcp/.deployment-state << EOF
CONFIGURED=true
TIMESTAMP=$(date +%s)
PROJECT_ID=$PROJECT_ID
SERVICE_NAME=$SERVICE_NAME
REGION=$REGION
EOF
}

# Run main function
main
