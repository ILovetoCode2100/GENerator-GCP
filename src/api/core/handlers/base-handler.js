/**
 * Base Handler Class
 * Provides common functionality for all API handlers
 */

const axios = require('axios');

class BaseHandler {
  /**
   * Initialize base handler
   * @param {Object} config - Handler configuration
   * @param {Object} services - Platform services (logger, secretManager, etc.)
   */
  constructor(config, services) {
    this.config = config;
    this.services = services;
    this.logger = services.logger;
    
    // Initialize HTTP client
    this.client = this.createHttpClient();
  }

  /**
   * Create configured HTTP client
   * @returns {Object} Axios instance
   */
  createHttpClient() {
    const client = axios.create({
      baseURL: this.config.baseUrl || 'https://api.virtuoso.qa/api',
      timeout: this.config.timeout || 30000,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
    });

    // Add request interceptor for authentication
    client.interceptors.request.use(
      async (config) => {
        try {
          const token = await this.getApiToken();
          config.headers.Authorization = `Bearer ${token}`;
        } catch (error) {
          this.logger.error('Failed to get API token', { error });
          throw error;
        }
        
        // Log request
        this.logger.debug('API request', {
          method: config.method,
          url: config.url,
          params: config.params
        });
        
        return config;
      },
      (error) => {
        this.logger.error('Request interceptor error', { error });
        return Promise.reject(error);
      }
    );

    // Add response interceptor for logging
    client.interceptors.response.use(
      (response) => {
        this.logger.debug('API response', {
          status: response.status,
          url: response.config.url
        });
        return response;
      },
      (error) => {
        if (error.response) {
          this.logger.error('API error response', {
            status: error.response.status,
            url: error.config?.url,
            data: error.response.data
          });
        } else {
          this.logger.error('API request failed', {
            message: error.message,
            url: error.config?.url
          });
        }
        return Promise.reject(error);
      }
    );

    return client;
  }

  /**
   * Get API token from secret manager
   * @returns {Promise<string>} API token
   */
  async getApiToken() {
    if (this.cachedToken && this.tokenExpiry > Date.now()) {
      return this.cachedToken;
    }

    const token = await this.services.secretManager.getApiToken();
    
    // Cache for 5 minutes
    this.cachedToken = token;
    this.tokenExpiry = Date.now() + (5 * 60 * 1000);
    
    return token;
  }

  /**
   * Make API request with retry logic
   * @param {Object} options - Request options
   * @returns {Promise<Object>} Response data
   */
  async request(options) {
    const retryConfig = this.config.retryConfig || {
      retries: 3,
      minTimeout: 1000,
      maxTimeout: 5000
    };

    let lastError;
    
    for (let attempt = 0; attempt <= retryConfig.retries; attempt++) {
      try {
        const response = await this.client.request(options);
        return response.data;
      } catch (error) {
        lastError = error;
        
        // Don't retry on client errors (4xx)
        if (error.response && error.response.status >= 400 && error.response.status < 500) {
          throw this.formatError(error);
        }
        
        // Calculate retry delay
        if (attempt < retryConfig.retries) {
          const delay = Math.min(
            retryConfig.minTimeout * Math.pow(2, attempt),
            retryConfig.maxTimeout
          );
          
          this.logger.warn(`Request failed, retrying in ${delay}ms`, {
            attempt: attempt + 1,
            maxAttempts: retryConfig.retries + 1,
            error: error.message
          });
          
          await this.sleep(delay);
        }
      }
    }
    
    throw this.formatError(lastError);
  }

  /**
   * Format error for consistent error handling
   * @param {Error} error - Original error
   * @returns {Error} Formatted error
   */
  formatError(error) {
    const formatted = new Error(error.message);
    
    if (error.response) {
      formatted.statusCode = error.response.status;
      formatted.details = error.response.data;
      
      // Extract specific error message from response
      if (error.response.data?.message) {
        formatted.message = error.response.data.message;
      } else if (error.response.data?.error) {
        formatted.message = error.response.data.error;
      }
    } else if (error.request) {
      formatted.statusCode = 503;
      formatted.code = 'SERVICE_UNAVAILABLE';
      formatted.message = 'Unable to reach Virtuoso API';
    } else {
      formatted.statusCode = 500;
      formatted.code = 'INTERNAL_ERROR';
    }
    
    return formatted;
  }

