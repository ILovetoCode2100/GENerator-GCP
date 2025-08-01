/**
 * Example: Using the API layer in AWS Lambda
 */

const { createApiService } = require('../index');

// Create Lambda handler
exports.handler = async (event, context) => {
  // Initialize API service for AWS Lambda
  const apiService = createApiService({
    platform: 'aws-lambda',
    tenantId: event.headers?.['x-tenant-id'] || 'default'
  });

  // Get Lambda runtime
  const runtime = apiService.getServices().runtime;

  // Create unified handler
  const unifiedHandler = async (request) => {
    // Route to appropriate handler based on path
    const route = `${request.method} ${request.path}`;
    
    switch (route) {
      case 'GET /projects':
        return apiService.projects.listProjects(request);
        
      case 'POST /projects':
        return apiService.projects.createProject(request);
        
      case 'GET /testsuites/{journeyId}':
        // Holy Grail endpoint
        return apiService.journeys.getJourneyDetails(request);
        
      case 'POST /executions':
        return apiService.executions.executeGoal(request);
        
      default:
        throw new Error(`Unknown route: ${route}`);
    }
  };

  // Handle request with Lambda runtime
  return runtime.handleRequest(unifiedHandler, event, context);
};

// Alternative: Create individual Lambda functions
const createProjectHandler = () => {
  const apiService = createApiService({
    platform: 'aws-lambda'
  });

  return apiService.createPlatformHandler('lambda');
};

// Example of using with Lambda layers
const createLayeredHandler = () => {
  // When deployed with Lambda layers, the dependencies are in /opt
  process.env.NODE_PATH = '/opt/nodejs/node_modules';
  
  const apiService = createApiService({
    platform: 'aws-lambda',
    config: {
      secrets: {
        parameterPrefix: '/virtuoso/prod'
      },
      logger: {
        serviceName: 'virtuoso-api-lambda'
      }
    }
  });

  return apiService.createPlatformHandler('lambda');
};

module.exports = {
  handler: exports.handler,
  createProjectHandler,
  createLayeredHandler
};