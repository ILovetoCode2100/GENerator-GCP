/**
 * Virtuoso API Service
 * Main entry point for the platform-agnostic API management layer
 */

const { platformFactory } = require('./abstractions/platform-factory');
const { TenantConfigManager } = require('./config/tenant-config');

// Import all handlers
const { 
  ProjectHandler,
  GoalHandler,
  JourneyHandler,
  CheckpointHandler,
  StepHandler,
  ExecutionHandler,
  LibraryHandler,
  DataHandler,
  EnvironmentHandler
} = require('./core/handlers');

/**
 * Virtuoso API Service
 * Provides a unified interface for all Virtuoso API operations
 */
class VirtuosoApiService {
  /**
   * Initialize the API service
   * @param {Object} options - Service options
   * @param {string} options.platform - Platform type (aws-lambda, generic, etc.)
   * @param {string} options.tenantId - Tenant identifier for multi-tenant support
   * @param {Object} options.config - Custom configuration
   * @param {Object} options.storage - Storage backend for tenant configs
   */
  constructor(options = {}) {
    // Set platform if specified
    if (options.platform) {
      platformFactory.setPlatform(options.platform);
    }

    // Initialize platform services
    this.services = platformFactory.createPlatformServices(options.config || {});
    
    // Initialize tenant configuration manager
    this.tenantManager = new TenantConfigManager(options.storage);
    this.tenantId = options.tenantId;

    // Initialize handlers
    this.handlers = {};
    this.initializeHandlers();
  }

  /**
   * Initialize all API handlers
   */
  async initializeHandlers() {
    // Get tenant-specific configuration
    const config = await this.tenantManager.createTenantAwareConfig(
      this.services.configManager.getApiConfig(),
      this.tenantId
    );

    // Create handler instances
    const handlerClasses = {
      projects: ProjectHandler,
      goals: GoalHandler,
      journeys: JourneyHandler,
      checkpoints: CheckpointHandler,
      steps: StepHandler,
      executions: ExecutionHandler,
      library: LibraryHandler,
      data: DataHandler,
      environments: EnvironmentHandler
    };

    for (const [name, HandlerClass] of Object.entries(handlerClasses)) {
      this.handlers[name] = new HandlerClass(config, this.services);
    }

    // Create convenience properties
    this.projects = this.createHandlerProxy('projects');
    this.goals = this.createHandlerProxy('goals');
    this.journeys = this.createHandlerProxy('journeys');
    this.checkpoints = this.createHandlerProxy('checkpoints');
    this.steps = this.createHandlerProxy('steps');
    this.executions = this.createHandlerProxy('executions');
    this.library = this.createHandlerProxy('library');
    this.data = this.createHandlerProxy('data');
    this.environments = this.createHandlerProxy('environments');
  }

  /**
   * Create a proxy for handler methods
   * @param {string} handlerName - Handler name
   * @returns {Object} Proxy object
   */
  createHandlerProxy(handlerName) {
    const handler = this.handlers[handlerName];
    const proxy = {};

    // Get all methods from the handler
    const methods = Object.getOwnPropertyNames(Object.getPrototypeOf(handler))
      .filter(name => name !== 'constructor' && typeof handler[name] === 'function');

    // Create proxy methods
    for (const method of methods) {
      // Skip internal methods
      if (method.startsWith('_') || ['get', 'post', 'put', 'delete'].includes(method)) {
        continue;
      }

      proxy[method] = async (...args) => {
        // Add tenant context if not provided
        if (args.length === 1 && typeof args[0] === 'object' && !args[0].context) {
          args[0].context = { tenantId: this.tenantId };
        }

        return handler[method](...args);
      };
    }

    return proxy;
  }

  /**
   * Get raw handler instance
   * @param {string} name - Handler name
   * @returns {Object} Handler instance
   */
  getHandler(name) {
    return this.handlers[name];
  }

