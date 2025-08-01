/**
 * Tenant Configuration Manager
 * Manages tenant-specific configurations for multi-tenant support
 */

/**
 * Default tenant configuration
 */
const DEFAULT_TENANT_CONFIG = {
  api: {
    baseUrl: process.env.VIRTUOSO_API_URL || 'https://api.virtuoso.qa/api',
    timeout: 30000,
    retry: {
      count: 3,
      minTimeout: 1000,
      maxTimeout: 5000
    }
  },
  features: {
    autoProjectCreation: false,
    parallelExecution: true,
    caching: true,
    rateLimit: {
      enabled: true,
      maxRequests: 1000,
      windowMs: 60000 // 1 minute
    }
  },
  logging: {
    level: 'info',
    includeDetails: true
  }
};

/**
 * Tenant configuration manager
 */
class TenantConfigManager {
  constructor(storage = null) {
    this.storage = storage;
    this.cache = new Map();
    this.defaultConfig = DEFAULT_TENANT_CONFIG;
  }

  /**
   * Get configuration for a specific tenant
   * @param {string} tenantId - Tenant identifier
   * @returns {Promise<Object>} Tenant configuration
   */
  async getTenantConfig(tenantId) {
    if (!tenantId) {
      return this.defaultConfig;
    }

    // Check cache first
    if (this.cache.has(tenantId)) {
      return this.cache.get(tenantId);
    }

    // Load from storage if available
    if (this.storage) {
      try {
        const storedConfig = await this.storage.get(`tenant:${tenantId}:config`);
        if (storedConfig) {
          const config = this.mergeWithDefaults(storedConfig);
          this.cache.set(tenantId, config);
          return config;
        }
      } catch (error) {
        console.warn(`Failed to load tenant config for ${tenantId}:`, error.message);
      }
    }

    // Return default config
    this.cache.set(tenantId, this.defaultConfig);
    return this.defaultConfig;
  }

  /**
   * Set configuration for a tenant
   * @param {string} tenantId - Tenant identifier
   * @param {Object} config - Tenant configuration
   * @returns {Promise<void>}
   */
  async setTenantConfig(tenantId, config) {
    if (!tenantId) {
      throw new Error('Tenant ID is required');
    }

    const mergedConfig = this.mergeWithDefaults(config);
    
    // Update cache
    this.cache.set(tenantId, mergedConfig);

    // Persist to storage if available
    if (this.storage) {
      await this.storage.set(`tenant:${tenantId}:config`, mergedConfig);
    }
  }

  /**
   * Get API configuration for a tenant
   * @param {string} tenantId - Tenant identifier
   * @returns {Promise<Object>} API configuration
   */
  async getApiConfig(tenantId) {
    const config = await this.getTenantConfig(tenantId);
    return config.api;
  }

  /**
   * Get feature flags for a tenant
   * @param {string} tenantId - Tenant identifier
   * @returns {Promise<Object>} Feature flags
   */
  async getFeatureFlags(tenantId) {
    const config = await this.getTenantConfig(tenantId);
    return config.features;
  }

  /**
   * Check if a feature is enabled for a tenant
   * @param {string} tenantId - Tenant identifier
   * @param {string} feature - Feature name
   * @returns {Promise<boolean>} True if enabled
   */
  async isFeatureEnabled(tenantId, feature) {
    const features = await this.getFeatureFlags(tenantId);
    return features[feature] === true;
  }

  /**
   * Get secret path for a tenant
   * @param {string} tenantId - Tenant identifier
   * @param {string} secretName - Secret name
   * @returns {string} Secret path
   */
  getTenantSecretPath(tenantId, secretName) {
    return `/virtuoso/tenant/${tenantId}/${secretName}`;
  }

  /**
   * Clear tenant configuration cache
   * @param {string} tenantId - Tenant identifier (optional)
   */
  clearCache(tenantId = null) {
    if (tenantId) {
      this.cache.delete(tenantId);
    } else {
      this.cache.clear();
    }
  }

  /**
   * Merge configuration with defaults
   * @param {Object} config - Partial configuration
   * @returns {Object} Merged configuration
   */
  mergeWithDefaults(config) {
    return this.deepMerge(this.defaultConfig, config);
  }

  /**
   * Deep merge objects
   * @param {Object} target - Target object
   * @param {Object} source - Source object
   * @returns {Object} Merged object
   */
  deepMerge(target, source) {
    const result = { ...target };
    
    for (const key in source) {
      if (source[key] && typeof source[key] === 'object' && !Array.isArray(source[key])) {
        result[key] = this.deepMerge(result[key] || {}, source[key]);
      } else {
        result[key] = source[key];
      }
    }
    
    return result;
  }

  /**
   * Validate tenant configuration
   * @param {Object} config - Configuration to validate
   * @returns {Object} Validation result
   */
  validateConfig(config) {
    const errors = [];

    // Validate API configuration
    if (config.api) {
      if (config.api.timeout && (typeof config.api.timeout !== 'number' || config.api.timeout < 0)) {
        errors.push('api.timeout must be a positive number');
      }
      
      if (config.api.retry) {
        if (config.api.retry.count && (typeof config.api.retry.count !== 'number' || config.api.retry.count < 0)) {
          errors.push('api.retry.count must be a positive number');
        }
      }
    }

    // Validate feature flags
    if (config.features) {
      Object.entries(config.features).forEach(([key, value]) => {
        if (typeof value !== 'boolean' && typeof value !== 'object') {
          errors.push(`features.${key} must be a boolean or object`);
        }
      });
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * Create tenant-aware configuration
   * @param {Object} baseConfig - Base configuration
   * @param {string} tenantId - Tenant identifier
   * @returns {Promise<Object>} Tenant-aware configuration
   */
  async createTenantAwareConfig(baseConfig, tenantId) {
    const tenantConfig = await this.getTenantConfig(tenantId);
    
    return {
      ...baseConfig,
      ...tenantConfig,
      tenant: {
        id: tenantId,
        secretPath: this.getTenantSecretPath(tenantId, 'api-token')
      }
    };
  }
}

/**
 * Create a tenant configuration manager
 * @param {Object} storage - Storage backend (optional)
 * @returns {TenantConfigManager} Manager instance
 */
function createTenantConfigManager(storage = null) {
  return new TenantConfigManager(storage);
}

module.exports = {
  TenantConfigManager,
  createTenantConfigManager,
  DEFAULT_TENANT_CONFIG
};