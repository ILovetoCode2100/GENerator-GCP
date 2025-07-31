const { Logger } = require('@aws-lambda-powertools/logger');

exports.createLogger = (serviceName) => {
  return new Logger({
    serviceName,
    logLevel: process.env.LOG_LEVEL || 'INFO'
  });
};