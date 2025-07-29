# Virtuoso CLI Deployment & FastAPI Wrapper Implementation Plan

## Objective

Deploy the Virtuoso API CLI as a production-ready service with a FastAPI wrapper, enabling HTTP-based access to all CLI functionality.

## Implementation Prompt

Please help me implement a production deployment of the Virtuoso API CLI with a FastAPI wrapper. Here's what needs to be done:

### Phase 1: Prepare CLI for Deployment

1. **Create Release Build**

   - Add a `make release` target that builds optimized binaries for multiple platforms
   - Include version information in the binary (use git tags/commits)
   - Strip debug symbols for smaller binary size
   - Create checksums for each platform build

2. **Package Configuration**

   - Create a production configuration template
   - Add environment variable support for all config values
   - Implement secure credential handling (no hardcoded tokens)
   - Add health check command: `api-cli health-check`

3. **Containerization**
   - Create a multi-stage Dockerfile that:
     - Builds the Go binary in first stage
     - Creates minimal runtime image with just the binary
     - Includes ca-certificates for HTTPS
     - Sets up non-root user for security
   - Add docker-compose.yml for local development

### Phase 2: FastAPI Wrapper Implementation

Create a FastAPI service (`app/main.py`) that wraps the CLI with these features:

1. **Core Structure**

   ```python
   # Key endpoints needed:
   POST /test/run - Run a test from YAML/JSON
   POST /commands/{command_group}/{subcommand} - Execute any CLI command
   GET /commands - List all available commands
   GET /health - Health check endpoint
   GET /sessions/{session_id} - Get session info
   POST /sessions - Create new session
   ```

2. **Command Execution Service**

   - Subprocess management with proper timeout handling
   - Stream command output for long-running operations
   - Parse CLI output based on requested format (json/yaml/human)
   - Handle concurrent command execution safely

3. **Request/Response Models**

   - Pydantic models for all command inputs
   - Standardized error responses
   - Support for async/sync execution modes
   - WebSocket support for real-time command output

4. **Authentication & Security**

   - API key authentication
   - Rate limiting per client
   - Request validation and sanitization
   - Audit logging for all commands

5. **Session Management**
   - Create/manage Virtuoso sessions via API
   - Session pooling for performance
   - Automatic session cleanup

### Phase 3: Deployment Infrastructure

1. **Kubernetes Deployment**

   ```yaml
   # Create manifests for:
   - Deployment with HPA (Horizontal Pod Autoscaler)
   - Service with LoadBalancer/Ingress
   - ConfigMap for CLI configuration
   - Secret for API credentials
   - PersistentVolume for test artifacts
   ```

2. **Monitoring & Observability**

   - Prometheus metrics (command execution time, success rate)
   - Structured logging with correlation IDs
   - Distributed tracing for command flow
   - Grafana dashboards

3. **CI/CD Pipeline**
   - GitHub Actions workflow for:
     - Running tests on PR
     - Building and pushing Docker images
     - Deploying to staging/production
     - Running integration tests post-deploy

### Phase 4: Additional Features

1. **Queue System**

   - Add Redis/RabbitMQ for async job processing
   - Background job status tracking
   - Webhook notifications on completion

2. **Storage Layer**

   - Store test results and artifacts
   - S3/MinIO integration for large files
   - Test history and analytics

3. **API Documentation**
   - Auto-generated OpenAPI/Swagger docs
   - Interactive API explorer
   - Code examples in multiple languages
   - SDK generation

### Technical Requirements

1. **FastAPI Dependencies**

   ```python
   fastapi==0.104.1
   uvicorn==0.24.0
   pydantic==2.5.0
   python-multipart==0.0.6
   aiofiles==23.2.1
   asyncio==3.4.3
   redis==5.0.1
   prometheus-client==0.19.0
   ```

2. **Directory Structure**

   ```
   virtuoso-GENerator/
   ├── api/
   │   ├── app/
   │   │   ├── __init__.py
   │   │   ├── main.py
   │   │   ├── models/
   │   │   ├── routes/
   │   │   ├── services/
   │   │   └── utils/
   │   ├── requirements.txt
   │   ├── Dockerfile.api
   │   └── tests/
   ├── deployment/
   │   ├── kubernetes/
   │   ├── docker-compose.yml
   │   └── scripts/
   └── .github/
       └── workflows/
           ├── build.yml
           └── deploy.yml
   ```

3. **Environment Variables**

   ```bash
   # API Service
   API_HOST=0.0.0.0
   API_PORT=8000
   API_WORKERS=4
   API_KEY_HEADER=X-API-Key

   # CLI Configuration
   VIRTUOSO_API_TOKEN=<encrypted>
   VIRTUOSO_BASE_URL=https://api-app2.virtuoso.qa/api
   VIRTUOSO_ORG_ID=2242

   # Redis (for queuing)
   REDIS_URL=redis://localhost:6379

   # Monitoring
   ENABLE_METRICS=true
   LOG_LEVEL=INFO
   ```

### Implementation Steps

1. **Week 1**: CLI preparation and containerization
2. **Week 2**: FastAPI wrapper core functionality
3. **Week 3**: Authentication, monitoring, and testing
4. **Week 4**: Kubernetes deployment and CI/CD
5. **Week 5**: Queue system and advanced features
6. **Week 6**: Documentation and production readiness

### Key Considerations

1. **Performance**

   - CLI binary should be cached in memory
   - Use connection pooling for API calls
   - Implement request queuing for rate limiting

2. **Security**

   - Never expose raw CLI credentials
   - Validate and sanitize all inputs
   - Use least-privilege principles

3. **Scalability**

   - Horizontal scaling via Kubernetes
   - Stateless API design
   - External session storage (Redis)

4. **Reliability**
   - Circuit breakers for external calls
   - Retry logic with exponential backoff
   - Graceful degradation

### Success Criteria

- [ ] All 70 CLI commands accessible via REST API
- [ ] 99.9% uptime SLA
- [ ] < 100ms latency for command initiation
- [ ] Comprehensive API documentation
- [ ] Full test coverage (unit + integration)
- [ ] Production monitoring and alerting
- [ ] Automated deployment pipeline

## Next Steps

1. Review and refine this plan
2. Set up the basic FastAPI structure
3. Implement core command execution endpoint
4. Add authentication and error handling
5. Create Docker images and test locally
6. Deploy to staging environment
7. Load test and optimize
8. Deploy to production

This implementation will transform the CLI into a scalable, production-ready API service while maintaining all existing functionality.
