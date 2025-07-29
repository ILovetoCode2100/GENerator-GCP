# Virtuoso API CLI - GCP Quick Start Guide

Deploy the Virtuoso API CLI to Google Cloud Platform in 5 minutes!

## Prerequisites

Ensure you have:

- Google Cloud account with billing enabled
- `gcloud` CLI installed and authenticated
- Virtuoso API credentials

## 5-Minute Deployment

### 1. Clone and Configure (1 minute)

```bash
# Clone the repository
git clone https://github.com/your-org/virtuoso-GENerator.git
cd virtuoso-GENerator/gcp

# Set required environment variables
export GCP_PROJECT_ID="my-virtuoso-project"  # Choose your project ID
export VIRTUOSO_API_KEY="your-api-key"       # Your Virtuoso API key
export VIRTUOSO_ORG_ID="your-org-id"         # Your Virtuoso org ID
```

### 2. One-Command Setup (2 minutes)

```bash
# Create project and set up everything
./setup-project.sh --project-id $GCP_PROJECT_ID --create-project && \
./secrets-setup.sh --non-interactive && \
./deploy.sh --skip-monitoring
```

### 3. Verify Deployment (1 minute)

```bash
# Get your service URL
SERVICE_URL=$(gcloud run services describe virtuoso-api-cli \
  --region us-central1 --format 'value(status.url)')

# Test it
curl "$SERVICE_URL/health"
```

### 4. Start Using (1 minute)

```bash
# Test the CLI through the API
curl -X POST "$SERVICE_URL/api/v1/execute" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"command": "list-projects"}'
```

## Common Configurations

### Development Environment

```bash
# Minimal setup for development
./deploy.sh \
  --environment development \
  --skip-monitoring \
  --skip-functions
```

### Production with Monitoring

```bash
# Full production setup
./deploy.sh \
  --environment production \
  --alert-email "ops@example.com"
```

### Local Development Only

```bash
# Just run locally with emulators
./deploy-local.sh
```

## Quick Commands

### View Logs

```bash
gcloud run logs read --service virtuoso-api-cli --limit 50
```

### Update Service

```bash
./deploy.sh --skip-terraform --skip-monitoring
```

### Emergency Rollback

```bash
./rollback.sh --type service
```

### Check Costs

```bash
gcloud billing projects describe $GCP_PROJECT_ID
```

## Environment Variables Reference

```bash
# Minimal required setup
export GCP_PROJECT_ID="your-project"
export VIRTUOSO_API_KEY="your-key"
export VIRTUOSO_ORG_ID="your-org"

# Optional customization
export GCP_REGION="us-east1"              # Default: us-central1
export ENVIRONMENT="staging"               # Default: production
export ALERT_EMAIL="alerts@example.com"    # For monitoring alerts
```

## Troubleshooting

### "Project already exists"

```bash
# Use existing project
./setup-project.sh --project-id $GCP_PROJECT_ID  # Without --create-project
```

### "Billing account required"

```bash
# List billing accounts
gcloud billing accounts list

# Link billing
gcloud billing projects link $GCP_PROJECT_ID --billing-account=ACCOUNT_ID
```

### "Permission denied"

```bash
# Ensure you're authenticated
gcloud auth login
gcloud config set project $GCP_PROJECT_ID
```

## Next Steps

1. **Add monitoring**: Run `./monitoring-setup.sh`
2. **Set up CI/CD**: Configure GitHub integration
3. **Customize**: Edit `terraform/terraform.tfvars`
4. **Scale**: Adjust Cloud Run settings

## Cost Estimate

Minimal deployment: ~$50-100/month

- Cloud Run: Pay per request
- Firestore: Free tier usually sufficient
- Monitoring: Basic tier free

## Clean Up

```bash
# Delete everything (WARNING: This removes all resources)
gcloud projects delete $GCP_PROJECT_ID
```

---

Need help? Check the [full deployment guide](README.md) or open an issue!
