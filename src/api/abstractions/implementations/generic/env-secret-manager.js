/**
 * Environment Variable Secret Manager Implementation
 * Simple secret management using environment variables
 */

const { SecretManagerInterface } = require('../../interfaces/secret.interface');

class EnvSecretManager extends SecretManagerInterface {
  constructor(config = {}) {
    super(config);
    this.prefix = config.prefix || 'VIRTUOSO_';
    this.secrets = new Map();
    
    // Load initial secrets from environment
    this.loadFromEnvironment();
  }

  /**
   * Load secrets from environment variables
   */
  loadFromEnvironment() {
    Object.entries(process.env).forEach(([key, value]) => {
      if (key.startsWith(this.prefix)) {
        const secretName = key.substring(this.prefix.length).toLowerCase().replace(/_/g, '-');
        this.secrets.set(secretName, value);
      }
    });
  }

  /**
   * Get a secret value
   * @param {string} name - Secret name
   * @param {Object} options - Additional options
   * @returns {Promise<string>} Secret value
   */
  async getSecret(name, options = {}) {
    // Check in-memory secrets first
    if (this.secrets.has(name)) {
      return this.secrets.get(name);
    }

    // Check environment variable directly
    const envName = this.formatEnvName(name);
    const value = process.env[envName];
    
    if (value !== undefined) {
      this.secrets.set(name, value);
      return value;
    }

    // Check without prefix
    const directValue = process.env[name];
    if (directValue !== undefined) {
      return directValue;
    }

    const error = new Error(`Secret not found: ${name}`);
    error.code = 'SecretNotFound';
    throw error;
  }

  /**
   * Store a secret value
   * @param {string} name - Secret name
   * @param {string} value - Secret value
   * @param {Object} options - Additional options
   * @returns {Promise<void>}
   */
  async setSecret(name, value, options = {}) {
    // Store in memory
    this.secrets.set(name, value);
    
    // Optionally set in process.env
    if (options.setInProcess !== false) {
      const envName = this.formatEnvName(name);
      process.env[envName] = value;
    }
  }

  /**
   * Delete a secret
   * @param {string} name - Secret name
   * @param {Object} options - Additional options
   * @returns {Promise<void>}
   */
  async deleteSecret(name, options = {}) {
    this.secrets.delete(name);
    
    // Optionally delete from process.env
    if (options.deleteFromProcess !== false) {
      const envName = this.formatEnvName(name);
      delete process.env[envName];
    }
  }

  /**
   * List secrets
   * @param {Object} filter - Filter criteria
   * @returns {Promise<Array>} List of secret metadata
   */
  async listSecrets(filter = {}) {
    const secrets = [];
    
    // List from memory
    for (const [name, value] of this.secrets.entries()) {
      if (filter.prefix && !name.startsWith(filter.prefix)) {
        continue;
      }
      
      secrets.push({
        name,
        source: 'memory',
        hasValue: true
      });
    }
    
    // List from environment
    Object.keys(process.env).forEach(key => {
      if (key.startsWith(this.prefix)) {
        const name = key.substring(this.prefix.length).toLowerCase().replace(/_/g, '-');
        
        if (!this.secrets.has(name)) {
          if (filter.prefix && !name.startsWith(filter.prefix)) {
            return;
          }
          
          secrets.push({
            name,
            source: 'environment',
            envVar: key,
            hasValue: true
          });
        }
      }
    });
    
    return secrets;
  }

  /**
   * Format secret name for environment variable
   * @param {string} name - Secret name
   * @returns {string} Environment variable name
   */
  formatEnvName(name) {
    const formatted = name.toUpperCase().replace(/-/g, '_').replace(/\//g, '_');
    return `${this.prefix}${formatted}`;
  }

  /**
   * Get secret with fallback chain
   * @param {string} name - Secret name
   * @param {Array<string>} fallbacks - Fallback names
   * @returns {Promise<string>} Secret value
   */
  async getSecretWithFallback(name, fallbacks = []) {
    try {
      return await this.getSecret(name);
    } catch (error) {
      if (error.code === 'SecretNotFound' && fallbacks.length > 0) {
        const [next, ...remaining] = fallbacks;
        return this.getSecretWithFallback(next, remaining);
      }
      throw error;
    }
  }

  /**
   * Get API token helper
   * @returns {Promise<string>} API token
   */
  async getApiToken() {
    const possibleNames = [
      'api-token',
      'api_token',
      'API_TOKEN',
      'VIRTUOSO_API_TOKEN'
    ];

    for (const name of possibleNames) {
      try {
        return await this.getSecret(name);
      } catch (error) {
        // Continue to next
      }
    }

    throw new Error('API token not found in environment variables');
  }

  /**
   * Get tenant-specific secret
   * @param {string} tenantId - Tenant identifier
   * @param {string} secretName - Secret name
   * @returns {Promise<string>} Secret value
   */
  async getTenantSecret(tenantId, secretName) {
    const tenantSecretName = `tenant-${tenantId}-${secretName}`;
    return this.getSecret(tenantSecretName);
  }

  /**
   * Load secrets from file
   * @param {string} filePath - Path to secrets file
   * @returns {Promise<void>}
   */
  async loadFromFile(filePath) {
    try {
      const fs = require('fs').promises;
      const content = await fs.readFile(filePath, 'utf8');
      
      // Support both JSON and .env formats
      if (filePath.endsWith('.json')) {
        const secrets = JSON.parse(content);
        Object.entries(secrets).forEach(([key, value]) => {
          this.secrets.set(key, value);
        });
      } else {
        // Parse .env format
        content.split('\n').forEach(line => {
          const trimmed = line.trim();
          if (trimmed && !trimmed.startsWith('#')) {
            const [key, ...valueParts] = trimmed.split('=');
            const value = valueParts.join('=').replace(/^["']|["']$/g, '');
            
            if (key.startsWith(this.prefix)) {
              const secretName = key.substring(this.prefix.length).toLowerCase().replace(/_/g, '-');
              this.secrets.set(secretName, value);
            }
          }
        });
      }
    } catch (error) {
      // Ignore file load errors in production
      if (process.env.NODE_ENV !== 'production') {
        console.warn(`Failed to load secrets from file: ${error.message}`);
      }
    }
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    return 'env';
  }
}

module.exports = { EnvSecretManager };