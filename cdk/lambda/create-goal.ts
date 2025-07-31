import { APIGatewayProxyEventV2, APIGatewayProxyResultV2, Context } from 'aws-lambda';
import { BaseHandler, createHandler } from './shared/base-handler';
import { VirtuosoGoal } from './shared/types';

interface CreateGoalRequest {
  name: string;
  projectId: string;
}

class CreateGoalHandler extends BaseHandler {
  async handle(event: APIGatewayProxyEventV2, context: Context): Promise<APIGatewayProxyResultV2> {
    return this.handleRequest(event, context, async () => {
      const authToken = this.extractAuthToken(event);
      const requestBody = this.parseRequestBody<CreateGoalRequest>(event);
      
      if (!authToken) {
        return this.createErrorResponse(401, 'Missing authorization token');
      }

      if (!requestBody || !requestBody.name || !requestBody.projectId) {
        return this.createErrorResponse(400, 'Missing required fields: name, projectId');
      }

      const goal = await this.makeVirtuosoRequest<VirtuosoGoal>(
        'POST',
        '/goals',
        requestBody,
        authToken
      );

      return this.createResponse(201, goal);
    });
  }
}

export const handler = createHandler(CreateGoalHandler);