# 🤖 Automated Deployment & Testing

I've created an automated deployment system that allows you to deploy and test with minimal effort!

## 🚀 What I've Set Up

### 1. **GitHub Actions Workflow**

- Automatically deploys when you push to main
- Generates fresh API keys each deployment
- Runs health checks and tests
- Supports multiple environments (dev, staging, prod)

### 2. **One-Click Deploy Script**

```bash
./deploy.sh
```

This single command:

- Commits any changes
- Pushes to GitHub
- Triggers automatic deployment
- Shows you where to monitor progress

### 3. **Automated Testing**

```bash
./test-deployment.sh
```

Tests your deployed service automatically

## 📋 What You Need to Provide

To enable automated deployment, you need to provide these 3 secrets:

### 1. **Render API Key**

- Get from: https://dashboard.render.com/account/api-keys
- Click "Create API Key"
- Copy the key

### 2. **Virtuoso API Token**

- Get from: https://app.virtuoso.qa
- Go to Settings > API
- Generate token

### 3. **GitHub Repository**

- Your GitHub repo URL
- Example: https://github.com/yourusername/virtuoso-generator

## 🔧 Quick Setup (5 minutes)

Run this command and follow the prompts:

```bash
./setup-automated-deployment.sh
```

This will:

1. Ask for your credentials
2. Set up GitHub secrets (if you have GitHub CLI)
3. Create deployment scripts
4. Configure everything for you

## 🎯 After Setup

### Deploy Anytime

```bash
# Make changes to your code
git add .
git commit -m "Update API"

# Deploy with one command
./deploy.sh
```

### Monitor Deployment

- GitHub Actions: See real-time logs
- Render Dashboard: View service status
- API Endpoint: Test your deployed service

### Automatic Features

- ✅ Generates new API keys each deployment
- ✅ Runs health checks
- ✅ Tests endpoints
- ✅ Saves API keys as artifacts
- ✅ Updates deployment status

## 🔄 Continuous Deployment

Once set up:

1. **Every push to main** → Automatic deployment
2. **Manual trigger** → Go to Actions tab, click "Run workflow"
3. **One command** → Use `./deploy.sh`

## 🧪 Testing After Deployment

The workflow automatically:

- Waits for service to be ready
- Checks health endpoint
- Tests API endpoints
- Validates authentication

You can also manually test:

```bash
# Get your service URL from Render dashboard
SERVICE_URL="https://virtuoso-api-production.onrender.com"

# Test health
curl $SERVICE_URL/health

# Test with API key (check GitHub Actions artifacts for keys)
curl -H "X-API-Key: your-key" $SERVICE_URL/api/v1/commands
```

## 🔐 Security

- API keys are generated fresh each deployment
- Credentials are stored as GitHub secrets
- Keys are masked in logs
- Old keys are automatically invalidated

## 📊 What Happens on Each Deploy

1. **Build Phase**

   - Builds Docker container
   - Installs dependencies
   - Compiles Go binary

2. **Deploy Phase**

   - Pushes to Render
   - Sets environment variables
   - Starts services

3. **Validation Phase**

   - Health checks
   - API tests
   - Status reporting

4. **Notification Phase**
   - GitHub status update
   - Deployment URL provided
   - API keys saved as artifacts

## 🆘 Troubleshooting

If deployment fails:

1. Check GitHub Actions logs
2. Verify secrets are set correctly
3. Check Render dashboard for errors
4. Run `./test-deployment.sh` to debug

## 🎉 Summary

With this setup:

- **You push code** → It deploys automatically
- **You run `./deploy.sh`** → It handles everything
- **Tests run automatically** → You know it's working
- **API keys generated** → Fresh keys each time

No manual steps needed after initial setup!
