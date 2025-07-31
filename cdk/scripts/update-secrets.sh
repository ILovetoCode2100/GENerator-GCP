#!/bin/bash

# Update Virtuoso API configuration in AWS Secrets Manager
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
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

# Default values
SECRET_NAME="virtuoso-api-config"
BASE_URL="https://api-app2.virtuoso.qa/api"
ORG_ID="2242"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --api-key)
            API_KEY="$2"
            shift 2
            ;;
        --base-url)
            BASE_URL="$2"
            shift 2
            ;;
        --org-id)
            ORG_ID="$2"
            shift 2
            ;;
        --secret-name)
            SECRET_NAME="$2"
            shift 2
            ;;
        --region)
            AWS_REGION="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 --api-key YOUR_API_KEY [options]"
            echo ""
            echo "Required:"
            echo "  --api-key KEY      Your Virtuoso API key"
            echo ""
            echo "Optional:"
            echo "  --base-url URL     API base URL (default: $BASE_URL)"
            echo "  --org-id ID        Organization ID (default: $ORG_ID)"
            echo "  --secret-name NAME Secret name (default: $SECRET_NAME)"
            echo "  --region REGION    AWS region"
            echo "  -h, --help         Show this help"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Validate required parameters
if [ -z "$API_KEY" ]; then
    print_error "API key is required. Use --api-key parameter."
    echo "Example: $0 --api-key your-virtuoso-api-key-here"
    exit 1
fi

# Set AWS region if not provided
if [ -z "$AWS_REGION" ]; then
    AWS_REGION=$(aws configure get region || echo "us-east-1")
fi

print_status "Configuration:"
print_status "  Secret Name: $SECRET_NAME"
print_status "  Base URL: $BASE_URL"
print_status "  Organization ID: $ORG_ID"
print_status "  AWS Region: $AWS_REGION"
print_status "  API Key: ${API_KEY:0:8}... (hidden)"

# Create the secret value JSON
SECRET_VALUE=$(cat <<EOF
{
  "virtuosoApiBaseUrl": "$BASE_URL",
  "organizationId": "$ORG_ID",
  "apiKey": "$API_KEY"
}
EOF
)

# Check if secret exists
print_status "Checking if secret exists..."
if aws secretsmanager describe-secret --secret-id "$SECRET_NAME" --region "$AWS_REGION" &> /dev/null; then
    print_status "Secret exists. Updating..."
    
    aws secretsmanager update-secret \
        --secret-id "$SECRET_NAME" \
        --secret-string "$SECRET_VALUE" \
        --region "$AWS_REGION"
    
    print_success "Secret updated successfully!"
else
    print_warning "Secret does not exist. This script is intended to update existing secrets."
    print_warning "The secret should be created during CDK deployment."
    print_error "Please deploy the CDK stack first, then run this script."
    exit 1
fi

# Verify the update
print_status "Verifying secret update..."
STORED_CONFIG=$(aws secretsmanager get-secret-value \
    --secret-id "$SECRET_NAME" \
    --region "$AWS_REGION" \
    --query 'SecretString' \
    --output text)

if echo "$STORED_CONFIG" | jq -e '.virtuosoApiBaseUrl' > /dev/null 2>&1; then
    STORED_BASE_URL=$(echo "$STORED_CONFIG" | jq -r '.virtuosoApiBaseUrl')
    STORED_ORG_ID=$(echo "$STORED_CONFIG" | jq -r '.organizationId')
    
    print_success "Secret verification completed:"
    print_success "  Base URL: $STORED_BASE_URL"
    print_success "  Organization ID: $STORED_ORG_ID"
    print_success "  API Key: Updated (not shown for security)"
else
    print_error "Failed to verify secret format. Please check the secret manually."
fi

echo ""
print_success "Configuration update completed!"
print_status "Your API Gateway is now configured to use the updated Virtuoso API settings."
print_status "You can now test your API endpoints."