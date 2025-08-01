/**
 * JSON Configuration Manager Implementation
 * Manages configuration from multiple sources (env, files, objects)
 */

const { ConfigManagerInterface } = require('../../interfaces/config.interface');
const fs = require('fs');
const path = require('path');

class JsonConfigManager extends ConfigManagerInterface {
  constructor(options = {}) {
    super(options);
    
    this.config = {};
    this.sources = options.sources || ['env', 'file'];
    this.configFile = options.configFile || 'config.json';
    this.environment = options.environment || process.env.NODE_ENV || 'development';
    
    // Load configuration on initialization
    this.loadSync();
  }

  /**
   * Load configuration synchronously
   */
  loadSync() {
    this.config = {};
    
    // Load from each source in order (later sources override earlier)
    this.sources.forEach(source => {
      switch (source) {
        case 'env':
          this.loadFromEnv();
          break;
        case 'file':
          this.loadFromFileSync();
          break;
        case 'defaults':
          this.loadDefaults();
          break;
      }
    });
  }

  /**
   * Load configuration from source
   * @param {string} source - Configuration source
   * @returns {Promise<void>}
   */
  async load(source) {
    if (!source) {
      // Reload all sources
      this.config = {};
      
      for (const src of this.sources) {
        switch (src) {
          case 'env':
            this.loadFromEnv();
            break;
          case 'file':
            await this.loadFromFile();
            break;
          case 'defaults':
            this.loadDefaults();
            break;
        }
      }
    } else {
      // Load specific source
      switch (source) {
        case 'env':
          this.loadFromEnv();
          break;
        case 'file':
          await this.loadFromFile();
          break;
        case 'defaults':
          this.loadDefaults();
          break;
        default:
          throw new Error(`Unknown configuration source: ${source}`);
      }
    }
  }

  /**
   * Load configuration from environment variables
   */
  loadFromEnv() {
    // Load VIRTUOSO_ prefixed variables
    Object.entries(process.env).forEach(([key, value]) => {
      if (key.startsWith('VIRTUOSO_')) {
        const configKey = key
          .substring(9)
          .toLowerCase()
          .replace(/_/g, '.');
        
        this.setNested(configKey, this.parseValue(value));
      }
    });
    
    // Load specific known variables
    const mappings = {
      'NODE_ENV': 'environment',
      'LOG_LEVEL': 'logging.level',
      'API_TIMEOUT': 'api.timeout',
      'RETRY_COUNT': 'api.retry.count',
      'RETRY_MIN_TIMEOUT': 'api.retry.minTimeout',
      'RETRY_MAX_TIMEOUT': 'api.retry.maxTimeout'
    };
    
    Object.entries(mappings).forEach(([envKey, configKey]) => {
      if (process.env[envKey]) {
        this.setNested(configKey, this.parseValue(process.env[envKey]));
      }
    });
  }

  /**
   * Load configuration from file synchronously
   */
  loadFromFileSync() {
    try {
      // Try multiple file locations
      const locations = [
        path.join(process.cwd(), this.configFile),
        path.join(process.cwd(), 'config', this.configFile),
        path.join(process.cwd(), `config.${this.environment}.json`),
        path.join(__dirname, '../../../../config', this.configFile)
      ];
      
      for (const location of locations) {
        if (fs.existsSync(location)) {
          const content = fs.readFileSync(location, 'utf8');
          const fileConfig = JSON.parse(content);
          
          // Merge with existing config
          this.deepMerge(this.config, fileConfig);
          break;
        }
      }
    } catch (error) {
      // Ignore file errors in production
      if (process.env.NODE_ENV !== 'production') {
        console.warn(`Failed to load config file: ${error.message}`);
      }
    }
  }

  /**
   * Load configuration from file asynchronously
   */
  async loadFromFile() {
    try {
      const fs = require('fs').promises;
      
      // Try multiple file locations
      const locations = [
        path.join(process.cwd(), this.configFile),
        path.join(process.cwd(), 'config', this.configFile),
        path.join(process.cwd(), `config.${this.environment}.json`),
        path.join(__dirname, '../../../../config', this.configFile)
      ];
      
      for (const location of locations) {
        try {
          const content = await fs.readFile(location, 'utf8');
          const fileConfig = JSON.parse(content);
          
          // Merge with existing config
          this.deepMerge(this.config, fileConfig);
          break;
        } catch (e) {
          // Try next location
        }
      }
    } catch (error) {
      // Ignore file errors in production
      if (process.env.NODE_ENV !== 'production') {
        console.warn(`Failed to load config file: ${error.message}`);
      }
    }
  }

