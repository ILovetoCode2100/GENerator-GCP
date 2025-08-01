/**
 * Secret Manager Interface
 * Abstracts platform-specific secret/credential management
 */

/**
 * Secret manager interface for handling credentials and sensitive data
 * @interface
 */
class SecretManagerInterface {
  /**
   * Initialize the secret manager with configuration
   * @param {Object} config - Secret manager configuration
   */
  constructor(config) {
    if (new.target === SecretManagerInterface) {
      throw new Error('SecretManagerInterface is an abstract class');
    }
    this.config = config;
  }

  /**
   * Get a secret value by name
   * @param {string} name - Secret name/identifier
   * @param {Object} options - Additional options (version, decryption, etc.)
   * @returns {Promise<string>} Secret value
   */
  async getSecret(name, options = {}) {
    throw new Error('getSecret must be implemented by subclass');
  }

  /**
   * Store a secret value
   * @param {string} name - Secret name/identifier
   * @param {string} value - Secret value
   * @param {Object} options - Additional options (encryption, tags, etc.)
   * @returns {Promise<void>}
   */
  async setSecret(name, value, options = {}) {
    throw new Error('setSecret must be implemented by subclass');
  }

  /**
   * Delete a secret
   * @param {string} name - Secret name/identifier
   * @param {Object} options - Additional options (force delete, etc.)
   * @returns {Promise<void>}
   */
  async deleteSecret(name, options = {}) {
    throw new Error('deleteSecret must be implemented by subclass');
  }

  /**
   * List secrets with optional filtering
   * @param {Object} filter - Filter criteria
   * @returns {Promise<Array>} List of secret metadata
   */
  async listSecrets(filter = {}) {
    throw new Error('listSecrets must be implemented by subclass');
  }

  /**
   * Check if a secret exists
   * @param {string} name - Secret name/identifier
   * @returns {Promise<boolean>} True if secret exists
   */
  async secretExists(name) {
    try {
      await this.getSecret(name);
      return true;
    } catch (error) {
      if (error.code === 'SecretNotFound' || error.code === 'ParameterNotFound') {
        return false;
      }
      throw error;
    }
  }

  /**
   * Get secret with caching support
   * @param {string} name - Secret name/identifier
   * @param {number} ttl - Cache TTL in seconds
   * @returns {Promise<string>} Secret value
   */
  async getCachedSecret(name, ttl = 300) {
    // Default implementation without caching
    // Subclasses can override for caching support
    return this.getSecret(name);
  }

  /**
   * Rotate a secret
   * @param {string} name - Secret name/identifier
   * @param {string} newValue - New secret value
   * @returns {Promise<void>}
   */
  async rotateSecret(name, newValue) {
    await this.setSecret(name, newValue, { overwrite: true });
  }

  /**
   * Get platform name
   * @returns {string} Platform identifier
   */
  getPlatform() {
    throw new Error('getPlatform must be implemented by subclass');
  }
}

module.exports = { SecretManagerInterface };