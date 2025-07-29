# 📋 Currently Deployed Endpoints

## API URL

`https://virtuoso-api-936111683985.us-central1.run.app`

## Available Endpoints (70 Total)

### Core Health & Monitoring (6 endpoints)

#### 1. **GET /**

- **Description**: Root endpoint with service information
- **Response**:
  ```json
  {
    "service": "Virtuoso CLI",
    "version": "1.0.0",
    "description": "RESTful API for Virtuoso CLI",
    "documentation": {
      "swagger": "/docs",
      "redoc": "/redoc",
      "openapi": "/openapi.json"
    }
  }
  ```
- **Authentication**: None required

#### 2. **GET /health**

- **Description**: Comprehensive health check endpoint
- **Response**:
  ```json
  {
    "healthy": true,
    "api_version": "1.0.0",
    "environment": "production",
    "timestamp": "2024-07-24T22:00:00Z"
  }
  ```
- **Authentication**: None required

#### 3. **GET /health/ready**

- **Description**: Readiness check endpoint
- **Response**:
  ```json
  {
    "ready": true,
    "checks": {
      "cli": true,
      "redis": true,
      "config": true
    }
  }
  ```
- **Authentication**: None required

#### 4. **GET /health/live**

- **Description**: Liveness check endpoint
- **Response**:
  ```json
  {
    "alive": true,
    "uptime_seconds": 3600,
    "pid": 1234
  }
  ```
- **Authentication**: None required

#### 5. **GET /health/metrics**

- **Description**: Prometheus metrics endpoint
- **Response**: Prometheus text format metrics
- **Authentication**: None required

#### 6. **GET /health/stats**

- **Description**: Comprehensive monitoring statistics
- **Response**:
  ```json
  {
    "system": {
      "cpu_usage_percent": 23.5,
      "memory": {...},
      "disk": {...}
    },
    "monitoring": {
      "active_tasks": 5,
      "health_history": {...}
    }
  }
  ```
- **Authentication**: None required

### Command Execution (12 endpoints)

#### 7. **GET /api/v1/commands**

- **Description**: List available command groups
- **Response**:
  ```json
  {
    "commands": [
      "step-assert",
      "step-interact",
      "step-navigate",
      "step-wait",
      "step-data",
      "step-window",
      "step-dialog",
      "step-misc",
      "run-test"
    ],
    "total": 9
  }
  ```
- **Authentication**: Optional API key

#### 8. **GET /api/v1/commands/list**

- **Description**: List all available CLI commands with permissions
- **Response**: Array of command details
- **Authentication**: Required

#### 9. **GET /api/v1/commands/{command}/help**

- **Description**: Get detailed help for a specific command
- **Response**: Command help information
- **Authentication**: Required

#### 10. **POST /api/v1/commands/step/{command_group}/{subcommand}**

- **Description**: Execute a single step command
- **Request Body**:
  ```json
  {
    "checkpoint_id": "cp_123",
    "args": ["arg1", "arg2"],
    "position": 1,
    "description": "Click button"
  }
  ```
- **Response**: Step execution result
- **Authentication**: Required

#### 11. **POST /api/v1/commands/batch**

- **Description**: Execute multiple commands in batch
- **Request Body**: Array of commands
- **Response**: Batch execution result
- **Authentication**: Required

#### 12. **POST /api/v1/commands/execute-async**

- **Description**: Execute command asynchronously using Cloud Tasks
- **Response**: Task ID and status
- **Authentication**: Required

#### 13. **POST /api/v1/commands/step-assert/{subcommand}**

- **Description**: Execute assertion commands
- **Authentication**: Required

#### 14. **POST /api/v1/commands/step-interact/{subcommand}**

- **Description**: Execute interaction commands
- **Authentication**: Required

#### 15. **POST /api/v1/commands/step-navigate/{subcommand}**

- **Description**: Execute navigation commands
- **Authentication**: Required

#### 16. **POST /api/v1/commands/step-wait/{subcommand}**

- **Description**: Execute wait commands
- **Authentication**: Required

#### 17. **POST /api/v1/commands/step-data/{subcommand}**

- **Description**: Execute data commands
- **Authentication**: Required

#### 18. **POST /api/v1/commands/run-test**

- **Description**: Execute test run command
- **Authentication**: Required

### Test Management (12 endpoints)

#### 19. **POST /api/v1/tests/run**

- **Description**: Run a test from definition or ID
- **Request Body**: Test definition
- **Response**: Test run result
- **Authentication**: Required

#### 20. **POST /api/v1/tests/upload**

- **Description**: Upload and run a test file (YAML/JSON)
- **Request**: Multipart file upload
- **Response**: Test run result
- **Authentication**: Required

#### 21. **GET /api/v1/tests/templates**

- **Description**: List available test templates
- **Response**: Array of test templates
- **Authentication**: Required

#### 22. **GET /api/v1/tests/{test_id}**

