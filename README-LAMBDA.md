# Virtuoso API Lambda Functions

## Overview

This project contains AWS Lambda functions that wrap the Virtuoso API endpoints, organized by resource type for optimal performance and maintainability.

## Architecture

- **Shared Layer**: Common utilities, authentication, error handling, and retry logic
- **Lambda Functions**: Grouped by resource type (projects, goals, journeys, etc.)
- **API Gateway**: RESTful API interface with path-based routing

## Deployment

### Prerequisites

- AWS CLI configured with appropriate credentials
- AWS SAM CLI installed
- Node.js 18.x or later

### Quick Deploy

```bash
# Deploy with your Virtuoso API token
./deploy.sh YOUR_VIRTUOSO_API_TOKEN

# Or with custom settings
STACK_NAME=my-virtuoso-api AWS_REGION=eu-west-1 ./deploy.sh YOUR_API_TOKEN
```

### Manual Deployment

1. Install dependencies:
   ```bash
   cd lambda-layer/nodejs && npm install
   ```

2. Deploy using SAM:
   ```bash
   sam deploy --guided --parameter-overrides ApiTokenValue=YOUR_TOKEN
   ```

## Usage

Each Lambda function handles multiple related endpoints. Call them with an `action` parameter:

```javascript
// Example: Create a project
const response = await lambda.invoke({
  FunctionName: 'VirtuosoProjectHandler',
  Payload: JSON.stringify({
    action: 'createProject',
    body: {
      name: 'My Test Project',
      description: 'Automated testing project'
    }
  })
}).promise();

// Example: Execute a goal
const response = await lambda.invoke({
  FunctionName: 'VirtuosoExecutionHandler',
  Payload: JSON.stringify({
    action: 'executeGoal',
    body: {
      goalId: 123,
      environment: 'staging'
    }
  })
}).promise();
```

## Function Groups


### VirtuosoProjectHandler
Endpoints:
- POST /projects (action: `createProject`)
- GET /projects (action: `listProjects`)
- GET /projects/{projectId}/goals (action: `listProjectGoals`)

### VirtuosoGoalHandler
Endpoints:
- POST /goals (action: `createGoal`)
- GET /goals/{goalId}/versions (action: `getGoalVersions`)
- POST /goals/{goalId}/snapshots/{snapshotId}/execute (action: `executeGoalSnapshot`)

### VirtuosoJourneyHandler
Endpoints:
- POST /testsuites (action: `createJourney`)
- POST /journeys (action: `createJourneyAlt`)
- GET /testsuites/latest_status (action: `listJourneysWithStatus`)
- GET /testsuites/{journeyId} (action: `getJourneyDetails`)
- PUT /testsuites/{journeyId} (action: `updateJourney`)
- POST /testsuites/{journeyId}/checkpoints/attach (action: `attachCheckpoint`)
- POST /journeys/attach-library (action: `attachLibraryCheckpoint`)

### VirtuosoCheckpointHandler
Endpoints:
- POST /testcases (action: `createCheckpoint`)
- POST /checkpoints (action: `createCheckpointAlt`)
- GET /testcases/{checkpointId} (action: `getCheckpointDetails`)
- GET /checkpoints/{checkpointId}/teststeps (action: `getCheckpointSteps`)
- POST /testcases/{checkpointId}/add-to-library (action: `addCheckpointToLibrary`)

### VirtuosoStepHandler
Endpoints:
- POST /teststeps (action: `addTestStep`)
- POST /teststeps?envelope=false (action: `addTestStepNoEnvelope`)
- POST /steps (action: `addTestStepAlt`)
- GET /teststeps/{stepId} (action: `getStepDetails`)
- PUT /teststeps/{stepId}/properties (action: `updateStepProperties`)

### VirtuosoExecutionHandler
Endpoints:
- POST /executions (action: `executeGoal`)
- GET /executions/{executionId} (action: `getExecutionStatus`)
- GET /executions/analysis/{executionId} (action: `getExecutionAnalysis`)

### VirtuosoLibraryHandler
Endpoints:
- POST /library/checkpoints (action: `addToLibrary`)
- GET /library/checkpoints/{libraryCheckpointId} (action: `getLibraryCheckpoint`)
- PUT /library/checkpoints/{libraryCheckpointId} (action: `updateLibraryCheckpoint`)
- DELETE /library/checkpoints/{libraryCheckpointId}/steps/{testStepId} (action: `removeLibraryStep`)
- POST /library/checkpoints/{libraryCheckpointId}/steps/{testStepId}/move (action: `moveLibraryStep`)

### VirtuosoDataHandler
Endpoints:
- POST /testdata/tables/create (action: `createDataTable`)
- GET /testdata/tables/{tableId} (action: `getDataTable`)
- POST /testdata/tables/{tableId}/import (action: `importDataToTable`)

### VirtuosoEnvironmentHandler
Endpoints:
- POST /environments (action: `createEnvironment`)

## Environment Variables

- `VIRTUOSO_API_URL`: Base URL for Virtuoso API
- `API_TOKEN_PARAM`: SSM parameter name for API token
- `LOG_LEVEL`: Logging level (DEBUG, INFO, WARN, ERROR)

## Monitoring

CloudWatch Logs and X-Ray tracing are automatically configured for all functions.

## Cost Optimization

- Functions are grouped to minimize cold starts
- Shared layer reduces deployment size
- Memory allocated based on typical workload

## Security

- API token stored in AWS Systems Manager Parameter Store
- IAM roles follow least privilege principle
- All functions use VPC endpoints when configured
