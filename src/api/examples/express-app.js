/**
 * Example: Using API layer with Express.js
 */

const express = require('express');
const { createApiService } = require('../index');

// Create Express app
const app = express();
app.use(express.json());

// Initialize API service
const apiService = createApiService({
  platform: 'express',
  config: {
    secrets: {
      parameterPrefix: '/virtuoso/express'
    },
    logger: {
      serviceName: 'virtuoso-express-api'
    }
  }
});

// Middleware to extract tenant ID
app.use((req, res, next) => {
  req.tenantId = req.headers['x-tenant-id'] || 'default';
  next();
});

// Create route handlers
const createRouteHandler = (handlerName, methodName) => {
  return async (req, res, next) => {
    try {
      // Set tenant context
      await apiService.setTenant(req.tenantId);
      
      // Create request object
      const request = {
        params: req.params,
        body: req.body,
        query: req.query,
        headers: req.headers,
        context: {
          tenantId: req.tenantId
        }
      };
      
      // Call handler method
      const result = await apiService[handlerName][methodName](request);
      
      res.json(result);
    } catch (error) {
      next(error);
    }
  };
};

// Project routes
app.post('/api/projects', createRouteHandler('projects', 'createProject'));
app.get('/api/projects', createRouteHandler('projects', 'listProjects'));
app.get('/api/projects/:projectId/goals', createRouteHandler('projects', 'listProjectGoals'));

// Goal routes
app.post('/api/goals', createRouteHandler('goals', 'createGoal'));
app.get('/api/goals/:goalId/versions', createRouteHandler('goals', 'getGoalVersions'));
app.post('/api/goals/:goalId/snapshots/:snapshotId/execute', createRouteHandler('goals', 'executeGoalSnapshot'));

// Journey routes
app.post('/api/testsuites', createRouteHandler('journeys', 'createJourney'));
app.post('/api/journeys', createRouteHandler('journeys', 'createJourneyAlt'));
app.get('/api/testsuites/latest_status', createRouteHandler('journeys', 'listJourneysWithStatus'));
app.get('/api/testsuites/:journeyId', createRouteHandler('journeys', 'getJourneyDetails')); // Holy Grail
app.put('/api/testsuites/:journeyId', createRouteHandler('journeys', 'updateJourney'));
app.post('/api/testsuites/:journeyId/checkpoints/attach', createRouteHandler('journeys', 'attachCheckpoint'));
app.post('/api/journeys/attach-library', createRouteHandler('journeys', 'attachLibraryCheckpoint'));

// Checkpoint routes
app.post('/api/testcases', createRouteHandler('checkpoints', 'createCheckpoint'));
app.post('/api/checkpoints', createRouteHandler('checkpoints', 'createCheckpointAlt'));
app.get('/api/testcases/:checkpointId', createRouteHandler('checkpoints', 'getCheckpointDetails'));
app.get('/api/checkpoints/:checkpointId/teststeps', createRouteHandler('checkpoints', 'getCheckpointSteps'));
app.post('/api/testcases/:checkpointId/add-to-library', createRouteHandler('checkpoints', 'addCheckpointToLibrary'));

// Step routes
app.post('/api/teststeps', createRouteHandler('steps', 'addTestStep'));
app.post('/api/steps', createRouteHandler('steps', 'addTestStepAlt'));
app.get('/api/teststeps/:stepId', createRouteHandler('steps', 'getStepDetails'));
app.put('/api/teststeps/:stepId/properties', createRouteHandler('steps', 'updateStepProperties'));

// Execution routes
app.post('/api/executions', createRouteHandler('executions', 'executeGoal'));
app.get('/api/executions/:executionId', createRouteHandler('executions', 'getExecutionStatus'));
app.get('/api/executions/analysis/:executionId', createRouteHandler('executions', 'getExecutionAnalysis'));

// Library routes
app.post('/api/library/checkpoints', createRouteHandler('library', 'addToLibrary'));
app.get('/api/library/checkpoints/:libraryCheckpointId', createRouteHandler('library', 'getLibraryCheckpoint'));
app.put('/api/library/checkpoints/:libraryCheckpointId', createRouteHandler('library', 'updateLibraryCheckpoint'));
app.delete('/api/library/checkpoints/:libraryCheckpointId/steps/:testStepId', createRouteHandler('library', 'removeLibraryStep'));
app.post('/api/library/checkpoints/:libraryCheckpointId/steps/:testStepId/move', createRouteHandler('library', 'moveLibraryStep'));

// Data routes
app.post('/api/testdata/tables/create', createRouteHandler('data', 'createDataTable'));
app.get('/api/testdata/tables/:tableId', createRouteHandler('data', 'getDataTable'));
app.post('/api/testdata/tables/:tableId/import', createRouteHandler('data', 'importDataToTable'));

// Environment routes
app.post('/api/environments', createRouteHandler('environments', 'createEnvironment'));

// Health check
app.get('/api/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: 'virtuoso-api-layer',
    platform: 'express',
    timestamp: new Date().toISOString()
  });
});

// Error handler
app.use((err, req, res, next) => {
  const logger = apiService.getServices().logger;
  logger.error('Request failed', { error: err, path: req.path });
  
  res.status(err.statusCode || 500).json({
    error: err.code || 'INTERNAL_ERROR',
    message: err.message,
    details: err.details
  });
});

// Start server
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Virtuoso API layer running on port ${PORT}`);
});

module.exports = app;