- **Description**: Get test details by ID
- **Response**: Test information
- **Authentication**: Required

#### 23. **GET /api/v1/tests/{test_id}/status**

- **Description**: Get test execution status
- **Response**: Status information
- **Authentication**: Required

#### 24. **GET /api/v1/tests/{test_id}/results**

- **Description**: Get test execution results
- **Response**: Detailed results
- **Authentication**: Required

#### 25. **DELETE /api/v1/tests/{test_id}**

- **Description**: Delete a test
- **Response**: 204 No Content
- **Authentication**: Required

#### 26. **GET /api/v1/tests/history**

- **Description**: Get test execution history
- **Response**: Array of past test runs
- **Authentication**: Required

#### 27. **POST /api/v1/tests/{test_id}/retry**

- **Description**: Retry a failed test
- **Response**: New test run result
- **Authentication**: Required

#### 28. **GET /api/v1/tests/{test_id}/logs**

- **Description**: Get test execution logs
- **Response**: Log entries
- **Authentication**: Required

#### 29. **POST /api/v1/tests/validate**

- **Description**: Validate test YAML/JSON without running
- **Response**: Validation result
- **Authentication**: Required

#### 30. **GET /api/v1/tests/templates/{template_id}**

- **Description**: Get specific test template
- **Response**: Template details
- **Authentication**: Required

### Session Management (10 endpoints)

#### 31. **GET /api/v1/sessions**

- **Description**: List all sessions for current user
- **Response**: Array of sessions
- **Authentication**: Required

#### 32. **POST /api/v1/sessions**

- **Description**: Create a new session
- **Request Body**: Session details
- **Response**: Created session
- **Authentication**: Required

#### 33. **GET /api/v1/sessions/{session_id}**

- **Description**: Get specific session details
- **Response**: Session information
- **Authentication**: Required

#### 34. **PATCH /api/v1/sessions/{session_id}**

- **Description**: Update session information
- **Request Body**: Update fields
- **Response**: Updated session
- **Authentication**: Required

#### 35. **DELETE /api/v1/sessions/{session_id}**

- **Description**: Delete a session
- **Response**: 204 No Content
- **Authentication**: Required

#### 36. **POST /api/v1/sessions/{session_id}/activate**

- **Description**: Activate a session (set as current)
- **Response**: Activation result
- **Authentication**: Required

#### 37. **GET /api/v1/sessions/current**

- **Description**: Get current active session
- **Response**: Current session or null
- **Authentication**: Required

#### 38. **GET /api/v1/sessions/{session_id}/analytics**

- **Description**: Get analytics for a session
- **Response**: Session analytics data
- **Authentication**: Required

#### 39. **POST /api/v1/sessions/{session_id}/checkpoint**

- **Description**: Create checkpoint in session
- **Response**: Checkpoint details
- **Authentication**: Required

#### 40. **GET /api/v1/sessions/{session_id}/checkpoints**

- **Description**: List session checkpoints
- **Response**: Array of checkpoints
- **Authentication**: Required

### Webhook Management (10 endpoints)

#### 41. **GET /api/v1/webhooks**

- **Description**: List all webhooks
- **Response**: Array of webhooks
- **Authentication**: Required

#### 42. **POST /api/v1/webhooks**

- **Description**: Create a new webhook
- **Request Body**: Webhook configuration
- **Response**: Created webhook
- **Authentication**: Required

#### 43. **GET /api/v1/webhooks/{webhook_id}**

- **Description**: Get specific webhook
- **Response**: Webhook details
- **Authentication**: Required

#### 44. **PATCH /api/v1/webhooks/{webhook_id}**

- **Description**: Update webhook configuration
- **Request Body**: Update fields
- **Response**: Updated webhook
- **Authentication**: Required

#### 45. **DELETE /api/v1/webhooks/{webhook_id}**

- **Description**: Delete a webhook
- **Response**: 204 No Content
- **Authentication**: Required

#### 46. **POST /api/v1/webhooks/{webhook_id}/test**

- **Description**: Test webhook with sample payload
- **Response**: Test result
- **Authentication**: Required

#### 47. **GET /api/v1/webhooks/{webhook_id}/deliveries**

- **Description**: Get webhook delivery history
- **Response**: Array of deliveries
- **Authentication**: Required

#### 48. **POST /api/v1/webhooks/pubsub/receive**

- **Description**: Internal endpoint for Pub/Sub push
- **Response**: Acknowledgment
- **Authentication**: Internal only

#### 49. **POST /api/v1/webhooks/{webhook_id}/enable**

- **Description**: Enable a disabled webhook
- **Response**: Updated webhook
- **Authentication**: Required

#### 50. **POST /api/v1/webhooks/{webhook_id}/disable**

- **Description**: Disable an active webhook
- **Response**: Updated webhook
- **Authentication**: Required

### Analytics & Reporting (8 endpoints)

