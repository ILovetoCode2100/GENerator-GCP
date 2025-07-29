# 🏗️ GCP Architecture Documentation

## Overview

The Virtuoso API is deployed on Google Cloud Platform (GCP) using a modern, scalable, serverless architecture. This document describes the complete infrastructure setup, service integrations, and deployment pipeline.

## Architecture Diagram

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Client Apps   │────▶│  Load Balancer  │────▶│   Cloud Run     │
└─────────────────┘     └─────────────────┘     └────────┬────────┘
                                                          │
                           ┌──────────────────────────────┴──────────────────────────────┐
                           │                                                             │
                           ▼                                                             ▼
                  ┌─────────────────┐                                          ┌─────────────────┐
                  │   Cloud Tasks   │                                          │    Firestore    │
                  └────────┬────────┘                                          └────────┬────────┘
                           │                                                             │
                           ▼                                                             ▼
                  ┌─────────────────┐     ┌─────────────────┐            ┌─────────────────┐
                  │     Pub/Sub     │────▶│    BigQuery     │            │  Cloud Storage  │
                  └─────────────────┘     └─────────────────┘            └─────────────────┘
                           │
                           ▼
                  ┌─────────────────┐     ┌─────────────────┐
                  │ Cloud Functions │     │ Secret Manager  │
                  └─────────────────┘     └─────────────────┘
```

## Core Services

### 1. Cloud Run (Primary Application Host)

- **Service Name**: `virtuoso-api`
- **Region**: `us-central1`
- **Configuration**:
  ```yaml
  apiVersion: serving.knative.dev/v1
  kind: Service
  metadata:
    name: virtuoso-api
    labels:
      cloud.googleapis.com/location: us-central1
  spec:
    template:
      metadata:
        annotations:
          autoscaling.knative.dev/minScale: "0"
          autoscaling.knative.dev/maxScale: "100"
          run.googleapis.com/execution-environment: gen2
      spec:
        containerConcurrency: 80
        timeoutSeconds: 300
        serviceAccountName: virtuoso-api-sa@virtuoso-generator.iam.gserviceaccount.com
        containers:
          - image: gcr.io/virtuoso-generator/virtuoso-api:latest
            resources:
              limits:
                cpu: "1"
                memory: 512Mi
            env:
              - name: PORT
                value: "8080"
              - name: GCP_PROJECT_ID
                value: virtuoso-generator
  ```

### 2. Firestore (NoSQL Database)

- **Purpose**: Session management, caching, user data
- **Collections**:
  - `sessions` - Active user sessions
  - `checkpoints` - Test checkpoints
  - `webhooks` - Webhook configurations
  - `cache` - Temporary data with TTL
  - `users` - User profiles and settings
- **Indexes**:
  ```
  - sessions: (user_id, created_at DESC)
  - checkpoints: (session_id, position ASC)
  - webhooks: (user_id, active, created_at DESC)
  ```

### 3. Cloud Storage (Object Storage)

- **Buckets**:
  - `virtuoso-tests` - Test files and templates
  - `virtuoso-reports` - Generated reports
  - `virtuoso-artifacts` - Build artifacts
- **Lifecycle Rules**:
  - Test files: Delete after 30 days
  - Reports: Archive to Coldline after 90 days
  - Artifacts: Keep latest 10 versions

### 4. BigQuery (Data Warehouse)

- **Dataset**: `analytics`
- **Tables**:
  - `command_executions` - Command execution logs
  - `test_runs` - Test execution history
  - `user_activity` - User activity tracking
  - `performance_metrics` - API performance data
- **Scheduled Queries**:
  - Daily aggregation of command statistics
  - Weekly performance reports
  - Monthly usage summaries

### 5. Cloud Tasks (Async Processing)

- **Queues**:
  - `default` - General async tasks
  - `command-execution` - Long-running commands
  - `test-suite` - Test suite execution
  - `webhook-delivery` - Webhook deliveries
- **Configuration**:
  ```yaml
  rateLimits:
    maxDispatchesPerSecond: 100
    maxConcurrentDispatches: 50
  retryConfig:
    maxAttempts: 5
    maxBackoff: 600s
    minBackoff: 10s
  ```

### 6. Pub/Sub (Event Streaming)

- **Topics**:
  - `command-events` - Command execution events
  - `test-events` - Test run events
  - `webhook-events` - Webhook trigger events
  - `system-events` - System monitoring events
- **Subscriptions**:
  - Push subscriptions for webhook delivery
  - Pull subscriptions for analytics processing

### 7. Secret Manager (Credentials)

- **Secrets**:
  - `virtuoso-api-key` - Main API key
  - `api-keys` - Additional API keys
  - `database-url` - Database connection string
  - `encryption-key` - Data encryption key
  - `webhook-signing-key` - Webhook signature key

### 8. Cloud Build (CI/CD)

- **Trigger**: Push to main branch
- **Build Configuration**: `cloudbuild.yaml`
- **Steps**:
  1. Build Docker image
  2. Run tests
  3. Push to Container Registry
  4. Deploy to Cloud Run
  5. Run smoke tests

## Security Configuration

### IAM Roles

```yaml
Service Account: virtuoso-api-sa@virtuoso-generator.iam.gserviceaccount.com
Roles:
  - roles/firestore.user
  - roles/storage.objectAdmin (virtuoso-* buckets only)
  - roles/bigquery.dataEditor
  - roles/cloudtasks.enqueuer
  - roles/pubsub.publisher
  - roles/secretmanager.secretAccessor
  - roles/monitoring.metricWriter
  - roles/logging.logWriter
