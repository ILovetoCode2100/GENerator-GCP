/**
 * Runtime Interface
 * Abstracts the platform-specific runtime handling (Lambda, Bedrock, etc.)
 */

/**
 * @typedef {Object} HttpRequest
 * @property {string} method - HTTP method (GET, POST, PUT, DELETE, etc.)
 * @property {string} path - Request path
 * @property {Object} params - Path parameters
 * @property {Object} query - Query string parameters
 * @property {Object} headers - Request headers
 * @property {*} body - Request body (parsed)
 * @property {Object} context - Additional context (tenant, user, etc.)
 */

/**
 * @typedef {Object} HttpResponse
 * @property {number} status - HTTP status code
 * @property {Object} headers - Response headers
 * @property {*} data - Response data
 */

/**
 * Runtime interface for handling platform-specific request/response
 * @interface
 */
class RuntimeInterface {
  /**
   * Initialize the runtime with configuration
   * @param {Object} config - Runtime configuration
   */
  constructor(config) {
    if (new.target === RuntimeInterface) {
      throw new Error('RuntimeInterface is an abstract class');
    }
    this.config = config;
  }

  /**
   * Parse platform-specific event into standard HttpRequest
   * @param {*} event - Platform-specific event object
   * @returns {HttpRequest} Standard HTTP request
   */
  parseRequest(event) {
    throw new Error('parseRequest must be implemented by subclass');
  }

  /**
   * Format standard response for platform
   * @param {HttpResponse} response - Standard HTTP response
   * @returns {*} Platform-specific response format
   */
  formatResponse(response) {
    throw new Error('formatResponse must be implemented by subclass');
  }

  /**
   * Handle request with given handler function
   * @param {Function} handler - Request handler function
   * @param {*} event - Platform-specific event
   * @param {*} context - Platform-specific context
   * @returns {Promise<*>} Platform-specific response
   */
  async handleRequest(handler, event, context) {
    try {
      const request = this.parseRequest(event);
      request.context = { ...request.context, ...this.extractContext(context) };
      
      const response = await handler(request);
      
      return this.formatResponse(response);
    } catch (error) {
      return this.formatErrorResponse(error);
    }
  }

  /**
   * Extract platform-specific context
   * @param {*} context - Platform context
   * @returns {Object} Extracted context
   */
  extractContext(context) {
    return {};
  }

  /**
   * Format error response for platform
   * @param {Error} error - Error object
   * @returns {*} Platform-specific error response
   */
  formatErrorResponse(error) {
    throw new Error('formatErrorResponse must be implemented by subclass');
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    throw new Error('getPlatform must be implemented by subclass');
  }
}

module.exports = { RuntimeInterface };