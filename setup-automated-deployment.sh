#!/bin/bash

# Automated Deployment Setup Script
# This sets up everything needed for automated deployments

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Automated Deployment Setup${NC}"
echo "=============================="
echo ""
echo "This script will set up automated deployment to Render."
echo "You'll be able to deploy with a single git push!"
echo ""

# Check if git repo exists
if [ ! -d .git ]; then
    echo -e "${YELLOW}Initializing git repository...${NC}"
    git init
    git add .
    git commit -m "Initial commit with Render deployment"
fi

# Get GitHub repo info
echo -e "${YELLOW}Step 1: GitHub Repository${NC}"
echo "Please enter your GitHub repository URL"
echo "Format: https://github.com/username/repo-name"
read -p "GitHub Repo URL: " GITHUB_REPO

# Extract owner and repo name
REPO_OWNER=$(echo $GITHUB_REPO | sed -E 's/.*github.com[:/]([^/]+)\/.*/\1/')
REPO_NAME=$(echo $GITHUB_REPO | sed -E 's/.*\/([^.]+)(\.git)?$/\1/')

echo "Repository: $REPO_OWNER/$REPO_NAME"

# Set up git remote
if ! git remote | grep -q origin; then
    git remote add origin $GITHUB_REPO
else
    git remote set-url origin $GITHUB_REPO
fi

# Get credentials
echo ""
echo -e "${YELLOW}Step 2: API Credentials${NC}"
echo ""
echo "You need the following credentials:"
echo "1. Render API Key - Get from: https://dashboard.render.com/account/api-keys"
echo "2. Virtuoso API Token - Get from: https://app.virtuoso.qa > Settings > API"
echo ""

read -sp "Render API Key: " RENDER_API_KEY
echo ""
read -sp "Virtuoso API Token: " VIRTUOSO_API_TOKEN
echo ""
read -p "Virtuoso Org ID [2242]: " VIRTUOSO_ORG_ID
VIRTUOSO_ORG_ID=${VIRTUOSO_ORG_ID:-2242}

# Create GitHub CLI script to set secrets
echo ""
echo -e "${YELLOW}Step 3: Setting up GitHub Secrets${NC}"
echo ""
echo "To set GitHub secrets, you can either:"
echo ""
echo "Option 1: Use GitHub CLI (if installed)"
echo "Option 2: Set manually in GitHub UI"
echo ""

# Check if GitHub CLI is installed
if command -v gh &> /dev/null; then
    echo "GitHub CLI detected. Would you like to set secrets automatically?"
    read -p "Set secrets with GitHub CLI? (y/N): " USE_GH_CLI

    if [[ $USE_GH_CLI =~ ^[Yy]$ ]]; then
        echo "Setting GitHub secrets..."
        gh secret set RENDER_API_KEY -b "$RENDER_API_KEY" -R "$REPO_OWNER/$REPO_NAME"
        gh secret set VIRTUOSO_API_TOKEN -b "$VIRTUOSO_API_TOKEN" -R "$REPO_OWNER/$REPO_NAME"
        gh secret set VIRTUOSO_ORG_ID -b "$VIRTUOSO_ORG_ID" -R "$REPO_OWNER/$REPO_NAME"
        echo -e "${GREEN}✅ Secrets set successfully!${NC}"
    fi
else
    echo -e "${YELLOW}GitHub CLI not found.${NC}"
fi

# Create script for manual setup
cat > setup-github-secrets.md << EOF
# GitHub Secrets Setup

Go to: https://github.com/$REPO_OWNER/$REPO_NAME/settings/secrets/actions

Add these secrets:

1. **RENDER_API_KEY**
   Value: $RENDER_API_KEY

2. **VIRTUOSO_API_TOKEN**
   Value: $VIRTUOSO_API_TOKEN

3. **VIRTUOSO_ORG_ID**
   Value: $VIRTUOSO_ORG_ID

After adding these secrets, you can deploy by:
1. Pushing to main branch (auto-deploy)
2. Going to Actions tab and clicking "Run workflow"
EOF

