# Virtuoso API CLI - GCP Cloud Architecture

## Executive Summary

This document outlines a comprehensive Google Cloud Platform (GCP) architecture for the Virtuoso API CLI, designed to maximize the use of managed services while minimizing operational overhead. The architecture emphasizes serverless computing, event-driven patterns, and cost optimization through intelligent use of GCP's free tiers and auto-scaling capabilities.

## Architecture Overview

### Core Principles

1. **Serverless First**: Prioritize serverless services (Cloud Run, Cloud Functions) to eliminate infrastructure management
2. **Managed Services**: Use fully managed services for databases, caching, and messaging
3. **Event-Driven**: Implement asynchronous patterns for scalability and resilience
4. **Cost Optimization**: Leverage free tiers and pay-per-use pricing models
5. **Security by Design**: Implement defense-in-depth with managed security services

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Internet                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        Cloud CDN (Global Caching)                        │
│                    • Static content caching                              │
│                    • API response caching                                │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Cloud Load Balancer (Global)                          │
│                    • SSL termination                                     │
│                    • Path-based routing                                  │
│                    • DDoS protection                                     │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    ▼                               ▼
┌─────────────────────────────────┐ ┌─────────────────────────────────┐
│      Cloud Run (FastAPI)        │ │    Cloud Functions              │
│  • Main API service             │ │  • Health checks                │
│  • Auto-scaling (0-1000)        │ │  • Webhook handlers             │
│  • Request handling             │ │  • Lightweight operations       │
└─────────────────────────────────┘ └─────────────────────────────────┘
            │                                   │
            ├───────────────────────────────────┤
            ▼                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Service Layer                                    │
├─────────────────────┬───────────────────────┬───────────────────────────┤
│   Firestore         │  Memorystore          │  Cloud Tasks              │
│ • Session state     │ • High-speed cache    │ • Async execution         │
│ • Test metadata     │ • Command results     │ • Retry logic             │
│ • User preferences  │ • API responses       │ • Rate limiting           │
└─────────────────────┴───────────────────────┴───────────────────────────┘
            │                                   │
            └───────────────┬───────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Pub/Sub                                         │
│                    • Event streaming                                     │
│                    • Service decoupling                                  │
│                    • Fan-out messaging                                   │
└─────────────────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│ Cloud Storage │ │Secret Manager │ │Identity       │
│ • Logs        │ │ • API keys    │ │Platform       │
│ • Artifacts   │ │ • Credentials │ │ • Auth        │
│ • Backups     │ │ • Certs       │ │ • API keys    │
└───────────────┘ └───────────────┘ └───────────────┘
```

## Component Details

### 1. Cloud Run (Main FastAPI Service)

**Purpose**: Host the primary Virtuoso API CLI service with automatic scaling and zero infrastructure management.

**Configuration**:

```yaml
service: virtuoso-api-cli
region: us-central1
platform: managed
spec:
  containers:
    - image: gcr.io/PROJECT_ID/virtuoso-api-cli:latest
      resources:
        limits:
          cpu: "2"
          memory: "4Gi"
      env:
        - name: FIRESTORE_PROJECT
          value: PROJECT_ID
        - name: MEMORYSTORE_HOST
          valueFrom:
            secretKeyRef:
              name: memorystore-host
              key: latest
  traffic:
    - percent: 100
      latestRevision: true
  scaling:
    minInstances: 0
    maxInstances: 1000
    concurrency: 100
```

**Key Features**:

- Auto-scales from 0 to 1000 instances
- Pay only for actual usage
- Built-in HTTPS and load balancing
- Integrated with GCP services via service accounts

### 2. Firestore (Persistent State Management)

**Purpose**: NoSQL database for session management, test metadata, and user preferences.

**Collections Structure**:

```
firestore/
├── sessions/
│   ├── {session_id}/
│   │   ├── checkpoint_id: string
│   │   ├── created_at: timestamp
│   │   ├── updated_at: timestamp
│   │   ├── user_id: string
│   │   └── metadata: map
├── test_runs/
│   ├── {run_id}/
│   │   ├── status: string
│   │   ├── steps: array
│   │   ├── results: map
│   │   └── timestamps: map
├── user_preferences/
│   └── {user_id}/
│       ├── default_output: string
│       ├── api_key_hash: string
│       └── settings: map
└── command_history/
    └── {user_id}/
        └── {command_id}/
            ├── command: string
            ├── timestamp: timestamp
            └── result: map