```

### Network Security

- **Ingress**: Allow all traffic (public API)
- **Egress**: VPC connector for internal services
- **SSL**: Managed SSL certificates
- **DDoS Protection**: Cloud Armor policies

### API Security

- **Authentication**: API keys and JWT tokens
- **Rate Limiting**: Per-user and per-IP limits
- **CORS**: Configurable allowed origins
- **Request Validation**: Input sanitization

## Monitoring & Observability

### Cloud Monitoring

- **Metrics**:
  - Request latency (p50, p95, p99)
  - Error rate by endpoint
  - Active connections
  - Memory and CPU usage
- **Dashboards**:
  - API Performance Dashboard
  - Error Analysis Dashboard
  - Usage Analytics Dashboard

### Cloud Logging

- **Log Types**:
  - Application logs (structured JSON)
  - Access logs
  - Error logs with stack traces
  - Audit logs
- **Log Retention**: 30 days
- **Log Sinks**: BigQuery for long-term analysis

### Alerting Policies

```yaml
- name: High Error Rate
  condition: error_rate > 5%
  duration: 5 minutes
  notification: PagerDuty, Email

- name: High Latency
  condition: p95_latency > 1000ms
  duration: 10 minutes
  notification: Email

- name: Cloud Run Instance Down
  condition: instance_count == 0
  duration: 2 minutes
  notification: PagerDuty
```

## Deployment Pipeline

### Continuous Deployment

```yaml
# cloudbuild.yaml
steps:
  # Build the container image
  - name: "gcr.io/cloud-builders/docker"
    args: ["build", "-t", "gcr.io/$PROJECT_ID/virtuoso-api:$COMMIT_SHA", "."]

  # Push to Container Registry
  - name: "gcr.io/cloud-builders/docker"
    args: ["push", "gcr.io/$PROJECT_ID/virtuoso-api:$COMMIT_SHA"]

  # Deploy to Cloud Run
  - name: "gcr.io/google.com/cloudsdktool/cloud-sdk"
    entrypoint: gcloud
    args:
      - "run"
      - "deploy"
      - "virtuoso-api"
      - "--image"
      - "gcr.io/$PROJECT_ID/virtuoso-api:$COMMIT_SHA"
      - "--region"
      - "us-central1"
      - "--platform"
      - "managed"
      - "--allow-unauthenticated"

  # Tag as latest
  - name: "gcr.io/cloud-builders/docker"
    args:
      [
        "tag",
        "gcr.io/$PROJECT_ID/virtuoso-api:$COMMIT_SHA",
        "gcr.io/$PROJECT_ID/virtuoso-api:latest",
      ]

  - name: "gcr.io/cloud-builders/docker"
    args: ["push", "gcr.io/$PROJECT_ID/virtuoso-api:latest"]
