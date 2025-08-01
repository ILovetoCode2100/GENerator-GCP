/**
 * Error Handler Utility
 * Platform-agnostic error handling
 */

/**
 * Custom error class for Virtuoso API errors
 */
class VirtuosoError extends Error {
  constructor(message, statusCode = 500, details = {}) {
    super(message);
    this.name = 'VirtuosoError';
    this.statusCode = statusCode;
    this.details = details;
    this.timestamp = new Date().toISOString();
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message,
      statusCode: this.statusCode,
      details: this.details,
      timestamp: this.timestamp,
      stack: this.stack
    };
  }
}

/**
 * Format error for HTTP response
 * @param {Error} error - Error to format
 * @param {Object} logger - Logger instance (optional)
 * @returns {Object} Formatted error response
 */
function handleError(error, logger = null) {
  // Log error if logger provided
  if (logger) {
    logger.error('Error occurred', { error });
  }
  
  // Handle custom VirtuosoError
  if (error instanceof VirtuosoError) {
    return {
      status: error.statusCode,
      data: {
        error: error.message,
        code: error.code || 'VIRTUOSO_ERROR',
        details: error.details
      }
    };
  }
  
  // Handle Axios errors (from API calls)
  if (error.response) {
    return {
      status: error.response.status,
      data: {
        error: error.response.data?.message || error.response.data?.error || 'API request failed',
        code: error.response.data?.code || 'API_ERROR',
        details: error.response.data
      }
    };
  }
  
  // Handle request errors (network issues)
  if (error.request) {
    return {
      status: 503,
      data: {
        error: 'Service unavailable',
        code: 'SERVICE_UNAVAILABLE',
        message: error.message
      }
    };
  }
  
  // Handle validation errors
  if (error.name === 'ValidationError') {
    return {
      status: 400,
      data: {
        error: 'Validation failed',
        code: 'VALIDATION_ERROR',
        message: error.message,
        details: error.details || {}
      }
    };
  }
  
  // Default error response
  return {
    status: 500,
    data: {
      error: 'Internal server error',
      code: 'INTERNAL_ERROR',
      message: error.message
    }
  };
}

/**
 * Create validation error
 * @param {string} message - Error message
 * @param {Object} details - Validation details
 * @returns {Error} Validation error
 */
function createValidationError(message, details = {}) {
  const error = new Error(message);
  error.name = 'ValidationError';
  error.statusCode = 400;
  error.details = details;
  return error;
}

/**
 * Create not found error
 * @param {string} resource - Resource type
 * @param {string} identifier - Resource identifier
 * @returns {VirtuosoError} Not found error
 */
function createNotFoundError(resource, identifier) {
  return new VirtuosoError(
    `${resource} not found: ${identifier}`,
    404,
    { resource, identifier }
  );
}

/**
 * Create unauthorized error
 * @param {string} message - Error message
 * @returns {VirtuosoError} Unauthorized error
 */
function createUnauthorizedError(message = 'Unauthorized') {
  return new VirtuosoError(message, 401);
}

/**
 * Create forbidden error
 * @param {string} message - Error message
 * @returns {VirtuosoError} Forbidden error
 */
function createForbiddenError(message = 'Forbidden') {
  return new VirtuosoError(message, 403);
}

/**
 * Wrap async handler with error handling
 * @param {Function} handler - Async handler function
 * @param {Object} logger - Logger instance
 * @returns {Function} Wrapped handler
 */
function withErrorHandling(handler, logger = null) {
  return async (...args) => {
    try {
      return await handler(...args);
    } catch (error) {
      const formattedError = handleError(error, logger);
      
      // Re-throw with formatted error
      const wrappedError = new Error(formattedError.data.error);
      wrappedError.status = formattedError.status;
      wrappedError.data = formattedError.data;
      throw wrappedError;
    }
  };
}

module.exports = {
  VirtuosoError,
  handleError,
  createValidationError,
  createNotFoundError,
  createUnauthorizedError,
  createForbiddenError,
  withErrorHandling
};

// Backward compatibility exports
exports.VirtuosoError = VirtuosoError;
exports.handleError = (error) => {
  // For AWS Lambda compatibility, return in Lambda response format
  const formatted = handleError(error);
  return {
    statusCode: formatted.status,
    body: JSON.stringify(formatted.data)
  };
};