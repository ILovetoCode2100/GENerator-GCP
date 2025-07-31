#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Lambda function template with API Gateway integration
const lambdaTemplate = (groupName, endpoints) => `const axios = require('axios');
const { getApiToken } = require('/opt/utils/auth');
const { handleError, VirtuosoError } = require('/opt/utils/error-handler');
const { retryableRequest } = require('/opt/utils/retry');
const { createLogger } = require('/opt/utils/logger');
const config = require('/opt/config');

const logger = createLogger('${groupName}');

// Initialize axios instance for Virtuoso API
const virtuosoApi = axios.create({
  baseURL: config.baseUrl,
  timeout: config.timeout,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
});

// Add auth interceptor
virtuosoApi.interceptors.request.use(async (config) => {
  const token = await getApiToken();
  config.headers.Authorization = \`Bearer \${token}\`;
  return config;
});

// Route mapping
const routes = {
${endpoints.map(ep => `  '${ep.method} ${ep.path}': '${ep.handler}'`).join(',\n')}
};

// Handler implementations
${endpoints.map(endpoint => `
const ${endpoint.handler} = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '${endpoint.path}';
  
  // Replace path parameters
  if (pathParameters) {
    Object.keys(pathParameters).forEach(key => {
      url = url.replace(\`{\${key}}\`, pathParameters[key]);
    });
  }
  
  // Add query parameters
  if (queryStringParameters) {
    const queryString = new URLSearchParams(queryStringParameters).toString();
    url += \`?\${queryString}\`;
  }
  
  const requestConfig = {
    method: '${endpoint.method}',
    url
  };
  
  if (body && ['POST', 'PUT', 'PATCH'].includes('${endpoint.method}')) {
    requestConfig.data = body;
  }
  
  logger.info('Making Virtuoso API request', { method: '${endpoint.method}', url, body });
  
  try {
    const response = await virtuosoApi(requestConfig);
    return response.data;
  } catch (error) {
    logger.error('Virtuoso API error', { error: error.response?.data || error.message });
    throw error;
  }
};`).join('\n')}

// Main handler for API Gateway
exports.handler = async (event) => {
  logger.info('Received event', { 
    httpMethod: event.httpMethod,
    resource: event.resource,
    pathParameters: event.pathParameters,
    queryStringParameters: event.queryStringParameters
  });
  
  try {
    // Check if this is an API Gateway event
    if (event.httpMethod && event.resource) {
      // Normalize the resource path for matching
      let routeKey = \`\${event.httpMethod} \${event.resource}\`;
      
      // Find the matching handler
      const handlerName = routes[routeKey];
      
      if (!handlerName) {
        logger.error('No handler found for route', { routeKey, availableRoutes: Object.keys(routes) });
        return {
          statusCode: 404,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ error: \`No handler found for \${routeKey}\` })
        };
      }
      
      const handlers = {
${endpoints.map(endpoint => `        '${endpoint.handler}': ${endpoint.handler}`).join(',\n')}
      };
      
      const handler = handlers[handlerName];
      if (!handler) {
        throw new VirtuosoError(\`Handler function not found: \${handlerName}\`, 500);
      }
      
      const result = await retryableRequest(
        () => handler(event),
        config.retryConfig
      );
      
      return {
        statusCode: 200,
        headers: { 
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*'
        },
        body: JSON.stringify(result)
      };
    } 
    // Support for direct Lambda invocation with action parameter
    else if (event.action) {
      const handlers = {
${endpoints.map(endpoint => `        '${endpoint.handler}': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return ${endpoint.handler}(apiGatewayEvent);
        }`).join(',\n')}
      };
      
      if (!handlers[event.action]) {
        throw new VirtuosoError(\`Unknown action: \${event.action}\`, 400);
      }
      
      const result = await retryableRequest(
        () => handlers[event.action](),
        config.retryConfig
      );
      
      return {
        statusCode: 200,
        body: JSON.stringify(result)
      };
    } else {
      throw new VirtuosoError('Invalid event format', 400);
    }
  } catch (error) {
    return handleError(error);
  }
};
`;

// Endpoint groups definition (same as in generate-lambdas.js)
const endpointGroups = {
  project: {
    name: 'VirtuosoProjectHandler',
    endpoints: [
      { method: 'POST', path: '/projects', handler: 'createProject' },
      { method: 'GET', path: '/projects', handler: 'listProjects' },
      { method: 'GET', path: '/projects/{projectId}/goals', handler: 'listProjectGoals' }
    ]
  },
  goal: {
    name: 'VirtuosoGoalHandler',
    endpoints: [
      { method: 'POST', path: '/goals', handler: 'createGoal' },
      { method: 'GET', path: '/goals/{goalId}/versions', handler: 'getGoalVersions' },
      { method: 'POST', path: '/goals/{goalId}/snapshots/{snapshotId}/execute', handler: 'executeGoalSnapshot' }
    ]
  },
  journey: {
    name: 'VirtuosoJourneyHandler',
    endpoints: [
      { method: 'POST', path: '/testsuites', handler: 'createJourney' },
      { method: 'POST', path: '/journeys', handler: 'createJourneyAlt' },
      { method: 'GET', path: '/testsuites/latest_status', handler: 'listJourneysWithStatus' },
      { method: 'GET', path: '/testsuites/{journeyId}', handler: 'getJourneyDetails' },
      { method: 'PUT', path: '/testsuites/{journeyId}', handler: 'updateJourney' },
      { method: 'POST', path: '/testsuites/{journeyId}/checkpoints/attach', handler: 'attachCheckpoint' },
      { method: 'POST', path: '/journeys/attach-library', handler: 'attachLibraryCheckpoint' }
    ]
  },
  checkpoint: {
    name: 'VirtuosoCheckpointHandler',
    endpoints: [
      { method: 'POST', path: '/testcases', handler: 'createCheckpoint' },
      { method: 'POST', path: '/checkpoints', handler: 'createCheckpointAlt' },
      { method: 'GET', path: '/testcases/{checkpointId}', handler: 'getCheckpointDetails' },
      { method: 'GET', path: '/checkpoints/{checkpointId}/teststeps', handler: 'getCheckpointSteps' },
      { method: 'POST', path: '/testcases/{checkpointId}/add-to-library', handler: 'addCheckpointToLibrary' }
    ]
  },
  step: {
    name: 'VirtuosoStepHandler',
    endpoints: [
      { method: 'POST', path: '/teststeps', handler: 'addTestStep' },
      { method: 'POST', path: '/teststeps?envelope=false', handler: 'addTestStepNoEnvelope' },
      { method: 'POST', path: '/steps', handler: 'addTestStepAlt' },
      { method: 'GET', path: '/teststeps/{stepId}', handler: 'getStepDetails' },
      { method: 'PUT', path: '/teststeps/{stepId}/properties', handler: 'updateStepProperties' }
    ]
  },
  execution: {
    name: 'VirtuosoExecutionHandler',
    endpoints: [
      { method: 'POST', path: '/executions', handler: 'executeGoal' },
      { method: 'GET', path: '/executions/{executionId}', handler: 'getExecutionStatus' },
      { method: 'GET', path: '/executions/analysis/{executionId}', handler: 'getExecutionAnalysis' }
    ]
  },
  library: {
    name: 'VirtuosoLibraryHandler',
    endpoints: [
      { method: 'POST', path: '/library/checkpoints', handler: 'addToLibrary' },
      { method: 'GET', path: '/library/checkpoints/{libraryCheckpointId}', handler: 'getLibraryCheckpoint' },
      { method: 'PUT', path: '/library/checkpoints/{libraryCheckpointId}', handler: 'updateLibraryCheckpoint' },
      { method: 'DELETE', path: '/library/checkpoints/{libraryCheckpointId}/steps/{testStepId}', handler: 'removeLibraryStep' },
      { method: 'POST', path: '/library/checkpoints/{libraryCheckpointId}/steps/{testStepId}/move', handler: 'moveLibraryStep' }
    ]
  },
  data: {
    name: 'VirtuosoDataHandler',
    endpoints: [
      { method: 'POST', path: '/testdata/tables/create', handler: 'createDataTable' },
      { method: 'GET', path: '/testdata/tables/{tableId}', handler: 'getDataTable' },
      { method: 'POST', path: '/testdata/tables/{tableId}/import', handler: 'importDataToTable' }
    ]
  },
  environment: {
    name: 'VirtuosoEnvironmentHandler',
    endpoints: [
      { method: 'POST', path: '/environments', handler: 'createEnvironment' }
    ]
  }
};

// Update all Lambda functions
console.log('🚀 Updating all Lambda functions with API Gateway integration...\n');

Object.entries(endpointGroups).forEach(([key, group]) => {
  const functionDir = path.join('./lambda-functions', key);
  const indexPath = path.join(functionDir, 'index.js');
  
  // Generate updated function code
  const functionCode = lambdaTemplate(group.name, group.endpoints);
  
  // Write the updated code
  fs.writeFileSync(indexPath, functionCode);
  
  console.log(`✅ Updated ${group.name} (${group.endpoints.length} endpoints)`);
});

console.log('\n✅ All Lambda functions updated successfully!');
console.log('\n📦 Now you need to:');
console.log('1. Redeploy the Lambda functions to AWS');
console.log('2. Run the test suite to verify everything works');