/**
 * Configuration Manager Interface
 * Abstracts platform-specific configuration management
 */

/**
 * Configuration manager interface for handling application configuration
 * @interface
 */
class ConfigManagerInterface {
  /**
   * Initialize the configuration manager
   * @param {Object} options - Configuration options
   */
  constructor(options = {}) {
    if (new.target === ConfigManagerInterface) {
      throw new Error('ConfigManagerInterface is an abstract class');
    }
    this.options = options;
    this.cache = new Map();
  }

  /**
   * Get a configuration value by key
   * @param {string} key - Configuration key (supports dot notation)
   * @param {*} defaultValue - Default value if key not found
   * @returns {*} Configuration value
   */
  get(key, defaultValue = undefined) {
    throw new Error('get must be implemented by subclass');
  }

  /**
   * Set a configuration value
   * @param {string} key - Configuration key (supports dot notation)
   * @param {*} value - Configuration value
   */
  set(key, value) {
    throw new Error('set must be implemented by subclass');
  }

  /**
   * Check if a configuration key exists
   * @param {string} key - Configuration key
   * @returns {boolean} True if key exists
   */
  has(key) {
    return this.get(key) !== undefined;
  }

  /**
   * Get multiple configuration values
   * @param {Array<string>} keys - Array of configuration keys
   * @returns {Object} Object with key-value pairs
   */
  getMultiple(keys) {
    const result = {};
    keys.forEach(key => {
      result[key] = this.get(key);
    });
    return result;
  }

  /**
   * Get all configuration as object
   * @returns {Object} Complete configuration object
   */
  getAll() {
    throw new Error('getAll must be implemented by subclass');
  }

  /**
   * Merge configuration from object
   * @param {Object} config - Configuration object to merge
   */
  merge(config) {
    Object.entries(config).forEach(([key, value]) => {
      this.set(key, value);
    });
  }

  /**
   * Get configuration for a specific environment
   * @param {string} env - Environment name
   * @returns {Object} Environment-specific configuration
   */
  getEnvironmentConfig(env) {
    return this.get(`environments.${env}`, {});
  }

  /**
   * Get current environment name
   * @returns {string} Current environment
   */
  getCurrentEnvironment() {
    return this.get('environment', 'development');
  }

  /**
   * Validate configuration against schema
   * @param {Object} schema - Validation schema
   * @returns {Object} Validation result
   */
  validate(schema) {
    // Default implementation - subclasses can override
    return { valid: true, errors: [] };
  }

  /**
   * Load configuration from source
   * @param {string} source - Configuration source identifier
   * @returns {Promise<void>}
   */
  async load(source) {
    throw new Error('load must be implemented by subclass');
  }

  /**
   * Reload configuration from sources
   * @returns {Promise<void>}
   */
  async reload() {
    this.cache.clear();
    await this.load(this.options.source);
  }

  /**
   * Watch for configuration changes
   * @param {Function} callback - Callback for configuration changes
   * @returns {Function} Unwatch function
   */
  watch(callback) {
    // Default implementation - no watching
    // Subclasses can implement actual watching
    return () => {};
  }

  /**
   * Get required configuration value (throws if not found)
   * @param {string} key - Configuration key
   * @returns {*} Configuration value
   * @throws {Error} If configuration key not found
   */
  require(key) {
    const value = this.get(key);
    if (value === undefined) {
      throw new Error(`Required configuration key not found: ${key}`);
    }
    return value;
  }

  /**
   * Get typed configuration value
   * @param {string} key - Configuration key
   * @param {string} type - Expected type (string, number, boolean, object, array)
   * @param {*} defaultValue - Default value
   * @returns {*} Typed configuration value
   */
  getTyped(key, type, defaultValue) {
    const value = this.get(key, defaultValue);
    
    switch (type) {
      case 'string':
        return String(value);
      case 'number':
        return Number(value);
      case 'boolean':
        return value === 'true' || value === true;
      case 'object':
        return typeof value === 'string' ? JSON.parse(value) : value;
      case 'array':
        return Array.isArray(value) ? value : [value];
      default:
        return value;
    }
  }

  /**
   * Get configuration with environment variable override
   * @param {string} key - Configuration key
   * @param {string} envVar - Environment variable name
   * @param {*} defaultValue - Default value
   * @returns {*} Configuration value
   */
  getWithEnvOverride(key, envVar, defaultValue) {
    // Check environment variable first
    if (process.env[envVar]) {
      return process.env[envVar];
    }
    return this.get(key, defaultValue);
  }

  /**
   * Get API configuration
   * @returns {Object} API-specific configuration
   */
  getApiConfig() {
    return {
      baseUrl: this.get('api.baseUrl', 'https://api.virtuoso.qa/api'),
      timeout: this.getTyped('api.timeout', 'number', 30000),
      retryConfig: {
        retries: this.getTyped('api.retry.count', 'number', 3),
        minTimeout: this.getTyped('api.retry.minTimeout', 'number', 1000),
        maxTimeout: this.getTyped('api.retry.maxTimeout', 'number', 5000)
      }
    };
  }

  /**
   * Get tenant-specific configuration
   * @param {string} tenantId - Tenant identifier
   * @returns {Object} Tenant configuration
   */
  getTenantConfig(tenantId) {
    return this.get(`tenants.${tenantId}`, {});
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    throw new Error('getPlatform must be implemented by subclass');
  }
}

module.exports = { ConfigManagerInterface };