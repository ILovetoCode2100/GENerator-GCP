# GCP Integration for Virtuoso API

This module provides comprehensive Google Cloud Platform integration for the Virtuoso API, enabling enterprise-grade features for scalability, monitoring, and reliability.

## Services Integrated

### 1. Firestore (`firestore_client.py`)

- **Session Management**: Replace Redis with a managed NoSQL solution
- **API Key Storage**: Secure storage and validation of API keys
- **Command History**: Track all executed commands with metadata
- **Caching Layer**: Distributed cache with TTL support

### 2. Cloud Tasks (`cloud_tasks_client.py`)

- **Async Command Execution**: Queue commands for background processing
- **Batch Processing**: Execute multiple commands in parallel
- **Long-running Test Suites**: Handle tests that exceed HTTP timeout limits
- **Retry Configuration**: Automatic retry with exponential backoff

### 3. Pub/Sub (`pubsub_client.py`)

- **Event Publishing**: Publish command execution and test completion events
- **Webhook Support**: Trigger external webhooks on specific events
- **Dead Letter Queue**: Handle failed message processing
- **Event Subscriptions**: Subscribe to events for real-time updates

### 4. Secret Manager (`secret_manager_client.py`)

- **Credential Management**: Securely store Virtuoso API credentials
- **API Key Rotation**: Support for automatic key rotation
- **Secret Caching**: Cache secrets with configurable TTL
- **Multi-environment Support**: Different secrets per environment

### 5. Cloud Storage (`cloud_storage_client.py`)

- **Test Results**: Store detailed test execution results
- **Command Logs**: Archive command execution logs
- **Static Assets**: Serve test reports and documentation
- **Lifecycle Management**: Automatic archival and deletion policies

### 6. Cloud Operations (`monitoring_client.py`)

- **Custom Metrics**: Track command execution metrics
- **Distributed Tracing**: Trace requests across services
- **Error Reporting**: Automatic error reporting with context
- **Performance Profiling**: Monitor API performance

## Configuration

### Environment Variables

```bash
# GCP Project Settings
GCP_PROJECT_ID=your-project-id
GCP_LOCATION=us-central1
GCP_SERVICE_ACCOUNT_EMAIL=service-account@project.iam.gserviceaccount.com

# Enable Services (set to true to activate)
USE_FIRESTORE=true
USE_CLOUD_TASKS=true
USE_PUBSUB=true
USE_SECRET_MANAGER=true
USE_CLOUD_STORAGE=true
USE_CLOUD_MONITORING=true

# Local Development (optional)
FIRESTORE_EMULATOR_HOST=localhost:8080
PUBSUB_EMULATOR_HOST=localhost:8085
```

## Local Development

### Using Emulators

1. Install GCP emulators:

```bash
gcloud components install cloud-firestore-emulator
gcloud components install pubsub-emulator
```

2. Start emulators:

```bash
# Firestore
gcloud emulators firestore start --host-port=localhost:8080

# Pub/Sub
gcloud emulators pubsub start --host-port=localhost:8085
```

3. Set environment variables:

```bash
export FIRESTORE_EMULATOR_HOST=localhost:8080
export PUBSUB_EMULATOR_HOST=localhost:8085
```

## Usage Examples

### Session Management with Firestore

```python
from app.gcp import get_firestore_client

firestore = get_firestore_client()

# Create session
session = await firestore.create_session(
    session_id="unique-id",
    user_id="user-123",
    checkpoint_id="cp_456"
)

# Update session
await firestore.update_session(
    session_id="unique-id",
    checkpoint_id="cp_789"
)

# Get session
session = await firestore.get_session("unique-id")
```

### Async Command Execution

```python
from app.gcp import get_cloud_tasks_client

tasks = get_cloud_tasks_client()

# Queue a command
task = await tasks.create_command_task(
    command="step-navigate",
    args=["to", "https://example.com"],
    checkpoint_id="cp_123",
    user_id="user-456"
)

# Create batch task
batch_task = await tasks.create_batch_task(
    commands=[
        {"command": "step-navigate", "args": ["to", "https://example.com"]},
        {"command": "step-interact", "args": ["click", "button.submit"]},
        {"command": "step-assert", "args": ["exists", "Success message"]}
    ],
    batch_id="batch-001",
    parallel=True
)
```

