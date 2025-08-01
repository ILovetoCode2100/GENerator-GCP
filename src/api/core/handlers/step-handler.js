const { BaseHandler } = require('./base-handler');

/**
 * Step Handler - Manages Virtuoso test step operations
 * Handles test step creation, management, and property updates
 */
class StepHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Add a new test step
   * @param {Object} request - Request object
   * @param {Object} request.body - Test step data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created test step data
   */
  async addTestStep(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Adding test step', { body, query });
    
    return this.post('/teststeps', body, {}, query);
  }

  /**
   * Add a new test step without envelope wrapping
   * @param {Object} request - Request object
   * @param {Object} request.body - Test step data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created test step data
   */
  async addTestStepNoEnvelope(request) {
    const { body = {}, query = {} } = request;
    
    // Merge envelope=false with existing query parameters
    const finalQuery = { ...query, envelope: 'false' };
    
    this.logger.info('Adding test step (no envelope)', { body, query: finalQuery });
    
    return this.post('/teststeps', body, {}, finalQuery);
  }

  /**
   * Add a new test step using alternative endpoint
   * @param {Object} request - Request object
   * @param {Object} request.body - Test step data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created test step data
   */
  async addTestStepAlt(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Adding test step (alt endpoint)', { body, query });
    
    return this.post('/steps', body, {}, query);
  }

  /**
   * Get detailed information for a specific test step
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.stepId - Step ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Test step details
   */
  async getStepDetails(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['stepId']);
    
    this.logger.info('Getting step details', { 
      stepId: params.stepId,
      query 
    });
    
    return this.get('/teststeps/{stepId}', params, query);
  }

  /**
   * Update properties of a specific test step
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.stepId - Step ID
   * @param {Object} request.body - Updated step properties
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Updated step data
   */
  async updateStepProperties(request) {
    const { params = {}, body = {}, query = {} } = request;
    
    this.validateRequired(params, ['stepId']);
    
    this.logger.info('Updating step properties', { 
      stepId: params.stepId,
      body,
      query 
    });
    
    return this.put('/teststeps/{stepId}/properties', body, params, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      addTestStep: this.createHandler(this.addTestStep),
      addTestStepNoEnvelope: this.createHandler(this.addTestStepNoEnvelope),
      addTestStepAlt: this.createHandler(this.addTestStepAlt),
      getStepDetails: this.createHandler(this.getStepDetails),
      updateStepProperties: this.createHandler(this.updateStepProperties)
    };
  }
}

module.exports = { StepHandler };