/**
 * AWS Lambda Runtime Implementation
 * Handles AWS Lambda and API Gateway specific request/response formats
 */

const { RuntimeInterface } = require('../../interfaces/runtime.interface');

class LambdaRuntime extends RuntimeInterface {
  constructor(config = {}) {
    super(config);
    this.corsHeaders = config.corsHeaders || {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token',
      'Access-Control-Allow-Methods': 'GET,POST,PUT,DELETE,OPTIONS'
    };
  }

  /**
   * Parse API Gateway event into standard HttpRequest
   * @param {Object} event - API Gateway event
   * @returns {HttpRequest} Standard HTTP request
   */
  parseRequest(event) {
    // Handle both v1 and v2 API Gateway event formats
    const isV2 = event.version === '2.0';
    
    let method, path, pathParameters, queryStringParameters, headers, body;
    
    if (isV2) {
      // API Gateway v2 format
      method = event.requestContext.http.method;
      path = event.requestContext.http.path;
      pathParameters = event.pathParameters || {};
      queryStringParameters = event.queryStringParameters || {};
      headers = event.headers || {};
      body = event.body;
    } else {
      // API Gateway v1 format
      method = event.httpMethod;
      path = event.resource || event.path;
      pathParameters = event.pathParameters || {};
      queryStringParameters = event.queryStringParameters || {};
      headers = event.headers || {};
      body = event.body;
    }

    // Parse body if it's a string
    let parsedBody = body;
    if (body && typeof body === 'string') {
      try {
        parsedBody = JSON.parse(body);
      } catch (e) {
        // Keep as string if not valid JSON
        parsedBody = body;
      }
    }

    // Extract context information
    const context = {
      requestId: event.requestContext?.requestId,
      accountId: event.requestContext?.accountId,
      apiId: event.requestContext?.apiId,
      stage: event.requestContext?.stage,
      sourceIp: event.requestContext?.identity?.sourceIp,
      userAgent: event.requestContext?.identity?.userAgent
    };

    return {
      method,
      path,
      params: pathParameters,
      query: queryStringParameters,
      headers,
      body: parsedBody,
      context,
      rawEvent: event
    };
  }

  /**
   * Format standard response for API Gateway
   * @param {HttpResponse} response - Standard HTTP response
   * @returns {Object} API Gateway response format
   */
  formatResponse(response) {
    const apiGatewayResponse = {
      statusCode: response.status || 200,
      headers: {
        'Content-Type': 'application/json',
        ...this.corsHeaders,
        ...response.headers
      }
    };

    // Handle different response data types
    if (response.data !== undefined) {
      if (typeof response.data === 'string') {
        apiGatewayResponse.body = response.data;
      } else {
        apiGatewayResponse.body = JSON.stringify(response.data);
      }
    } else {
      apiGatewayResponse.body = '';
    }

    // Add isBase64Encoded if needed
    if (response.isBase64Encoded) {
      apiGatewayResponse.isBase64Encoded = true;
    }

    return apiGatewayResponse;
  }

  /**
   * Format error response for API Gateway
   * @param {Error} error - Error object
   * @returns {Object} API Gateway error response
   */
  formatErrorResponse(error) {
    const statusCode = error.statusCode || error.status || 500;
    const errorResponse = {
      error: error.message || 'Internal Server Error',
      code: error.code || 'INTERNAL_ERROR'
    };

    // Add details if available
    if (error.details) {
      errorResponse.details = error.details;
    }

    // Add request ID if available
    if (error.requestId) {
      errorResponse.requestId = error.requestId;
    }

    // Add stack trace in development
    if (this.config.includeStackTrace && error.stack) {
      errorResponse.stack = error.stack.split('\n');
    }

    return {
      statusCode,
      headers: {
        'Content-Type': 'application/json',
        ...this.corsHeaders
      },
      body: JSON.stringify(errorResponse)
    };
  }

  /**
   * Extract Lambda context information
   * @param {Object} context - Lambda context object
   * @returns {Object} Extracted context
   */
  extractContext(context) {
    if (!context) return {};

    return {
      functionName: context.functionName,
      functionVersion: context.functionVersion,
      invokedFunctionArn: context.invokedFunctionArn,
      memoryLimitInMB: context.memoryLimitInMB,
      awsRequestId: context.awsRequestId,
      logGroupName: context.logGroupName,
      logStreamName: context.logStreamName,
      remainingTime: context.getRemainingTimeInMillis ? context.getRemainingTimeInMillis() : null
    };
  }

  /**
   * Handle preflight CORS requests
   * @param {HttpRequest} request - Parsed request
   * @returns {HttpResponse|null} CORS response or null
   */
  handleCors(request) {
    if (request.method === 'OPTIONS') {
      return {
        status: 200,
        headers: this.corsHeaders,
        data: ''
      };
    }
    return null;
  }

  /**
   * Override handleRequest to add CORS handling
   */
  async handleRequest(handler, event, context) {
    try {
      const request = this.parseRequest(event);
      request.context = { ...request.context, ...this.extractContext(context) };
      
      // Handle CORS preflight
      const corsResponse = this.handleCors(request);
      if (corsResponse) {
        return this.formatResponse(corsResponse);
      }
      
      const response = await handler(request);
      
      return this.formatResponse(response);
    } catch (error) {
      // Add request context to error
      if (event.requestContext?.requestId) {
        error.requestId = event.requestContext.requestId;
      }
      return this.formatErrorResponse(error);
    }
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    return 'aws-lambda';
  }

  /**
   * Create a Lambda handler function
   * @param {Function} handler - Core handler logic
   * @returns {Function} Lambda handler
   */
  createHandler(handler) {
    return async (event, context) => {
      return this.handleRequest(handler, event, context);
    };
  }
}

module.exports = { LambdaRuntime };