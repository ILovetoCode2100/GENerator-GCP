/**
 * AWS Lambda PowerTools Logger Implementation
 * Provides structured logging with AWS Lambda PowerTools
 */

const { Logger } = require('@aws-lambda-powertools/logger');
const { LoggerInterface, LogLevel } = require('../../interfaces/logger.interface');

class PowerToolsLogger extends LoggerInterface {
  constructor(config = {}) {
    super(config);
    
    // Initialize PowerTools logger
    this.logger = new Logger({
      serviceName: config.serviceName || 'virtuoso-api',
      logLevel: config.logLevel || process.env.LOG_LEVEL || 'INFO',
      sampleRateValue: config.sampleRate || 0.1
    });

    // Map our log levels to PowerTools levels
    this.levelMap = {
      [LogLevel.TRACE]: 'debug', // PowerTools doesn't have trace
      [LogLevel.DEBUG]: 'debug',
      [LogLevel.INFO]: 'info',
      [LogLevel.WARN]: 'warn',
      [LogLevel.ERROR]: 'error',
      [LogLevel.FATAL]: 'error' // PowerTools doesn't have fatal
    };
  }

  /**
   * Core logging method
   * @param {string} level - Log level
   * @param {string} message - Log message
   * @param {Object} data - Additional data
   */
  log(level, message, data = {}) {
    const powerToolsLevel = this.levelMap[level] || 'info';
    const logData = {
      ...this.context,
      ...data
    };

    // Special handling for fatal level
    if (level === LogLevel.FATAL) {
      logData.fatal = true;
    }

    switch (powerToolsLevel) {
      case 'debug':
        this.logger.debug(message, logData);
        break;
      case 'info':
        this.logger.info(message, logData);
        break;
      case 'warn':
        this.logger.warn(message, logData);
        break;
      case 'error':
        this.logger.error(message, logData);
        break;
      default:
        this.logger.info(message, logData);
    }
  }

  /**
   * Create a child logger with additional context
   * @param {Object} context - Additional context
   * @returns {PowerToolsLogger} Child logger instance
   */
  child(context) {
    const childConfig = {
      ...this.config,
      serviceName: this.logger.serviceName
    };
    
    const childLogger = new PowerToolsLogger(childConfig);
    childLogger.context = { ...this.context, ...context };
    
    // Copy PowerTools persistent attributes
    Object.entries(this.logger.getPersistentLogAttributes()).forEach(([key, value]) => {
      childLogger.logger.appendKeys({ [key]: value });
    });
    
    // Add new context
    childLogger.logger.appendKeys(context);
    
    return childLogger;
  }

  /**
   * Check if a log level is enabled
   * @param {string} level - Log level to check
   * @returns {boolean} True if level is enabled
   */
  isLevelEnabled(level) {
    const currentLevel = this.logger.level;
    const levelPriority = {
      [LogLevel.TRACE]: 0,
      [LogLevel.DEBUG]: 0,
      [LogLevel.INFO]: 1,
      [LogLevel.WARN]: 2,
      [LogLevel.ERROR]: 3,
      [LogLevel.FATAL]: 3
    };

    const currentPriority = {
      'DEBUG': 0,
      'INFO': 1,
      'WARN': 2,
      'ERROR': 3
    }[currentLevel] || 1;

    return levelPriority[level] >= currentPriority;
  }

  /**
   * Add Lambda context to logger
   * @param {Object} lambdaContext - Lambda context object
   */
  addLambdaContext(lambdaContext) {
    if (lambdaContext && lambdaContext.awsRequestId) {
      this.logger.addContext(lambdaContext);
    }
  }

  /**
   * Inject Lambda context for correlation
   * @param {Object} event - Lambda event
   * @param {Object} context - Lambda context
   */
  injectLambdaContext(event, context) {
    this.logger.injectLambdaContext(event, context);
  }

  /**
   * Create cold start metric
   */
  logColdStart() {
    this.logger.info('Cold start', { 
      coldStart: true,
      runtime: process.version,
      memoryLimit: process.env.AWS_LAMBDA_FUNCTION_MEMORY_SIZE
    });
  }

  /**
   * Log metric
   * @param {string} name - Metric name
   * @param {number} value - Metric value
   * @param {string} unit - Metric unit
   */
  logMetric(name, value, unit = 'Count') {
    this.logger.info('Metric', {
      metric: {
        name,
        value,
        unit
      }
    });
  }

  /**
   * Start a segment for tracing
   * @param {string} name - Segment name
   * @returns {Object} Segment object
   */
  startSegment(name) {
    const segment = {
      name,
      startTime: Date.now(),
      correlationId: this.logger.correlationId
    };

    this.logger.debug(`Segment started: ${name}`, { segment });

    return {
      ...segment,
      end: (metadata = {}) => {
        const duration = Date.now() - segment.startTime;
        this.logger.debug(`Segment ended: ${name}`, {
          segment: {
            ...segment,
            duration,
            ...metadata
          }
        });
        return duration;
      }
    };
  }

  /**
   * Set correlation ID for distributed tracing
   * @param {string} correlationId - Correlation ID
   */
  setCorrelationId(correlationId) {
    this.logger.appendKeys({ correlationId });
  }

  /**
   * Clear all persistent attributes
   */
  clearAttributes() {
    this.logger.resetKeys();
    this.context = {};
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    return 'aws-powertools';
  }

  /**
   * Log API operation with PowerTools conventions
   * @param {string} operation - Operation name
   * @param {Object} details - Operation details
   */
  logApiOperation(operation, details) {
    this.logger.info(`API Operation: ${operation}`, {
      apiOperation: operation,
      ...details
    });
  }

  /**
   * Create structured error log
   * @param {Error} error - Error object
   * @param {Object} context - Additional context
   */
  logStructuredError(error, context = {}) {
    this.logger.error(error.message, {
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack,
        code: error.code,
        statusCode: error.statusCode,
        details: error.details
      },
      ...context
    });
  }
}

module.exports = { PowerToolsLogger };