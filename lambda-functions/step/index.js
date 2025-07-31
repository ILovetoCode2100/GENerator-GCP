const axios = require('axios');
const { getApiToken } = require('/opt/utils/auth');
const { handleError, VirtuosoError } = require('/opt/utils/error-handler');
const { retryableRequest } = require('/opt/utils/retry');
const { createLogger } = require('/opt/utils/logger');
const config = require('/opt/config');

const logger = createLogger('VirtuosoStepHandler');

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
  'POST /teststeps': 'addTestStep',
  'POST /teststeps?envelope=false': 'addTestStepNoEnvelope',
  'POST /steps': 'addTestStepAlt',
  'GET /teststeps/{stepId}': 'getStepDetails',
  'PUT /teststeps/{stepId}/properties': 'updateStepProperties'
};

// Handler implementations

const addTestStep = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '/teststeps';
  
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

const addTestStepNoEnvelope = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '/teststeps?envelope=false';
  
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

const addTestStepAlt = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '/steps';
  
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

const getStepDetails = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '/teststeps/{stepId}';
  
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

const updateStepProperties = async (event) => {
  const { pathParameters = {}, body: bodyString, queryStringParameters } = event;
  const body = bodyString ? JSON.parse(bodyString) : null;
  
  let url = '/teststeps/{stepId}/properties';
  
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
    method: 'PUT',
    url
  };
  
  if (body && ['POST', 'PUT', 'PATCH'].includes('PUT')) {
    requestConfig.data = body;
  }
  
  logger.info('Making Virtuoso API request', { method: 'PUT', url, body });
  
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
        'addTestStep': addTestStep,
        'addTestStepNoEnvelope': addTestStepNoEnvelope,
        'addTestStepAlt': addTestStepAlt,
        'getStepDetails': getStepDetails,
        'updateStepProperties': updateStepProperties
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
        'addTestStep': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return addTestStep(apiGatewayEvent);
        },
        'addTestStepNoEnvelope': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return addTestStepNoEnvelope(apiGatewayEvent);
        },
        'addTestStepAlt': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return addTestStepAlt(apiGatewayEvent);
        },
        'getStepDetails': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return getStepDetails(apiGatewayEvent);
        },
        'updateStepProperties': async () => {
          const params = event.params || {};
          const body = event.body;
          const queryStringParameters = event.queryStringParameters;
          
          // Simulate API Gateway event
          const apiGatewayEvent = {
            pathParameters: params,
            body: body ? JSON.stringify(body) : null,
            queryStringParameters
          };
          
          return updateStepProperties(apiGatewayEvent);
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
