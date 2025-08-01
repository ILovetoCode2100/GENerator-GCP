const { BaseHandler } = require('./base-handler');

/**
 * Execution Handler - Manages Virtuoso test execution operations
 * Handles goal execution, status monitoring, and analysis retrieval
 */
class ExecutionHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Execute a goal or test scenario
   * @param {Object} request - Request object
   * @param {Object} request.body - Execution parameters
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Execution initiation result
   */
  async executeGoal(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Executing goal', { body, query });
    
    return this.post('/executions', body, {}, query);
  }

  /**
   * Get the status of a specific execution
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.executionId - Execution ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Execution status details
   */
  async getExecutionStatus(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['executionId']);
    
    this.logger.info('Getting execution status', { 
      executionId: params.executionId,
      query 
    });
    
    return this.get('/executions/{executionId}', params, query);
  }

  /**
   * Get detailed analysis for a completed execution
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.executionId - Execution ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Execution analysis data
   */
  async getExecutionAnalysis(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['executionId']);
    
    this.logger.info('Getting execution analysis', { 
      executionId: params.executionId,
      query 
    });
    
    return this.get('/executions/analysis/{executionId}', params, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      executeGoal: this.createHandler(this.executeGoal),
      getExecutionStatus: this.createHandler(this.getExecutionStatus),
      getExecutionAnalysis: this.createHandler(this.getExecutionAnalysis)
    };
  }
}

module.exports = { ExecutionHandler };