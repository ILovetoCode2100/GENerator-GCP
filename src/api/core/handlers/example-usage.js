/**
 * Example Usage of Core API Handlers
 * Demonstrates how to use the platform-agnostic handlers
 */

const { platformFactory } = require('../../abstractions/platform-factory');
const { ProjectHandler, GoalHandler, JourneyHandler } = require('./index');

/**
 * Example configuration
 */
const config = {
  baseUrl: 'https://api.virtuoso.qa/api',
  timeout: 30000,
  retryConfig: {
    retries: 3,
    minTimeout: 1000,
    maxTimeout: 5000
  }
};

/**
 * Initialize handlers with platform services
 */
async function initializeHandlers() {
  // Create platform-specific services
  const services = platformFactory.createPlatformServices({
    logger: { serviceName: 'virtuoso-api' },
    secrets: { region: 'us-east-1' }
  });

  // Initialize handlers
  const projectHandler = new ProjectHandler(config, services);
  const goalHandler = new GoalHandler(config, services);
  const journeyHandler = new JourneyHandler(config, services);

  return {
    projectHandler,
    goalHandler,
    journeyHandler,
    services
  };
}

/**
 * Example: Project Operations
 */
async function projectExamples() {
  const { projectHandler } = await initializeHandlers();

  try {
    // List all projects
    const projects = await projectHandler.listProjects({
      query: { limit: 10 }
    });
    console.log('Projects:', projects);

    // Create a new project
    const newProject = await projectHandler.createProject({
      body: {
        name: 'My Test Project',
        description: 'A project created via API'
      }
    });
    console.log('Created project:', newProject);

    // List goals for a project
    if (projects.data && projects.data.length > 0) {
      const projectGoals = await projectHandler.listProjectGoals({
        params: { projectId: projects.data[0].id }
      });
      console.log('Project goals:', projectGoals);
    }

  } catch (error) {
    console.error('Project operations failed:', error);
  }
}

/**
 * Example: Goal Operations
 */
async function goalExamples() {
  const { goalHandler } = await initializeHandlers();

  try {
    // Create a new goal
    const newGoal = await goalHandler.createGoal({
      body: {
        name: 'My Test Goal',
        description: 'A goal created via API'
      }
    });
    console.log('Created goal:', newGoal);

    // Get goal versions (if goal exists)
    if (newGoal.id) {
      const versions = await goalHandler.getGoalVersions({
        params: { goalId: newGoal.id }
      });
      console.log('Goal versions:', versions);
    }

  } catch (error) {
    console.error('Goal operations failed:', error);
  }
}

/**
 * Example: Journey Operations (Including Holy Grail)
 */
async function journeyExamples() {
  const { journeyHandler } = await initializeHandlers();

  try {
    // List journeys with status
    const journeysWithStatus = await journeyHandler.listJourneysWithStatus({
      query: { limit: 5 }
    });
    console.log('Journeys with status:', journeysWithStatus);

    // Get detailed journey information (Holy Grail endpoint)
    if (journeysWithStatus.data && journeysWithStatus.data.length > 0) {
      const journeyId = journeysWithStatus.data[0].id;
      
      const journeyDetails = await journeyHandler.getJourneyDetails({
        params: { journeyId }
      });
      console.log('Journey details (Holy Grail):', journeyDetails);
    }

    // Create a new journey
    const newJourney = await journeyHandler.createJourney({
      body: {
        name: 'My Test Journey',
        description: 'A journey created via API'
      }
    });
    console.log('Created journey:', newJourney);

  } catch (error) {
    console.error('Journey operations failed:', error);
  }
}

/**
 * Example: Using Handler Methods with Platform Runtime
 */
async function runtimeExample() {
  const { projectHandler, services } = await initializeHandlers();

  // Get wrapped handler methods
  const handlers = projectHandler.getHandlers();

  // Example request (simulating platform-specific request format)
  const request = {
    method: 'GET',
    path: '/projects',
    params: {},
    query: { limit: 10 },
    body: {}
  };

  try {
    // Use the wrapped handler
    const response = await handlers.listProjects(request);
    console.log('Handler response:', response);

  } catch (error) {
    console.error('Handler execution failed:', error);
  }
}

// Export examples for testing
module.exports = {
  initializeHandlers,
  projectExamples,
  goalExamples,
  journeyExamples,
  runtimeExample
};

// Run examples if called directly
if (require.main === module) {
  console.log('Running Virtuoso API Handler Examples...\n');
  
  console.log('Current Platform:', platformFactory.getPlatform());
  console.log('Is AWS:', platformFactory.isAWS());
  console.log('Is Serverless:', platformFactory.isServerless());
  console.log('');

  // Note: Uncomment to run examples (requires valid API token)
  // projectExamples();
  // goalExamples();
  // journeyExamples();
  // runtimeExample();
}