```

**Benefits**:

- Real-time synchronization
- Strong consistency for single documents
- Automatic scaling
- 1GB free tier per day

### 3. Memorystore (Redis Protocol Cache)

**Purpose**: High-performance caching layer for API responses and frequently accessed data.

**Configuration**:

```yaml
instance:
  name: virtuoso-cache
  tier: STANDARD_HA
  memorySizeGb: 1
  region: us-central1
  redisVersion: REDIS_6_X
  authEnabled: true
  transitEncryptionMode: TLS
```

**Cache Strategy**:

```python
# Cache keys structure
cache_keys = {
    "session:{session_id}": "Session data (TTL: 1 hour)",
    "api:response:{endpoint}:{params_hash}": "API responses (TTL: 5 minutes)",
    "command:result:{command_id}": "Command results (TTL: 30 minutes)",
    "rate_limit:{user_id}:{endpoint}": "Rate limiting counters (TTL: 1 minute)"
}
```

### 4. Cloud Tasks (Async Command Execution)

**Purpose**: Queue and execute long-running commands asynchronously with built-in retry logic.

**Queue Configuration**:

```yaml
queues:
  - name: command-execution
    rateLimits:
      maxDispatchesPerSecond: 100
      maxConcurrentDispatches: 1000
    retryConfig:
      maxAttempts: 5
      maxBackoff: 3600s
      minBackoff: 10s
      maxDoublings: 5

  - name: test-execution
    rateLimits:
      maxDispatchesPerSecond: 50
      maxConcurrentDispatches: 500
    retryConfig:
      maxAttempts: 3
      maxBackoff: 1800s
```

**Task Examples**:

```python
# Create test execution task
task = {
    "http_request": {
        "http_method": "POST",
        "url": f"https://api-cli-{project_id}.run.app/execute-test",
        "headers": {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {service_token}"
        },
        "body": json.dumps({
            "test_id": test_id,
            "checkpoint_id": checkpoint_id,
            "steps": steps
        })
    }
}
```

### 5. Pub/Sub (Event-Driven Architecture)

**Purpose**: Decouple services and enable event-driven patterns for scalability.

**Topics and Subscriptions**:

```yaml
topics:
  - name: command-events
    subscriptions:
      - name: command-processor
        pushEndpoint: https://command-processor.run.app
      - name: command-logger
        pushEndpoint: https://logger.run.app

  - name: test-events
    subscriptions:
      - name: test-executor
        pushEndpoint: https://test-executor.run.app
      - name: test-analyzer
        pushEndpoint: https://analyzer.run.app

  - name: system-events
    subscriptions:
      - name: monitoring
        pushEndpoint: https://monitoring.run.app
      - name: alerting
        pushEndpoint: https://alerting.run.app
```

**Event Schema**:

```json
{
  "eventId": "uuid",
  "eventType": "command.executed",
  "timestamp": "2025-01-24T10:00:00Z",
  "data": {
    "commandId": "string",
    "userId": "string",
    "result": "object"
  },
  "metadata": {
    "source": "api-cli",
    "version": "1.0.0"
  }
}
```

### 6. Cloud Functions (Lightweight Operations)

**Purpose**: Handle lightweight, event-driven operations without the overhead of a full service.

**Functions**:

1. **Health Check Function**

```python
import functions_framework
from google.cloud import firestore

@functions_framework.http
def health_check(request):
    """Lightweight health check for all services"""
    try:
        # Check Firestore
        db = firestore.Client()
        db.collection('health').document('check').set({'timestamp': firestore.SERVER_TIMESTAMP})

        # Check other services
        return {
            'status': 'healthy',
            'services': {
                'firestore': 'ok',
                'memorystore': 'ok',
                'cloud_run': 'ok'
            }
        }, 200
    except Exception as e:
        return {'status': 'unhealthy', 'error': str(e)}, 503
```

2. **Webhook Handler**

```python
@functions_framework.http
def webhook_handler(request):
    """Handle incoming webhooks from Virtuoso"""
    data = request.get_json()

    # Publish to Pub/Sub for processing
    publisher = pubsub.PublisherClient()
    topic = f"projects/{PROJECT_ID}/topics/webhook-events"

    publisher.publish(topic, json.dumps(data).encode())
    return {'status': 'accepted'}, 202
```

### 7. Secret Manager (Credentials Management)

**Purpose**: Securely store and manage sensitive configuration data.

**Secrets Structure**:

```yaml
secrets:
  - name: virtuoso-api-key
    replication:
      automatic: {}
    versions:
      - enabled: true

  - name: memorystore-host
    replication:
      automatic: {}

  - name: service-account-key
    replication:
      automatic: {}

  - name: tls-certificates
    replication:
      userManaged:
        replicas:
          - location: us-central1
          - location: us-east1
