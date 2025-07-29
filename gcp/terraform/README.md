# Virtuoso API CLI - GCP Terraform Configuration

This directory contains comprehensive Terraform configuration for deploying the Virtuoso API CLI to Google Cloud Platform (GCP).

## Overview

The configuration deploys a fully-managed, serverless architecture leveraging:

- **Cloud Run** for the main API service
- **Firestore** for NoSQL data storage
- **Memorystore (Redis)** for high-performance caching
- **Cloud Functions** for lightweight operations
- **Pub/Sub** for event-driven architecture
- **Cloud Tasks** for async processing
- **Cloud Load Balancing** with CDN for global distribution
- **Cloud Monitoring** for observability

## Architecture

```
Internet → Cloud CDN → Load Balancer → Cloud Run (API)
                                          ↓
                                    VPC Connector
                                          ↓
                          ┌───────────────┼───────────────┐
                          ↓               ↓               ↓
                     Firestore       Memorystore      Cloud Tasks
                          ↓               ↓               ↓
                          └───────────────┴───────────────┘
                                          ↓
                                      Pub/Sub
                                          ↓
                                  Cloud Functions
```

## Prerequisites

1. **GCP Account**: Active GCP account with billing enabled
2. **Terraform**: Version 1.5.0 or higher
3. **gcloud CLI**: Authenticated with appropriate permissions
4. **Required Permissions**:
   - Project Editor or Owner
   - Service Account Admin
   - Security Admin (if using Cloud Armor)

## Quick Start

1. **Clone and navigate to the terraform directory**:

   ```bash
   cd gcp/terraform
   ```

2. **Copy and configure variables**:

   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your values
   ```

3. **Initialize Terraform**:

   ```bash
   terraform init
   ```

4. **Review the plan**:

   ```bash
   terraform plan
   ```

5. **Apply the configuration**:
   ```bash
   terraform apply
   ```

## Configuration Files

### Core Infrastructure

- **`main.tf`** - Provider setup, API enablement, and base resources
- **`services.tf`** - Cloud Run, Firestore, Memorystore, and storage
- **`networking.tf`** - VPC, load balancer, CDN, and network security
- **`security.tf`** - IAM, secrets, identity platform, and security policies

### Async Processing

- **`async.tf`** - Pub/Sub topics, Cloud Tasks queues, Cloud Scheduler jobs
- **`functions.tf`** - Cloud Functions for health checks, webhooks, cleanup, and analytics

### Observability

- **`monitoring.tf`** - Logging, monitoring, alerts, and dashboards

### Configuration

- **`variables.tf`** - Input variable definitions
- **`outputs.tf`** - Output values
- **`terraform.tfvars.example`** - Example configuration

## Key Variables

### Required

- `project_id` - Your GCP project ID
- `virtuoso_api_key` - API key for Virtuoso (stored in Secret Manager)

### Important Optional

- `environment` - Environment name (development/staging/production)
- `region` - GCP region (default: us-central1)
- `api_domains` - Custom domains for HTTPS access
- `cloud_run_min_instances` - Minimum instances (0 for scale-to-zero)
- `cloud_run_max_instances` - Maximum instances for auto-scaling

## Cost Optimization

The configuration is optimized for cost with:

- **Scale-to-zero** Cloud Run instances
- **Basic tier** Redis for development
- **Lifecycle policies** for storage
- **Minimal VPC connector** sizing
- **Pay-per-use** pricing for most services

Estimated monthly costs:

- Development: ~$50-100
- Production (low traffic): ~$200-500
- Production (high traffic): ~$500-2000

## Security Features

- **Secret Manager** for sensitive data
- **Service accounts** with minimal permissions
- **Cloud Armor** DDoS protection (optional)
- **Binary Authorization** for container security (optional)
- **VPC Service Controls** for network isolation (optional)
- **Private Service Connect** for private endpoints (optional)

## Monitoring & Alerts

The configuration includes:

- **Custom dashboards** for service health
- **Alert policies** for errors, latency, and resource usage
- **Log-based metrics** for custom tracking
- **Uptime checks** for availability monitoring
- **SLOs** for service reliability (optional)

## Deployment Environments

### Development

```hcl
environment = "development"
cloud_run_min_instances = 0
redis_tier = "BASIC"
enable_cloud_armor = false
```

### Staging

```hcl
environment = "staging"
cloud_run_min_instances = 1
redis_tier = "BASIC"
enable_cloud_armor = true
```

### Production

```hcl
environment = "production"
cloud_run_min_instances = 2
cloud_run_max_instances = 1000
redis_tier = "STANDARD_HA"
enable_cloud_armor = true
enable_binary_authorization = true
```

## Post-Deployment Steps

1. **Configure DNS**:

   - Point your domain A records to the load balancer IP
   - Wait for SSL certificate provisioning (10-15 minutes)

2. **Set Virtuoso API Key**:

   ```bash
   gcloud secrets versions add virtuoso-api-key --data-file=- < api-key.txt
   ```

3. **Deploy Application**:

   ```bash
   # Build and push container
   docker build -t gcr.io/$PROJECT_ID/virtuoso-api-cli .
   docker push gcr.io/$PROJECT_ID/virtuoso-api-cli

   # Update Cloud Run service
   gcloud run deploy virtuoso-api-cli \
     --image gcr.io/$PROJECT_ID/virtuoso-api-cli \
     --region $REGION
   ```

4. **Configure Monitoring**:
   - Set up notification channels
   - Customize alert thresholds
   - Create custom dashboards

## Maintenance

### Updating the Infrastructure

```bash
terraform plan
terraform apply
```

### Scaling Resources

Edit `terraform.tfvars`:

```hcl
cloud_run_max_instances = 500
redis_memory_gb = 2
```

### Destroying Resources

```bash
terraform destroy
```

## Troubleshooting

### Common Issues

1. **API enablement errors**:

   ```bash
   gcloud services enable compute.googleapis.com
   ```

2. **Permission errors**:

   ```bash
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="user:your-email@example.com" \
     --role="roles/editor"
   ```

3. **SSL certificate pending**:
   - Ensure DNS is properly configured
   - Wait 10-15 minutes for provisioning

### Debug Commands

```bash
# Check Cloud Run logs
gcloud run logs read --service=virtuoso-api-cli

# Check function logs
gcloud functions logs read health-check

# Test health endpoint
curl https://your-domain.com/health
```

## Advanced Features

### Enable Binary Authorization

```hcl
enable_binary_authorization = true
```

### Enable VPC Service Controls

```hcl
enable_vpc_service_controls = true
access_policy_id = "your-policy-id"
```

### Enable Private Service Connect

```hcl
enable_private_service_connect = true
```

## Support

For issues or questions:

1. Check the [architecture documentation](../architecture/ARCHITECTURE.md)
2. Review Cloud Console logs
3. Contact the platform team

## License

This configuration is part of the Virtuoso API CLI project.