```

### Environment Configuration

```bash
# Production Environment Variables
PORT=8080
GCP_PROJECT_ID=virtuoso-generator
ENVIRONMENT=production
USE_CLOUD_MONITORING=true
USE_FIRESTORE=true
USE_BIGQUERY=true
USE_CLOUD_TASKS=true
USE_PUBSUB=true
USE_CLOUD_STORAGE=true
RATE_LIMIT_ENABLED=true
LOG_LEVEL=INFO
```

## Scaling Strategy

### Auto-scaling Configuration

- **Min Instances**: 0 (scale to zero)
- **Max Instances**: 100
- **Target CPU Utilization**: 60%
- **Target Concurrent Requests**: 80
- **Scale Down Delay**: 300 seconds

### Performance Optimization

- **Cold Start Mitigation**:
  - Keep-alive health checks every 5 minutes
  - Minimum instance during business hours
  - Prewarming for predictable traffic
- **Caching Strategy**:
  - Firestore for hot data (< 1MB)
  - Cloud CDN for static assets
  - In-memory cache for frequent queries

## Disaster Recovery

### Backup Strategy

- **Firestore**: Daily automated backups
- **BigQuery**: Table snapshots every 6 hours
- **Cloud Storage**: Cross-region replication
- **Secret Manager**: Version history maintained

### Recovery Procedures

1. **Service Failure**: Automatic failover to healthy instances
2. **Region Failure**: Manual failover to backup region
3. **Data Loss**: Restore from latest backup
4. **Corruption**: Rollback to previous version

## Cost Optimization

### Resource Allocation

- **Cloud Run**: Pay per request, scale to zero
- **Firestore**: Optimize queries, use caching
- **BigQuery**: Partition tables, use clustering
- **Cloud Storage**: Lifecycle policies for archival

### Monitoring Costs

- **Budget Alerts**: $500/month warning, $1000/month critical
- **Cost Breakdown**: Weekly reports by service
- **Optimization**: Monthly review and adjustments

## Migration Notes

### From Previous Architecture

- **Docker Compose**: Replaced with Cloud Run
- **PostgreSQL**: Migrated to Firestore
- **Redis**: Replaced with Firestore caching
- **Nginx**: Replaced with Cloud Load Balancer

### Future Enhancements

1. Multi-region deployment for global availability
2. GraphQL API endpoint
3. WebSocket support for real-time updates
4. ML-powered command suggestions
5. Advanced analytics with Dataflow

## Troubleshooting Guide

### Common Issues

1. **Cold Start Latency**

   - Solution: Increase minimum instances
   - Monitor: Cold start frequency metrics

2. **Rate Limit Errors**

   - Solution: Adjust rate limit configuration
   - Monitor: Rate limit hit metrics

3. **Memory Pressure**

   - Solution: Increase memory allocation
   - Monitor: Memory usage metrics

4. **Timeout Errors**
   - Solution: Increase timeout or use async
   - Monitor: Request duration metrics

### Debug Commands

```bash
# View Cloud Run logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=virtuoso-api" --limit 50

# Check service status
gcloud run services describe virtuoso-api --region us-central1

# View metrics
gcloud monitoring time-series list --filter='metric.type="run.googleapis.com/request_count"'

# Test deployment
curl https://virtuoso-api-936111683985.us-central1.run.app/health
```

## Contact & Support

- **Team**: Virtuoso Platform Team
- **Email**: platform@virtuoso.dev
- **Slack**: #virtuoso-api-support
- **On-Call**: PagerDuty rotation

## References

- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Firestore Best Practices](https://cloud.google.com/firestore/docs/best-practices)
- [BigQuery Optimization](https://cloud.google.com/bigquery/docs/best-practices-performance-overview)
- [GCP Security Best Practices](https://cloud.google.com/security/best-practices)