### Event Publishing

```python
from app.gcp import get_pubsub_client
from app.gcp.pubsub_client import EventType

pubsub = get_pubsub_client()

# Publish event
await pubsub.publish_event(
    EventType.COMMAND_EXECUTED,
    {
        "command": "step-navigate",
        "checkpoint_id": "cp_123",
        "duration_ms": 1234,
        "success": True
    }
)

# Subscribe to events
def handle_command_event(event_data):
    print(f"Command executed: {event_data}")

pubsub.register_handler(EventType.COMMAND_EXECUTED, handle_command_event)
await pubsub.subscribe("command-events-sub")
```

### Secret Management

```python
from app.gcp import get_secret_manager_client

secrets = get_secret_manager_client()

# Load Virtuoso credentials
creds = await secrets.load_virtuoso_credentials()

# Store API key
await secrets.create_secret(
    SecretType.VIRTUOSO_API_KEY,
    {"api_key": "secret-key", "org_id": "1234"}
)

# Rotate secret
await secrets.rotate_secret(
    SecretType.VIRTUOSO_API_KEY,
    {"api_key": "new-secret-key", "org_id": "1234"}
)
```

### Cloud Storage

```python
from app.gcp import get_cloud_storage_client

storage = get_cloud_storage_client()

# Store test results
url = await storage.store_test_result(
    test_id="test-123",
    result_data={
        "status": "passed",
        "duration": 5432,
        "steps": [...]
    }
)

# Generate signed URL
signed_url = await storage.generate_signed_url(
    bucket_name="test-results",
    blob_name="test-123/result.json",
    expiration_minutes=60
)
```

### Monitoring and Tracing

```python
from app.gcp import get_monitoring_client

monitoring = get_monitoring_client()

# Track command execution
await monitoring.track_command_execution(
    command="step-navigate",
    duration_ms=1234,
    success=True,
    user_id="user-123"
)

# Use distributed tracing
async with monitoring.trace_span("process_test_suite") as span:
    span.set_attribute("test.suite.id", "suite-123")
    # Process test suite
    await process_suite()
```

## Best Practices

1. **Use Feature Flags**: Enable only the services you need
2. **Handle Degradation**: Code should work even if GCP services are unavailable
3. **Use Emulators**: Test locally with emulators before deploying
4. **Monitor Costs**: Set up budget alerts for GCP services
5. **Implement Caching**: Use Firestore cache to reduce API calls
6. **Set TTLs**: Configure appropriate TTLs for sessions and cache
7. **Use Batch Operations**: Batch Firestore operations when possible
8. **Enable Tracing**: Use distributed tracing in production
9. **Set Up Alerts**: Configure monitoring alerts for critical metrics
10. **Rotate Secrets**: Implement regular secret rotation

## Error Handling

All GCP clients implement:

- Exponential backoff retry logic
- Graceful degradation when services are unavailable
- Comprehensive error logging
- Timeout handling

## Security Considerations

1. **Service Account**: Use dedicated service accounts with minimal permissions
2. **API Keys**: Store in Secret Manager, never in code
3. **Encryption**: All data is encrypted at rest and in transit
4. **Access Control**: Use IAM roles for fine-grained access control
5. **Audit Logging**: Enable audit logs for all GCP services

## Monitoring

Key metrics tracked:

- Command execution count and duration
- API request latency
- Error rates by type
- Session activity
- Cache hit/miss rates
- Queue depths (Cloud Tasks)
- Message processing rates (Pub/Sub)

## Cost Optimization

1. **Firestore**: Use batch operations and efficient queries
2. **Cloud Tasks**: Set appropriate retry limits
3. **Pub/Sub**: Configure message retention policies
4. **Storage**: Set lifecycle rules for automatic archival
5. **Monitoring**: Sample traces in high-traffic scenarios
