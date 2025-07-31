# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains a multi-deployment Virtuoso API proxy implementation that simplifies and standardizes access to the Virtuoso test automation platform. The project supports three deployment targets:
- **AWS**: CDK-based Lambda + API Gateway deployment
- **GCP**: Cloud Run + Cloud Functions deployment
- **Local/Docker**: FastAPI-based REST service

## Commands

### AWS CDK Deployment

```bash
# Navigate to CDK directory
cd cdk

# Install dependencies
npm install

# Build TypeScript
npm run build

# Deploy stack (requires proper AWS credentials)
npm run deploy

# Other CDK commands
npm run synth    # Synthesize CloudFormation template
npm run diff     # Show deployment differences
npm run destroy  # Remove all resources
npm run test     # Run CDK tests
```

### GCP Deployment

```bash
# Navigate to GCP directory
cd gcp

# Quick deployment (interactive wizard)
./deploy-wizard.sh

# One-click deployment
./one-click-deploy.sh

# Manual deployment steps
./setup-project.sh --project-id PROJECT_ID --create-project
./secrets-setup.sh
./deploy.sh --skip-monitoring

# Local development
./deploy-local.sh

# Rollback if needed
./rollback.sh --type all
```

### Local API Development

```bash
# Navigate to API directory
cd api

# Install Python dependencies
pip install -r requirements.txt

# Run development server
uvicorn app.main:app --reload --port 8000

# Run tests
pytest
pytest --cov=app tests/

# Docker deployment
docker build -f Dockerfile.api -t virtuoso-api .
docker run -p 8000:8000 --env-file .env virtuoso-api
```

### CLI Binary Commands

```bash
# Build CLI binary
make build

# Run quality checks
make check
make lint
make fmt

# Run tests
make test
make test-commands
make test-library

# Integration tests
./test-scripts/test-all-69-commands.sh [checkpoint-id]
./test-all-commands-simple.sh [checkpoint-id]
```

## Architecture

### Repository Structure

```
api-lambdav2/
├── cdk/                    # AWS CDK infrastructure
│   ├── lib/               # CDK stack definitions
│   ├── lambda/            # Lambda function handlers
│   └── scripts/           # Deployment scripts
├── gcp/                    # Google Cloud Platform deployment
│   ├── terraform/         # Infrastructure as code
│   ├── functions/         # Cloud Functions
│   └── cloudbuild/        # CI/CD configurations
├── api/                    # FastAPI REST service
│   ├── app/               # Application code
│   │   ├── routes/        # API endpoints
│   │   ├── services/      # Business logic
│   │   └── middleware/    # Request processing
│   └── tests/             # API tests
├── pkg/api-cli/           # Go CLI implementation
│   ├── client/            # Virtuoso API client
│   ├── commands/          # CLI commands
│   └── yaml-layer/        # YAML test parser
└── examples/              # Usage examples
```

### Deployment Architecture

#### AWS Architecture
- **API Gateway (HTTP API)**: Cost-efficient routing layer at ~$1/million requests
- **Lambda Functions**: One function per endpoint, Node.js 20.x on ARM64
- **Secrets Manager**: Secure storage for Virtuoso API keys
- **CloudWatch**: Logging and monitoring
- **Custom Authorizer**: Bearer token validation

#### GCP Architecture
- **Cloud Run**: Main API service with auto-scaling
- **Cloud Functions**: Background tasks and webhooks
- **Firestore**: Session and test data storage
- **Secret Manager**: Credential management
- **Cloud Build**: CI/CD pipeline
- **Monitoring**: Integrated logging and metrics

#### Local/Docker Architecture
- **FastAPI**: High-performance Python web framework
- **Uvicorn**: ASGI server
- **Session Management**: In-memory or Redis-backed
- **Rate Limiting**: Configurable per-endpoint limits

### Key Design Patterns

1. **Proxy Pattern**: All deployments act as a simplified proxy to the complex Virtuoso API
2. **Request Simplification**: Strip unnecessary fields, provide defaults
3. **Response Minimization**: Return only essential data
4. **Unified Error Handling**: Consistent error format across deployments
5. **Session Context**: Maintain state across multiple API calls

## API Simplification Strategy

### Example: Goal Execution Endpoint

