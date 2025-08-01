const { BaseHandler } = require('./base-handler');

/**
 * Environment Handler - Manages Virtuoso environment operations
 * Handles environment configuration and setup
 */
class EnvironmentHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Create a new environment configuration
   * @param {Object} request - Request object
   * @param {Object} request.body - Environment configuration data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created environment configuration
   */
  async createEnvironment(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Creating environment', { body, query });
    
    return this.post('/environments', body, {}, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      createEnvironment: this.createHandler(this.createEnvironment)
    };
  }
}

module.exports = { EnvironmentHandler };