  /**
   * Load default configuration
   */
  loadDefaults() {
    const defaults = {
      environment: 'development',
      api: {
        baseUrl: 'https://api.virtuoso.qa/api',
        timeout: 30000,
        retry: {
          count: 3,
          minTimeout: 1000,
          maxTimeout: 5000
        }
      },
      logging: {
        level: 'info',
        format: 'json'
      },
      features: {
        caching: true,
        parallelExecution: true,
        autoProjectCreation: false
      }
    };
    
    this.deepMerge(this.config, defaults);
  }

  /**
   * Get configuration value by key
   * @param {string} key - Configuration key (supports dot notation)
   * @param {*} defaultValue - Default value if not found
   * @returns {*} Configuration value
   */
  get(key, defaultValue = undefined) {
    const keys = key.split('.');
    let value = this.config;
    
    for (const k of keys) {
      if (value && typeof value === 'object' && k in value) {
        value = value[k];
      } else {
        return defaultValue;
      }
    }
    
    return value;
  }

  /**
   * Set configuration value
   * @param {string} key - Configuration key (supports dot notation)
   * @param {*} value - Configuration value
   */
  set(key, value) {
    this.setNested(key, value);
  }

  /**
   * Get all configuration
   * @returns {Object} Complete configuration
   */
  getAll() {
    return this.deepClone(this.config);
  }

  /**
   * Set nested configuration value
   * @param {string} key - Dot notation key
   * @param {*} value - Value to set
   */
  setNested(key, value) {
    const keys = key.split('.');
    let current = this.config;
    
    for (let i = 0; i < keys.length - 1; i++) {
      const k = keys[i];
      if (!(k in current) || typeof current[k] !== 'object') {
        current[k] = {};
      }
      current = current[k];
    }
    
    current[keys[keys.length - 1]] = value;
  }

  /**
   * Deep merge objects
   * @param {Object} target - Target object
   * @param {Object} source - Source object
   */
  deepMerge(target, source) {
    Object.keys(source).forEach(key => {
      if (source[key] && typeof source[key] === 'object' && !Array.isArray(source[key])) {
        if (!target[key] || typeof target[key] !== 'object') {
          target[key] = {};
        }
        this.deepMerge(target[key], source[key]);
      } else {
        target[key] = source[key];
      }
    });
  }

  /**
   * Deep clone object
   * @param {Object} obj - Object to clone
   * @returns {Object} Cloned object
   */
  deepClone(obj) {
    if (obj === null || typeof obj !== 'object') return obj;
    if (obj instanceof Date) return new Date(obj);
    if (obj instanceof Array) return obj.map(item => this.deepClone(item));
    
    const cloned = {};
    Object.keys(obj).forEach(key => {
      cloned[key] = this.deepClone(obj[key]);
    });
    
    return cloned;
  }

  /**
   * Parse string value to appropriate type
   * @param {string} value - String value
   * @returns {*} Parsed value
   */
  parseValue(value) {
    // Boolean
    if (value === 'true') return true;
    if (value === 'false') return false;
    
    // Number
    if (/^\d+$/.test(value)) return parseInt(value, 10);
    if (/^\d*\.\d+$/.test(value)) return parseFloat(value);
    
    // JSON
    if (value.startsWith('{') || value.startsWith('[')) {
      try {
        return JSON.parse(value);
      } catch (e) {
        // Return as string if not valid JSON
      }
    }
    
    return value;
  }

  /**
   * Validate configuration
   * @param {Object} schema - Validation schema
   * @returns {Object} Validation result
   */
  validate(schema) {
    const errors = [];
    
    // Simple required field validation
    if (schema.required) {
      schema.required.forEach(field => {
        if (this.get(field) === undefined) {
          errors.push({
            field,
            message: `Required field missing: ${field}`
          });
        }
      });
    }
    
    // Type validation
    if (schema.types) {
      Object.entries(schema.types).forEach(([field, expectedType]) => {
        const value = this.get(field);
        if (value !== undefined) {
          const actualType = Array.isArray(value) ? 'array' : typeof value;
          if (actualType !== expectedType) {
            errors.push({
              field,
              message: `Type mismatch: ${field} should be ${expectedType}, got ${actualType}`
            });
          }
        }
      });
    }
    
    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    return 'json';
  }

  /**
   * Export configuration to file
   * @param {string} filePath - Export file path
   * @returns {Promise<void>}
   */
  async exportToFile(filePath) {
    const fs = require('fs').promises;
    const content = JSON.stringify(this.config, null, 2);
    await fs.writeFile(filePath, content, 'utf8');
  }
}

module.exports = { JsonConfigManager };