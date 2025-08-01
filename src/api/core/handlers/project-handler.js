const { BaseHandler } = require('./base-handler');

/**
 * Project Handler - Manages Virtuoso project operations
 * Handles project creation, listing, and goal retrieval
 */
class ProjectHandler extends BaseHandler {
  constructor(config, services) {
    super(config, services);
  }

  /**
   * Create a new project
   * @param {Object} request - Request object
   * @param {Object} request.body - Project data
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} Created project data
   */
  async createProject(request) {
    const { body = {}, query = {} } = request;
    
    this.logger.info('Creating project', { body, query });
    
    return this.post('/projects', body, {}, query);
  }

  /**
   * List all projects
   * @param {Object} request - Request object  
   * @param {Object} request.query - Query parameters for filtering/pagination
   * @returns {Promise<Object>} List of projects
   */
  async listProjects(request) {
    const { query = {} } = request;
    
    this.logger.info('Listing projects', { query });
    
    return this.get('/projects', {}, query);
  }

  /**
   * List goals for a specific project
   * @param {Object} request - Request object
   * @param {Object} request.params - Path parameters
   * @param {string} request.params.projectId - Project ID
   * @param {Object} request.query - Query parameters
   * @returns {Promise<Object>} List of project goals
   */
  async listProjectGoals(request) {
    const { params = {}, query = {} } = request;
    
    this.validateRequired(params, ['projectId']);
    
    this.logger.info('Listing project goals', { 
      projectId: params.projectId,
      query 
    });
    
    return this.get('/projects/{projectId}/goals', params, query);
  }

  /**
   * Get all available handler methods
   * @returns {Object} Map of method names to handler functions
   */
  getHandlers() {
    return {
      createProject: this.createHandler(this.createProject),
      listProjects: this.createHandler(this.listProjects),
      listProjectGoals: this.createHandler(this.listProjectGoals)
    };
  }
}

module.exports = { ProjectHandler };