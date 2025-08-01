const { handleError, VirtuosoError } = require('./error-handler');
const { retryableRequest } = require('./retry');
const config = require('../config');

/**
 * Helper function to determine which handler to use based on API Gateway event
 */
function getHandlerFromEvent(event, routeMap) {
  const { httpMethod, resource } = event;
  const routeKey = `${httpMethod}:${resource}`;
  return routeMap[routeKey];
}

/**
 * Helper function to parse API Gateway event into normalized format
 */
function parseApiGatewayEvent(event) {
  const { pathParameters, queryStringParameters, body } = event;
  
  return {
    params: pathParameters || {},
    queryStringParameters: queryStringParameters || {},
    body: body ? (typeof body === 'string' ? JSON.parse(body) : body) : null
  };
}

/**
 * Create a unified handler that supports both API Gateway proxy events and direct action calls
 */
function createApiGatewayHandler(routeMap, handlers, logger) {
  return async (event) => {
    logger.info('Received event', { event });
    
    try {
      // Check if this is a direct action call (for backward compatibility) or API Gateway event
      let handlerName, normalizedEvent;
      
      if (event.action) {
        // Direct action call (backward compatibility)
        handlerName = event.action;
        normalizedEvent = event;
      } else {
        // API Gateway proxy integration event
        handlerName = getHandlerFromEvent(event, routeMap);
        normalizedEvent = parseApiGatewayEvent(event);
      }
      
      if (!handlers[handlerName]) {
        throw new VirtuosoError(
          `Unknown handler: ${handlerName} for ${event.httpMethod || 'ACTION'} ${event.resource || event.action}`, 
          400
        );
      }
      
      const result = await retryableRequest(
        () => handlers[handlerName](normalizedEvent),
        config.retryConfig
      );
      
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
      const errorResponse = handleError(error);
      return {
        statusCode: errorResponse.statusCode || 500,
        headers: {
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*'
        },
        body: JSON.stringify(errorResponse.body ? JSON.parse(errorResponse.body) : { error: 'Internal server error' })
      };
    }
  };
}

module.exports = {
  createApiGatewayHandler,
  parseApiGatewayEvent,
  getHandlerFromEvent
};