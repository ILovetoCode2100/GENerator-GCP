const axios = require('axios');
const { getApiToken } = require('/opt/utils/auth');
const { handleError, VirtuosoError } = require('/opt/utils/error-handler');
const { retryableRequest } = require('/opt/utils/retry');
const { createLogger } = require('/opt/utils/logger');
const config = require('/opt/config');

const logger = createLogger('VirtuosoProjectHandler');

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
  config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// Route mapping
const routes = {
  'POST /projects': 'createProject',
  'GET /projects': 'listProjects',
  'GET /projects/{projectId}/goals': 'listProjectGoals'
};

// Handler implementations

const createProject = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '/projects';
  
  // Replace path parameters
  if (pathParameters) {
    Object.keys(pathParameters).forEach(key => {
      url = url.replace(`{${key}}`, pathParameters[key]);
    });
  }
  
  // Add query parameters
  if (queryStringParameters) {
    const queryString = new URLSearchParams(queryStringParameters).toString();
    url += `?${queryString}`;
  }
  
  const requestConfig = {
    method: 'POST',
    url
  };
  
  if (body && ['POST', 'PUT', 'PATCH'].includes('POST')) {
    requestConfig.data = body;
  }
  
  logger.info('Making Virtuoso API request', { method: 'POST', url, body });
  
  try {
    const response = await virtuosoApi(requestConfig);
    return response.data;
  } catch (error) {
    logger.error('Virtuoso API error', { error: error.response?.data || error.message });
    throw error;
  }
};

const listProjects = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '/projects';
  
  // Replace path parameters
  if (pathParameters) {
    Object.keys(pathParameters).forEach(key => {
      url = url.replace(`{${key}}`, pathParameters[key]);
    });
  }
  
  // Add query parameters
  if (queryStringParameters) {
    const queryString = new URLSearchParams(queryStringParameters).toString();
    url += `?${queryString}`;
  }
  
  const requestConfig = {
    method: 'GET',
    url
  };
  
  if (body && ['POST', 'PUT', 'PATCH'].includes('GET')) {
    requestConfig.data = body;
  }
  
  logger.info('Making Virtuoso API request', { method: 'GET', url, body });
  
  try {
    const response = await virtuosoApi(requestConfig);
    return response.data;
  } catch (error) {
    logger.error('Virtuoso API error', { error: error.response?.data || error.message });
    throw error;
  }
};

const listProjectGoals = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '/projects/{projectId}/goals';
  
  // Replace path parameters
  if (pathParameters) {
    Object.keys(pathParameters).forEach(key => {
      url = url.replace(`{${key}}`, pathParameters[key]);
    });
  }
  
  // Add query parameters
  if (queryStringParameters) {
    const queryString = new URLSearchParams(queryStringParameters).toString();
    url += `?${queryString}`;
  }
  
  const requestConfig = {
    method: 'GET',
    url
  };
  
  if (body && ['POST', 'PUT', 'PATCH'].includes('GET')) {
    requestConfig.data = body;
  }
  
  logger.info('Making Virtuoso API request', { method: 'GET', url, body });
  
  try {
    const response = await virtuosoApi(requestConfig);
    return response.data;
  } catch (error) {
    logger.error('Virtuoso API error', { error: error.response?.data || error.message });
    throw error;
  }
};

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
      let routeKey = `${event.httpMethod} ${event.resource}`;
      
      // Find the matching handler
      const handlerName = routes[routeKey];
      
      if (!handlerName) {
        logger.error('No handler found for route', { routeKey, availableRoutes: Object.keys(routes) });
        return {
          statusCode: 404,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ error: `No handler found for ${routeKey}` })
        };
      }
      
      const handlers = {
        'createProject': createProject,
        'listProjects': listProjects,
        'listProjectGoals': listProjectGoals
      };
      
      const handler = handlers[handlerName];
      if (!handler) {
        throw new VirtuosoError(`Handler function not found: ${handlerName}`, 500);
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
        'createProject': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return createProject(apiGatewayEvent);
        },
        'listProjects': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return listProjects(apiGatewayEvent);
        },
        'listProjectGoals': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return listProjectGoals(apiGatewayEvent);
        }
      };
      
      if (!handlers[event.action]) {
        throw new VirtuosoError(`Unknown action: ${event.action}`, 400);
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