  /**
   * Sleep helper for retry delays
   * @param {number} ms - Milliseconds to sleep
   * @returns {Promise<void>}
   */
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Process path parameters in URL
   * @param {string} path - URL path with parameters
   * @param {Object} params - Parameter values
   * @returns {string} Processed path
   */
  processPath(path, params = {}) {
    let processedPath = path;
    
    Object.entries(params).forEach(([key, value]) => {
      processedPath = processedPath.replace(`{${key}}`, encodeURIComponent(value));
    });
    
    return processedPath;
  }

  /**
   * Build query string from parameters
   * @param {Object} params - Query parameters
   * @returns {string} Query string
   */
  buildQueryString(params) {
    if (!params || Object.keys(params).length === 0) {
      return '';
    }
    
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, value);
      }
    });
    
    const queryString = searchParams.toString();
    return queryString ? `?${queryString}` : '';
  }

  /**
   * Validate required fields in data
   * @param {Object} data - Data to validate
   * @param {Array<string>} required - Required field names
   * @throws {Error} If validation fails
   */
  validateRequired(data, required) {
    const missing = required.filter(field => !data || data[field] === undefined);
    
    if (missing.length > 0) {
      const error = new Error(`Missing required fields: ${missing.join(', ')}`);
      error.statusCode = 400;
      error.code = 'VALIDATION_ERROR';
      throw error;
    }
  }

  /**
   * Generic GET request
   * @param {string} path - API path
   * @param {Object} params - Path parameters
   * @param {Object} query - Query parameters
   * @returns {Promise<Object>} Response data
   */
  async get(path, params = {}, query = {}) {
    const processedPath = this.processPath(path, params);
    const queryString = this.buildQueryString(query);
    
    return this.request({
      method: 'GET',
      url: `${processedPath}${queryString}`
    });
  }

  /**
   * Generic POST request
   * @param {string} path - API path
   * @param {Object} data - Request body
   * @param {Object} params - Path parameters
   * @param {Object} query - Query parameters
   * @returns {Promise<Object>} Response data
   */
  async post(path, data = {}, params = {}, query = {}) {
    const processedPath = this.processPath(path, params);
    const queryString = this.buildQueryString(query);
    
    return this.request({
      method: 'POST',
      url: `${processedPath}${queryString}`,
      data
    });
  }

  /**
   * Generic PUT request
   * @param {string} path - API path
   * @param {Object} data - Request body
   * @param {Object} params - Path parameters
   * @param {Object} query - Query parameters
   * @returns {Promise<Object>} Response data
   */
  async put(path, data = {}, params = {}, query = {}) {
    const processedPath = this.processPath(path, params);
    const queryString = this.buildQueryString(query);
    
    return this.request({
      method: 'PUT',
      url: `${processedPath}${queryString}`,
      data
    });
  }

  /**
   * Generic DELETE request
   * @param {string} path - API path
   * @param {Object} params - Path parameters
   * @param {Object} query - Query parameters
   * @returns {Promise<Object>} Response data
   */
  async delete(path, params = {}, query = {}) {
    const processedPath = this.processPath(path, params);
    const queryString = this.buildQueryString(query);
    
    return this.request({
      method: 'DELETE',
      url: `${processedPath}${queryString}`
    });
  }

  /**
   * Create handler method that wraps business logic
   * @param {Function} fn - Handler function
   * @returns {Function} Wrapped handler
   */
  createHandler(fn) {
    return async (request) => {
      const timer = this.logger.startTimer(`${this.constructor.name}.${fn.name}`);
      
      try {
        this.logger.info(`Handling ${fn.name}`, {
          method: request.method,
          path: request.path,
          params: request.params
        });
        
        const result = await fn.call(this, request);
        
        timer({ status: 'success' });
        
        return {
          status: 200,
          data: result
        };
      } catch (error) {
        timer({ status: 'error', error: error.message });
        
        throw error;
      }
    };
  }
}

module.exports = { BaseHandler };