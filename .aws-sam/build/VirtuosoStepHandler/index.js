const axios = require('axios');
const { getApiToken } = require('/opt/utils/auth');
const { handleError, VirtuosoError } = require('/opt/utils/error-handler');
const { retryableRequest } = require('/opt/utils/retry');
const { createLogger } = require('/opt/utils/logger');
const config = require('/opt/config');

const logger = createLogger('VirtuosoStepHandler');

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

const addTestStep = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/teststeps';
  
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

const addTestStepNoEnvelope = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/teststeps?envelope=false';
  
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

const addTestStepAlt = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/steps';
  
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

const getStepDetails = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/teststeps/{stepId}';
  
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

const updateStepProperties = async (event) => {
  const { params = {}, body, queryStringParameters } = event;
  let url = '/teststeps/{stepId}/properties';
  
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
    method: 'PUT',
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
      'addTestStep': addTestStep,
      'addTestStepNoEnvelope': addTestStepNoEnvelope,
      'addTestStepAlt': addTestStepAlt,
      'getStepDetails': getStepDetails,
      'updateStepProperties': updateStepProperties
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
