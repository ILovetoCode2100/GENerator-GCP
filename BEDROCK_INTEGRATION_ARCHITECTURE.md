# Bedrock Integration Architecture

## Overview

This document outlines the architecture for integrating the lambda-api-gen API management layer into the virtuoso-GENerator-bedrock project, creating a unified platform with clear separation of concerns.

## Integrated Architecture Design

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Virtuoso GENerator Bedrock Platform                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                           Multi-Agent Orchestration Layer                    │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐          │
│  │  Analyzer  │  │ Converter  │  │   Healer   │  │  Learner   │          │
│  │   Agent    │  │   Agent    │  │   Agent    │  │   Agent    │          │
│  └──────┬─────┘  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘          │
│         └───────────────┴───────────────┴───────────────┘                  │
│                                    │                                        │
├────────────────────────────────────┼────────────────────────────────────────┤
│                          Service Integration Layer                           │
│                                    │                                        │
│  ┌─────────────────────────────────┴────────────────────────────────────┐ │
│  │                        API Management Service                          │ │
│  │  ┌───────────────┐  ┌──────────────────┐  ┌───────────────────────┐ │ │
│  │  │Resource Manager│  │Operation Executor│  │  Response Transformer │ │ │
│  │  └───────┬───────┘  └────────┬─────────┘  └──────────┬────────────┘ │ │
│  │          └────────────────────┴────────────────────────┘             │ │
│  │                               │                                       │ │
│  │  ┌────────────────────────────┴─────────────────────────────────┐   │ │
│  │  │                    Virtuoso API Client                        │   │ │
│  │  │  ┌─────────┐ ┌─────────┐ ┌──────────┐ ┌──────────┐         │   │ │
│  │  │  │ Project │ │  Goal   │ │ Journey  │ │Checkpoint│  ...    │   │ │
│  │  │  │ Handler │ │ Handler │ │ Handler  │ │ Handler  │         │   │ │
│  │  │  └─────────┘ └─────────┘ └──────────┘ └──────────┘         │   │ │
│  │  └──────────────────────────────────────────────────────────────┘   │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                    │                                        │
├────────────────────────────────────┼────────────────────────────────────────┤
│                          Platform Abstraction Layer                          │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │   Runtime    │  │    Secret    │  │   Logger    │  │  Configuration  │ │
│  │   Adapter    │  │   Manager    │  │  Interface  │  │     Manager     │ │
│  └─────────────┘  └──────────────┘  └─────────────┘  └─────────────────┘ │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                         Infrastructure Layer                                 │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │  AWS Lambda │  │      S3      │  │  DynamoDB   │  │ Secrets Manager │ │
│  │  Functions  │  │   Storage    │  │  Database   │  │                 │ │
│  └─────────────┘  └──────────────┘  └─────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Directory Structure in Bedrock

```
virtuoso-generator-bedrock/
├── src/
│   ├── api/                      # New API Management Layer
│   │   ├── core/                 # Core API logic (portable)
│   │   │   ├── handlers/         # API endpoint handlers
│   │   │   │   ├── project.js
│   │   │   │   ├── goal.js
│   │   │   │   ├── journey.js
│   │   │   │   ├── checkpoint.js
│   │   │   │   ├── step.js
│   │   │   │   ├── execution.js
│   │   │   │   ├── library.js
│   │   │   │   ├── data.js
│   │   │   │   └── environment.js
│   │   │   ├── services/         # Business logic services
│   │   │   │   ├── resource-manager.js
│   │   │   │   ├── operation-executor.js
│   │   │   │   └── response-transformer.js
│   │   │   └── client/           # Virtuoso API client
│   │   │       ├── virtuoso-client.js
│   │   │       └── endpoint-registry.js
│   │   │
│   │   ├── abstractions/         # Platform abstractions
│   │   │   ├── interfaces/       # Interface definitions
│   │   │   │   ├── runtime.interface.js
│   │   │   │   ├── secret.interface.js
│   │   │   │   ├── logger.interface.js
│   │   │   │   └── config.interface.js
│   │   │   └── implementations/  # Platform-specific implementations
│   │   │       ├── aws/
│   │   │       │   ├── lambda-runtime.js
│   │   │       │   ├── ssm-secret-manager.js
│   │   │       │   └── powertools-logger.js
│   │   │       └── bedrock/
│   │   │           ├── bedrock-runtime.js
│   │   │           ├── bedrock-secret-manager.js
│   │   │           └── bedrock-logger.js
│   │   │
│   │   ├── utils/                # Shared utilities (from lambda-layer)
│   │   │   ├── error-handler.js
│   │   │   ├── retry.js
│   │   │   └── validator.js
│   │   │
│   │   ├── config/               # Configuration management
│   │   │   ├── api-config.js
│   │   │   ├── tenant-config.js
│   │   │   └── endpoint-groups.js
│   │   │
│   │   └── index.js              # Main API service export
│   │
│   ├── agents/                   # Existing Bedrock agents
│   ├── converters/               # Existing converters
│   ├── processors/               # Existing processors
│   └── lambdas/                  # Existing Lambda functions
```

