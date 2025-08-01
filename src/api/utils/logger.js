/**
 * Logger Utility
 * Platform-agnostic logger factory
 */

let platformLogger = null;

/**
 * Set the platform logger instance
 * @param {Object} logger - Logger instance from platform factory
 */
function setPlatformLogger(logger) {
  platformLogger = logger;
}

/**
 * Get the platform logger instance
 * @returns {Object} Logger instance
 */
function getPlatformLogger() {
  if (!platformLogger) {
    // Fallback to console logger if platform logger not set
    console.warn('Platform logger not configured, using console logger');
    return createConsoleLogger('default');
  }
  return platformLogger;
}

/**
 * Create a logger instance
 * @param {string} serviceName - Name of the service
 * @param {Object} logger - Optional logger instance to use
 * @returns {Object} Logger instance
 */
function createLogger(serviceName, logger = null) {
  if (logger) {
    return logger.child({ service: serviceName });
  }
  
  // Try to get platform logger
  const platformLogger = getPlatformLogger();
  if (platformLogger && platformLogger.child) {
    return platformLogger.child({ service: serviceName });
  }
  
  // Fallback to simple logger
  return createConsoleLogger(serviceName);
}

/**
 * Create a simple console logger
 * @param {string} serviceName - Name of the service
 * @returns {Object} Console logger instance
 */
function createConsoleLogger(serviceName) {
  const logLevel = process.env.LOG_LEVEL || 'INFO';
  const levels = ['TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'];
  const currentLevelIndex = levels.indexOf(logLevel.toUpperCase());

  const shouldLog = (level) => {
    return levels.indexOf(level.toUpperCase()) >= currentLevelIndex;
  };

  const formatLog = (level, message, data = {}) => {
    const timestamp = new Date().toISOString();
    const logEntry = {
      timestamp,
      level,
      service: serviceName,
      message,
      ...data
    };
    
    if (process.env.NODE_ENV === 'production') {
      return JSON.stringify(logEntry);
    }
    
    return `[${timestamp}] [${level}] [${serviceName}] ${message} ${Object.keys(data).length > 0 ? JSON.stringify(data, null, 2) : ''}`;
  };

  return {
    trace: (message, data) => shouldLog('TRACE') && console.debug(formatLog('TRACE', message, data)),
    debug: (message, data) => shouldLog('DEBUG') && console.debug(formatLog('DEBUG', message, data)),
    info: (message, data) => shouldLog('INFO') && console.info(formatLog('INFO', message, data)),
    warn: (message, data) => shouldLog('WARN') && console.warn(formatLog('WARN', message, data)),
    error: (message, data) => shouldLog('ERROR') && console.error(formatLog('ERROR', message, data)),
    fatal: (message, data) => shouldLog('FATAL') && console.error(formatLog('FATAL', message, data)),
    
    child: (context) => {
      return createConsoleLogger(`${serviceName}:${context.service || context.namespace || 'child'}`);
    },
    
    startTimer: (label) => {
      const start = Date.now();
      return (data = {}) => {
        const duration = Date.now() - start;
        if (shouldLog('INFO')) {
          console.info(formatLog('INFO', `${label} completed`, { duration, ...data }));
        }
        return duration;
      };
    }
  };
}

module.exports = {
  createLogger,
  setPlatformLogger,
  getPlatformLogger,
  createConsoleLogger
};

// Backward compatibility for AWS Lambda
// Try to create PowerTools logger if available
if (typeof exports !== 'undefined') {
  try {
    const { Logger } = require('@aws-lambda-powertools/logger');
    exports.createLogger = (serviceName) => {
      return new Logger({
        serviceName,
        logLevel: process.env.LOG_LEVEL || 'INFO'
      });
    };
  } catch (e) {
    // PowerTools not available, use fallback
    exports.createLogger = createLogger;
  }
}