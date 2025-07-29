# 🚀 Virtuoso API CLI - 5-Minute GCP Deployment

Deploy the Virtuoso API CLI to Google Cloud in under 5 minutes!

## Prerequisites ✅

You need:

1. Google Cloud account with billing enabled
2. `gcloud` CLI installed ([Download here](https://cloud.google.com/sdk/docs/install))
3. Docker Desktop running ([Download here](https://www.docker.com/products/docker-desktop))
4. Your Virtuoso API token

## Step 1: Authenticate (30 seconds) 🔐

```bash
gcloud auth login
```

## Step 2: Get Your Virtuoso API Token (1 minute) 🔑

1. Log into Virtuoso: https://app.virtuoso.qa
2. Go to Settings → API Keys
3. Copy your API token

## Step 3: Deploy (3 minutes) 🚀

```bash
# Clone the repository (if not already done)
git clone https://github.com/your-org/virtuoso-GENerator.git
cd virtuoso-GENerator

# Make scripts executable
chmod +x gcp/*.sh

# Run the deployment wizard
./gcp/deploy-wizard.sh
```

When prompted:

- Press Enter to use defaults (recommended)
- Paste your Virtuoso API token when asked
- Type "y" to create a new project (if needed)

## Step 4: Complete Deployment (1 minute) ⚡

```bash
# Run the automated deployment
./gcp/one-click-deploy.sh
```

## Done! 🎉

Your API is now live! The script will show you:

- **Service URL**: `https://virtuoso-api-cli-xxxxx.run.app`
- **Test command**: Copy and run to verify

## Test Your Deployment 🧪

```bash
# Copy the URL from the deployment output and test:
curl https://your-service-url.run.app/health
```

## What Just Happened? 🤔

You deployed:

- ✅ Containerized API service on Cloud Run
- ✅ Automatic HTTPS endpoint
- ✅ Auto-scaling (0 to 100 instances)
- ✅ Secure secret storage
- ✅ Pay-per-use pricing (scales to zero)

## Total Cost: ~$0/month when idle 💰

## Need Help? 🆘

- **Deployment failed?** Check `gcp/deployment-report-*.txt`
- **View logs:** `gcloud run logs tail --service=virtuoso-api-cli`
- **Common issues:** See [Troubleshooting](README.md#troubleshooting)

---

🎯 **Pro tip:** Save your deployment configuration for future updates:

```bash
cp gcp/auto-deploy-config.yaml ~/virtuoso-deployment-backup.yaml
```
