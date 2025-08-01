# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains a Lambda API Generator for the Virtuoso API, which converts Virtuoso API endpoints into AWS Lambda functions with automated deployment.

- **Purpose**: Transform Virtuoso API endpoints into grouped AWS Lambda functions
- **Main Components**: Lambda function generator, deployment scripts, test suite
- **Language**: JavaScript (Node.js) for Lambda functions
- **Infrastructure**: AWS Lambda, API Gateway, SSM Parameter Store, CloudFormation/SAM

## Commands

### Build and Generate

```bash
# Generate Lambda functions from Virtuoso API endpoints
node generate-lambdas.js

# Update all Lambda functions with latest code
node update-all-lambdas.js

# Quick start - generates everything
./QUICK_START_LAMBDA.sh
```

### Deploy

```bash
# Deploy to AWS with SAM CLI
./deploy.sh YOUR_VIRTUOSO_API_TOKEN

# Deploy to specific region
AWS_REGION=eu-west-1 ./deploy.sh YOUR_API_TOKEN

# Redeploy Lambda functions (without SAM)
./redeploy-lambdas.sh

# Deploy with custom stack name
STACK_NAME=virtuoso-prod ./deploy.sh YOUR_API_TOKEN

```

### Test

```bash
# Run Lambda API test suite
node test-virtuoso-api.js

# Run holy grail test (GET /testsuites/{journeyId})
node test-holy-grail.js
```

## Architecture

### Lambda API Generator

The system organizes 29 Virtuoso API endpoints into 9 Lambda functions by resource type:

1. **VirtuosoProjectHandler** - Project management (3 endpoints)
2. **VirtuosoGoalHandler** - Goal operations (3 endpoints)  
3. **VirtuosoJourneyHandler** - Journey/testsuite operations (7 endpoints)
4. **VirtuosoCheckpointHandler** - Checkpoint/testcase operations (5 endpoints)
5. **VirtuosoStepHandler** - Test step operations (5 endpoints)
6. **VirtuosoExecutionHandler** - Test execution (3 endpoints)
7. **VirtuosoLibraryHandler** - Library management (5 endpoints)
8. **VirtuosoDataHandler** - Test data operations (3 endpoints)
9. **VirtuosoEnvironmentHandler** - Environment configuration (1 endpoint)

### Shared Layer Architecture

The Lambda functions share a common layer (`/opt/nodejs/`) containing:
- **Authentication**: `utils/auth.js` - Retrieves API token from SSM
- **Error Handling**: `utils/error-handler.js` - Standardized error responses
- **Retry Logic**: `utils/retry.js` - Automatic retry with exponential backoff
- **Logging**: `utils/logger.js` - AWS Lambda Powertools integration
- **Configuration**: `config.js` - Centralized configuration


## Key Implementation Details

### Lambda Function Generation

The `generate-lambdas.js` script:
1. Defines endpoint groups in `endpointGroups` object
2. Generates Lambda functions with API Gateway integration
3. Creates SAM template for deployment
4. Packages shared utilities in Lambda layer

### Authentication Flow

1. API token stored in AWS Systems Manager Parameter Store at `/virtuoso/api-token`
2. Lambda functions retrieve token using `getApiToken()`
3. Token added to Authorization header for Virtuoso API calls


## Important Endpoints

### Holy Grail Endpoint

`GET /testsuites/{journeyId}` - Returns complete test structure including:
- Journey/TestSuite details
- List of checkpoints/test cases
- List of steps within each checkpoint

## Configuration

### Environment Variables

- `VIRTUOSO_API_URL` - Base URL for Virtuoso API
- `API_TOKEN_PARAM` - SSM parameter name
- `VIRTUOSO_SESSION_ID` - Session checkpoint ID
- `AWS_REGION` - AWS region for deployment


## Testing Strategy

1. **Lambda Tests**: `test-virtuoso-api.js`, `test-holy-grail.js`
2. **Integration Tests**: Test scripts verify Lambda function deployment and API Gateway routes

## Known Issues

1. **API Gateway Routes**: Some routes may return 403 if not properly configured
2. **Cold Start Latency**: Initial Lambda invocations may have higher latency

## Update: 2025-07-31 19:03

### Changes Summary
- Major refactoring to focus exclusively on AWS deployment
- Removed all GCP, Kubernetes, Docker, and FastAPI components
- Removed Go CLI and all multi-cloud deployment features
- Streamlined project to AWS Lambda and SAM deployment only

### Modified Components
- **Features**: AWS-only deployment focus
- **Fixes**: Removed multi-cloud complexity
- **Dependencies**: Reduced to AWS SDK and Node.js only
- **Configuration**: Simplified to AWS-specific settings

### Security Considerations
- **CRITICAL**: Hardcoded API token found in `test-virtuoso-api.js` (line 8)
- This token should be moved to environment variables or AWS SSM Parameter Store
- Never commit credentials to version control

### Performance Impact
- Reduced build times by removing unnecessary components
- Smaller deployment package size
- Faster CI/CD pipeline execution

### Notes for Claude Code
- Project is now AWS-focused only - do not add multi-cloud features
- Use AWS SSM Parameter Store for all secrets management
- Follow AWS Lambda best practices for all new functions
- Ensure all API tokens are retrieved from SSM, never hardcoded

## Update: 2025-08-01 06:30

### Changes Summary
- Created comprehensive platform-agnostic API layer in `src/api/`
- Prepared project for integration with virtuoso-GENerator-bedrock
- Maintained backward compatibility while enabling forward flexibility
- Added multi-tenant support and abstraction layers

### Modified Components
- **Features**: Platform-agnostic API layer with 9 handlers for 29 endpoints
- **Architecture**: Created abstraction interfaces for runtime, secrets, logging, configuration
- **Dependencies**: Added abstractions for AWS Lambda, Bedrock, Express, and generic platforms
- **Configuration**: Multi-tenant configuration management system

### New Directory Structure
```
src/api/
├── abstractions/         # Platform abstraction layer
├── config/              # Configuration and tenant management
├── core/handlers/       # API endpoint handlers
├── utils/               # Shared utilities
├── examples/            # Integration examples
└── scripts/             # Build and validation scripts
```

### Integration Documentation
- **BEDROCK_INTEGRATION_ARCHITECTURE.md**: Complete integration design
- **PORTING_GUIDE.md**: Step-by-step integration instructions
- **BEDROCK_PORTING_SUMMARY.md**: Summary of changes and benefits

### Security Considerations
- Removed hardcoded API token from test-virtuoso-api.js
- All secrets now retrieved from SSM Parameter Store
- Platform-agnostic secret management through abstraction layer

### Performance Impact
- Modular architecture enables selective loading
- Abstraction layer adds minimal overhead
- Retry logic and error handling built into base handler

### Notes for Claude Code
- API layer can be used standalone or integrated with bedrock project
- Use `createApiService()` as main entry point
- Platform detection is automatic but can be overridden
- Tenant configuration enables multi-customer deployments
- All handlers extend BaseHandler for consistent behavior