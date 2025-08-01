# Virtuoso API Layer

A platform-agnostic API management layer for Virtuoso test automation, designed to work seamlessly across AWS Lambda, Bedrock AI agents, and other platforms.

## Features

- **Platform Agnostic**: Works with AWS Lambda, Express, Fastify, Koa, and custom platforms
- **Multi-tenant Support**: Built-in tenant configuration management
- **Complete API Coverage**: All 29 Virtuoso API endpoints organized into 9 logical handlers
- **Abstraction Layer**: Clean separation of platform-specific code from business logic
- **Retry Logic**: Automatic retry with exponential backoff
- **Error Handling**: Standardized error responses across platforms

## Installation

```bash
npm install @virtuoso/api-layer
```

## Quick Start

### Basic Usage

```javascript
const { createApiService } = require('@virtuoso/api-layer');

// Create API service
const apiService = createApiService({
  platform: 'generic',
  tenantId: 'default'
});

// List projects
const projects = await apiService.projects.listProjects();

// Get journey details (Holy Grail endpoint)
const journey = await apiService.journeys.getJourneyDetails({
  params: { journeyId: '12345' }
});
```

### AWS Lambda Integration

```javascript
const { createApiService } = require('@virtuoso/api-layer');

exports.handler = async (event, context) => {
  const apiService = createApiService({
    platform: 'aws-lambda',
    tenantId: event.headers?.['x-tenant-id'] || 'default'
  });

  return apiService.createPlatformHandler('lambda')(event, context);
};
```

### Bedrock AI Integration

```javascript
const { BedrockApiIntegration } = require('@virtuoso/api-layer/examples/bedrock-integration');

const integration = new BedrockApiIntegration({
  tenantId: 'tenant-123',
  storage: dynamoDbStorage
});

// Create test resources from AI-converted test
const resources = await integration.createTestResourcesFromConversion(
  convertedTest,
  tenantId
);

// Execute test
const execution = await integration.executeConvertedTest(
  resources.goal.id,
  environmentId,
  tenantId
);
```

## API Handlers

### Projects Handler
- `createProject(request)` - Create a new project
- `listProjects(request)` - List all projects
- `listProjectGoals(request)` - List goals in a project

### Goals Handler
- `createGoal(request)` - Create a new goal
- `getGoalVersions(request)` - Get goal version history
- `executeGoalSnapshot(request)` - Execute a specific goal version

### Journeys Handler
- `createJourney(request)` - Create a new journey/test suite
- `getJourneyDetails(request)` - Get complete journey structure (Holy Grail)
- `updateJourney(request)` - Update journey details
- `attachCheckpoint(request)` - Attach checkpoint to journey

### Checkpoints Handler
- `createCheckpoint(request)` - Create a new checkpoint/test case
- `getCheckpointDetails(request)` - Get checkpoint details
- `getCheckpointSteps(request)` - List steps in a checkpoint
- `addCheckpointToLibrary(request)` - Add checkpoint to library

### Steps Handler
- `addTestStep(request)` - Add a step to a checkpoint
- `getStepDetails(request)` - Get step details
- `updateStepProperties(request)` - Update step properties

### Executions Handler
- `executeGoal(request)` - Execute a goal
- `getExecutionStatus(request)` - Get execution status
- `getExecutionAnalysis(request)` - Get execution analysis

### Library Handler
- `addToLibrary(request)` - Add checkpoint to library
- `getLibraryCheckpoint(request)` - Get library checkpoint
- `updateLibraryCheckpoint(request)` - Update library checkpoint
- `removeLibraryStep(request)` - Remove step from library
- `moveLibraryStep(request)` - Move step position

### Data Handler
- `createDataTable(request)` - Create test data table
- `getDataTable(request)` - Get data table
- `importDataToTable(request)` - Import data to table

### Environments Handler
- `createEnvironment(request)` - Create test environment

## Platform Support

### Built-in Platforms

- **aws-lambda**: AWS Lambda with API Gateway
- **generic**: Generic HTTP request/response
- **express**: Express.js middleware
- **fastify**: Fastify plugin
- **koa**: Koa middleware
- **bedrock**: AWS Bedrock AI agents

### Custom Platforms

Implement the platform interfaces to add support for your platform:

```javascript
class CustomRuntime {
  parseRequest(event) {
    // Parse platform-specific event
  }
  
  formatResponse(response) {
    // Format platform-specific response
  }
  
  async handleRequest(handler, event, context) {
    // Handle request lifecycle
  }
}
```

## Multi-tenant Configuration

```javascript
const apiService = createApiService({
  platform: 'generic',
  tenantId: 'tenant-123',
  storage: {
    async get(key) { /* retrieve config */ },
    async set(key, value) { /* store config */ }
  }
});

// Switch tenant context
await apiService.setTenant('tenant-456');
```

### Tenant Configuration Options

```javascript
{
  api: {
    baseUrl: 'https://api.virtuoso.qa/api',
    timeout: 30000,
    retry: {
      count: 3,
      minTimeout: 1000,
      maxTimeout: 5000
    }
  },
  features: {
    autoProjectCreation: false,
    parallelExecution: true,
    caching: true,
    rateLimit: {
      enabled: true,
      maxRequests: 1000,
      windowMs: 60000
    }
  }
}
```

## Request Format

All handler methods accept a request object:

```javascript
{
  params: {}, // URL parameters
  body: {},   // Request body
  query: {},  // Query parameters
  headers: {}, // Request headers
  context: {   // Additional context
    tenantId: 'tenant-123'
  }
}
```

## Error Handling

The API layer provides standardized error responses:

```javascript
{
  error: 'VALIDATION_ERROR',
  message: 'Invalid project name',
  details: {
    field: 'name',
    reason: 'Required field missing'
  }
}
```

## Environment Variables

- `VIRTUOSO_API_URL` - Base URL for Virtuoso API
- `API_TOKEN_PARAM` - SSM parameter name for API token
- `NODE_ENV` - Environment (development/production)

## Examples

See the `examples/` directory for complete integration examples:
- `lambda-usage.js` - AWS Lambda integration
- `bedrock-integration.js` - Bedrock AI agent integration
- `express-app.js` - Express.js integration

## License

MIT