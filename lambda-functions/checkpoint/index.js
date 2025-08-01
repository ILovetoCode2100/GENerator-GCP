const axios = require('axios');
const { getApiToken } = require('/opt/utils/auth');
const { handleError, VirtuosoError } = require('/opt/utils/error-handler');
const { retryableRequest } = require('/opt/utils/retry');
const { createLogger } = require('/opt/utils/logger');
const config = require('/opt/config');

const logger = createLogger('VirtuosoCheckpointHandler');

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

const createCheckpoint = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/testcases';
  
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

const createCheckpointAlt = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/checkpoints';
  
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

const getCheckpointDetails = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/testcases/{checkpointId}';
  
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

const getCheckpointSteps = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/checkpoints/{checkpointId}/teststeps';
  
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

const addCheckpointToLibrary = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/testcases/{checkpointId}/add-to-library';
  
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

// Main handler
exports.handler = async (event) => {
  logger.info('Received event', { event });
  
  try {
    const { action } = event;
    
    const handlers = {
      'createCheckpoint': createCheckpoint,
      'createCheckpointAlt': createCheckpointAlt,
      'getCheckpointDetails': getCheckpointDetails,
      'getCheckpointSteps': getCheckpointSteps,
      'addCheckpointToLibrary': addCheckpointToLibrary
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
