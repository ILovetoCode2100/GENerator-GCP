# Virtuoso API CLI - GCP Deployment Guide

## Overview

This directory contains comprehensive deployment scripts and infrastructure as code for deploying the Virtuoso API CLI to Google Cloud Platform (GCP). The deployment creates a scalable, secure, and monitored cloud infrastructure.

## Architecture

![Architecture Diagram](architecture/architecture-diagram.png)

The deployment creates the following GCP resources:

### Core Services

- **Cloud Run**: Hosts the main API service (auto-scaling, serverless)
- **Cloud Functions**: Background tasks and webhook handlers
- **Firestore**: NoSQL database for session and test data
- **Cloud Storage**: Binary storage and test artifacts

### Supporting Services

- **Cloud Build**: CI/CD pipeline
- **Secret Manager**: Secure credential storage
- **Cloud Scheduler**: Cron jobs and periodic tasks
- **Cloud Tasks**: Asynchronous task queue
- **Pub/Sub**: Event-driven messaging

### Monitoring & Operations

- **Cloud Monitoring**: Metrics and dashboards
- **Cloud Logging**: Centralized log management
- **Cloud Trace**: Distributed tracing
- **Error Reporting**: Automatic error tracking

## Prerequisites

1. **Tools Required**:

   - Google Cloud SDK (`gcloud`)
   - Terraform >= 1.0
   - Docker
   - Go 1.21+
   - jq

2. **GCP Account**:

   - Active GCP account with billing enabled
   - Appropriate IAM permissions (Project Editor or specific roles)

3. **Virtuoso Credentials**:
   - Virtuoso API key
   - Virtuoso Organization ID

## Quick Start

### 1. Initial Setup (One-time)

```bash
# Set environment variables
export GCP_PROJECT_ID="your-project-id"
export VIRTUOSO_API_KEY="your-virtuoso-api-key"
export VIRTUOSO_ORG_ID="your-virtuoso-org-id"

# Create and configure GCP project
./setup-project.sh --project-id $GCP_PROJECT_ID --create-project

# Configure secrets
./secrets-setup.sh
```

### 2. Deploy Application

```bash
# Full deployment (recommended for first time)
./deploy.sh

# Or with specific options
./deploy.sh --project-id $GCP_PROJECT_ID --region us-central1 --environment production
```

### 3. Verify Deployment

The deployment script will output the service URL. Test it:

```bash
# Get service URL
SERVICE_URL=$(gcloud run services describe virtuoso-api-cli --region us-central1 --format 'value(status.url)')

# Test health endpoint
curl "$SERVICE_URL/health"

# Test API (requires authentication)
curl -H "X-API-Key: your-api-key" "$SERVICE_URL/api/v1/commands/list"
```

## Deployment Scripts

### `setup-project.sh`

Initial GCP project setup including:

- Project creation (optional)
- API enablement
- Service account creation
- IAM role configuration
- Base infrastructure (VPC, storage buckets, Firestore)

**Usage**:

```bash
./setup-project.sh [OPTIONS]
  --project-id ID          GCP project ID
  --project-name NAME      Project display name
  --billing-account ID     Billing account ID
  --organization ID        Organization ID (optional)
  --folder ID             Folder ID (optional)
  --region REGION         GCP region (default: us-central1)
  --create-project        Create new project
```

### `deploy.sh`

Master deployment script that orchestrates the entire deployment:

- Prerequisites check
- Terraform infrastructure
- CLI binary build
- Cloud Run deployment
- Cloud Functions deployment
- CI/CD setup
- Monitoring configuration
- Smoke tests

**Usage**:

```bash
./deploy.sh [OPTIONS]
  --project-id ID      GCP project ID
  --region REGION      GCP region (default: us-central1)
  --environment ENV    Environment (default: production)
  --dry-run           Show what would be done
  --skip-terraform    Skip Terraform setup
  --skip-build        Skip CLI build
  --skip-functions    Skip Cloud Functions
  --skip-monitoring   Skip monitoring setup
```

### `deploy-local.sh`

Sets up local development environment with GCP emulators:

- Firestore emulator
- Pub/Sub emulator
- Local API server
- Test data initialization

**Usage**:

```bash
./deploy-local.sh [OPTIONS]
  --skip-emulators    Skip GCP emulators
  --skip-build        Skip CLI build
  --skip-services     Skip Docker services
  --api-port PORT     API port (default: 8000)
```

### `rollback.sh`

Emergency rollback procedures:

- Cloud Run revision rollback
- Terraform state restoration
- Function redeployment
- Cache invalidation

**Usage**:

```bash
./rollback.sh [OPTIONS]
  --project-id ID         GCP project ID
  --region REGION         GCP region
  --service NAME          Service name
  --type TYPE            Rollback type: all|service|terraform|functions
  --revision REV          Specific revision to rollback to
  --skip-verification     Skip rollback verification
```

### `monitoring-setup.sh`

Comprehensive monitoring configuration:

- Alert policies
- Custom dashboards
- SLOs (Service Level Objectives)
- Log exports
- Synthetic monitoring