  /**
   * Create HTTP handlers for all endpoints
   * @returns {Object} Map of endpoints to handlers
   */
  createHttpHandlers() {
    const handlers = {};

    // Define endpoint mappings
    const endpoints = {
      // Projects
      'POST /projects': { handler: 'projects', method: 'createProject' },
      'GET /projects': { handler: 'projects', method: 'listProjects' },
      'GET /projects/:projectId/goals': { handler: 'projects', method: 'listProjectGoals' },
      
      // Goals
      'POST /goals': { handler: 'goals', method: 'createGoal' },
      'GET /goals/:goalId/versions': { handler: 'goals', method: 'getGoalVersions' },
      'POST /goals/:goalId/snapshots/:snapshotId/execute': { handler: 'goals', method: 'executeGoalSnapshot' },
      
      // Journeys
      'POST /testsuites': { handler: 'journeys', method: 'createJourney' },
      'POST /journeys': { handler: 'journeys', method: 'createJourneyAlt' },
      'GET /testsuites/latest_status': { handler: 'journeys', method: 'listJourneysWithStatus' },
      'GET /testsuites/:journeyId': { handler: 'journeys', method: 'getJourneyDetails' },
      'PUT /testsuites/:journeyId': { handler: 'journeys', method: 'updateJourney' },
      'POST /testsuites/:journeyId/checkpoints/attach': { handler: 'journeys', method: 'attachCheckpoint' },
      'POST /journeys/attach-library': { handler: 'journeys', method: 'attachLibraryCheckpoint' },
      
      // Checkpoints
      'POST /testcases': { handler: 'checkpoints', method: 'createCheckpoint' },
      'POST /checkpoints': { handler: 'checkpoints', method: 'createCheckpointAlt' },
      'GET /testcases/:checkpointId': { handler: 'checkpoints', method: 'getCheckpointDetails' },
      'GET /checkpoints/:checkpointId/teststeps': { handler: 'checkpoints', method: 'getCheckpointSteps' },
      'POST /testcases/:checkpointId/add-to-library': { handler: 'checkpoints', method: 'addCheckpointToLibrary' },
      
      // Steps
      'POST /teststeps': { handler: 'steps', method: 'addTestStep' },
      'POST /steps': { handler: 'steps', method: 'addTestStepAlt' },
      'GET /teststeps/:stepId': { handler: 'steps', method: 'getStepDetails' },
      'PUT /teststeps/:stepId/properties': { handler: 'steps', method: 'updateStepProperties' },
      
      // Executions
      'POST /executions': { handler: 'executions', method: 'executeGoal' },
      'GET /executions/:executionId': { handler: 'executions', method: 'getExecutionStatus' },
      'GET /executions/analysis/:executionId': { handler: 'executions', method: 'getExecutionAnalysis' },
      
      // Library
      'POST /library/checkpoints': { handler: 'library', method: 'addToLibrary' },
      'GET /library/checkpoints/:libraryCheckpointId': { handler: 'library', method: 'getLibraryCheckpoint' },
      'PUT /library/checkpoints/:libraryCheckpointId': { handler: 'library', method: 'updateLibraryCheckpoint' },
      'DELETE /library/checkpoints/:libraryCheckpointId/steps/:testStepId': { handler: 'library', method: 'removeLibraryStep' },
      'POST /library/checkpoints/:libraryCheckpointId/steps/:testStepId/move': { handler: 'library', method: 'moveLibraryStep' },
      
      // Data
      'POST /testdata/tables/create': { handler: 'data', method: 'createDataTable' },
      'GET /testdata/tables/:tableId': { handler: 'data', method: 'getDataTable' },
      'POST /testdata/tables/:tableId/import': { handler: 'data', method: 'importDataToTable' },
      
      // Environments
      'POST /environments': { handler: 'environments', method: 'createEnvironment' }
    };

    // Create handler functions
    for (const [endpoint, config] of Object.entries(endpoints)) {
      const handler = this.handlers[config.handler];
      const method = handler[config.method].bind(handler);
      
      handlers[endpoint] = handler.createHandler(method);
    }

    return handlers;
  }

  /**
   * Create platform-specific handler
   * @param {string} type - Handler type (express, lambda, etc.)
   * @returns {Function} Platform handler
   */
  createPlatformHandler(type = 'generic') {
    const runtime = this.services.runtime;
    const httpHandlers = this.createHttpHandlers();

    switch (type) {
      case 'express':
        return runtime.createExpressMiddleware(async (request) => {
          const endpoint = `${request.method} ${request.path}`;
          const handler = httpHandlers[endpoint];
          
          if (!handler) {
            throw new Error(`No handler for endpoint: ${endpoint}`);
          }
          
          return handler(request);
        });

      case 'lambda':
        return runtime.createHandler(async (request) => {
          const endpoint = `${request.method} ${request.path}`;
          const handler = httpHandlers[endpoint];
          
          if (!handler) {
            throw new Error(`No handler for endpoint: ${endpoint}`);
          }
          
          return handler(request);
        });

      default:
        return async (request) => {
          const endpoint = `${request.method} ${request.path}`;
          const handler = httpHandlers[endpoint];
          
          if (!handler) {
            throw new Error(`No handler for endpoint: ${endpoint}`);
          }
          
          return handler(request);
        };
    }
  }

  /**
   * Update tenant context
   * @param {string} tenantId - New tenant ID
   */
  async setTenant(tenantId) {
    this.tenantId = tenantId;
    await this.initializeHandlers();
  }

  /**
   * Get current platform services
   * @returns {Object} Platform services
   */
  getServices() {
    return this.services;
  }

  /**
   * Get tenant configuration manager
   * @returns {TenantConfigManager} Tenant manager
   */
  getTenantManager() {
    return this.tenantManager;
  }
}

/**
 * Create a new Virtuoso API service instance
 * @param {Object} options - Service options
 * @returns {VirtuosoApiService} API service instance
 */
function createApiService(options = {}) {
  return new VirtuosoApiService(options);
}

module.exports = {
  VirtuosoApiService,
  createApiService,
  platformFactory
};