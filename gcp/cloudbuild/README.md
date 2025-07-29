# Cloud Build CI/CD Pipeline for Virtuoso API CLI

This directory contains the complete CI/CD pipeline configuration for deploying the Virtuoso API CLI to Google Cloud Platform using Cloud Build.

## Overview

The pipeline provides:

- Automated builds and tests on every commit
- Preview deployments for pull requests
- Blue-green deployments with gradual traffic shifting
- Infrastructure as Code with Terraform
- Comprehensive monitoring and rollback capabilities

## Directory Structure

```
cloudbuild/
├── cloudbuild.yaml           # Main production deployment pipeline
├── cloudbuild-pr.yaml        # Pull request preview builds
├── cloudbuild-terraform.yaml # Infrastructure deployment
├── buildtriggers.tf         # Terraform config for build triggers
├── substitutions/           # Environment-specific variables
│   ├── dev.yaml
│   ├── staging.yaml
│   └── prod.yaml
├── scripts/                 # Helper scripts
│   ├── test.sh             # Comprehensive test runner
│   ├── deploy.sh           # Blue-green deployment script
│   ├── rollback.sh         # Emergency rollback script
│   └── smoke-test.sh       # Post-deployment validation
└── README.md               # This file
```

## Prerequisites

1. **Google Cloud Project** with the following APIs enabled:

   - Cloud Build API
   - Cloud Run API
   - Artifact Registry API
   - Cloud Functions API
   - Cloud Scheduler API

2. **GitHub Repository** connected to Cloud Build

3. **Service Account** with appropriate permissions

4. **Artifact Registry** repository created:
   ```bash
   gcloud artifacts repositories create virtuoso-artifacts \
     --repository-format=docker \
     --location=us-central1
   ```

## Setup Instructions

### 1. Initial Setup

```bash
# Set your project ID
export PROJECT_ID=your-project-id

# Enable required APIs
gcloud services enable \
  cloudbuild.googleapis.com \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  cloudfunctions.googleapis.com \
  cloudscheduler.googleapis.com

# Create service account for Cloud Build
gcloud iam service-accounts create virtuoso-cloudbuild \
  --display-name="Virtuoso Cloud Build"

# Grant necessary permissions
for role in \
  roles/cloudbuild.builds.builder \
  roles/run.admin \
  roles/storage.admin \
  roles/artifactregistry.admin \
  roles/cloudfunctions.admin; do

  gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:virtuoso-cloudbuild@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="${role}"
done
```

### 2. Connect GitHub Repository

1. Go to Cloud Build triggers page in Console
2. Click "Connect Repository"
3. Select GitHub and authenticate
4. Choose your repository
5. Note the connection name for Terraform

### 3. Create Build Triggers

Apply the Terraform configuration:

```bash
cd gcp/cloudbuild
terraform init
terraform apply -var="project_id=${PROJECT_ID}" -var="github_owner=your-github-username"
```

### 4. Configure Secrets

Store sensitive values in Secret Manager:

```bash
# API key for Virtuoso
echo -n "your-virtuoso-api-key" | gcloud secrets create virtuoso-api-key --data-file=-

# Slack webhook (optional)
echo -n "your-slack-webhook" | gcloud secrets create slack-webhook --data-file=-
```

## Usage

### Manual Trigger

```bash
# Trigger main deployment
gcloud builds submit --config=gcp/cloudbuild/cloudbuild.yaml \
  --substitutions=_ENVIRONMENT=prod

# Trigger PR build
gcloud builds submit --config=gcp/cloudbuild/cloudbuild-pr.yaml \
  --substitutions=_PR_NUMBER=123

# Trigger Terraform plan
gcloud builds submit --config=gcp/cloudbuild/cloudbuild-terraform.yaml \
  --substitutions=_ACTION=plan,_ENVIRONMENT=dev
```

### Automatic Triggers

The following triggers are configured:

1. **Main branch** → Production deployment
2. **Pull requests** → Preview deployments
3. **Tags (v\*)** → Release deployments
4. **staging branch** → Staging deployment
5. **develop branch** → Development deployment

## Pipeline Features

### 1. Build Caching

- Go module cache is preserved between builds
- Docker layer caching for faster builds
- Build artifacts stored in GCS

### 2. Blue-Green Deployments

- New versions deployed with 0% traffic
- Gradual traffic shifting (5% → 10% → 25% → 50% → 75% → 100%)
- Automatic rollback on error rate threshold
- Smoke tests at each stage

### 3. Testing Strategy

- Unit tests with coverage requirements
- Integration tests against real API
- Security scanning with gosec
- Performance testing with k6
- Smoke tests after deployment

### 4. Monitoring

- Build notifications to Slack/Teams
- Deployment metrics in Cloud Monitoring
- Error rate monitoring during rollout
- Automatic incident reports on rollback

## Environment Configuration

### Development

- Minimal resources (256Mi RAM, 1 CPU)
- No approval required
- Debug mode enabled
- 7-day log retention

### Staging

- Medium resources (512Mi RAM, 1 CPU)
- Performance testing enabled
- 30-day log retention
- Blue-green deployment

### Production

- High resources (1Gi RAM, 2 CPUs)
- Manual approval required
- Multi-region deployment
- 90-day log retention
- Full monitoring suite

## Customization

### Adding New Substitutions

Edit the appropriate file in `substitutions/`:

```yaml
substitutions:
  _MY_CUSTOM_VAR: "value"
```

### Modifying Build Steps

Edit the relevant Cloud Build YAML file and add your step:

```yaml
- name: "gcr.io/cloud-builders/gcloud"
  id: "my-custom-step"
  args: ["your", "command"]
```

### Adding New Environments

1. Create new substitution file: `substitutions/new-env.yaml`
2. Add new build trigger in `buildtriggers.tf`
3. Create environment-specific patches

## Troubleshooting

### Build Failures

```bash
# View build logs
gcloud builds log BUILD_ID

# List recent builds
gcloud builds list --limit=10

# View detailed build info
gcloud builds describe BUILD_ID
```

### Deployment Issues

```bash
# Check Cloud Run service
gcloud run services describe virtuoso-api-cli --region=us-central1

# View service logs
gcloud run services logs read virtuoso-api-cli --region=us-central1

# Force rollback
./scripts/rollback.sh
```

### Common Issues

1. **Permission Denied**: Check service account permissions
2. **Build Timeout**: Increase timeout in cloudbuild.yaml
3. **Image Not Found**: Verify Artifact Registry configuration
4. **Traffic Shift Failed**: Check service health before shifting

## Best Practices

1. **Always test in staging** before production
2. **Monitor error rates** during deployments
3. **Keep build steps idempotent**
4. **Use substitutions** for environment-specific values
5. **Tag releases** for easy rollback
6. **Document changes** in commit messages

## Security Considerations

1. **Never commit secrets** - use Secret Manager
2. **Restrict service account** permissions
3. **Enable vulnerability scanning** on images
4. **Use least privilege** for all resources
5. **Audit build logs** regularly

## Support

For issues or questions:

1. Check build logs first
2. Review this README
3. Check GCP documentation
4. Contact the DevOps team

## Next Steps

1. Review and customize the configurations for your needs
2. Set up monitoring dashboards
3. Configure alerting policies
4. Plan your branching strategy
5. Train team on the deployment process
