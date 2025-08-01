# Porting Guide: Integrating API Layer into Bedrock Project

This guide explains how to integrate the Virtuoso API management layer from `lambda-api-gen` into the `virtuoso-GENerator-bedrock` project.

## Overview

The API layer has been refactored into a platform-agnostic module (`@virtuoso/api-layer`) that can be seamlessly integrated into the bedrock project to provide resource management capabilities beneath the AI conversion layer.

## Architecture Integration

```
virtuoso-GENerator-bedrock/
├── AI Conversion Layer (existing)
│   ├── Test Parser Agent
│   ├── Test Converter Agent
│   └── Validation Agent
└── API Management Layer (new)
    ├── Resource Manager (from lambda-api-gen)
    ├── Operation Executor
    └── Response Transformer
```

## Integration Steps

### 1. Copy the API Module

Copy the entire `/src/api` directory from `lambda-api-gen` to the bedrock project:

```bash
cp -r lambda-api-gen/src/api bedrock-project/src/api-layer
```

### 2. Install Dependencies

Add the API layer dependencies to your bedrock project:

```json
{
  "dependencies": {
    "axios": "^1.6.0",
    "p-retry": "^5.1.2"
  }
}
```

### 3. Configure Platform Support

In your bedrock Lambda functions, initialize the API service:

```javascript
const { createApiService } = require('./api-layer');

// In your Lambda handler
const apiService = createApiService({
  platform: 'bedrock',
  tenantId: event.tenantId,
  storage: dynamoDbStorage // Your existing storage
});
```

### 4. Integrate with AI Agents

Use the provided integration example to connect AI agents with the API layer:

```javascript
const { BedrockApiIntegration } = require('./api-layer/examples/bedrock-integration');

class TestConverterAgent {
  constructor() {
    this.apiIntegration = new BedrockApiIntegration({
      storage: dynamoDbStorage
    });
  }

  async convertAndCreateTest(sourceTest, tenantId) {
    // AI conversion logic
    const convertedTest = await this.convertTest(sourceTest);
    
    // Create resources via API layer
    const resources = await this.apiIntegration.createTestResourcesFromConversion(
      convertedTest,
      tenantId
    );
    
    return resources;
  }
}
```

### 5. Update SAM Template

Add the API layer to your Lambda functions:

```yaml
Resources:
  BedrockFunction:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: src/
      Handler: handlers/bedrock.handler
      Runtime: nodejs18.x
      Layers:
        - !Ref ApiLayer
      Environment:
        Variables:
          API_TOKEN_PARAM: /virtuoso/bedrock/api-token

  ApiLayer:
    Type: AWS::Serverless::LayerVersion
    Properties:
      LayerName: virtuoso-api-layer
      ContentUri: src/api-layer/
      CompatibleRuntimes:
        - nodejs18.x
```

### 6. Multi-tenant Configuration

Configure tenant-specific settings:

```javascript
// Store tenant configuration
await apiService.getTenantManager().setTenantConfig('tenant-123', {
  api: {
    baseUrl: 'https://api.customer.virtuoso.qa/api',
    timeout: 60000
  },
  features: {
    parallelExecution: true,
    autoProjectCreation: true
  }
});
```

### 7. Use in Agent Tools

Expose API operations as tools for Bedrock agents:

```javascript
const agentTools = apiIntegration.createAgentTools();

// In your agent definition
const tools = {
  createProject: {
    description: "Create a new Virtuoso project",
    inputSchema: { /* ... */ },
    handler: agentTools.createProject
  },
  getTestStructure: {
    description: "Get complete test structure (Holy Grail)",
    inputSchema: { /* ... */ },
    handler: agentTools.getJourneyDetails
  }
};
```

## Key Benefits

1. **Separation of Concerns**: AI conversion logic remains separate from API management
2. **Reusable Components**: All 29 API endpoints available as modular handlers
3. **Platform Agnostic**: Works with Lambda, Bedrock agents, or any other platform
4. **Multi-tenant Ready**: Built-in support for tenant isolation and configuration
5. **Battle-tested**: Inherits retry logic, error handling, and optimizations from lambda-api-gen

## Migration Checklist

- [ ] Copy `/src/api` directory to bedrock project
- [ ] Install required dependencies
- [ ] Update Lambda function handlers to use API service
- [ ] Configure multi-tenant storage backend
- [ ] Update SAM template with API layer
- [ ] Test holy grail endpoint integration
- [ ] Verify tenant isolation works correctly
- [ ] Update CI/CD pipeline to include API layer

## Example: Complete Integration

```javascript
// bedrock-project/src/handlers/convert-and-create.js
const { BedrockApiIntegration } = require('../api-layer/examples/bedrock-integration');
const { TestConverterAgent } = require('../agents/test-converter');

exports.handler = async (event, context) => {
  const integration = new BedrockApiIntegration({
    tenantId: event.tenantId,
    storage: getDynamoStorage()
  });
  
  const converter = new TestConverterAgent();
  
  try {
    // Convert test using AI
    const convertedTest = await converter.convert(event.sourceTest);
    
    // Create resources in Virtuoso
    const resources = await integration.createTestResourcesFromConversion(
      convertedTest,
      event.tenantId
    );
    
    // Execute if requested
    if (event.autoExecute) {
      const execution = await integration.executeConvertedTest(
        resources.goal.id,
        event.environmentId,
        event.tenantId
      );
      
      // Monitor execution
      const result = await integration.monitorExecution(
        execution.id,
        event.tenantId
      );
      
      return {
        resources,
        execution: result
      };
    }
    
    return { resources };
  } catch (error) {
    console.error('Conversion failed:', error);
    throw error;
  }
};
```

## Testing

Run the validation script to ensure all handlers are properly set up:

```bash
cd src/api-layer
npm run validate
```

## Support

For questions about the API layer integration:
1. Review the comprehensive README.md in `/src/api`
2. Check the examples in `/src/api/examples`
3. Refer to the original lambda-api-gen documentation

## Next Steps

After integration:
1. Add API layer metrics to your monitoring
2. Configure rate limiting per tenant
3. Set up API token rotation
4. Enable caching for frequently accessed resources
5. Add custom handlers for bedrock-specific operations