**Original Virtuoso API Request**:
```json
{
  "goalId": "123",
  "startingUrl": "https://example.com",
  "includeDataDrivenJourneys": true,
  "includeDisabledJourneys": false,
  "parallelExecution": true,
  "maxParallelExecutions": 5,
  "environment": "production",
  "initialData": {...},
  "headers": {...},
  "cookies": [...]
}
```

**Simplified API Request**:
```json
{
  "startingUrl": "https://example.com"  // Optional, all other fields use defaults
}
```

**Simplified Response**:
```json
{
  "jobId": "job123",
  "status": "started"
}
```

## Configuration

### AWS Configuration
```bash
# Required IAM permissions (see iam-policy-virtuoso-cdk.json)
- CloudFormation full access
- Lambda management
- API Gateway v2 management
- IAM role creation
- S3 for CDK assets
- Secrets Manager access
```

### GCP Configuration
```bash
# Environment variables
export GCP_PROJECT_ID="your-project-id"
export VIRTUOSO_API_KEY="your-api-key"
export VIRTUOSO_ORG_ID="your-org-id"
export GCP_REGION="us-central1"
```

### Local API Configuration
```yaml
# virtuoso-config.yaml
api:
  auth_token: your-api-key-here
  base_url: https://api-app2.virtuoso.qa/api
organization:
  id: "2242"
```

## Common Development Tasks

### Adding a New Endpoint

1. **AWS Lambda Handler**:
   ```typescript
   // Copy pattern from cdk/lambda/handlers/execute-goal.ts
   // Implement request simplification
   // Add to endpoint list in virtuoso-api-stack.ts
   ```

2. **GCP Function**:
   ```go
   // Add handler in functions/
   // Update terraform configuration
   // Deploy with ./deploy.sh
   ```

3. **FastAPI Route**:
   ```python
   # Add route in api/app/routes/
   # Implement service in api/app/services/
   # Add tests in api/tests/
   ```

### Testing Changes

```bash
# Test specific endpoint locally
curl -X POST http://localhost:8000/api/v1/commands/execute \
  -H "X-API-Key: your-key" \
  -H "Content-Type: application/json" \
  -d '{"command": "step-navigate", "args": ["to", "https://example.com"]}'

# Run full test suite
./test-scripts/test-all-69-commands.sh

# Test YAML functionality
./bin/api-cli run-test examples/simple-login-test.yaml
```

## Deployment Considerations

### AWS Deployment Issues
- Requires extensive IAM permissions
- CDK bootstrap needed (can fail with limited permissions)
- Use `virtuoso-dev` profile or create new IAM user
- Alternative: Export CDK template and deploy manually

### GCP Deployment (Recommended)
- Simpler permission model
- Built-in monitoring and logging
- Free tier available
- One-click deployment scripts

### Production Checklist
1. Set appropriate API rate limits
2. Configure CORS for your domains
3. Enable monitoring and alerting
4. Rotate API keys regularly
5. Use custom domain with SSL
6. Set up backup procedures
7. Configure auto-scaling limits

## Known Limitations

1. **File Uploads**: Only URL-based uploads supported (no local files)
2. **Browser Navigation**: `back`, `forward`, `refresh` not supported by API
3. **Window Operations**: Close and frame switching by index/name unavailable
4. **CDK Bootstrap**: May fail with insufficient IAM permissions
5. **Rate Limits**: Virtuoso API has undocumented rate limits

## Troubleshooting

### AWS CDK Errors
```bash
# TypeScript compilation errors
npm run build  # Check for syntax errors

# Permission errors
aws sts get-caller-identity  # Verify credentials
# Use admin credentials or fix IAM permissions

# Bootstrap failures
# Delete failed stack in CloudFormation console
# Re-run with proper permissions
```

### GCP Deployment Errors
```bash
# Check prerequisites
./pre-deployment-check.sh

# View logs
gcloud logging read "severity>=ERROR" --limit 50

# Check service status
gcloud run services describe virtuoso-api-cli --region us-central1
```

### API Development Issues
```bash
# Module import errors
pip install -r requirements.txt

# CLI binary not found
export CLI_PATH=/path/to/api-cli
# Or copy binary to expected location

# Session errors
# Check VIRTUOSO_SESSION_ID environment variable
```