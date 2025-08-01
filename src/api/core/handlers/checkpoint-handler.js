const { BaseHandler } = require('./base-handler');

/**
 * Checkpoint Handler - Manages Virtuoso checkpoint/testcase operations
 * Handles checkpoint creation, management, step retrieval, and library operations
 */
class CheckpointHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Create a new checkpoint (testcase)
   * @param {Object} request - Request object
   * @param {Object} request.body - Checkpoint data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created checkpoint data
   */
  async createCheckpoint(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Creating checkpoint', { body, query });
    
    return this.post('/testcases', body, {}, query);
  }

  /**
   * Create a new checkpoint using alternative endpoint
   * @param {Object} request - Request object
   * @param {Object} request.body - Checkpoint data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created checkpoint data
   */
  async createCheckpointAlt(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Creating checkpoint (alt endpoint)', { body, query });
    
    return this.post('/checkpoints', body, {}, query);
  }

  /**
   * Get detailed checkpoint information
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.checkpointId - Checkpoint ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Checkpoint details
   */
  async getCheckpointDetails(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['checkpointId']);
    
    this.logger.info('Getting checkpoint details', { 
      checkpointId: params.checkpointId,
      query 
    });
    
    return this.get('/testcases/{checkpointId}', params, query);
  }

  /**
   * Get test steps for a specific checkpoint
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.checkpointId - Checkpoint ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} List of checkpoint test steps
   */
  async getCheckpointSteps(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['checkpointId']);
    
    this.logger.info('Getting checkpoint steps', { 
      checkpointId: params.checkpointId,
      query 
    });
    
    return this.get('/checkpoints/{checkpointId}/teststeps', params, query);
  }

  /**
   * Add a checkpoint to the library
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.checkpointId - Checkpoint ID
   * @param {Object} request.body - Library addition data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Library addition result
   */
  async addCheckpointToLibrary(request) {
    const { params = {}, body = {}, query = {} } = request;
    
    this.validateRequired(params, ['checkpointId']);
    
    this.logger.info('Adding checkpoint to library', { 
      checkpointId: params.checkpointId,
      body,
      query 
    });
    
    return this.post('/testcases/{checkpointId}/add-to-library', body, params, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      createCheckpoint: this.createHandler(this.createCheckpoint),
      createCheckpointAlt: this.createHandler(this.createCheckpointAlt),
      getCheckpointDetails: this.createHandler(this.getCheckpointDetails),
      getCheckpointSteps: this.createHandler(this.getCheckpointSteps),
      addCheckpointToLibrary: this.createHandler(this.addCheckpointToLibrary)
    };
  }
}

module.exports = { CheckpointHandler };