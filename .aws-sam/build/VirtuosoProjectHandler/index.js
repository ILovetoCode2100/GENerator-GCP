const axios = require('axios');
const { getApiToken } = require('/opt/utils/auth');
const { handleError, VirtuosoError } = require('/opt/utils/error-handler');
const { retryableRequest } = require('/opt/utils/retry');
const { createLogger } = require('/opt/utils/logger');
const config = require('/opt/config');

const logger = createLogger('VirtuosoProjectHandler');

// Initialize axios instance
const api = axios.create({
  baseURL: config.baseUrl,
  timeout: config.timeout,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add auth interceptor
api.interceptors.request.use(async (config) => {
  const token = await getApiToken();
  config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// Handler implementations

const createProject = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/projects';
  
  // Replace path parameters
  Object.keys(params).forEach(key => {
    url = url.replace(`{${key}}`, params[key]);
  });
  
  // Add query parameters
  if (queryStringParameters) {
    const queryString = new URLSearchParams(queryStringParameters).toString();
    url += `?${queryString}`;
  }
  
  const requestConfig = {
    method: 'POST',
    url
  };
  
  if (body) {
    requestConfig.data = body;
  }
  
  const response = await api(requestConfig);
  return response.data;
};

const listProjects = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/projects';
  
  // Replace path parameters
  Object.keys(params).forEach(key => {
    url = url.replace(`{${key}}`, params[key]);
  });
  
  // Add query parameters
  if (queryStringParameters) {
    const queryString = new URLSearchParams(queryStringParameters).toString();
    url += `?${queryString}`;
  }
  
  const requestConfig = {
    method: 'GET',
    url
  };
  
  if (body) {
    requestConfig.data = body;
  }
  
  const response = await api(requestConfig);
  return response.data;
};

const listProjectGoals = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/projects/{projectId}/goals';
  
  // Replace path parameters
  Object.keys(params).forEach(key => {
    url = url.replace(`{${key}}`, params[key]);
  });
  
  // Add query parameters
  if (queryStringParameters) {
    const queryString = new URLSearchParams(queryStringParameters).toString();
    url += `?${queryString}`;
  }
  
  const requestConfig = {
    method: 'GET',
    url
  };
  
  if (body) {
    requestConfig.data = body;
  }
  
  const response = await api(requestConfig);
  return response.data;
};

// Main handler
exports.handler = async (event) => {
  logger.info('Received event', { event });
  
  try {
    // Handle API Gateway events
    const httpMethod = event.httpMethod;
    const path = event.path || event.resource;
    
    let handler;
    let action;
    
    // Route based on HTTP method and path
    if (httpMethod === 'POST' && path.includes('/projects')) {
      handler = createProject;
      action = 'createProject';
    } else if (httpMethod === 'GET' && path.includes('/projects') && path.includes('/goals')) {
      handler = listProjectGoals;
      action = 'listProjectGoals';
    } else if (httpMethod === 'GET' && path.includes('/projects')) {
      handler = listProjects;
      action = 'listProjects';
    } else {
      // Fallback to action-based routing for direct invocation
      const { action: directAction } = event;
      const handlers = {
        'createProject': createProject,
        'listProjects': listProjects,
        'listProjectGoals': listProjectGoals
      };
      
      if (!handlers[directAction]) {
        throw new VirtuosoError(`Unsupported request: ${httpMethod} ${path}`, 400);
      }
      
      handler = handlers[directAction];
      action = directAction;
    }
    
    // Parse body for API Gateway events
    let parsedEvent = event;
    if (event.body && typeof event.body === 'string') {
      try {
        parsedEvent = {
          ...event,
          body: JSON.parse(event.body)
        };
      } catch (e) {
        parsedEvent = {
          ...event,
          body: event.body
        };
      }
    }
    
    // Extract path parameters from API Gateway
    if (event.pathParameters) {
      parsedEvent.params = event.pathParameters;
    }
    
    const result = await retryableRequest(
      () => handler(parsedEvent),
      config.retryConfig
    );
    
    // Return proper API Gateway response
    return {
      statusCode: 200,
      headers: {
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token',
        'Access-Control-Allow-Methods': 'GET,POST,PUT,DELETE,OPTIONS'
      },
      body: JSON.stringify(result)
    };
  } catch (error) {
    logger.error('Handler error', { error: error.message, stack: error.stack });
    const errorResponse = handleError(error);
    
    // Ensure proper API Gateway error response format
    return {
      statusCode: errorResponse.statusCode || 500,
      headers: {
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token',
        'Access-Control-Allow-Methods': 'GET,POST,PUT,DELETE,OPTIONS'
      },
      body: typeof errorResponse.body === 'string' ? errorResponse.body : JSON.stringify(errorResponse.body || { error: 'Internal server error' })
    };
  }
};
