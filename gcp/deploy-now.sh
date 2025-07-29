#!/bin/bash

# Ultra-Simple GCP Deployment Script
# Just run this and follow the prompts!

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

clear

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          🚀 Virtuoso API - Google Cloud Deployment 🚀          ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}This script will deploy your API to Google Cloud in ~15 minutes.${NC}"
echo -e "${GREEN}You only need two things:${NC}"
echo -e "  1. A Google account (for GCP)"
echo -e "  2. Your Virtuoso API token"
echo ""
echo -e "${YELLOW}Press Enter to start...${NC}"
read

# Step 1: Check if gcloud is installed
echo -e "\n${BLUE}Step 1: Checking Google Cloud CLI...${NC}"
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ Google Cloud CLI not found${NC}"
    echo ""
    echo "Please install it first:"
    echo "  Mac: brew install google-cloud-sdk"
    echo "  Other: curl https://sdk.cloud.google.com | bash"
    echo ""
    echo "Then run this script again."
    exit 1
fi
echo -e "${GREEN}✓ Google Cloud CLI found${NC}"

# Step 2: Authenticate
echo -e "\n${BLUE}Step 2: Authenticating with Google Cloud...${NC}"
echo "A browser window will open. Please log in with your Google account."
echo -e "${YELLOW}Press Enter to continue...${NC}"
read

gcloud auth login

# Step 3: Check if we have a project
echo -e "\n${BLUE}Step 3: Setting up GCP Project...${NC}"
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null)

if [ -z "$CURRENT_PROJECT" ]; then
    echo "No project selected. Would you like to:"
    echo "  1) Create a new project (recommended)"
    echo "  2) Select an existing project"
    read -p "Choose (1 or 2): " PROJECT_CHOICE

    if [ "$PROJECT_CHOICE" = "1" ]; then
        PROJECT_ID="virtuoso-api-$(date +%s)"
        echo -e "\n${YELLOW}Creating project: $PROJECT_ID${NC}"
        gcloud projects create $PROJECT_ID --name="Virtuoso API"
        gcloud config set project $PROJECT_ID
    else
        echo -e "\n${YELLOW}Available projects:${NC}"
        gcloud projects list
        read -p "Enter project ID: " PROJECT_ID
        gcloud config set project $PROJECT_ID
    fi
else
    echo -e "${GREEN}✓ Using project: $CURRENT_PROJECT${NC}"
    PROJECT_ID=$CURRENT_PROJECT
fi

# Step 4: Enable billing (required for some services)
echo -e "\n${BLUE}Step 4: Checking billing...${NC}"
BILLING_ENABLED=$(gcloud beta billing projects describe $PROJECT_ID --format="value(billingEnabled)" 2>/dev/null || echo "false")

if [ "$BILLING_ENABLED" != "True" ]; then
    echo -e "${YELLOW}⚠️  Billing is not enabled. Some features will be limited.${NC}"
    echo "To enable full features, visit:"
    echo "https://console.cloud.google.com/billing/linkedaccount?project=$PROJECT_ID"
    echo ""
    echo "Continue with free tier only? (y/n)"
    read -p "> " CONTINUE
    if [ "$CONTINUE" != "y" ]; then
        exit 0
    fi
fi

# Step 5: Get Virtuoso API Token
echo -e "\n${BLUE}Step 5: Virtuoso API Configuration${NC}"
echo "Please enter your Virtuoso API token"
echo "(Get it from: https://app.virtuoso.qa > Settings > API)"
read -sp "API Token: " VIRTUOSO_TOKEN
echo ""

# Step 6: Quick deployment
echo -e "\n${BLUE}Step 6: Deploying your API...${NC}"
echo "This will take about 10-15 minutes. Starting deployment..."
echo ""

# Create a temporary deployment script with all the automation
cat > /tmp/deploy-virtuoso.sh << 'DEPLOY_SCRIPT'
#!/bin/bash
set -e

PROJECT_ID="$1"
VIRTUOSO_TOKEN="$2"