```

### 8. Cloud Build (CI/CD Pipeline)

**Purpose**: Automated build and deployment pipeline for continuous delivery.

**cloudbuild.yaml**:

```yaml
steps:
  # Run tests
  - name: "golang:1.21"
    entrypoint: "go"
    args: ["test", "./..."]

  # Build container
  - name: "gcr.io/cloud-builders/docker"
    args:
      ["build", "-t", "gcr.io/$PROJECT_ID/virtuoso-api-cli:$COMMIT_SHA", "."]

  # Push to registry
  - name: "gcr.io/cloud-builders/docker"
    args: ["push", "gcr.io/$PROJECT_ID/virtuoso-api-cli:$COMMIT_SHA"]

  # Deploy to Cloud Run
  - name: "gcr.io/cloud-builders/gcloud"
    args:
      - "run"
      - "deploy"
      - "virtuoso-api-cli"
      - "--image=gcr.io/$PROJECT_ID/virtuoso-api-cli:$COMMIT_SHA"
      - "--region=us-central1"
      - "--platform=managed"

  # Update traffic
  - name: "gcr.io/cloud-builders/gcloud"
    args:
      - "run"
      - "services"
      - "update-traffic"
      - "virtuoso-api-cli"
      - "--to-latest"
      - "--region=us-central1"

timeout: 1200s
```

### 9. Cloud CDN (Content Delivery)

**Purpose**: Cache API responses globally for improved performance.

**Configuration**:

```yaml
backend:
  type: CLOUD_RUN
  service: virtuoso-api-cli

cachePolicy:
  defaultTtl: 300
  maxTtl: 3600
  clientTtl: 300
  negativeCaching: true
  negativeCachingPolicy:
    - code: 404
      ttl: 120
    - code: 500
      ttl: 10

cacheKeyPolicy:
  includeHost: true
  includeProtocol: true
  includeQueryString: true
  queryStringWhitelist:
    - session_id
    - checkpoint_id
    - format
```

### 10. Cloud Load Balancing

**Purpose**: Distribute traffic and provide global access points.

**Configuration**:

```yaml
loadBalancer:
  name: virtuoso-api-lb
  type: HTTPS

  frontendConfig:
    ipAddress: EPHEMERAL
    port: 443
    protocol: HTTPS
    certificate: virtuoso-api-cert

  backendConfig:
    - name: api-backend
      type: CLOUD_RUN
      service: virtuoso-api-cli

  urlMap:
    defaultService: api-backend
    pathMatchers:
      - name: api
        defaultService: api-backend
        pathRules:
          - paths: ["/api/*"]
            service: api-backend
          - paths: ["/health"]
            service: health-check-function
```

### 11. Cloud Storage (Logs and Artifacts)

**Purpose**: Store logs, test artifacts, and backups.

**Bucket Structure**:

```
gs://virtuoso-api-storage/
├── logs/
│   ├── api/
│   │   └── {date}/
│   │       └── {hour}/
│   ├── functions/
│   └── system/
├── artifacts/
│   ├── test-results/
│   │   └── {test_id}/
│   └── reports/
│       └── {date}/
└── backups/
    ├── firestore/
    │   └── {date}/
    └── configs/
        └── {date}/
```

**Lifecycle Policies**:

```yaml
lifecycle:
  rules:
    - action:
        type: Delete
      condition:
        age: 30
        matchesPrefix:
          - logs/

    - action:
        type: SetStorageClass
        storageClass: NEARLINE
      condition:
        age: 7
        matchesPrefix:
          - artifacts/

    - action:
        type: SetStorageClass
        storageClass: COLDLINE
      condition:
        age: 90
        matchesPrefix:
          - backups/
```

### 12. Identity Platform (API Key Management)

**Purpose**: Manage API keys and user authentication.

**Configuration**:

```yaml
identityPlatform:
  apiKeys:
    restrictions:
      - type: SERVER
        allowedIps:
          - 0.0.0.0/0  # For CLI usage

  quotas:
    - name: requests-per-minute
      limit: 1000
      - name: requests-per-day
      limit: 100000

  monitoring:
    - metric: api_key_usage
      threshold: 80
      action: alert