#### 51. **POST /api/v1/analytics/query**

- **Description**: Query analytics data
- **Request Body**: Query parameters
- **Response**: Analytics report
- **Authentication**: Required

#### 52. **GET /api/v1/analytics/dashboard**

- **Description**: Get dashboard statistics
- **Response**: Dashboard data
- **Authentication**: Required

#### 53. **POST /api/v1/analytics/report/generate**

- **Description**: Generate analytics report
- **Request Body**: Report parameters
- **Response**: Report ID and status
- **Authentication**: Required

#### 54. **GET /api/v1/analytics/report/{report_id}/status**

- **Description**: Get report generation status
- **Response**: Status information
- **Authentication**: Required

#### 55. **GET /api/v1/analytics/report/{report_id}/download**

- **Description**: Download generated report
- **Response**: Report file
- **Authentication**: Required

#### 56. **GET /api/v1/analytics/usage**

- **Description**: Get usage statistics
- **Response**: Usage data
- **Authentication**: Required

#### 57. **GET /api/v1/analytics/metrics**

- **Description**: Get custom metrics
- **Response**: Metrics data
- **Authentication**: Required

#### 58. **POST /api/v1/analytics/export**

- **Description**: Export analytics data
- **Response**: Export job ID
- **Authentication**: Required

### Project Management (6 endpoints)

#### 59. **GET /api/v1/projects**

- **Description**: List projects
- **Response**: Array of projects
- **Authentication**: Required

#### 60. **POST /api/v1/projects**

- **Description**: Create new project
- **Request Body**: Project details
- **Response**: Created project
- **Authentication**: Required

#### 61. **GET /api/v1/projects/{project_id}**

- **Description**: Get project details
- **Response**: Project information
- **Authentication**: Required

#### 62. **PATCH /api/v1/projects/{project_id}**

- **Description**: Update project
- **Request Body**: Update fields
- **Response**: Updated project
- **Authentication**: Required

#### 63. **DELETE /api/v1/projects/{project_id}**

- **Description**: Delete project
- **Response**: 204 No Content
- **Authentication**: Required

#### 64. **GET /api/v1/projects/{project_id}/tests**

- **Description**: Get project tests
- **Response**: Array of tests
- **Authentication**: Required

### Documentation (3 endpoints)

#### 65. **GET /docs**

- **Description**: Swagger UI documentation
- **Response**: HTML page
- **Authentication**: None required

#### 66. **GET /redoc**

- **Description**: ReDoc documentation
- **Response**: HTML page
- **Authentication**: None required

#### 67. **GET /openapi.json**

- **Description**: OpenAPI schema
- **Response**: JSON schema
- **Authentication**: None required

### Authentication & User Management (3 endpoints)

#### 68. **POST /api/v1/auth/token**

- **Description**: Generate API token
- **Request Body**: Credentials
- **Response**: Access token
- **Authentication**: Basic auth

#### 69. **POST /api/v1/auth/refresh**

- **Description**: Refresh API token
- **Request Body**: Refresh token
- **Response**: New access token
- **Authentication**: Refresh token

#### 70. **GET /api/v1/auth/validate**

- **Description**: Validate current token
- **Response**: Token validation result
- **Authentication**: Required

## 🔐 Authentication Details

### API Key Authentication

- Header: `X-API-Key: your-api-key`
- Environment variable: `VIRTUOSO_API_KEY`

### Bearer Token Authentication

- Header: `Authorization: Bearer your-token`
- Token expiry: 24 hours

### Rate Limiting

- Default: 200 requests per minute per user
- Batch operations: 10 requests per minute
- Async operations: 50 requests per minute

## 🚀 Deployment Status

### Currently Deployed

- Full FastAPI application with all 70 endpoints
- GCP Cloud Run service: `virtuoso-api`
- Region: `us-central1`
- Automatic scaling: 0-100 instances
- Memory: 512MB per instance
- CPU: 1 vCPU per instance

### GCP Services Integration

- **Cloud Storage**: Test files and reports
- **Firestore**: Session management and caching
- **Pub/Sub**: Event streaming and webhooks
- **Cloud Tasks**: Async command execution
- **BigQuery**: Analytics data warehouse
- **Cloud Monitoring**: Metrics and logging
- **Secret Manager**: API keys and credentials

## 📊 Monitoring & Observability

### Health Check Endpoints

- `/health` - Overall health
- `/health/ready` - Readiness probe
- `/health/live` - Liveness probe
- `/health/metrics` - Prometheus metrics

### Logging

- Structured JSON logging
- Request/response logging
- Error tracking with stack traces
- Performance metrics

### Metrics

- Request count by endpoint
- Response time percentiles
- Error rates by type
- Active connections
- Command execution statistics

## 🔄 Version Information

- **API Version**: 1.0.0
- **Deployed**: July 24, 2024
- **Last Updated**: July 24, 2024
- **Revision**: Latest
