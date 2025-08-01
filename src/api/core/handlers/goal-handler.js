const { BaseHandler } = require('./base-handler');

/**
 * Goal Handler - Manages Virtuoso goal operations
 * Handles goal creation, version management, and snapshot execution
 */
class GoalHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Create a new goal
   * @param {Object} request - Request object
   * @param {Object} request.body - Goal data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created goal data
   */
  async createGoal(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Creating goal', { body, query });
    
    return this.post('/goals', body, {}, query);
  }

  /**
   * Get versions for a specific goal
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.goalId - Goal ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} List of goal versions
   */
  async getGoalVersions(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['goalId']);
    
    this.logger.info('Getting goal versions', { 
      goalId: params.goalId,
      query 
    });
    
    return this.get('/goals/{goalId}/versions', params, query);
  }

  /**
   * Execute a goal snapshot
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.goalId - Goal ID
   * @param {string} request.params.snapshotId - Snapshot ID
   * @param {Object} request.body - Execution parameters
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Execution result
   */
  async executeGoalSnapshot(request) {
    const { params = {}, body = {}, query = {} } = request;
    
    this.validateRequired(params, ['goalId', 'snapshotId']);
    
    this.logger.info('Executing goal snapshot', { 
      goalId: params.goalId,
      snapshotId: params.snapshotId,
      body,
      query 
    });
    
    return this.post('/goals/{goalId}/snapshots/{snapshotId}/execute', body, params, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      createGoal: this.createHandler(this.createGoal),
      getGoalVersions: this.createHandler(this.getGoalVersions),
      executeGoalSnapshot: this.createHandler(this.executeGoalSnapshot)
    };
  }
}

module.exports = { GoalHandler };