```

## Security Architecture

### Defense in Depth

1. **Network Security**

   - Cloud Armor DDoS protection
   - VPC Service Controls for internal services
   - Private Google Access for GCP API calls

2. **Identity and Access**

   - Service accounts with minimal permissions
   - Workload Identity for Kubernetes workloads
   - API key rotation via Identity Platform

3. **Data Protection**

   - Encryption at rest (default for all services)
   - Encryption in transit (TLS 1.3)
   - Customer-managed encryption keys (CMEK) option

4. **Compliance**
   - Cloud Security Command Center monitoring
   - Asset inventory tracking
   - Vulnerability scanning

### IAM Roles

```yaml
serviceAccounts:
  - name: api-cli-runner
    roles:
      - roles/run.invoker
      - roles/datastore.user
      - roles/redis.editor
      - roles/cloudtasks.enqueuer
      - roles/pubsub.publisher

  - name: function-executor
    roles:
      - roles/cloudfunctions.invoker
      - roles/datastore.viewer
      - roles/secretmanager.secretAccessor

  - name: ci-cd-builder
    roles:
      - roles/cloudbuild.builds.builder
      - roles/run.developer
      - roles/storage.objectAdmin
```

## Monitoring and Observability

### Cloud Monitoring

**Metrics**:

```yaml
customMetrics:
  - name: command_execution_time
    type: DISTRIBUTION
    unit: ms

  - name: api_request_count
    type: CUMULATIVE
    unit: 1

  - name: cache_hit_rate
    type: GAUGE
    unit: ratio
```

**Dashboards**:

- Service Health Overview
- API Performance Metrics
- Cost Analysis Dashboard
- Security Monitoring

### Cloud Logging

**Log Routing**:

```yaml
sinks:
  - name: security-logs
    destination: bigquery.security_dataset
    filter: 'severity >= WARNING AND resource.type="cloud_run_revision"'

  - name: api-analytics
    destination: storage.api-logs-bucket
    filter: 'jsonPayload.endpoint =~ "^/api/"'
```

### Cloud Trace

**Distributed Tracing**:

- Automatic trace collection for Cloud Run
- Custom spans for command execution
- Integration with OpenTelemetry

## Disaster Recovery

### Backup Strategy

1. **Firestore**: Daily automated backups to Cloud Storage
2. **Configuration**: Version controlled in Cloud Source Repositories
3. **Secrets**: Backed up to separate project with cross-region replication

### Recovery Procedures

**RTO (Recovery Time Objective)**: < 30 minutes
**RPO (Recovery Point Objective)**: < 1 hour

**Failover Process**:

1. Traffic automatically routes to healthy regions
2. Cloud Run scales up in alternate regions
3. Firestore multi-region replication ensures data availability

## Development Workflow

### Local Development

```bash
# Install Cloud SDK
gcloud auth application-default login

# Set up local environment
export GOOGLE_CLOUD_PROJECT=virtuoso-dev
export FIRESTORE_EMULATOR_HOST=localhost:8080

# Run Firestore emulator
gcloud emulators firestore start

# Run application locally
go run cmd/api-cli/main.go
```

### Deployment Pipeline

```mermaid
graph LR
    A[Code Push] --> B[Cloud Build Trigger]
    B --> C[Run Tests]
    C --> D[Build Container]
    D --> E[Push to Registry]
    E --> F[Deploy to Staging]
    F --> G[Run Integration Tests]
    G --> H[Deploy to Production]
    H --> I[Update Traffic Split]
```

## Performance Optimization

### Caching Strategy

1. **CDN Level**: Cache static content and common API responses
2. **Memorystore**: Cache session data and frequently accessed records
3. **Application Level**: In-memory caching for hot paths

### Auto-scaling Configuration

```yaml
scaling:
  cloud_run:
    minInstances: 0
    maxInstances: 1000
    targetCPUUtilization: 80
    targetConcurrentRequests: 100

  cloud_functions:
    minInstances: 0
    maxInstances: 1000

  memorystore:
    maxmemoryPolicy: allkeys-lru
    maxmemoryGB: 1
```

## Migration Strategy

### Phase 1: Core Services (Week 1-2)

1. Deploy FastAPI service to Cloud Run
2. Set up Firestore collections
3. Configure Cloud Load Balancer

### Phase 2: Async Processing (Week 3-4)

1. Implement Cloud Tasks queues
2. Set up Pub/Sub topics
3. Deploy Cloud Functions

### Phase 3: Optimization (Week 5-6)

1. Configure Cloud CDN
2. Set up Memorystore
3. Implement monitoring

### Phase 4: Production Readiness (Week 7-8)

1. Security hardening
2. Performance testing
3. Documentation and training

## Conclusion

This architecture provides a robust, scalable, and cost-effective platform for the Virtuoso API CLI. By leveraging GCP's managed services, we eliminate operational overhead while maintaining high availability and performance. The serverless approach ensures we only pay for actual usage, making this solution ideal for both development and production workloads.