echo ""
echo -e "${GREEN}✅ Setup script created: setup-github-secrets.md${NC}"

# Create one-click deployment script
cat > deploy.sh << 'EOF'
#!/bin/bash
# One-click deployment script

echo "🚀 Deploying to Render..."

# Check if we have uncommitted changes
if [[ -n $(git status -s) ]]; then
    echo "📝 Committing changes..."
    git add .
    git commit -m "Deploy to Render - $(date +%Y-%m-%d-%H:%M:%S)"
fi

# Push to trigger deployment
echo "📤 Pushing to GitHub..."
git push origin main

echo ""
echo "✅ Deployment triggered!"
echo ""
echo "Monitor deployment at:"
echo "  GitHub Actions: https://github.com/$(git remote get-url origin | sed -E 's/.*github.com[:/]([^.]+)(\.git)?$/\1/')/actions"
echo "  Render Dashboard: https://dashboard.render.com"
echo ""
echo "Your API will be available at: https://virtuoso-api-production.onrender.com"
EOF

chmod +x deploy.sh

# Create test script
cat > test-deployment.sh << 'EOF'
#!/bin/bash
# Test deployed service

SERVICE_URL="https://virtuoso-api-production.onrender.com"

echo "🧪 Testing deployment at $SERVICE_URL"
echo ""

# Test health
echo "Testing health endpoint..."
curl -f "$SERVICE_URL/health" | jq . || echo "❌ Health check failed"

echo ""
echo "Testing detailed health..."
curl -f "$SERVICE_URL/health?detailed=true" | jq . || echo "❌ Detailed health check failed"

echo ""
echo "To test authenticated endpoints, use:"
echo "curl -H 'X-API-Key: YOUR_API_KEY' $SERVICE_URL/api/v1/commands"
EOF

chmod +x test-deployment.sh

# Create render-dashboard.yaml for one-click Render deploy
cat > render-dashboard.yaml << EOF
# Deploy this blueprint directly from Render Dashboard
# Go to: https://dashboard.render.com/blueprints

services:
  - type: web
    name: virtuoso-api-quick
    runtime: docker
    dockerfilePath: ./Dockerfile.render
    dockerContext: .
    envVars:
      - key: VIRTUOSO_API_TOKEN
        value: "$VIRTUOSO_API_TOKEN"
      - key: VIRTUOSO_ORG_ID
        value: "$VIRTUOSO_ORG_ID"
      - key: API_KEYS
        generateValue: true
      - key: REDIS_URL
        fromService:
          type: redis
          name: virtuoso-redis-quick
          property: connectionString
    healthCheckPath: /health
    autoDeploy: false

  - type: redis
    name: virtuoso-redis-quick
    ipAllowList: []
    maxmemoryPolicy: allkeys-lru
    plan: free
EOF

echo ""
echo -e "${GREEN}✅ Automated deployment setup complete!${NC}"
echo ""
echo -e "${BLUE}📋 What's been created:${NC}"
echo "  - GitHub Actions workflow (.github/workflows/deploy-render.yml)"
echo "  - One-click deploy script (./deploy.sh)"
echo "  - Test script (./test-deployment.sh)"
echo "  - Manual setup guide (setup-github-secrets.md)"
echo "  - Render dashboard config (render-dashboard.yaml)"
echo ""
echo -e "${BLUE}🚀 How to deploy:${NC}"
echo ""
echo "Option 1: Automated (recommended)"
echo "  1. Set GitHub secrets (see setup-github-secrets.md)"
echo "  2. Run: ./deploy.sh"
echo ""
echo "Option 2: GitHub Actions"
echo "  1. Push to main branch"
echo "  2. Or go to Actions tab and click 'Run workflow'"
echo ""
echo "Option 3: Render Dashboard"
echo "  1. Go to https://dashboard.render.com/blueprints"
echo "  2. Create new blueprint from render-dashboard.yaml"
echo ""
echo -e "${YELLOW}⚠️  Important: Save your credentials securely!${NC}"
echo "Your credentials have been masked in this output."
echo "Check setup-github-secrets.md for the values to set in GitHub."
