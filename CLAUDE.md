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