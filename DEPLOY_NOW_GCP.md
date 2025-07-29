# 🚀 Deploy to GCP Now - 3 Simple Steps

I've created an automated deployment system for you. Here's all you need to do:

## Prerequisites (One-time setup)

```bash
# 1. Install Google Cloud CLI (if not already installed)
# Mac:
brew install google-cloud-sdk

# Other OS:
curl https://sdk.cloud.google.com | bash
```

## 🎯 Deploy in 3 Steps

### Step 1: Authenticate with Google Cloud

```bash
gcloud auth login
```

_This opens your browser - just click "Allow"_

### Step 2: Get Your Virtuoso API Token

1. Go to https://app.virtuoso.qa
2. Navigate to Settings > API
3. Click "Generate API Token"
4. Copy the token

### Step 3: Run the Deployment

```bash
cd gcp
./one-click-deploy.sh
```

**That's it!** The script will:

- ✅ Auto-detect or create a GCP project
- ✅ Enable all required APIs
- ✅ Build and deploy your API
- ✅ Set up all cloud services
- ✅ Configure monitoring
- ✅ Run tests
- ✅ Give you the API URL

## 📋 What Happens During Deployment

```
🚀 Virtuoso API GCP Deployment
==============================
✓ Checking prerequisites...
✓ Authenticating with GCP...
✓ Creating/selecting project...
✓ Enabling APIs...
✓ Building container...
✓ Creating secrets...
✓ Deploying to Cloud Run...
✓ Setting up Firestore...
✓ Configuring monitoring...
✓ Running tests...

✅ Deployment Complete!

Your API is live at:
https://virtuoso-api-abc123-uc.a.run.app

Test commands:
curl https://virtuoso-api-abc123-uc.a.run.app/health
curl -H "X-API-Key: your-key" https://virtuoso-api-abc123-uc.a.run.app/api/v1/commands

View logs:
gcloud run logs read --service=virtuoso-api

Monitor:
https://console.cloud.google.com/run
```

## 🔑 Generated Credentials

The deployment automatically:

- Creates 3 API keys for your clients
- Stores them in Secret Manager
- Shows them in the deployment report
- Saves them to `deployment-report.txt`

## 💰 Cost

- **While idle**: $0.00/month
- **Active usage**: ~$0.50 per million requests
- **Free tier**: First 2M requests/month free
- **Estimated monthly**: $0-50 for most use cases

## 🆘 If Something Goes Wrong

The script handles most errors automatically, but if needed:

```bash
# Check what went wrong
cat deployment-report.txt

# View detailed logs
gcloud run logs read --service=virtuoso-api --limit=50

# Rollback if needed
./rollback.sh

# Get help
echo "Error details:" && gcloud run services describe virtuoso-api
```

## 📊 After Deployment

1. **Test your API**:

   ```bash
   # Use the URL from deployment output
   curl https://your-api-url.run.app/health
   ```

2. **View API docs**:

   ```
   https://your-api-url.run.app/docs
   ```

3. **Monitor usage**:
   ```
   https://console.cloud.google.com/run
   ```

## 🎉 That's It!

Your Virtuoso API is now:

- ✅ Live on Google Cloud
- ✅ Auto-scaling from 0 to millions
- ✅ Secured with API keys
- ✅ Monitored 24/7
- ✅ Costing $0 when idle

**Total time**: ~15 minutes
**Your effort**: 3 simple steps
**Result**: Production-ready API

---

## 🤖 What I Can Do After Deployment

Once deployed, I can help you:

- Monitor performance and costs
- Debug any issues
- Scale based on traffic
- Add new features
- Optimize configurations

Just share your project ID and I can assist with management!
