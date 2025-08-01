/**
 * Generic Runtime Implementation
 * Platform-agnostic runtime for standard HTTP servers
 */

const { RuntimeInterface } = require('../../interfaces/runtime.interface');

class GenericRuntime extends RuntimeInterface {
  constructor(config = {}) {
    super(config);
  }

  /**
   * Parse generic HTTP request into standard HttpRequest
   * @param {Object} event - Generic HTTP request event
   * @returns {HttpRequest} Standard HTTP request
   */
  parseRequest(event) {
    // Handle Express-like request format
    if (event.method && event.url) {
      return this.parseExpressRequest(event);
    }

    // Handle generic format
    return {
      method: event.method || 'GET',
      path: event.path || '/',
      params: event.params || {},
      query: event.query || {},
      headers: event.headers || {},
      body: event.body,
      context: event.context || {}
    };
  }

  /**
   * Parse Express-like request
   * @param {Object} req - Express request object
   * @returns {HttpRequest} Standard HTTP request
   */
  parseExpressRequest(req) {
    return {
      method: req.method,
      path: req.route?.path || req.path || req.url,
      params: req.params || {},
      query: req.query || {},
      headers: req.headers || {},
      body: req.body,
      context: {
        ip: req.ip,
        protocol: req.protocol,
        hostname: req.hostname,
        originalUrl: req.originalUrl
      }
    };
  }

  /**
   * Format standard response for generic HTTP
   * @param {HttpResponse} response - Standard HTTP response
   * @returns {Object} Generic HTTP response format
   */
  formatResponse(response) {
    return {
      status: response.status || 200,
      headers: {
        'Content-Type': 'application/json',
        ...response.headers
      },
      body: response.data
    };
  }

  /**
   * Format error response
   * @param {Error} error - Error object
   * @returns {Object} Generic error response
   */
  formatErrorResponse(error) {
    const status = error.statusCode || error.status || 500;
    const errorResponse = {
      error: error.message || 'Internal Server Error',
      code: error.code || 'INTERNAL_ERROR'
    };

    if (error.details) {
      errorResponse.details = error.details;
    }

    if (this.config.includeStackTrace && error.stack) {
      errorResponse.stack = error.stack.split('\n');
    }

    return {
      status,
      headers: {
        'Content-Type': 'application/json'
      },
      body: errorResponse
    };
  }

  /**
   * Extract generic context information
   * @param {Object} context - Generic context object
   * @returns {Object} Extracted context
   */
  extractContext(context) {
    if (!context) return {};

    return {
      requestId: context.requestId || this.generateRequestId(),
      timestamp: context.timestamp || new Date().toISOString(),
      ...context
    };
  }

  /**
   * Generate a request ID
   * @returns {string} Request ID
   */
  generateRequestId() {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    return 'generic';
  }

  /**
   * Create an Express middleware
   * @param {Function} handler - Core handler logic
   * @returns {Function} Express middleware
   */
  createExpressMiddleware(handler) {
    return async (req, res, next) => {
      try {
        const request = this.parseExpressRequest(req);
        request.context.requestId = this.generateRequestId();
        
        const response = await handler(request);
        const formatted = this.formatResponse(response);
        
        res.status(formatted.status);
        Object.entries(formatted.headers).forEach(([key, value]) => {
          res.setHeader(key, value);
        });
        
        if (typeof formatted.body === 'object') {
          res.json(formatted.body);
        } else {
          res.send(formatted.body);
        }
      } catch (error) {
        const errorResponse = this.formatErrorResponse(error);
        res.status(errorResponse.status).json(errorResponse.body);
      }
    };
  }

  /**
   * Create a generic HTTP handler
   * @param {Function} handler - Core handler logic
   * @returns {Function} Generic handler
   */
  createHandler(handler) {
    return async (event, context = {}) => {
      return this.handleRequest(handler, event, context);
    };
  }

  /**
   * Create a Fastify handler
   * @param {Function} handler - Core handler logic
   * @returns {Function} Fastify handler
   */
  createFastifyHandler(handler) {
    return async (request, reply) => {
      try {
        const parsedRequest = {
          method: request.method,
          path: request.url,
          params: request.params || {},
          query: request.query || {},
          headers: request.headers || {},
          body: request.body,
          context: {
            requestId: request.id,
            ip: request.ip,
            hostname: request.hostname
          }
        };

        const response = await handler(parsedRequest);
        const formatted = this.formatResponse(response);
        
        reply
          .code(formatted.status)
          .headers(formatted.headers)
          .send(formatted.body);
      } catch (error) {
        const errorResponse = this.formatErrorResponse(error);
        reply
          .code(errorResponse.status)
          .headers(errorResponse.headers)
          .send(errorResponse.body);
      }
    };
  }

  /**
   * Create a Koa handler
   * @param {Function} handler - Core handler logic
   * @returns {Function} Koa middleware
   */
  createKoaMiddleware(handler) {
    return async (ctx, next) => {
      try {
        const request = {
          method: ctx.method,
          path: ctx.path,
          params: ctx.params || {},
          query: ctx.query || {},
          headers: ctx.headers || {},
          body: ctx.request.body,
          context: {
            requestId: ctx.state.requestId || this.generateRequestId(),
            ip: ctx.ip,
            protocol: ctx.protocol,
            host: ctx.host
          }
        };

        const response = await handler(request);
        const formatted = this.formatResponse(response);
        
        ctx.status = formatted.status;
        Object.entries(formatted.headers).forEach(([key, value]) => {
          ctx.set(key, value);
        });
        ctx.body = formatted.body;
      } catch (error) {
        const errorResponse = this.formatErrorResponse(error);
        ctx.status = errorResponse.status;
        ctx.body = errorResponse.body;
      }
    };
  }
}

module.exports = { GenericRuntime };