**Usage**:

```bash
./monitoring-setup.sh [OPTIONS]
  --project-id ID         GCP project ID
  --region REGION         GCP region
  --alert-email EMAIL     Email for alerts
  --slack-webhook URL     Slack webhook URL
```

### `secrets-setup.sh`

Secure credential management:

- Interactive secret configuration
- Secret rotation setup
- Local development export
- Backup procedures

**Usage**:

```bash
./secrets-setup.sh [OPTIONS]
  --project-id ID      GCP project ID
  --non-interactive    Don't prompt for values
  --rotate            Rotate existing secrets
  --export-local      Export for local development
  --backup            Backup secret metadata
  --verify-only       Only verify secrets
```

## Configuration

### Environment Variables

```bash
# Required
export GCP_PROJECT_ID="your-project-id"
export VIRTUOSO_API_KEY="your-api-key"
export VIRTUOSO_ORG_ID="your-org-id"

# Optional
export GCP_REGION="us-central1"
export ENVIRONMENT="production"
export ALERT_EMAIL="alerts@example.com"
export SLACK_WEBHOOK="https://hooks.slack.com/..."
export GITHUB_OWNER="your-github-username"
```

### Terraform Variables

Edit `terraform/terraform.tfvars`:

```hcl
project_id = "your-project-id"
region = "us-central1"
environment = "production"

# Resource configuration
cloud_run_cpu = 2
cloud_run_memory = "2Gi"
cloud_run_max_instances = 10

# Networking
enable_cdn = true
enable_armor = true
```

## Cost Optimization

### Estimated Monthly Costs

| Service         | Usage              | Estimated Cost     |
| --------------- | ------------------ | ------------------ |
| Cloud Run       | 1M requests/month  | $50-100            |
| Firestore       | 10GB storage       | $20-30             |
| Cloud Functions | 100K invocations   | $10-20             |
| Monitoring      | Standard tier      | $20-50             |
| **Total**       | **Moderate usage** | **$100-200/month** |

### Cost Optimization Tips

1. **Cloud Run**:

   - Set appropriate max instances
   - Use minimum instances = 0 for dev/staging
   - Configure concurrency limits

2. **Storage**:

   - Enable lifecycle policies
   - Use nearline storage for backups
   - Clean up old artifacts

3. **Monitoring**:
   - Use log sampling for high-volume logs
   - Set appropriate retention periods
   - Disable unused metrics

## Security Best Practices

1. **Authentication**:

   - All secrets in Secret Manager
   - Service accounts with minimal permissions
   - API key rotation enabled

2. **Network Security**:

   - Cloud Armor DDoS protection
   - VPC with private subnets
   - Cloud NAT for egress

3. **Data Protection**:
   - Encryption at rest
   - Encrypted backups
   - Audit logging enabled

## Troubleshooting

### Common Issues

1. **Authentication Errors**:

   ```bash
   # Check active account
   gcloud auth list

   # Re-authenticate
   gcloud auth login
   ```

2. **API Not Enabled**:

   ```bash
   # Enable missing API
   gcloud services enable SERVICE_NAME.googleapis.com
   ```

3. **Insufficient Permissions**:
   ```bash
   # Grant required role
   gcloud projects add-iam-policy-binding PROJECT_ID \
     --member="user:EMAIL" \
     --role="roles/ROLE_NAME"
   ```

### Debug Commands

```bash
# View Cloud Run logs
gcloud run logs read --service virtuoso-api-cli --region us-central1

# Check service status
gcloud run services describe virtuoso-api-cli --region us-central1

# List recent errors
gcloud logging read "severity>=ERROR" --limit 50

# View metrics
gcloud monitoring dashboards list
```

## Maintenance

### Regular Tasks

1. **Weekly**:

   - Review monitoring alerts
   - Check error logs
   - Verify backups

2. **Monthly**:

   - Review costs
   - Update dependencies
   - Rotate secrets

3. **Quarterly**:
   - Security audit
   - Performance review
   - Disaster recovery test

### Backup and Recovery

```bash
# Backup Firestore
gcloud firestore export gs://PROJECT-backups/firestore/$(date +%Y%m%d)

# Backup Terraform state
gsutil cp gs://PROJECT-terraform-state/terraform.tfstate \
  gs://PROJECT-backups/terraform/terraform.tfstate.$(date +%Y%m%d)

# Restore from backup
./rollback.sh --type all
```

## CI/CD Integration

The deployment automatically sets up Cloud Build triggers for:

1. **Main Branch**: Auto-deploy on push
2. **Pull Requests**: Build and test
3. **Tags**: Deploy releases

### Manual Deployment

```bash
# Deploy from specific commit
git checkout COMMIT_SHA
./deploy.sh

# Deploy specific service only
gcloud run deploy virtuoso-api-cli --source .
```

## Support

For issues or questions:

1. Check the [troubleshooting guide](#troubleshooting)
2. Review Cloud Console logs and metrics
3. Consult GCP documentation
4. Open an issue in the repository

## License

This deployment configuration is part of the Virtuoso API CLI project and follows the same license terms.
