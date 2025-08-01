const { BaseHandler } = require('./base-handler');

/**
 * Journey Handler - Manages Virtuoso journey/testsuite operations
 * Handles journey creation, management, and checkpoint operations
 * Includes the "holy grail" endpoint for complete test structure retrieval
 */
class JourneyHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Create a new journey (testsuite)
   * @param {Object} request - Request object
   * @param {Object} request.body - Journey data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created journey data
   */
  async createJourney(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Creating journey', { body, query });
    
    return this.post('/testsuites', body, {}, query);
  }

  /**
   * Create a new journey using alternative endpoint
   * @param {Object} request - Request object
   * @param {Object} request.body - Journey data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created journey data
   */
  async createJourneyAlt(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Creating journey (alt endpoint)', { body, query });
    
    return this.post('/journeys', body, {}, query);
  }

  /**
   * List journeys with their latest status
   * @param {Object} request - Request object
   * @param {Object} request.query - Query parameters for filtering
   * @returns {Promise<Object>} List of journeys with status
   */
  async listJourneysWithStatus(request) {
    const { query = {} } = request;
    
    this.logger.info('Listing journeys with status', { query });
    
    return this.get('/testsuites/latest_status', {}, query);
  }

  /**
   * Get detailed journey information (Holy Grail endpoint)
   * Returns complete test structure including journey details, 
   * checkpoints/test cases, and steps within each checkpoint
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.journeyId - Journey/TestSuite ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Complete journey structure
   */
  async getJourneyDetails(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['journeyId']);
    
    this.logger.info('Getting journey details (Holy Grail)', { 
      journeyId: params.journeyId,
      query 
    });
    
    return this.get('/testsuites/{journeyId}', params, query);
  }

  /**
   * Update an existing journey
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.journeyId - Journey ID
   * @param {Object} request.body - Updated journey data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Updated journey data
   */
  async updateJourney(request) {
    const { params = {}, body = {}, query = {} } = request;
    
    this.validateRequired(params, ['journeyId']);
    
    this.logger.info('Updating journey', { 
      journeyId: params.journeyId,
      body,
      query 
    });
    
    return this.put('/testsuites/{journeyId}', body, params, query);
  }

  /**
   * Attach a checkpoint to a journey
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.journeyId - Journey ID
   * @param {Object} request.body - Checkpoint attachment data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Attachment result
   */
  async attachCheckpoint(request) {
    const { params = {}, body = {}, query = {} } = request;
    
    this.validateRequired(params, ['journeyId']);
    
    this.logger.info('Attaching checkpoint to journey', { 
      journeyId: params.journeyId,
      body,
      query 
    });
    
    return this.post('/testsuites/{journeyId}/checkpoints/attach', body, params, query);
  }

  /**
   * Attach a library checkpoint to a journey
   * @param {Object} request - Request object
   * @param {Object} request.body - Library checkpoint data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Attachment result
   */
  async attachLibraryCheckpoint(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Attaching library checkpoint', { body, query });
    
    return this.post('/journeys/attach-library', body, {}, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      createJourney: this.createHandler(this.createJourney),
      createJourneyAlt: this.createHandler(this.createJourneyAlt),
      listJourneysWithStatus: this.createHandler(this.listJourneysWithStatus),
      getJourneyDetails: this.createHandler(this.getJourneyDetails), // Holy Grail endpoint
      updateJourney: this.createHandler(this.updateJourney),
      attachCheckpoint: this.createHandler(this.attachCheckpoint),
      attachLibraryCheckpoint: this.createHandler(this.attachLibraryCheckpoint)
    };
  }
}

module.exports = { JourneyHandler };