# Enable APIs
echo "🔧 Enabling required APIs..."
gcloud services enable run.googleapis.com \
    cloudbuild.googleapis.com \
    secretmanager.googleapis.com \
    firestore.googleapis.com \
    --project=$PROJECT_ID

# Create secrets
echo "🔐 Setting up secrets..."
echo -n "$VIRTUOSO_TOKEN" | gcloud secrets create virtuoso-api-token \
    --data-file=- \
    --replication-policy="automatic" \
    --project=$PROJECT_ID 2>/dev/null || \
    echo -n "$VIRTUOSO_TOKEN" | gcloud secrets versions add virtuoso-api-token --data-file=-

# Generate API keys
API_KEY_1=$(openssl rand -hex 32)
API_KEY_2=$(openssl rand -hex 32)
echo -n "[\"$API_KEY_1\",\"$API_KEY_2\"]" | gcloud secrets create api-keys \
    --data-file=- \
    --replication-policy="automatic" \
    --project=$PROJECT_ID 2>/dev/null || \
    echo -n "[\"$API_KEY_1\",\"$API_KEY_2\"]" | gcloud secrets versions add api-keys --data-file=-

# Build and deploy
echo "🚀 Building and deploying to Cloud Run..."
cd "$(dirname "$0")"

# Create a minimal Cloud Run service with gcloud (no need for complex build)
gcloud run deploy virtuoso-api \
    --source . \
    --platform managed \
    --region us-central1 \
    --allow-unauthenticated \
    --memory 1Gi \
    --cpu 1 \
    --min-instances 0 \
    --max-instances 100 \
    --set-env-vars="GCP_PROJECT_ID=$PROJECT_ID" \
    --set-secrets="VIRTUOSO_API_TOKEN=virtuoso-api-token:latest,API_KEYS=api-keys:latest" \
    --project=$PROJECT_ID

# Get service URL
SERVICE_URL=$(gcloud run services describe virtuoso-api \
    --platform managed \
    --region us-central1 \
    --format 'value(status.url)' \
    --project=$PROJECT_ID)

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📋 Deployment Summary"
echo "===================="
echo "API URL: $SERVICE_URL"
echo "Project: $PROJECT_ID"
echo ""
echo "🔑 Your API Keys:"
echo "Client 1: $API_KEY_1"
echo "Client 2: $API_KEY_2"
echo ""
echo "📝 Save these API keys - you won't see them again!"
echo ""
echo "🧪 Test your API:"
echo "curl $SERVICE_URL/health"
echo ""
echo "curl -H \"X-API-Key: $API_KEY_1\" $SERVICE_URL/api/v1/commands"
echo ""
echo "📚 View API docs at:"
echo "$SERVICE_URL/docs"

# Save deployment info
cat > deployment-info.txt << EOF
Virtuoso API Deployment
======================
Date: $(date)
Project: $PROJECT_ID
URL: $SERVICE_URL
Region: us-central1

API Keys:
Client 1: $API_KEY_1
Client 2: $API_KEY_2

Test Commands:
curl $SERVICE_URL/health
curl -H "X-API-Key: $API_KEY_1" $SERVICE_URL/api/v1/commands

Management:
View logs: gcloud run logs read --service=virtuoso-api
Update: gcloud run services update virtuoso-api
Delete: gcloud run services delete virtuoso-api
EOF

echo ""
echo "💾 Deployment details saved to: deployment-info.txt"
DEPLOY_SCRIPT

# Make it executable and run
chmod +x /tmp/deploy-virtuoso.sh
cd "$(dirname "$0")/.."  # Go to project root
/tmp/deploy-virtuoso.sh "$PROJECT_ID" "$VIRTUOSO_TOKEN"

# Cleanup
rm -f /tmp/deploy-virtuoso.sh

echo ""
echo -e "${GREEN}🎉 Congratulations! Your Virtuoso API is now live on Google Cloud!${NC}"
echo ""
echo "Next steps:"
echo "1. Test your API with the curl commands above"
echo "2. View the API documentation in your browser"
echo "3. Monitor usage at: https://console.cloud.google.com/run"
echo ""
echo -e "${BLUE}Questions? Check deployment-info.txt for all details.${NC}"
