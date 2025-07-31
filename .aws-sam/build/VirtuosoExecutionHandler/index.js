const axios = require('axios');
const { getApiToken } = require('/opt/utils/auth');
const { handleError, VirtuosoError } = require('/opt/utils/error-handler');
const { retryableRequest } = require('/opt/utils/retry');
const { createLogger } = require('/opt/utils/logger');
const config = require('/opt/config');

const logger = createLogger('VirtuosoExecutionHandler');

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

const executeGoal = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/executions';
  
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

const getExecutionStatus = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/executions/{executionId}';
  
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

const getExecutionAnalysis = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/executions/analysis/{executionId}';
  
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
    const { action } = event;
    
    const handlers = {
      'executeGoal': executeGoal,
      'getExecutionStatus': getExecutionStatus,
      'getExecutionAnalysis': getExecutionAnalysis
    };
    
    if (!handlers[action]) {
      throw new VirtuosoError(`Unknown action: ${action}`, 400);
    }
    
    const result = await retryableRequest(
      () => handlers[action](event),
      config.retryConfig
    );
    
    return {
      statusCode: 200,
      body: JSON.stringify(result)
    };
  } catch (error) {
    return handleError(error);
  }
};
