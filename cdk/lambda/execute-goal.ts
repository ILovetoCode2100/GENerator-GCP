import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoExecutionJob } from './shared/types';

interface ExecuteGoalRequest {
  startingUrl?: string;
}

class ExecuteGoalHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const pathParams = this.extractPathParameters(event);
      const requestBody = this.parseRequestBody<ExecuteGoalRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      const goalId = pathParams.goal_id;
      if (!goalId) {
        return this.createErrorResponse(400, 'Missing goal_id path parameter');
      }

      const execution = await this.makeVirtuosoRequest<VirtuosoExecutionJob>(
        'POST',
        `/goals/${goalId}/execute`,
        requestBody || {},
        authToken
      );

      return this.createResponse(200, execution);
    });
  }
}

export const handler = createHandler(ExecuteGoalHandler);