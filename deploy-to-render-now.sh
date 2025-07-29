#!/bin/bash

# Quick Deploy to Render Script
# This script will guide you through deploying to Render in minutes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Virtuoso API CLI - Quick Deploy to Render${NC}"
echo "============================================"
echo ""

# Step 1: Check if user has Render CLI
echo -e "${YELLOW}Step 1: Checking Render CLI...${NC}"
if ! command -v render &> /dev/null; then
    echo -e "${RED}❌ Render CLI not found${NC}"
    echo ""
    echo "Please install Render CLI first:"
    echo "  Mac: brew install render"
    echo "  Other: npm install -g @render-oss/cli"
    echo ""
    echo "Or deploy manually through the Render Dashboard:"
    echo "1. Go to https://render.com"
    echo "2. Sign in and click 'New' > 'Blueprint'"
    echo "3. Connect your GitHub repo"
    echo "4. Render will auto-detect render.yaml"
    echo ""
    exit 1
else
    echo -e "${GREEN}✅ Render CLI found${NC}"
fi

# Step 2: Check if logged in
echo ""
echo -e "${YELLOW}Step 2: Checking Render authentication...${NC}"
if ! render auth whoami &> /dev/null; then
    echo "Please log in to Render:"
    render auth login
fi
echo -e "${GREEN}✅ Authenticated with Render${NC}"

# Step 3: Generate API keys
echo ""
echo -e "${YELLOW}Step 3: Generating API Keys for your service...${NC}"
API_KEY_1=$(openssl rand -hex 32)
API_KEY_2=$(openssl rand -hex 32)
API_KEY_3=$(openssl rand -hex 32)

echo -e "${GREEN}✅ Generated 3 API keys${NC}"
echo ""
echo "Your API Keys (save these!):"
echo "  Client 1: $API_KEY_1"
echo "  Client 2: $API_KEY_2"
echo "  Client 3: $API_KEY_3"

# Step 4: Get Virtuoso credentials
echo ""
echo -e "${YELLOW}Step 4: Virtuoso API Credentials${NC}"
echo "Please enter your Virtuoso API Token"
echo "(Get it from: https://app.virtuoso.qa > Settings > API)"
read -sp "Virtuoso API Token: " VIRTUOSO_TOKEN
echo ""

# Optional: Organization ID
echo ""
echo "Enter your Virtuoso Organization ID (press Enter for default: 2242):"
read -p "Organization ID [2242]: " ORG_ID
ORG_ID=${ORG_ID:-2242}

# Step 5: Create environment file
echo ""
echo -e "${YELLOW}Step 5: Creating environment configuration...${NC}"
cat > .render-env.tmp << EOF
VIRTUOSO_API_TOKEN=$VIRTUOSO_TOKEN
API_KEYS=["$API_KEY_1", "$API_KEY_2", "$API_KEY_3"]
VIRTUOSO_ORG_ID=$ORG_ID
LOG_LEVEL=INFO
RATE_LIMIT_PER_MINUTE=60
CORS_ALLOWED_ORIGINS=*
EOF
echo -e "${GREEN}✅ Environment configuration created${NC}"

# Step 6: Deploy
echo ""
echo -e "${YELLOW}Step 6: Ready to deploy!${NC}"
echo ""
echo "This will create:"
echo "  - Web Service (FastAPI)"
echo "  - Redis (for caching)"
echo "  - Auto-scaling (1-5 instances)"
echo "  - Health monitoring"
echo ""
echo -e "${BLUE}Estimated cost:${NC}"
echo "  - Free tier: $0/month (good for testing)"
echo "  - Starter: ~$25/month (small teams)"
echo "  - Pro: ~$85/month (production)"
echo ""
read -p "Deploy now? (y/N): " DEPLOY_CONFIRM

if [[ $DEPLOY_CONFIRM =~ ^[Yy]$ ]]; then
    echo ""
    echo -e "${YELLOW}🚀 Deploying to Render...${NC}"

    # Create the blueprint deployment
    render blueprint deploy \
        --file render.yaml \
        --env-file .render-env.tmp \
        --name "virtuoso-api-${USER}-$(date +%s)"

    echo ""
    echo -e "${GREEN}✅ Deployment initiated!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Check deployment status: https://dashboard.render.com"
    echo "2. Your API will be available at: https://virtuoso-api-*.onrender.com"
    echo "3. Test with: curl https://your-url.onrender.com/health"
    echo ""
    echo "Your API Keys (save these!):"
    echo "  $API_KEY_1"
    echo "  $API_KEY_2"
    echo "  $API_KEY_3"
else
    echo ""
    echo "Deployment cancelled."
    echo ""
    echo "To deploy manually:"
    echo "1. Go to https://render.com"
    echo "2. Click 'New' > 'Blueprint'"
    echo "3. Connect your GitHub repo"
    echo "4. Add these environment variables:"
    cat .render-env.tmp
fi

# Cleanup
rm -f .render-env.tmp

echo ""
echo -e "${BLUE}📚 Documentation:${NC}"
echo "  - Full guide: deployment/render/README.md"
echo "  - Troubleshooting: RENDER_SETUP_GUIDE.md"
echo "  - API docs will be at: https://your-url.onrender.com/docs"
