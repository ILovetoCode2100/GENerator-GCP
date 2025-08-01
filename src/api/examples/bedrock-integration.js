/**
 * Example: Integrating API layer with virtuoso-GENerator-bedrock
 */

const { createApiService } = require('../index');

/**
 * Create API service for Bedrock platform
 */
class BedrockApiIntegration {
  constructor(config = {}) {
    this.apiService = createApiService({
      platform: 'bedrock',
      tenantId: config.tenantId,
      storage: config.storage, // DynamoDB or other storage for tenant configs
      config: {
        secrets: {
          prefix: 'bedrock-virtuoso'
        },
        logger: {
          serviceName: 'bedrock-api-layer'
        }
      }
    });
  }

  /**
   * Use in AI conversion agent
   */
  async createTestResourcesFromConversion(convertedTest, tenantId) {
    // Set tenant context
    await this.apiService.setTenant(tenantId);

    try {
      // Create project if it doesn't exist
      const project = await this.apiService.projects.createProject({
        body: {
          name: convertedTest.projectName,
          description: `AI-converted from ${convertedTest.sourceType}`
        }
      });

      // Create goal
      const goal = await this.apiService.goals.createGoal({
        body: {
          projectId: project.id,
          name: convertedTest.goalName || 'Converted Test Goal'
        }
      });

      // Create journey (test suite)
      const journey = await this.apiService.journeys.createJourney({
        body: {
          goalId: goal.id,
          name: convertedTest.journeyName || 'Converted Test Journey',
          description: convertedTest.description
        }
      });

      // Create checkpoints and steps
      const checkpoints = [];
      for (const checkpoint of convertedTest.checkpoints) {
        const cp = await this.apiService.checkpoints.createCheckpoint({
          body: {
            journeyId: journey.id,
            name: checkpoint.name,
            description: checkpoint.description
          }
        });

        // Add steps to checkpoint
        for (const step of checkpoint.steps) {
          await this.apiService.steps.addTestStep({
            body: {
              checkpointId: cp.id,
              action: step.action,
              target: step.target,
              value: step.value,
              description: step.description
            }
          });
        }

        checkpoints.push(cp);
      }

      return {
        project,
        goal,
        journey,
        checkpoints
      };
    } catch (error) {
      this.apiService.getServices().logger.error('Failed to create test resources', {
        error,
        tenantId,
        testName: convertedTest.projectName
      });
      throw error;
    }
  }

  /**
   * Execute converted test
   */
  async executeConvertedTest(goalId, environmentId, tenantId) {
    await this.apiService.setTenant(tenantId);

    const execution = await this.apiService.executions.executeGoal({
      body: {
        goalId,
        environmentId,
        tags: ['ai-converted', 'bedrock']
      }
    });

    return execution;
  }

  /**
   * Monitor test execution
   */
  async monitorExecution(executionId, tenantId) {
    await this.apiService.setTenant(tenantId);

    const checkStatus = async () => {
      const status = await this.apiService.executions.getExecutionStatus({
        params: { executionId }
      });

      if (status.status === 'completed' || status.status === 'failed') {
        // Get detailed analysis
        const analysis = await this.apiService.executions.getExecutionAnalysis({
          params: { executionId }
        });

        return {
          status,
          analysis
        };
      }

      // Still running, check again in 5 seconds
      await new Promise(resolve => setTimeout(resolve, 5000));
      return checkStatus();
    };

    return checkStatus();
  }

  /**
   * Get test structure (Holy Grail endpoint)
   */
  async getCompleteTestStructure(journeyId, tenantId) {
    await this.apiService.setTenant(tenantId);

    return this.apiService.journeys.getJourneyDetails({
      params: { journeyId }
    });
  }

  /**
   * Integration with Bedrock agents
   */
  createAgentTools() {
    return {
      createProject: async (params) => {
        return this.apiService.projects.createProject({
          body: params
        });
      },

      createGoal: async (params) => {
        return this.apiService.goals.createGoal({
          body: params
        });
      },

      createJourney: async (params) => {
        return this.apiService.journeys.createJourney({
          body: params
        });
      },

      getJourneyDetails: async (journeyId) => {
        return this.apiService.journeys.getJourneyDetails({
          params: { journeyId }
        });
      },

      executeTest: async (goalId, environmentId) => {
        return this.apiService.executions.executeGoal({
          body: { goalId, environmentId }
        });
      }
    };
  }
}

/**
 * Example usage in Bedrock Lambda function
 */
exports.handler = async (event, context) => {
  const integration = new BedrockApiIntegration({
    tenantId: event.tenantId,
    storage: dynamoDbStorage // Your DynamoDB storage implementation
  });

  switch (event.action) {
    case 'convertAndCreate':
      return integration.createTestResourcesFromConversion(
        event.convertedTest,
        event.tenantId
      );

    case 'execute':
      return integration.executeConvertedTest(
        event.goalId,
        event.environmentId,
        event.tenantId
      );

    case 'monitor':
      return integration.monitorExecution(
        event.executionId,
        event.tenantId
      );

    case 'getStructure':
      return integration.getCompleteTestStructure(
        event.journeyId,
        event.tenantId
      );

    default:
      throw new Error(`Unknown action: ${event.action}`);
  }
};

module.exports = {
  BedrockApiIntegration,
  handler: exports.handler
};