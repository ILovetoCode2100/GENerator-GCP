#!/bin/bash

# Virtuoso API Gateway CDK Deployment Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if AWS CLI is installed and configured
check_aws_cli() {
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install it first."
        exit 1
    fi

    if ! aws sts get-caller-identity &> /dev/null; then
        print_error "AWS CLI is not configured. Please run 'aws configure' first."
        exit 1
    fi

    print_success "AWS CLI is configured"
}

# Check if Node.js and npm are installed
check_node() {
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed. Please install Node.js 18+ first."
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed. Please install npm first."
        exit 1
    fi

    NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
    if [ "$NODE_VERSION" -lt 18 ]; then
        print_error "Node.js version 18 or higher is required. Current version: $(node -v)"
        exit 1
    fi

    print_success "Node.js $(node -v) is installed"
}

# Check if CDK CLI is installed
check_cdk() {
    if ! command -v cdk &> /dev/null; then
        print_warning "CDK CLI is not installed. Installing globally..."
        npm install -g aws-cdk
    fi

    print_success "CDK CLI is available: $(cdk --version)"
}

# Install dependencies
install_dependencies() {
    print_status "Installing CDK dependencies..."
    npm install

    print_status "Installing Lambda dependencies..."
    cd lambda
    npm install
    cd ..

    print_success "Dependencies installed"
}

# Build the project
build_project() {
    print_status "Building TypeScript code..."
    npm run build
    print_success "Build completed"
}

# Bootstrap CDK (if needed)
bootstrap_cdk() {
    print_status "Checking if CDK bootstrap is needed..."
    
    # Get AWS account and region
    AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
    AWS_REGION=$(aws configure get region || echo "us-east-1")
    
    print_status "AWS Account: $AWS_ACCOUNT"
    print_status "AWS Region: $AWS_REGION"
    
    # Check if bootstrap stack exists
    if ! aws cloudformation describe-stacks --stack-name CDKToolkit --region $AWS_REGION &> /dev/null; then
        print_status "Bootstrapping CDK for account $AWS_ACCOUNT in region $AWS_REGION..."
        cdk bootstrap aws://$AWS_ACCOUNT/$AWS_REGION
        print_success "CDK bootstrap completed"
    else
        print_success "CDK is already bootstrapped"
    fi
}

# Deploy the stack
deploy_stack() {
    print_status "Deploying Virtuoso API Stack..."
    
    # Set environment variables for deployment
    export CDK_DEFAULT_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
    export CDK_DEFAULT_REGION=$(aws configure get region || echo "us-east-1")
    export ENVIRONMENT=${ENVIRONMENT:-development}
    
    print_status "Environment: $ENVIRONMENT"
    print_status "Account: $CDK_DEFAULT_ACCOUNT"
    print_status "Region: $CDK_DEFAULT_REGION"
    
    # Deploy with approval bypass for non-interactive environments
    if [ "$CI" = "true" ] || [ "$SKIP_APPROVAL" = "true" ]; then
        cdk deploy --require-approval never
    else
        cdk deploy
    fi
    
    print_success "Stack deployment completed!"
}

# Post-deployment instructions
post_deployment_instructions() {
    print_success "Deployment completed successfully!"
    echo ""
    print_warning "IMPORTANT: Post-deployment steps:"
    echo "1. Update the Secrets Manager secret 'virtuoso-api-config' with your actual API key"
    echo "2. Configure CORS origins if needed (currently set to '*')"
    echo "3. Test the API endpoints with your Bearer token"
    echo ""
    
    # Get the API Gateway URL from CDK outputs
    print_status "Getting API Gateway URL..."
    STACK_NAME="VirtuosoApiStack"
    API_URL=$(aws cloudformation describe-stacks \
        --stack-name $STACK_NAME \
        --query 'Stacks[0].Outputs[?OutputKey==`ApiGatewayUrl`].OutputValue' \
        --output text 2>/dev/null || echo "Not available")
    
    if [ "$API_URL" != "Not available" ] && [ "$API_URL" != "" ]; then
        echo "API Gateway URL: $API_URL"
        echo ""
        echo "Test with:"
        echo "curl -H \"Authorization: Bearer YOUR_TOKEN\" \"$API_URL/api/user\""
    fi
}

# Main deployment process
main() {
    print_status "Starting Virtuoso API Gateway deployment..."
    echo ""
    
    check_aws_cli
    check_node  
    check_cdk
    install_dependencies
    build_project
    bootstrap_cdk
    deploy_stack
    post_deployment_instructions
    
    echo ""
    print_success "Deployment process completed!"
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "destroy")
        print_warning "Destroying the stack..."
        cdk destroy
        print_success "Stack destroyed"
        ;;
    "diff")
        print_status "Showing stack differences..."
        cdk diff
        ;;
    "synth")
        print_status "Synthesizing CloudFormation template..."
        cdk synth
        ;;
    *)
        echo "Usage: $0 [deploy|destroy|diff|synth]"
        echo "  deploy  - Deploy the stack (default)"
        echo "  destroy - Destroy the stack"
        echo "  diff    - Show differences"
        echo "  synth   - Synthesize template"
        exit 1
        ;;
esac