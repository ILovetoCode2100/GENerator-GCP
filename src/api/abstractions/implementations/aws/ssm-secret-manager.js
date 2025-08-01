/**
 * AWS SSM Parameter Store Secret Manager Implementation
 * Handles secure parameter storage and retrieval using AWS Systems Manager
 */

const { SSMClient, GetParameterCommand, PutParameterCommand, DeleteParameterCommand, GetParametersByPathCommand } = require('@aws-sdk/client-ssm');
const { SecretManagerInterface } = require('../../interfaces/secret.interface');

class SSMSecretManager extends SecretManagerInterface {
  constructor(config = {}) {
    super(config);
    
    // Initialize SSM client
    this.client = new SSMClient({
      region: config.region || process.env.AWS_REGION || 'us-east-1',
      ...config.clientConfig
    });
    
    // Cache configuration
    this.cache = new Map();
    this.cacheEnabled = config.cacheEnabled !== false;
    this.cacheTTL = config.cacheTTL || 300000; // 5 minutes default
  }

  /**
   * Get a secret value from SSM Parameter Store
   * @param {string} name - Parameter name
   * @param {Object} options - Additional options
   * @returns {Promise<string>} Parameter value
   */
  async getSecret(name, options = {}) {
    try {
      // Check cache first
      if (this.cacheEnabled && !options.bypassCache) {
        const cached = this.getFromCache(name);
        if (cached !== null) {
          return cached;
        }
      }

      // Get from SSM
      const command = new GetParameterCommand({
        Name: this.formatParameterName(name),
        WithDecryption: options.decrypt !== false
      });

      const response = await this.client.send(command);
      const value = response.Parameter.Value;

      // Cache the value
      if (this.cacheEnabled) {
        this.setCache(name, value);
      }

      return value;
    } catch (error) {
      if (error.name === 'ParameterNotFound') {
        const notFoundError = new Error(`Secret not found: ${name}`);
        notFoundError.code = 'SecretNotFound';
        throw notFoundError;
      }
      throw error;
    }
  }

  /**
   * Store a secret in SSM Parameter Store
   * @param {string} name - Parameter name
   * @param {string} value - Parameter value
   * @param {Object} options - Additional options
   * @returns {Promise<void>}
   */
  async setSecret(name, value, options = {}) {
    try {
      const command = new PutParameterCommand({
        Name: this.formatParameterName(name),
        Value: value,
        Type: options.type || 'SecureString',
        Overwrite: options.overwrite !== false,
        Description: options.description,
        Tags: options.tags
      });

      await this.client.send(command);

      // Clear cache for this parameter
      if (this.cacheEnabled) {
        this.cache.delete(name);
      }
    } catch (error) {
      throw new Error(`Failed to set secret ${name}: ${error.message}`);
    }
  }

  /**
   * Delete a secret from SSM Parameter Store
   * @param {string} name - Parameter name
   * @param {Object} options - Additional options
   * @returns {Promise<void>}
   */
  async deleteSecret(name, options = {}) {
    try {
      const command = new DeleteParameterCommand({
        Name: this.formatParameterName(name)
      });

      await this.client.send(command);

      // Clear cache for this parameter
      if (this.cacheEnabled) {
        this.cache.delete(name);
      }
    } catch (error) {
      if (error.name === 'ParameterNotFound') {
        // Ignore if already deleted
        return;
      }
      throw new Error(`Failed to delete secret ${name}: ${error.message}`);
    }
  }

  /**
   * List secrets with optional filtering
   * @param {Object} filter - Filter criteria
   * @returns {Promise<Array>} List of parameter metadata
   */
  async listSecrets(filter = {}) {
    try {
      const parameters = [];
      let nextToken;

      do {
        const command = new GetParametersByPathCommand({
          Path: filter.path || this.config.basePath || '/',
          Recursive: filter.recursive !== false,
          MaxResults: filter.maxResults || 50,
          NextToken: nextToken,
          ParameterFilters: filter.filters
        });

        const response = await this.client.send(command);
        
        parameters.push(...response.Parameters.map(param => ({
          name: param.Name,
          type: param.Type,
          lastModifiedDate: param.LastModifiedDate,
          version: param.Version,
          description: param.Description
        })));

        nextToken = response.NextToken;
      } while (nextToken && (!filter.limit || parameters.length < filter.limit));

      return parameters;
    } catch (error) {
      throw new Error(`Failed to list secrets: ${error.message}`);
    }
  }

  /**
   * Get secret with caching support
   * @param {string} name - Parameter name
   * @param {number} ttl - Cache TTL in seconds
   * @returns {Promise<string>} Parameter value
   */
  async getCachedSecret(name, ttl) {
    const customTTL = ttl ? ttl * 1000 : this.cacheTTL;
    const cached = this.getFromCache(name, customTTL);
    
    if (cached !== null) {
      return cached;
    }

    return this.getSecret(name);
  }

  /**
   * Format parameter name with prefix if configured
   * @param {string} name - Parameter name
   * @returns {string} Formatted parameter name
   */
  formatParameterName(name) {
    // Handle absolute paths
    if (name.startsWith('/')) {
      return name;
    }

    // Apply prefix if configured
    const prefix = this.config.parameterPrefix || '';
    return prefix ? `${prefix}/${name}` : name;
  }

  /**
   * Get value from cache
   * @param {string} key - Cache key
   * @param {number} customTTL - Custom TTL
   * @returns {*} Cached value or null
   */
  getFromCache(key, customTTL = null) {
    const cached = this.cache.get(key);
    if (!cached) return null;

    const ttl = customTTL || this.cacheTTL;
    if (Date.now() - cached.timestamp > ttl) {
      this.cache.delete(key);
      return null;
    }

    return cached.value;
  }

  /**
   * Set value in cache
   * @param {string} key - Cache key
   * @param {*} value - Value to cache
   */
  setCache(key, value) {
    this.cache.set(key, {
      value,
      timestamp: Date.now()
    });
  }

  /**
   * Clear all cached values
   */
  clearCache() {
    this.cache.clear();
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    return 'aws-ssm';
  }

  /**
   * Get API token helper method
   * @returns {Promise<string>} API token
   */
  async getApiToken() {
    const tokenPath = this.config.apiTokenPath || process.env.API_TOKEN_PARAM || '/virtuoso/api-token';
    return this.getSecret(tokenPath);
  }

  /**
   * Get tenant-specific secret
   * @param {string} tenantId - Tenant identifier
   * @param {string} secretName - Secret name
   * @returns {Promise<string>} Secret value
   */
  async getTenantSecret(tenantId, secretName) {
    const tenantSecretPath = `/virtuoso/tenant/${tenantId}/${secretName}`;
    return this.getSecret(tenantSecretPath);
  }
}

module.exports = { SSMSecretManager };