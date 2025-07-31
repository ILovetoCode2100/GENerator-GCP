const { Logger } = require('@aws-lambda-powertools/logger');

const logger = new Logger();

class VirtuosoError extends Error {
  constructor(message, statusCode = 500, details = {}) {
    super(message);
    this.statusCode = statusCode;
    this.details = details;
  }
}

exports.VirtuosoError = VirtuosoError;

exports.handleError = (error) => {
  logger.error('Error occurred', { error });
  
  if (error instanceof VirtuosoError) {
    return {
      statusCode: error.statusCode,
      body: JSON.stringify({
        error: error.message,
        details: error.details
      })
    };
  }
  
  if (error.response) {
    return {
      statusCode: error.response.status,
      body: JSON.stringify({
        error: error.response.data?.message || 'API request failed',
        details: error.response.data
      })
    };
  }
  
  return {
    statusCode: 500,
    body: JSON.stringify({
      error: 'Internal server error',
      message: error.message
    })
  };
};