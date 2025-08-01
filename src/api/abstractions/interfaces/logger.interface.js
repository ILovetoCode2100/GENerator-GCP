/**
 * Logger Interface
 * Abstracts platform-specific logging implementations
 */

/**
 * Log levels enumeration
 */
const LogLevel = {
  TRACE: 'trace',
  DEBUG: 'debug',
  INFO: 'info',
  WARN: 'warn',
  ERROR: 'error',
  FATAL: 'fatal'
};

/**
 * Logger interface for structured logging
 * @interface
 */
class LoggerInterface {
  /**
   * Initialize the logger with configuration
   * @param {Object} config - Logger configuration
   * @param {string} config.serviceName - Service name for log context
   * @param {string} config.logLevel - Minimum log level
   */
  constructor(config) {
    if (new.target === LoggerInterface) {
      throw new Error('LoggerInterface is an abstract class');
    }
    this.config = config;
    this.context = {};
  }

  /**
   * Log a trace level message
   * @param {string} message - Log message
   * @param {Object} data - Additional data to log
   */
  trace(message, data = {}) {
    this.log(LogLevel.TRACE, message, data);
  }

  /**
   * Log a debug level message
   * @param {string} message - Log message
   * @param {Object} data - Additional data to log
   */
  debug(message, data = {}) {
    this.log(LogLevel.DEBUG, message, data);
  }

  /**
   * Log an info level message
   * @param {string} message - Log message
   * @param {Object} data - Additional data to log
   */
  info(message, data = {}) {
    this.log(LogLevel.INFO, message, data);
  }

  /**
   * Log a warning level message
   * @param {string} message - Log message
   * @param {Object} data - Additional data to log
   */
  warn(message, data = {}) {
    this.log(LogLevel.WARN, message, data);
  }

  /**
   * Log an error level message
   * @param {string} message - Log message
   * @param {Object|Error} data - Additional data or error to log
   */
  error(message, data = {}) {
    if (data instanceof Error) {
      data = {
        error: {
          message: data.message,
          stack: data.stack,
          code: data.code,
          ...data
        }
      };
    }
    this.log(LogLevel.ERROR, message, data);
  }

  /**
   * Log a fatal level message
   * @param {string} message - Log message
   * @param {Object|Error} data - Additional data or error to log
   */
  fatal(message, data = {}) {
    if (data instanceof Error) {
      data = {
        error: {
          message: data.message,
          stack: data.stack,
          code: data.code,
          ...data
        }
      };
    }
    this.log(LogLevel.FATAL, message, data);
  }

  /**
   * Core logging method - must be implemented by subclass
   * @param {string} level - Log level
   * @param {string} message - Log message
   * @param {Object} data - Additional data
   */
  log(level, message, data) {
    throw new Error('log must be implemented by subclass');
  }

  /**
   * Set persistent context for all logs
   * @param {Object} context - Context object
   */
  setContext(context) {
    this.context = { ...this.context, ...context };
  }

  /**
   * Clear persistent context
   */
  clearContext() {
    this.context = {};
  }

  /**
   * Create a child logger with additional context
   * @param {Object} context - Additional context
   * @returns {LoggerInterface} Child logger instance
   */
  child(context) {
    throw new Error('child must be implemented by subclass');
  }

  /**
   * Start a timer for performance measurement
   * @param {string} label - Timer label
   * @returns {Function} Function to stop timer and log duration
   */
  startTimer(label) {
    const start = Date.now();
    return (additionalData = {}) => {
      const duration = Date.now() - start;
      this.info(`${label} completed`, {
        duration,
        durationUnit: 'ms',
        ...additionalData
      });
      return duration;
    };
  }

  /**
   * Log API request
   * @param {Object} request - Request details
   */
  logApiRequest(request) {
    this.info('API request', {
      method: request.method,
      path: request.path,
      params: request.params,
      query: request.query,
      headers: this.sanitizeHeaders(request.headers)
    });
  }

  /**
   * Log API response
   * @param {Object} response - Response details
   * @param {number} duration - Request duration in ms
   */
  logApiResponse(response, duration) {
    const level = response.status >= 400 ? LogLevel.ERROR : LogLevel.INFO;
    this.log(level, 'API response', {
      status: response.status,
      duration,
      durationUnit: 'ms'
    });
  }

  /**
   * Sanitize headers for logging (remove sensitive data)
   * @param {Object} headers - Request headers
   * @returns {Object} Sanitized headers
   */
  sanitizeHeaders(headers) {
    const sanitized = { ...headers };
    const sensitiveHeaders = ['authorization', 'x-api-key', 'cookie'];
    
    sensitiveHeaders.forEach(header => {
      if (sanitized[header]) {
        sanitized[header] = '[REDACTED]';
      }
    });
    
    return sanitized;
  }

  /**
   * Check if a log level is enabled
   * @param {string} level - Log level to check
   * @returns {boolean} True if level is enabled
   */
  isLevelEnabled(level) {
    throw new Error('isLevelEnabled must be implemented by subclass');
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    throw new Error('getPlatform must be implemented by subclass');
  }
}

module.exports = { LoggerInterface, LogLevel };