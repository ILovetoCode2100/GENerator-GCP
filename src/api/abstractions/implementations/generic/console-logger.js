/**
 * Console Logger Implementation
 * Simple structured logging using console methods
 */

const { LoggerInterface, LogLevel } = require('../../interfaces/logger.interface');

class ConsoleLogger extends LoggerInterface {
  constructor(config = {}) {
    super(config);
    
    this.serviceName = config.serviceName || 'virtuoso-api';
    this.logLevel = config.logLevel || process.env.LOG_LEVEL || 'info';
    this.timestamp = config.timestamp !== false;
    this.colors = config.colors !== false && process.stdout.isTTY;
    
    // Log level priorities
    this.levelPriorities = {
      [LogLevel.TRACE]: 0,
      [LogLevel.DEBUG]: 1,
      [LogLevel.INFO]: 2,
      [LogLevel.WARN]: 3,
      [LogLevel.ERROR]: 4,
      [LogLevel.FATAL]: 5
    };
    
    this.currentPriority = this.levelPriorities[this.logLevel.toLowerCase()] || 2;
    
    // Color codes for different log levels
    this.levelColors = {
      [LogLevel.TRACE]: '\x1b[90m',  // gray
      [LogLevel.DEBUG]: '\x1b[36m',  // cyan
      [LogLevel.INFO]: '\x1b[32m',   // green
      [LogLevel.WARN]: '\x1b[33m',   // yellow
      [LogLevel.ERROR]: '\x1b[31m',  // red
      [LogLevel.FATAL]: '\x1b[35m'   // magenta
    };
    
    this.resetColor = '\x1b[0m';
  }

  /**
   * Core logging method
   * @param {string} level - Log level
   * @param {string} message - Log message
   * @param {Object} data - Additional data
   */
  log(level, message, data = {}) {
    if (!this.isLevelEnabled(level)) {
      return;
    }

    const logEntry = this.formatLogEntry(level, message, data);
    const output = this.formatOutput(logEntry);
    
    // Choose console method based on level
    switch (level) {
      case LogLevel.TRACE:
      case LogLevel.DEBUG:
        console.debug(output);
        break;
      case LogLevel.INFO:
        console.info(output);
        break;
      case LogLevel.WARN:
        console.warn(output);
        break;
      case LogLevel.ERROR:
      case LogLevel.FATAL:
        console.error(output);
        break;
      default:
        console.log(output);
    }
  }

  /**
   * Format log entry
   * @param {string} level - Log level
   * @param {string} message - Log message
   * @param {Object} data - Additional data
   * @returns {Object} Formatted log entry
   */
  formatLogEntry(level, message, data) {
    const entry = {
      level: level.toUpperCase(),
      message,
      service: this.serviceName
    };

    if (this.timestamp) {
      entry.timestamp = new Date().toISOString();
    }

    // Merge context and data
    const mergedData = {
      ...this.context,
      ...data
    };

    // Add non-empty data
    if (Object.keys(mergedData).length > 0) {
      entry.data = mergedData;
    }

    return entry;
  }

  /**
   * Format output string
   * @param {Object} logEntry - Log entry object
   * @returns {string} Formatted output
   */
  formatOutput(logEntry) {
    const { level, timestamp, service, message, data } = logEntry;
    
    if (process.env.NODE_ENV === 'production' || process.env.LOG_FORMAT === 'json') {
      // JSON format for production
      return JSON.stringify(logEntry);
    }

    // Human-readable format for development
    const parts = [];
    
    if (timestamp) {
      parts.push(`[${timestamp}]`);
    }
    
    // Add colored level
    const levelStr = this.colors 
      ? `${this.levelColors[level.toLowerCase()] || ''}${level}${this.resetColor}`
      : level;
    parts.push(`[${levelStr}]`);
    
    parts.push(`[${service}]`);
    parts.push(message);
    
    // Format data if present
    if (data && Object.keys(data).length > 0) {
      const dataStr = this.formatData(data);
      parts.push(dataStr);
    }
    
    return parts.join(' ');
  }

  /**
   * Format data object for display
   * @param {Object} data - Data object
   * @returns {string} Formatted data
   */
  formatData(data) {
    // Handle errors specially
    if (data.error && data.error instanceof Error) {
      const error = data.error;
      data.error = {
        message: error.message,
        stack: error.stack,
        code: error.code,
        ...error
      };
    }
    
    // In development, pretty print
    if (process.env.NODE_ENV !== 'production') {
      try {
        return '\n' + JSON.stringify(data, null, 2)
          .split('\n')
          .map(line => '  ' + line)
          .join('\n');
      } catch (e) {
        // Fallback for circular references
        return `[Data: ${Object.keys(data).join(', ')}]`;
      }
    }
    
    // In production, single line
    try {
      return JSON.stringify(data);
    } catch (e) {
      return `[Data: ${Object.keys(data).join(', ')}]`;
    }
  }

  /**
   * Create a child logger with additional context
   * @param {Object} context - Additional context
   * @returns {ConsoleLogger} Child logger instance
   */
  child(context) {
    const childLogger = new ConsoleLogger(this.config);
    childLogger.context = { ...this.context, ...context };
    return childLogger;
  }

  /**
   * Check if a log level is enabled
   * @param {string} level - Log level to check
   * @returns {boolean} True if level is enabled
   */
  isLevelEnabled(level) {
    const levelPriority = this.levelPriorities[level] || 0;
    return levelPriority >= this.currentPriority;
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    return 'console';
  }

  /**
   * Create a progress logger
   * @param {string} task - Task name
   * @returns {Object} Progress logger
   */
  createProgress(task) {
    let count = 0;
    let lastUpdate = Date.now();
    
    return {
      update: (current, total, message = '') => {
        count++;
        const now = Date.now();
        
        // Update at most once per second in production
        if (process.env.NODE_ENV === 'production' && now - lastUpdate < 1000) {
          return;
        }
        
        lastUpdate = now;
        const percent = total ? Math.round((current / total) * 100) : 0;
        
        this.info(`${task} progress`, {
          current,
          total,
          percent,
          message
        });
      },
      
      complete: (message = 'Complete') => {
        this.info(`${task} ${message}`, {
          operations: count
        });
      }
    };
  }

  /**
   * Log with structured context
   * @param {string} action - Action being performed
   * @param {Object} details - Action details
   */
  logAction(action, details = {}) {
    this.info(action, {
      action,
      ...details
    });
  }

  /**
   * Create a namespace logger
   * @param {string} namespace - Namespace
   * @returns {ConsoleLogger} Namespaced logger
   */
  namespace(namespace) {
    return this.child({ namespace });
  }
}

module.exports = { ConsoleLogger };