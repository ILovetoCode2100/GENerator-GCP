const { BaseHandler } = require('./base-handler');

/**
 * Data Handler - Manages Virtuoso test data operations
 * Handles test data table creation, retrieval, and data import operations
 */
class DataHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Create a new test data table
   * @param {Object} request - Request object
   * @param {Object} request.body - Data table creation parameters
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created data table information
   */
  async createDataTable(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Creating data table', { body, query });
    
    return this.post('/testdata/tables/create', body, {}, query);
  }

  /**
   * Get a specific test data table
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.tableId - Data table ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Data table details and contents
   */
  async getDataTable(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['tableId']);
    
    this.logger.info('Getting data table', { 
      tableId: params.tableId,
      query 
    });
    
    return this.get('/testdata/tables/{tableId}', params, query);
  }

  /**
   * Import data into a test data table
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.tableId - Data table ID
   * @param {Object} request.body - Data import payload (CSV data, mapping, etc.)
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Data import result
   */
  async importDataToTable(request) {
    const { params = {}, body = {}, query = {} } = request;
    
    this.validateRequired(params, ['tableId']);
    
    this.logger.info('Importing data to table', { 
      tableId: params.tableId,
      body,
      query 
    });
    
    return this.post('/testdata/tables/{tableId}/import', body, params, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      createDataTable: this.createHandler(this.createDataTable),
      getDataTable: this.createHandler(this.getDataTable),
      importDataToTable: this.createHandler(this.importDataToTable)
    };
  }
}

module.exports = { DataHandler };