## Integration Points

### 1. Agent-to-API Integration

```javascript
// In converter agent
const apiService = require('../api');

async function createTestResources(convertedTest) {
  // Create project if needed
  const project = await apiService.projects.create({
    name: convertedTest.projectName,
    description: 'AI-generated test project'
  });
  
  // Create goal
  const goal = await apiService.goals.create({
    projectId: project.id,
    name: convertedTest.goalName
  });
  
  // Create journey
  const journey = await apiService.journeys.create({
    goalId: goal.id,
    name: convertedTest.journeyName
  });
  
  // Add checkpoints and steps
  for (const checkpoint of convertedTest.checkpoints) {
    const cp = await apiService.checkpoints.create({
      journeyId: journey.id,
      name: checkpoint.name
    });
    
    for (const step of checkpoint.steps) {
      await apiService.steps.add({
        checkpointId: cp.id,
        ...step
      });
    }
  }
  
  return { project, goal, journey };
}
```

### 2. Multi-Tenant Support

```javascript
// Tenant-aware API client
class TenantAwareApiClient {
  constructor(tenantId, config) {
    this.tenantId = tenantId;
    this.config = config;
    this.client = new VirtuosoClient(this.getApiToken());
  }
  
  async getApiToken() {
    const secretManager = PlatformFactory.getSecretManager();
    return secretManager.getSecret(`virtuoso/tenant/${this.tenantId}/api-token`);
  }
  
  // Add tenant context to all requests
  async request(method, endpoint, data) {
    const headers = {
      'X-Tenant-ID': this.tenantId,
      'X-Request-ID': generateRequestId()
    };
    
    return this.client.request(method, endpoint, data, headers);
  }
}
```

### 3. Platform Abstraction Example

```javascript
// Runtime abstraction
class RuntimeAdapter {
  static create(platform) {
    switch (platform) {
      case 'aws-lambda':
        return new LambdaRuntime();
      case 'bedrock':
        return new BedrockRuntime();
      default:
        throw new Error(`Unsupported platform: ${platform}`);
    }
  }
}

// Lambda implementation
class LambdaRuntime {
  async handleRequest(handler, event) {
    const request = this.parseApiGatewayEvent(event);
    const response = await handler(request);
    return this.formatLambdaResponse(response);
  }
  
  parseApiGatewayEvent(event) {
    return {
      method: event.httpMethod,
      path: event.resource,
      params: event.pathParameters || {},
      query: event.queryStringParameters || {},
      body: event.body ? JSON.parse(event.body) : null,
      headers: event.headers || {}
    };
  }
  
  formatLambdaResponse(response) {
    return {
      statusCode: response.status || 200,
      headers: {
        'Content-Type': 'application/json',
        ...response.headers
      },
      body: JSON.stringify(response.data)
    };
  }
}

// Bedrock implementation
class BedrockRuntime {
  async handleRequest(handler, event) {
    const request = this.parseBedrockEvent(event);
    const response = await handler(request);
    return this.formatBedrockResponse(response);
  }
  
  parseBedrockEvent(event) {
    // Bedrock-specific event parsing
    return {
      method: event.method,
      path: event.path,
      params: event.params,
      query: event.query,
      body: event.body,
      headers: event.headers
    };
  }
  
  formatBedrockResponse(response) {
    // Bedrock-specific response formatting
    return {
      success: true,
      status: response.status,
      data: response.data
    };
  }
}
```

## Service Layer Design

### Resource Manager

```javascript
class ResourceManager {
  constructor(apiClient) {
    this.apiClient = apiClient;
    this.cache = new Map();
  }
  
  async getOrCreateProject(name, options = {}) {
    const cacheKey = `project:${name}`;
    
    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey);
    }
    
    // Check if project exists
    const projects = await this.apiClient.projects.list();
    const existing = projects.find(p => p.name === name);
    
    if (existing) {
      this.cache.set(cacheKey, existing);
      return existing;
    }
    
    // Create new project
    const project = await this.apiClient.projects.create({
      name,
      description: options.description || 'Created by API Management Layer',
      ...options
    });
    
    this.cache.set(cacheKey, project);
    return project;
  }
  
  // Similar methods for goals, journeys, etc.
}
```

