const { BaseHandler } = require('./base-handler');

/**
 * Library Handler - Manages Virtuoso library operations
 * Handles library checkpoint management, step operations, and updates
 */
class LibraryHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Add a checkpoint to the library
   * @param {Object} request - Request object
   * @param {Object} request.body - Library checkpoint data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Library checkpoint creation result
   */
  async addToLibrary(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Adding checkpoint to library', { body, query });
    
    return this.post('/library/checkpoints', body, {}, query);
  }

  /**
   * Get a specific library checkpoint
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.libraryCheckpointId - Library checkpoint ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Library checkpoint details
   */
  async getLibraryCheckpoint(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['libraryCheckpointId']);
    
    this.logger.info('Getting library checkpoint', { 
      libraryCheckpointId: params.libraryCheckpointId,
      query 
    });
    
    return this.get('/library/checkpoints/{libraryCheckpointId}', params, query);
  }

  /**
   * Update a library checkpoint
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.libraryCheckpointId - Library checkpoint ID
   * @param {Object} request.body - Updated checkpoint data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Updated library checkpoint
   */
  async updateLibraryCheckpoint(request) {
    const { params = {}, body = {}, query = {} } = request;
    
    this.validateRequired(params, ['libraryCheckpointId']);
    
    this.logger.info('Updating library checkpoint', { 
      libraryCheckpointId: params.libraryCheckpointId,
      body,
      query 
    });
    
    return this.put('/library/checkpoints/{libraryCheckpointId}', body, params, query);
  }

  /**
   * Remove a test step from a library checkpoint
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.libraryCheckpointId - Library checkpoint ID
   * @param {string} request.params.testStepId - Test step ID to remove
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Step removal result
   */
  async removeLibraryStep(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['libraryCheckpointId', 'testStepId']);
    
    this.logger.info('Removing library step', { 
      libraryCheckpointId: params.libraryCheckpointId,
      testStepId: params.testStepId,
      query 
    });
    
    return this.delete('/library/checkpoints/{libraryCheckpointId}/steps/{testStepId}', params, query);
  }

  /**
   * Move a test step within a library checkpoint
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.libraryCheckpointId - Library checkpoint ID
   * @param {string} request.params.testStepId - Test step ID to move
   * @param {Object} request.body - Move operation data (position, etc.)
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Step move result
   */
  async moveLibraryStep(request) {
    const { params = {}, body = {}, query = {} } = request;
    
    this.validateRequired(params, ['libraryCheckpointId', 'testStepId']);
    
    this.logger.info('Moving library step', { 
      libraryCheckpointId: params.libraryCheckpointId,
      testStepId: params.testStepId,
      body,
      query 
    });
    
    return this.post('/library/checkpoints/{libraryCheckpointId}/steps/{testStepId}/move', body, params, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      addToLibrary: this.createHandler(this.addToLibrary),
      getLibraryCheckpoint: this.createHandler(this.getLibraryCheckpoint),
      updateLibraryCheckpoint: this.createHandler(this.updateLibraryCheckpoint),
      removeLibraryStep: this.createHandler(this.removeLibraryStep),
      moveLibraryStep: this.createHandler(this.moveLibraryStep)
    };
  }
}

module.exports = { LibraryHandler };