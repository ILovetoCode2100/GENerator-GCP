# Virtuoso API Lambda Generator

Transform Virtuoso API endpoints into AWS Lambda functions with automated deployment using AWS SAM.

## Overview

This project automatically generates and deploys AWS Lambda functions that proxy requests to the Virtuoso API. It groups 29 Virtuoso endpoints into 9 Lambda functions based on resource type, providing a serverless interface to Virtuoso's test automation platform.

## Features

- **Automated Lambda Generation**: Converts Virtuoso API endpoints into Lambda functions
- **Resource-Based Grouping**: 9 Lambda functions handling related endpoints
- **Shared Layer Architecture**: Common utilities packaged as Lambda layer
- **AWS SAM Deployment**: Infrastructure as code using CloudFormation
- **Secure Token Management**: API tokens stored in AWS SSM Parameter Store
- **Built-in Retry Logic**: Automatic retry with exponential backoff
- **Structured Logging**: AWS Lambda Powertools integration

## Prerequisites

- Node.js 18.x or later
- AWS CLI configured with appropriate credentials
- AWS SAM CLI installed
- Access to Virtuoso API with valid API token

## Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd lambda-api-gen

# Generate Lambda functions and SAM template
node generate-lambdas.js

# Deploy to AWS (you'll be prompted for your Virtuoso API token)
./deploy.sh YOUR_VIRTUOSO_API_TOKEN
```

## Architecture

### Lambda Functions

The system organizes Virtuoso API endpoints into 9 Lambda functions:

1. **VirtuosoProjectHandler** - Project management (3 endpoints)
2. **VirtuosoGoalHandler** - Goal operations (3 endpoints)  
3. **VirtuosoJourneyHandler** - Journey/testsuite operations (7 endpoints)
4. **VirtuosoCheckpointHandler** - Checkpoint/testcase operations (5 endpoints)
5. **VirtuosoStepHandler** - Test step operations (5 endpoints)
6. **VirtuosoExecutionHandler** - Test execution (3 endpoints)
7. **VirtuosoLibraryHandler** - Library management (5 endpoints)
8. **VirtuosoDataHandler** - Test data operations (3 endpoints)
9. **VirtuosoEnvironmentHandler** - Environment configuration (1 endpoint)

### Shared Layer

All Lambda functions share a common layer containing:
- **Authentication**: SSM-based token retrieval
- **Error Handling**: Standardized error responses
- **Retry Logic**: Configurable retry with backoff
- **Logging**: Structured logging with AWS Lambda Powertools

## Deployment

### Basic Deployment

```bash
./deploy.sh YOUR_VIRTUOSO_API_TOKEN
```

### Custom Configuration

```bash
# Deploy to specific region
AWS_REGION=eu-west-1 ./deploy.sh YOUR_API_TOKEN

# Deploy with custom stack name
STACK_NAME=virtuoso-prod ./deploy.sh YOUR_API_TOKEN
```

### Update Functions

After making code changes:

```bash
# Regenerate Lambda functions
node generate-lambdas.js

# Update deployed functions
node update-all-lambdas.js
```

## Testing

```bash
# Test Lambda functions via API Gateway
node test-virtuoso-api.js

# Test the holy grail endpoint
node test-holy-grail.js
```

## Configuration

### Environment Variables

- `AWS_REGION` - Target AWS region (default: us-east-1)
- `VIRTUOSO_API_URL` - Base URL for Virtuoso API
- `API_TOKEN_PARAM` - SSM parameter name for API token
- `STACK_NAME` - CloudFormation stack name

### Lambda Configuration

Edit `generate-lambdas.js` to modify:
- Memory allocation (default: 256 MB)
- Timeout (default: 30 seconds)
- Runtime version
- Environment variables

## API Gateway Routes

Each Lambda function is exposed via API Gateway with routes like:

```
POST   /virtuoso/projects
GET    /virtuoso/projects
GET    /virtuoso/projects/{projectId}/goals
POST   /virtuoso/goals
GET    /virtuoso/testsuites/{journeyId}
POST   /virtuoso/checkpoints
POST   /virtuoso/teststeps
...
```

## Security

- API tokens are stored securely in AWS Systems Manager Parameter Store
- Lambda functions use IAM roles with least privilege access
- All API requests require authentication
- Sensitive data is never logged

## Monitoring

- CloudWatch Logs for all Lambda invocations
- X-Ray tracing enabled for performance monitoring
- CloudWatch metrics for invocations, errors, and duration
- Structured logging for easy querying

## Troubleshooting

### Common Issues

1. **403 Forbidden from API Gateway**
   - Check IAM permissions
   - Verify API Gateway deployment
   - Ensure Lambda function has proper execution role

2. **Token Not Found**
   - Verify SSM parameter exists: `/virtuoso/api-token`
   - Check Lambda IAM role has SSM read permissions

3. **High Latency**
   - Cold start on first invocation
   - Consider increasing Lambda memory
   - Enable provisioned concurrency for critical functions

## Project Structure

```
lambda-api-gen/
├── generate-lambdas.js      # Lambda function generator
├── update-all-lambdas.js    # Update deployed functions
├── template.yaml            # SAM template (generated)
├── deploy.sh               # Deployment script
├── lambda-functions/       # Generated Lambda functions
│   ├── project/
│   ├── goal/
│   ├── journey/
│   └── ...
├── lambda-layer/          # Shared utilities
│   └── nodejs/
│       ├── utils/
│       ├── config.js
│       └── package.json
└── test-*.js             # Test scripts
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details