### Operation Executor

```javascript
class OperationExecutor {
  constructor(apiClient, config) {
    this.apiClient = apiClient;
    this.config = config;
    this.queue = new OperationQueue(config.concurrency);
  }
  
  async executeBatch(operations) {
    const results = [];
    
    for (const op of operations) {
      const result = await this.queue.add(async () => {
        try {
          return await this.executeOperation(op);
        } catch (error) {
          if (this.config.continueOnError) {
            return { success: false, error: error.message };
          }
          throw error;
        }
      });
      
      results.push(result);
    }
    
    return results;
  }
  
  async executeOperation(operation) {
    const { type, resource, action, data } = operation;
    const handler = this.apiClient[resource];
    
    if (!handler || !handler[action]) {
      throw new Error(`Invalid operation: ${resource}.${action}`);
    }
    
    return handler[action](data);
  }
}
```

## Configuration System

### Multi-Tenant Configuration

```javascript
// config/tenant-config.js
class TenantConfigManager {
  constructor(storage) {
    this.storage = storage; // DynamoDB, S3, etc.
  }
  
  async getTenantConfig(tenantId) {
    const config = await this.storage.get(`tenant:${tenantId}:config`);
    
    return {
      apiEndpoint: config.apiEndpoint || process.env.VIRTUOSO_API_URL,
      apiTokenSecret: config.apiTokenSecret || `/virtuoso/tenant/${tenantId}/api-token`,
      retryConfig: {
        retries: config.maxRetries || 3,
        minTimeout: config.retryMinTimeout || 1000,
        maxTimeout: config.retryMaxTimeout || 5000
      },
      features: {
        autoProjectCreation: config.autoProjectCreation !== false,
        parallelExecution: config.parallelExecution !== false,
        caching: config.caching !== false
      }
    };
  }
}
```

### Environment-Based Configuration

```javascript
// config/api-config.js
module.exports = {
  development: {
    baseUrl: 'https://api.virtuoso.qa/api',
    timeout: 30000,
    logLevel: 'debug'
  },
  staging: {
    baseUrl: process.env.VIRTUOSO_API_URL,
    timeout: 60000,
    logLevel: 'info'
  },
  production: {
    baseUrl: process.env.VIRTUOSO_API_URL,
    timeout: 120000,
    logLevel: 'warn'
  }
};
```

## Testing Strategy

### Unit Tests

```javascript
// tests/unit/api/handlers/project.test.js
describe('ProjectHandler', () => {
  let handler;
  let mockClient;
  
  beforeEach(() => {
    mockClient = createMockApiClient();
    handler = new ProjectHandler(mockClient);
  });
  
  test('should create project with valid data', async () => {
    const projectData = { name: 'Test Project' };
    const expectedResponse = { id: '123', ...projectData };
    
    mockClient.post.mockResolvedValue(expectedResponse);
    
    const result = await handler.create(projectData);
    
    expect(mockClient.post).toHaveBeenCalledWith('/projects', projectData);
    expect(result).toEqual(expectedResponse);
  });
});
```

### Integration Tests

```javascript
// tests/integration/api/api-service.test.js
describe('API Service Integration', () => {
  let apiService;
  
  beforeAll(async () => {
    apiService = await createApiService({
      platform: 'test',
      tenant: 'test-tenant'
    });
  });
  
  test('should create complete test hierarchy', async () => {
    const project = await apiService.projects.create({
      name: 'Integration Test Project'
    });
    
    const goal = await apiService.goals.create({
      projectId: project.id,
      name: 'Test Goal'
    });
    
    const journey = await apiService.journeys.create({
      goalId: goal.id,
      name: 'Test Journey'
    });
    
    expect(project.id).toBeDefined();
    expect(goal.projectId).toBe(project.id);
    expect(journey.goalId).toBe(goal.id);
  });
});
```

## Migration Benefits

1. **Unified Platform**: Complete test lifecycle management
2. **Separation of Concerns**: AI conversion vs API management
3. **Reusability**: Shared across all agents and converters
4. **Scalability**: Multi-tenant ready with proper abstractions
5. **Maintainability**: Clear interfaces and modular design
6. **Testability**: Comprehensive test coverage possible

## Next Steps

1. Implement core abstractions
2. Port handlers to platform-agnostic format
3. Create bedrock-specific implementations
4. Integrate with existing agent architecture
5. Add comprehensive test coverage
6. Document API usage for agent developers